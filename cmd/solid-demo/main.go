// solid-demo demonstrates the solid modeling SDK ported from the original
// C++ VRaniML library. It exercises construction, sweeps, queries,
// geometric clean-up, structural verification, and plane-split operations.
package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func banner(title string) {
	fmt.Printf("\n%s\n%s\n", title, strings.Repeat("-", len(title)))
}

func dumpStats(label string, s *solid.Solid) {
	f, e, v := s.Stats()
	fmt.Printf("  %-22s faces=%d  edges=%d  verts=%d\n", label, f, e, v)
}

func dumpExtents(s *solid.Solid) {
	mn, mx := s.Extents()
	fmt.Printf("  extents  min=(%.2f, %.2f, %.2f)  max=(%.2f, %.2f, %.2f)\n",
		mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z)
}

func makeCube() *solid.Solid {
	positions := []vec.SFVec3f{
		{X: -1, Y: 1, Z: 1},
		{X: 1, Y: 1, Z: 1},
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: 1},
	}
	indices := []int64{0, 1, 2, 3, -1}
	color := vec.SFColor{R: 0.8, G: 0.2, B: 0.2, A: 1}
	s := solid.BuildFromIndexSet(positions, indices, color)
	if s == nil {
		return nil
	}
	s.TranslationalSweep(s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: -2})
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func makeTrianglePrism() *solid.Solid {
	positions := []vec.SFVec3f{
		{X: 0, Y: 1, Z: 1},
		{X: 1, Y: -1, Z: 1},
		{X: -1, Y: -1, Z: 1},
	}
	indices := []int64{0, 1, 2, -1}
	color := vec.SFColor{R: 0.2, G: 0.6, B: 0.8, A: 1}
	s := solid.BuildFromIndexSet(positions, indices, color)
	if s == nil {
		return nil
	}
	s.TranslationalSweep(s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: -2})
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func main() {
	banner("1. Build a cube [-1,1]^3")
	cube := makeCube()
	if cube == nil {
		fmt.Fprintln(os.Stderr, "failed to build cube")
		os.Exit(1)
	}
	dumpStats("cube:", cube)
	dumpExtents(cube)
	fmt.Printf("  volume = %.2f\n", cube.Volume())

	banner("2. Structural verification (VerifyDetailed)")
	if errs := cube.VerifyDetailed(); len(errs) > 0 {
		for _, e := range errs {
			fmt.Printf("  error: %v\n", e)
		}
	} else {
		fmt.Println("  cube passes all structural checks")
	}

	banner("3. Vertex neighborhood queries")
	v0 := cube.Verts
	faces := v0.VertexFaces()
	neighbors := v0.VertexNeighbors()
	edges := v0.VertexEdges()
	fmt.Printf("  vertex %d at %v\n", v0.Index, v0.Loc)
	fmt.Printf("    adjacent faces:    %d\n", len(faces))
	fmt.Printf("    adjacent edges:    %d\n", len(edges))
	fmt.Printf("    neighbor vertices: %d\n", len(neighbors))

	banner("4. Face neighborhood queries")
	f0 := cube.Faces
	faceNeighbors := f0.FaceNeighbors()
	fmt.Printf("  face %d  normal=(%.2f, %.2f, %.2f)\n",
		f0.Index, f0.Normal.X, f0.Normal.Y, f0.Normal.Z)
	fmt.Printf("    neighbor faces: %d\n", len(faceNeighbors))

	banner("5. Coplanarity and collinearity checks")
	_, _, hasCop := cube.Faces.HasCoplanarNeighbor()
	fmt.Printf("  first face has coplanar neighbor? %v\n", hasCop)
	fmt.Printf("  solid has colinear verts?         %v\n", cube.HasColinearVerts())
	fmt.Printf("  degenerate faces?                 %v\n", cube.HasDegenerateFaces() != nil)

	banner("6. Copy + translate")
	shifted := cube.Copy()
	shifted.TransformGeometry(vec.TranslationMatrix(3, 0, 0))
	shifted.CalcPlaneEquations()
	shifted.Renumber()
	dumpStats("shifted cube:", shifted)
	dumpExtents(shifted)

	banner("7. Merge (original + shifted copy)")
	merged := cube.Copy()
	shiftedCopy := shifted.Copy()
	merged.Merge(shiftedCopy)
	merged.Renumber()
	dumpStats("merged:", merged)

	banner("8. Triangulate a cube copy")
	triCube := cube.Copy()
	triCube.Triangulate()
	triCube.Renumber()
	dumpStats("triangulated cube:", triCube)

	banner("9. Build a triangular prism")
	prism := makeTrianglePrism()
	if prism == nil {
		fmt.Fprintln(os.Stderr, "failed to build prism")
		os.Exit(1)
	}
	dumpStats("prism:", prism)
	dumpExtents(prism)
	fmt.Printf("  volume = %.2f\n", prism.Volume())

	banner("10. Split cube along z=0 plane")
	splitCube := makeCube()
	zPlane := solid.Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: 0}
	above, below, ok := splitCube.Split(zPlane)
	if !ok {
		fmt.Println("  split produced no intersection (unexpected)")
	} else {
		above.Renumber()
		below.Renumber()
		dumpStats("above (z>0):", above)
		dumpExtents(above)
		fmt.Printf("  volume = %.2f\n", above.Volume())
		dumpStats("below (z<0):", below)
		dumpExtents(below)
		fmt.Printf("  volume = %.2f\n", below.Volume())
	}

	banner("11. Split cube along x=0 plane")
	splitCube2 := makeCube()
	xPlane := solid.Plane{Normal: vec.SFVec3f{X: 1, Y: 0, Z: 0}, D: 0}
	left, right, ok := splitCube2.Split(xPlane)
	if !ok {
		fmt.Println("  split produced no intersection (unexpected)")
	} else {
		left.Renumber()
		right.Renumber()
		dumpStats("left  (x>0):", left)
		dumpStats("right (x<0):", right)
	}

	banner("12. Split prism along z=0 plane")
	splitPrism := makeTrianglePrism()
	above2, below2, ok := splitPrism.Split(zPlane)
	if !ok {
		fmt.Println("  split produced no intersection (unexpected)")
	} else {
		above2.Renumber()
		below2.Renumber()
		dumpStats("above (z>0):", above2)
		dumpStats("below (z<0):", below2)
	}

	banner("13. Split cube with non-intersecting plane")
	noHitCube := makeCube()
	farPlane := solid.Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: -10}
	_, _, ok = noHitCube.Split(farPlane)
	fmt.Printf("  plane at z=10 intersects cube? %v (expected false)\n", ok)

	banner("14. Verify split results")
	if above != nil {
		if errs := above.VerifyDetailed(); len(errs) > 0 {
			fmt.Printf("  above half: %d errors\n", len(errs))
			for _, e := range errs {
				fmt.Printf("    %v\n", e)
			}
		} else {
			fmt.Println("  above half passes verification")
		}
	}
	if below != nil {
		if errs := below.VerifyDetailed(); len(errs) > 0 {
			fmt.Printf("  below half: %d errors\n", len(errs))
			for _, e := range errs {
				fmt.Printf("    %v\n", e)
			}
		} else {
			fmt.Println("  below half passes verification")
		}
	}

	banner("15. Export animated split to VRML")
	outDir := "."
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	// ── Plane 1: tilted 25° from Z toward X (green, diagonal cut) ──
	tilt1Deg := 25.0
	tilt1Rad := tilt1Deg * math.Pi / 180.0
	n1x := float64(math.Sin(tilt1Rad))
	n1z := float64(math.Cos(tilt1Rad))
	plane1 := solid.Plane{Normal: vec.SFVec3f{X: n1x, Y: 0, Z: n1z}, D: 0}

	// ── Plane 2: tilted 30° from Y toward X (red, perpendicular-ish cut) ──
	tilt2Deg := 30.0
	tilt2Rad := tilt2Deg * math.Pi / 180.0
	n2x := float64(math.Sin(tilt2Rad))
	n2y := float64(math.Cos(tilt2Rad))
	plane2 := solid.Plane{Normal: vec.SFVec3f{X: n2x, Y: n2y, Z: 0}, D: 0}

	// ── Plane 3: Y=0 (blue, horizontal cut) ──
	plane3 := solid.Plane{Normal: vec.SFVec3f{X: 0, Y: 1, Z: 0}, D: 0}

	yellow := vec.SFColor{R: 1, G: 0.85, B: 0.1, A: 1}

	// Build whole cube (Switch choice 0).
	wholeCube := makeCube()
	wholeCube.SetColor(yellow)
	wholeCube.Renumber()

	// First split → two halves (Switch choice 1).
	cube1 := makeCube()
	cube1.SetColor(yellow)
	half1, half2, splitOk := cube1.Split(plane1)
	if !splitOk {
		fmt.Fprintln(os.Stderr, "  first split failed")
		fmt.Println()
		return
	}
	half1.SetColor(yellow)
	half1.Renumber()
	half2.SetColor(yellow)
	half2.Renumber()

	// Second split → four quarters (Switch choice 2).
	q1a, q1b, ok1 := half1.Copy().Split(plane2)
	q2a, q2b, ok2 := half2.Copy().Split(plane2)
	if !ok1 || !ok2 {
		fmt.Fprintln(os.Stderr, "  second split failed")
		fmt.Println()
		return
	}
	quarters := []*solid.Solid{q1a, q1b, q2a, q2b}
	for _, q := range quarters {
		q.SetColor(yellow)
		q.Renumber()
	}

	// Third split → eight octants (Switch choice 3).
	octants := make([]*solid.Solid, 8)
	octNames := []string{"o1", "o2", "o3", "o4", "o5", "o6", "o7", "o8"}
	allOk := true
	for i, q := range quarters {
		a, b, ok := q.Copy().Split(plane3)
		if !ok {
			fmt.Fprintf(os.Stderr, "  third split of quarter %d failed\n", i)
			allOk = false
			break
		}
		a.SetColor(yellow)
		a.Renumber()
		b.SetColor(yellow)
		b.Renumber()
		octants[i*2] = a
		octants[i*2+1] = b
	}
	if !allOk {
		fmt.Println()
		return
	}

	var oValid [8]bool
	for i, o := range octants {
		o.CalcPlaneEquations()
		o.Renumber()
		errs := o.VerifyDetailed()
		oValid[i] = len(errs) == 0
		f, e, v := o.Stats()
		mn, mx := o.Extents()
		euler := f + v - e
		status := "VALID"
		if !oValid[i] {
			status = "INVALID"
		}
		fmt.Printf("  %s: %s  F=%d E=%d V=%d  euler=%d (want 2)  vol=%.3f\n",
			octNames[i], status, f, e, v, euler, o.Volume())
		fmt.Printf("    extents min=(%.2f,%.2f,%.2f) max=(%.2f,%.2f,%.2f)\n",
			mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z)
		for _, err := range errs {
			fmt.Printf("    error: %v\n", err)
		}
	}

	outPath := filepath.Join(outDir, "examples", "split_anim.wrl")
	params := animParams{
		whole: wholeCube,
		half1: half1, half2: half2,
		quarters: quarters,
		octants:  octants,
		oValid:   oValid,
		n1x: n1x, n1z: n1z, tilt1Rad: float64(tilt1Rad),
		n2x: n2x, n2y: n2y, tilt2Rad: float64(tilt2Rad),
	}
	if err := writeAnimatedSplit(outPath, params); err != nil {
		fmt.Fprintf(os.Stderr, "  export error: %v\n", err)
	} else {
		fmt.Printf("  wrote %s\n", outPath)
		fmt.Println("  view with: go run ./cmd/viewer/ " + outPath)
	}

	fmt.Println()
}

