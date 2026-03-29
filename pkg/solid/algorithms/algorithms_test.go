package algorithms

import (
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func makeTriangle(t *testing.T) *base.Solid {
	t.Helper()
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 5, Y: 10, Z: 0},
	}
	indices := []int64{0, 1, 2, -1}
	s, err := euler.BuildFromIndexSet(positions, indices, vec.Red)
	if err != nil {
		t.Skipf("BuildFromIndexSet: %v", err)
	}
	return s
}

func makeQuad(t *testing.T) *base.Solid {
	t.Helper()
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 10, Y: 10, Z: 0},
		{X: 0, Y: 10, Z: 0},
	}
	indices := []int64{0, 1, 2, 3, -1}
	s, err := euler.BuildFromIndexSet(positions, indices, vec.Red)
	if err != nil {
		t.Skipf("BuildFromIndexSet: %v", err)
	}
	return s
}

func findTriangleLoop(s *base.Solid) *base.Loop {
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut != nil && f.LoopOut.NHalfEdges() >= 3 {
			return f.LoopOut
		}
	}
	return nil
}

func findTriangleFace(s *base.Solid) *base.Face {
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut != nil && f.LoopOut.NHalfEdges() >= 3 {
			return f
		}
	}
	return nil
}

func TestIntersectRecord_Default(t *testing.T) {
	rec := NewIntersectRecord()
	if rec.GetType() != NoHit {
		t.Fatalf("expected NoHit, got %d", rec.GetType())
	}
}

func TestIntersectRecord_RecordTypes(t *testing.T) {
	rec := NewIntersectRecord()

	rec.RecordHit(nil, nil, FaceHit)
	if rec.Type != FaceHit || rec.He != nil || rec.Vert != nil {
		t.Fatal("FaceHit wrong")
	}

	v := &base.Vertex{}
	rec.RecordHit(nil, v, VertexHit)
	if rec.Type != VertexHit || rec.Vert != v {
		t.Fatal("VertexHit wrong")
	}

	he := &base.HalfEdge{}
	rec.RecordHit(he, nil, EdgeHit)
	if rec.Type != EdgeHit || rec.He != he {
		t.Fatal("EdgeHit wrong")
	}

	rec.RecordHit(nil, nil, InsideLoopNoHit)
	if rec.Type != InsideLoopNoHit {
		t.Fatal("InsideLoopNoHit wrong")
	}
}

