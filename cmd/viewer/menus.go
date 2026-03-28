package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/converter"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// findBoolDemoDir locates the examples/bool_demos directory by trying
// several candidate paths (CWD-relative, parent dir, relative to loaded file).
func findBoolDemoDir(vs *viewerState) string {
	candidates := []string{
		"examples/bool_demos",
		"trueblocks-vranimal/examples/bool_demos",
	}
	// Also try relative to the loaded WRL file's directory
	if vs.baseDir != "" {
		candidates = append(candidates, filepath.Join(vs.baseDir, "..", "..", "examples", "bool_demos"))
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}

// ────────────────────────── Bool case definitions ──────────────────────────

var (
	yellow    = vec.SFColor{R: 1.0, G: 0.9, B: 0.2, A: 1}
	lightBlue = vec.SFColor{R: 0.4, G: 0.6, B: 1.0, A: 1}
)

type boolCase struct {
	name  string
	makeA func() *solid.Solid
	makeB func() *solid.Solid
}

var boolCaseMap = map[string]*boolCase{
	"partial_penetration": {
		name:  "partial_penetration",
		makeA: func() *solid.Solid { return solid.MakeCube(1.0, yellow) },
		makeB: func() *solid.Solid {
			s := solid.MakeCube(0.5, lightBlue)
			boolDoScale(s, 1, 1, 2)
			boolDoTranslate(s, 0.25, 0.25, -0.25)
			return s
		},
	},
	"fully_contained": {
		name:  "fully_contained",
		makeA: func() *solid.Solid { return solid.MakeCube(1.0, yellow) },
		makeB: func() *solid.Solid {
			s := solid.MakeCube(0.5, lightBlue)
			boolDoTranslate(s, 0.25, 0.25, 0.25)
			return s
		},
	},
	"through": {
		name:  "through",
		makeA: func() *solid.Solid { return solid.MakeCube(1.0, yellow) },
		makeB: func() *solid.Solid {
			s := solid.MakeCube(0.5, lightBlue)
			boolDoScale(s, 1, 1, 4)
			boolDoTranslate(s, 0.25, 0.25, -0.15)
			return s
		},
	},
	"edge_on_edge": {
		name:  "edge_on_edge",
		makeA: func() *solid.Solid { return solid.MakeCube(1.0, yellow) },
		makeB: func() *solid.Solid {
			s := solid.MakeCube(1.0, lightBlue)
			boolDoTranslate(s, 0, -1, -1)
			return s
		},
	},
	"rotated_elongated": {
		name:  "rotated_elongated",
		makeA: func() *solid.Solid { return solid.MakeCube(1.0, yellow) },
		makeB: func() *solid.Solid {
			s := solid.MakeCube(0.5, lightBlue)
			boolDoScale(s, 1, 1, 4)
			boolDoTranslate(s, -0.25, -0.25, -0.5)
			boolDoTranslate(s, 0.5, 0, -0.25)
			boolDoRotateCenter(s, 55.0, vec.XAxis)
			boolDoTranslate(s, 0, -0.07, 0)
			return s
		},
	},
	"hexagon_prism": {
		name: "hexagon_prism",
		makeA: func() *solid.Solid {
			return boolMakeSweptLamina(boolMakeHexVerts(0.8), vec.SFVec3f{X: 0, Y: 0, Z: -1.5}, yellow)
		},
		makeB: func() *solid.Solid {
			s := boolMakeSweptLamina(boolMakeHexVerts(0.8), vec.SFVec3f{X: 0, Y: 0, Z: -1.5}, lightBlue)
			boolDoTranslate(s, 0.5, 0.3, -0.4)
			return s
		},
	},
	"sphere_vs_cube": {
		name:  "sphere_vs_cube",
		makeA: func() *solid.Solid { return solid.MakeSphere(1.0, 10, 10, yellow) },
		makeB: func() *solid.Solid {
			s := solid.MakeCube(1.0, lightBlue)
			boolDoTranslate(s, 0, 0, -1)
			return s
		},
	},
}

func boolDoTranslate(s *solid.Solid, x, y, z float64) {
	s.TransformGeometry(vec.TranslationMatrix(x, y, z))
}

func boolDoScale(s *solid.Solid, sx, sy, sz float64) {
	s.TransformGeometry(vec.ScaleMatrix(sx, sy, sz))
}

func boolDoRotateCenter(s *solid.Solid, degrees float64, axis vec.SFVec3f) {
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

func boolMakeHexVerts(radius float64) []vec.SFVec3f {
	verts := make([]vec.SFVec3f, 6)
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3.0
		verts[i] = vec.SFVec3f{
			X: radius * math.Cos(angle),
			Y: radius * math.Sin(angle),
			Z: 0,
		}
	}
	return verts
}

func boolMakeSweptLamina(verts []vec.SFVec3f, dir vec.SFVec3f, color vec.SFColor) *solid.Solid {
	s := solid.MakeLamina(verts, color)
	s.TranslationalSweep(s.GetFirstFace(), dir)
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

// parseBoolFilename extracts case name and operation from a pass/fail filename.
// e.g. "pass_partial_penetration_union" → ("partial_penetration", solid.BoolUnion)
func parseBoolFilename(stem string) (caseName string, op int, ok bool) {
	// Strip pass_ or fail_ prefix
	s := stem
	if strings.HasPrefix(s, "pass_") {
		s = s[5:]
	} else if strings.HasPrefix(s, "fail_") {
		s = s[5:]
	} else {
		return "", 0, false
	}

	// The operation is the last _word
	if strings.HasSuffix(s, "_union") {
		return s[:len(s)-6], solid.BoolUnion, true
	}
	if strings.HasSuffix(s, "_intersection") {
		return s[:len(s)-13], solid.BoolIntersection, true
	}
	if strings.HasSuffix(s, "_difference") {
		return s[:len(s)-11], solid.BoolDifference, true
	}
	return "", 0, false
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

	// Bool demo state
	currentCase *boolCase    // selected case (nil = none)
	solidA      *solid.Solid // current A solid
	solidB      *solid.Solid // current B solid
	meshANode   *core.Node   // g3n node for A mesh
	meshBNode   *core.Node   // g3n node for B mesh
	resultNodes []*core.Node // g3n nodes for bool results
	resultSpan  float64      // offset for result positioning

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
	mb.AddMenu("Debug", buildDebugMenu(vs))

	return mb
}

// ────────────────────────── File ──────────────────────────

func buildFileMenu(vs *viewerState) *gui.Menu {
	m := gui.NewMenu()

	open := m.AddOption("Open WRL...")
	open.SetShortcut(window.ModControl, window.KeyO)
	open.Subscribe(gui.OnClick, func(string, interface{}) {
		go openFileDialog(vs)
	})

	m.AddSeparator()

	exportWRL := m.AddOption("Export WRL...")
	exportWRL.SetShortcut(window.ModControl|window.ModShift, window.KeyS)
	exportWRL.Subscribe(gui.OnClick, func(string, interface{}) {
		go exportScene(vs)
	})

	m.AddSeparator()

	quit := m.AddOption("Quit")
	quit.SetShortcut(window.ModControl, window.KeyQ)
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

// exportScene uses osascript to pick a save path and writes the current file.
func exportScene(vs *viewerState) {
	out, err := exec.Command("osascript", "-e",
		`POSIX path of (choose file name default name "scene.wrl" with prompt "Export WRL")`).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export cancelled or failed: %v\n", err)
		return
	}
	path := strings.TrimSpace(string(out))
	if path == "" {
		return
	}
	// Copy current file to chosen path
	src := vs.wrlPath
	data, err := os.ReadFile(filepath.Clean(src))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Export read error: %v\n", err)
		return
	}
	if err := os.WriteFile(filepath.Clean(path), data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Export write error: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "Exported to: %s\n", path)
}

// ────────────────────────── View ──────────────────────────

func buildViewMenu(vs *viewerState) *gui.Menu {
	m := gui.NewMenu()

	wire := m.AddOption("Toggle Wireframe")
	wire.SetShortcut(window.ModControl, window.KeyW)
	wire.Subscribe(gui.OnClick, func(string, interface{}) {
		vs.wireframe = !vs.wireframe
		toggleWireframe(vs.scene, vs.wireframe)
		fmt.Fprintf(os.Stderr, "Wireframe: %v\n", vs.wireframe)
	})

	axesItem := m.AddOption("Toggle Axes")
	axesItem.SetShortcut(window.ModControl, window.KeyA)
	axesItem.Subscribe(gui.OnClick, func(string, interface{}) {
		vs.axesVisible = !vs.axesVisible
		vs.axes.SetVisible(vs.axesVisible)
		fmt.Fprintf(os.Stderr, "Axes: %v\n", vs.axesVisible)
	})

	m.AddSeparator()

	resetCam := m.AddOption("Reset Camera")
	resetCam.SetShortcut(window.ModControl, window.KeyR)
	resetCam.Subscribe(gui.OnClick, func(string, interface{}) {
		resetCamera(vs)
		fmt.Fprintf(os.Stderr, "Camera reset\n")
	})

	return m
}

// toggleWireframe recursively sets wireframe on all materials in the scene.
func toggleWireframe(n core.INode, on bool) {
	if gr, ok := n.(*graphic.Mesh); ok {
		for _, gm := range gr.Materials() {
			if mat, ok := gm.IMaterial().(*material.Standard); ok {
				mat.SetWireframe(on)
			}
		}
	}
	node := n.GetNode()
	for _, child := range node.Children() {
		toggleWireframe(child, on)
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

	demoDir := findBoolDemoDir(vs)
	if demoDir == "" {
		fmt.Fprintf(os.Stderr, "Bool menu: cannot find examples/bool_demos directory\n")
		m.AddOption("(no bool demos found)")
		return m
	}

	// Scan bool_demos directory for pass/fail WRL files
	entries, err := os.ReadDir(demoDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bool menu: cannot read %s: %v\n", demoDir, err)
		return m
	}

	var passFiles, failFiles []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".wrl") {
			continue
		}
		if strings.HasPrefix(name, "pass_") {
			passFiles = append(passFiles, name)
		} else if strings.HasPrefix(name, "fail_") {
			failFiles = append(failFiles, name)
		}
	}

	fmt.Fprintf(os.Stderr, "Bool menu: %d pass + %d fail files from %s\n", len(passFiles), len(failFiles), demoDir)

	// Each pass/fail file becomes a menu item that selects A+B and runs the operation
	addBoolFileItems := func(files []string) {
		for _, name := range files {
			stem := strings.TrimSuffix(name, ".wrl")
			caseName, _, ok := parseBoolFilename(stem)
			if !ok {
				continue
			}
			bc := boolCaseMap[caseName]
			if bc == nil {
				fmt.Fprintf(os.Stderr, "Bool menu: unknown case '%s' from file %s\n", caseName, name)
				continue
			}
			capturedCase := bc
			opt := m.AddOption(stem)
			opt.Subscribe(gui.OnClick, func(string, interface{}) {
				selectBoolCase(vs, capturedCase)
			})
		}
	}

	addBoolFileItems(passFiles)
	if len(passFiles) > 0 && len(failFiles) > 0 {
		m.AddSeparator()
	}
	addBoolFileItems(failFiles)

	m.AddSeparator()

	union := m.AddOption("Union (A ∪ B)")
	union.SetShortcut(window.ModControl, window.KeyU)
	union.Subscribe(gui.OnClick, func(string, interface{}) {
		runBoolOp(vs, solid.BoolUnion, "Union")
	})

	inter := m.AddOption("Intersection (A ∩ B)")
	inter.SetShortcut(window.ModControl, window.KeyI)
	inter.Subscribe(gui.OnClick, func(string, interface{}) {
		runBoolOp(vs, solid.BoolIntersection, "Intersection")
	})

	diff := m.AddOption("Difference (A − B)")
	diff.SetShortcut(window.ModControl, window.KeyD)
	diff.Subscribe(gui.OnClick, func(string, interface{}) {
		runBoolOp(vs, solid.BoolDifference, "Difference")
	})

	return m
}

