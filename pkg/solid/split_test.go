package solid

import (
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Plane
// ---------------------------------------------------------------------------

func TestPlane_GetDistance(t *testing.T) {
	// Plane: z = 0 (Normal=(0,0,1), D=0)
	pl := Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: 0}

	// Point above
	d := pl.GetDistance(vec.SFVec3f{X: 0, Y: 0, Z: 5})
	if d != 5 {
		t.Fatalf("expected 5, got %g", d)
	}

	// Point below
	d = pl.GetDistance(vec.SFVec3f{X: 0, Y: 0, Z: -3})
	if d != -3 {
		t.Fatalf("expected -3, got %g", d)
	}

	// Point on plane
	d = pl.GetDistance(vec.SFVec3f{X: 5, Y: 5, Z: 0})
	if d != 0 {
		t.Fatalf("expected 0, got %g", d)
	}
}

func TestPlane_GetDistance_Offset(t *testing.T) {
	// Plane: z = 2  →  Normal=(0,0,1), D=-2
	pl := Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: -2}

	d := pl.GetDistance(vec.SFVec3f{X: 0, Y: 0, Z: 2})
	if math.Abs(float64(d)) > 1e-6 {
		t.Fatalf("expected 0, got %g", d)
	}

	d = pl.GetDistance(vec.SFVec3f{X: 0, Y: 0, Z: 5})
	if math.Abs(float64(d)-3) > 1e-6 {
		t.Fatalf("expected 3, got %g", d)
	}
}

// ---------------------------------------------------------------------------
// HalfEdge mark helpers
// ---------------------------------------------------------------------------

func TestHalfEdge_SetMark_Marked(t *testing.T) {
	he := &HalfEdge{}
	he.SetMark(LOOSE)
	if !he.Marked(LOOSE) {
		t.Fatal("should be LOOSE")
	}
	he.SetMark(NOT_LOOSE)
	if he.Marked(LOOSE) {
		t.Fatal("should be NOT_LOOSE")
	}
}

// ---------------------------------------------------------------------------
// floatCompare
// ---------------------------------------------------------------------------

func TestFloatCompare(t *testing.T) {
	if floatCompare(1.0, 0) != 1 {
		t.Fatal("positive should return 1")
	}
	if floatCompare(-1.0, 0) != -1 {
		t.Fatal("negative should return -1")
	}
	if floatCompare(0.000001, 0) != 0 {
		t.Fatal("near-zero should return 0")
	}
}

// ---------------------------------------------------------------------------
// Split — cube split by Z=0 plane
// ---------------------------------------------------------------------------

func makeCube() *Solid {
	// Build a cube from [-1,-1,-1] to [1,1,1] via sweeping.
	positions := []vec.SFVec3f{
		{X: -1, Y: -1, Z: -1},
		{X: 1, Y: -1, Z: -1},
		{X: 1, Y: 1, Z: -1},
		{X: -1, Y: 1, Z: -1},
	}
	indices := []int64{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		return nil
	}
	// Sweep upward to create the cube body.
	var bottomFace *Face
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut != nil && f.LoopOut.NHalfEdges() >= 4 {
			bottomFace = f
			break
		}
	}
	if bottomFace == nil {
		return nil
	}
	s.TranslationalSweep(bottomFace, vec.SFVec3f{X: 0, Y: 0, Z: 2})
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func TestSplit_CubeByZPlane(t *testing.T) {
	s := makeCube()
	if s == nil {
		t.Skip("failed to build cube")
	}

	// Split by z=0 plane: Normal=(0,0,1), D=0
	sp := Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: 0}
	above, below, ok := s.Split(sp)
	if !ok {
		t.Fatal("split should succeed")
	}
	if above == nil || below == nil {
		t.Fatal("both halves should be non-nil")
	}

	// Both halves should have faces and vertices.
	if above.NFaces() == 0 {
		t.Fatal("above should have faces")
	}
	if below.NFaces() == 0 {
		t.Fatal("below should have faces")
	}
	if above.NVerts() == 0 {
		t.Fatal("above should have vertices")
	}
	if below.NVerts() == 0 {
		t.Fatal("below should have vertices")
	}
}

