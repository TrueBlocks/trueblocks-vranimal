package solid

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Euler operators - fundamental topological operations on B-rep solids.
// Ported from vraniml/src/solid/eulerops.cpp
// ---------------------------------------------------------------------------

// Mvfs creates a new solid with one vertex and one face (Make Vertex Face Solid).
func Mvfs(loc vec.SFVec3f, color vec.SFColor) (*Solid, *Vertex, *Face) {
	s := NewSolid()
	v := NewVertexVec(loc)
	s.AddVertex(v)
	f := NewFace(s, color)
	s.AddFace(f)
	l := NewLoop(f, true)
	he := NewHalfEdge(l, v)
	l.AddHalfEdge(he)
	v.He = he
	return s, v, f
}

// Kvfs destroys a solid with one vertex and one face (Kill Vertex Face Solid).
func Kvfs(s *Solid) {
	s.Faces = nil
	s.Edges = nil
	s.Verts = nil
	s.nFaces = 0
	s.nEdges = 0
	s.nVerts = 0
}

// insertBefore inserts nhe immediately before he in the circular half-edge ring.
func insertBefore(he, nhe *HalfEdge) {
	nhe.Next = he
	nhe.Prev = he.Prev
	he.Prev.Next = nhe
	he.Prev = nhe
}

// Lmev creates a new vertex and edge at a half-edge (Low Make Edge Vertex).
// Ported from vraniml/src/solid/eulerops.cpp lmev (he1==he2 case).
func Lmev(he *HalfEdge, loc vec.SFVec3f) (*Vertex, *Edge) {
	s := he.GetFace().Solid
	nv := NewVertexVec(loc)
	s.AddVertex(nv)

	ne := NewEdge()
	s.AddEdge(ne)

	if he.Edge == nil {
		// First edge from bare Mvfs vertex: reuse he for He2, create new for He1.
		ne.He2 = he
		he.Edge = ne

		nhe := &HalfEdge{Vertex: nv, Edge: ne, Loop: he.Loop}
		ne.He1 = nhe
		insertBefore(he, nhe)

		nv.He = nhe
		he.Vertex.He = he
	} else {
		// Subsequent edges: create two new half-edges before he.
		nhe2 := &HalfEdge{Vertex: he.Vertex, Edge: ne, Loop: he.Loop}
		ne.He2 = nhe2
		insertBefore(he, nhe2)

		nhe1 := &HalfEdge{Vertex: nv, Edge: ne, Loop: he.Loop}
		ne.He1 = nhe1
		insertBefore(he, nhe1)

		nv.He = nhe1
		he.Vertex.He = he
	}

	return nv, ne
}

// Lmev2 splits a vertex between two half-edges in potentially different loops,
// creating a new vertex and edge. This is the two-argument form from the C++ lmev.
// he1 and he2 must share the same vertex; half-edges from he1 to he2 (via mate orbit)
// are reassigned to the new vertex.
// Ported from vraniml/src/solid/eulerops.cpp lmev(he1, he2, ...).
func Lmev2(he1, he2 *HalfEdge, loc vec.SFVec3f) (*Vertex, *Edge) {
	if he1.Vertex != he2.Vertex {
		return nil, nil
	}
	s := he1.GetFace().Solid
	nv := NewVertexVec(loc)
	s.AddVertex(nv)

	ne := NewEdge()
	s.AddEdge(ne)

	// Walk from he1 around the vertex orbit until he2, reassigning to nv
	he := he1
	for he != he2 {
		he.Vertex = nv
		m := he.GetMate()
		if m == nil {
			break
		}
		he = m.Next
	}

	// Insert new half-edges using the edge's AddHalfEdge pattern
	nhe2 := &HalfEdge{Vertex: he2.Vertex, Edge: ne, Loop: he1.Loop}
	ne.He2 = nhe2
	insertBefore(he1, nhe2)

	nhe1 := &HalfEdge{Vertex: nv, Edge: ne, Loop: he2.Loop}
	ne.He1 = nhe1
	insertBefore(he2, nhe1)

	nv.He = nhe1
	he2.Vertex.He = he2

	return nv, ne
}

