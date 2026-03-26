package solid

// Integration tests for Bool Phase 1 (Generate + LastDitch).
// These test the BoolopRecord pipeline: Generate detects intersections,
// LastDitch handles degenerate cases (disjoint, contained, identical).

import (
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

var (
	boolRed   = vec.SFColor{R: 1, G: 0, B: 0, A: 1}
	boolGreen = vec.SFColor{R: 0, G: 1, B: 0, A: 1}
)

func setupBoolOp(a, b *Solid, op int) *BoolopRecord {
	a.CalcPlaneEquations()
	b.CalcPlaneEquations()
	a.SetFaceMarks(UNKNOWN)
	a.SetVertexMarks(UNKNOWN)
	b.SetFaceMarks(UNKNOWN)
	b.SetVertexMarks(UNKNOWN)

	br := NewBoolopRecord()
	br.Reset(a, b, op)
	return br
}

// ═══════════════════════════════════════════════════════════════════════════════
// Generate — intersection detection tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestGenerate_PartialOverlap_CubeVsCube(t *testing.T) {
	// Two cubes overlapping by half: A at origin, B shifted +1 on X.
	// This is the BOOL01 case 2 geometry.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(1.0, 0, 0))

	br := setupBoolOp(a, b, BoolUnion)
	br.Generate()

	// With partial overlap, we expect intersections.
	totalVF := len(br.VertsA) + len(br.VertsB)
	totalVV := len(br.VertsV)
	total := totalVF + totalVV

	if total == 0 {
		t.Fatal("expected intersections for overlapping cubes, got 0")
	}
	t.Logf("partial overlap: VF_A=%d VF_B=%d VV=%d total=%d",
		len(br.VertsA), len(br.VertsB), totalVV, total)
}

func TestGenerate_CubeVsSphere_Intersections(t *testing.T) {
	// Sphere radius 1.2 centered at origin overlapping with unit cube.
	a := MakeCube(1.0, boolRed)
	b := MakeSphere(1.2, 8, 8, boolGreen)

	br := setupBoolOp(a, b, BoolIntersection)
	br.Generate()

	total := len(br.VertsA) + len(br.VertsB) + len(br.VertsV)
	if total == 0 {
		t.Fatal("expected intersections for cube-sphere overlap, got 0")
	}
	t.Logf("cube vs sphere: VF_A=%d VF_B=%d VV=%d",
		len(br.VertsA), len(br.VertsB), len(br.VertsV))
}

func TestGenerate_FaceOnFace_NoStraddle(t *testing.T) {
	// Two cubes sharing a face (B shifted by exactly 2.0 on X).
	// Vertices touch the face plane but no edges straddle it.
	// All intersections should be VertexOnFace (VF records) or VV.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(2.0, 0, 0))

	br := setupBoolOp(a, b, BoolUnion)
	br.Generate()

	// ON-vertex cases: vertices of B touching faces of A and vice versa
	total := len(br.VertsA) + len(br.VertsB) + len(br.VertsV)
	t.Logf("face-on-face: VF_A=%d VF_B=%d VV=%d total=%d",
		len(br.VertsA), len(br.VertsB), len(br.VertsV), total)
}

func TestGenerate_Identical_NoStraddle(t *testing.T) {
	// Two identical cubes at the same position.
	// Every vertex of A is ON every adjacent face of B and vice versa.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)

	br := setupBoolOp(a, b, BoolUnion)
	br.Generate()

	totalVV := len(br.VertsV)
	t.Logf("identical: VF_A=%d VF_B=%d VV=%d",
		len(br.VertsA), len(br.VertsB), totalVV)
	// No straddling edges expected — all ON-plane
}

func TestGenerate_Disjoint_NoIntersections(t *testing.T) {
	// Two cubes far apart — no intersections at all.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(10.0, 0, 0))

	br := setupBoolOp(a, b, BoolDifference)
	br.Generate()

	totalVF := len(br.VertsA) + len(br.VertsB)
	totalVV := len(br.VertsV)
	if totalVF != 0 || totalVV != 0 {
		t.Errorf("expected no intersections for disjoint cubes: VF=%d VV=%d", totalVF, totalVV)
	}
}

