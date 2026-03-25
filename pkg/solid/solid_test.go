package solid

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Mvfs / Kvfs - basic solid creation
// ---------------------------------------------------------------------------

func TestMvfs(t *testing.T) {
	s, v, f := Mvfs(vec.SFVec3f{X: 1, Y: 2, Z: 3}, vec.Red)
	if s == nil || v == nil || f == nil {
		t.Fatal("nil result")
	}
	if s.NVerts() != 1 {
		t.Fatalf("expected 1 vert, got %d", s.NVerts())
	}
	if s.NFaces() != 1 {
		t.Fatalf("expected 1 face, got %d", s.NFaces())
	}
	if v.Loc != (vec.SFVec3f{X: 1, Y: 2, Z: 3}) {
		t.Fatalf("vertex loc = %v", v.Loc)
	}
}

func TestKvfs(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.White)
	Kvfs(s)
	if s.NFaces() != 0 || s.NEdges() != 0 || s.NVerts() != 0 {
		t.Fatal("should be empty after Kvfs")
	}
}

// ---------------------------------------------------------------------------
// Lmev - split half-edge, create vertex+edge
// ---------------------------------------------------------------------------

func TestLmev(t *testing.T) {
	s, v, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	nv, ne := Lmev(v.He, vec.SFVec3f{X: 1})
	if nv == nil || ne == nil {
		t.Fatal("nil result")
	}
	if s.NVerts() != 2 {
		t.Fatalf("expected 2 verts, got %d", s.NVerts())
	}
	if s.NEdges() != 1 {
		t.Fatalf("expected 1 edge, got %d", s.NEdges())
	}
	if nv.Loc.X != 1 {
		t.Fatalf("new vert X = %g", nv.Loc.X)
	}
}

// ---------------------------------------------------------------------------
// Solid list operations
// ---------------------------------------------------------------------------

func TestSolid_AddRemoveFace(t *testing.T) {
	s := NewSolid()
	f1 := &Face{Data: &ColorData{}}
	f2 := &Face{Data: &ColorData{}}
	s.AddFace(f1)
	s.AddFace(f2)
	if s.NFaces() != 2 {
		t.Fatalf("expected 2, got %d", s.NFaces())
	}
	s.RemoveFace(f1)
	if s.NFaces() != 1 {
		t.Fatalf("expected 1 after remove, got %d", s.NFaces())
	}
	s.RemoveFace(f2)
	if s.NFaces() != 0 {
		t.Fatalf("expected 0, got %d", s.NFaces())
	}
}

func TestSolid_AddRemoveEdge(t *testing.T) {
	s := NewSolid()
	e1 := NewEdge()
	e2 := NewEdge()
	s.AddEdge(e1)
	s.AddEdge(e2)
	if s.NEdges() != 2 {
		t.Fatalf("expected 2, got %d", s.NEdges())
	}
	s.RemoveEdge(e2)
	if s.NEdges() != 1 {
		t.Fatalf("expected 1, got %d", s.NEdges())
	}
}

func TestSolid_AddRemoveVertex(t *testing.T) {
	s := NewSolid()
	v1 := NewVertex(0, 0, 0)
	v2 := NewVertex(1, 1, 1)
	s.AddVertex(v1)
	s.AddVertex(v2)
	if s.NVerts() != 2 {
		t.Fatalf("expected 2, got %d", s.NVerts())
	}
	s.RemoveVertex(v1)
	if s.NVerts() != 1 {
		t.Fatalf("expected 1, got %d", s.NVerts())
	}
}

// ---------------------------------------------------------------------------
// Iterators
// ---------------------------------------------------------------------------

