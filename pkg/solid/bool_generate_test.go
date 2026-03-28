package solid

// Pre-implementation tests for Bool Phase 1 (Edge-Face Intersection Generation).
// These exercise the exact primitives and edge cases that Generate will rely on.
//
// Section 1: Topology validation — tests existing APIs under stress
// Section 2: Intersection-specific — tests that will validate Generate output

import (
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

var (
	testRed   = vec.SFColor{R: 1, G: 0, B: 0, A: 1}
	testGreen = vec.SFColor{R: 0, G: 1, B: 0, A: 1}
)

// vertexAt returns the first vertex at the given location (within tolerance).
// func vertexAt(s *Solid, x, y, z float64) *Vertex {
// 	target := vec.SFVec3f{X: x, Y: y, Z: z}
// 	var found *Vertex
// 	s.ForEachVertex(func(v *Vertex) bool {
// 		d := v.Loc.Sub(target)
// 		if d.Length() < 1e-7 {
// 			found = v
// 			return false
// 		}
// 		return true
// 	})
// 	return found
// }

// faceWithNormal returns the first face whose normal is closest to n.
func faceWithNormal(s *Solid, n vec.SFVec3f) *Face {
	n = n.Normalize()
	var best *Face
	bestDot := -2.0
	s.ForEachFace(func(f *Face) bool {
		d := f.Normal.Dot(n)
		if d > bestDot {
			bestDot = d
			best = f
		}
		return true
	})
	return best
}

// countVertices returns number of vertices in a solid.
// func countVertices(s *Solid) int {
// 	_, _, v := s.Stats()
// 	return v
// }

// countEdges returns number of edges in a solid.
// func countEdges(s *Solid) int {
// 	_, e, _ := s.Stats()
// 	return e
// }

// ═══════════════════════════════════════════════════════════════════════════════
// SECTION 1: Topology Validation Tests
// Exercise existing APIs under degenerate/stressful inputs
// ═══════════════════════════════════════════════════════════════════════════════

// ---------------------------------------------------------------------------
// TestLmev2_NearEndpoint: split an edge very close to its endpoints.
// Validates that sliver edges maintain valid topology.
// ---------------------------------------------------------------------------

func TestLmev2_NearEndpoint(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	initialF, initialE, initialV := cube.Stats()

	// Find an edge and split it at t=0.001 (very near v1)
	e := cube.Edges
	v1 := e.He1.Vertex
	v2 := e.He2.Vertex

	splitPt := v1.Loc.Add(v2.Loc.Sub(v1.Loc).Scale(0.001))
	newV, newE := Lmev2(e.He1, e.He2.Next, splitPt)

	if newV == nil || newE == nil {
		t.Fatal("Lmev2 returned nil for near-endpoint split")
		return
	}

	// Topology should gain exactly 1 vertex and 1 edge, same faces
	f2, e2, v2count := cube.Stats()
	if f2 != initialF {
		t.Errorf("face count changed: %d -> %d", initialF, f2)
	}
	if e2 != initialE+1 {
		t.Errorf("edge count: got %d, want %d", e2, initialE+1)
	}
	if v2count != initialV+1 {
		t.Errorf("vertex count: got %d, want %d", v2count, initialV+1)
	}

	// New vertex should be at the split point
	dist := newV.Loc.Sub(splitPt).Length()
	if dist > 1e-12 {
		t.Errorf("new vertex at wrong location, distance=%e", dist)
	}

	// Verify Euler consistency
	if !cube.Verify() {
		t.Error("Verify failed after near-endpoint split")
	}

	// Now split again near the other endpoint (t=0.999 of original edge)
	cube2 := MakeCube(1.0, testRed)
	e = cube2.Edges
	v1 = e.He1.Vertex
	v2loc := e.He2.Vertex.Loc
	splitPt2 := v1.Loc.Add(v2loc.Sub(v1.Loc).Scale(0.999))
	newV2, _ := Lmev2(e.He1, e.He2.Next, splitPt2)
	if newV2 == nil {
		t.Fatal("Lmev2 returned nil for near-far-endpoint split")
	}
	if !cube2.Verify() {
		t.Error("Verify failed after near-far-endpoint split")
	}
}

// ---------------------------------------------------------------------------
// TestFaceContains_OnEdge: point exactly on a face edge.
// Must return EdgeHit, not FaceHit.
// ---------------------------------------------------------------------------

func TestFaceContains_OnEdge(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	// Find the +Z face (normal pointing in +Z direction)
	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	// Get the midpoint of the first edge of the top face
	he := topFace.GetFirstHe()
	mid := he.Vertex.Loc.Add(he.Next.Vertex.Loc).Scale(0.5)

	testV := NewVertexVec(mid)
	rec := NewIntersectRecord()
	hit := topFace.FaceContains(testV, &rec)

	if !hit {
		t.Fatal("FaceContains returned false for point on edge")
	}
	if rec.Type != EdgeHit {
		t.Errorf("expected EdgeHit (%d), got %d", EdgeHit, rec.Type)
	}
	if rec.He == nil {
		t.Error("EdgeHit but He is nil")
	}
}

// ---------------------------------------------------------------------------
// TestFaceContains_OnVertex: point exactly on a face vertex.
// Must return VertexHit, not EdgeHit or FaceHit.
// ---------------------------------------------------------------------------

func TestFaceContains_OnVertex(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	// The point IS an existing vertex location
	cornerLoc := topFace.GetFirstHe().Vertex.Loc
	testV := NewVertexVec(cornerLoc)

	rec := NewIntersectRecord()
	hit := topFace.FaceContains(testV, &rec)

	if !hit {
		t.Fatal("FaceContains returned false for point on vertex")
	}
	if rec.Type != VertexHit {
		t.Errorf("expected VertexHit (%d), got %d", VertexHit, rec.Type)
	}
	if rec.Vert == nil {
		t.Error("VertexHit but Vert is nil")
	}
}

// ---------------------------------------------------------------------------
// TestFaceContains_NearEdge: point at 1e-8 from a face edge.
// Tests the tolerance boundary between EdgeHit and FaceHit.
// ---------------------------------------------------------------------------

func TestFaceContains_NearEdge(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
		return
	}

	// Get edge midpoint, then offset slightly into the face interior
	he := topFace.GetFirstHe()
	mid := he.Vertex.Loc.Add(he.Next.Vertex.Loc).Scale(0.5)

	// Compute inward normal of the edge within the face plane
	edgeDir := he.Next.Vertex.Loc.Sub(he.Vertex.Loc).Normalize()
	inward := topFace.Normal.Cross(edgeDir).Normalize()

	// Point at 1e-8 inside the face from the edge midpoint
	nearPt := mid.Add(inward.Scale(1e-8))
	testV := NewVertexVec(nearPt)

	rec := NewIntersectRecord()
	hit := topFace.FaceContains(testV, &rec)

	if !hit {
		t.Fatal("FaceContains returned false for point near edge")
	}

	// Either EdgeHit (snapped) or FaceHit (not snapped) are acceptable outcomes,
	// but it must not return NoHit or crash.
	t.Logf("Near-edge point (1e-8 offset): hit type = %d", rec.Type)
}

