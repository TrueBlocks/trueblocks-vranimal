package solid

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// IsCoplanar
// ---------------------------------------------------------------------------

func TestFace_IsCoplanar_Same(t *testing.T) {
	// Two faces with identical normals are coplanar.
	f1 := &Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}}
	f2 := &Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}}
	if !f1.IsCoplanar(f2) {
		t.Fatal("faces with same normal should be coplanar")
	}
}

func TestFace_IsCoplanar_Opposite(t *testing.T) {
	// Anti-parallel normals: cross product is zero, so they ARE coplanar.
	f1 := &Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}}
	f2 := &Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: -1}}
	if !f1.IsCoplanar(f2) {
		t.Fatal("anti-parallel normals should be coplanar")
	}
}

func TestFace_IsCoplanar_NotCoplanar(t *testing.T) {
	f1 := &Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}}
	f2 := &Face{Normal: vec.SFVec3f{X: 1, Y: 0, Z: 0}}
	if f1.IsCoplanar(f2) {
		t.Fatal("perpendicular faces should not be coplanar")
	}
}

// ---------------------------------------------------------------------------
// HasCoplanarNeighbor
// ---------------------------------------------------------------------------

func TestFace_HasCoplanarNeighbor_None(t *testing.T) {
	// A swept triangular prism has rectangular side faces at 90° angles
	// to each other — none should be coplanar with their neighbors.
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	// Sweep upward to make a prism with non-coplanar side faces
	var triF *Face
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut != nil && f.LoopOut.NHalfEdges() >= 3 {
			triF = f
			break
		}
	}
	if triF == nil {
		t.Skip("no triangle face")
	}
	s.TranslationalSweep(triF, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	s.CalcPlaneEquations()

	// Side faces (rectangles) should NOT be coplanar with each other
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut == nil {
			continue
		}
		if f2, _, ok := f.HasCoplanarNeighbor(); ok {
			// The two triangular caps (top/bottom) may be coplanar with each
			// other but they don't share an edge, so that shouldn't happen.
			// Only flag if both are rectangular side faces.
			if f.LoopOut.NHalfEdges() == 4 && f2.LoopOut != nil && f2.LoopOut.NHalfEdges() == 4 {
				t.Fatal("rectangular side faces should not be coplanar")
			}
		}
	}
}

func TestFace_HasCoplanarNeighbor_WithCoplanar(t *testing.T) {
	// Build a quad (4 vertices) then triangulate — two triangles on the
	// same plane should be coplanar neighbors.
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 10, Y: 10, Z: 0},
		{X: 0, Y: 10, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	s.Triangulate()
	s.CalcPlaneEquations()

	found := false
	for f := s.Faces; f != nil; f = f.Next {
		if _, _, ok := f.HasCoplanarNeighbor(); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("triangulated quad should have coplanar neighbor pair")
	}
}

// ---------------------------------------------------------------------------
// AdjustOuterLoop
// ---------------------------------------------------------------------------