// selectBoolCase builds A and B solids for the chosen case and displays them.
func selectBoolCase(vs *viewerState, bc *boolCase) {
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

	// Remove previous bool meshes
	if vs.meshANode != nil {
		vs.scene.Remove(vs.meshANode)
		vs.meshANode = nil
	}
	if vs.meshBNode != nil {
		vs.scene.Remove(vs.meshBNode)
		vs.meshBNode = nil
	}
	for _, rn := range vs.resultNodes {
		vs.scene.Remove(rn)
	}
	vs.resultNodes = nil

	// Add meshes for A (yellow) and B (blue)
	meshA := solidToMesh(vs.solidA, &math32.Color{R: 1, G: 0.9, B: 0.2})
	meshB := solidToMesh(vs.solidB, &math32.Color{R: 0.4, G: 0.6, B: 1.0})

	if meshA != nil {
		vs.meshANode = core.NewNode()
		vs.meshANode.Add(meshA)
		vs.scene.Add(vs.meshANode)
	}
	if meshB != nil {
		vs.meshBNode = core.NewNode()
		vs.meshBNode.Add(meshB)
		vs.scene.Add(vs.meshBNode)
	}

	statsA := solidStatsStr(vs.solidA)
	statsB := solidStatsStr(vs.solidB)
	fmt.Fprintf(os.Stderr, "Bool case '%s': A(%s), B(%s)\n", bc.name, statsA, statsB)
}

