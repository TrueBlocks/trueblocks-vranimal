package euler

import (
	"errors"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// validateTopology checks structural invariants assuming genus-0 (V-E+2F-L=2).
func validateTopology(t *testing.T, s *base.Solid) {
	t.Helper()
	validateTopologyEP(t, s, 2)
}

// validateTopologyEP checks structural invariants with an explicit expected
// Euler-Poincaré value (2 for genus-0, 0 for genus-1, etc.).
func validateTopologyEP(t *testing.T, s *base.Solid, wantEP int) {
	t.Helper()

	// Collect the set of faces actually in the solid (for loop-ownership check).
	faceSet := make(map[*base.Face]bool)
	faceCount := 0
	for f := s.Faces; f != nil; f = f.Next {
		faceSet[f] = true
		faceCount++
	}

	// Check 4a: face count consistency.
	if faceCount != s.NFaces() {
		t.Errorf("face list walk counted %d, NFaces()=%d", faceCount, s.NFaces())
	}

	// Check 4b: edge count consistency.
	edgeCount := 0
	for e := s.Edges; e != nil; e = e.Next {
		edgeCount++
	}
	if edgeCount != s.NEdges() {
		t.Errorf("edge list walk counted %d, NEdges()=%d", edgeCount, s.NEdges())
	}

	// Check 4c: vertex count consistency.
	vertCount := 0
	for v := s.Verts; v != nil; v = v.Next {
		vertCount++
	}
	if vertCount != s.NVerts() {
		t.Errorf("vertex list walk counted %d, NVerts()=%d", vertCount, s.NVerts())
	}

	// Track which vertices are reachable from half-edge rings (check 3).
	reachable := make(map[*base.Vertex]bool)

	L := 0
	for f := s.Faces; f != nil; f = f.Next {
		for _, l := range f.Loops {
			L++

			// Check 5: loop ownership — loop's Face must be in the solid.
			if !faceSet[l.Face] {
				t.Errorf("loop in face %d has .Face not in solid's face list", f.Index)
			}
			if l.Face != f {
				t.Errorf("loop in face %d has .Face pointing to face %d", f.Index, l.Face.Index)
			}

			if l.HalfEdges == nil {
				continue
			}
			he := l.HalfEdges
			start := he
			n := 0
			for {
				if he.Next == nil {
					t.Fatalf("he.Next is nil in loop of face %d", f.Index)
					return
				}
				if he.Prev == nil {
					t.Fatalf("he.Prev is nil in loop of face %d", f.Index)
					return
				}
				if he.Next.Prev != he {
					t.Errorf("he.Next.Prev != he in face %d", f.Index)
				}
				if he.Prev.Next != he {
					t.Errorf("he.Prev.Next != he in face %d", f.Index)
				}
				if he.Loop != l {
					t.Errorf("he.Loop mismatch in face %d", f.Index)
				}
				reachable[he.Vertex] = true
				n++
				he = he.Next
				if he == start {
					break
				}
				if n > 10000 {
					t.Fatalf("infinite loop in half-edge ring of face %d", f.Index)
					return
				}
			}
		}
	}

	// Euler-Poincaré: V - E + 2F - L = wantEP
	V := s.NVerts()
	E := s.NEdges()
	F := s.NFaces()
	if ep := V - E + 2*F - L; ep != wantEP {
		t.Errorf("Euler-Poincaré: V(%d)-E(%d)+2F(%d)-L(%d) = %d, want %d", V, E, F, L, ep, wantEP)
	}

	// Check 2: mate symmetry + both half-edges non-nil.
	for e := s.Edges; e != nil; e = e.Next {
		if e.He1 == nil || e.He2 == nil {
			t.Errorf("edge %d has nil half-edge", e.Index)
			continue
		}
		if e.He1.GetMate() != e.He2 {
			t.Errorf("edge %d: He1.GetMate() != He2", e.Index)
		}
		if e.He2.GetMate() != e.He1 {
			t.Errorf("edge %d: He2.GetMate() != He1", e.Index)
		}
	}

	// Check 1: every vertex's .He points back to itself.
	for v := s.Verts; v != nil; v = v.Next {
		if v.He == nil {
			t.Errorf("vertex %d has nil He", v.Index)
		} else if v.He.Vertex != v {
			t.Errorf("vertex %d: He.Vertex != self (points to vertex %d)", v.Index, v.He.Vertex.Index)
		}
	}

	// Check 3: every vertex in the solid is reachable from some loop.
	for v := s.Verts; v != nil; v = v.Next {
		if !reachable[v] {
			t.Errorf("vertex %d not reachable from any half-edge ring", v.Index)
		}
	}
}

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
	validateTopology(t, s)
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
	validateTopology(t, s)
}

