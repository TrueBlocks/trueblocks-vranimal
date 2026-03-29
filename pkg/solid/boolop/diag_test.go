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

func tracedBoolOpCore(t *testing.T, a, b *base.Solid, op int, perturb bool, label string) (*base.Solid, bool, bool) {
	workA := a.Copy()
	workB := b.Copy()
	if perturb {
		const eps = 1e-4
		workB.TransformGeometry(vec.TranslationMatrix(eps, eps, eps))
	}
	workA.CalcPlaneEquations()
	workB.CalcPlaneEquations()
	workA.SetFaceMarks(base.UNKNOWN)
	workA.SetVertexMarks(base.UNKNOWN)
	workB.SetFaceMarks(base.UNKNOWN)
	workB.SetVertexMarks(base.UNKNOWN)

	fa, va, ea := solidStats(workA)
	fb, vb, eb := solidStats(workB)
	fmt.Printf("[%s] INPUT: A(F=%d V=%d E=%d) B(F=%d V=%d E=%d) perturb=%v\n", label, fa, va, ea, fb, vb, eb, perturb)

	br := NewBoolopRecord()
	br.Reset(workA, workB, op)
	br.Generate()

	fa, va, ea = solidStats(workA)
	fb, vb, eb = solidStats(workB)
	fmt.Printf("[%s] AFTER GENERATE: A(F=%d V=%d E=%d) B(F=%d V=%d E=%d)\n", label, fa, va, ea, fb, vb, eb)
	fmt.Printf("[%s]   VF-A=%d VF-B=%d VV=%d NoVV=%v Quit=%v\n",
		label, len(br.VertsA), len(br.VertsB), len(br.VertsV), br.NoVV, br.Quit)

	if br.Quit {
		fmt.Printf("[%s] QUIT after Generate\n", label)
		return nil, false, false
	}

	classified := br.Classify()
	fmt.Printf("[%s] AFTER CLASSIFY: classified=%v EdgesA=%d EdgesB=%d NoVV=%v Quit=%v\n",
		label, classified, len(br.EdgesA), len(br.EdgesB), br.NoVV, br.Quit)

	if !classified {
		if br.Quit {
			return nil, false, false
		}
		ok := br.LastDitch()
		fmt.Printf("[%s] LASTDITCH: ok=%v\n", label, ok)
		if ok && br.Result != nil {
			rf, rv, re := solidStats(br.Result)
			mn, mx := br.Result.Extents()
			fmt.Printf("[%s] LASTDITCH RESULT: F=%d V=%d E=%d bbox [%.4f,%.4f,%.4f]-[%.4f,%.4f,%.4f]\n",
				label, rf, rv, re, mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z)
		}
		if ok {
			br.Complete()
		}
		return br.Result, ok, false
	}

	br.Connect()
	fmt.Printf("[%s] AFTER CONNECT: FacesA=%d FacesB=%d VertsA=%d VertsB=%d NoVV=%v Quit=%v\n",
		label, len(br.FacesA), len(br.FacesB), len(br.VertsA), len(br.VertsB), br.NoVV, br.Quit)

	// Print face pair details
	for i := range br.FacesA {
		fA := br.FacesA[i]
		fB := br.FacesB[i]
		nLoopsA := len(fA.Loops)
		nLoopsB := len(fB.Loops)
		vertsA := countLoopVerts(fA)
		vertsB := countLoopVerts(fB)
		fmt.Printf("[%s]   pair %d: A(loops=%d outerVerts=%d) B(loops=%d outerVerts=%d)\n",
			label, i, nLoopsA, vertsA, nLoopsB, vertsB)
		// Print each loop's verts
		for li, l := range fA.Loops {
			isOuter := l == fA.LoopOut
			he := l.GetFirstHe()
			if he == nil {
				continue
			}
			fmt.Printf("[%s]     A-loop%d (outer=%v): ", label, li, isOuter)
			start := he
			for {
				fmt.Printf("(%.2f,%.2f,%.2f) ", he.Vertex.Loc.X, he.Vertex.Loc.Y, he.Vertex.Loc.Z)
				he = he.Next
				if he == start {
					break
				}
			}
			fmt.Println()
		}
		for li, l := range fB.Loops {
			isOuter := l == fB.LoopOut
			he := l.GetFirstHe()
			if he == nil {
				continue
			}
			fmt.Printf("[%s]     B-loop%d (outer=%v): ", label, li, isOuter)
			start := he
			for {
				fmt.Printf("(%.2f,%.2f,%.2f) ", he.Vertex.Loc.X, he.Vertex.Loc.Y, he.Vertex.Loc.Z)
				he = he.Next
				if he == start {
					break
				}
			}
			fmt.Println()
		}
	}

	if br.Quit {
		return nil, false, false
	}

	// Check degenerate
	if !perturb {
		for i := range br.FacesA {
			if countLoopVerts(br.FacesA[i]) < 3 {
				fmt.Printf("[%s] DEGENERATE: FacesA[%d] has < 3 verts\n", label, i)
				return nil, false, true
			}
		}
	}

	br.Finish()
	fmt.Printf("[%s] AFTER FINISH: Quit=%v\n", label, br.Quit)
	if br.Quit {
		return nil, false, false
	}
	br.Complete()
	if br.Result != nil {
		rf, rv, re := solidStats(br.Result)
		mn, mx := br.Result.Extents()
		fmt.Printf("[%s] FINAL RESULT: F=%d V=%d E=%d bbox [%.4f,%.4f,%.4f]-[%.4f,%.4f,%.4f]\n",
			label, rf, rv, re, mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z)
		errs := algorithms.VerifyDetailed(br.Result)
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Printf("[%s]   EULER ERR: %v\n", label, e)
			}
		}
	}
	return br.Result, br.Result != nil, false
}

func faceBBox(f *base.Face) (mn, mx vec.SFVec3f) {
	first := true
	for _, l := range f.Loops {
		he := l.GetFirstHe()
		if he == nil {
			continue
		}
		start := he
		for {
			v := he.Vertex.Loc
			if first {
				mn, mx = v, v
				first = false
			} else {
				if v.X < mn.X {
					mn.X = v.X
				}
				if v.Y < mn.Y {
					mn.Y = v.Y
				}
				if v.Z < mn.Z {
					mn.Z = v.Z
				}
				if v.X > mx.X {
					mx.X = v.X
				}
				if v.Y > mx.Y {
					mx.Y = v.Y
				}
				if v.Z > mx.Z {
					mx.Z = v.Z
				}
			}
			he = he.Next
			if he == start {
				break
			}
		}
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