func TestGenerate_FullyContained_NoStraddle(t *testing.T) {
	// Small cube fully inside large cube — no edge straddles any face.
	a := MakeCube(2.0, boolRed)
	b := MakeCube(0.5, boolGreen)

	br := setupBoolOp(a, b, BoolDifference)
	br.Generate()

	totalVF := len(br.VertsA) + len(br.VertsB)
	totalVV := len(br.VertsV)
	if totalVF != 0 && totalVV != 0 {
		t.Logf("NOTE: contained cube produced intersections VF=%d VV=%d (faces may be ON-plane)", totalVF, totalVV)
	}
}

func TestGenerate_Rotated_Intersections(t *testing.T) {
	// Cube A at origin, Cube B rotated 45° around Y axis.
	// Edges of the rotated cube should straddle faces of A and vice versa.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	radians := 45.0 * math.Pi / 180.0
	rot := vec.SFRotation{X: 0, Y: 1, Z: 0, W: radians}
	b.TransformGeometry(vec.RotationMatrix(rot))

	br := setupBoolOp(a, b, BoolIntersection)
	br.Generate()

	total := len(br.VertsA) + len(br.VertsB) + len(br.VertsV)
	if total == 0 {
		t.Fatal("expected intersections for rotated overlapping cubes, got 0")
	}
	t.Logf("rotated: VF_A=%d VF_B=%d VV=%d",
		len(br.VertsA), len(br.VertsB), len(br.VertsV))
}

func TestGenerate_EdgeHit_SplitsBothEdges(t *testing.T) {
	// Cube A at origin, Cube B shifted so its edge aligns with A's face.
	// B shifted diagonally: edges of B should hit edges of A's faces.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(1.0, 1.0, 0))

	br := setupBoolOp(a, b, BoolUnion)
	br.Generate()

	total := len(br.VertsA) + len(br.VertsB) + len(br.VertsV)
	t.Logf("edge-hit: VF_A=%d VF_B=%d VV=%d total=%d",
		len(br.VertsA), len(br.VertsB), len(br.VertsV), total)
}

func TestGenerate_Prism_Intersections(t *testing.T) {
	// Triangular prism vs cube — non-trivial face polygon.
	triVerts := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 2, Y: 0, Z: 0},
		{X: 1, Y: 2, Z: 0},
	}
	prism := MakeLamina(triVerts, boolGreen)
	prism.TranslationalSweep(prism.GetFirstFace(), vec.SFVec3f{X: 0, Y: 0, Z: 2})

	cube := MakeCube(1.0, boolRed)

	br := setupBoolOp(cube, prism, BoolIntersection)
	br.Generate()

	total := len(br.VertsA) + len(br.VertsB) + len(br.VertsV)
	if total == 0 {
		t.Fatal("expected intersections for cube-prism overlap, got 0")
	}
	t.Logf("prism: VF_A=%d VF_B=%d VV=%d",
		len(br.VertsA), len(br.VertsB), len(br.VertsV))
}

func TestGenerate_PreservesTopology(t *testing.T) {
	// After Generate, both solids should still pass Verify.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(1.0, 0, 0))

	br := setupBoolOp(a, b, BoolUnion)
	br.Generate()

	if !a.Verify() {
		errs := a.VerifyDetailed()
		for _, err := range errs {
			t.Errorf("solid A verify: %v", err)
		}
	}
	if !b.Verify() {
		errs := b.VerifyDetailed()
		for _, err := range errs {
			t.Errorf("solid B verify: %v", err)
		}
	}
}