func TestLmef(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
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
	validateTopology(t, s)
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
	validateTopology(t, s)
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
	validateTopology(t, s)
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
	// Flat quad faces have opposing normals (dot≈-1), so with angle 0.5:
	// -1 < 1-0.5=0.5 → all edges should be creased
	MarkCreases(s, 0.5)
	creased := 0
	for e := s.Edges; e != nil; e = e.Next {
		if e.Mark&base.CREASE != 0 {
			creased++
		}
	}
	if creased != s.NEdges() {
		t.Errorf("expected all %d edges creased with angle 0.5, got %d", s.NEdges(), creased)
	}
	// Clear marks; with very large angle (3.0): -1 < 1-3=-2 is false → no creases
	for e := s.Edges; e != nil; e = e.Next {
		e.Mark = 0
	}
	MarkCreases(s, 3.0)
	creased = 0
	for e := s.Edges; e != nil; e = e.Next {
		if e.Mark&base.CREASE != 0 {
			creased++
		}
	}
	if creased != 0 {
		t.Errorf("expected 0 creased edges with angle 3.0, got %d", creased)
	}
}

// makeQuadWithHole builds a quad face with a 2-vertex inner ring.
// Creates the inner ring correctly using Lkfmrh (not Lkemr).
// Returns the solid, the face with the hole, and the inner loop.
func makeQuadWithHole(t *testing.T) (*base.Solid, *base.Face, *base.Loop) {
	t.Helper()
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 3, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 3, Y: 3, Z: 0})
	v3, _, _ := Lmev(v2.He, vec.SFVec3f{X: 0, Y: 3, Z: 0})
	outerFace, _, _ := Lmef(v0.He, v3.He)

	// Add inner vertices to outerFace's loop (v0.He is in outerFace after Lmef).
	v4, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})
	v5, _, _ := Lmev(v4.He, vec.SFVec3f{X: 2, Y: 1, Z: 0})
	innerFace, _, err := Lmef(v4.He, v5.He)
	if err != nil {
		t.Fatalf("makeQuadWithHole: Lmef inner: %v", err)
	}

	// Convert innerFace's loop into an inner ring of outerFace.
	if err := Lkfmrh(innerFace, outerFace); err != nil {
		t.Fatalf("makeQuadWithHole: Lkfmrh: %v", err)
	}
	if len(outerFace.Loops) < 2 {
		t.Fatalf("makeQuadWithHole: expected 2+ loops, got %d", len(outerFace.Loops))
	}

	var innerLoop *base.Loop
	for _, l := range outerFace.Loops {
		if l != outerFace.LoopOut {
			innerLoop = l
			break
		}
	}
	if innerLoop == nil {
		t.Fatal("makeQuadWithHole: no inner loop")
	}
	return s, outerFace, innerLoop
}

func TestLmfkrh(t *testing.T) {
	s, faceWithHole, innerLoop := makeQuadWithHole(t)
	facesBefore := s.NFaces()
	nf, err := Lmfkrh(innerLoop)
	if err != nil {
		t.Fatalf("Lmfkrh failed: %v", err)
	}
	if nf == nil {
		t.Fatal("Lmfkrh returned nil face")
	}
	if s.NFaces() != facesBefore+1 {
		t.Errorf("expected %d faces after Lmfkrh, got %d", facesBefore+1, s.NFaces())
	}
	if len(nf.Loops) != 1 {
		t.Errorf("new face should have 1 loop, got %d", len(nf.Loops))
	}
	_ = faceWithHole
	validateTopology(t, s)
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

func TestLmev_NilSolid(t *testing.T) {
	f := &base.Face{}
	l := &base.Loop{Face: f}
	f.Loops = []*base.Loop{l}
	f.LoopOut = l
	v := base.NewVertexVec(vec.SFVec3f{})
	he := base.NewHalfEdge(l, v)
	l.AddHalfEdge(he)
	_, _, err := Lmev(he, vec.SFVec3f{X: 1})
	if !errors.Is(err, ErrNilSolid) {
		t.Errorf("expected ErrNilSolid, got %v", err)
	}
}

func TestLringmv(t *testing.T) {
	s, v0, f1 := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})
	v3, _, _ := Lmev(v2.He, vec.SFVec3f{X: 0, Y: 1, Z: 0})
	f2, _, _ := Lmef(v0.He, v3.He)
	_ = f1
	l := f2.LoopOut
	if l == nil {
		t.Fatal("f2 has no outer loop")
	}
	initialF2Loops := len(f2.Loops)
	initialF1Loops := len(f1.Loops)
	err := Lringmv(l, f1, false)
	if err != nil {
		t.Fatalf("Lringmv returned error: %v", err)
	}
	if len(f1.Loops) != initialF1Loops+1 {
		t.Errorf("expected f1 to gain a loop, got %d", len(f1.Loops))
	}
	if len(f2.Loops) != initialF2Loops-1 {
		t.Errorf("expected f2 to lose a loop, got %d", len(f2.Loops))
	}
	validateTopology(t, s)
}

