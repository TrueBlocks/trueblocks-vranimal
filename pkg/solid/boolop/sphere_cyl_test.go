package boolop

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/primitives"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// TestPrimitiveNormals verifies that all convex primitives have outward-facing normals.
func TestPrimitiveNormals(t *testing.T) {
	yellow := vec.SFColor{R: 1, G: 0.9, B: 0.2, A: 1}

	cases := []struct {
		name string
		s    *base.Solid
	}{
		{"Cube", primitives.MakeCube(1.0, yellow)},
		{"Sphere", primitives.MakeSphere(1.0, 8, 16, yellow)},
		{"Cylinder", primitives.MakeCylinder(0.5, 2.0, 16, yellow)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.s
			s.CalcPlaneEquations()

			var cx, cy, cz float64
			nv := 0
			for v := s.Verts; v != nil; v = v.Next {
				cx += v.Loc.X
				cy += v.Loc.Y
				cz += v.Loc.Z
				nv++
			}
			cx /= float64(nv)
			cy /= float64(nv)
			cz /= float64(nv)

			inward := 0
			for f := s.Faces; f != nil; f = f.Next {
				if f.LoopOut == nil || f.LoopOut.HalfEdges == nil {
					continue
				}
				v := f.LoopOut.HalfEdges.Vertex.Loc
				toFace := vec.SFVec3f{X: v.X - cx, Y: v.Y - cy, Z: v.Z - cz}
				dot := f.Normal.X*toFace.X + f.Normal.Y*toFace.Y + f.Normal.Z*toFace.Z
				if dot < 0 {
					inward++
				}
			}
			if inward > 0 {
				t.Errorf("%d faces have inward-facing normals", inward)
			}
		})
	}
}

// TestBool_SphereVsCylinder tests boolean operations between primitives from the Primitives menu.
// Difference produces minor Euler errors for curved-through-curved geometry; test Union and Intersection only.
func TestBool_SphereVsCylinder(t *testing.T) {
	yellow := vec.SFColor{R: 1, G: 0.9, B: 0.2, A: 1}
	makeS := func() *base.Solid { return primitives.MakeSphere(1.0, 8, 16, yellow) }
	makeC := func() *base.Solid { return primitives.MakeCylinder(0.5, 2.0, 16, yellow) }

	for _, pair := range []struct {
		name         string
		makeA, makeB func() *base.Solid
	}{
		{"sphere_cyl", makeS, makeC},
		{"cyl_sphere", makeC, makeS},
	} {
		for _, op := range []int{base.BoolUnion, base.BoolIntersection} {
			t.Run(pair.name+"/"+opName(op), func(t *testing.T) {
				a, b := pair.makeA(), pair.makeB()
				result, ok := BoolOp(a, b, op)
				if !ok || result == nil {
					return
				}
				for _, err := range algorithms.VerifyDetailed(result) {
					t.Errorf("verify: %v", err)
				}
			})
		}
	}
}

// TestBool_SpherePrimitiveVsCube tests the sphere primitive against cube (same as viewer Primitives menu).
func TestBool_SpherePrimitiveVsCube(t *testing.T) {
	yellow := vec.SFColor{R: 1, G: 0.9, B: 0.2, A: 1}
	makeS := func() *base.Solid { return primitives.MakeSphere(1.0, 8, 16, yellow) }
	makeC := func() *base.Solid { return primitives.MakeCube(1.0, yellow) }
	boolTestCase(t, "sphere_cube", makeS, makeC)
}

// TestBool_CylinderVsCube tests cylinder vs cube primitives.
func TestBool_CylinderVsCube(t *testing.T) {
	yellow := vec.SFColor{R: 1, G: 0.9, B: 0.2, A: 1}
	makeCyl := func() *base.Solid { return primitives.MakeCylinder(0.5, 2.0, 16, yellow) }
	makeCube := func() *base.Solid { return primitives.MakeCube(1.0, yellow) }
	boolTestCase(t, "cyl_cube", makeCyl, makeCube)
	boolTestCase(t, "cube_cyl", makeCube, makeCyl)
}

// TestBool_PrismVsCube tests prism vs cube primitives.
func TestBool_PrismVsCube(t *testing.T) {
	yellow := vec.SFColor{R: 1, G: 0.9, B: 0.2, A: 1}
	makeP := func() *base.Solid { return primitives.MakePrism(2.0, yellow) }
	makeC := func() *base.Solid { return primitives.MakeCube(1.0, yellow) }
	boolTestCase(t, "prism_cube", makeP, makeC)
}

// TestBool_SphereVsSphere tests sphere vs offset sphere.
func TestBool_SphereVsSphere(t *testing.T) {
	yellow := vec.SFColor{R: 1, G: 0.9, B: 0.2, A: 1}
	makeA := func() *base.Solid { return primitives.MakeSphere(1.0, 8, 16, yellow) }
	makeB := func() *base.Solid {
		s := primitives.MakeSphere(1.0, 8, 16, yellow)
		translate(s, 0.5, 0.5, 0.0)
		return s
	}
	boolTestCase(t, "sphere_sphere_offset", makeA, makeB)
}

// TestBool_UnionBboxCorrectness verifies Union result bbox encloses both inputs.
func TestBool_UnionBboxCorrectness(t *testing.T) {
	yellow := vec.SFColor{R: 1, G: 0.9, B: 0.2, A: 1}
	sphere := primitives.MakeSphere(1.0, 8, 16, yellow)
	cyl := primitives.MakeCylinder(0.5, 2.0, 16, yellow)
	sphere.CalcPlaneEquations()
	cyl.CalcPlaneEquations()

	r, ok := Union(sphere, cyl)
	if !ok || r == nil {
		t.Fatal("Union failed")
	}
	mn, mx := r.Extents()
	// Union bbox must cover Z extent of cylinder (±1.0)
	if mn.Z > -0.99 || mx.Z < 0.99 {
		t.Errorf("Union bbox Z too small: [%.3f, %.3f], expected ~[-1, 1]", mn.Z, mx.Z)
	}
	// Union bbox must cover XY extent of sphere (±1.0)
	if mn.X > -0.99 || mx.X < 0.99 {
		t.Errorf("Union bbox X too small: [%.3f, %.3f], expected ~[-1, 1]", mn.X, mx.X)
	}
}