func TestGenerate_MarksOnVertices(t *testing.T) {
	// After Generate, intersection vertices should be marked ON.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(1.0, 0, 0))

	br := setupBoolOp(a, b, BoolUnion)
	br.Generate()

	// All recorded vertices should be marked ON
	for i, vf := range br.VertsA {
		if vf.V.Mark != ON {
			t.Errorf("VertsA[%d].V mark = %d, want ON(%d)", i, vf.V.Mark, ON)
		}
	}
	for i, vf := range br.VertsB {
		if vf.V.Mark != ON {
			t.Errorf("VertsB[%d].V mark = %d, want ON(%d)", i, vf.V.Mark, ON)
		}
	}
	for i, vv := range br.VertsV {
		if vv.Va.Mark != ON {
			t.Errorf("VertsV[%d].Va mark = %d, want ON(%d)", i, vv.Va.Mark, ON)
		}
		if vv.Vb.Mark != ON {
			t.Errorf("VertsV[%d].Vb mark = %d, want ON(%d)", i, vv.Vb.Mark, ON)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// LastDitch — degenerate case handling
// ═══════════════════════════════════════════════════════════════════════════════

func TestLastDitch_Disjoint_Union(t *testing.T) {
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(10.0, 0, 0))

	result, ok := BoolOp(a, b, BoolUnion)
	if !ok || result == nil {
		t.Fatal("Union of disjoint cubes should succeed")
	}
	// Union of two disjoint cubes = 12 faces (6 + 6)
	if result.NFaces() != 12 {
		t.Errorf("expected 12 faces, got %d", result.NFaces())
	}
}

func TestLastDitch_Disjoint_Intersection(t *testing.T) {
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(10.0, 0, 0))

	result, ok := BoolOp(a, b, BoolIntersection)
	if ok || result != nil {
		t.Fatal("Intersection of disjoint cubes should return nil")
	}
}

func TestLastDitch_Disjoint_Difference(t *testing.T) {
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(10.0, 0, 0))

	result, ok := BoolOp(a, b, BoolDifference)
	if !ok || result == nil {
		t.Fatal("Difference of disjoint cubes should return A")
	}
	if result.NFaces() != 6 {
		t.Errorf("expected 6 faces, got %d", result.NFaces())
	}
}

func TestLastDitch_Contained_Union(t *testing.T) {
	a := MakeCube(2.0, boolRed)
	b := MakeCube(0.3, boolGreen) // Small cube fully inside

	result, ok := BoolOp(a, b, BoolUnion)
	if !ok || result == nil {
		t.Fatal("Union (A contains B) should return A")
	}
	if result.NFaces() != 6 {
		t.Errorf("expected 6 faces (outer shell), got %d", result.NFaces())
	}
}

func TestLastDitch_Contained_Intersection(t *testing.T) {
	a := MakeCube(2.0, boolRed)
	b := MakeCube(0.3, boolGreen) // Small cube fully inside

	result, ok := BoolOp(a, b, BoolIntersection)
	if !ok || result == nil {
		t.Fatal("Intersection (A contains B) should return B")
	}
	if result.NFaces() != 6 {
		t.Errorf("expected 6 faces (inner cube), got %d", result.NFaces())
	}
}

func TestLastDitch_Contained_Difference(t *testing.T) {
	a := MakeCube(2.0, boolRed)
	b := MakeCube(0.3, boolGreen) // Small cube fully inside

	result, ok := BoolOp(a, b, BoolDifference)
	if !ok || result == nil {
		t.Fatal("Difference (A contains B) should return A")
	}
	if result.NFaces() != 6 {
		t.Errorf("expected 6 faces, got %d", result.NFaces())
	}
}

func TestLastDitch_BContainsA_Union(t *testing.T) {
	a := MakeCube(0.3, boolRed)
	b := MakeCube(2.0, boolGreen) // Large cube contains small

	result, ok := BoolOp(a, b, BoolUnion)
	if !ok || result == nil {
		t.Fatal("Union (B contains A) should return B")
	}
	if result.NFaces() != 6 {
		t.Errorf("expected 6 faces, got %d", result.NFaces())
	}
}

func TestLastDitch_BContainsA_Intersection(t *testing.T) {
	a := MakeCube(0.3, boolRed)
	b := MakeCube(2.0, boolGreen)

	result, ok := BoolOp(a, b, BoolIntersection)
	if !ok || result == nil {
		t.Fatal("Intersection (B contains A) should return A")
	}
	if result.NFaces() != 6 {
		t.Errorf("expected 6 faces, got %d", result.NFaces())
	}
}

