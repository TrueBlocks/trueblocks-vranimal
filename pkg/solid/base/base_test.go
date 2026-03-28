package base

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func TestFloatCompare(t *testing.T) {
	if FloatCompare(0.1) != 1 {
		t.Error("expected 1 for positive")
	}
	if FloatCompare(-0.1) != -1 {
		t.Error("expected -1 for negative")
	}
	if FloatCompare(0.000001) != 0 {
		t.Error("expected 0 for near-zero")
	}
}

func TestGetDominantComp(t *testing.T) {
	if GetDominantComp(5, 1, 1) != 0 {
		t.Error("expected X dominant")
	}
	if GetDominantComp(1, 5, 1) != 1 {
		t.Error("expected Y dominant")
	}
	if GetDominantComp(1, 1, 5) != 2 {
		t.Error("expected Z dominant")
	}
}

func TestCollinear(t *testing.T) {
	if !Collinear(0, 0, 0, 1, 0, 0, 2, 0, 0) {
		t.Error("expected collinear")
	}
	if Collinear(0, 0, 0, 1, 0, 0, 0, 1, 0) {
		t.Error("expected not collinear")
	}
}

func TestVecEq(t *testing.T) {
	if !VecEq(1, 2, 3, 1, 2, 3) {
		t.Error("expected equal")
	}
	if VecEq(1, 2, 3, 1, 2, 4) {
		t.Error("expected not equal")
	}
}

func TestCrossingsTest(t *testing.T) {
	square := [][2]float64{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	if !CrossingsTest(square, [2]float64{0.5, 0.5}) {
		t.Error("expected inside")
	}
	if CrossingsTest(square, [2]float64{2, 2}) {
		t.Error("expected outside")
	}
}

func TestNewSolid(t *testing.T) {
	s := NewSolid()
	if s.NFaces() != 0 || s.NEdges() != 0 || s.NVerts() != 0 {
		t.Error("new solid should be empty")
	}
}

func TestSolid_AddRemoveVertex(t *testing.T) {
	s := NewSolid()
	v := NewVertex(1, 2, 3)
	s.AddVertex(v)
	if s.NVerts() != 1 {
		t.Errorf("expected 1 vert, got %d", s.NVerts())
	}
	s.RemoveVertex(v)
	if s.NVerts() != 0 {
		t.Error("expected 0 verts after remove")
	}
}

func TestSolid_AddRemoveFace(t *testing.T) {
	s := NewSolid()
	f := NewFace(s, vec.White)
	s.AddFace(f)
	if s.NFaces() != 1 {
		t.Errorf("expected 1 face, got %d", s.NFaces())
	}
	s.RemoveFace(f)
	if s.NFaces() != 0 {
		t.Error("expected 0 faces after remove")
	}
}

func TestHalfEdge_GetMate(t *testing.T) {
	he1 := &HalfEdge{}
	he2 := &HalfEdge{}
	e := &Edge{He1: he1, He2: he2}
	he1.Edge = e
	he2.Edge = e
	if he1.GetMate() != he2 {
		t.Error("he1 mate should be he2")
	}
	if he2.GetMate() != he1 {
		t.Error("he2 mate should be he1")
	}
}

func TestLoop_AddHalfEdge(t *testing.T) {
	f := &Face{}
	l := &Loop{Face: f}
	f.Loops = append(f.Loops, l)
	f.LoopOut = l
	v1 := NewVertex(0, 0, 0)
	v2 := NewVertex(1, 0, 0)
	v3 := NewVertex(0, 1, 0)
	he1 := NewHalfEdge(l, v1)
	he2 := NewHalfEdge(l, v2)
	he3 := NewHalfEdge(l, v3)
	l.AddHalfEdge(he1)
	l.AddHalfEdge(he2)
	l.AddHalfEdge(he3)
	if l.NHalfEdges() != 3 {
		t.Errorf("expected 3 half-edges, got %d", l.NHalfEdges())
	}
	if he1.Next != he2 || he2.Next != he3 || he3.Next != he1 {
		t.Error("circular list broken")
	}
}

func TestFace_CalcEquation(t *testing.T) {
	s := NewSolid()
	f := NewFace(s, vec.White)
	s.AddFace(f)
	l := NewLoop(f, true)
	he1 := NewHalfEdge(l, NewVertex(0, 0, 0))
	he2 := NewHalfEdge(l, NewVertex(1, 0, 0))
	he3 := NewHalfEdge(l, NewVertex(0, 1, 0))
	l.AddHalfEdge(he1)
	l.AddHalfEdge(he2)
	l.AddHalfEdge(he3)
	f.CalcEquation()
	if f.Normal.Z < 0.99 {
		t.Errorf("expected Z-up normal, got %v", f.Normal)
	}
}

func TestPlane_GetDistance(t *testing.T) {
	pl := Plane{Normal: vec.SFVec3f{X: 0, Y: 0, Z: 1}, D: 0}
	d := pl.GetDistance(vec.SFVec3f{X: 0, Y: 0, Z: 5})
	if d != 5 {
		t.Errorf("expected 5, got %f", d)
	}
}

func TestSolid_Renumber(t *testing.T) {
	s := NewSolid()
	v1 := NewVertex(0, 0, 0)
	v2 := NewVertex(1, 0, 0)
	s.AddVertex(v1)
	s.AddVertex(v2)
	s.Renumber()
	if v2.Index != 0 || v1.Index != 1 {
		t.Errorf("unexpected indices: v2=%d v1=%d", v2.Index, v1.Index)
	}
}
