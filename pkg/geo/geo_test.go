package geo

import (
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

const eps = 1e-6

func approx(a, b float64) bool {
	return math.Abs(a-b) < eps
}

func approxVec(a, b vec.SFVec3f) bool {
	return approx(a.X, b.X) && approx(a.Y, b.Y) && approx(a.Z, b.Z)
}

// ===========================================================================
// Ray
// ===========================================================================

func TestNewRay(t *testing.T) {
	r := NewRay(vec.SFVec3f{X: 1}, vec.SFVec3f{Y: 1})
	if r.Loc.X != 1 || r.Dir.Y != 1 {
		t.Fatal("constructor failed")
	}
}

func TestRay_Evaluate(t *testing.T) {
	r := NewRay(vec.SFVec3f{}, vec.SFVec3f{X: 1, Y: 2, Z: 3})
	pt := r.Evaluate(2)
	if !approxVec(pt, vec.SFVec3f{X: 2, Y: 4, Z: 6}) {
		t.Fatalf("expected (2,4,6), got %v", pt)
	}
}

func TestRay_Evaluate_Origin(t *testing.T) {
	r := NewRay(vec.SFVec3f{X: 5, Y: 5, Z: 5}, vec.SFVec3f{X: 1})
	pt := r.Evaluate(0)
	if !approxVec(pt, vec.SFVec3f{X: 5, Y: 5, Z: 5}) {
		t.Fatalf("t=0 should return origin, got %v", pt)
	}
}

func TestRay_Evaluate_Negative(t *testing.T) {
	r := NewRay(vec.SFVec3f{}, vec.SFVec3f{X: 1})
	pt := r.Evaluate(-5)
	if !approxVec(pt, vec.SFVec3f{X: -5}) {
		t.Fatalf("negative t: %v", pt)
	}
}

func TestRay_GetDistance(t *testing.T) {
	r := NewRay(vec.SFVec3f{}, vec.SFVec3f{X: 3, Y: 4})
	// |dir*2| = |(6,8,0)| = 10
	d := r.GetDistance(2)
	if !approx(d, 10) {
		t.Fatalf("expected 10, got %g", d)
	}
}

func TestRay_Extrapolate(t *testing.T) {
	r := NewRay(vec.SFVec3f{}, vec.SFVec3f{X: 10})
	r2 := r.Extrapolate()
	if !approxVec(r2.Loc, vec.SFVec3f{X: 10}) {
		t.Fatalf("extrapolated origin wrong: %v", r2.Loc)
	}
	if r2.Dir != r.Dir {
		t.Fatal("direction should be preserved")
	}
}

func TestRay_Interpolate(t *testing.T) {
	r := NewRay(vec.SFVec3f{}, vec.SFVec3f{X: 10})
	mid := r.Interpolate()
	if !approxVec(mid, vec.SFVec3f{X: 5}) {
		t.Fatalf("midpoint should be (5,0,0), got %v", mid)
	}
}

func TestRay_ReflectRay(t *testing.T) {
	// Ray going down-right hitting a horizontal floor (normal = up)
	r := NewRay(vec.SFVec3f{Y: 1}, vec.SFVec3f{X: 1, Y: -1})
	reflected := r.ReflectRay(vec.SFVec3f{Y: 1})
	// Reflected direction should go up-right
	if !approx(reflected.Dir.X, 1/math.Sqrt(2)) {
		t.Fatalf("X should be positive: %g", reflected.Dir.X)
	}
	if reflected.Dir.Y < 0 {
		t.Fatalf("Y should be positive (reflected up): %g", reflected.Dir.Y)
	}
}

func TestRay_ApplyTransform(t *testing.T) {
	r := NewRay(vec.SFVec3f{}, vec.SFVec3f{X: 1})
	m := vec.TranslationMatrix(5, 0, 0)
	tr := r.ApplyTransform(m)
	if !approxVec(tr.Loc, vec.SFVec3f{X: 5}) {
		t.Fatalf("translated origin should be (5,0,0), got %v", tr.Loc)
	}
}

// ===========================================================================
// Plane
// ===========================================================================

func TestNewPlane(t *testing.T) {
	p := NewPlane(vec.SFVec3f{Y: 1}, 0)
	if p.Normal.Y != 1 || p.D != 0 {
		t.Fatal("constructor failed")
	}
}

func TestNewPlaneFromPoints(t *testing.T) {
	// XZ plane (normal should be +Y or -Y)
	p := NewPlaneFromPoints(
		vec.SFVec3f{X: 0, Y: 0, Z: 0},
		vec.SFVec3f{X: 1, Y: 0, Z: 0},
		vec.SFVec3f{X: 0, Y: 0, Z: 1},
	)
	// Normal should be along Y axis
	if math.Abs(p.Normal.Y) < 0.99 {
		t.Fatalf("expected Y-aligned normal, got %v", p.Normal)
	}
}

func TestPlane_GetDistance(t *testing.T) {
	// Y=0 plane (floor)
	p := NewPlane(vec.SFVec3f{Y: 1}, 0)
	d := p.GetDistance(vec.SFVec3f{Y: 5})
	if !approx(d, 5) {
		t.Fatalf("expected 5, got %g", d)
	}
	d = p.GetDistance(vec.SFVec3f{Y: -3})
	if !approx(d, -3) {
		t.Fatalf("expected -3, got %g", d)
	}
}

func TestPlane_GetDistance_OnPlane(t *testing.T) {
	p := NewPlane(vec.SFVec3f{Y: 1}, 0)
	d := p.GetDistance(vec.SFVec3f{X: 100, Y: 0, Z: -50})
	if !approx(d, 0) {
		t.Fatalf("point on plane should have distance 0, got %g", d)
	}
}

func TestPlane_IntersectRay(t *testing.T) {
	// Floor plane at Y=0
	p := NewPlane(vec.SFVec3f{Y: 1}, 0)
	// Ray from (0,10,0) going down
	r := NewRay(vec.SFVec3f{Y: 10}, vec.SFVec3f{Y: -1})
	tVal, ok := p.IntersectRay(r)
	if !ok {
		t.Fatal("should intersect")
	}
	if !approx(tVal, 10) {
		t.Fatalf("expected t=10, got %g", tVal)
	}
}

func TestPlane_IntersectRay_Parallel(t *testing.T) {
	p := NewPlane(vec.SFVec3f{Y: 1}, 0)
	// Horizontal ray (parallel to Y=0 plane)
	r := NewRay(vec.SFVec3f{Y: 5}, vec.SFVec3f{X: 1})
	_, ok := p.IntersectRay(r)
	if ok {
		t.Fatal("parallel ray should not intersect")
	}
}

func TestPlane_IntersectRay_Negative_T(t *testing.T) {
	// Plane behind the ray
	p := NewPlane(vec.SFVec3f{Y: 1}, 0)
	r := NewRay(vec.SFVec3f{Y: 10}, vec.SFVec3f{Y: 1}) // moving away from plane
	tVal, ok := p.IntersectRay(r)
	if !ok {
		t.Fatal("should compute t even if negative")
	}
	if tVal >= 0 {
		t.Fatalf("t should be negative (plane behind), got %g", tVal)
	}
}

func TestPlane_IntersectPlane(t *testing.T) {
	// XZ plane (Y=0) and XY plane (Z=0)
	p1 := NewPlane(vec.SFVec3f{Y: 1}, 0)
	p2 := NewPlane(vec.SFVec3f{Z: 1}, 0)
	ray, ok := p1.IntersectPlane(p2)
	if !ok {
		t.Fatal("should intersect")
	}
	// Intersection line should be along X axis
	if math.Abs(ray.Dir.X) < 0.99 {
		t.Fatalf("expected X-aligned direction, got %v", ray.Dir)
	}
}

func TestPlane_IntersectPlane_Parallel(t *testing.T) {
	p1 := NewPlane(vec.SFVec3f{Y: 1}, 0)
	p2 := NewPlane(vec.SFVec3f{Y: 1}, 5)
	_, ok := p1.IntersectPlane(p2)
	if ok {
		t.Fatal("parallel planes should not intersect")
	}
}

func TestPlane_Intercepts(t *testing.T) {
	// Plane: x + y + z = 6
	// normal=(1,1,1), D=-6
	p := Plane{Normal: vec.SFVec3f{X: 1, Y: 1, Z: 1}, D: -6}
	// X-intercept: X = 6 (when Y=Z=0)
	if !approx(p.XIntercept(), 6) {
		t.Fatalf("X intercept: %g", p.XIntercept())
	}
	if !approx(p.YIntercept(), 6) {
		t.Fatalf("Y intercept: %g", p.YIntercept())
	}
	if !approx(p.ZIntercept(), 6) {
		t.Fatalf("Z intercept: %g", p.ZIntercept())
	}
}

func TestPlane_Intercepts_ZeroComponent(t *testing.T) {
	p := Plane{Normal: vec.SFVec3f{X: 0, Y: 1, Z: 0}, D: -5}
	if p.XIntercept() != 0 {
		t.Fatal("X intercept with zero normal.X should be 0")
	}
	if !approx(p.YIntercept(), 5) {
		t.Fatalf("Y intercept: %g", p.YIntercept())
	}
	if p.ZIntercept() != 0 {
		t.Fatal("Z intercept with zero normal.Z should be 0")
	}
}

func TestPlane_FromPoints_Order(t *testing.T) {
	a := vec.SFVec3f{X: 0, Y: 0, Z: 0}
	b := vec.SFVec3f{X: 1, Y: 0, Z: 0}
	c := vec.SFVec3f{X: 0, Y: 1, Z: 0}
	p := NewPlaneFromPoints(a, b, c)
	// Normal should point in +Z (right-hand rule: ab x ac = +Z)
	if p.Normal.Z < 0.99 {
		t.Fatalf("expected +Z normal, got %v", p.Normal)
	}
}

// ===========================================================================
// BoundingBox
// ===========================================================================

func TestDefaultBBox(t *testing.T) {
	b := DefaultBBox()
	if !b.IsDefault() {
		t.Fatal("default bbox should be 'default' (min > max)")
	}
}

func TestNewBBox(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	if b.IsDefault() {
		t.Fatal("should not be default")
	}
}

func TestBBox_Center(t *testing.T) {
	b := NewBBox(vec.SFVec3f{}, vec.SFVec3f{X: 10, Y: 20, Z: 30})
	c := b.Center()
	if !approxVec(c, vec.SFVec3f{X: 5, Y: 10, Z: 15}) {
		t.Fatalf("center: %v", c)
	}
}

func TestBBox_Size(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -2, Z: -3}, vec.SFVec3f{X: 1, Y: 2, Z: 3})
	s := b.Size()
	if !approxVec(s, vec.SFVec3f{X: 2, Y: 4, Z: 6}) {
		t.Fatalf("size: %v", s)
	}
}