func TestLringmv_NilLoop(t *testing.T) {
	err := Lringmv(nil, &base.Face{}, false)
	if !errors.Is(err, ErrNilLoop) {
		t.Errorf("expected ErrNilLoop, got %v", err)
	}
}

func TestLringmv_NilFace(t *testing.T) {
	l := base.NewLoop(&base.Face{}, true)
	err := Lringmv(l, nil, false)
	if !errors.Is(err, ErrNilFace) {
		t.Errorf("expected ErrNilFace, got %v", err)
	}
}

func TestLkfmrh_DropsInnerLoops(t *testing.T) {
	s, faceWithHole, _ := makeQuadWithHole(t)

	keepFace := s.Faces
	if keepFace == faceWithHole {
		keepFace = keepFace.Next
	}
	if keepFace == nil {
		t.Fatal("no keepFace available")
		return
	}

	innerLoopCount := len(faceWithHole.Loops) - 1
	totalLoopsBefore := len(keepFace.Loops)

	err := Lkfmrh(faceWithHole, keepFace)
	if err != nil {
		t.Fatalf("Lkfmrh failed: %v", err)
	}

	expectedLoops := totalLoopsBefore + 1 + innerLoopCount
	if len(keepFace.Loops) != expectedLoops {
		t.Errorf("Lkfmrh: keepFace has %d loops, expected %d (inner loops dropped!)",
			len(keepFace.Loops), expectedLoops)
	}
}

// --- Dedicated operator tests ---

func TestLkef(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})
	_, closingEdge, err := Lmef(v0.He, v2.He)
	if err != nil {
		t.Fatalf("Lmef: %v", err)
	}
	validateTopology(t, s)

	if err := Lkef(closingEdge.He1); err != nil {
		t.Fatalf("Lkef: %v", err)
	}
	if s.NVerts() != 3 {
		t.Errorf("expected 3 verts after Lkef, got %d", s.NVerts())
	}
	if s.NEdges() != 2 {
		t.Errorf("expected 2 edges after Lkef, got %d", s.NEdges())
	}
	if s.NFaces() != 1 {
		t.Errorf("expected 1 face after Lkef, got %d", s.NFaces())
	}
	validateTopology(t, s)
}

func TestLkemr(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 3, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 3, Y: 3, Z: 0})
	v3, _, _ := Lmev(v2.He, vec.SFVec3f{X: 0, Y: 3, Z: 0})
	_, _, _ = Lmef(v0.He, v3.He)
	v4, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})
	v5, _, _ := Lmev(v4.He, vec.SFVec3f{X: 2, Y: 1, Z: 0})
	_, eInner, _ := Lmef(v4.He, v5.He)

	edgesBefore := s.NEdges()
	err := Lkemr(eInner.He1)
	if err != nil {
		t.Fatalf("Lkemr: %v", err)
	}
	if s.NEdges() != edgesBefore-1 {
		t.Errorf("expected %d edges after Lkemr, got %d", edgesBefore-1, s.NEdges())
	}
	// eInner's half-edges were in different loops (cross-face, from Lmef).
	// Lkemr merges the two loops into one — no inner ring is created.
	// The mate's face is absorbed into he's face.
	validateTopology(t, s)
}

func TestLmekr(t *testing.T) {
	s, faceWithHole, innerLoop := makeQuadWithHole(t)

	he1 := faceWithHole.LoopOut.HalfEdges
	he2 := innerLoop.HalfEdges

	edgesBefore := s.NEdges()
	ne, err := Lmekr(he1, he2)
	if err != nil {
		t.Fatalf("Lmekr: %v", err)
	}
	if ne == nil {
		t.Fatal("Lmekr returned nil edge")
	}
	if s.NEdges() != edgesBefore+1 {
		t.Errorf("expected %d edges after Lmekr, got %d", edgesBefore+1, s.NEdges())
	}
	validateTopologyEP(t, s, 0) // genus-1: face-with-hole topology
}