// ---------------------------------------------------------------------------
// TestGetDistance_Coplanar: vertex in the face plane but outside the face.
// Distance=0 but FaceContains=false.
// ---------------------------------------------------------------------------

func TestGetDistance_Coplanar(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	// Point at (10, 10, z_of_top_face) — in the plane but far outside
	topZ := topFace.GetFirstHe().Vertex.Loc.Z
	farPt := vec.SFVec3f{X: 10, Y: 10, Z: topZ}

	dist := topFace.GetDistance(farPt)
	if math.Abs(dist) > 1e-10 {
		t.Errorf("expected distance ≈ 0 for coplanar point, got %e", dist)
	}

	testV := NewVertexVec(farPt)
	rec := NewIntersectRecord()
	hit := topFace.FaceContains(testV, &rec)
	if hit {
		t.Error("FaceContains returned true for coplanar point outside face")
	}
}

// ---------------------------------------------------------------------------
// TestSolidContains_Touching: Solid B shares a face plane with A but doesn't
// penetrate. Tests the ambiguous boundary case.
// ---------------------------------------------------------------------------

func TestSolidContains_Touching(t *testing.T) {
	a := MakeCube(1.0, testRed)
	b := MakeCube(1.0, testGreen)

	// Translate B so its -Z face sits on A's +Z face
	b.TransformGeometry(vec.TranslationMatrix(0, 0, 2.0)) // face-touching, not overlapping

	a.CalcPlaneEquations()
	b.CalcPlaneEquations()

	// Neither should contain the other
	if a.SolidContains(b) {
		t.Error("A should not contain face-touching B")
	}
	if b.SolidContains(a) {
		t.Error("B should not contain face-touching A")
	}
}