func TestSolid_ForEachVertex(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	Lmev(v0.He, vec.SFVec3f{X: 2})
	count := 0
	s.ForEachVertex(func(v *Vertex) bool {
		count++
		return true
	})
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestSolid_ForEachVertex_EarlyStop(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	Lmev(v0.He, vec.SFVec3f{X: 2})
	count := 0
	s.ForEachVertex(func(v *Vertex) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Fatalf("expected 2 (early stop), got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Find operations
// ---------------------------------------------------------------------------

func TestSolid_FindVertex(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	s.Renumber()
	v := s.FindVertex(0)
	if v == nil {
		t.Fatal("should find vertex 0")
	}
	if s.FindVertex(999) != nil {
		t.Fatal("should not find vertex 999")
	}
}

func TestSolid_FindFace(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	s.Renumber()
	f := s.FindFace(0)
	if f == nil {
		t.Fatal("should find face 0")
	}
	if s.FindFace(999) != nil {
		t.Fatal("should not find face 999")
	}
}

// ---------------------------------------------------------------------------
// Extents
// ---------------------------------------------------------------------------

func TestSolid_Extents(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: -1, Y: -2, Z: -3}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 4, Y: 5, Z: 6})
	min, max := s.Extents()
	if min.X != -1 || min.Y != -2 || min.Z != -3 {
		t.Fatalf("min = %v", min)
	}
	if max.X != 4 || max.Y != 5 || max.Z != 6 {
		t.Fatalf("max = %v", max)
	}
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------

func TestSolid_Stats(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	faces, edges, verts := s.Stats()
	if faces != 1 || edges != 1 || verts != 2 {
		t.Fatalf("stats: f=%d e=%d v=%d", faces, edges, verts)
	}
}

// ---------------------------------------------------------------------------
// Renumber
// ---------------------------------------------------------------------------

func TestSolid_Renumber(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	s.Renumber()
	// After renumber, indices should be sequential
	var vertIndices []uint32
	s.ForEachVertex(func(v *Vertex) bool {
		vertIndices = append(vertIndices, v.Index)
		return true
	})
	if len(vertIndices) != 2 {
		t.Fatalf("expected 2 verts, got %d", len(vertIndices))
	}
}

// ---------------------------------------------------------------------------
// SetColor / SetMarks
// ---------------------------------------------------------------------------

func TestSolid_SetColor(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	s.SetColor(vec.Blue)
	s.ForEachFace(func(f *Face) bool {
		c := f.GetColor(vec.White)
		if c != vec.Blue {
			t.Fatalf("expected blue, got %v", c)
		}
		return true
	})
}

func TestSolid_SetFaceMarks(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	s.SetFaceMarks(42)
	s.ForEachFace(func(f *Face) bool {
		if f.Mark1 != 42 {
			t.Fatalf("expected 42, got %d", f.Mark1)
		}
		return true
	})
}

func TestSolid_SetVertexMarks(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	s.SetVertexMarks(7)
	s.ForEachVertex(func(v *Vertex) bool {
		if v.Mark != 7 {
			t.Fatalf("expected 7, got %d", v.Mark)
		}
		return true
	})
}

// ---------------------------------------------------------------------------
// TransformGeometry
// ---------------------------------------------------------------------------

func TestSolid_TransformGeometry(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 1, Y: 0, Z: 0}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 2, Y: 0, Z: 0})
	m := vec.ScaleMatrix(2, 2, 2)
	s.TransformGeometry(m)
	// All vertices should be doubled
	s.ForEachVertex(func(v *Vertex) bool {
		if v.Loc.X != 2 && v.Loc.X != 4 {
			t.Fatalf("unexpected X = %g", v.Loc.X)
		}
		return true
	})
}

// ---------------------------------------------------------------------------
// Merge
// ---------------------------------------------------------------------------

func TestSolid_Merge(t *testing.T) {
	s1, _, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	s2, _, _ := Mvfs(vec.SFVec3f{X: 5}, vec.Blue)
	s1.Merge(s2)
	if s1.NVerts() != 2 {
		t.Fatalf("expected 2 verts after merge, got %d", s1.NVerts())
	}
	if s1.NFaces() != 2 {
		t.Fatalf("expected 2 faces after merge, got %d", s1.NFaces())
	}
}

func TestSolid_MergeNil(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	s.Merge(nil) // should not panic
	if s.NVerts() != 1 {
		t.Fatal("merge nil should be no-op")
	}
}