func TestLmev2(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	_, _, _ = Lmev(v1.He, vec.SFVec3f{X: 2, Y: 0, Z: 0})
	// V=3, E=2, F=1

	var heA, heB *base.HalfEdge
	for f := s.Faces; f != nil; f = f.Next {
		for _, l := range f.Loops {
			he := l.HalfEdges
			start := he
			for {
				if he.Vertex == v1 {
					if heA == nil {
						heA = he
					} else if he != heA && heB == nil {
						heB = he
					}
				}
				he = he.Next
				if he == start {
					break
				}
			}
		}
	}
	if heA == nil || heB == nil {
		t.Fatal("could not find two half-edges at v1")
		return
	}

	nv, ne, err := Lmev2(heA, heB, vec.SFVec3f{X: 1, Y: 0.5, Z: 0})
	if err != nil {
		t.Fatalf("Lmev2: %v", err)
	}
	if nv == nil || ne == nil {
		t.Fatal("Lmev2 returned nil")
		return
	}
	if s.NVerts() != 4 {
		t.Errorf("expected 4 verts after Lmev2, got %d", s.NVerts())
	}
	if s.NEdges() != 3 {
		t.Errorf("expected 3 edges after Lmev2, got %d", s.NEdges())
	}
	validateTopology(t, s)
}

func TestLmev2_ErrorCases(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1})
	_, _, err := Lmev2(nil, v0.He, vec.SFVec3f{})
	if !errors.Is(err, ErrNilHalfEdge) {
		t.Errorf("expected ErrNilHalfEdge, got %v", err)
	}
	_, _, err = Lmev2(v0.He, v1.He, vec.SFVec3f{})
	if !errors.Is(err, ErrDifferentVerts) {
		t.Errorf("expected ErrDifferentVerts, got %v", err)
	}
}

// --- Round-trip tests ---

func TestRoundTrip_LmevLkev(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	validateTopology(t, s)

	_, ne, err := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	if err != nil {
		t.Fatalf("Lmev: %v", err)
	}
	if s.NVerts() != 2 || s.NEdges() != 1 {
		t.Fatalf("after Lmev: V=%d E=%d", s.NVerts(), s.NEdges())
	}
	validateTopology(t, s)

	if err := Lkev(ne.He1); err != nil {
		t.Fatalf("Lkev: %v", err)
	}
	if s.NVerts() != 1 || s.NEdges() != 0 || s.NFaces() != 1 {
		t.Errorf("round-trip: V=%d E=%d F=%d, want 1/0/1",
			s.NVerts(), s.NEdges(), s.NFaces())
	}
	validateTopology(t, s)
}

func TestRoundTrip_LmefLkef(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})
	validateTopology(t, s)

	_, closingEdge, err := Lmef(v0.He, v2.He)
	if err != nil {
		t.Fatalf("Lmef: %v", err)
	}
	if s.NFaces() != 2 || s.NEdges() != 3 {
		t.Fatalf("after Lmef: F=%d E=%d", s.NFaces(), s.NEdges())
	}
	validateTopology(t, s)

	if err := Lkef(closingEdge.He1); err != nil {
		t.Fatalf("Lkef: %v", err)
	}
	if s.NVerts() != 3 || s.NEdges() != 2 || s.NFaces() != 1 {
		t.Errorf("round-trip: V=%d E=%d F=%d, want 3/2/1",
			s.NVerts(), s.NEdges(), s.NFaces())
	}
	validateTopology(t, s)
}

func TestRoundTrip_LkemrLmekr(t *testing.T) {
	s, faceWithHole, innerLoop := makeQuadWithHole(t)

	he1 := faceWithHole.LoopOut.HalfEdges
	he2 := innerLoop.HalfEdges

	// Bridge the two loops with Lmekr → one loop, one bridge edge.
	bridge, err := Lmekr(he1, he2)
	if err != nil {
		t.Fatalf("Lmekr: %v", err)
	}
	eBefore := s.NEdges()
	validateTopologyEP(t, s, 0) // genus-1 after Lmekr bridges hole

	// Lkemr on the bridge edge (same-loop) → splits back into outer+inner.
	if err := Lkemr(bridge.He1); err != nil {
		t.Fatalf("Lkemr: %v", err)
	}
	if s.NEdges() != eBefore-1 {
		t.Fatalf("after Lkemr: E=%d, want %d", s.NEdges(), eBefore-1)
	}

	// Should have an inner loop again.
	found := false
	for f := s.Faces; f != nil; f = f.Next {
		if len(f.Loops) > 1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected face with inner loop after same-loop Lkemr")
	}
	validateTopologyEP(t, s, 0) // genus-1: hole restored
}