func TestBBox_Include(t *testing.T) {
	b := DefaultBBox()
	b.Include(vec.SFVec3f{X: 1, Y: 2, Z: 3})
	b.Include(vec.SFVec3f{X: -1, Y: -2, Z: -3})
	if !approxVec(b.Min, vec.SFVec3f{X: -1, Y: -2, Z: -3}) {
		t.Fatalf("min: %v", b.Min)
	}
	if !approxVec(b.Max, vec.SFVec3f{X: 1, Y: 2, Z: 3}) {
		t.Fatalf("max: %v", b.Max)
	}
}

func TestBBox_Include_SinglePoint(t *testing.T) {
	b := DefaultBBox()
	b.Include(vec.SFVec3f{X: 7, Y: 8, Z: 9})
	if !approxVec(b.Min, vec.SFVec3f{X: 7, Y: 8, Z: 9}) {
		t.Fatalf("min should equal point: %v", b.Min)
	}
	if !approxVec(b.Max, vec.SFVec3f{X: 7, Y: 8, Z: 9}) {
		t.Fatalf("max should equal point: %v", b.Max)
	}
}

func TestBBox_IncludeBox(t *testing.T) {
	a := NewBBox(vec.SFVec3f{}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	b := NewBBox(vec.SFVec3f{X: 2, Y: 2, Z: 2}, vec.SFVec3f{X: 3, Y: 3, Z: 3})
	a.IncludeBox(b)
	if !approxVec(a.Min, vec.SFVec3f{}) {
		t.Fatalf("min: %v", a.Min)
	}
	if !approxVec(a.Max, vec.SFVec3f{X: 3, Y: 3, Z: 3}) {
		t.Fatalf("max: %v", a.Max)
	}
}

func TestBBox_IsInside(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	if !b.IsInside(vec.SFVec3f{}) {
		t.Fatal("origin should be inside")
	}
	if !b.IsInside(vec.SFVec3f{X: 1, Y: 1, Z: 1}) {
		t.Fatal("corner should be inside (inclusive)")
	}
	if b.IsInside(vec.SFVec3f{X: 2}) {
		t.Fatal("outside point should not be inside")
	}
}

func TestBBox_Intersect_Hit(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	r := NewRay(vec.SFVec3f{X: -5}, vec.SFVec3f{X: 1})
	tVal, ok := b.Intersect(r)
	if !ok {
		t.Fatal("should hit the box")
	}
	if !approx(tVal, 4) {
		t.Fatalf("expected t=4, got %g", tVal)
	}
}

func TestBBox_Intersect_Miss(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	// Ray going up, missing the box
	r := NewRay(vec.SFVec3f{X: -5, Y: 5}, vec.SFVec3f{X: 1})
	_, ok := b.Intersect(r)
	if ok {
		t.Fatal("should miss")
	}
}

func TestBBox_Intersect_InsideBox(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	// Ray starting inside box
	r := NewRay(vec.SFVec3f{}, vec.SFVec3f{X: 1})
	_, ok := b.Intersect(r)
	if !ok {
		t.Fatal("ray from inside should hit")
	}
}

func TestBBox_Intersect_BehindRay(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	// Ray pointing away from box
	r := NewRay(vec.SFVec3f{X: 5}, vec.SFVec3f{X: 1})
	_, ok := b.Intersect(r)
	if ok {
		t.Fatal("box behind ray should miss")
	}
}

func TestBBox_Intersect_ParallelAxis(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	// Ray parallel to X axis but outside Y range
	r := NewRay(vec.SFVec3f{X: -5, Y: 5, Z: 0}, vec.SFVec3f{X: 1})
	_, ok := b.Intersect(r)
	if ok {
		t.Fatal("should miss - outside Y range")
	}
}

func TestBBox_Intersect_ZParallel(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	// Ray along Z, inside X/Y but outside Z
	r := NewRay(vec.SFVec3f{X: 0, Y: 0, Z: 5}, vec.SFVec3f{Z: 1})
	_, ok := b.Intersect(r)
	if ok {
		t.Fatal("should miss - moving away on Z")
	}
}

func TestBBox_Intersect_ZeroDir_Inside(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	// Ray has zero X direction but origin is inside X range, moving along Y
	r := NewRay(vec.SFVec3f{X: 0, Y: -5, Z: 0}, vec.SFVec3f{Y: 1})
	tVal, ok := b.Intersect(r)
	if !ok {
		t.Fatal("should hit - origin within X range, moving into Y range")
	}
	if !approx(tVal, 4) {
		t.Fatalf("expected t=4, got %g", tVal)
	}
}

func TestBBox_Intersect_ZeroDir_Outside(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	// Ray has zero Z direction and origin outside Z range
	r := NewRay(vec.SFVec3f{X: 0, Y: 0, Z: 5}, vec.SFVec3f{X: 1})
	_, ok := b.Intersect(r)
	if ok {
		t.Fatal("should miss - origin outside Z range with zero Z dir")
	}
}

func TestBBox_SurfaceArea(t *testing.T) {
	// 2x2x2 cube: SA = 2*(4+4+4) = 24
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	if !approx(b.SurfaceArea(), 24) {
		t.Fatalf("expected 24, got %g", b.SurfaceArea())
	}
}

func TestBBox_SurfaceArea_Zero(t *testing.T) {
	b := NewBBox(vec.SFVec3f{}, vec.SFVec3f{})
	if b.SurfaceArea() != 0 {
		t.Fatalf("zero-size box SA should be 0, got %g", b.SurfaceArea())
	}
}

func TestBBox_SurfaceArea_Flat(t *testing.T) {
	// Flat box (zero height): SA = 2*(2*0 + 0*2 + 2*2) = 8
	b := NewBBox(vec.SFVec3f{X: -1, Z: -1}, vec.SFVec3f{X: 1, Z: 1})
	if !approx(b.SurfaceArea(), 8) {
		t.Fatalf("flat box SA: %g", b.SurfaceArea())
	}
}

func TestUnion(t *testing.T) {
	a := NewBBox(vec.SFVec3f{}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{})
	u := Union(a, b)
	if !approxVec(u.Min, vec.SFVec3f{X: -1, Y: -1, Z: -1}) {
		t.Fatalf("union min: %v", u.Min)
	}
	if !approxVec(u.Max, vec.SFVec3f{X: 1, Y: 1, Z: 1}) {
		t.Fatalf("union max: %v", u.Max)
	}
}

func TestOverlap_True(t *testing.T) {
	a := NewBBox(vec.SFVec3f{}, vec.SFVec3f{X: 2, Y: 2, Z: 2})
	b := NewBBox(vec.SFVec3f{X: 1, Y: 1, Z: 1}, vec.SFVec3f{X: 3, Y: 3, Z: 3})
	if !Overlap(a, b) {
		t.Fatal("should overlap")
	}
}

func TestOverlap_False(t *testing.T) {
	a := NewBBox(vec.SFVec3f{}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	b := NewBBox(vec.SFVec3f{X: 5, Y: 5, Z: 5}, vec.SFVec3f{X: 6, Y: 6, Z: 6})
	if Overlap(a, b) {
		t.Fatal("should not overlap")
	}
}

func TestOverlap_Touching(t *testing.T) {
	a := NewBBox(vec.SFVec3f{}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	b := NewBBox(vec.SFVec3f{X: 1, Y: 1, Z: 1}, vec.SFVec3f{X: 2, Y: 2, Z: 2})
	if !Overlap(a, b) {
		t.Fatal("touching boxes should overlap (inclusive)")
	}
}

func TestTransformBox(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	m := vec.TranslationMatrix(10, 0, 0)
	tb := TransformBox(b, m)
	if !approxVec(tb.Center(), vec.SFVec3f{X: 10}) {
		t.Fatalf("translated center: %v", tb.Center())
	}
}

func TestTransformBox_Scale(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -1, Y: -1, Z: -1}, vec.SFVec3f{X: 1, Y: 1, Z: 1})
	m := vec.ScaleMatrix(2, 2, 2)
	tb := TransformBox(b, m)
	s := tb.Size()
	if !approxVec(s, vec.SFVec3f{X: 4, Y: 4, Z: 4}) {
		t.Fatalf("scaled size: %v", s)
	}
}

func TestTransformBox_Identity(t *testing.T) {
	b := NewBBox(vec.SFVec3f{X: -2, Y: -3, Z: -4}, vec.SFVec3f{X: 2, Y: 3, Z: 4})
	m := vec.Identity()
	tb := TransformBox(b, m)
	if !approxVec(tb.Min, b.Min) || !approxVec(tb.Max, b.Max) {
		t.Fatalf("identity transform changed box: %v -> %v", b, tb)
	}
}

// ===========================================================================
// Rect2D
// ===========================================================================

func TestNewRect2D(t *testing.T) {
	r := NewRect2D(10, 20, 100, 200)
	if r.X != 10 || r.Y != 20 || r.W != 100 || r.H != 200 {
		t.Fatalf("got %+v", r)
	}
}

func TestRect2D_Zero(t *testing.T) {
	r := NewRect2D(0, 0, 0, 0)
	if r.W != 0 || r.H != 0 {
		t.Fatal("should be zero")
	}
}