// ---------------------------------------------------------------------------
// TestGetDistance_SignedDistances: verify signed distance for all cube faces.
// ---------------------------------------------------------------------------

func TestGetDistance_SignedDistances(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	// Center point should be inside — negative distance to all faces?
	// Actually with outward normals, center should have negative distance.
	center := vec.SFVec3f{X: 0, Y: 0, Z: 0}
	cube.ForEachFace(func(f *Face) bool {
		d := f.GetDistance(center)
		// With outward-pointing normals and D set so plane goes through vertices,
		// center should be "below" (inside) every face
		if d > 1e-5 {
			t.Errorf("center has positive distance %f to face with normal %v", d, f.Normal)
		}
		return true
	})

	// Point far outside should have positive distance to at least one face
	outside := vec.SFVec3f{X: 5, Y: 0, Z: 0}
	anyPositive := false
	cube.ForEachFace(func(f *Face) bool {
		d := f.GetDistance(outside)
		if d > 0 {
			anyPositive = true
			return false
		}
		return true
	})
	if !anyPositive {
		t.Error("point at (5,0,0) should have positive distance to at least one face")
	}
}

// ---------------------------------------------------------------------------
// TestEdgeStraddle_Classification: verify signed distance correctly classifies
// edge endpoints as straddling a face.
// ---------------------------------------------------------------------------

func TestEdgeStraddle_Classification(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	topZ := topFace.GetFirstHe().Vertex.Loc.Z

	// Edge that clearly straddles the top face
	above := vec.SFVec3f{X: 0, Y: 0, Z: topZ + 0.5}
	below := vec.SFVec3f{X: 0, Y: 0, Z: topZ - 0.5}

	d1 := topFace.GetDistance(above)
	d2 := topFace.GetDistance(below)

	if d1 <= 0 {
		t.Errorf("point above face should have positive distance, got %f", d1)
	}
	if d2 >= 0 {
		t.Errorf("point below face should have negative distance, got %f", d2)
	}

	// Compute intersection parameter
	tParam := d1 / (d1 - d2)
	if tParam < 0 || tParam > 1 {
		t.Errorf("intersection t should be in [0,1], got %f", tParam)
	}

	// Check the intersection point
	pt := above.Add(below.Sub(above).Scale(tParam))
	ptDist := math.Abs(topFace.GetDistance(pt))
	if ptDist > 1e-10 {
		t.Errorf("intersection point should be on face plane, distance=%e", ptDist)
	}
}

// ---------------------------------------------------------------------------
// TestEdgeStraddle_ParallelToFace: edge parallel to face, no intersection.
// ---------------------------------------------------------------------------

func TestEdgeStraddle_ParallelToFace(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	topZ := topFace.GetFirstHe().Vertex.Loc.Z

	// Edge parallel to face, at slight offset above
	p1 := vec.SFVec3f{X: -2, Y: 0, Z: topZ + 0.001}
	p2 := vec.SFVec3f{X: 2, Y: 0, Z: topZ + 0.001}

	d1 := topFace.GetDistance(p1)
	d2 := topFace.GetDistance(p2)

	// Both should be same sign — no straddling
	if (d1 > 0) != (d2 > 0) {
		t.Error("parallel edge endpoints should have same sign distance")
	}
}

// ---------------------------------------------------------------------------
// TestEdgeStraddle_OnFacePlane: edge endpoint exactly on face plane.
// This is the s1==0 or s2==0 degenerate case from boolgenerate.
// ---------------------------------------------------------------------------