func TestRoundTrip_LmfkrhLkfmrh(t *testing.T) {
	s, faceWithHole, innerLoop := makeQuadWithHole(t)
	fBefore := s.NFaces()

	nf, err := Lmfkrh(innerLoop)
	if err != nil {
		t.Fatalf("Lmfkrh: %v", err)
	}
	if s.NFaces() != fBefore+1 {
		t.Fatalf("after Lmfkrh: F=%d, want %d", s.NFaces(), fBefore+1)
	}
	validateTopology(t, s)

	if err := Lkfmrh(nf, faceWithHole); err != nil {
		t.Fatalf("Lkfmrh: %v", err)
	}
	if s.NFaces() != fBefore {
		t.Errorf("round-trip: F=%d, want %d", s.NFaces(), fBefore)
	}
	validateTopologyEP(t, s, 0) // genus-1: hole restored
}

// --- Shape diversity tests (#61) ---

func TestTriangleTopology(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 0.5, Y: 1, Z: 0})
	_, _, err := Lmef(v0.He, v2.He)
	if err != nil {
		t.Fatalf("Lmef: %v", err)
	}
	if s.NVerts() != 3 {
		t.Errorf("V=%d, want 3", s.NVerts())
	}
	if s.NEdges() != 3 {
		t.Errorf("E=%d, want 3", s.NEdges())
	}
	if s.NFaces() != 2 {
		t.Errorf("F=%d, want 2", s.NFaces())
	}
	validateTopology(t, s)
}

func TestPentagonTopology(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{X: 1, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 0.309, Y: 0.951, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: -0.809, Y: 0.588, Z: 0})
	v3, _, _ := Lmev(v2.He, vec.SFVec3f{X: -0.809, Y: -0.588, Z: 0})
	v4, _, _ := Lmev(v3.He, vec.SFVec3f{X: 0.309, Y: -0.951, Z: 0})
	_, _, err := Lmef(v0.He, v4.He)
	if err != nil {
		t.Fatalf("Lmef: %v", err)
	}
	if s.NVerts() != 5 {
		t.Errorf("V=%d, want 5", s.NVerts())
	}
	if s.NEdges() != 5 {
		t.Errorf("E=%d, want 5", s.NEdges())
	}
	if s.NFaces() != 2 {
		t.Errorf("F=%d, want 2", s.NFaces())
	}
	validateTopology(t, s)
}

// makeCube, fan, and pyramid tests are deferred — building 3D multi-face
// topology from raw Euler ops requires TranslationalSweep or a working
// buildFace (#56), neither of which is in this package yet. See #61.

// --- Deep error path / edge-case tests (#63) ---

func TestLmef_SameHalfEdge(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	_, v0, _ := Mvfs(vec.SFVec3f{}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1})
	// he1 == he2: degenerate — creates a zero-length self-loop edge.
	// Must not panic.
	_, _, _ = Lmef(v1.He, v1.He)
}

func TestLkef_NilEdge(t *testing.T) {
	f := &base.Face{Solid: base.NewSolid()}
	l := base.NewLoop(f, true)
	he := base.NewHalfEdge(l, base.NewVertexVec(vec.SFVec3f{}))
	l.AddHalfEdge(he)
	err := Lkef(he)
	if !errors.Is(err, ErrNilEdge) {
		t.Errorf("expected ErrNilEdge, got %v", err)
	}
}

func TestLkemr_NilEdge(t *testing.T) {
	f := &base.Face{Solid: base.NewSolid()}
	l := base.NewLoop(f, true)
	he := base.NewHalfEdge(l, base.NewVertexVec(vec.SFVec3f{}))
	l.AddHalfEdge(he)
	err := Lkemr(he)
	if !errors.Is(err, ErrNilEdge) {
		t.Errorf("expected ErrNilEdge, got %v", err)
	}
}

func TestLkev_LastEdge(t *testing.T) {
	s, v0, _ := Mvfs(vec.SFVec3f{}, vec.White)
	_, _, _ = Lmev(v0.He, vec.SFVec3f{X: 1})
	err := Lkev(v0.He)
	if err != nil {
		t.Fatalf("Lkev last edge: %v", err)
	}
	if s.NVerts() != 1 {
		t.Errorf("V=%d, want 1", s.NVerts())
	}
	if s.NEdges() != 0 {
		t.Errorf("E=%d, want 0", s.NEdges())
	}
	validateTopology(t, s)
}

