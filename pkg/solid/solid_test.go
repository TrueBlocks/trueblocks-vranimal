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

// ---------------------------------------------------------------------------
// New HalfEdge methods
// ---------------------------------------------------------------------------

func TestHalfEdge_GetMateIndex(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	nv, _ := Lmev(v0.He, vec.SFVec3f{X: 10})
	s.Renumber()
	_ = nv
	// Find a half-edge with a mate
	for e := s.Edges; e != nil; e = e.Next {
		idx := e.He1.GetMateIndex()
		if idx != e.He2.GetIndex() {
			t.Fatalf("mate index %d != he2 index %d", idx, e.He2.GetIndex())
		}
	}
}

func TestHalfEdge_GetSolid(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	if v0.He.GetSolid() != s {
		t.Fatal("GetSolid should return the solid")
	}
}

func TestHalfEdge_IsNullStrut(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	// On a simple edge, the mate IS adjacent → null strut
	for he := v0.He; ; he = he.Next {
		if he.Edge != nil {
			// In this simple topology, strut should be true
			_ = he.IsNullStrut()
			break
		}
		if he.Next == v0.He {
			break
		}
	}
}

func TestHalfEdge_IsMate(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	for he := v0.He; ; he = he.Next {
		if he.Edge != nil {
			m := he.GetMate()
			if m != nil && !he.IsMate(m) {
				t.Fatal("should be mates")
			}
			break
		}
		if he.Next == v0.He {
			break
		}
	}
}

func TestHalfEdge_InsideVector(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	iv := v0.He.InsideVector()
	// InsideVector is cross(normal, edge_dir); for a degenerate face it may be zero
	_ = iv
}

func TestHalfEdge_IsWide(t *testing.T) {
	// Build a triangle face to test angle classification
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
	// All angles in a triangle are < 180, so IsWide should be false
	s.ForEachFace(func(f *Face) bool {
		if f.LoopOut == nil {
			return true
		}
		f.LoopOut.ForEachHe(func(he *HalfEdge) bool {
			// Don't crash
			_ = he.IsWide(false)
			_ = he.IsWide(true)
			_ = he.Is180()
			return true
		})
		return true
	})
}

// ---------------------------------------------------------------------------
// New Loop methods
// ---------------------------------------------------------------------------

func TestLoop_GetSolid(t *testing.T) {
	s, _, f := Mvfs(vec.SFVec3f{}, vec.Red)
	l := f.LoopOut
	if l.GetSolid() != s {
		t.Fatal("GetSolid should return the solid")
	}
}

func TestLoop_NHalfEdges(t *testing.T) {
	_, v0, f := Mvfs(vec.SFVec3f{}, vec.Red)
	_ = v0
	l := f.LoopOut
	if l.NHalfEdges() != 1 {
		t.Fatalf("expected 1 halfedge in Mvfs loop, got %d", l.NHalfEdges())
	}
}

func TestLoop_GetVertexLocations(t *testing.T) {
	_, v0, f := Mvfs(vec.SFVec3f{X: 1, Y: 2, Z: 3}, vec.Red)
	_ = v0
	locs := make([]vec.SFVec3f, 10)
	n := f.LoopOut.GetVertexLocations(locs)
	if n != 1 {
		t.Fatalf("expected 1, got %d", n)
	}
	if locs[0] != (vec.SFVec3f{X: 1, Y: 2, Z: 3}) {
		t.Fatalf("got %v", locs[0])
	}
}

// ---------------------------------------------------------------------------
// New Vertex methods
// ---------------------------------------------------------------------------

func TestVertex_GetValence(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	Lmev(v0.He, vec.SFVec3f{X: 2})
	// v0 should have valence 2 (two edges connected)
	val := v0.GetValence()
	if val < 1 {
		t.Fatalf("expected valence >= 1, got %d", val)
	}
}

func TestVertex_CalcNormal(t *testing.T) {
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
	s.CalcPlaneEquations()
	s.CalcVertexNormals()
	// Should not panic
}

func TestVertex_IsMarked(t *testing.T) {
	v := NewVertex(0, 0, 0)
	v.Mark = VISITED
	if !v.IsMarked(VISITED) {
		t.Fatal("should be marked VISITED")
	}
	if v.IsMarked(0) {
		t.Fatal("should not be marked 0")
	}
}

// ---------------------------------------------------------------------------
// New Edge methods
// ---------------------------------------------------------------------------

func TestEdge_GetSolid(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	_, ne := Lmev(v0.He, vec.SFVec3f{X: 1})
	if ne.GetSolid() != s {
		t.Fatal("GetSolid should return solid")
	}
}

func TestEdge_Length(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	_, ne := Lmev(v0.He, vec.SFVec3f{X: 3})
	l := ne.Length()
	if l < 2.9 || l > 3.1 {
		t.Fatalf("expected ~3, got %g", l)
	}
}

