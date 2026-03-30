package main

import (
	"encoding/json"
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
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/browser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/converter"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/boolop"
	solidExport "github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/export"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/primitives"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/traverser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/writer"
)

// ────────────────────────── Bool case definitions ──────────────────────────

var (
	yellow    = vec.SFColor{R: 1.0, G: 0.9, B: 0.2, A: 1}
	lightBlue = vec.SFColor{R: 0.4, G: 0.6, B: 1.0, A: 1}
	selRed    = vec.SFColor{R: 1, G: 0, B: 0, A: 1}
	selPurple = vec.SFColor{R: 0.6, G: 0, B: 0.8, A: 1}
	selPink   = vec.SFColor{R: 1, G: 0.4, B: 0.7, A: 1}
)

type boolCase struct {
	name  string
	makeA func() *base.Solid
	makeB func() *base.Solid
}

var boolCaseMap = map[string]*boolCase{
	"same_size_partial_overlap": {
		name:  "same_size_partial_overlap",
		makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
		makeB: func() *base.Solid {
			s := primitives.MakeCube(1.0, lightBlue)
			boolDoTranslate(s, 0.5, 0.5, 0)
			return s
		},
	},
	"same_size_edge_on_edge": {
		name:  "same_size_edge_on_edge",
		makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
		makeB: func() *base.Solid {
			s := primitives.MakeCube(1.0, lightBlue)
			boolDoTranslate(s, 1.0, 1.0, 0)
			return s
		},
	},
	"same_size_slight_twist": {
		name:  "same_size_slight_twist",
		makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
		makeB: func() *base.Solid {
			s := primitives.MakeCube(1.0, lightBlue)
			boolDoRotateCenter(s, 15, vec.ZAxis)
			boolDoTranslate(s, 0.3, 0.3, 0)
			return s
		},
	},
	"sphere_vs_cube": {
		name:  "sphere_vs_cube",
		makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
		makeB: func() *base.Solid {
			s := primitives.MakeSphere(0.5, 4, 8, lightBlue)
			boolDoTranslate(s, -0.25, -0.25, -0.25)
			return s
		},
	},
	"sphere_cube_same_size": {
		name:  "sphere_cube_same_size",
		makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
		makeB: func() *base.Solid { return primitives.MakeSphere(1.0, 8, 16, lightBlue) },
	},
}

func boolDoTranslate(s *base.Solid, x, y, z float64) {
	s.TransformGeometry(vec.TranslationMatrix(x, y, z))
}

func boolDoRotateCenter(s *base.Solid, degrees float64, axis vec.SFVec3f) {
	mn, mx := s.Extents()
	cx := (mn.X + mx.X) / 2
	cy := (mn.Y + mx.Y) / 2
	cz := (mn.Z + mx.Z) / 2
	boolDoTranslate(s, -cx, -cy, -cz)
	radians := degrees * math.Pi / 180.0
	rot := vec.SFRotation{X: axis.X, Y: axis.Y, Z: axis.Z, W: radians}
	s.TransformGeometry(vec.RotationMatrix(rot))
	boolDoTranslate(s, cx, cy, cz)
}

// ────────────────────────── Settings persistence ──────────────────────────

type viewerSettings struct {
	Wireframe   bool   `json:"wireframe"`
	AxesVisible bool   `json:"axesVisible"`
	LastCase    string `json:"lastCase,omitempty"`
	LastSaveDir string `json:"lastSaveDir,omitempty"`
	WinWidth    int    `json:"winWidth,omitempty"`
	WinHeight   int    `json:"winHeight,omitempty"`
	WinX        int    `json:"winX,omitempty"`
	WinY        int    `json:"winY,omitempty"`
}

func settingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "vrml-viewer", "settings.json")
}

func loadSettings() viewerSettings {
	s := viewerSettings{AxesVisible: true} // default: axes on
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, &s)
	return s
}

func saveSettings(s viewerSettings) {
	p := settingsPath()
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(p, data, 0o644)
}

// ────────────────────────── Viewer state ──────────────────────────

// viewerState holds mutable state toggled by menu actions.
type viewerState struct {
	wireframe    bool
	axesVisible  bool
	normalsShown bool
	axes         *helper.Axes
	scene        *core.Node
	cam          *camera.Camera
	oc           *camera.OrbitControl
	vrmlNodes    []node.Node
	wrlPath      string
	baseDir      string
	normalsNode  core.INode        // face-normals helper, if shown
	reloadFn     func(path string) // callback to reload a file
	lastSaveDir  string            // last directory used for export

	// Browser and picker (shared with main for route/sensor integration)
	browser *browser.Browser
	picker  *traverser.Picker

	// Node map — links VRML scene graph to g3n for rendering sync
	nodeMap *converter.NodeMap

	// Bool demo state
	currentCase  *boolCase         // selected case (nil = none)
	solidA       *base.Solid       // current A solid
	solidB       *base.Solid       // current B solid
	meshANode    *core.Node        // g3n node for A mesh
	meshBNode    *core.Node        // g3n node for B mesh
	xformA       *node.Transform   // VRML Transform for A mesh
	xformB       *node.Transform   // VRML Transform for B mesh
	resultNodes  []*core.Node      // g3n nodes for bool results
	resultXforms []*node.Transform // VRML Transforms for bool results
	resultSolids []*base.Solid     // result solids (for export)
	resultNames  []string          // result operation names
	resultSpan   float64           // offset for result positioning

	// Pick selection (persistent until deselected)
	pickTargets    []pickTarget
	pickRoutes     []*node.Route
	selectedPick   *pickTarget // currently selected solid (nil = none)
	selectionColor vec.SFColor // color applied to selected solid

	// Cursor tracking for red-selection rotation
	lastCursorX float64
	lastCursorY float64

	// Wireframe: saved original materials keyed by mesh pointer
	savedMaterials map[*graphic.Mesh][]material.IMaterial

	// Thread-safe pending load (OpenGL must run on main thread)
	pendingLoad chan string
}