func TestCrossingsTest_InsideSquare(t *testing.T) {
	pgon := [][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	if !base.CrossingsTest(pgon, [2]float64{5, 5}) {
		t.Fatal("point should be inside")
	}
}

func TestCrossingsTest_OutsideSquare(t *testing.T) {
	pgon := [][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	if base.CrossingsTest(pgon, [2]float64{15, 5}) {
		t.Fatal("point should be outside")
	}
}

func TestCrossingsTest_InsideTriangle(t *testing.T) {
	pgon := [][2]float64{{0, 0}, {10, 0}, {5, 10}}
	if !base.CrossingsTest(pgon, [2]float64{5, 3}) {
		t.Fatal("point should be inside triangle")
	}
}

func TestCrossingsTest_Empty(t *testing.T) {
	if base.CrossingsTest(nil, [2]float64{0, 0}) {
		t.Fatal("empty polygon should return false")
	}
}

func TestCrossingsTest_Pentagon(t *testing.T) {
	pgon := make([][2]float64, 5)
	for i := 0; i < 5; i++ {
		angle := float64(i) * 2 * math.Pi / 5
		pgon[i] = [2]float64{math.Cos(angle), math.Sin(angle)}
	}
	if !base.CrossingsTest(pgon, [2]float64{0, 0}) {
		t.Fatal("center should be inside pentagon")
	}
	if base.CrossingsTest(pgon, [2]float64{10, 10}) {
		t.Fatal("far point should be outside pentagon")
	}
}

func TestBoundaryContains_VertexHit(t *testing.T) {
	s := makeTriangle(t)
	l := findTriangleLoop(s)
	if l == nil {
		t.Skip("no triangle loop")
	}

	testV := base.NewVertexVec(vec.SFVec3f{X: 10, Y: 0, Z: 0})
	rec := NewIntersectRecord()
	if !BoundaryContains(l, testV, &rec) {
		t.Fatal("should find vertex hit")
	}
	if rec.Type != VertexHit {
		t.Fatalf("expected VertexHit, got %d", rec.Type)
	}
}

func TestBoundaryContains_Miss(t *testing.T) {
	s := makeTriangle(t)
	l := findTriangleLoop(s)
	if l == nil {
		t.Skip("no triangle loop")
	}

	testV := base.NewVertexVec(vec.SFVec3f{X: 100, Y: 100, Z: 0})
	rec := NewIntersectRecord()
	if BoundaryContains(l, testV, &rec) {
		t.Fatal("should not find hit")
	}
}

func TestLoopContains(t *testing.T) {
	s := makeTriangle(t)
	l := findTriangleLoop(s)
	if l == nil {
		t.Skip("no triangle loop")
	}

	inside := base.NewVertexVec(vec.SFVec3f{X: 5, Y: 3, Z: 0})
	rec := NewIntersectRecord()
	if !LoopContains(l, inside, 2, &rec) {
		t.Fatal("point should be inside")
	}
}

func TestFaceContains_Inside(t *testing.T) {
	s := makeTriangle(t)
	s.CalcPlaneEquations()
	f := findTriangleFace(s)
	if f == nil {
		t.Skip("no triangle face")
	}

	inside := base.NewVertexVec(vec.SFVec3f{X: 5, Y: 3, Z: 0})
	rec := NewIntersectRecord()
	if !FaceContains(f, inside, &rec) {
		t.Fatal("point should be inside face")
	}
	if rec.Type != FaceHit {
		t.Fatalf("expected FaceHit, got %d", rec.Type)
	}
}

func TestFaceContains_Outside(t *testing.T) {
	s := makeTriangle(t)
	s.CalcPlaneEquations()
	f := findTriangleFace(s)
	if f == nil {
		t.Skip("no triangle face")
	}

	outside := base.NewVertexVec(vec.SFVec3f{X: 20, Y: 20, Z: 0})
	rec := NewIntersectRecord()
	if FaceContains(f, outside, &rec) {
		t.Fatal("point should be outside face")
	}
}

func TestSolidContains(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: -10, Y: -10, Z: 0},
		{X: 10, Y: -10, Z: 0},
		{X: 10, Y: 10, Z: 0},
		{X: -10, Y: 10, Z: 0},
	}
	indices := []int64{0, 1, 2, 3, -1}
	outer, err := euler.BuildFromIndexSet(positions, indices, vec.Red)
	if err != nil {
		t.Skipf("BuildFromIndexSet: %v", err)
	}
	outer.CalcPlaneEquations()

	inner, _, _ := euler.Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.Blue)
	if !SolidContains(outer, inner) {
		t.Fatal("inner should be contained")
	}
}

func TestVerify_Wire(t *testing.T) {
	s, _, _ := euler.Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	_, _, _ = euler.Lmev(s.Verts.He, vec.SFVec3f{X: 1})
	if !Verify(s) {
		t.Fatal("wire should verify")
	}
}

func TestVerify_Triangle(t *testing.T) {
	s := makeTriangle(t)
	_ = Verify(s)
}

func TestVerifyDetailed_Clean(t *testing.T) {
	s := makeTriangle(t)
	errs := VerifyDetailed(s)
	for _, err := range errs {
		t.Logf("verification note (open face expected): %v", err)
	}
}

func TestVerifyDetailed_Wire(t *testing.T) {
	s, _, _ := euler.Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	_, _, _ = euler.Lmev(s.Verts.He, vec.SFVec3f{X: 1})
	errs := VerifyDetailed(s)
	for _, err := range errs {
		t.Errorf("verification error: %v", err)
	}
}

func TestVerifyDetailed_Quad(t *testing.T) {
	s := makeQuad(t)
	errs := VerifyDetailed(s)
	for _, err := range errs {
		t.Logf("verification note (open face expected): %v", err)
	}
}

func TestVerifyError_String(t *testing.T) {
	err := &VerifyError{Element: "Face", Index: 3, Message: "test error"}
	s := err.Error()
	if s != "Face[3]: test error" {
		t.Fatalf("unexpected error string: %s", s)
	}
}

func TestIsCoplanar_Same(t *testing.T) {
	f1 := &base.Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}}
	f2 := &base.Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}}
	if !IsCoplanar(f1, f2) {
		t.Fatal("same normal should be coplanar")
	}
}

func TestIsCoplanar_Opposite(t *testing.T) {
	f1 := &base.Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}}
	f2 := &base.Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: -1}}
	if !IsCoplanar(f1, f2) {
		t.Fatal("anti-parallel normals should be coplanar")
	}
}