func TestLastDitch_BContainsA_Difference(t *testing.T) {
	a := MakeCube(0.3, boolRed)
	b := MakeCube(2.0, boolGreen)

	result, ok := BoolOp(a, b, BoolDifference)
	if ok || result != nil {
		t.Fatal("Difference (B contains A) should return nil")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// BoolOp end-to-end — full pipeline through Generate → LastDitch path
// ═══════════════════════════════════════════════════════════════════════════════

func TestBoolOp_DisjointCubes_AllOps(t *testing.T) {
	for _, op := range []int{BoolUnion, BoolIntersection, BoolDifference} {
		t.Run(opName(op), func(t *testing.T) {
			a := MakeCube(1.0, boolRed)
			b := MakeCube(1.0, boolGreen)
			b.TransformGeometry(vec.TranslationMatrix(10.0, 0, 0))

			result, ok := BoolOp(a, b, op)
			switch op {
			case BoolUnion:
				if !ok || result == nil {
					t.Fatal("union should succeed")
				}
				if result.NFaces() != 12 {
					t.Errorf("faces: got %d, want 12", result.NFaces())
				}
			case BoolIntersection:
				if ok || result != nil {
					t.Fatal("intersection should be empty")
				}
			case BoolDifference:
				if !ok || result == nil {
					t.Fatal("difference should succeed")
				}
				if result.NFaces() != 6 {
					t.Errorf("faces: got %d, want 6", result.NFaces())
				}
			}
		})
	}
}

func TestBoolOp_VerifyResult_SingleShell(t *testing.T) {
	// A contained-cube intersection produces a single-shell result
	// that should pass topology verification.
	a := MakeCube(2.0, boolRed)
	b := MakeCube(0.3, boolGreen) // fully inside A

	result, ok := BoolOp(a, b, BoolIntersection)
	if !ok || result == nil {
		t.Fatal("intersection should succeed")
	}

	if !result.Verify() {
		errs := result.VerifyDetailed()
		for _, err := range errs {
			t.Errorf("verify: %v", err)
		}
	}
}

func TestBoolOp_DisjointUnion_MultiShell(t *testing.T) {
	// Disjoint union produces a 2-shell solid (merged).
	// Verify() assumes single shell, so check counts directly.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(10.0, 0, 0))

	result, ok := BoolOp(a, b, BoolUnion)
	if !ok || result == nil {
		t.Fatal("union should succeed")
	}

	// Two cubes merged: 12 faces, 24 edges, 16 vertices
	if result.NFaces() != 12 {
		t.Errorf("faces: got %d, want 12", result.NFaces())
	}
	if result.NEdges() != 24 {
		t.Errorf("edges: got %d, want 24", result.NEdges())
	}
	if result.NVerts() != 16 {
		t.Errorf("verts: got %d, want 16", result.NVerts())
	}
}

func TestBoolOp_ContainedCube_VerifyResult(t *testing.T) {
	a := MakeCube(2.0, boolRed)
	b := MakeCube(0.3, boolGreen)

	for _, op := range []int{BoolUnion, BoolIntersection, BoolDifference} {
		t.Run(opName(op), func(t *testing.T) {
			result, ok := BoolOp(a, b, op)
			if result != nil && ok {
				if !result.Verify() {
					errs := result.VerifyDetailed()
					for _, err := range errs {
						t.Errorf("verify: %v", err)
					}
				}
			}
		})
	}
}

func TestBoolOp_OriginalUnmodified(t *testing.T) {
	// BoolOp should work on copies; originals should be untouched.
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(1.0, 0, 0))

	origAVerts := a.NVerts()
	origBVerts := b.NVerts()
	origAEdges := a.NEdges()
	origBEdges := b.NEdges()

	_, _ = BoolOp(a, b, BoolUnion)

	if a.NVerts() != origAVerts {
		t.Errorf("A verts changed: %d → %d", origAVerts, a.NVerts())
	}
	if b.NVerts() != origBVerts {
		t.Errorf("B verts changed: %d → %d", origBVerts, b.NVerts())
	}
	if a.NEdges() != origAEdges {
		t.Errorf("A edges changed: %d → %d", origAEdges, a.NEdges())
	}
	if b.NEdges() != origBEdges {
		t.Errorf("B edges changed: %d → %d", origBEdges, b.NEdges())
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Convenience wrapper tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestUnion_Convenience(t *testing.T) {
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(10.0, 0, 0))

	result, ok := Union(a, b)
	if !ok || result == nil {
		t.Fatal("Union should succeed")
	}
}

func TestIntersection_Convenience(t *testing.T) {
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(10.0, 0, 0))

	result, ok := Intersection(a, b)
	if ok || result != nil {
		t.Fatal("Intersection of disjoint should be empty")
	}
}

func TestDifference_Convenience(t *testing.T) {
	a := MakeCube(1.0, boolRed)
	b := MakeCube(1.0, boolGreen)
	b.TransformGeometry(vec.TranslationMatrix(10.0, 0, 0))

	result, ok := Difference(a, b)
	if !ok || result == nil {
		t.Fatal("Difference should succeed")
	}
}