// setupMenuBar creates the menu bar and adds it to the scene.
// Returns the menu bar panel (caller adds to scene for rendering).
func setupMenuBar(vs *viewerState) *gui.Menu {
	mb := gui.NewMenuBar()
	mb.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockTop})

	mb.AddMenu("File", buildFileMenu(vs))
	mb.AddMenu("View", buildViewMenu(vs))
	mb.AddMenu("Bool", buildBoolMenu(vs))
	mb.AddMenu("Primitives", buildPrimitivesMenu(vs))
	mb.AddMenu("Debug", buildDebugMenu(vs))

	return mb
}

func persistSettings(vs *viewerState) {
	s := viewerSettings{
		Wireframe:   vs.wireframe,
		AxesVisible: vs.axesVisible,
		LastSaveDir: vs.lastSaveDir,
	}
	if vs.currentCase != nil {
		s.LastCase = vs.currentCase.name
	}
	// Capture current window geometry
	a := app.App()
	if gw, ok := a.IWindow.(*window.GlfwWindow); ok {
		s.WinWidth, s.WinHeight = gw.GetSize()
		s.WinX, s.WinY = gw.GetPos()
	}
	saveSettings(s)
}

func updateWindowTitle(vs *viewerState) {
	title := "VRML Viewer"
	if vs.currentCase != nil {
		title = "VRML Viewer — " + vs.currentCase.name
	}
	a := app.App()
	if gw, ok := a.IWindow.(*window.GlfwWindow); ok {
		gw.SetTitle(title)
	}
}

// handleCmdShortcut processes Cmd+key shortcuts directly from the window
// key-down handler. Bool cases (Cmd+1-5) and bool operations (Cmd+U/I/D)
// are handled by the menu SetShortcut mechanism — do NOT duplicate them here.
func handleCmdShortcut(vs *viewerState, kev *window.KeyEvent) {
	switch kev.Key {
	case window.KeyW:
		vs.wireframe = !vs.wireframe
		toggleWireframe(vs, vs.wireframe)
		persistSettings(vs)
	case window.KeyA:
		vs.axesVisible = !vs.axesVisible
		vs.axes.SetVisible(vs.axesVisible)
		persistSettings(vs)
	case window.KeyR:
		resetCamera(vs)
	case window.KeyO:
		go openFileDialog(vs)
	case window.KeyS:
		if kev.Mods&window.ModShift != 0 {
			go exportScene(vs)
		}
	case window.KeyE:
		showEulerStats(vs)
	case window.KeyN:
		toggleNormals(vs)
	case window.KeyH:
		showHalfEdgeInfo(vs)
	case window.KeyQ:
		os.Exit(0)
	}
}

// ────────────────────────── File ──────────────────────────

func buildFileMenu(vs *viewerState) *gui.Menu {
	m := gui.NewMenu()

	open := m.AddOption("Open WRL...")
	open.SetShortcut(window.ModSuper, window.KeyO)
	open.Subscribe(gui.OnClick, func(string, interface{}) {
		go openFileDialog(vs)
	})

	m.AddSeparator()

	exportWRL := m.AddOption("Export WRL...")
	exportWRL.SetShortcut(window.ModSuper|window.ModShift, window.KeyS)
	exportWRL.Subscribe(gui.OnClick, func(string, interface{}) {
		go exportScene(vs)
	})

	m.AddSeparator()

	quit := m.AddOption("Quit")
	quit.SetShortcut(window.ModSuper, window.KeyQ)
	quit.Subscribe(gui.OnClick, func(string, interface{}) {
		os.Exit(0)
	})

	return m
}

// openFileDialog uses macOS osascript to pick a .wrl file.
// Sends the path to vs.pendingLoad so the main thread does the actual load.
func openFileDialog(vs *viewerState) {
	out, err := exec.Command("osascript", "-e",
		`POSIX path of (choose file of type {"wrl"} with prompt "Open VRML File")`).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Open cancelled or failed: %v\n", err)
		return
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "Opening: %s\n", path)
	vs.pendingLoad <- path
}

// exportScene uses osascript to pick a save path and serializes the current scene to VRML.
func exportScene(vs *viewerState) {
	defaultName := "scene.wrl"
	if vs.currentCase != nil {
		defaultName = vs.currentCase.name + ".wrl"
	}
	defaultDir := vs.lastSaveDir
	if defaultDir == "" {
		home, _ := os.UserHomeDir()
		defaultDir = filepath.Join(home, "Desktop")
	}

	script := fmt.Sprintf(
		`POSIX path of (choose file name default name "%s" default location POSIX file "%s" with prompt "Export WRL")`,
		defaultName, defaultDir)
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export cancelled or failed: %v\n", err)
		return
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return
	}

	vs.lastSaveDir = filepath.Dir(path)
	persistSettings(vs)

	if vs.currentCase != nil {
		if err := exportBoolScene(vs, path); err != nil {
			fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
			return
		}
	} else if len(vs.vrmlNodes) > 0 {
		if err := exportVRMLScene(vs, path); err != nil {
			fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
			return
		}
	} else {
		fmt.Fprintf(os.Stderr, "Export: nothing to export\n")
		return
	}
	fmt.Fprintf(os.Stderr, "Exported to: %s\n", path)
}

