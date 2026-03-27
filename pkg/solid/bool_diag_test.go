package solid

import (
	"fmt"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// TestBoolDiagnostic_Through traces every step of the bool pipeline for the
// simplest failing case (Group0/Case6 "through" Union) and validates after
// each operation, dying at the first invalid solid.
func TestBoolDiagnostic_Through(t *testing.T) {
	red := vec.SFColor{R: 1, G: 0, B: 0}
	green := vec.SFColor{R: 0, G: 1, B: 0}

	a := MakeCube(1.0, red)
	b := MakeCube(0.5, green)
	scale(b, 1, 1, 4)
	translate(b, 0.25, 0.25, -0.15)

	workA := a.Copy()
	workB := b.Copy()
	workA.CalcPlaneEquations()
	workB.CalcPlaneEquations()
	workA.SetFaceMarks(UNKNOWN)
	workA.SetVertexMarks(UNKNOWN)
	workB.SetFaceMarks(UNKNOWN)
	workB.SetVertexMarks(UNKNOWN)

	check := func(label string, solids ...*Solid) {
		for i, s := range solids {
			name := fmt.Sprintf("%c", 'A'+i)
			if len(solids) == 1 {
				name = "Result"
			}
			errs := s.VerifyDetailed()
			stats := solidStats(s)
			if len(errs) > 0 {
				t.Logf("INVALID after %s — %s: %s", label, name, stats)
				for _, e := range errs {
					t.Logf("  -> %v", e)
				}
				t.Fatalf("First invalid solid detected after: %s (%s)", label, name)
			}
			t.Logf("  ok after %-45s %s: %s", label, name, stats)
		}
	}

	br := NewBoolopRecord()
	br.Reset(workA, workB, BoolUnion)

	check("initial", workA, workB)

	br.Generate()
	if br.Quit {
		t.Fatal("Generate quit")
	}
	check("Generate", workA, workB)
	t.Logf("  VertsV=%d VertsA=%d VertsB=%d", len(br.VertsV), len(br.VertsA), len(br.VertsB))

	ok := br.Classify()
	if !ok {
		t.Fatal("Classify returned false")
	}
	check("Classify", workA, workB)
	t.Logf("  EdgesA=%d EdgesB=%d", len(br.EdgesA), len(br.EdgesB))

	br.Connect()
	if br.Quit {
		t.Fatal("Connect quit")
	}
	check("Connect", workA, workB)
	t.Logf("  FacesA=%d FacesB=%d NoVV=%v", len(br.FacesA), len(br.FacesB), br.NoVV)

	// --- Now replicate Finish step by step (Union path) ---
	nFacesA := len(br.FacesA)
	t.Logf("=== Finish (Union) with %d face pairs ===", nFacesA)

	// Step 1: Lmfkrh — create mirrors (temporarily invalidates Euler)
	mirrorsA := make([]*Face, nFacesA)
	mirrorsB := make([]*Face, nFacesA)
	for i := 0; i < nFacesA; i++ {
		fA := br.FacesA[i]
		if len(fA.Loops) >= 2 && fA.Loops[1] == fA.LoopOut {
			fA.LoopOut = fA.Loops[0]
		}
		sl := fA.GetSecondLoop()
		if sl != nil {
			mirrorsA[i] = Lmfkrh(sl)
		}
		t.Logf("  after Lmfkrh(A[%d]): A %s (temp invalid expected)", i, solidStats(workA))

		fB := br.FacesB[i]
		if len(fB.Loops) >= 2 && fB.Loops[1] == fB.LoopOut {
			fB.LoopOut = fB.Loops[0]
		}
		sl = fB.GetSecondLoop()
		if sl != nil {
			mirrorsB[i] = Lmfkrh(sl)
		}
		t.Logf("  after Lmfkrh(B[%d]): B %s (temp invalid expected)", i, solidStats(workB))

		if br.NoVV {
			// SKIP NoVV swap
			t.Logf("  NoVV swap SKIPPED for pair %d", i)
		}
	}

	t.Log("--- Lmfkrh complete. Faces/mirrors setup: ---")
	for i := 0; i < nFacesA; i++ {
		t.Logf("  pair[%d]: FacesA loops=%d, mirrorsA loops=%d, FacesB loops=%d, mirrorsB loops=%d",
			i, br.FacesA[i].NLoops(), nilSafeLoops(mirrorsA[i]), br.FacesB[i].NLoops(), nilSafeLoops(mirrorsB[i]))
	}

	// Step 2: MoveFace
	result := NewSolid()
	workA.ClearFaceMarks2()
	workB.ClearFaceMarks2()
	for i := 0; i < nFacesA; i++ {
		workA.MoveFace(br.FacesA[i], result)
		t.Logf("  after MoveFace(FacesA[%d]): result %s, A %s", i, solidStats(result), solidStats(workA))

		workB.MoveFace(br.FacesB[i], result)
		t.Logf("  after MoveFace(FacesB[%d]): result %s, B %s", i, solidStats(result), solidStats(workB))
	}

	// Step 3: Cleanup
	result.Cleanup()
	t.Logf("  after Cleanup: result %s", solidStats(result))
	checkNonFatal(t, "Cleanup", result)

	// Step 4: Lkfmrh + LoopGlue (with param swap fix)
	for i := 0; i < nFacesA; i++ {
		t.Logf("  before Lkfmrh[%d]: FacesA[%d] loops=%d, FacesB[%d] loops=%d",
			i, i, br.FacesA[i].NLoops(), i, br.FacesB[i].NLoops())

		Lkfmrh(br.FacesB[i], br.FacesA[i]) // SWAPPED: kill B, keep A (matches Union fix)
		t.Logf("  after  Lkfmrh[%d]: result %s", i, solidStats(result))
		checkNonFatal(t, fmt.Sprintf("Lkfmrh[%d]", i), result)

		t.Logf("  before LoopGlue[%d]: face loops=%d", i, br.FacesB[i].NLoops())
		result.LoopGlue(br.FacesB[i])
		t.Logf("  after  LoopGlue[%d]: result %s", i, solidStats(result))
		checkNonFatal(t, fmt.Sprintf("LoopGlue[%d]", i), result)
	}

	// Step 5: final cleanup
	workA.Cleanup()
	workB.Cleanup()
	if result != nil {
		for f := result.Faces; f != nil; f = f.Next {
			if f.LoopOut == nil && len(f.Loops) > 0 {
				f.LoopOut = f.Loops[0]
			}
		}
	}

	errs := result.VerifyDetailed()
	t.Logf("=== Final result: %s ===", solidStats(result))
	for _, e := range errs {
		t.Errorf("verify: %v", e)
	}
	if len(errs) == 0 {
		t.Log("PASS — result is valid!")
	}
}

func solidStats(s *Solid) string {
	f, v, e, h := 0, 0, 0, 0
	for face := s.Faces; face != nil; face = face.Next {
		f++
		h += face.NLoops() - 1
	}
	for vert := s.Verts; vert != nil; vert = vert.Next {
		v++
	}
	for edge := s.Edges; edge != nil; edge = edge.Next {
		e++
	}
	diff := (f + v - 2) - (e + h)
	return fmt.Sprintf("F=%d V=%d E=%d H=%d euler_diff=%+d", f, v, e, h, diff)
}

func checkNonFatal(t *testing.T, label string, s *Solid) {
	t.Helper()
	errs := s.VerifyDetailed()
	if len(errs) > 0 {
		for _, e := range errs {
			t.Logf("  !! %s: %v", label, e)
		}
	}
}

func nilSafeLoops(f *Face) int {
	if f == nil {
		return -1
	}
	return f.NLoops()
}