// animParams bundles the geometry and plane parameters for the three-stage animation.
type animParams struct {
	whole                *solid.Solid
	half1, half2         *solid.Solid
	quarters             []*solid.Solid // [4]
	octants              []*solid.Solid // [8]
	oValid               [8]bool
	n1x, n1z, tilt1Rad   float64
	n2x, n2y, tilt2Rad   float64
}

// writeAnimatedSplit creates a VRML97 three-stage animated split demo.
//
// Switch choices:
//
//	0 — whole yellow cube
//	1 — two halves (green plane)
//	2 — four quarters (red plane)
//	3 — eight octants (blue plane)
//	4 — eight colored octants (blue=valid, orange=invalid)
//
// Timeline (12-second loop):
//
//	0.00–0.04  whole cube
//	0.04–0.10  green plane slides in
//	0.10–0.12  pause
//	0.12       Switch 0→1
//	0.12–0.20  halves slide apart
//	0.20–0.22  pause
//	0.22–0.28  red plane slides in
//	0.28–0.30  pause
//	0.30       Switch 1→2
//	0.30–0.38  quarters fly apart
//	0.38–0.40  pause
//	0.40–0.46  blue plane slides in
//	0.46–0.48  pause
//	0.48       Switch 2→3
//	0.48–0.56  octants fly apart
//	0.60       Switch 3→4 (validation colors)
//	0.60–1.00  hold colored view
func writeAnimatedSplit(path string, p animParams) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := func(s string) { fmt.Fprintln(f, s) }
	wf := func(format string, args ...any) { fmt.Fprintf(f, format+"\n", args...) }

	// Separation vectors.
	sep1 := float64(1.5)
	h1End := vec.SFVec3f{X: -p.n1x * sep1, Y: 0, Z: -p.n1z * sep1}
	h2End := vec.SFVec3f{X: p.n1x * sep1, Y: 0, Z: p.n1z * sep1}

	sep2 := float64(1.2)
	q_up := vec.SFVec3f{X: p.n2x * sep2, Y: p.n2y * sep2, Z: 0}
	q_dn := vec.SFVec3f{X: -p.n2x * sep2, Y: -p.n2y * sep2, Z: 0}

	// Quarter end positions (half offset + plane2 offset).
	qEnds := [4]vec.SFVec3f{
		{X: h1End.X + q_dn.X, Y: h1End.Y + q_dn.Y, Z: h1End.Z + q_dn.Z},
		{X: h1End.X + q_up.X, Y: h1End.Y + q_up.Y, Z: h1End.Z + q_up.Z},
		{X: h2End.X + q_dn.X, Y: h2End.Y + q_dn.Y, Z: h2End.Z + q_dn.Z},
		{X: h2End.X + q_up.X, Y: h2End.Y + q_up.Y, Z: h2End.Z + q_up.Z},
	}

	// Octant separation: +Y or -Y from parent quarter.
	sep3 := float64(0.8)
	oUp := vec.SFVec3f{X: 0, Y: sep3, Z: 0}
	oDn := vec.SFVec3f{X: 0, Y: -sep3, Z: 0}

	// Octant end positions = quarter offset + Y offset.
	// octants[i*2] = "above" (a), octants[i*2+1] = "below" (b) from plane3 (Y=0).
	var oEnds [8]vec.SFVec3f
	for i := 0; i < 4; i++ {
		qe := qEnds[i]
		oEnds[i*2] = vec.SFVec3f{X: qe.X + oDn.X, Y: qe.Y + oDn.Y, Z: qe.Z + oDn.Z}
		oEnds[i*2+1] = vec.SFVec3f{X: qe.X + oUp.X, Y: qe.Y + oUp.Y, Z: qe.Z + oUp.Z}
	}

	// Plane approach positions.
	plane1Far := vec.SFVec3f{X: p.n1x * 5, Y: 0, Z: p.n1z * 5}
	plane2Far := vec.SFVec3f{X: p.n2x * 5, Y: p.n2y * 5, Z: 0}
	plane3Far := vec.SFVec3f{X: 0, Y: 5, Z: 0}

	w("#VRML V2.0 utf8")
	w("")
	w("# Three-stage animated plane-split demo")
	w("# Generated by solid-demo from the trueblocks-vranimal SDK")
	w("")
	w("NavigationInfo { type \"EXAMINE\" headlight TRUE }")
	w("Viewpoint { position 0 5 11 orientation 1 0 0 -0.35 description \"Front\" }")
	w("Background { skyColor [ 0.08 0.08 0.15 ] }")
	w("DirectionalLight { direction -0.5 -1 -0.5 intensity 0.8 }")
	w("DirectionalLight { direction 0.5 0.5 1 intensity 0.4 }")
	w("")

	// ── Switch: 5 choices ──
	w("DEF SPLITTER Switch {")
	w("  whichChoice 0")
	w("  choice [")
	w("")

	// Choice 0: whole cube
	w("    # Choice 0: whole cube")
	w("    Group { children [")
	if err := p.whole.ExportVRMLShape(f, "      "); err != nil {
		return err
	}
	w("    ] }")
	w("")

	// Choice 1: two halves
	w("    # Choice 1: two halves")
	w("    Group { children [")
	w("      DEF HALF1 Transform { translation 0 0 0 children [")
	if err := p.half1.ExportVRMLShape(f, "        "); err != nil {
		return err
	}
	w("      ] }")
	w("      DEF HALF2 Transform { translation 0 0 0 children [")
	if err := p.half2.ExportVRMLShape(f, "        "); err != nil {
		return err
	}
	w("      ] }")
	w("    ] }")
	w("")

	// Choice 2: four quarters
	qDEFs := []string{"Q0", "Q1", "Q2", "Q3"}
	w("    # Choice 2: four quarters")
	w("    Group { children [")
	for i, q := range p.quarters {
		wf("      DEF %s Transform { translation 0 0 0 children [", qDEFs[i])
		if err := q.ExportVRMLShape(f, "        "); err != nil {
			return err
		}
		w("      ] }")
	}
	w("    ] }")
	w("")

	// Choice 3: eight octants
	oDEFs := []string{"O0", "O1", "O2", "O3", "O4", "O5", "O6", "O7"}
	w("    # Choice 3: eight octants")
	w("    Group { children [")
	for i, o := range p.octants {
		wf("      DEF %s Transform { translation 0 0 0 children [", oDEFs[i])
		if err := o.ExportVRMLShape(f, "        "); err != nil {
			return err
		}
		w("      ] }")
	}
	w("    ] }")
	w("")

	// Choice 4: eight colored octants (blue=valid, orange=invalid)
	yellow := vec.SFColor{R: 1, G: 0.85, B: 0.1, A: 1}
	blue := vec.SFColor{R: 0.2, G: 0.4, B: 0.95, A: 1}
	orange := vec.SFColor{R: 1.0, G: 0.5, B: 0.1, A: 1}
	coDEFs := []string{"CO0", "CO1", "CO2", "CO3", "CO4", "CO5", "CO6", "CO7"}
	w("    # Choice 4: verified octants (blue=valid, orange=invalid)")
	w("    Group { children [")
	for i, o := range p.octants {
		clr := blue
		if !p.oValid[i] {
			clr = orange
		}
		o.SetColor(clr)
		wf("      DEF %s Transform { translation 0 0 0 children [", coDEFs[i])
		if err := o.ExportVRMLShape(f, "        "); err != nil {
			return err
		}
		w("      ] }")
		o.SetColor(yellow)
	}
	w("    ] }")
	w("")

	w("  ]")
	w("}")
	w("")

	// ── Green cutting plane (plane 1) ──
	w("DEF CUTTER1 Transform {")
	wf("  translation %g %g %g", plane1Far.X, plane1Far.Y, plane1Far.Z)
	wf("  rotation 0 1 0 %g", p.tilt1Rad)
	w("  children [ Shape {")
	w("    appearance Appearance { material Material { diffuseColor 0.2 0.9 0.2  transparency 0.5 } }")
	w("    geometry Box { size 3.5 3.5 0.02 }")
	w("  } ]")
	w("}")
	w("")

	// ── Red cutting plane (plane 2) ──
	w("DEF CUTTER2 Transform {")
	wf("  translation %g %g %g", plane2Far.X, plane2Far.Y, plane2Far.Z)
	wf("  rotation 0 0 1 %g", -p.tilt2Rad)
	w("  children [ Shape {")
	w("    appearance Appearance { material Material { diffuseColor 0.9 0.2 0.2  transparency 0.5 } }")
	w("    geometry Box { size 7 0.02 7 }")
	w("  } ]")
	w("}")
	w("")

	// ── Blue cutting plane (plane 3: Y=0) ──
	w("DEF CUTTER3 Transform {")
	wf("  translation %g %g %g", plane3Far.X, plane3Far.Y, plane3Far.Z)
	w("  children [ Shape {")
	w("    appearance Appearance { material Material { diffuseColor 0.2 0.4 0.95  transparency 0.5 } }")
	w("    geometry Box { size 7 0.02 7 }")
	w("  } ]")
	w("}")
	w("")

	// ── Timer: 12-second loop ──
	w("DEF CLOCK TimeSensor { cycleInterval 12  loop TRUE }")
	w("")

	// ── Switch selector: 0→1→2→3→4 ──
	w("DEF SWITCH_SEL ScalarInterpolator {")
	w("  key      [ 0, 0.119, 0.12, 0.299, 0.30, 0.479, 0.48, 0.599, 0.60, 1.0 ]")
	w("  keyValue [ 0, 0,     1,    1,     2,    2,     3,    3,     4,    4   ]")
	w("}")
	w("")

	// ── Green plane: far → center → far ──
	w("DEF CUT1_POS PositionInterpolator {")
	w("  key [ 0, 0.04, 0.10, 0.12, 0.18, 1.0 ]")
	w("  keyValue [")
	wf("    %g %g %g,", plane1Far.X, plane1Far.Y, plane1Far.Z)
	wf("    %g %g %g,", plane1Far.X, plane1Far.Y, plane1Far.Z)
	w("    0 0 0,")
	w("    0 0 0,")
	wf("    %g %g %g,", plane1Far.X, plane1Far.Y, plane1Far.Z)
	wf("    %g %g %g", plane1Far.X, plane1Far.Y, plane1Far.Z)
	w("  ]")
	w("}")
	w("")

	// ── Red plane: far → center → far ──
	w("DEF CUT2_POS PositionInterpolator {")
	w("  key [ 0, 0.22, 0.28, 0.30, 0.36, 1.0 ]")
	w("  keyValue [")
	wf("    %g %g %g,", plane2Far.X, plane2Far.Y, plane2Far.Z)
	wf("    %g %g %g,", plane2Far.X, plane2Far.Y, plane2Far.Z)
	w("    0 0 0,")
	w("    0 0 0,")
	wf("    %g %g %g,", plane2Far.X, plane2Far.Y, plane2Far.Z)
	wf("    %g %g %g", plane2Far.X, plane2Far.Y, plane2Far.Z)
	w("  ]")
	w("}")
	w("")

	// ── Blue plane: far → center → far ──
	w("DEF CUT3_POS PositionInterpolator {")
	w("  key [ 0, 0.40, 0.46, 0.48, 0.54, 1.0 ]")
	w("  keyValue [")
	wf("    %g %g %g,", plane3Far.X, plane3Far.Y, plane3Far.Z)
	wf("    %g %g %g,", plane3Far.X, plane3Far.Y, plane3Far.Z)
	w("    0 0 0,")
	w("    0 0 0,")
	wf("    %g %g %g,", plane3Far.X, plane3Far.Y, plane3Far.Z)
	wf("    %g %g %g", plane3Far.X, plane3Far.Y, plane3Far.Z)
	w("  ]")
	w("}")
	w("")

	// ── Half positions ──
	writePos := func(name string, keys string, kv []vec.SFVec3f) {
		wf("DEF %s PositionInterpolator {", name)
		wf("  key %s", keys)
		w("  keyValue [")
		for i, v := range kv {
			sep := ","
			if i == len(kv)-1 {
				sep = ""
			}
			wf("    %g %g %g%s", v.X, v.Y, v.Z, sep)
		}
		w("  ]")
		w("}")
		w("")
	}

	zero := vec.SFVec3f{}
	writePos("HALF1_POS", "[ 0, 0.12, 0.20, 1.0 ]", []vec.SFVec3f{zero, zero, h1End, h1End})
	writePos("HALF2_POS", "[ 0, 0.12, 0.20, 1.0 ]", []vec.SFVec3f{zero, zero, h2End, h2End})

	// ── Quarter positions: start at parent half's end, fly to quarter offset ──
	// q0,q1 come from half1; q2,q3 come from half2.
	halfEnds := [4]vec.SFVec3f{h1End, h1End, h2End, h2End}
	for i, name := range qDEFs {
		writePos(name+"_POS", "[ 0, 0.30, 0.38, 1.0 ]",
			[]vec.SFVec3f{halfEnds[i], halfEnds[i], qEnds[i], qEnds[i]})
	}

	// ── Octant positions: start at parent quarter offset, fly to octant offset ──
	for i, name := range oDEFs {
		parentQ := i / 2
		writePos(name+"_POS", "[ 0, 0.48, 0.56, 1.0 ]",
			[]vec.SFVec3f{qEnds[parentQ], qEnds[parentQ], oEnds[i], oEnds[i]})
	}

	// ── Colored octant positions (choice 4): same as octant positions ──
	for i, name := range coDEFs {
		parentQ := i / 2
		writePos(name+"_POS", "[ 0, 0.48, 0.56, 1.0 ]",
			[]vec.SFVec3f{qEnds[parentQ], qEnds[parentQ], oEnds[i], oEnds[i]})
	}

	// ── Routes ──
	w("# Switch selector")
	w("ROUTE CLOCK.fraction_changed TO SWITCH_SEL.set_fraction")
	w("ROUTE SWITCH_SEL.value_changed TO SPLITTER.whichChoice")
	w("")
	w("# Green plane (cut 1)")
	w("ROUTE CLOCK.fraction_changed TO CUT1_POS.set_fraction")
	w("ROUTE CUT1_POS.value_changed TO CUTTER1.set_translation")
	w("")
	w("# Red plane (cut 2)")
	w("ROUTE CLOCK.fraction_changed TO CUT2_POS.set_fraction")
	w("ROUTE CUT2_POS.value_changed TO CUTTER2.set_translation")
	w("")
	w("# Blue plane (cut 3)")
	w("ROUTE CLOCK.fraction_changed TO CUT3_POS.set_fraction")
	w("ROUTE CUT3_POS.value_changed TO CUTTER3.set_translation")
	w("")
	w("# Halves (choice 1)")
	w("ROUTE CLOCK.fraction_changed TO HALF1_POS.set_fraction")
	w("ROUTE HALF1_POS.value_changed TO HALF1.set_translation")
	w("ROUTE CLOCK.fraction_changed TO HALF2_POS.set_fraction")
	w("ROUTE HALF2_POS.value_changed TO HALF2.set_translation")
	w("")

	w("# Quarters (choice 2)")
	for _, name := range qDEFs {
		wf("ROUTE CLOCK.fraction_changed TO %s_POS.set_fraction", name)
		wf("ROUTE %s_POS.value_changed TO %s.set_translation", name, name)
	}
	w("")

	w("# Octants (choice 3)")
	for _, name := range oDEFs {
		wf("ROUTE CLOCK.fraction_changed TO %s_POS.set_fraction", name)
		wf("ROUTE %s_POS.value_changed TO %s.set_translation", name, name)
	}
	w("")

	w("# Colored octants (choice 4)")
	for _, name := range coDEFs {
		wf("ROUTE CLOCK.fraction_changed TO %s_POS.set_fraction", name)
		wf("ROUTE %s_POS.value_changed TO %s.set_translation", name, name)
	}

	return nil
}