// exportBoolScene writes the bool demo solids (A, B, result) to a VRML file.
func exportBoolScene(vs *viewerState, path string) (retErr error) {
	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer func() {
		if cErr := f.Close(); cErr != nil && retErr == nil {
			retErr = cErr
		}
	}()

	if _, err := fmt.Fprintln(f, "#VRML V2.0 utf8"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(f, "# Bool case: %s\n", vs.currentCase.name); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(f, `NavigationInfo { type "EXAMINE" }`); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(f, `Viewpoint { position 0 0 6 description "Front" }`); err != nil {
		return err
	}

	if vs.solidA != nil {
		if _, err := fmt.Fprintln(f, "\n# Input A"); err != nil {
			return err
		}
		if err := solidExport.Shape(vs.solidA, f, ""); err != nil {
			return err
		}
	}
	if vs.solidB != nil {
		if _, err := fmt.Fprintln(f, "\n# Input B"); err != nil {
			return err
		}
		if err := solidExport.Wireframe(vs.solidB, f, "", lightBlue); err != nil {
			return err
		}
	}
	for i, rs := range vs.resultSolids {
		name := "Result"
		if i < len(vs.resultNames) {
			name = vs.resultNames[i]
		}
		if _, err := fmt.Fprintf(f, "\n# %s\n", name); err != nil {
			return err
		}
		if err := solidExport.Shape(rs, f, ""); err != nil {
			return err
		}
	}
	return nil
}

// exportVRMLScene writes the current VRML node tree to a file.
func exportVRMLScene(vs *viewerState, path string) (retErr error) {
	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer func() {
		if cErr := f.Close(); cErr != nil && retErr == nil {
			retErr = cErr
		}
	}()
	writer.New(f).WriteScene(vs.vrmlNodes)
	return nil
}

// ────────────────────────── View ──────────────────────────

func buildViewMenu(vs *viewerState) *gui.Menu {
	m := gui.NewMenu()

	wire := m.AddOption("Toggle Wireframe")
	wire.SetShortcut(window.ModSuper, window.KeyW)
	wire.Subscribe(gui.OnClick, func(string, interface{}) {
		vs.wireframe = !vs.wireframe
		toggleWireframe(vs, vs.wireframe)
		persistSettings(vs)
		fmt.Fprintf(os.Stderr, "Wireframe: %v\n", vs.wireframe)
	})

	axesItem := m.AddOption("Toggle Axes")
	axesItem.SetShortcut(window.ModSuper, window.KeyA)
	axesItem.Subscribe(gui.OnClick, func(string, interface{}) {
		vs.axesVisible = !vs.axesVisible
		vs.axes.SetVisible(vs.axesVisible)
		persistSettings(vs)
		fmt.Fprintf(os.Stderr, "Axes: %v\n", vs.axesVisible)
	})

	m.AddSeparator()

	resetCam := m.AddOption("Reset Camera")
	resetCam.SetShortcut(window.ModSuper, window.KeyR)
	resetCam.Subscribe(gui.OnClick, func(string, interface{}) {
		resetCamera(vs)
		fmt.Fprintf(os.Stderr, "Camera reset\n")
	})

	return m
}

// toggleWireframe swaps mesh materials between their originals and white
// unlit wireframe. Saved originals are stored on vs.savedMaterials.
func toggleWireframe(vs *viewerState, on bool) {
	if vs.savedMaterials == nil {
		vs.savedMaterials = make(map[*graphic.Mesh][]material.IMaterial)
	}
	if on {
		applyWireframeRecursive(vs, vs.scene)
	} else {
		restoreMaterialsRecursive(vs, vs.scene)
	}
}

func applyWireframeRecursive(vs *viewerState, n core.INode) {
	if gr, ok := n.(*graphic.Mesh); ok {
		// Save originals if not already saved
		if _, saved := vs.savedMaterials[gr]; !saved {
			var orig []material.IMaterial
			for _, gm := range gr.Materials() {
				orig = append(orig, gm.IMaterial())
			}
			vs.savedMaterials[gr] = orig
		}
		// Replace with white unlit wireframe
		gr.ClearMaterials()
		wm := material.NewStandard(&math32.Color{R: 1, G: 1, B: 1})
		wm.SetEmissiveColor(&math32.Color{R: 1, G: 1, B: 1})
		wm.SetWireframe(true)
		wm.SetSide(material.SideDouble)
		gr.AddMaterial(wm, 0, 0)
	}
	for _, child := range n.GetNode().Children() {
		applyWireframeRecursive(vs, child)
	}
}

func restoreMaterialsRecursive(vs *viewerState, n core.INode) {
	if gr, ok := n.(*graphic.Mesh); ok {
		if orig, saved := vs.savedMaterials[gr]; saved {
			gr.ClearMaterials()
			for _, m := range orig {
				gr.AddMaterial(m, 0, 0)
			}
			delete(vs.savedMaterials, gr)
		}
	}
	for _, child := range n.GetNode().Children() {
		restoreMaterialsRecursive(vs, child)
	}
}

// ────────────────────────── Pick highlighting ──────────────────────────

// pickTarget pairs a g3n node with its VRML TouchSensor/Material for pick-selection.
type pickTarget struct {
	gNode     *core.Node
	sensor    *node.TouchSensor
	mat       *node.Material
	vrmlXform *node.Transform    // VRML Transform wrapping this mesh (owns rotation/translation/scale)
	origColor vec.SFColor        // color when unselected
	makeFn    func() *base.Solid // factory to produce a fresh copy of the underlying solid
}

// addPickTarget creates a TouchSensor + Material for a mesh node (no routes needed).
func addPickTarget(vs *viewerState, meshNode *core.Node, xform *node.Transform, color vec.SFColor, makeFn func() *base.Solid) {
	if meshNode == nil {
		return
	}
	ts := node.NewTouchSensor()
	mat := &node.Material{DiffuseColor: color}
	tagMeshChildren(meshNode, ts)

	vs.pickTargets = append(vs.pickTargets, pickTarget{gNode: meshNode, sensor: ts, mat: mat, vrmlXform: xform, origColor: color, makeFn: makeFn})
}

// tagMeshChildren sets UserData on each graphic.Mesh child of parent so the
// Picker's raycaster hit object has the sensor directly attached.
func tagMeshChildren(parent *core.Node, sensor *node.TouchSensor) {
	if parent == nil {
		return
	}
	data := []node.Node{sensor}
	for _, child := range parent.Children() {
		if _, ok := child.(*graphic.Mesh); ok {
			child.GetNode().SetUserData(data)
		}
	}
}

// clearPickSensors removes pick-related sensors from the viewer.
func clearPickSensors(vs *viewerState) {
	if vs.browser != nil && len(vs.pickRoutes) > 0 {
		rm := make(map[*node.Route]bool, len(vs.pickRoutes))
		for _, r := range vs.pickRoutes {
			rm[r] = true
		}
		kept := vs.browser.Routes[:0]
		for _, r := range vs.browser.Routes {
			if !rm[r] {
				kept = append(kept, r)
			}
		}
		vs.browser.Routes = kept
	}
	vs.pickRoutes = nil
	vs.pickTargets = nil
	vs.selectedPick = nil
	vs.selectionColor = vec.SFColor{}
}

// deselectAll clears the current pick selection, restoring the original color.
func deselectAll(vs *viewerState) {
	if vs.selectedPick != nil {
		vs.selectedPick.mat.DiffuseColor = vs.selectedPick.origColor
		vs.selectedPick = nil
		vs.selectionColor = vec.SFColor{}
		vs.oc.SetEnabled(camera.OrbitAll)
	}
}

// deleteSelected removes the currently selected solid from the scene.
func deleteSelected(vs *viewerState) {
	sel := vs.selectedPick
	if sel == nil {
		return
	}

	// Remove from scene
	gn := sel.gNode
	vs.scene.Remove(gn)

	// Unregister its VRML Transform from the nodeMap
	if sel.vrmlXform != nil {
		delete(vs.nodeMap.Transforms, sel.vrmlXform)
	}

	// Clear viewerState references if this is mesh A or B
	if gn == vs.meshANode {
		vs.meshANode = nil
		vs.xformA = nil
		vs.solidA = nil
	}
	if gn == vs.meshBNode {
		vs.meshBNode = nil
		vs.xformB = nil
		vs.solidB = nil
	}

	// Remove from result lists
	for i, rn := range vs.resultNodes {
		if rn == gn {
			vs.resultNodes = append(vs.resultNodes[:i], vs.resultNodes[i+1:]...)
			vs.resultXforms = append(vs.resultXforms[:i], vs.resultXforms[i+1:]...)
			vs.resultSolids = append(vs.resultSolids[:i], vs.resultSolids[i+1:]...)
			vs.resultNames = append(vs.resultNames[:i], vs.resultNames[i+1:]...)
			break
		}
	}

	// Remove from pickTargets
	for i, pt := range vs.pickTargets {
		if pt.gNode == gn {
			vs.pickTargets = append(vs.pickTargets[:i], vs.pickTargets[i+1:]...)
			break
		}
	}

	// Clear selection state
	vs.selectedPick = nil
	vs.selectionColor = vec.SFColor{}
	vs.oc.SetEnabled(camera.OrbitAll)
}

// selectPickTarget toggles selection on the pick target matching the given
// sensor. If the target is already selected with the same color, it is
// deselected (restored to its original color). Otherwise any previous
// selection is deselected first, then the new target is selected.
func selectPickTarget(vs *viewerState, ts *node.TouchSensor, color vec.SFColor) {
	// Find which target was clicked.
	var clicked *pickTarget
	for i := range vs.pickTargets {
		if vs.pickTargets[i].sensor == ts {
			clicked = &vs.pickTargets[i]
			break
		}
	}
	if clicked == nil {
		return
	}

	// If the same target is already selected (any color), deselect.
	if vs.selectedPick == clicked {
		deselectAll(vs)
		return
	}

	// Deselect any previous selection.
	deselectAll(vs)

	// Select the new target.
	clicked.mat.DiffuseColor = color
	vs.selectedPick = clicked
	vs.selectionColor = color

	// Disable orbit rotation so mouse drag manipulates the solid, not the scene.
	vs.oc.SetEnabled(camera.OrbitZoom | camera.OrbitKeys)
}

// rotateSelectedSolid rotates the red-selected solid around its center
// based on mouse movement deltas (in screen pixels).
func rotateSelectedSolid(vs *viewerState, dx, dy float64) {
	if vs.selectedPick == nil || vs.selectedPick.vrmlXform == nil {
		return
	}
	xf := vs.selectedPick.vrmlXform
	const sensitivity = 0.005 // radians per pixel
	// Horizontal mouse → rotate around Y axis; vertical → rotate around X axis.
	cur := xf.Rotation
	if dx != 0 {
		deltaY := vec.NewRotation(0, 1, 0, dx*sensitivity)
		cur = vec.ComposeRotations(cur, deltaY)
	}
	if dy != 0 {
		deltaX := vec.NewRotation(1, 0, 0, dy*sensitivity)
		cur = vec.ComposeRotations(cur, deltaX)
	}
	xf.Rotation = cur
	// UpdateDynamic() in the render loop will sync this to g3n.
}

// translateSelectedSolid moves the purple-selected solid in the XY plane
// based on mouse movement deltas (in screen pixels).
func translateSelectedSolid(vs *viewerState, dx, dy float64) {
	if vs.selectedPick == nil || vs.selectedPick.vrmlXform == nil {
		return
	}
	xf := vs.selectedPick.vrmlXform
	const sensitivity = 0.01 // world units per pixel
	xf.Translation.X += dx * sensitivity
	xf.Translation.Y -= dy * sensitivity // screen Y is inverted
}

// scaleSelectedSolid scales the pink-selected solid uniformly based on
// vertical mouse movement (drag up = bigger, drag down = smaller).
func scaleSelectedSolid(vs *viewerState, dx, dy float64) {
	if vs.selectedPick == nil || vs.selectedPick.vrmlXform == nil {
		return
	}
	xf := vs.selectedPick.vrmlXform
	const sensitivity = 0.005 // scale factor per pixel
	// Vertical drag: up (negative dy) grows, down shrinks. Also use horizontal.
	factor := 1.0 - dy*sensitivity + dx*sensitivity
	if factor < 0.1 {
		factor = 0.1
	}
	xf.Scale.X *= factor
	xf.Scale.Y *= factor
	xf.Scale.Z *= factor
}

// syncPickColors propagates VRML Material.DiffuseColor (set by routes) to
// the g3n mesh materials so the highlight is visible in the renderer.
func syncPickColors(vs *viewerState) {
	for _, pt := range vs.pickTargets {
		syncPickMesh(pt.mat, pt.gNode, vs.wireframe)
	}
}

func syncPickMesh(vrmlMat *node.Material, meshNode *core.Node, wireframe bool) {
	if vrmlMat == nil || meshNode == nil {
		return
	}
	c := &math32.Color{
		R: float32(vrmlMat.DiffuseColor.R),
		G: float32(vrmlMat.DiffuseColor.G),
		B: float32(vrmlMat.DiffuseColor.B),
	}
	for _, child := range meshNode.Children() {
		if mesh, ok := child.(*graphic.Mesh); ok {
			for _, gm := range mesh.Materials() {
				if stdMat, ok := gm.IMaterial().(*material.Standard); ok {
					stdMat.SetColor(c)
					if wireframe {
						stdMat.SetEmissiveColor(c)
					}
				}
			}
		}
	}
}

// resetCamera moves camera to default position.
func resetCamera(vs *viewerState) {
	vp := converter.GetViewpoint(vs.vrmlNodes)
	if vp != nil {
		vs.cam.SetPosition(float32(vp.Position.X), float32(vp.Position.Y), float32(vp.Position.Z))
	} else {
		vs.cam.SetPosition(0, 2, 10)
	}
	vs.cam.LookAt(&math32.Vector3{X: 0, Y: 0, Z: 0}, &math32.Vector3{X: 0, Y: 1, Z: 0})
	vs.oc.Reset()
}

// ────────────────────────── Bool ──────────────────────────

func buildBoolMenu(vs *viewerState) *gui.Menu {
	m := gui.NewMenu()

	// Add the test cases with Cmd+1, Cmd+2, etc.
	caseOrder := []string{"same_size_partial_overlap", "same_size_edge_on_edge", "same_size_slight_twist", "sphere_vs_cube", "sphere_cube_same_size"}
	caseKeys := []window.Key{window.Key1, window.Key2, window.Key3, window.Key4, window.Key5}
	for i, name := range caseOrder {
		bc := boolCaseMap[name]
		capturedCase := bc
		opt := m.AddOption(name)
		opt.SetShortcut(window.ModSuper, caseKeys[i])
		opt.Subscribe(gui.OnClick, func(string, interface{}) {
			selectBoolCase(vs, capturedCase)
		})
	}

	m.AddSeparator()

	union := m.AddOption("Union (A ∪ B)")
	union.SetShortcut(window.ModSuper, window.KeyU)
	union.Subscribe(gui.OnClick, func(string, interface{}) {
		runBoolOp(vs, base.BoolUnion, "Union")
	})

	inter := m.AddOption("Intersection (A ∩ B)")
	inter.SetShortcut(window.ModSuper, window.KeyI)
	inter.Subscribe(gui.OnClick, func(string, interface{}) {
		runBoolOp(vs, base.BoolIntersection, "Intersection")
	})

	diff := m.AddOption("Difference (A − B)")
	diff.SetShortcut(window.ModSuper, window.KeyD)
	diff.Subscribe(gui.OnClick, func(string, interface{}) {
		runBoolOp(vs, base.BoolDifference, "Difference")
	})

	return m
}

// buildPrimitivesMenu creates a menu to display individual primitives.
func buildPrimitivesMenu(vs *viewerState) *gui.Menu {
	m := gui.NewMenu()

	type primDef struct {
		name string
		make func() *base.Solid
	}
	prims := []primDef{
		{"Cube", func() *base.Solid { return primitives.MakeCube(1.0, yellow) }},
		{"Sphere", func() *base.Solid { return primitives.MakeSphere(1.0, 8, 16, yellow) }},
		{"Cylinder", func() *base.Solid { return primitives.MakeCylinder(0.5, 2.0, 16, yellow) }},
		{"Prism", func() *base.Solid { return primitives.MakePrism(2.0, yellow) }},
		{"Torus", func() *base.Solid { return primitives.MakeTorus(1.0, 0.3, 16, 8, yellow) }},
		{"Circle", func() *base.Solid { return primitives.MakeCircle(0, 0, 1.0, 0.2, 16, yellow) }},
	}

	for _, p := range prims {
		pd := p
		opt := m.AddOption(pd.name)
		opt.Subscribe(gui.OnClick, func(string, interface{}) {
			showPrimitive(vs, pd.name, pd.make)
		})
	}

	return m
}

// showPrimitive adds a single primitive solid to the scene.
var lastPrimitiveTime time.Time

func showPrimitive(vs *viewerState, name string, makeFn func() *base.Solid) {
	// Debounce: g3n SetShortcut fires the callback twice per keystroke.
	if time.Since(lastPrimitiveTime) < 200*time.Millisecond {
		return
	}
	lastPrimitiveTime = time.Now()

	s := makeFn()
	if s == nil {
		return
	}
	s.CalcPlaneEquations()

	mesh := solidToMesh(s, &math32.Color{R: 1, G: 0.9, B: 0.2})
	if mesh == nil {
		return
	}

	xf := node.NewTransform()
	gn := core.NewNode()
	gn.Add(mesh)
	vs.scene.Add(gn)
	vs.nodeMap.Transforms[xf] = gn
	addPickTarget(vs, gn, xf, yellow, makeFn)

	if vs.wireframe {
		applyWireframeRecursive(vs, gn)
	}
}

// selectBoolCase builds A and B solids for the chosen case and adds them to the scene.
var lastBoolCaseTime time.Time

func selectBoolCase(vs *viewerState, bc *boolCase) {
	// Debounce: g3n SetShortcut fires the callback twice per keystroke.
	if time.Since(lastBoolCaseTime) < 200*time.Millisecond {
		return
	}
	lastBoolCaseTime = time.Now()

	vs.currentCase = bc

	// Build fresh solids
	vs.solidA = bc.makeA()
	vs.solidB = bc.makeB()
	vs.solidA.CalcPlaneEquations()
	vs.solidB.CalcPlaneEquations()

	// Compute result offset from extents
	mnA, mxA := vs.solidA.Extents()
	mnB, mxB := vs.solidB.Extents()
	xmin := mnA.X
	if mnB.X < xmin {
		xmin = mnB.X
	}
	xmax := mxA.X
	if mxB.X > xmax {
		xmax = mxB.X
	}
	span := xmax - xmin
	if span < 1 {
		span = 1
	}
	vs.resultSpan = span*0.5 + span*0.8

	// Add meshes for A (yellow) and B (blue), each wrapped in a VRML Transform
	meshA := solidToMesh(vs.solidA, &math32.Color{R: 1, G: 0.9, B: 0.2})
	meshB := solidToMesh(vs.solidB, &math32.Color{R: 0.4, G: 0.6, B: 1.0})

	if meshA != nil {
		vs.xformA = node.NewTransform()
		vs.meshANode = core.NewNode()
		vs.meshANode.Add(meshA)
		vs.scene.Add(vs.meshANode)
		vs.nodeMap.Transforms[vs.xformA] = vs.meshANode
		addPickTarget(vs, vs.meshANode, vs.xformA, yellow, bc.makeA)
	}
	if meshB != nil {
		vs.xformB = node.NewTransform()
		vs.meshBNode = core.NewNode()
		vs.meshBNode.Add(meshB)
		vs.scene.Add(vs.meshBNode)
		vs.nodeMap.Transforms[vs.xformB] = vs.meshBNode
		addPickTarget(vs, vs.meshBNode, vs.xformB, lightBlue, bc.makeB)
	}

	statsA := solidStatsStr(vs.solidA)
	statsB := solidStatsStr(vs.solidB)
	fmt.Fprintf(os.Stderr, "Bool case '%s': A(%s), B(%s)\n", bc.name, statsA, statsB)

	updateWindowTitle(vs)
	persistSettings(vs)
	if vs.wireframe {
		toggleWireframe(vs, true)
	}
}

// runBoolOp runs a boolean operation on all non-result solids in the scene.
// Chains pairwise: solids[0] op solids[1] → result, result op solids[2] → …
var lastBoolOpTime time.Time

func runBoolOp(vs *viewerState, op int, name string) {
	// Debounce: g3n SetShortcut fires the callback twice per keystroke.
	if time.Since(lastBoolOpTime) < 200*time.Millisecond {
		return
	}
	lastBoolOpTime = time.Now()

	// Build set of result nodes for fast lookup.
	isResult := make(map[*core.Node]bool, len(vs.resultNodes))
	for _, rn := range vs.resultNodes {
		isResult[rn] = true
	}

	// Collect fresh solids from all non-result pick targets, applying transforms.
	var solids []*base.Solid
	for i := range vs.pickTargets {
		pt := &vs.pickTargets[i]
		if isResult[pt.gNode] || pt.makeFn == nil {
			continue
		}
		s := pt.makeFn()
		if s == nil {
			continue
		}
		s.CalcPlaneEquations()
		if pt.vrmlXform != nil {
			m := pt.vrmlXform.GetLocalMatrix()
			if m != vec.Identity() {
				s.TransformGeometry(m)
				s.CalcPlaneEquations()
			}
		}
		solids = append(solids, s)
	}

	if len(solids) < 2 {
		fmt.Fprintf(os.Stderr, "Bool %s: need at least 2 non-result solids in scene (have %d)\n", name, len(solids))
		return
	}

	fmt.Fprintf(os.Stderr, "Bool %s: chaining %d solids\n", name, len(solids))
	boolRes := boolop.ChainBoolOpEx(solids, op)
	if !boolRes.Ok || boolRes.Solid == nil {
		fmt.Fprintf(os.Stderr, "Bool %s: operation failed or empty result\n", name)
		return
	}

	result := boolRes.Solid
	result.CalcPlaneEquations()

	stats := solidStatsStr(result)
	if boolRes.UsedLastDitch {
		fmt.Fprintf(os.Stderr, "Bool %s: %s (LastDitch)\n", name, stats)
	} else {
		fmt.Fprintf(os.Stderr, "Bool %s: %s\n", name, stats)
	}

	errs := algorithms.VerifyDetailed(result)
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Bool %s: %d errors\n", name, len(errs))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  %v\n", e)
		}
	}

	// Build g3n mesh from result solid and add to scene — teal if LastDitch
	var meshColor *math32.Color
	var pickColor vec.SFColor
	if boolRes.UsedLastDitch {
		meshColor = &math32.Color{R: 0, G: 0.8, B: 0.8}
		pickColor = vec.SFColor{R: 0, G: 0.8, B: 0.8, A: 1}
	} else {
		meshColor = &math32.Color{R: 0.2, G: 0.8, B: 0.3}
		pickColor = vec.SFColor{R: 0.2, G: 0.8, B: 0.3, A: 1}
	}
	mesh := solidToMesh(result, meshColor)
	if mesh == nil {
		fmt.Fprintf(os.Stderr, "Bool %s: could not build mesh from result\n", name)
		return
	}

	// Position result via a VRML Transform (state lives in VRML, not g3n)
	resultXform := node.NewTransform()
	span := vs.resultSpan
	if span == 0 {
		span = 2.0
	}
	switch op {
	case base.BoolUnion:
		resultXform.Translation = vec.SFVec3f{X: span, Y: 0, Z: 0}
	case base.BoolIntersection:
		resultXform.Translation = vec.SFVec3f{X: 0, Y: span, Z: 0}
	case base.BoolDifference:
		resultXform.Translation = vec.SFVec3f{X: -span, Y: 0, Z: 0}
	}

	resultGroup := core.NewNode()
	resultGroup.Add(mesh)
	vs.scene.Add(resultGroup)
	vs.nodeMap.Transforms[resultXform] = resultGroup

	vs.resultNodes = append(vs.resultNodes, resultGroup)
	vs.resultXforms = append(vs.resultXforms, resultXform)
	vs.resultSolids = append(vs.resultSolids, result)
	vs.resultNames = append(vs.resultNames, name)

	if vs.wireframe {
		applyWireframeRecursive(vs, resultGroup)
	}

	// Make result mesh pickable
	addPickTarget(vs, resultGroup, resultXform, pickColor, nil)

	fmt.Fprintf(os.Stderr, "Bool %s: result at (%.1f, %.1f, %.1f)\n",
		name, resultXform.Translation.X, resultXform.Translation.Y, resultXform.Translation.Z)
}

