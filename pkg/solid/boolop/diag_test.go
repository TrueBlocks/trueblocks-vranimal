package boolop

import (
	"fmt"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/primitives"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func solidStats(s *base.Solid) (nf, nv, ne int) {
	for f := s.Faces; f != nil; f = f.Next {
		nf++
	}
	for v := s.Verts; v != nil; v = v.Next {
		nv++
	}
	for e := s.Edges; e != nil; e = e.Next {
		ne++
	}
	return
}

func TestDiag_PartialOverlapUnion(t *testing.T) {
	red := vec.SFColor{R: 1, A: 1}
	green := vec.SFColor{G: 1, A: 1}

	a := primitives.MakeCube(1.0, red)
	b := primitives.MakeCube(1.0, green)
	b.TransformGeometry(vec.TranslationMatrix(0.5, 0.5, 0))

	// Test via BoolOp (the real API) for correctness
	fmt.Println("=== VIA BoolOp ===")
	for _, tc := range []struct {
		op   int
		name string
	}{
		{base.BoolUnion, "Union"},
		{base.BoolIntersection, "Intersection"},
		{base.BoolDifference, "Difference"},
	} {
		result, ok := BoolOp(a.Copy(), b.Copy(), tc.op)
		if !ok || result == nil {
			fmt.Printf("BoolOp %s: failed or nil\n", tc.name)
			continue
		}
		nf, nv, ne := solidStats(result)
		mn, mx := result.Extents()
		errs := algorithms.VerifyDetailed(result)
		fmt.Printf("BoolOp %s: F=%d V=%d E=%d bbox [%.4f,%.4f,%.4f]-[%.4f,%.4f,%.4f] euler_errs=%d\n",
			tc.name, nf, nv, ne, mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z, len(errs))
		for _, e := range errs {
			fmt.Printf("  %v\n", e)
		}
	}

	// Through case via BoolOp
	fmt.Println("\n=== THROUGH VIA BoolOp ===")
	throughA := primitives.MakeCube(1.0, red)
	throughB := primitives.MakeCube(0.5, green)
	throughB.TransformGeometry(vec.ScaleMatrix(1, 1, 4))
	throughB.TransformGeometry(vec.TranslationMatrix(0.25, 0.25, -0.15))
	for _, tc := range []struct {
		op   int
		name string
	}{
		{base.BoolUnion, "Union"},
		{base.BoolIntersection, "Intersection"},
		{base.BoolDifference, "Difference"},
	} {
		result, ok := BoolOp(throughA.Copy(), throughB.Copy(), tc.op)
		if !ok || result == nil {
			fmt.Printf("Through %s: failed or nil\n", tc.name)
			continue
		}
		nf, nv, ne := solidStats(result)
		mn, mx := result.Extents()
		errs := algorithms.VerifyDetailed(result)
		fmt.Printf("Through %s: F=%d V=%d E=%d bbox [%.4f,%.4f,%.4f]-[%.4f,%.4f,%.4f] euler_errs=%d\n",
			tc.name, nf, nv, ne, mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z, len(errs))
		for _, e := range errs {
			fmt.Printf("  %v\n", e)
		}
	}

	// Sphere vs cube via BoolOp
	fmt.Println("\n=== SPHERE VS CUBE VIA BoolOp ===")
	cubeA := primitives.MakeCube(1.0, red)
	sphereB := primitives.MakeSphere(0.5, 4, 8, green)
	sphereB.TransformGeometry(vec.TranslationMatrix(-0.25, -0.25, -0.25))
	for _, tc := range []struct {
		op   int
		name string
	}{
		{base.BoolUnion, "Union"},
		{base.BoolIntersection, "Intersection"},
		{base.BoolDifference, "Difference"},
	} {
		result, ok := BoolOp(cubeA.Copy(), sphereB.Copy(), tc.op)
		if !ok || result == nil {
			fmt.Printf("Sphere %s: failed or nil\n", tc.name)
			continue
		}
		nf, nv, ne := solidStats(result)
		mn, mx := result.Extents()
		errs := algorithms.VerifyDetailed(result)
		fmt.Printf("Sphere %s: F=%d V=%d E=%d bbox [%.4f,%.4f,%.4f]-[%.4f,%.4f,%.4f] euler_errs=%d\n",
			tc.name, nf, nv, ne, mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z, len(errs))
		for _, e := range errs {
			fmt.Printf("  %v\n", e)
		}
	}
}