// runBoolOp runs a boolean operation on the current A and B and displays the result.
func runBoolOp(vs *viewerState, op int, name string) {
	if vs.currentCase == nil {
		fmt.Fprintf(os.Stderr, "Bool %s: no case selected\n", name)
		return
	}

	// Build fresh solids from the case definition
	a := vs.currentCase.makeA()
	b := vs.currentCase.makeB()

	result, ok := solid.BoolOp(a, b, op)
	if !ok || result == nil {
		fmt.Fprintf(os.Stderr, "Bool %s: operation failed or empty result\n", name)
		return
	}

	result.CalcPlaneEquations()

	stats := solidStatsStr(result)
	mn, mx := result.Extents()
	fmt.Fprintf(os.Stderr, "Bool %s: %s, bbox [%.2f,%.2f,%.2f]-[%.2f,%.2f,%.2f]\n",
		name, stats, mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z)

	errs := result.VerifyDetailed()
	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Bool %s: %d errors\n", name, len(errs))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  %v\n", e)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Bool %s: valid!\n", name)
	}

	// Build g3n mesh from result solid and add to scene
	mesh := solidToMesh(result, &math32.Color{R: 0.2, G: 0.8, B: 0.3})
	if mesh == nil {
		fmt.Fprintf(os.Stderr, "Bool %s: could not build mesh from result\n", name)
		return
	}

	resultGroup := core.NewNode()
	span := float32(vs.resultSpan)
	switch op {
	case solid.BoolUnion:
		resultGroup.SetPosition(span, 0, 0)
	case solid.BoolIntersection:
		resultGroup.SetPosition(0, span, 0)
	case solid.BoolDifference:
		resultGroup.SetPosition(-span, 0, 0)
	}
	resultGroup.Add(mesh)
	vs.scene.Add(resultGroup)
	vs.resultNodes = append(vs.resultNodes, resultGroup)

	pos := resultGroup.Position()
	fmt.Fprintf(os.Stderr, "Bool %s: result displayed at (%.1f, %.1f, %.1f)\n", name, pos.X, pos.Y, pos.Z)
}