func TestSplit_CubeByXPlane(t *testing.T) {
	s := makeCube()
	if s == nil {
		t.Skip("failed to build cube")
	}

	// Split by x=0 plane
	sp := Plane{Normal: vec.SFVec3f{X: 1, Y: 0, Z: 0}, D: 0}
	above, below, ok := s.Split(sp)
	if !ok {
		t.Fatal("split should succeed")
	}
	if above == nil || below == nil {
		t.Fatal("both halves should be non-nil")
	}
	if above.NFaces() == 0 || below.NFaces() == 0 {
		t.Fatal("both halves should have faces")
	}
}

func TestSplit_NoIntersection(t *testing.T) {
	s := makeCube()
	if s == nil {
		t.Skip("failed to build cube")
	}

	// Plane far above — no edge crosses.
	sp := Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: -100}
	_, _, ok := s.Split(sp)
	if ok {
		t.Fatal("split should fail when plane doesn't intersect")
	}
}

// ---------------------------------------------------------------------------
// Split — triangle split
// ---------------------------------------------------------------------------

func TestSplit_TrianglePrism(t *testing.T) {
	// Build a triangular prism and split it.
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 2, Y: 0, Z: 0},
		{X: 1, Y: 2, Z: 0},
	}
	indices := []int64{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
		return
	}
	var triF *Face
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut != nil && f.LoopOut.NHalfEdges() >= 3 {
			triF = f
			break
		}
	}
	if triF == nil {
		t.Skip("no triangle face")
	}
	s.TranslationalSweep(triF, vec.SFVec3f{X: 0, Y: 0, Z: 4})
	s.CalcPlaneEquations()
	s.Renumber()

	// Split at z=2
	sp := Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: -2}
	above, below, ok := s.Split(sp)
	if !ok {
		t.Fatal("split should succeed")
	}
	if above == nil || below == nil {
		t.Fatal("both halves should be non-nil")
	}
	if above.NVerts() == 0 || below.NVerts() == 0 {
		t.Fatal("both halves should have vertices")
	}
}

// ---------------------------------------------------------------------------
// Split record internals
// ---------------------------------------------------------------------------

func TestSplitRecord_AddVert_Deduplicate(t *testing.T) {
	sr := NewSplitRecord()
	sr.verts = nil
	v := &Vertex{Mark: UNKNOWN}
	sr.addVert(v)
	sr.addVert(v) // should be no-op (already marked ON)
	if len(sr.verts) != 1 {
		t.Fatalf("expected 1 vert, got %d", len(sr.verts))
	}
}

func TestSplitRecord_CanJoin_NoMatch(t *testing.T) {
	sr := NewSplitRecord()
	sr.looseEnds = nil
	he := &HalfEdge{}
	result := sr.canJoin(he)
	if result != nil {
		t.Fatal("should return nil when no loose ends")
	}
	if len(sr.looseEnds) != 1 {
		t.Fatalf("should have 1 loose end, got %d", len(sr.looseEnds))
	}
}

// ---------------------------------------------------------------------------
// SetFaceMarks2
// ---------------------------------------------------------------------------

func TestSolid_SetFaceMarks2(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int64{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
		return
	}

	s.SetFaceMarks2(VISITED)
	for f := s.Faces; f != nil; f = f.Next {
		if f.Mark2 != VISITED {
			t.Fatalf("face %d mark2 should be VISITED", f.Index)
		}
	}
}

