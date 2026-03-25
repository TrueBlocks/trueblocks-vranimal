package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/browser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/converter"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/traverser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file.wrl>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nVRML97 3D Viewer - loads and renders .wrl files\n")
		os.Exit(1)
	}

	wrlPath := os.Args[1]
	baseDir := filepath.Dir(wrlPath)

	// Parse the VRML file
	f, err := os.Open(filepath.Clean(wrlPath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot open %s: %v\n", wrlPath, err)
		os.Exit(1)
	}

	p := parser.NewParser(f)
	p.SetBaseDir(baseDir)
	vrmlNodes := p.Parse()
	f.Close()

	if errs := p.Errors(); len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Parse warnings:\n")
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  %s\n", e)
		}
	}

	if len(vrmlNodes) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no nodes parsed from %s\n", wrlPath)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Parsed %d top-level nodes from %s\n", len(vrmlNodes), wrlPath)

	// Create g3n application
	a := app.App()
	scene := core.NewNode()

	// Convert VRML scene to g3n (returns node map for animation)
	nodeMap := converter.Convert(vrmlNodes, scene, baseDir)

	// Set up browser event engine
	br := browser.NewBrowser()
	br.Children = vrmlNodes
	br.Routes = p.GetRoutes()
	br.CollectTimeSensors()

	// Set up action traverser for sensors, LOD, etc.
	at := traverser.NewActionTraverser()
	at.CollectSensors(vrmlNodes)

	if len(br.TimeSensors) > 0 || len(br.Routes) > 0 {
		fmt.Fprintf(os.Stderr, "Event engine: %d TimeSensors, %d routes\n",
			len(br.TimeSensors), len(br.Routes))
	}
	if len(at.ProximitySensors) > 0 || len(at.TouchSensors) > 0 || len(at.LODs) > 0 {
		fmt.Fprintf(os.Stderr, "Action traverser: %d ProximitySensors, %d TouchSensors, %d LODs\n",
			len(at.ProximitySensors), len(at.TouchSensors), len(at.LODs))
	}

	// Set up camera from VRML Viewpoint if available
	cam := camera.New(1)
	vp := converter.GetViewpoint(vrmlNodes)
	if vp != nil {
		cam.SetPosition(vp.Position.X, vp.Position.Y, vp.Position.Z)
		cam.SetFov(float32(float64(vp.FieldOfView) * 180.0 / math.Pi))
	} else {
		cam.SetPosition(0, 2, 10)
	}
	cam.LookAt(&math32.Vector3{X: 0, Y: 0, Z: 0}, &math32.Vector3{X: 0, Y: 1, Z: 0})
	scene.Add(cam)

	// Orbit control for mouse interaction
	camera.NewOrbitControl(cam)

	// Add default lighting if no lights in the scene
	nav := converter.GetNavigationInfo(vrmlNodes)
	headlight := nav == nil || nav.Headlight // default is true per VRML97
	if !hasLights(vrmlNodes) {
		ambLight := light.NewAmbient(&math32.Color{R: 0.3, G: 0.3, B: 0.3}, 1.0)
		scene.Add(ambLight)
		if headlight {
			dirLight := light.NewDirectional(&math32.Color{R: 1, G: 1, B: 1}, 0.8)
			dirLight.SetPosition(5, 10, 5)
			scene.Add(dirLight)
		}
	} else if headlight {
		// Scene has lights but headlight is also requested
		dirLight := light.NewDirectional(&math32.Color{R: 1, G: 1, B: 1}, 0.5)
		dirLight.SetPosition(0, 0, 10)
		scene.Add(dirLight)
	}

	// Axes helper
	axes := helper.NewAxes(1.0)
	scene.Add(axes)

	// Background color
	bg := converter.GetBackground(vrmlNodes)
	if bg != nil && len(bg.SkyColor) > 0 {
		c := bg.SkyColor[0]
		a.Gls().ClearColor(c.R, c.G, c.B, 1.0)
	} else {
		a.Gls().ClearColor(0.2, 0.2, 0.3, 1.0)
	}

	// Handle window resize — use framebuffer size for viewport (Retina/HiDPI)
	onResize := func(evname string, ev interface{}) {
		fbW, fbH := a.GetFramebufferSize()
		a.Gls().Viewport(0, 0, int32(fbW), int32(fbH))
		cam.SetAspect(float32(fbW) / float32(fbH))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	fmt.Fprintf(os.Stderr, "Viewer ready. Drag to rotate, scroll to zoom.\n")

	// Render loop
	a.Run(func(rend *renderer.Renderer, deltaTime time.Duration) {
		// Get camera position for action traverser
		camPos := cam.Position()
		viewerPos := vec.SFVec3f{X: camPos.X, Y: camPos.Y, Z: camPos.Z}

		// Process action traverser (ProximitySensors, LOD distance, etc.)
		at.Update(viewerPos, br.SimTime())

		// Process VRML events (TimeSensors, routes, interpolators)
		br.Update(deltaTime)
		nodeMap.UpdateDynamic()

		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		_ = rend.Render(scene, cam)
	})
}

func hasLights(nodes []node.Node) bool {
	for _, n := range nodes {
		switch v := n.(type) {
		case *node.DirectionalLight, *node.PointLight, *node.SpotLight:
			return true
		case *node.Transform:
			if hasLights(v.Children) {
				return true
			}
		case *node.Group:
			if hasLights(v.Children) {
				return true
			}
		}
	}
	return false
}