// ---------------------------------------------------------------------------
// ColorData
// ---------------------------------------------------------------------------

func TestColorData_Flags(t *testing.T) {
	d := &ColorData{}
	if d.HasColor() || d.HasNormal() || d.HasTexCoord() {
		t.Fatal("fresh ColorData should have no flags")
	}
	d.SetColor(vec.Red)
	if !d.HasColor() {
		t.Fatal("should have color")
	}
	if d.GetColor() != vec.Red {
		t.Fatalf("color = %v", d.GetColor())
	}
	d.SetNormal(vec.YAxis)
	if !d.HasNormal() {
		t.Fatal("should have normal")
	}
	d.SetTexCoord(vec.SFVec2f{X: 0.5, Y: 0.5})
	if !d.HasTexCoord() {
		t.Fatal("should have texcoord")
	}
}

// ---------------------------------------------------------------------------
// HalfEdge helpers
// ---------------------------------------------------------------------------

func TestHalfEdge_SetGetColor(t *testing.T) {
	l := &Loop{}
	v := NewVertex(0, 0, 0)
	he := NewHalfEdge(l, v)
	// Default should return fallback
	if he.GetColor(vec.White) != vec.White {
		t.Fatal("default should return fallback")
	}
	he.SetColor(vec.Green)
	if he.GetColor(vec.White) != vec.Green {
		t.Fatal("should return set color")
	}
}

func TestHalfEdge_Bisect(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	nv, _ := Lmev(v0.He, vec.SFVec3f{X: 10})
	_ = s
	// Find the half-edge connecting them
	he := v0.He
	for he.GetMate() == nil {
		he = he.Next
		if he == v0.He {
			break
		}
	}
	_ = nv
	// Bisect should give midpoint
	if he.GetMate() != nil {
		mid := he.Bisect()
		if mid.X < 4.9 || mid.X > 5.1 {
			t.Fatalf("expected ~5, got %g", mid.X)
		}
	}
}

// ---------------------------------------------------------------------------
// Face helpers
// ---------------------------------------------------------------------------

func TestFace_GetSetColor(t *testing.T) {
	s, _, f := Mvfs(vec.SFVec3f{}, vec.Red)
	_ = s
	c := f.GetColor(vec.White)
	if c != vec.Red {
		t.Fatalf("expected red, got %v", c)
	}
	f.SetColor(vec.Blue)
	if f.GetColor(vec.White) != vec.Blue {
		t.Fatal("should be blue")
	}
}

func TestFace_NLoops(t *testing.T) {
	_, _, f := Mvfs(vec.SFVec3f{}, vec.Red)
	if f.NLoops() != 1 {
		t.Fatalf("expected 1 loop, got %d", f.NLoops())
	}
}

func TestFace_Marked(t *testing.T) {
	_, _, f := Mvfs(vec.SFVec3f{}, vec.Red)
	f.Mark1 = 5
	f.Mark2 = 10
	if !f.Marked1(5) {
		t.Fatal("mark1 should be 5")
	}
	if !f.Marked2(10) {
		t.Fatal("mark2 should be 10")
	}
}

// ---------------------------------------------------------------------------
// Edge helpers
// ---------------------------------------------------------------------------

func TestEdge_SwapHes(t *testing.T) {
	he1 := &HalfEdge{Vertex: NewVertex(0, 0, 0)}
	he2 := &HalfEdge{Vertex: NewVertex(1, 1, 1)}
	e := &Edge{He1: he1, He2: he2}
	e.SwapHes()
	if e.He1 != he2 || e.He2 != he1 {
		t.Fatal("swap failed")
	}
}

func TestEdge_Marked(t *testing.T) {
	e := &Edge{Mark: CREASE}
	if !e.Marked(CREASE) {
		t.Fatal("should be marked crease")
	}
	if e.Marked(0) {
		t.Fatal("should not be marked 0")
	}
}

// ---------------------------------------------------------------------------
// Loop operations
// ---------------------------------------------------------------------------

