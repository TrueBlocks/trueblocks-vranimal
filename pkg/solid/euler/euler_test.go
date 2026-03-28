package euler

import (
	"errors"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func TestMvfs(t *testing.T) {
	s, v, f := Mvfs(vec.SFVec3f{X: 1, Y: 2, Z: 3}, vec.Red)
	if s == nil || v == nil || f == nil {
		t.Fatal("Mvfs returned nil")
	}
	if s.NFaces() != 1 {
		t.Errorf("expected 1 face, got %d", s.NFaces())
	}
	if s.NVerts() != 1 {
		t.Errorf("expected 1 vertex, got %d", s.NVerts())
	}
	if s.NEdges() != 0 {
		t.Errorf("expected 0 edges, got %d", s.NEdges())
	}
	if v.Loc.X != 1 || v.Loc.Y != 2 || v.Loc.Z != 3 {
		t.Errorf("vertex location wrong: %v", v.Loc)
	}
}

func TestKvfs(t *testing.T) {
	s, _, _ := Mvfs(vec.SFVec3f{}, vec.White)
	Kvfs(s)
	if s.NFaces() != 0 || s.NEdges() != 0 || s.NVerts() != 0 {
		t.Error("Kvfs did not clear solid")
	}
}

func TestLmev(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, e, err := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	if err != nil {
		t.Fatalf("Lmev returned error: %v", err)
	}
	if v1 == nil || e == nil {
		t.Fatal("Lmev returned nil")
	}
	if s.NVerts() != 2 {
		t.Errorf("expected 2 verts, got %d", s.NVerts())
	}
	if s.NEdges() != 1 {
		t.Errorf("expected 1 edge, got %d", s.NEdges())
	}
}

func TestLmef(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, err := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	if err != nil {
		t.Fatalf("Lmev returned error: %v", err)
	}
	v2, _, err := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})
	if err != nil {
		t.Fatalf("Lmev returned error: %v", err)
	}
	nf, ne, err := Lmef(v0.He, v2.He)
	if err != nil {
		t.Fatalf("Lmef returned error: %v", err)
	}
	if nf == nil || ne == nil {
		t.Fatal("Lmef returned nil")
	}
}

func TestLkev(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	_, _, err := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	if err != nil {
		t.Fatalf("Lmev returned error: %v", err)
	}
	if s.NVerts() != 2 || s.NEdges() != 1 {
		t.Fatalf("wrong counts before Lkev: v=%d e=%d", s.NVerts(), s.NEdges())
	}
	he := v0.He
	if err := Lkev(he); err != nil {
		t.Fatalf("Lkev returned error: %v", err)
	}
	if s.NVerts() != 1 {
		t.Errorf("expected 1 vert after Lkev, got %d", s.NVerts())
	}
	if s.NEdges() != 0 {
		t.Errorf("expected 0 edges after Lkev, got %d", s.NEdges())
	}
}

func makeQuad() *base.Solid {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})
	v3, _, _ := Lmev(v2.He, vec.SFVec3f{X: 0, Y: 1, Z: 0})
	_, _, _ = Lmef(v0.He, v3.He)
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func TestQuadTopology(t *testing.T) {
	s := makeQuad()
	if s.NFaces() != 2 {
		t.Errorf("expected 2 faces, got %d", s.NFaces())
	}
	if s.NEdges() != 4 {
		t.Errorf("expected 4 edges, got %d", s.NEdges())
	}
	if s.NVerts() != 4 {
		t.Errorf("expected 4 verts, got %d", s.NVerts())
	}
}

func TestLmef_RejectsDifferentLoops(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})
	v3, _, _ := Lmev(v2.He, vec.SFVec3f{X: 0, Y: 1, Z: 0})
	_, _, _ = Lmef(v0.He, v3.He)

	v4, _, _ := Lmev(v0.He, vec.SFVec3f{X: 0, Y: 0, Z: 1})
	v5, _, _ := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 0, Z: 1})

	_, _, err := Lmef(v4.He, v5.He)
	if !errors.Is(err, ErrDifferentLoops) {
		t.Errorf("expected ErrDifferentLoops, got %v", err)
	}
}

func TestBuildFromIndexSet(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int64{0, 1, 2, 3, -1}
	s, err := BuildFromIndexSet(positions, indices, vec.White)
	if err != nil {
		t.Fatalf("BuildFromIndexSet returned error: %v", err)
	}
	if s == nil {
		t.Fatal("BuildFromIndexSet returned nil")
	}
	if s.NVerts() != 4 {
		t.Errorf("expected 4 verts, got %d", s.NVerts())
	}
}

func TestMarkCreases(t *testing.T) {
	s := makeQuad()
	s.CalcPlaneEquations()
	MarkCreases(s, 0.5)
}

func TestLmfkrh(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})
	v3, _, _ := Lmev(v2.He, vec.SFVec3f{X: 0, Y: 1, Z: 0})
	_, _, _ = Lmef(v0.He, v3.He)
	if s.NFaces() < 2 {
		t.Error("expected at least 2 faces")
	}
}

func TestLmev_NilHalfEdge(t *testing.T) {
	_, _, err := Lmev(nil, vec.SFVec3f{})
	if !errors.Is(err, ErrNilHalfEdge) {
		t.Errorf("expected ErrNilHalfEdge, got %v", err)
	}
}

func TestLmef_NilHalfEdge(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{}, vec.White)
	_, _, err := Lmef(nil, v0.He)
	if !errors.Is(err, ErrNilHalfEdge) {
		t.Errorf("expected ErrNilHalfEdge, got %v", err)
	}
}

func TestBuildFromIndexSet_Empty(t *testing.T) {
	_, err := BuildFromIndexSet(nil, nil, vec.White)
	if err == nil {
		t.Error("expected error for empty inputs")
	}
}