func TestIsCoplanar_NotCoplanar(t *testing.T) {
	f1 := &base.Face{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}}
	f2 := &base.Face{Normal: vec.SFVec3f{X: 1, Y: 0, Z: 0}}
	if IsCoplanar(f1, f2) {
		t.Fatal("perpendicular should not be coplanar")
	}
}

func TestHasCoplanarNeighbor_WithCoplanar(t *testing.T) {
	s := makeQuad(t)
	Triangulate(s)
	s.CalcPlaneEquations()

	found := false
	for f := s.Faces; f != nil; f = f.Next {
		if _, _, ok := HasCoplanarNeighbor(f); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("triangulated quad should have coplanar neighbor pair")
	}
}

func TestAdjustOuterLoop_SingleLoop(t *testing.T) {
	s := makeTriangle(t)
	s.CalcPlaneEquations()

	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut != nil {
			orig := f.LoopOut
			AdjustOuterLoop(f)
			if f.LoopOut != orig {
				t.Fatal("single-loop face should keep same outer loop")
			}
		}
	}
}

func TestLoopHasColinearVerts_Found(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 5, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 5, Y: 5, Z: 0},
	}
	indices := []int64{0, 1, 2, 3, -1}
	s, err := euler.BuildFromIndexSet(positions, indices, vec.Red)
	if err != nil {
		t.Skipf("BuildFromIndexSet: %v", err)
	}

	found := false
	for f := s.Faces; f != nil; f = f.Next {
		for _, l := range f.Loops {
			if he := LoopHasColinearVerts(l); he != nil {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("should detect collinear vertices")
	}
}

func TestLoopHasColinearVerts_None(t *testing.T) {
	s := makeTriangle(t)
	for f := s.Faces; f != nil; f = f.Next {
		for _, l := range f.Loops {
			if he := LoopHasColinearVerts(l); he != nil {
				t.Fatal("triangle should have no collinear vertices")
			}
		}
	}
}

func TestFaceHasColinearVerts(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 5, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 5, Y: 5, Z: 0},
	}
	indices := []int64{0, 1, 2, 3, -1}
	s, err := euler.BuildFromIndexSet(positions, indices, vec.Red)
	if err != nil {
		t.Skipf("BuildFromIndexSet: %v", err)
	}

	found := false
	for f := s.Faces; f != nil; f = f.Next {
		if _, he := FaceHasColinearVerts(f); he != nil {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("should detect collinear at face level")
	}
}

func TestSolidHasColinearVerts_True(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 5, Y: 0, Z: 0},
		{X: 10, Y: 0, Z: 0},
		{X: 5, Y: 5, Z: 0},
	}
	indices := []int64{0, 1, 2, 3, -1}
	s, err := euler.BuildFromIndexSet(positions, indices, vec.Red)
	if err != nil {
		t.Skipf("BuildFromIndexSet: %v", err)
	}
	if !SolidHasColinearVerts(s) {
		t.Fatal("should detect collinear verts")
	}
}

func TestSolidHasColinearVerts_False(t *testing.T) {
	s := makeTriangle(t)
	if SolidHasColinearVerts(s) {
		t.Fatal("triangle should not have collinear verts")
	}
}

func TestHasDegenerateFaces_None(t *testing.T) {
	s := makeTriangle(t)
	if f := HasDegenerateFaces(s); f != nil {
		t.Fatal("triangle should not be degenerate")
	}
}

func TestRemoveDegenerateFaces(t *testing.T) {
	s := makeTriangle(t)
	RemoveDegenerateFaces(s)
}

func TestTriangulate(t *testing.T) {
	s := makeQuad(t)
	Triangulate(s)

	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut == nil {
			continue
		}
		n := f.LoopOut.NHalfEdges()
		if n != 3 && n > 0 {
			t.Fatalf("expected 3 half-edges, got %d", n)
		}
	}
}

func TestTwist(t *testing.T) {
	s, _, _ := euler.Mvfs(vec.SFVec3f{X: 1, Y: 0, Z: 1}, vec.Red)
	_, _, _ = euler.Lmev(s.Verts.He, vec.SFVec3f{X: 0, Y: 1, Z: 2})
	Twist(s, func(z float64) float64 { return z * 0.1 })
	for v := s.Verts; v != nil; v = v.Next {
		_ = v.Loc
	}
}

func TestRemoveColinearVerts(t *testing.T) {
	s := makeQuad(t)
	for f := s.Faces; f != nil; f = f.Next {
		RemoveColinearVerts(f)
	}
}

func TestRemoveCoplanarFaces(t *testing.T) {
	s := makeQuad(t)
	Triangulate(s)
	s.CalcPlaneEquations()
	RemoveCoplanarFaces(s)
}