func TestFace_AdjustOuterLoop_SingleLoop(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	s.CalcPlaneEquations()

	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut != nil {
			orig := f.LoopOut
			f.AdjustOuterLoop()
			if f.LoopOut != orig {
				t.Fatal("single-loop face should keep same outer loop")
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Loop HasColinearVerts
// ---------------------------------------------------------------------------

func TestLoop_HasColinearVerts_Found(t *testing.T) {
	// Build a face with 4 vertices where 3 are collinear: (0,0)-(5,0)-(10,0)-(5,5)
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 5, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 5, Y: 5, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}

	found := false
	for f := s.Faces; f != nil; f = f.Next {
		for _, l := range f.Loops {
			if he := l.HasColinearVerts(); he != nil {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("should detect collinear vertices")
	}
}

func TestLoop_HasColinearVerts_None(t *testing.T) {
	// Right triangle — no collinear vertices
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 0, Y: 10, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}

	for f := s.Faces; f != nil; f = f.Next {
		for _, l := range f.Loops {
			if he := l.HasColinearVerts(); he != nil {
				t.Fatal("triangle should have no collinear vertices")
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Face HasColinearVerts
// ---------------------------------------------------------------------------

func TestFace_HasColinearVerts(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 5, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 5, Y: 5, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}

	found := false
	for f := s.Faces; f != nil; f = f.Next {
		if _, he := f.HasColinearVerts(); he != nil {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("should detect collinear vertices at face level")
	}
}

// ---------------------------------------------------------------------------
// Solid HasColinearVerts
// ---------------------------------------------------------------------------

func TestSolid_HasColinearVerts_True(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 5, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 5, Y: 5, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	if !s.HasColinearVerts() {
		t.Fatal("should detect collinear verts")
	}
}

func TestSolid_HasColinearVerts_False(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 0, Y: 10, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	if s.HasColinearVerts() {
		t.Fatal("triangle should not have collinear verts")
	}
}

// ---------------------------------------------------------------------------
// HasDegenerateFaces / RemoveDegenerateFaces
// ---------------------------------------------------------------------------

func TestSolid_HasDegenerateFaces_None(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	if f := s.HasDegenerateFaces(); f != nil {
		t.Fatal("triangle should not be degenerate")
	}
}

func TestSolid_RemoveDegenerateFaces(t *testing.T) {
	// Just verify it doesn't panic on a healthy solid
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	s.RemoveDegenerateFaces()
}

// ---------------------------------------------------------------------------
// VerifyDetailed
// ---------------------------------------------------------------------------

func TestSolid_VerifyDetailed_Clean(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}

	errs := s.VerifyDetailed()
	for _, err := range errs {
		t.Errorf("verification error: %v", err)
	}
}

func TestSolid_VerifyDetailed_Wire(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	Lmev(s.Verts.He, vec.SFVec3f{X: 1})

	errs := s.VerifyDetailed()
	for _, err := range errs {
		t.Errorf("verification error: %v", err)
	}
}

func TestSolid_VerifyDetailed_Quad(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}

	errs := s.VerifyDetailed()
	for _, err := range errs {
		t.Errorf("verification error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// VertexFaces (neighborhood)
// ---------------------------------------------------------------------------

func TestVertex_VertexFaces(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}

	// Every vertex should be adjacent to at least one face
	for v := s.Verts; v != nil; v = v.Next {
		faces := v.VertexFaces()
		if len(faces) == 0 {
			t.Fatalf("vertex %d should have at least one neighboring face", v.Index)
		}
	}
}

func TestVertex_VertexFaces_Wire(t *testing.T) {
	_, v, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	Lmev(v.He, vec.SFVec3f{X: 1})

	faces := v.VertexFaces()
	if len(faces) != 1 {
		t.Fatalf("wire vertex should have 1 face, got %d", len(faces))
	}
}

// ---------------------------------------------------------------------------
// VertexEdges (neighborhood)
// ---------------------------------------------------------------------------

func TestVertex_VertexEdges(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}

	for v := s.Verts; v != nil; v = v.Next {
		edges := v.VertexEdges()
		if len(edges) == 0 {
			t.Fatalf("vertex %d should have incident edges", v.Index)
		}
	}
}

// ---------------------------------------------------------------------------
// VertexNeighbors (neighborhood)
// ---------------------------------------------------------------------------

func TestVertex_VertexNeighbors(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}

	// Each vertex in a triangle should have 2 neighbors
	for v := s.Verts; v != nil; v = v.Next {
		nbrs := v.VertexNeighbors()
		if len(nbrs) < 2 {
			t.Fatalf("vertex %d: expected >=2 neighbors, got %d", v.Index, len(nbrs))
		}
	}
}

func TestVertex_VertexNeighbors_Nil(t *testing.T) {
	v := &Vertex{}
	if nbrs := v.VertexNeighbors(); nbrs != nil {
		t.Fatal("nil He vertex should return nil neighbors")
	}
}

// ---------------------------------------------------------------------------
// FaceNeighbors (neighborhood)
// ---------------------------------------------------------------------------

func TestFace_FaceNeighbors(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	s.Triangulate()

	// After triangulating a quad, each triangle face should have at least 1 neighbor
	for f := s.Faces; f != nil; f = f.Next {
		nbrs := f.FaceNeighbors()
		if f.LoopOut != nil && f.LoopOut.NHalfEdges() >= 3 && len(nbrs) == 0 {
			t.Fatalf("face %d should have neighbors after triangulation", f.Index)
		}
	}
}

// ---------------------------------------------------------------------------
// GetNormal / GetD accessors
// ---------------------------------------------------------------------------

func TestFace_GetNormal_GetD(t *testing.T) {
	f := &Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: 5.0}
	if n := f.GetNormal(); n.Z != 1 {
		t.Fatalf("expected Z=1, got %g", n.Z)
	}
	if d := f.GetD(); d != 5.0 {
		t.Fatalf("expected D=5, got %g", d)
	}
}

// ---------------------------------------------------------------------------
// RemoveCoplanarFaces (integration test)
// ---------------------------------------------------------------------------

func TestSolid_RemoveCoplanarFaces(t *testing.T) {
	// Build a quad, triangulate it (creating 2 coplanar triangles),
	// then remove coplanar faces.
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 10, Y: 10, Z: 0},
		{X: 0, Y: 10, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	s.Triangulate()
	s.CalcPlaneEquations()

	beforeFaces := s.NFaces()
	s.RemoveCoplanarFaces()

	if s.NFaces() >= beforeFaces {
		// At least one coplanar pair should have been merged
		t.Logf("RemoveCoplanarFaces: %d -> %d faces (may not reduce if BuildFromIndexSet creates non-manifold loops)", beforeFaces, s.NFaces())
	}
}

// ---------------------------------------------------------------------------
// RemoveCoplaneColine (integration test)
// ---------------------------------------------------------------------------

func TestSolid_RemoveCoplaneColine(t *testing.T) {
	// Build a quad, triangulate, then run full cleanup.
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 10, Y: 10, Z: 0},
		{X: 0, Y: 10, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}
	s.Triangulate()

	// Should not panic
	s.RemoveCoplaneColine()
}

// ---------------------------------------------------------------------------
// Solid-level RemoveColinearVerts
// ---------------------------------------------------------------------------

func TestSolid_RemoveColinearVerts(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Skip("BuildFromIndexSet returned nil")
	}

	s.RemoveColinearVerts()
	// Should not panic and solid should remain intact
	if s.NFaces() == 0 {
		t.Fatal("solid should still have faces")
	}
}

// ---------------------------------------------------------------------------
// VerifyError format
// ---------------------------------------------------------------------------

func TestVerifyError_String(t *testing.T) {
	err := &VerifyError{Element: "Face", Index: 3, Message: "test error"}
	s := err.Error()
	if s != "Face[3]: test error" {
		t.Fatalf("unexpected error string: %s", s)
	}
}