// solidToMesh converts a B-rep Solid into a g3n Mesh for rendering.
func solidToMesh(s *solid.Solid, color *math32.Color) *graphic.Mesh {
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
		f.LoopOut.ForEachHe(func(he *solid.HalfEdge) bool {
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
func solidStatsStr(s *solid.Solid) string {
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
	vs.resultNodes = nil
	vs.solidA = nil
	vs.solidB = nil
	vs.normalsNode = nil
	vs.normalsShown = false

	fmt.Fprintf(os.Stderr, "Cleared %d geometry nodes from scene\n", len(toRemove))
}

// ────────────────────────── Debug ──────────────────────────

func buildDebugMenu(vs *viewerState) *gui.Menu {
	m := gui.NewMenu()

	euler := m.AddOption("Show Euler Stats")
	euler.SetShortcut(window.ModControl, window.KeyE)
	euler.Subscribe(gui.OnClick, func(string, interface{}) {
		showEulerStats(vs)
	})

	normals := m.AddOption("Toggle Face Normals")
	normals.SetShortcut(window.ModControl, window.KeyN)
	normals.Subscribe(gui.OnClick, func(string, interface{}) {
		toggleNormals(vs)
	})

	halfedge := m.AddOption("Show Half-Edge Info")
	halfedge.SetShortcut(window.ModControl, window.KeyH)
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
	for i, s := range []*solid.Solid{vs.solidA, vs.solidB} {
		label := "A"
		if i == 1 {
			label = "B"
		}
		if s == nil {
			continue
		}
		stats := solidStatsStr(s)
		fmt.Fprintf(os.Stderr, "Solid %s: %s\n", label, stats)
		errs := s.VerifyDetailed()
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
	for i, sol := range []*solid.Solid{vs.solidA, vs.solidB} {
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
	f.Close()

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
	converter.Convert(vrmlNodes, vs.scene, baseDir)

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