func TestLkev_DoubleKill(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic on double-kill: %v", r)
		}
	}()
	_, v0, _ := Mvfs(vec.SFVec3f{}, vec.White)
	_, _, _ = Lmev(v0.He, vec.SFVec3f{X: 1})
	he := v0.He
	_ = Lkev(he)
	// Second kill on the same (now detached) half-edge — must not panic.
	_ = Lkev(he)
}

func TestLkfmrh_SameFace(t *testing.T) {
	_, v0, _ := Mvfs(vec.SFVec3f{}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 1})
	v3, _, _ := Lmev(v2.He, vec.SFVec3f{X: 0, Y: 1})
	f2, _, _ := Lmef(v0.He, v3.He)
	err := Lkfmrh(f2, f2)
	if !errors.Is(err, ErrSameFace) {
		t.Errorf("expected ErrSameFace, got %v", err)
	}
}

func TestLmfkrh_OuterLoop(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	s, _, f := Mvfs(vec.SFVec3f{}, vec.White)
	// Promoting the sole outer loop to a new face leaves original face empty.
	// Must not panic.
	nf, err := Lmfkrh(f.LoopOut)
	if err != nil {
		return // error is acceptable
	}
	// If it succeeds, the new face should exist and the solid should have 2 faces.
	if nf == nil {
		t.Error("Lmfkrh returned nil face without error")
	}
	_ = s
}

func TestLkef_SameFaceBothSides(t *testing.T) {
	// Create a face with an inner loop via makeQuadWithHole, then bridge with Lmekr.
	// The bridge edge has both half-edges in the same loop.
	// Lkef on it should return ErrSameLoop.
	_, faceWithHole, innerLoop := makeQuadWithHole(t)

	he1 := faceWithHole.LoopOut.HalfEdges
	bridge, err := Lmekr(he1, innerLoop.HalfEdges)
	if err != nil {
		t.Fatalf("Lmekr: %v", err)
	}

	// Both half-edges of bridge are in the same loop of the same face.
	// Lkef can't split a loop — it returns ErrSameLoop (use Lkemr instead).
	err = Lkef(bridge.He1)
	if !errors.Is(err, ErrSameLoop) {
		t.Errorf("expected ErrSameLoop, got %v", err)
	}
}

// --- Half-edge selection stress tests (#62) ---

// findHeAt walks every half-edge ring in s and returns a half-edge at vertex v
// that is NOT v.He. Returns nil if v has valence < 2 or no alternative exists.
func findHeAt(s *base.Solid, v *base.Vertex) *base.HalfEdge {
	for f := s.Faces; f != nil; f = f.Next {
		for _, l := range f.Loops {
			if l.HalfEdges == nil {
				continue
			}
			he := l.HalfEdges
			start := he
			for {
				if he.Vertex == v && he != v.He {
					return he
				}
				he = he.Next
				if he == start {
					break
				}
			}
		}
	}
	return nil
}

func TestLmev2_HighValence_NonDefaultHe(t *testing.T) {
	// Build a 4-spoke star: center vertex v0 with 4 edges radiating outward.
	// v0 has valence 4.
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	v2, _, _ := Lmev(v0.He, vec.SFVec3f{X: 0, Y: 1, Z: 0})
	v3, _, _ := Lmev(v0.He, vec.SFVec3f{X: -1, Y: 0, Z: 0})
	v4, _, _ := Lmev(v0.He, vec.SFVec3f{X: 0, Y: -1, Z: 0})
	_ = v1
	_ = v2
	_ = v3
	_ = v4

	if v0.GetValence() < 4 {
		t.Fatalf("expected valence >= 4, got %d", v0.GetValence())
	}

	// Find two distinct half-edges at v0 (neither is v0.He).
	heA := findHeAt(s, v0)
	if heA == nil {
		t.Fatal("could not find non-default half-edge at v0")
	}
	heB := v0.He
	if heA == heB {
		t.Fatal("heA == heB, need distinct half-edges")
	}

	// Lmev2 splits v0 between heA and heB.
	nv, ne, err := Lmev2(heA, heB, vec.SFVec3f{X: 0.5, Y: 0.5, Z: 0})
	if err != nil {
		t.Fatalf("Lmev2: %v", err)
	}
	if nv == nil || ne == nil {
		t.Fatal("Lmev2 returned nil")
	}
	if s.NVerts() != 6 {
		t.Errorf("V=%d, want 6", s.NVerts())
	}
	if s.NEdges() != 5 {
		t.Errorf("E=%d, want 5", s.NEdges())
	}
	validateTopology(t, s)
}