// solidToMesh converts a B-rep Solid into a g3n Mesh for rendering.
func solidToMesh(s *base.Solid, color *math32.Color) *graphic.Mesh {
	s.Renumber()

	positions := math32.NewArrayF32(0, 0)
	normals := math32.NewArrayF32(0, 0)
	indices := math32.NewArrayU32(0, 0)

	vi := uint32(0)
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut == nil {
			continue
		}
		// Collect face vertices
		var faceVerts []vec.SFVec3f
		f.LoopOut.ForEachHe(func(he *base.HalfEdge) bool {
			faceVerts = append(faceVerts, he.Vertex.Loc)
			return true
		})
		if len(faceVerts) < 3 {
			continue
		}

		// Face normal for flat shading
		nx := float32(f.Normal.X)
		ny := float32(f.Normal.Y)
		nz := float32(f.Normal.Z)

		// Add vertices with face normal
		startVI := vi
		for _, v := range faceVerts {
			positions.Append(float32(v.X), float32(v.Y), float32(v.Z))
			normals.Append(nx, ny, nz)
			vi++
		}

		// Fan triangulation
		for i := uint32(1); i+1 < uint32(len(faceVerts)); i++ {
			indices.Append(startVI, startVI+i, startVI+i+1)
		}
	}

	if vi == 0 {
		return nil
	}

	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.SetIndices(indices)

	if color == nil {
		color = &math32.Color{R: 0.8, G: 0.8, B: 0.8}
	}
	mat := material.NewStandard(color)
	mat.SetSide(material.SideDouble)
	mat.SetShininess(30)
	return graphic.NewMesh(geom, mat)
}