func TestLoop_AddHalfEdge(t *testing.T) {
	f := &Face{}
	l := &Loop{Face: f}
	v1 := NewVertex(0, 0, 0)
	v2 := NewVertex(1, 0, 0)
	v3 := NewVertex(0, 1, 0)
	he1 := NewHalfEdge(l, v1)
	he2 := NewHalfEdge(l, v2)
	he3 := NewHalfEdge(l, v3)
	l.AddHalfEdge(he1)
	l.AddHalfEdge(he2)
	l.AddHalfEdge(he3)
	// Should form circular list: he1->he2->he3->he1
	if he1.Next != he2 || he2.Next != he3 || he3.Next != he1 {
		t.Fatal("circular next chain broken")
	}
	if he1.Prev != he3 || he2.Prev != he1 || he3.Prev != he2 {
		t.Fatal("circular prev chain broken")
	}
}

func TestLoop_ForEachHe(t *testing.T) {
	f := &Face{}
	l := &Loop{Face: f}
	v := NewVertex(0, 0, 0)
	for i := 0; i < 4; i++ {
		l.AddHalfEdge(NewHalfEdge(l, v))
	}
	count := 0
	l.ForEachHe(func(he *HalfEdge) bool {
		count++
		return true
	})
	if count != 4 {
		t.Fatalf("expected 4, got %d", count)
	}
}

func TestLoop_ForEachHe_EarlyStop(t *testing.T) {
	f := &Face{}
	l := &Loop{Face: f}
	v := NewVertex(0, 0, 0)
	for i := 0; i < 4; i++ {
		l.AddHalfEdge(NewHalfEdge(l, v))
	}
	count := 0
	l.ForEachHe(func(he *HalfEdge) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestLoop_ForEachHe_Empty(t *testing.T) {
	l := &Loop{}
	count := 0
	l.ForEachHe(func(he *HalfEdge) bool {
		count++
		return true
	})
	if count != 0 {
		t.Fatal("empty loop should iterate 0")
	}
}

// ---------------------------------------------------------------------------
// Vertex constructors
// ---------------------------------------------------------------------------

func TestNewVertex(t *testing.T) {
	v := NewVertex(1, 2, 3)
	if v.Loc != (vec.SFVec3f{X: 1, Y: 2, Z: 3}) {
		t.Fatalf("got %v", v.Loc)
	}
}

func TestNewVertexVec(t *testing.T) {
	v := NewVertexVec(vec.SFVec3f{X: 4, Y: 5, Z: 6})
	if v.Loc.X != 4 || v.Loc.Y != 5 || v.Loc.Z != 6 {
		t.Fatalf("got %v", v.Loc)
	}
}

func TestVertex_ColorData(t *testing.T) {
	v := NewVertex(0, 0, 0)
	if v.GetColor(vec.White) != vec.White {
		t.Fatal("default should return fallback")
	}
	v.SetColor(vec.Red)
	if v.GetColor(vec.White) != vec.Red {
		t.Fatal("should return set color")
	}
	v.SetNormal(vec.YAxis)
	if v.GetNormal(vec.SFVec3f{}) != vec.YAxis {
		t.Fatal("normal mismatch")
	}
	v.SetTexCoord(vec.SFVec2f{X: 1, Y: 1})
	if v.GetTexCoord(vec.SFVec2f{}) != (vec.SFVec2f{X: 1, Y: 1}) {
		t.Fatal("texcoord mismatch")
	}
}

// ---------------------------------------------------------------------------
// BuildFromIndexSet
// ---------------------------------------------------------------------------

func TestBuildFromIndexSet_Nil(t *testing.T) {
	if BuildFromIndexSet(nil, nil, vec.Red) != nil {
		t.Fatal("nil input should return nil")
	}
	if BuildFromIndexSet([]vec.SFVec3f{}, []int32{}, vec.Red) != nil {
		t.Fatal("empty input should return nil")
	}
}

func TestBuildFromIndexSet_Triangle(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int32{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, vec.Red)
	if s == nil {
		t.Fatal("nil solid")
	}
	if s.NVerts() < 3 {
		t.Fatalf("expected >= 3 verts, got %d", s.NVerts())
	}
}