func TestLmef_NonDefaultHe_Valence3(t *testing.T) {
	// Build a chain: v0 -- v1 -- v2, then add a spoke v1 -- v3.
	// v1 has valence 3. Close a face using a specific (non-default) half-edge at v1.
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	v1, _, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	v2, _, _ := Lmev(v1.He, vec.SFVec3f{X: 2, Y: 0, Z: 0})
	v3, _, _ := Lmev(v1.He, vec.SFVec3f{X: 1, Y: 1, Z: 0})

	if v1.GetValence() < 3 {
		t.Fatalf("expected v1 valence >= 3, got %d", v1.GetValence())
	}

	// Walk to find half-edges at v0 and v3 in the same loop so Lmef can close.
	// v0.He and v3.He should both be in the single loop.
	if v0.He.Loop != v3.He.Loop {
		// Try alternate half-edges — find one at v3 in v0's loop.
		alt := findHeAt(s, v3)
		if alt != nil && alt.Loop == v0.He.Loop {
			v3.He = alt
		}
	}
	// Also try non-default he at v0 if it's in the same loop as v3.He.
	altV0 := findHeAt(s, v0)
	heForV0 := v0.He
	if altV0 != nil && altV0.Loop == v3.He.Loop {
		heForV0 = altV0
	}

	nf, ne, err := Lmef(heForV0, v3.He)
	if err != nil {
		t.Fatalf("Lmef with non-default he: %v", err)
	}
	if nf == nil || ne == nil {
		t.Fatal("Lmef returned nil")
	}

	_ = v2
	validateTopology(t, s)
}

func TestLkev_HighValence(t *testing.T) {
	// Build a 3-spoke star at v0, then kill one spoke.
	// Remaining topology must be valid.
	s, v0, _ := Mvfs(vec.SFVec3f{X: 0, Y: 0, Z: 0}, vec.White)
	_, e1, _ := Lmev(v0.He, vec.SFVec3f{X: 1, Y: 0, Z: 0})
	_, _, _ = Lmev(v0.He, vec.SFVec3f{X: 0, Y: 1, Z: 0})
	_, _, _ = Lmev(v0.He, vec.SFVec3f{X: -1, Y: 0, Z: 0})

	if v0.GetValence() < 3 {
		t.Fatalf("expected valence >= 3, got %d", v0.GetValence())
	}

	// Kill the first spoke using a specific half-edge (He1 of e1).
	err := Lkev(e1.He1)
	if err != nil {
		t.Fatalf("Lkev on high-valence vertex: %v", err)
	}
	if s.NVerts() != 3 {
		t.Errorf("V=%d, want 3", s.NVerts())
	}
	if s.NEdges() != 2 {
		t.Errorf("E=%d, want 2", s.NEdges())
	}
	validateTopology(t, s)
}

// --- Multi-face BuildFromIndexSet tests (#64) ---

func TestBuildFromIndexSet_Triangle(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0.5, Y: 1, Z: 0},
	}
	indices := []int64{0, 1, 2, -1}
	s, err := BuildFromIndexSet(positions, indices, vec.White)
	if err != nil {
		t.Fatalf("BuildFromIndexSet triangle: %v", err)
	}
	if s.NVerts() != 3 {
		t.Errorf("V=%d, want 3", s.NVerts())
	}
	if s.NEdges() != 3 {
		t.Errorf("E=%d, want 3", s.NEdges())
	}
	if s.NFaces() != 1 {
		t.Errorf("F=%d, want 1", s.NFaces())
	}
	// Single face is an open surface — boundary edges have nil He1.
	// Full validateTopology not applicable.
}

func TestBuildFromIndexSet_NoTrailingSeparator(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
	}
	// No trailing -1
	indices := []int64{0, 1, 2}
	s, err := BuildFromIndexSet(positions, indices, vec.White)
	if err != nil {
		t.Fatalf("BuildFromIndexSet no trailing -1: %v", err)
	}
	if s.NVerts() != 3 {
		t.Errorf("V=%d, want 3", s.NVerts())
	}
	if s.NFaces() != 1 {
		t.Errorf("F=%d, want 1", s.NFaces())
	}
}

func TestBuildFromIndexSet_DegenerateFacesSkipped(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	// First segment has only 2 verts (degenerate, should be skipped).
	// Second segment is a valid quad.
	indices := []int64{0, 1, -1, 0, 1, 2, 3, -1}
	s, err := BuildFromIndexSet(positions, indices, vec.White)
	if err != nil {
		t.Fatalf("BuildFromIndexSet degenerate: %v", err)
	}
	if s.NVerts() != 4 {
		t.Errorf("V=%d, want 4", s.NVerts())
	}
	if s.NFaces() != 1 {
		t.Errorf("F=%d, want 1", s.NFaces())
	}
}