// solidStatsStr returns a compact stats string for a solid.
func solidStatsStr(s *base.Solid) string {
	nF, nV, nE, nH := 0, 0, 0, 0
	for f := s.Faces; f != nil; f = f.Next {
		nF++
		nH += f.NLoops() - 1
	}
	for v := s.Verts; v != nil; v = v.Next {
		nV++
	}
	for e := s.Edges; e != nil; e = e.Next {
		nE++
	}
	diff := (nF + nV - 2) - (nE + nH)
	return fmt.Sprintf("F=%d V=%d E=%d H=%d diff=%+d", nF, nV, nE, nH, diff)
}

// clearGeometry removes all geometry (meshes, bool results, converted VRML content)
// from the scene, keeping camera, lights, axes, and the menu bar.
func clearGeometry(vs *viewerState) {
	var toRemove []core.INode
	for _, child := range vs.scene.Children() {
		// Keep camera, lights, axes helper, and GUI elements
		if child == vs.cam {
			continue
		}
		if child == vs.axes {
			continue
		}
		switch child.(type) {
		case *gui.Menu:
			continue
		case *light.Ambient, *light.Directional, *light.Point, *light.Spot:
			continue
		}
		toRemove = append(toRemove, child)
	}
	for _, child := range toRemove {
		vs.scene.Remove(child)
	}

	// Clear state
	vs.meshANode = nil
	vs.meshBNode = nil
	if vs.xformA != nil {
		delete(vs.nodeMap.Transforms, vs.xformA)
		vs.xformA = nil
	}
	if vs.xformB != nil {
		delete(vs.nodeMap.Transforms, vs.xformB)
		vs.xformB = nil
	}
	for _, rx := range vs.resultXforms {
		delete(vs.nodeMap.Transforms, rx)
	}
	vs.resultNodes = nil
	vs.resultXforms = nil
	vs.solidA = nil
	vs.solidB = nil
	vs.normalsNode = nil
	vs.normalsShown = false
	clearPickSensors(vs)
}