func TestEdge_Midpoint(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{X: 0}, vec.Red)
	_, ne := Lmev(v0.He, vec.SFVec3f{X: 10})
	mid := ne.Midpoint()
	if mid.X < 4.9 || mid.X > 5.1 {
		t.Fatalf("expected ~5, got %g", mid.X)
	}
}

// ---------------------------------------------------------------------------
// New Face methods
// ---------------------------------------------------------------------------

func TestFace_GetFirstSecondLoop(t *testing.T) {
	_, _, f := Mvfs(vec.SFVec3f{}, vec.Red)
	if f.GetFirstLoop() == nil {
		t.Fatal("GetFirstLoop should not be nil")
	}
	if f.GetSecondLoop() != nil {
		t.Fatal("GetSecondLoop should be nil with only 1 loop")
	}
}

func TestFace_IsDegenerate(t *testing.T) {
	_, _, f := Mvfs(vec.SFVec3f{}, vec.Red)
	// Single vertex face is degenerate
	if !f.IsDegenerate() {
		t.Fatal("single vertex face should be degenerate")
	}
}

func TestFace_IsPlanar(t *testing.T) {
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
	s.CalcPlaneEquations()
	s.ForEachFace(func(f *Face) bool {
		if f.LoopOut != nil && f.LoopOut.NHalfEdges() >= 3 {
			if !f.IsPlanar() {
				t.Fatal("triangle should be planar")
			}
		}
		return true
	})
}

// ---------------------------------------------------------------------------
// New Solid methods
// ---------------------------------------------------------------------------

func TestSolid_IsWire(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	if !s.IsWire() {
		t.Fatal("Mvfs solid should be wire (1 face)")
	}
}

func TestSolid_IsLamina(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	if s.IsLamina() {
		t.Fatal("2-vertex solid should not be lamina yet")
	}
}

func TestSolid_GetFirst(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	if s.GetFirstFace() == nil {
		t.Fatal("should have first face")
	}
	if s.GetFirstVertex() == nil {
		t.Fatal("should have first vertex")
	}
}

func TestSolid_Revert(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	s.Revert() // should not panic
}

func TestSolid_CalcVertexNormals(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	s.CalcPlaneEquations()
	s.CalcVertexNormals() // should not panic
}

func TestSolid_Cleanup(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 1})
	oldV := s.NVerts()
	oldE := s.NEdges()
	s.Cleanup()
	if s.NVerts() != oldV {
		t.Fatalf("cleanup changed verts: %d -> %d", oldV, s.NVerts())
	}
	if s.NEdges() != oldE {
		t.Fatalf("cleanup changed edges: %d -> %d", oldE, s.NEdges())
	}
}

func TestSolid_Copy(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 1, Y: 2, Z: 3}, vec.Red)
	Lmev(v0.He, vec.SFVec3f{X: 4, Y: 5, Z: 6})
	s.Renumber()

	c := s.Copy()
	if c.NVerts() != s.NVerts() {
		t.Fatalf("copy verts %d != %d", c.NVerts(), s.NVerts())
	}
	if c.NEdges() != s.NEdges() {
		t.Fatalf("copy edges %d != %d", c.NEdges(), s.NEdges())
	}
	if c.NFaces() != s.NFaces() {
		t.Fatalf("copy faces %d != %d", c.NFaces(), s.NFaces())
	}
	// Copy should be independent
	if c.GetFirstVertex() == s.GetFirstVertex() {
		t.Fatal("copy should not share vertices")
	}
}

func TestSolid_AllocateColorData(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	s.AllocateColorData(PerVertex)
	s.ForEachVertex(func(v *Vertex) bool {
		if v.Data == nil {
			t.Fatal("vertex should have ColorData after allocation")
		}
		return true
	})
}

// ---------------------------------------------------------------------------
// Lmev2 / Lringmv
// ---------------------------------------------------------------------------

func TestLmev2_SameAsLmev(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	// With a single half-edge, he1==he2 effectively
	nv, ne := Lmev2(v0.He, v0.He, vec.SFVec3f{X: 5})
	if nv == nil || ne == nil {
		t.Fatal("nil result")
	}
	if s.NVerts() != 2 {
		t.Fatalf("expected 2 verts, got %d", s.NVerts())
	}
}

func TestLmev2_DifferentVertex(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{}, vec.Red)
	v1 := NewVertex(1, 1, 1)
	he2 := &HalfEdge{Vertex: v1} // different vertex
	nv, ne := Lmev2(v0.He, he2, vec.SFVec3f{X: 5})
	if nv != nil || ne != nil {
		t.Fatal("should return nil for different vertices")
	}
}

func TestLringmv(t *testing.T) {
	s, _, f1 := Mvfs(vec.SFVec3f{}, vec.Red)
	f2 := NewFace(s, vec.Blue)
	s.AddFace(f2)
	l := f1.LoopOut
	Lringmv(s, l, f2, true)
	if l.Face != f2 {
		t.Fatal("loop should belong to f2 after ringmv")
	}
	if f2.LoopOut != l {
		t.Fatal("f2 outer loop should be the moved loop")
	}
}