func TestBuildFromIndexSet_AllDegenerate(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
	}
	// Only degenerate faces (< 3 verts each).
	indices := []int64{0, 1, -1, 0, -1}
	_, err := BuildFromIndexSet(positions, indices, vec.White)
	if err == nil {
		t.Error("expected error for all-degenerate faces")
	}
}

func TestBuildFromIndexSet_OutOfBounds(t *testing.T) {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
	}
	indices := []int64{0, 1, 99, -1}
	_, err := BuildFromIndexSet(positions, indices, vec.White)
	if err == nil {
		t.Error("expected error for out-of-bounds index")
	}
}

func TestBuildFromIndexSet_TwoFaces(t *testing.T) {
	// Two triangles sharing edge 1-2 with consistent winding:
	//   face0: [0,1,2]  face1: [2,1,3]
	// Edge 1→2 from face0 pairs with 2→1 from face1.
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0.5, Y: 1, Z: 0},
		{X: 1.5, Y: 1, Z: 0},
	}
	indices := []int64{0, 1, 2, -1, 2, 1, 3, -1}
	s, err := BuildFromIndexSet(positions, indices, vec.White)
	if err != nil {
		t.Fatalf("BuildFromIndexSet two faces: %v", err)
	}
	if s.NVerts() != 4 {
		t.Errorf("V=%d, want 4", s.NVerts())
	}
	if s.NEdges() != 5 {
		t.Errorf("E=%d, want 5", s.NEdges())
	}
	if s.NFaces() != 2 {
		t.Errorf("F=%d, want 2", s.NFaces())
	}
	// Open surface — 4 boundary edges, 1 shared. Full topology check N/A.
}

func TestBuildFromIndexSet_Tetrahedron(t *testing.T) {
	// 4 triangular faces with consistent outward-facing winding.
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0.5, Y: 1, Z: 0},
		{X: 0.5, Y: 0.5, Z: 1},
	}
	indices := []int64{
		0, 2, 1, -1, // bottom
		0, 1, 3, -1, // front
		1, 2, 3, -1, // right
		0, 3, 2, -1, // left
	}
	s, err := BuildFromIndexSet(positions, indices, vec.White)
	if err != nil {
		t.Fatalf("BuildFromIndexSet tetrahedron: %v", err)
	}
	// Tetrahedron: V=4, E=6, F=4
	if s.NVerts() != 4 {
		t.Errorf("V=%d, want 4", s.NVerts())
	}
	if s.NEdges() != 6 {
		t.Errorf("E=%d, want 6", s.NEdges())
	}
	if s.NFaces() != 4 {
		t.Errorf("F=%d, want 4", s.NFaces())
	}
	validateTopology(t, s)
}

func TestBuildFromIndexSet_Cube(t *testing.T) {
	// 6 quad faces with consistent outward-facing winding.
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0}, // 0
		{X: 1, Y: 0, Z: 0}, // 1
		{X: 1, Y: 1, Z: 0}, // 2
		{X: 0, Y: 1, Z: 0}, // 3
		{X: 0, Y: 0, Z: 1}, // 4
		{X: 1, Y: 0, Z: 1}, // 5
		{X: 1, Y: 1, Z: 1}, // 6
		{X: 0, Y: 1, Z: 1}, // 7
	}
	indices := []int64{
		0, 3, 2, 1, -1, // bottom (outward = -Z)
		4, 5, 6, 7, -1, // top (outward = +Z)
		0, 1, 5, 4, -1, // front (outward = -Y)
		2, 3, 7, 6, -1, // back (outward = +Y)
		0, 4, 7, 3, -1, // left (outward = -X)
		1, 2, 6, 5, -1, // right (outward = +X)
	}
	s, err := BuildFromIndexSet(positions, indices, vec.White)
	if err != nil {
		t.Fatalf("BuildFromIndexSet cube: %v", err)
	}
	// Cube: V=8, E=12, F=6
	if s.NVerts() != 8 {
		t.Errorf("V=%d, want 8", s.NVerts())
	}
	if s.NEdges() != 12 {
		t.Errorf("E=%d, want 12", s.NEdges())
	}
	if s.NFaces() != 6 {
		t.Errorf("F=%d, want 6", s.NFaces())
	}
	validateTopology(t, s)
}
