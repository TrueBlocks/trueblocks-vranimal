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
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/boolop"
	solidExport "github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/export"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/primitives"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/writer"
)

// ────────────────────────── Bool case definitions ──────────────────────────

var (
	yellow    = vec.SFColor{R: 1.0, G: 0.9, B: 0.2, A: 1}
	lightBlue = vec.SFColor{R: 0.4, G: 0.6, B: 1.0, A: 1}
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
	currentCase  *boolCase     // selected case (nil = none)
	solidA       *base.Solid   // current A solid
	solidB       *base.Solid   // current B solid
	meshANode    *core.Node    // g3n node for A mesh
	meshBNode    *core.Node    // g3n node for B mesh
	resultNodes  []*core.Node  // g3n nodes for bool results
	resultSolids []*base.Solid // result solids (for export)
	resultNames  []string      // result operation names
	resultSpan   float64       // offset for result positioning

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

// exportScene uses osascript to pick a save path and serializes the current scene to VRML.
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

	// Add the three failing test cases directly
	caseOrder := []string{"same_size_partial_overlap", "same_size_edge_on_edge", "same_size_slight_twist"}
	for _, name := range caseOrder {
		bc := boolCaseMap[name]
		capturedCase := bc
		opt := m.AddOption(name)
		opt.Subscribe(gui.OnClick, func(string, interface{}) {
			selectBoolCase(vs, capturedCase)
		})
	}

	m.AddSeparator()

	union := m.AddOption("Union (A ∪ B)")
	union.SetShortcut(window.ModControl, window.KeyU)
	union.Subscribe(gui.OnClick, func(string, interface{}) {
		runBoolOp(vs, base.BoolUnion, "Union")
	})

	inter := m.AddOption("Intersection (A ∩ B)")
	inter.SetShortcut(window.ModControl, window.KeyI)
	inter.Subscribe(gui.OnClick, func(string, interface{}) {
		runBoolOp(vs, base.BoolIntersection, "Intersection")
	})

	diff := m.AddOption("Difference (A − B)")
	diff.SetShortcut(window.ModControl, window.KeyD)
	diff.Subscribe(gui.OnClick, func(string, interface{}) {
		runBoolOp(vs, base.BoolDifference, "Difference")
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
	vs.resultSolids = nil
	vs.resultNames = nil

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

	result, ok := boolop.BoolOp(a, b, op)
	if !ok || result == nil {
		fmt.Fprintf(os.Stderr, "Bool %s: operation failed or empty result\n", name)
		return
	}

	result.CalcPlaneEquations()

	stats := solidStatsStr(result)
	mn, mx := result.Extents()
	fmt.Fprintf(os.Stderr, "Bool %s: %s, bbox [%.2f,%.2f,%.2f]-[%.2f,%.2f,%.2f]\n",
		name, stats, mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z)

	errs := algorithms.VerifyDetailed(result)
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
	case base.BoolUnion:
		resultGroup.SetPosition(span, 0, 0)
	case base.BoolIntersection:
		resultGroup.SetPosition(0, span, 0)
	case base.BoolDifference:
		resultGroup.SetPosition(-span, 0, 0)
	}
	resultGroup.Add(mesh)
	vs.scene.Add(resultGroup)
	vs.resultNodes = append(vs.resultNodes, resultGroup)
	vs.resultSolids = append(vs.resultSolids, result)
	vs.resultNames = append(vs.resultNames, name)

	pos := resultGroup.Position()
	fmt.Fprintf(os.Stderr, "Bool %s: result displayed at (%.1f, %.1f, %.1f)\n", name, pos.X, pos.Y, pos.Z)
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