// Lmef creates a new edge and face by splitting a loop (Low Make Edge Face).
// he1 and he2 are half-edges in the same loop; a new edge connects their vertices,
// splitting the loop into two faces.
// Ported from vraniml/src/solid/eulerops.cpp lmef.
func Lmef(he1, he2 *HalfEdge) (*Face, *Edge) {
	old := he1.GetFace()
	s := old.Solid

	nf := NewFace(s, old.GetColor(vec.White))
	s.AddFace(nf)
	ne := NewEdge()
	s.AddEdge(ne)
	nl := NewLoop(nf, true)

	// Step 1: Walk from he1 to he2, reassigning to the new loop.
	for he := he1; he != he2; he = he.Next {
		he.Loop = nl
	}

	// Step 2: Insert two new half-edges into the ring (before splitting).
	// nhe1 goes before he1 (which is now in nl).
	nhe1 := &HalfEdge{Vertex: he2.Vertex, Edge: ne, Loop: he1.Loop}
	ne.He2 = nhe1
	nhe1.Edge = ne
	insertBefore(he1, nhe1)

	// nhe2 goes before he2 (which stays in the old loop).
	nhe2 := &HalfEdge{Vertex: he1.Vertex, Edge: ne, Loop: he2.Loop}
	ne.He1 = nhe2
	nhe2.Edge = ne
	insertBefore(he2, nhe2)

	// Step 3: Split the ring by swapping prev/next connections.
	nhe1.Prev.Next = nhe2
	nhe2.Prev.Next = nhe1

	temp := nhe2.Prev
	nhe2.Prev = nhe1.Prev
	nhe1.Prev = temp

	// Step 4: Set loop heads.
	he2.Loop.SetFirstHe(nhe2) // old loop
	nl.SetFirstHe(nhe1)       // new loop

	return nf, ne
}

// Lkev kills an edge and vertex, merging two vertices (Low Kill Edge Vertex).
// he.Vertex is killed; its mate's vertex survives. All half-edges that
// referenced the killed vertex are redirected to the surviving vertex.
func Lkev(he *HalfEdge) {
	if he.Edge == nil {
		return
	}
	s := he.GetFace().Solid
	e := he.Edge
	mate := he.GetMate()

	killV := he.Vertex
	keepV := mate.Vertex

	// Walk the vertex orbit of killV, reassigning all half-edges to keepV.
	// Only reassign half-edges that actually reference killV.
	cur := mate.Next
	for cur != he {
		if cur.Vertex != killV {
			break
		}
		cur.Vertex = keepV
		m := cur.GetMate()
		if m == nil {
			break
		}
		cur = m.Next
	}

	// Update surviving vertex's half-edge reference.
	if he.Prev == mate {
		// Null strut: he and mate are adjacent.
		keepV.He = mate.Next.Next
	} else {
		keepV.He = mate.Next
	}
	if keepV.He != nil && keepV.He.Edge == nil {
		keepV.He = nil
	}

	// Remove he from its loop.
	he.Prev.Next = he.Next
	he.Next.Prev = he.Prev
	if he.Loop.HalfEdges == he {
		he.Loop.HalfEdges = he.Next
	}

	// Remove mate from its loop.
	if mate != nil && mate != he {
		mate.Prev.Next = mate.Next
		mate.Next.Prev = mate.Prev
		if mate.Loop.HalfEdges == mate {
			mate.Loop.HalfEdges = mate.Next
		}
	}

	s.RemoveEdge(e)
	s.RemoveVertex(killV)
}

// Lkef kills an edge and face, merging two faces (Low Kill Edge Face).
// he's face survives; he's mate's face is killed.
func Lkef(he *HalfEdge) {
	if he.Edge == nil {
		return
	}
	s := he.GetFace().Solid
	e := he.Edge
	mate := he.GetMate()
	keepF := he.GetFace()
	killF := mate.GetFace()

	// Move all loops from the killed face to the keeping face.
	if killF != keepF {
		for l := killF.LoopOut; l != nil; l = killF.LoopOut {
			killF.RemoveLoop(l)
			keepF.AddLoop(l, false)
		}
	}

	// Reassign all half-edges in the killed loop to the keeping loop.
	keepLoop := he.Loop
	killLoop := mate.Loop
	cur := mate
	for {
		cur.Loop = keepLoop
		cur = cur.Next
		if cur == mate {
			break
		}
	}

	// Splice out he and mate, joining the two loops into one ring.
	he.Prev.Next = mate.Next
	mate.Next.Prev = he.Prev
	mate.Prev.Next = he.Next
	he.Next.Prev = mate.Prev

	// Update vertex half-edge references (they might point to removed HEs).
	v1 := he.Vertex
	v2 := mate.Vertex
	v1.He = mate.Next
	if v1.He != nil && v1.He.Edge == nil {
		v1.He = nil
	}
	v2.He = he.Next
	if v2.He != nil && v2.He.Edge == nil {
		v2.He = nil
	}

	keepLoop.SetFirstHe(he.Prev)

	// Remove the now-empty killed loop from the keeping face.
	killLoop.HalfEdges = nil
	keepF.RemoveLoop(killLoop)

	s.RemoveEdge(e)
	if killF != keepF {
		s.RemoveFace(killF)
	}
}