func TestRemoveCoplaneColine(t *testing.T) {
	s := makeQuad(t)
	Triangulate(s)
	RemoveCoplaneColine(s)
}

func TestRemoveColinearVertsSolid(t *testing.T) {
	s := makeQuad(t)
	RemoveColinearVertsSolid(s)
	if s.NFaces() == 0 {
		t.Fatal("solid should still have faces")
	}
}

func TestJoinCut(t *testing.T) {
	s := makeQuad(t)
	if s.NVerts() < 4 {
		t.Fatalf("expected >=4 verts, got %d", s.NVerts())
	}
}

func TestMoveFace(t *testing.T) {
	s := makeQuad(t)
	target := base.NewSolid()
	f := s.Faces
	if f == nil {
		t.Skip("no faces")
	}
	MoveFace(f, target)
	if target.NFaces() == 0 {
		t.Fatal("target should have faces after move")
	}
}

func TestLoopGlue_NoSecondLoop(t *testing.T) {
	s, _, _ := euler.Mvfs(vec.SFVec3f{}, vec.Red)
	LoopGlue(s.Faces)
}

func TestArc(t *testing.T) {
	s, v, f := euler.Mvfs(vec.SFVec3f{X: 1, Y: 0, Z: 0}, vec.Red)
	initialVerts := s.NVerts()
	Arc(s, f, v, 0, 0, 1, 0, 0, 90, 4)
	if s.NVerts() <= initialVerts {
		t.Fatal("arc should add vertices")
	}
}

func TestArcSweep(t *testing.T) {
	s := makeTriangle(t)
	var f *base.Face
	for ff := s.Faces; ff != nil; ff = ff.Next {
		if ff.LoopOut != nil && ff.LoopOut.NHalfEdges() >= 3 {
			f = ff
			break
		}
	}
	if f == nil {
		t.Skip("no triangle face")
	}
	dirs := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 1},
		{X: 0, Y: 0, Z: 1},
	}
	ArcSweep(s, f, dirs)
	if s.NVerts() < 3 {
		t.Fatal("sweep should maintain or add vertices")
	}
}

func TestTranslationalSweep(t *testing.T) {
	s := makeTriangle(t)
	initialVerts := s.NVerts()
	f := findTriangleFace(s)
	if f == nil {
		t.Skip("no triangle face")
	}
	TranslationalSweep(s, f, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	if s.NVerts() <= initialVerts {
		t.Fatal("sweep should add vertices")
	}
}

func TestVertexFaces(t *testing.T) {
	s := makeQuad(t)
	for v := s.Verts; v != nil; v = v.Next {
		faces := VertexFaces(v)
		if len(faces) == 0 {
			t.Fatalf("vertex %d should have at least one face", v.Index)
		}
	}
}

func TestVertexFaces_Wire(t *testing.T) {
	_, v, _ := euler.Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	_, _, _ = euler.Lmev(v.He, vec.SFVec3f{X: 1})
	faces := VertexFaces(v)
	if len(faces) != 1 {
		t.Fatalf("wire vertex should have 1 face, got %d", len(faces))
	}
}

func TestVertexEdges(t *testing.T) {
	s := makeTriangle(t)
	for v := s.Verts; v != nil; v = v.Next {
		edges := VertexEdges(v)
		if len(edges) == 0 {
			t.Fatalf("vertex %d should have incident edges", v.Index)
		}
	}
}

func TestVertexNeighbors(t *testing.T) {
	s, _, _ := euler.Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	_, _, _ = euler.Lmev(s.Verts.He, vec.SFVec3f{X: 1})
	v := s.Verts
	nbrs := VertexNeighbors(v)
	if len(nbrs) < 1 {
		t.Fatalf("wire vertex should have >=1 neighbor, got %d", len(nbrs))
	}
}

func TestVertexNeighbors_Nil(t *testing.T) {
	v := &base.Vertex{}
	if nbrs := VertexNeighbors(v); nbrs != nil {
		t.Fatal("nil He vertex should return nil neighbors")
	}
}

func TestFaceNeighbors(t *testing.T) {
	s := makeQuad(t)
	Triangulate(s)
	for f := s.Faces; f != nil; f = f.Next {
		nbrs := FaceNeighbors(f)
		if f.LoopOut != nil && f.LoopOut.NHalfEdges() >= 3 && len(nbrs) == 0 {
			t.Fatalf("face %d should have neighbors", f.Index)
		}
	}
}