// ────────────────────────── Debug ──────────────────────────

func buildDebugMenu(vs *viewerState) *gui.Menu {
	m := gui.NewMenu()

	euler := m.AddOption("Show Euler Stats")
	euler.SetShortcut(window.ModSuper, window.KeyE)
	euler.Subscribe(gui.OnClick, func(string, interface{}) {
		showEulerStats(vs)
	})

	normals := m.AddOption("Toggle Face Normals")
	normals.SetShortcut(window.ModSuper, window.KeyN)
	normals.Subscribe(gui.OnClick, func(string, interface{}) {
		toggleNormals(vs)
	})

	halfedge := m.AddOption("Show Half-Edge Info")
	halfedge.SetShortcut(window.ModSuper, window.KeyH)
	halfedge.Subscribe(gui.OnClick, func(string, interface{}) {
		showHalfEdgeInfo(vs)
	})

	return m
}

// showEulerStats prints Euler stats for the cached A and B solids.
func showEulerStats(vs *viewerState) {
	if vs.solidA == nil && vs.solidB == nil {
		fmt.Fprintf(os.Stderr, "Euler: no solids loaded (select a case from Bool menu)\n")
		return
	}
	for i, s := range []*base.Solid{vs.solidA, vs.solidB} {
		label := "A"
		if i == 1 {
			label = "B"
		}
		if s == nil {
			continue
		}
		stats := solidStatsStr(s)
		fmt.Fprintf(os.Stderr, "Solid %s: %s\n", label, stats)
		errs := algorithms.VerifyDetailed(s)
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  %v\n", e)
		}
		if len(errs) == 0 {
			fmt.Fprintf(os.Stderr, "  valid\n")
		}
	}
}