// Lkemr kills an edge, creating a new inner ring (Kill Edge Make Ring).
func Lkemr(he *HalfEdge) {
	if he.Edge == nil {
		return
	}
	s := he.GetFace().Solid
	e := he.Edge
	mate := he.GetMate()
	f := he.GetFace()

	// Create new inner loop
	nl := NewLoop(f, false)

	// Disconnect he and mate from the ring
	he.Prev.Next = mate.Next
	mate.Next.Prev = he.Prev

	mate.Prev.Next = he.Next
	he.Next.Prev = mate.Prev

	// Old loop keeps he.Prev side
	he.Loop.SetFirstHe(he.Prev)

	// New loop gets the mate.Prev side
	nl.HalfEdges = mate.Prev
	cur := mate.Prev
	for {
		cur.Loop = nl
		cur = cur.Next
		if cur == mate.Prev {
			break
		}
	}

	s.RemoveEdge(e)
}

// Lmekr makes an edge, killing an inner ring (Make Edge Kill Ring).
// he1 and he2 are in different loops of the same face. A new edge
// connecting he1.Vertex and he2.Vertex joins the two loops into one.
func Lmekr(he1, he2 *HalfEdge) *Edge {
	s := he1.GetFace().Solid
	ne := NewEdge()
	s.AddEdge(ne)

	nhe1 := NewHalfEdge(he1.Loop, he2.Vertex)
	nhe2 := NewHalfEdge(he1.Loop, he1.Vertex)
	ne.He1 = nhe1
	ne.He2 = nhe2
	nhe1.Edge = ne
	nhe2.Edge = ne

	// Save prev pointers before any splicing.
	he1prev := he1.Prev
	he2prev := he2.Prev

	// Splice: nhe1 bridges from loop2's tail to loop1's head.
	// nhe2 bridges from loop1's tail to loop2's head.
	// Result: ...he2prev → nhe1 → he1 → ... → he1prev → nhe2 → he2 → ... → he2prev...
	nhe1.Prev = he2prev
	nhe1.Next = he1
	he2prev.Next = nhe1
	he1.Prev = nhe1

	nhe2.Prev = he1prev
	nhe2.Next = he2
	he1prev.Next = nhe2
	he2.Prev = nhe2

	// All half-edges now in one loop
	cur := nhe1
	for {
		cur.Loop = he1.Loop
		cur = cur.Next
		if cur == nhe1 {
			break
		}
	}

	// Remove the inner loop from the face
	f := he1.GetFace()
	if he2.Loop != he1.Loop {
		f.RemoveLoop(he2.Loop)
	}

	he1.Loop.SetFirstHe(nhe1)
	return ne
}

// Lmfkrh promotes an inner ring to a new face (Make Face Kill Ring Hole).
func Lmfkrh(innerLoop *Loop) *Face {
	f := innerLoop.Face
	s := f.Solid
	nf := NewFace(s, f.GetColor(vec.White))
	s.AddFace(nf)

	f.RemoveLoop(innerLoop)
	nf.AddLoop(innerLoop, true)

	// Reassign half-edge loops
	innerLoop.ForEachHe(func(he *HalfEdge) bool {
		he.Loop = innerLoop
		return true
	})

	return nf
}

// Lkfmrh kills a face, creating an inner ring in another face (Kill Face Make Ring Hole).
func Lkfmrh(killFace, keepFace *Face) {
	s := killFace.Solid
	l := killFace.LoopOut
	killFace.RemoveLoop(l)
	keepFace.AddLoop(l, false)

	// Reassign half-edge loops
	l.ForEachHe(func(he *HalfEdge) bool {
		he.Loop = l
		return true
	})

	s.RemoveFace(killFace)
}

