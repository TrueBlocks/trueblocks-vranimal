package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
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
	var wrlPath string
	if len(os.Args) >= 2 {
		wrlPath = os.Args[1]
	} else {
		// No file specified — find the most recently modified .wrl in examples/bool_demos/
		wrlPath = findMostRecentBoolDemo()
		if wrlPath == "" {
			fmt.Fprintf(os.Stderr, "Usage: %s [file.wrl]\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "\nNo file specified and no bool demo files found in examples/bool_demos/\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "No file specified — loading most recent: %s\n", wrlPath)
	}
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
	if err := f.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "close error: %v\n", err)
	}

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

	// Drop pre-computed result transforms from bool_demo files (keep only A and B)
	vrmlNodes = dropResultTransforms(vrmlNodes)

	// Create g3n application
	a := app.App()
	scene := core.NewNode()

	// Restore persisted window geometry, or maximize on first run
	if gw, ok := a.IWindow.(*window.GlfwWindow); ok {
		ss := loadSettings()
		if ss.WinWidth > 0 && ss.WinHeight > 0 {
			gw.SetSize(ss.WinWidth, ss.WinHeight)
			gw.SetPos(ss.WinX, ss.WinY)
		} else {
			w, h := gw.ScreenResolution(nil)
			gw.SetSize(w, h)
			gw.SetPos(0, 0)
		}
	}

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
		cam.SetPosition(float32(vp.Position.X), float32(vp.Position.Y), float32(vp.Position.Z))
		cam.SetFov(float32(float64(vp.FieldOfView) * 180.0 / math.Pi))
	} else {
		cam.SetPosition(0, 2, 10)
	}
	cam.LookAt(&math32.Vector3{X: 0, Y: 0, Z: 0}, &math32.Vector3{X: 0, Y: 1, Z: 0})
	scene.Add(cam)

	// Orbit control for mouse interaction
	oc := camera.NewOrbitControl(cam)

	// Apply NavigationInfo settings to orbit control
	nav := converter.GetNavigationInfo(vrmlNodes)
	if nav != nil {
		// Speed scales zoom and pan
		oc.ZoomSpeed = float32(nav.Speed * 0.1)
		oc.KeyPanSpeed = float32(nav.Speed * 35.0)
		oc.KeyZoomSpeed = float32(nav.Speed * 2.0)

		// Type controls enabled interactions
		navType := "WALK"
		if len(nav.Type) > 0 {
			navType = nav.Type[0]
		}
		switch navType {
		case "NONE":
			oc.SetEnabled(camera.OrbitNone)
		case "EXAMINE":
			oc.SetEnabled(camera.OrbitAll)
		case "FLY":
			oc.SetEnabled(camera.OrbitRot | camera.OrbitZoom | camera.OrbitKeys)
		default: // "WALK" and others
			oc.SetEnabled(camera.OrbitAll)
		}

		// VisibilityLimit → MaxDistance
		if nav.VisibilityLimit > 0 {
			oc.MaxDistance = float32(nav.VisibilityLimit)
		}
	}

	// Set up mouse picker for sensor interaction
	picker := traverser.NewPicker(scene, cam)
	picker.OnAnchor = func(urls []string, description string) {
		if len(urls) > 0 {
			fmt.Fprintf(os.Stderr, "Anchor activated: %s\n", urls[0])
			if err := exec.Command("open", urls[0]).Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to open URL: %v\n", err)
			}
		}
	}
	// Always subscribe to mouse events — bool demo meshes get TouchSensors dynamically.
	a.Subscribe(window.OnMouseDown, func(evname string, ev interface{}) {
		mev := ev.(*window.MouseEvent)
		if mev.Button != window.MouseButtonLeft {
			return
		}
		picker.SimTime = br.SimTime()
		picker.HandlePointer(float64(mev.Xpos), float64(mev.Ypos), traverser.PointerDown)
	})
	a.Subscribe(window.OnMouseUp, func(evname string, ev interface{}) {
		mev := ev.(*window.MouseEvent)
		if mev.Button != window.MouseButtonLeft {
			return
		}
		picker.SimTime = br.SimTime()
		picker.HandlePointer(float64(mev.Xpos), float64(mev.Ypos), traverser.PointerUp)
	})
	a.Subscribe(window.OnCursor, func(evname string, ev interface{}) {
		cev := ev.(*window.CursorEvent)
		picker.SimTime = br.SimTime()
		picker.HandlePointer(float64(cev.Xpos), float64(cev.Ypos), traverser.PointerMove)
	})

	// Add default lighting if no lights in the scene
	headlight := nav == nil || nav.Headlight // default is true per VRML97
	if !hasLights(vrmlNodes) {
		ambLight := light.NewAmbient(&math32.Color{R: 0.3, G: 0.3, B: 0.3}, 1.0)
		scene.Add(ambLight)
		if headlight {
			// Headlight follows the camera (child of cam node)
			dirLight := light.NewDirectional(&math32.Color{R: 1, G: 1, B: 1}, 0.8)
			dirLight.SetPosition(0, 0, 1)
			cam.Add(dirLight)
		}
	} else if headlight {
		// Scene has lights but headlight is also requested
		dirLight := light.NewDirectional(&math32.Color{R: 1, G: 1, B: 1}, 0.5)
		dirLight.SetPosition(0, 0, 1)
		cam.Add(dirLight)
	}

	// Axes helper
	axes := helper.NewAxes(1.0)
	scene.Add(axes)

	// Set up GUI manager and menu bar
	gui.Manager().Set(scene)

	vs := &viewerState{
		wireframe:   false,
		axesVisible: true,
		axes:        axes,
		scene:       scene,
		cam:         cam,
		oc:          oc,
		vrmlNodes:   vrmlNodes,
		wrlPath:     wrlPath,
		baseDir:     baseDir,
		browser:     br,
		picker:      picker,
		pendingLoad: make(chan string, 1),
	}
	vs.reloadFn = func(path string) { loadScene(vs, path) }

	mb := setupMenuBar(vs)
	scene.Add(mb)

	// Restore persisted settings (wireframe, last bool case)
	restoreLastCase(vs)

	// Background color
	bg := converter.GetBackground(vrmlNodes)
	if bg != nil && len(bg.SkyColor) > 0 {
		c := bg.SkyColor[0]
		a.Gls().ClearColor(float32(c.R), float32(c.G), float32(c.B), 1.0)
	} else {
		a.Gls().ClearColor(0.2, 0.2, 0.3, 1.0)
	}

	// Handle window resize — use framebuffer size for viewport (Retina/HiDPI)
	onResize := func(evname string, ev interface{}) {
		fbW, fbH := a.GetFramebufferSize()
		a.Gls().Viewport(0, 0, int32(fbW), int32(fbH))
		cam.SetAspect(float32(float64(fbW) / float64(fbH)))
		// Picker uses window coordinates (mouse events), not framebuffer size
		wW, wH := a.GetSize()
		picker.SetSize(wW, wH)
		persistSettings(vs)
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	// Persist window position on move
	a.Subscribe(window.OnWindowPos, func(evname string, ev interface{}) {
		persistSettings(vs)
	})

	// Delete key clears all geometry from the scene
	a.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		if kev.Key == window.KeyDelete || kev.Key == window.KeyBackspace {
			clearGeometry(vs)
		}
	})

	fmt.Fprintf(os.Stderr, "Viewer ready. Drag to rotate, scroll to zoom.\n")

	// Render loop
	a.Run(func(rend *renderer.Renderer, deltaTime time.Duration) {
		// Check for pending file load (from goroutine — must run on main thread)
		select {
		case path := <-vs.pendingLoad:
			loadScene(vs, path)
		default:
		}

		// Get camera position for action traverser
		camPos := cam.Position()
		viewerPos := vec.SFVec3f{X: float64(camPos.X), Y: float64(camPos.Y), Z: float64(camPos.Z)}

		// Process action traverser (ProximitySensors, LOD distance, etc.)
		at.Update(viewerPos, br.SimTime())

		// Process VRML events (TimeSensors, routes, interpolators)
		br.Update(deltaTime)
		nodeMap.CameraPos = camPos
		nodeMap.UpdateDynamic()

		// Sync pick-highlight colors from VRML Material to g3n meshes
		syncPickColors(vs)

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

// findMostRecentBoolDemo returns the path to the most recently modified .wrl
// file in examples/bool_demos/, or "" if none found.
func findMostRecentBoolDemo() string {
	candidates := []string{
		"examples/bool_demos",
		"trueblocks-vranimal/examples/bool_demos",
	}
	var dir string
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			dir = c
			break
		}
	}
	if dir == "" {
		return ""
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var bestPath string
	var bestTime time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".wrl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(bestTime) {
			bestTime = info.ModTime()
			bestPath = filepath.Join(dir, e.Name())
		}
	}
	return bestPath
}