// toggleNormals adds/removes a face normals helper from the scene.
func toggleNormals(vs *viewerState) {
	if vs.normalsShown && vs.normalsNode != nil {
		vs.scene.Remove(vs.normalsNode)
		vs.normalsNode = nil
		vs.normalsShown = false
		fmt.Fprintf(os.Stderr, "Face normals: hidden\n")
		return
	}

	// Create normals helper for every mesh in the scene
	normalsGroup := core.NewNode()
	addNormalsHelpers(vs.scene, normalsGroup)
	vs.scene.Add(normalsGroup)
	vs.normalsNode = normalsGroup
	vs.normalsShown = true
	fmt.Fprintf(os.Stderr, "Face normals: shown\n")
}

// addNormalsHelpers recursively finds meshes and adds normal helpers.
func addNormalsHelpers(n core.INode, group *core.Node) {
	if mesh, ok := n.(*graphic.Mesh); ok {
		nh := helper.NewNormals(mesh, 0.3, &math32.Color{R: 1, G: 1, B: 0}, 1)
		group.Add(nh)
	}
	for _, child := range n.GetNode().Children() {
		addNormalsHelpers(child, group)
	}
}

// showHalfEdgeInfo prints half-edge details for the cached solids.
func showHalfEdgeInfo(vs *viewerState) {
	if vs.solidA == nil && vs.solidB == nil {
		fmt.Fprintf(os.Stderr, "HalfEdge: no solids loaded (select a case from Bool menu)\n")
		return
	}
	for i, sol := range []*base.Solid{vs.solidA, vs.solidB} {
		label := "A"
		if i == 1 {
			label = "B"
		}
		if sol == nil {
			continue
		}
		nF, nHE := 0, 0
		for f := sol.Faces; f != nil; f = f.Next {
			nF++
			for _, l := range f.Loops {
				he := l.HalfEdges
				if he == nil {
					continue
				}
				first := he
				for {
					nHE++
					he = he.Next
					if he == first {
						break
					}
				}
			}
		}
		nV, nE := 0, 0
		for v := sol.Verts; v != nil; v = v.Next {
			nV++
		}
		for e := sol.Edges; e != nil; e = e.Next {
			nE++
		}
		fmt.Fprintf(os.Stderr, "Solid %s: F=%d V=%d E=%d HalfEdges=%d (expect %d)\n",
			label, nF, nV, nE, nHE, nE*2)
	}
}