// Lringmv moves a loop from one face to another.
func Lringmv(s *Solid, l *Loop, toFace *Face, isOuter bool) {
	fromFace := l.Face
	fromFace.RemoveLoop(l)
	toFace.AddLoop(l, isOuter)
}

// ---------------------------------------------------------------------------
// High-level operators (by vertex index or position)
// ---------------------------------------------------------------------------

// Mev creates a new edge and vertex from an existing vertex (Make Edge Vertex).
func Mev(s *Solid, fromVertex *Vertex, loc vec.SFVec3f) (*Vertex, *Edge) {
	if fromVertex.He == nil {
		return nil, nil
	}
	return Lmev(fromVertex.He, loc)
}

// Mef creates a new edge and face connecting two vertices in the same face.
func Mef(s *Solid, v1, v2 *Vertex) (*Face, *Edge) {
	if v1.He == nil || v2.He == nil {
		return nil, nil
	}
	return Lmef(v1.He, v2.He)
}

// Smev is a convenience: Split vertex, Make Edge Vertex by index.
func Smev(s *Solid, vertIdx uint64, loc vec.SFVec3f) (*Vertex, *Edge) {
	v := s.FindVertex(vertIdx)
	if v == nil {
		return nil, nil
	}
	return Mev(s, v, loc)
}

// Smef is a convenience: Make Edge Face by vertex indices.
func Smef(s *Solid, idx1, idx2 uint64) (*Face, *Edge) {
	v1 := s.FindVertex(idx1)
	v2 := s.FindVertex(idx2)
	if v1 == nil || v2 == nil {
		return nil, nil
	}
	return Mef(s, v1, v2)
}

// ---------------------------------------------------------------------------
// BuildFromIndexSet - builds a solid from VRML-style indexed data
// ---------------------------------------------------------------------------

// BuildFromIndexSet constructs a solid from vertex positions and face indices.
// indices uses -1 as a face separator (VRML convention).
func BuildFromIndexSet(positions []vec.SFVec3f, indices []int64, color vec.SFColor) *Solid {
	if len(positions) == 0 || len(indices) == 0 {
		return nil
	}

	// Parse face index lists.
	var faces [][]int64
	var cur []int64
	for _, idx := range indices {
		if idx == -1 {
			if len(cur) >= 3 {
				faces = append(faces, append([]int64{}, cur...))
			}
			cur = cur[:0]
		} else {
			cur = append(cur, idx)
		}
	}
	if len(cur) >= 3 {
		faces = append(faces, append([]int64{}, cur...))
	}
	if len(faces) == 0 {
		return nil
	}

	first := faces[0]
	s, v0, _ := Mvfs(positions[first[0]], color)
	verts := make([]*Vertex, len(positions))
	verts[first[0]] = v0

	// Build the first face boundary: create vertices sequentially so
	// the loop contains exactly one half-edge per vertex.
	prev := v0
	for i := 1; i < len(first); i++ {
		nv, _ := Lmev(prev.He, positions[first[i]])
		verts[first[i]] = nv
		prev = nv
	}

	// Close the first face polygon.
	if len(first) >= 3 {
		Lmef(v0.He, prev.He)
	}

	// Create any remaining vertices not part of the first face.
	for i, pos := range positions {
		if verts[i] == nil {
			nv, _ := Lmev(v0.He, pos)
			verts[i] = nv
		}
	}

	// Build remaining faces.
	for _, face := range faces[1:] {
		buildFace(s, verts, face)
	}

	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func buildFace(s *Solid, verts []*Vertex, indices []int64) {
	_ = s
	if len(indices) < 3 {
		return
	}
	v0 := verts[indices[0]]
	v1 := verts[indices[1]]
	if v0.He == nil || v1.He == nil {
		return
	}
	Lmef(v0.He, v1.He)
}

// MarkCreases marks edges as creases based on the crease angle.
func MarkCreases(s *Solid, creaseAngle float64) {
	for e := s.Edges; e != nil; e = e.Next {
		f1 := e.He1.GetFace()
		f2 := e.He2.GetFace()
		if f1 == nil || f2 == nil {
			continue
		}
		dot := f1.Normal.Dot(f2.Normal)
		if dot < float64(1.0)-creaseAngle {
			e.Mark |= CREASE
		}
	}
}
