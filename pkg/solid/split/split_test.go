package split

import (
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/primitives"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func makeCube(t *testing.T) *base.Solid {
	t.Helper()
	s := primitives.MakeCube(1.0, vec.Red)
	if s == nil {
		t.Skip("MakeCube returned nil")
	}
	return s
}

func TestPlane_GetDistance(t *testing.T) {
	pl := base.Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: 0}

	d := pl.GetDistance(vec.SFVec3f{X: 0, Y: 0, Z: 5})
	if d != 5 {
		t.Fatalf("expected 5, got %g", d)
	}

	d = pl.GetDistance(vec.SFVec3f{X: 0, Y: 0, Z: -3})
	if d != -3 {
		t.Fatalf("expected -3, got %g", d)
	}

	d = pl.GetDistance(vec.SFVec3f{X: 5, Y: 5, Z: 0})
	if d != 0 {
		t.Fatalf("expected 0, got %g", d)
	}
}

func TestPlane_GetDistance_Offset(t *testing.T) {
	pl := base.Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: -2}

	d := pl.GetDistance(vec.SFVec3f{X: 0, Y: 0, Z: 2})
	if math.Abs(d) > 1e-6 {
		t.Fatalf("expected 0, got %g", d)
	}

	d = pl.GetDistance(vec.SFVec3f{X: 0, Y: 0, Z: 5})
	if math.Abs(d-3) > 1e-6 {
		t.Fatalf("expected 3, got %g", d)
	}
}

func TestHalfEdge_Marks(t *testing.T) {
	he := &base.HalfEdge{}
	he.Mark = base.LOOSE
	if he.Mark != base.LOOSE {
		t.Fatal("should be LOOSE")
	}
	he.Mark = base.NOT_LOOSE
	if he.Mark == base.LOOSE {
		t.Fatal("should be NOT_LOOSE")
	}
}

func TestFloatCompare(t *testing.T) {
	if base.FloatCompare(1.0) != 1 {
		t.Fatal("positive should return 1")
	}
	if base.FloatCompare(-1.0) != -1 {
		t.Fatal("negative should return -1")
	}
	if base.FloatCompare(0.000001) != 0 {
		t.Fatal("near-zero should return 0")
	}
}

func TestSplit_CubeByZPlane(t *testing.T) {
	s := makeCube(t)

	sp := base.Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: 0}
	above, below, ok := Split(s, sp)
	if !ok {
		t.Fatal("split should succeed")
	}
	if above == nil || below == nil {
		t.Fatal("both halves should be non-nil")
	}
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
	s := makeCube(t)

	sp := base.Plane{Normal: vec.SFVec3f{X: 1, Y: 0, Z: 0}, D: 0}
	above, below, ok := Split(s, sp)
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
	s := makeCube(t)

	sp := base.Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: -100}
	_, _, ok := Split(s, sp)
	if ok {
		t.Fatal("split should fail when plane doesn't intersect")
	}
}

func TestSplit_TrianglePrism(t *testing.T) {
	s := primitives.MakeLamina([]vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 2, Y: 0, Z: 0},
		{X: 1, Y: 2, Z: 0},
	}, vec.Red)
	if s == nil {
		t.Fatal("MakeLamina returned nil")
	}
	primitives.TranslationalSweep(s, s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: 4})
	s.CalcPlaneEquations()
	s.Renumber()

	sp := base.Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: -2}
	above, below, ok := Split(s, sp)
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

func TestSplitRecord_AddVert_Deduplicate(t *testing.T) {
	sr := NewSplitRecord()
	sr.verts = nil
	v := &base.Vertex{Mark: base.UNKNOWN}
	sr.addVert(v)
	sr.addVert(v)
	if len(sr.verts) != 1 {
		t.Fatalf("expected 1 vert, got %d", len(sr.verts))
	}
}

func TestSplitRecord_CanJoin_NoMatch(t *testing.T) {
	sr := NewSplitRecord()
	sr.looseEnds = nil
	he := &base.HalfEdge{}
	result := sr.canJoin(he)
	if result != nil {
		t.Fatal("should return nil when no loose ends")
	}
	if len(sr.looseEnds) != 1 {
		t.Fatalf("should have 1 loose end, got %d", len(sr.looseEnds))
	}
}

func TestSplit_DoubleSplit(t *testing.T) {
	s := makeCube(t)

	tiltRad := 25.0 * math.Pi / 180.0
	n1x := math.Sin(tiltRad)
	n1z := math.Cos(tiltRad)
	plane1 := base.Plane{Normal: vec.SFVec3f{X: n1x, Y: 0, Z: n1z}, D: 0}

	half1, half2, ok := Split(s, plane1)
	if !ok {
		t.Fatal("first split should succeed")
	}
	if errs := algorithms.VerifyDetailed(half1); len(errs) > 0 {
		t.Fatalf("half1 invalid: %v", errs)
	}
	if errs := algorithms.VerifyDetailed(half2); len(errs) > 0 {
		t.Fatalf("half2 invalid: %v", errs)
	}

	tilt2Rad := 30.0 * math.Pi / 180.0
	n2x := math.Sin(tilt2Rad)
	n2y := math.Cos(tilt2Rad)
	plane2 := base.Plane{Normal: vec.SFVec3f{X: n2x, Y: n2y, Z: 0}, D: 0}

	h1Copy := half1.Copy()
	h1Copy.CalcPlaneEquations()
	q1a, q1b, ok1 := Split(h1Copy, plane2)
	if !ok1 {
		t.Fatal("second split of half1 should succeed")
	}

	h2Copy := half2.Copy()
	h2Copy.CalcPlaneEquations()
	q2a, q2b, ok2 := Split(h2Copy, plane2)
	if !ok2 {
		t.Fatal("second split of half2 should succeed")
	}

	quarters := []*base.Solid{q1a, q1b, q2a, q2b}
	names := []string{"q1a", "q1b", "q2a", "q2b"}
	for i, q := range quarters {
		q.CalcPlaneEquations()
		q.Renumber()
		f, e, v := q.Stats()
		eu := f + v - e
		t.Logf("%s: F=%d E=%d V=%d euler=%d vol=%.3f", names[i], f, e, v, eu, q.Volume())
		if errs := algorithms.VerifyDetailed(q); len(errs) > 0 {
			t.Errorf("%s: verification failed (%d errors):", names[i], len(errs))
			for _, err := range errs {
				t.Errorf("  %v", err)
			}
		}
	}

	totalVol := float64(0)
	for _, q := range quarters {
		totalVol += q.Volume()
	}
	if math.Abs(totalVol-8.0) > 0.1 {
		t.Errorf("total volume %.3f, expected ~8.0", totalVol)
	}
}