// loadScene parses a WRL file and replaces the current scene content.
// If the file has 3+ top-level Transforms (bool_demo format), the 3rd+ are
// dropped (they're pre-computed results) — only A and B are kept.
func loadScene(vs *viewerState, path string) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", path, err)
		return
	}

	baseDir := filepath.Dir(path)
	p := parser.NewParser(f)
	p.SetBaseDir(baseDir)
	vrmlNodes := p.Parse()
	if err := f.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "close error: %v\n", err)
	}

	if len(vrmlNodes) == 0 {
		fmt.Fprintf(os.Stderr, "No nodes parsed from %s\n", path)
		return
	}

	// Drop the 3rd+ Transform nodes (pre-computed results in bool_demo files)
	vrmlNodes = dropResultTransforms(vrmlNodes)

	// Remove all scene children except camera, lights, axes, and the menu bar
	var keep []core.INode
	for _, child := range vs.scene.Children() {
		switch child.(type) {
		case *gui.Menu:
			keep = append(keep, child)
		default:
			if child == vs.cam {
				keep = append(keep, child)
			}
		}
	}
	vs.scene.DisposeChildren(true)
	for _, k := range keep {
		vs.scene.Add(k)
	}

	// Re-add axes
	vs.axes = helper.NewAxes(1.0)
	vs.axes.SetVisible(vs.axesVisible)
	vs.scene.Add(vs.axes)

	// Convert and add new scene content
	vs.nodeMap = converter.Convert(vrmlNodes, vs.scene, baseDir)

	vs.vrmlNodes = vrmlNodes
	vs.wrlPath = path
	vs.baseDir = baseDir
	vs.normalsNode = nil
	vs.normalsShown = false
	vs.resultNodes = nil

	fmt.Fprintf(os.Stderr, "Loaded %d nodes from %s\n", len(vrmlNodes), path)
}

// dropResultTransforms keeps only the first 2 Transform nodes and all
// non-Transform nodes (Viewpoint, NavigationInfo, etc.). This strips
// pre-computed bool results from bool_demo files.
func dropResultTransforms(nodes []node.Node) []node.Node {
	var result []node.Node
	transformCount := 0
	for _, n := range nodes {
		if _, ok := n.(*node.Transform); ok {
			transformCount++
			if transformCount > 2 {
				continue // drop 3rd+ transforms
			}
		}
		result = append(result, n)
	}
	if transformCount > 2 {
		fmt.Fprintf(os.Stderr, "Dropped %d result transform(s), keeping A and B\n", transformCount-2)
	}
	return result
}