func TestSplit_DebugGenerate(t *testing.T) {
	s := makeCube()
	if s == nil {
		t.Skip("failed to build cube")
		return
	}

	t.Logf("Cube: %d faces, %d edges, %d verts", s.NFaces(), s.NEdges(), s.NVerts())

	sp := Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: 0}

	// Check vertex distances
	for v := s.Verts; v != nil; v = v.Next {
		d := sp.GetDistance(v.Loc)
		t.Logf("v[%d] loc=%v dist=%g", v.Index, v.Loc, d)
	}

	// Check edges that straddle
	straddling := 0
	onPlane := 0
	for e := s.Edges; e != nil; e = e.Next {
		v1 := e.He1.Vertex
		v2 := e.He2.Vertex
		d1 := sp.GetDistance(v1.Loc)
		d2 := sp.GetDistance(v2.Loc)
		s1 := floatCompare(d1, 0)
		s2 := floatCompare(d2, 0)
		t.Logf("edge[%d]: v1[%d]=%v d1=%g s1=%d, v2[%d]=%v d2=%g s2=%d",
			e.Index, v1.Index, v1.Loc, d1, s1, v2.Index, v2.Loc, d2, s2)
		if (s1 == -1 && s2 == 1) || (s1 == 1 && s2 == -1) {
			straddling++
		}
		if s1 == 0 {
			onPlane++
		}
		if s2 == 0 {
			onPlane++
		}
	}
	t.Logf("Straddling edges: %d, On-plane vertex refs: %d", straddling, onPlane)

	// Run generate
	var above, below *Solid
	rec := NewSplitRecord()
	rec.reset(sp, s, &above, &below)
	s.SetFaceMarks(UNKNOWN)
	s.SetVertexMarks(UNKNOWN)

	// Count edges before generate
	edgeCount := 0
	for e := s.Edges; e != nil; e = e.Next {
		edgeCount++
	}
	t.Logf("Edge count before generate: %d", edgeCount)

	rec.generate()
	t.Logf("After generate: %d verts on plane, %d null edges", len(rec.verts), len(rec.edges))

	// Check Scratch values
	for v := s.Verts; v != nil; v = v.Next {
		t.Logf("v[%d] Scratch=%g", v.Index, v.Scratch)
	}

	// Run classify
	ok := rec.classify()
	t.Logf("After classify: ok=%v, %d null edges", ok, len(rec.edges))
}

// ---------------------------------------------------------------------------
// Double split — reproduces the bug where splitting a split result
// produces invalid solids (missing closing edges).
// ---------------------------------------------------------------------------

func TestSplit_DoubleSplit(t *testing.T) {
	s := makeCube()
	if s == nil {
		t.Fatal("failed to build cube")
	}

	// Plane 1: tilted 25° from Z toward X
	tiltRad := 25.0 * math.Pi / 180.0
	n1x := float64(math.Sin(tiltRad))
	n1z := float64(math.Cos(tiltRad))
	plane1 := Plane{Normal: vec.SFVec3f{X: n1x, Y: 0, Z: n1z}, D: 0}

	half1, half2, ok := s.Split(plane1)
	if !ok {
		t.Fatal("first split should succeed")
	}
	if errs := half1.VerifyDetailed(); len(errs) > 0 {
		t.Fatalf("half1 invalid: %v", errs)
	}
	if errs := half2.VerifyDetailed(); len(errs) > 0 {
		t.Fatalf("half2 invalid: %v", errs)
	}

	// Plane 2: tilted 30° in XY plane
	tilt2Rad := 30.0 * math.Pi / 180.0
	n2x := float64(math.Sin(tilt2Rad))
	n2y := float64(math.Cos(tilt2Rad))
	plane2 := Plane{Normal: vec.SFVec3f{X: n2x, Y: n2y, Z: 0}, D: 0}

	// Split both halves by plane2 to produce 4 quarters
	h1Copy := half1.Copy()
	h1Copy.CalcPlaneEquations()
	q1a, q1b, ok1 := h1Copy.Split(plane2)
	if !ok1 {
		t.Fatal("second split of half1 should succeed")
	}

	h2Copy := half2.Copy()
	h2Copy.CalcPlaneEquations()
	q2a, q2b, ok2 := h2Copy.Split(plane2)
	if !ok2 {
		t.Fatal("second split of half2 should succeed")
	}

	// All four quarters must pass Euler verification
	quarters := []*Solid{q1a, q1b, q2a, q2b}
	names := []string{"q1a", "q1b", "q2a", "q2b"}
	for i, q := range quarters {
		q.CalcPlaneEquations()
		q.Renumber()
		f, e, v := q.Stats()
		euler := f + v - e
		t.Logf("%s: F=%d E=%d V=%d euler=%d vol=%.3f", names[i], f, e, v, euler, q.Volume())
		if errs := q.VerifyDetailed(); len(errs) > 0 {
			t.Errorf("%s: verification failed (%d errors):", names[i], len(errs))
			for _, err := range errs {
				t.Errorf("  %v", err)
			}
		}
	}

	// Combined volume should equal original cube (8.0)
	totalVol := float64(0)
	for _, q := range quarters {
		totalVol += float64(q.Volume())
	}
	if math.Abs(totalVol-8.0) > 0.1 {
		t.Errorf("total volume %.3f, expected ~8.0", totalVol)
	}
}
