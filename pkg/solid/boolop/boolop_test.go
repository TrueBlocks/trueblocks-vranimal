package boolop

import (
	"fmt"
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/primitives"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

var (
	red   = vec.SFColor{R: 1, G: 0, B: 0, A: 1}
	green = vec.SFColor{R: 0, G: 1, B: 0, A: 1}
)

var allOps = []int{base.BoolUnion, base.BoolIntersection, base.BoolDifference}

func opName(op int) string {
	switch op {
	case base.BoolUnion:
		return "Union"
	case base.BoolIntersection:
		return "Intersection"
	case base.BoolDifference:
		return "Difference"
	default:
		return "Unknown"
	}
}

func translate(s *base.Solid, x, y, z float64) {
	s.TransformGeometry(vec.TranslationMatrix(x, y, z))
}

func scale(s *base.Solid, sx, sy, sz float64) {
	s.TransformGeometry(vec.ScaleMatrix(sx, sy, sz))
}

func rotate(s *base.Solid, degrees float64, axis vec.SFVec3f) {
	radians := degrees * math.Pi / 180.0
	rot := vec.SFRotation{X: axis.X, Y: axis.Y, Z: axis.Z, W: radians}
	s.TransformGeometry(vec.RotationMatrix(rot))
}

func rotateCenter(s *base.Solid, degrees float64, axis vec.SFVec3f) {
	mn, mx := s.Extents()
	cx := (mn.X + mx.X) / 2
	cy := (mn.Y + mx.Y) / 2
	cz := (mn.Z + mx.Z) / 2
	translate(s, -cx, -cy, -cz)
	rotate(s, degrees, axis)
	translate(s, cx, cy, cz)
}

func boolTestCase(t *testing.T, name string, makeA, makeB func() *base.Solid) {
	t.Helper()
	for _, op := range allOps {
		t.Run(fmt.Sprintf("%s/%s", name, opName(op)), func(t *testing.T) {
			a, b := makeA(), makeB()
			if errs := algorithms.VerifyDetailed(a); len(errs) > 0 {
				for _, err := range errs {
					t.Errorf("input A invalid: %v", err)
				}
				return
			}
			if errs := algorithms.VerifyDetailed(b); len(errs) > 0 {
				for _, err := range errs {
					t.Errorf("input B invalid: %v", err)
				}
				return
			}
			result, ok := BoolOp(a, b, op)
			if !ok {
				return
			}
			if result == nil {
				return
			}
			errs := algorithms.VerifyDetailed(result)
			for _, err := range errs {
				t.Errorf("verify: %v", err)
			}
		})
	}
}

func TestBoolOp_Convenience(t *testing.T) {
	a := primitives.MakeCube(1.0, red)
	b := primitives.MakeCube(0.5, green)
	translate(b, 0.25, 0.25, 0.25)
	r, ok := Union(a, b)
	if !ok || r == nil {
		t.Fatal("Union failed")
	}
	r, ok = Intersection(a, b)
	if !ok || r == nil {
		t.Fatal("Intersection failed")
	}
	r, ok = Difference(a, b)
	if !ok || r == nil {
		t.Fatal("Difference failed")
	}
}

func TestBoolOp_DisjointUnion(t *testing.T) {
	a := primitives.MakeCube(0.5, red)
	b := primitives.MakeCube(0.5, green)
	translate(b, 5, 0, 0)
	r, ok := Union(a, b)
	if !ok || r == nil {
		t.Fatal("disjoint union failed")
	}
}

func TestBoolOp_ContainedIntersection(t *testing.T) {
	a := primitives.MakeCube(1.0, red)
	b := primitives.MakeCube(0.3, green)
	r, ok := Intersection(a, b)
	if !ok || r == nil {
		t.Fatal("contained intersection failed")
	}
}

func TestBoolOp_DisjointIntersection(t *testing.T) {
	a := primitives.MakeCube(0.5, red)
	b := primitives.MakeCube(0.5, green)
	translate(b, 5, 0, 0)
	_, ok := Intersection(a, b)
	if ok {
		t.Fatal("disjoint intersection should return false")
	}
}

func TestBoolOp_DisjointDifference(t *testing.T) {
	a := primitives.MakeCube(0.5, red)
	b := primitives.MakeCube(0.5, green)
	translate(b, 5, 0, 0)
	r, ok := Difference(a, b)
	if !ok || r == nil {
		t.Fatal("disjoint difference failed")
	}
}

func TestBool_Group0_CubeVsScaledCube(t *testing.T) {
	type config struct {
		tx, ty, tz float64
		sx, sy, sz float64
		desc       string
	}
	cases := []config{
		{0.25, 0.25, -1.00, 1, 1, 2, "disjoint"},
		{0.25, 0.25, -0.50, 1, 1, 2, "face_touching"},
		{0.25, 0.25, -0.25, 1, 1, 2, "partial_penetration"},
		{0.25, 0.25, 0.00, 1, 1, 1, "interior_one_face"},
		{0.25, 0.25, 0.25, 1, 1, 1, "fully_contained"},
		{0.25, 0.25, 0.00, 1, 1, 2, "interior_two_faces"},
		{0.25, 0.25, -0.15, 1, 1, 4, "through"},
	}
	for i, c := range cases {
		makeA := func() *base.Solid { return primitives.MakeCube(1.0, red) }
		cfg := c
		makeB := func() *base.Solid {
			s := primitives.MakeCube(0.5, green)
			scale(s, cfg.sx, cfg.sy, cfg.sz)
			translate(s, cfg.tx, cfg.ty, cfg.tz)
			return s
		}
		boolTestCase(t, fmt.Sprintf("Case%d_%s", i, c.desc), makeA, makeB)
	}
}

func TestBool_Group1_CubeVsRotatedScaledCube(t *testing.T) {
	type config struct {
		tx, ty, tz float64
		sx, sy, sz float64
		desc       string
	}
	cases := []config{
		{0.25, 0.25, -1.00, 1, 1, 2, "disjoint"},
		{0.25, 0.25, -0.50, 1, 1, 2, "face_touching"},
		{0.25, 0.25, -0.25, 1, 1, 2, "partial_penetration"},
		{0.25, 0.25, 0.00, 1, 1, 1, "interior_one_face"},
		{0.25, 0.25, 0.25, 1, 1, 1, "fully_contained"},
		{0.25, 0.25, 0.00, 1, 1, 2, "interior_two_faces"},
		{0.25, 0.25, -0.15, 1, 1, 4, "through"},
	}
	for i, c := range cases {
		makeA := func() *base.Solid { return primitives.MakeCube(1.0, red) }
		cfg := c
		makeB := func() *base.Solid {
			s := primitives.MakeCube(0.5, green)
			scale(s, cfg.sx, cfg.sy, cfg.sz)
			rotate(s, 45.0, vec.ZAxis)
			translate(s, cfg.tx, cfg.ty, cfg.tz)
			return s
		}
		boolTestCase(t, fmt.Sprintf("Case%d_%s", i, c.desc), makeA, makeB)
	}
}

func TestBool_Group2_CubeVsRotCenterScaledCube(t *testing.T) {
	type config struct {
		tx, ty, tz float64
		sx, sy, sz float64
		desc       string
	}
	cases := []config{
		{0.25, 0.25, -1.00, 1, 1, 2, "disjoint"},
		{0.25, 0.25, -0.50, 1, 1, 2, "face_touching"},
		{0.25, 0.25, -0.25, 1, 1, 2, "partial_penetration"},
		{0.25, 0.25, 0.00, 1, 1, 1, "interior_one_face"},
		{0.25, 0.25, 0.25, 1, 1, 1, "fully_contained"},
		{0.25, 0.25, 0.00, 1, 1, 2, "interior_two_faces"},
		{0.25, 0.25, -0.15, 1, 1, 4, "through"},
	}
	for i, c := range cases {
		makeA := func() *base.Solid { return primitives.MakeCube(1.0, red) }
		cfg := c
		makeB := func() *base.Solid {
			s := primitives.MakeCube(0.5, green)
			scale(s, cfg.sx, cfg.sy, cfg.sz)
			rotateCenter(s, 45.0, vec.ZAxis)
			translate(s, cfg.tx, cfg.ty, cfg.tz)
			return s
		}
		boolTestCase(t, fmt.Sprintf("Case%d_%s", i, c.desc), makeA, makeB)
	}
}

func TestBool_Group5_CubeVsSphere(t *testing.T) {
	type config struct {
		radius           float64
		latSegs, lonSegs int
		desc             string
	}
	cases := []config{
		{1.0, 4, 4, "coarse_sphere"},
		{1.0, 9, 9, "medium_sphere"},
		{1.0, 16, 16, "fine_sphere"},
		{1.0, 25, 25, "highres_sphere"},
		{1.0, 20, 20, "large_sphere_containing"},
		{1.25, 20, 20, "larger_sphere"},
		{1.50, 20, 20, "largest_sphere"},
	}
	for i, c := range cases {
		makeA := func() *base.Solid { return primitives.MakeCube(1.0, red) }
		cfg := c
		makeB := func() *base.Solid {
			s := primitives.MakeSphere(cfg.radius, cfg.latSegs, cfg.lonSegs, green)
			translate(s, -0.25, -0.25, -0.25)
			return s
		}
		boolTestCase(t, fmt.Sprintf("Case%d_%s", i, c.desc), makeA, makeB)
	}
}

func TestBool_Group6_CubeVsSameSizeCube(t *testing.T) {
	translations := [][3]float64{
		{0.0, 0.0, 0.0},
		{0.5, 0.5, 0.0},
		{1.0, 1.0, 0.0},
		{0.5, 0.0, 0.0},
		{1.0, 0.0, 0.0},
		{2.0, 0.0, 0.0},
		{0.3, 0.3, 0.0},
	}
	descs := []string{
		"coincident", "partial_face_overlap", "edge_on_edge",
		"half_face", "face_on_face", "disjoint", "slight_twist",
	}
	for i, tr := range translations {
		makeA := func() *base.Solid { return primitives.MakeCube(1.0, red) }
		off := tr
		idx := i
		desc := descs[i]
		makeB := func() *base.Solid {
			s := primitives.MakeCube(1.0, green)
			if idx == 6 {
				rotateCenter(s, 15, vec.ZAxis)
			}
			translate(s, off[0], off[1], off[2])
			return s
		}
		if desc == "disjoint" {
			// Disjoint union produces a multi-shell solid; the single-shell
			// Euler checker cannot validate it. Test the other ops only.
			t.Run(fmt.Sprintf("Case%d_%s", i, desc), func(t *testing.T) {
				for _, op := range []int{base.BoolIntersection, base.BoolDifference} {
					t.Run(opName(op), func(t *testing.T) {
						a, b := makeA(), makeB()
						result, ok := BoolOp(a, b, op)
						if !ok || result == nil {
							return
						}
						for _, err := range algorithms.VerifyDetailed(result) {
							t.Errorf("verify: %v", err)
						}
					})
				}
			})
			continue
		}
		boolTestCase(t, fmt.Sprintf("Case%d_%s", i, desc), makeA, makeB)
	}
}

func TestBool_Precision_TinyCubes(t *testing.T) {
	makeA := func() *base.Solid { return primitives.MakeCube(0.001, red) }
	makeB := func() *base.Solid {
		s := primitives.MakeCube(0.0005, green)
		translate(s, 0.00025, 0.00025, 0.00025)
		return s
	}
	boolTestCase(t, "tiny_cubes", makeA, makeB)
}
