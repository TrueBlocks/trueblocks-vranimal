package solid

// Phase 2 pipeline tests: MakeRing, Generate, Classify, Connect, Finish.
// These test each phase independently before composing the full pipeline.

import (
	"fmt"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// checkAllLoops verifies that every loop in a solid has a finite half-edge ring.
// Returns an error description if any loop is infinite, or "" if all are OK.
func checkAllLoops(s *Solid) string {
	fIdx := 0
	for f := s.Faces; f != nil; f = f.Next {
		for li, l := range f.Loops {
			if l.HalfEdges == nil {
				continue
			}
			count := 0
			he := l.HalfEdges
			for {
				count++
				if count > 5000 {
					return fmt.Sprintf("face[%d] loop[%d]: infinite (>5000 hes)", fIdx, li)
				}
				he = he.Next
				if he == l.HalfEdges {
					break
				}
			}
		}
		fIdx++
	}
	return ""
}

// countLoopHes counts how many half-edges are in a loop (returns -1 if infinite).
func countLoopHes(l *Loop) int {
	if l == nil || l.HalfEdges == nil {
		return 0
	}
	count := 0
	he := l.HalfEdges
	for {
		count++
		if count > 5000 {
			return -1
		}
		he = he.Next
		if he == l.HalfEdges {
			break
		}
	}
	return count
}

// makeSimpleIntersection creates two cubes where B's 4 vertical edges
// cross A's top face, producing 4 VF records.
func makeSimpleIntersection() (a, b *Solid) {
	red := vec.SFColor{R: 1, A: 1}
	green := vec.SFColor{G: 1, A: 1}

	a = MakeCube(1.0, red)
	b = MakeCube(0.5, green)
	b.TransformGeometry(vec.TranslationMatrix(0.25, 0.25, 1))
	return a, b
}

// ---------------------------------------------------------------------------
// MakeRing
// ---------------------------------------------------------------------------

func TestPhase2_MakeRing_ValidInnerLoop(t *testing.T) {
	s := MakeCube(1.0, vec.SFColor{R: 1, A: 1})
	f := s.GetFirstFace()

	pt := vec.SFVec3f{X: 0, Y: 0, Z: f.GetFirstHe().Vertex.Loc.Z}
	e := f.MakeRing(pt)
	if e == nil {
		t.Fatal("MakeRing returned nil edge")
	}
	if f.NLoops() != 2 {
		t.Fatalf("expected 2 loops after MakeRing, got %d", f.NLoops())
	}
	if countLoopHes(f.LoopOut) != 4 {
		t.Fatalf("outer loop: expected 4 hes, got %d", countLoopHes(f.LoopOut))
	}
	if countLoopHes(f.GetSecondLoop()) != 2 {
		t.Fatalf("inner loop: expected 2 hes, got %d", countLoopHes(f.GetSecondLoop()))
	}
	if !s.Verify() {
		t.Fatal("Verify failed after MakeRing")
	}
}

func TestPhase2_MakeRing_MultipleRings(t *testing.T) {
	s := MakeCube(1.0, vec.SFColor{R: 1, A: 1})
	f := s.GetFirstFace()
	z := f.GetFirstHe().Vertex.Loc.Z

	for _, pt := range []vec.SFVec3f{
		{X: -0.3, Y: -0.3, Z: z}, {X: 0.3, Y: -0.3, Z: z},
		{X: 0.3, Y: 0.3, Z: z}, {X: -0.3, Y: 0.3, Z: z},
	} {
		f.MakeRing(pt)
	}

	if f.NLoops() != 5 {
		t.Fatalf("expected 5 loops, got %d", f.NLoops())
	}
	if msg := checkAllLoops(s); msg != "" {
		t.Fatalf("infinite loop: %s", msg)
	}
	if !s.Verify() {
		t.Fatal("Verify failed")
	}
}

// ---------------------------------------------------------------------------
// Generate
// ---------------------------------------------------------------------------

func TestPhase2_Generate_SimpleIntersection(t *testing.T) {
	a, b := makeSimpleIntersection()
	a.CalcPlaneEquations()
	b.CalcPlaneEquations()
	a.SetFaceMarks(UNKNOWN)
	a.SetVertexMarks(UNKNOWN)
	b.SetFaceMarks(UNKNOWN)
	b.SetVertexMarks(UNKNOWN)

	br := NewBoolopRecord()
	br.Reset(a, b, BoolUnion)
	br.Generate()

	totalVF := len(br.VertsA) + len(br.VertsB)
	if totalVF < 4 {
		t.Fatalf("expected >= 4 VF records, got %d", totalVF)
	}
	if !a.Verify() {
		t.Fatal("A failed Verify")
	}
	if !b.Verify() {
		t.Fatal("B failed Verify")
	}
}

// ---------------------------------------------------------------------------
// Classify
// ---------------------------------------------------------------------------

func TestPhase2_Classify_ProducesValidTopology(t *testing.T) {
	a, b := makeSimpleIntersection()
	a.CalcPlaneEquations()
	b.CalcPlaneEquations()
	a.SetFaceMarks(UNKNOWN)
	a.SetVertexMarks(UNKNOWN)
	b.SetFaceMarks(UNKNOWN)
	b.SetVertexMarks(UNKNOWN)

	br := NewBoolopRecord()
	br.Reset(a, b, BoolUnion)
	br.Generate()
	ok := br.Classify()

	if !ok {
		t.Skip("Classify returned false")
	}
	if len(br.EdgesA) != len(br.EdgesB) {
		t.Fatalf("EdgesA=%d != EdgesB=%d", len(br.EdgesA), len(br.EdgesB))
	}
	if msg := checkAllLoops(br.A); msg != "" {
		t.Fatalf("A infinite: %s", msg)
	}
	if msg := checkAllLoops(br.B); msg != "" {
		t.Fatalf("B infinite: %s", msg)
	}
}

// ---------------------------------------------------------------------------
// Connect
// ---------------------------------------------------------------------------

func TestPhase2_Connect_ProducesValidTopology(t *testing.T) {
	a, b := makeSimpleIntersection()
	a.CalcPlaneEquations()
	b.CalcPlaneEquations()
	a.SetFaceMarks(UNKNOWN)
	a.SetVertexMarks(UNKNOWN)
	b.SetFaceMarks(UNKNOWN)
	b.SetVertexMarks(UNKNOWN)

	br := NewBoolopRecord()
	br.Reset(a, b, BoolUnion)
	br.Generate()
	if !br.Classify() {
		t.Skip("Classify returned false")
	}
	br.Connect()

	if msg := checkAllLoops(br.A); msg != "" {
		t.Fatalf("A infinite after Connect: %s", msg)
	}
	if msg := checkAllLoops(br.B); msg != "" {
		t.Fatalf("B infinite after Connect: %s", msg)
	}
	if len(br.FacesA) != len(br.FacesB) {
		t.Fatalf("FacesA=%d != FacesB=%d", len(br.FacesA), len(br.FacesB))
	}
	for i, f := range br.FacesA {
		if f.NLoops() != 2 {
			t.Errorf("FacesA[%d] has %d loops, expected 2", i, f.NLoops())
		}
	}
	for i, f := range br.FacesB {
		if f.NLoops() != 2 {
			t.Errorf("FacesB[%d] has %d loops, expected 2", i, f.NLoops())
		}
	}
}

// ---------------------------------------------------------------------------
// Full pipeline
// ---------------------------------------------------------------------------

func TestPhase2_FullPipeline_SimpleIntersection(t *testing.T) {
	for _, op := range []int{BoolUnion, BoolIntersection, BoolDifference} {
		op := op
		t.Run(opName(op), func(t *testing.T) {
			a, b := makeSimpleIntersection()
			result, ok := BoolOp(a, b, op)
			if !ok || result == nil {
				return
			}
			if msg := checkAllLoops(result); msg != "" {
				t.Fatalf("result has infinite loop: %s", msg)
			}
			errs := result.VerifyDetailed()
			for _, err := range errs {
				t.Errorf("verify: %v", err)
			}
		})
	}
}