func TestEdgeStraddle_OnFacePlane(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	topZ := topFace.GetFirstHe().Vertex.Loc.Z

	// One endpoint ON the face plane, other above
	onPlane := vec.SFVec3f{X: 0, Y: 0, Z: topZ}
	abovePlane := vec.SFVec3f{X: 0, Y: 0, Z: topZ + 1.0}

	d1 := topFace.GetDistance(onPlane)
	d2 := topFace.GetDistance(abovePlane)

	if math.Abs(d1) > 1e-10 {
		t.Errorf("ON-plane point distance should be ≈ 0, got %e", d1)
	}
	if d2 <= 0 {
		t.Errorf("above point should have positive distance, got %f", d2)
	}

	// This is NOT straddling — both signs are >=0. Generate should handle
	// via VertexOnFace path, not edge splitting.
	s1 := floatCompare(d1, 0)
	s2 := floatCompare(d2, 0)

	straddling := (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0)
	if straddling {
		t.Error("ON-plane endpoint should not be classified as straddling")
	}
	if s1 != 0 {
		t.Errorf("ON-plane endpoint should classify as ON (0), got %d", s1)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// SECTION 2: Intersection-Specific Tests
// These test the Generate algorithm requirements directly.
// Tests that need BooleanOp are skipped; tests that validate the pieces are live.
// ═══════════════════════════════════════════════════════════════════════════════

// ---------------------------------------------------------------------------
// TestGenerate_EdgePiercesFace: single edge through center of a face.
// Validates the core straddle→interpolate→FaceContains path.
// ---------------------------------------------------------------------------

func TestGenerate_EdgePiercesFace(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	topZ := topFace.GetFirstHe().Vertex.Loc.Z

	// Edge that pierces through the middle of the top face
	above := vec.SFVec3f{X: 0.1, Y: 0.1, Z: topZ + 0.5}
	below := vec.SFVec3f{X: 0.1, Y: 0.1, Z: topZ - 0.5}

	// Step 1: Compute signed distances
	d1 := topFace.GetDistance(above)
	d2 := topFace.GetDistance(below)

	if d1 <= 0 || d2 >= 0 {
		t.Fatalf("edge does not straddle: d1=%f, d2=%f", d1, d2)
	}

	// Step 2: Interpolate
	tParam := d1 / (d1 - d2)
	pt := above.Add(below.Sub(above).Scale(tParam))

	// Step 3: Point should be on the face plane
	if math.Abs(topFace.GetDistance(pt)) > 1e-10 {
		t.Fatalf("interpolated point not on plane: dist=%e", topFace.GetDistance(pt))
	}

	// Step 4: Point should be inside the face
	testV := NewVertexVec(pt)
	rec := NewIntersectRecord()
	hit := topFace.FaceContains(testV, &rec)

	if !hit {
		t.Fatal("intersection point should be inside the face")
	}
	if rec.Type != FaceHit {
		t.Errorf("expected FaceHit for face interior, got type %d", rec.Type)
	}
}

// ---------------------------------------------------------------------------
// TestGenerate_EdgeGlancesFace: edge has one endpoint on face plane.
// Should trigger VertexOnFace path, not edge splitting.
// ---------------------------------------------------------------------------

func TestGenerate_EdgeGlancesFace(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	topZ := topFace.GetFirstHe().Vertex.Loc.Z

	// One endpoint exactly ON the face, inside the face polygon
	onFace := vec.SFVec3f{X: 0.1, Y: 0.1, Z: topZ}
	above := vec.SFVec3f{X: 0.1, Y: 0.1, Z: topZ + 1.0}

	d1 := topFace.GetDistance(onFace)
	d2 := topFace.GetDistance(above)

	s1 := floatCompare(d1, 0)
	s2 := floatCompare(d2, 0)

	// Not straddling — s1 is ON
	if (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0) {
		t.Fatal("should not classify as straddling when one endpoint is ON")
	}

	// The ON endpoint is inside the face polygon
	testV := NewVertexVec(onFace)
	rec := NewIntersectRecord()
	hit := topFace.FaceContains(testV, &rec)
	if !hit {
		t.Error("ON-plane point inside face polygon should be contained")
	}
	if rec.Type != FaceHit {
		t.Logf("ON-face hit type = %d (FaceHit=%d, EdgeHit=%d, VertexHit=%d)", rec.Type, FaceHit, EdgeHit, VertexHit)
	}
}

// ---------------------------------------------------------------------------
// TestGenerate_EdgeAlongFace: edge lies entirely in face plane.
// Degenerate case — both endpoints have zero distance.
// ---------------------------------------------------------------------------

func TestGenerate_EdgeAlongFace(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	topZ := topFace.GetFirstHe().Vertex.Loc.Z

	// Both endpoints on the face plane
	p1 := vec.SFVec3f{X: -0.3, Y: 0.0, Z: topZ}
	p2 := vec.SFVec3f{X: 0.3, Y: 0.0, Z: topZ}

	d1 := topFace.GetDistance(p1)
	d2 := topFace.GetDistance(p2)

	s1 := floatCompare(d1, 0)
	s2 := floatCompare(d2, 0)

	// Both should be ON — no straddling, no splitting
	if s1 != 0 {
		t.Errorf("p1 should be ON, classified as %d (dist=%e)", s1, d1)
	}
	if s2 != 0 {
		t.Errorf("p2 should be ON, classified as %d (dist=%e)", s2, d2)
	}

	// Neither should trigger edge splitting (no straddling)
	straddling := (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0)
	if straddling {
		t.Error("coplanar edge should not be classified as straddling")
	}
}

// ---------------------------------------------------------------------------
// TestGenerate_EdgeHitsFaceEdge: intersection lands on a face edge.
// The containment test should return EdgeHit.
// ---------------------------------------------------------------------------

func TestGenerate_EdgeHitsFaceEdge(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	topZ := topFace.GetFirstHe().Vertex.Loc.Z

	// Find the midpoint of the top face's first edge
	he := topFace.GetFirstHe()
	edgeMid := he.Vertex.Loc.Add(he.Next.Vertex.Loc).Scale(0.5)

	// Edge piercing through that edge midpoint
	above := vec.SFVec3f{X: edgeMid.X, Y: edgeMid.Y, Z: topZ + 0.5}
	below := vec.SFVec3f{X: edgeMid.X, Y: edgeMid.Y, Z: topZ - 0.5}

	d1 := topFace.GetDistance(above)
	d2 := topFace.GetDistance(below)
	tParam := d1 / (d1 - d2)
	pt := above.Add(below.Sub(above).Scale(tParam))

	testV := NewVertexVec(pt)
	rec := NewIntersectRecord()
	hit := topFace.FaceContains(testV, &rec)

	if !hit {
		t.Fatal("intersection on face edge should be contained")
	}
	if rec.Type != EdgeHit {
		t.Errorf("expected EdgeHit (%d), got type %d", EdgeHit, rec.Type)
	}
	if rec.He == nil {
		t.Error("EdgeHit should record the half-edge")
	}
}

// ---------------------------------------------------------------------------
// TestGenerate_EdgeHitsVertex: intersection lands on existing vertex.
// ---------------------------------------------------------------------------

func TestGenerate_EdgeHitsVertex(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
	}

	topZ := topFace.GetFirstHe().Vertex.Loc.Z

	// Edge piercing through an existing corner vertex of the top face
	cornerLoc := topFace.GetFirstHe().Vertex.Loc

	above := vec.SFVec3f{X: cornerLoc.X, Y: cornerLoc.Y, Z: topZ + 0.5}
	below := vec.SFVec3f{X: cornerLoc.X, Y: cornerLoc.Y, Z: topZ - 0.5}

	d1 := topFace.GetDistance(above)
	d2 := topFace.GetDistance(below)
	tParam := d1 / (d1 - d2)
	pt := above.Add(below.Sub(above).Scale(tParam))

	testV := NewVertexVec(pt)
	rec := NewIntersectRecord()
	hit := topFace.FaceContains(testV, &rec)

	if !hit {
		t.Fatal("intersection at face vertex should be contained")
	}
	if rec.Type != VertexHit {
		t.Errorf("expected VertexHit (%d), got type %d", VertexHit, rec.Type)
	}
	if rec.Vert == nil {
		t.Error("VertexHit should record the vertex")
	}
}

// ---------------------------------------------------------------------------
// TestGenerate_LongEdgeMultipleFaces: one edge pierces 3 faces of a cube.
// After Generate, expect 3 intersection points.
// ---------------------------------------------------------------------------

func TestGenerate_LongEdgeMultipleFaces(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	// Long diagonal edge that pierces multiple faces
	// From (-2, 0.1, 0.1) to (2, 0.1, 0.1) — pierces -X face and +X face
	p1 := vec.SFVec3f{X: -2, Y: 0.1, Z: 0.1}
	p2 := vec.SFVec3f{X: 2, Y: 0.1, Z: 0.1}

	// Count how many faces this edge straddles
	straddleCount := 0
	cube.ForEachFace(func(f *Face) bool {
		d1 := f.GetDistance(p1)
		d2 := f.GetDistance(p2)
		s1 := floatCompare(d1, 0)
		s2 := floatCompare(d2, 0)
		if (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0) {
			// Compute intersection point and check containment
			tParam := d1 / (d1 - d2)
			pt := p1.Add(p2.Sub(p1).Scale(tParam))
			testV := NewVertexVec(pt)
			rec := NewIntersectRecord()
			if f.FaceContains(testV, &rec) {
				straddleCount++
			}
		}
		return true
	})

	// Should pierce exactly the -X and +X faces
	if straddleCount != 2 {
		t.Errorf("long edge should pierce 2 faces, got %d", straddleCount)
	}
}

// ---------------------------------------------------------------------------
// TestGenerate_NearMissEdge: edge passes within 1e-8 of a face edge.
// The tolerance determines if this is EdgeHit or FaceHit.
// ---------------------------------------------------------------------------

func TestGenerate_NearMissEdge(t *testing.T) {
	cube := MakeCube(1.0, testRed)
	cube.CalcPlaneEquations()

	topFace := faceWithNormal(cube, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if topFace == nil {
		t.Fatal("could not find +Z face")
		return
	}

	topZ := topFace.GetFirstHe().Vertex.Loc.Z

	// Find edge midpoint and offset by 1e-8 into face interior
	he := topFace.GetFirstHe()
	edgeMid := he.Vertex.Loc.Add(he.Next.Vertex.Loc).Scale(0.5)
	edgeDir := he.Next.Vertex.Loc.Sub(he.Vertex.Loc).Normalize()
	inward := topFace.Normal.Cross(edgeDir).Normalize()
	nearPt := edgeMid.Add(inward.Scale(1e-8))

	// This point is ON the face plane, just barely inside
	above := vec.SFVec3f{X: nearPt.X, Y: nearPt.Y, Z: topZ + 0.5}
	below := vec.SFVec3f{X: nearPt.X, Y: nearPt.Y, Z: topZ - 0.5}

	d1 := topFace.GetDistance(above)
	d2 := topFace.GetDistance(below)
	tParam := d1 / (d1 - d2)
	pt := above.Add(below.Sub(above).Scale(tParam))

	testV := NewVertexVec(pt)
	rec := NewIntersectRecord()
	hit := topFace.FaceContains(testV, &rec)

	if !hit {
		t.Fatal("near-edge intersection should be contained")
	}
	// Log what we get — either EdgeHit or FaceHit is acceptable here
	t.Logf("near-miss (1e-8 from edge): hit type = %d", rec.Type)
}

// ---------------------------------------------------------------------------
// TestGenerate_CubeVsCube_Disjoint: two non-overlapping cubes.
// No edges of A straddle any faces of B, and vice versa.
// ---------------------------------------------------------------------------

func TestGenerate_CubeVsCube_Disjoint(t *testing.T) {
	a := MakeCube(1.0, testRed)
	b := MakeCube(1.0, testGreen)
	b.TransformGeometry(vec.TranslationMatrix(5, 0, 0))

	a.CalcPlaneEquations()
	b.CalcPlaneEquations()

	// Check that no edge of A straddles any face of B
	straddleCount := 0
	for e := a.Edges; e != nil; e = e.Next {
		v1 := e.He1.Vertex.Loc
		v2 := e.He2.Vertex.Loc
		for f := b.Faces; f != nil; f = f.Next {
			d1 := f.GetDistance(v1)
			d2 := f.GetDistance(v2)
			s1 := floatCompare(d1, 0)
			s2 := floatCompare(d2, 0)
			if (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0) {
				straddleCount++
			}
		}
	}

	if straddleCount != 0 {
		t.Errorf("disjoint cubes should have 0 straddling edges, got %d", straddleCount)
	}
}

// ---------------------------------------------------------------------------
// TestGenerate_CubeVsCube_PartialOverlap: BOOL01 case 2 geometry.
// Count expected straddle intersections.
// ---------------------------------------------------------------------------

func TestGenerate_CubeVsCube_PartialOverlap(t *testing.T) {
	a := MakeCube(1.0, testRed)
	b := MakeCube(0.5, testGreen)
	b.TransformGeometry(vec.ScaleMatrix(1, 1, 2))
	b.TransformGeometry(vec.TranslationMatrix(0.25, 0.25, -0.25))

	a.CalcPlaneEquations()
	b.CalcPlaneEquations()

	// Count A-edges straddling B-faces with actual containment
	hitCount := 0
	for e := a.Edges; e != nil; e = e.Next {
		v1 := e.He1.Vertex.Loc
		v2 := e.He2.Vertex.Loc
		for f := b.Faces; f != nil; f = f.Next {
			d1 := f.GetDistance(v1)
			d2 := f.GetDistance(v2)
			s1 := floatCompare(d1, 0)
			s2 := floatCompare(d2, 0)
			if (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0) {
				tParam := d1 / (d1 - d2)
				pt := v1.Add(v2.Sub(v1).Scale(tParam))
				testV := NewVertexVec(pt)
				rec := NewIntersectRecord()
				if f.FaceContains(testV, &rec) {
					hitCount++
				}
			}
		}
	}

	// Also count B-edges straddling A-faces
	for e := b.Edges; e != nil; e = e.Next {
		v1 := e.He1.Vertex.Loc
		v2 := e.He2.Vertex.Loc
		for f := a.Faces; f != nil; f = f.Next {
			d1 := f.GetDistance(v1)
			d2 := f.GetDistance(v2)
			s1 := floatCompare(d1, 0)
			s2 := floatCompare(d2, 0)
			if (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0) {
				tParam := d1 / (d1 - d2)
				pt := v1.Add(v2.Sub(v1).Scale(tParam))
				testV := NewVertexVec(pt)
				rec := NewIntersectRecord()
				if f.FaceContains(testV, &rec) {
					hitCount++
				}
			}
		}
	}

	if hitCount == 0 {
		t.Error("partial overlap should produce intersection points")
	}
	t.Logf("partial overlap: %d edge-face intersection hits", hitCount)
}

// ---------------------------------------------------------------------------
// TestGenerate_CubeVsCube_FaceOnFace: BOOL04 case 0 (face-on-face).
// No edges straddle — all contacts are coplanar.
// ---------------------------------------------------------------------------

func TestGenerate_CubeVsCube_FaceOnFace(t *testing.T) {
	a := MakeCube(1.0, testRed)
	b := MakeCube(1.0, testGreen)
	b.TransformGeometry(vec.TranslationMatrix(0, 0, -2.0))

	a.CalcPlaneEquations()
	b.CalcPlaneEquations()

	// Count straddling edges
	straddleCount := 0
	onCount := 0

	for e := a.Edges; e != nil; e = e.Next {
		v1 := e.He1.Vertex.Loc
		v2 := e.He2.Vertex.Loc
		for f := b.Faces; f != nil; f = f.Next {
			d1 := f.GetDistance(v1)
			d2 := f.GetDistance(v2)
			s1 := floatCompare(d1, 0)
			s2 := floatCompare(d2, 0)
			if (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0) {
				straddleCount++
			}
			if s1 == 0 {
				onCount++
			}
			if s2 == 0 {
				onCount++
			}
		}
	}

	// Face-on-face: shared face plane means some vertices are ON some faces of B.
	// But no edges should straddle — the Generate phase should produce
	// vertex-on-face records (via VertexOnFace path), not edge splits.
	t.Logf("face-on-face: %d straddling edges, %d ON-vertex cases", straddleCount, onCount)
	if onCount == 0 {
		t.Error("face-on-face should have ON-vertex cases")
	}
}

// ---------------------------------------------------------------------------
// TestGenerate_CubeVsCube_Identical: two cubes at same position.
// Everything is coplanar. No straddling edges.
// ---------------------------------------------------------------------------

func TestGenerate_CubeVsCube_Identical(t *testing.T) {
	a := MakeCube(1.0, testRed)
	b := MakeCube(1.0, testGreen)

	a.CalcPlaneEquations()
	b.CalcPlaneEquations()

	straddleCount := 0
	for e := a.Edges; e != nil; e = e.Next {
		v1 := e.He1.Vertex.Loc
		v2 := e.He2.Vertex.Loc
		for f := b.Faces; f != nil; f = f.Next {
			d1 := f.GetDistance(v1)
			d2 := f.GetDistance(v2)
			s1 := floatCompare(d1, 0)
			s2 := floatCompare(d2, 0)
			if (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0) {
				straddleCount++
			}
		}
	}

	if straddleCount != 0 {
		t.Errorf("identical cubes should have 0 straddling edges, got %d", straddleCount)
	}
}

// ---------------------------------------------------------------------------
// TestGenerate_CubeVsSphere_ManyHits: sphere inside cube, many intersections.
// Exercises high intersection count.
// ---------------------------------------------------------------------------

func TestGenerate_CubeVsSphere_ManyHits(t *testing.T) {
	a := MakeCube(1.0, testRed)
	// Sphere radius 1.2 protrudes past cube faces (cube vertices at ±1.0)
	b := MakeSphere(1.2, 16, 16, testGreen)

	a.CalcPlaneEquations()
	b.CalcPlaneEquations()

	// Count B-edges straddling A-faces
	hitCount := 0
	for e := b.Edges; e != nil; e = e.Next {
		v1 := e.He1.Vertex.Loc
		v2 := e.He2.Vertex.Loc
		for f := a.Faces; f != nil; f = f.Next {
			d1 := f.GetDistance(v1)
			d2 := f.GetDistance(v2)
			s1 := floatCompare(d1, 0)
			s2 := floatCompare(d2, 0)
			if (s1 > 0 && s2 < 0) || (s1 < 0 && s2 > 0) {
				tParam := d1 / (d1 - d2)
				pt := v1.Add(v2.Sub(v1).Scale(tParam))
				testV := NewVertexVec(pt)
				rec := NewIntersectRecord()
				if f.FaceContains(testV, &rec) {
					hitCount++
				}
			}
		}
	}

	t.Logf("sphere(16x16) vs cube: %d intersection hits", hitCount)
	if hitCount == 0 {
		t.Error("sphere inside cube should have many intersection points")
	}
}

// ---------------------------------------------------------------------------
// TestInterpolation_Precision: verify interpolation accuracy for various t values.
// The core computation: pt = v1 + t*(v2-v1) where t = d1/(d1-d2).
// ---------------------------------------------------------------------------

func TestInterpolation_Precision(t *testing.T) {
	// Test with large coordinate values (catastrophic cancellation risk)
	cases := []struct {
		name      string
		v1, v2    vec.SFVec3f
		planeNorm vec.SFVec3f
		planeD    float64
	}{
		{
			"unit_scale",
			vec.SFVec3f{X: 0, Y: 0, Z: -0.5},
			vec.SFVec3f{X: 0, Y: 0, Z: 0.5},
			vec.SFVec3f{X: 0, Y: 0, Z: 1},
			0, // plane at z=0
		},
		{
			"far_from_origin",
			vec.SFVec3f{X: 1e6, Y: 1e6, Z: 1e6 - 0.5},
			vec.SFVec3f{X: 1e6, Y: 1e6, Z: 1e6 + 0.5},
			vec.SFVec3f{X: 0, Y: 0, Z: 1},
			-1e6, // plane at z=1e6
		},
		{
			"near_endpoint",
			vec.SFVec3f{X: 0, Y: 0, Z: -0.001},
			vec.SFVec3f{X: 0, Y: 0, Z: 1.0},
			vec.SFVec3f{X: 0, Y: 0, Z: 1},
			0, // plane at z=0; t≈0.001
		},
		{
			"large_coordinates",
			vec.SFVec3f{X: 1e4, Y: 1e4, Z: -500},
			vec.SFVec3f{X: 1e4, Y: 1e4, Z: 500},
			vec.SFVec3f{X: 0, Y: 0, Z: 1},
			0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d1 := c.planeNorm.Dot(c.v1) + c.planeD
			d2 := c.planeNorm.Dot(c.v2) + c.planeD

			if (d1 > 0) == (d2 > 0) {
				t.Skip("endpoints don't straddle")
			}

			tParam := d1 / (d1 - d2)
			pt := c.v1.Add(c.v2.Sub(c.v1).Scale(tParam))

			// The intersection point should lie on the plane
			dist := c.planeNorm.Dot(pt) + c.planeD
			if math.Abs(dist) > 1e-10 {
				t.Errorf("intersection point off plane by %e", dist)
			}

			// t should be in [0,1]
			if tParam < 0 || tParam > 1 {
				t.Errorf("t=%f out of [0,1] range", tParam)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestContainsOnEdge_NearEndpoints: edge containment at extreme t values.
// ---------------------------------------------------------------------------

func TestContainsOnEdge_NearEndpoints(t *testing.T) {
	v1 := vec.SFVec3f{X: 0, Y: 0, Z: 0}
	v2 := vec.SFVec3f{X: 1, Y: 0, Z: 0}

	cases := []struct {
		name   string
		point  vec.SFVec3f
		expect bool
	}{
		{"at_v1", vec.SFVec3f{X: 0, Y: 0, Z: 0}, true},
		{"at_v2", vec.SFVec3f{X: 1, Y: 0, Z: 0}, true},
		{"midpoint", vec.SFVec3f{X: 0.5, Y: 0, Z: 0}, true},
		{"near_v1", vec.SFVec3f{X: 1e-8, Y: 0, Z: 0}, true},
		{"near_v2", vec.SFVec3f{X: 1 - 1e-8, Y: 0, Z: 0}, true},
		{"just_past_v1", vec.SFVec3f{X: -0.001, Y: 0, Z: 0}, false},
		{"just_past_v2", vec.SFVec3f{X: 1.001, Y: 0, Z: 0}, false},
		{"off_edge", vec.SFVec3f{X: 0.5, Y: 0.1, Z: 0}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ContainsOnEdge(v1, v2, c.point)
			if got != c.expect {
				t.Errorf("ContainsOnEdge(%v)=%v, want %v", c.point, got, c.expect)
			}
		})
	}
}
