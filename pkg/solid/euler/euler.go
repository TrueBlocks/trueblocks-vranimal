package euler

import (
	"errors"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

var (
	ErrNilHalfEdge    = errors.New("euler: nil half-edge")
	ErrNilEdge        = errors.New("euler: nil edge")
	ErrNilVertex      = errors.New("euler: nil vertex")
	ErrNilFace        = errors.New("euler: nil face")
	ErrNilLoop        = errors.New("euler: nil loop")
	ErrDifferentLoops = errors.New("euler: half-edges are in different loops")
	ErrDifferentVerts = errors.New("euler: half-edges reference different vertices")
)

// insertBefore inserts nhe immediately before he in the circular half-edge ring.
func insertBefore(he, nhe *base.HalfEdge) {
	nhe.Next = he
	nhe.Prev = he.Prev
	he.Prev.Next = nhe
	he.Prev = nhe
}

// Mvfs creates a new solid with one vertex and one face (Make Vertex Face Solid).
func Mvfs(loc vec.SFVec3f, color vec.SFColor) (*base.Solid, *base.Vertex, *base.Face) {
	s := base.NewSolid()
	v := base.NewVertexVec(loc)
	s.AddVertex(v)
	f := base.NewFace(s, color)
	s.AddFace(f)
	l := base.NewLoop(f, true)
	he := base.NewHalfEdge(l, v)
	l.AddHalfEdge(he)
	v.He = he
	return s, v, f
}

// Kvfs destroys a solid with one vertex and one face (Kill Vertex Face Solid).
func Kvfs(s *base.Solid) {
	s.Faces = nil
	s.Edges = nil
	s.Verts = nil
	s.NFace = 0
	s.NEdge = 0
	s.NVert = 0
}

// Lmev creates a new vertex and edge at a half-edge (Low Make Edge Vertex).
func Lmev(he *base.HalfEdge, loc vec.SFVec3f) (*base.Vertex, *base.Edge, error) {
	if he == nil {
		return nil, nil, ErrNilHalfEdge
	}
	if he.GetFace() == nil {
		return nil, nil, ErrNilFace
	}
	s := he.GetFace().Solid
	nv := base.NewVertexVec(loc)
	s.AddVertex(nv)

	ne := base.NewEdge()
	s.AddEdge(ne)

	if he.Edge == nil {
		ne.He2 = he
		he.Edge = ne

		nhe := &base.HalfEdge{Vertex: nv, Edge: ne, Loop: he.Loop}
		ne.He1 = nhe
		insertBefore(he, nhe)

		nv.He = nhe
		he.Vertex.He = he
	} else {
		nhe2 := &base.HalfEdge{Vertex: he.Vertex, Edge: ne, Loop: he.Loop}
		ne.He2 = nhe2
		insertBefore(he, nhe2)

		nhe1 := &base.HalfEdge{Vertex: nv, Edge: ne, Loop: he.Loop}
		ne.He1 = nhe1
		insertBefore(he, nhe1)

		nv.He = nhe1
		he.Vertex.He = he
	}

	return nv, ne, nil
}

// Lmev2 splits a vertex between two half-edges in potentially different loops.
func Lmev2(he1, he2 *base.HalfEdge, loc vec.SFVec3f) (*base.Vertex, *base.Edge, error) {
	if he1 == nil || he2 == nil {
		return nil, nil, ErrNilHalfEdge
	}
	if he1.Vertex != he2.Vertex {
		return nil, nil, ErrDifferentVerts
	}
	s := he1.GetFace().Solid
	nv := base.NewVertexVec(loc)
	s.AddVertex(nv)

	ne := base.NewEdge()
	s.AddEdge(ne)

	he := he1
	for he != he2 {
		he.Vertex = nv
		m := he.GetMate()
		if m == nil {
			break
		}
		he = m.Next
		if he == he1 {
			break
		}
	}

	nhe2 := &base.HalfEdge{Vertex: he2.Vertex, Edge: ne, Loop: he1.Loop}
	ne.He2 = nhe2
	insertBefore(he1, nhe2)

	nhe1 := &base.HalfEdge{Vertex: nv, Edge: ne, Loop: he2.Loop}
	ne.He1 = nhe1
	insertBefore(he2, nhe1)

	nv.He = nhe1
	he2.Vertex.He = he2

	return nv, ne, nil
}

// Lmef creates a new edge and face by splitting a loop (Low Make Edge Face).
// he1 and he2 must be in the same loop.
func Lmef(he1, he2 *base.HalfEdge) (*base.Face, *base.Edge, error) {
	if he1 == nil || he2 == nil {
		return nil, nil, ErrNilHalfEdge
	}
	if he1.Loop != he2.Loop {
		return nil, nil, ErrDifferentLoops
	}
	old := he1.GetFace()
	s := old.Solid

	nf := base.NewFace(s, old.GetColor(vec.White))
	s.AddFace(nf)
	ne := base.NewEdge()
	s.AddEdge(ne)
	nl := base.NewLoop(nf, true)

	for he := he1; he != he2; he = he.Next {
		he.Loop = nl
	}

	nhe1 := &base.HalfEdge{Vertex: he2.Vertex, Edge: ne, Loop: he1.Loop}
	ne.He2 = nhe1
	nhe1.Edge = ne
	insertBefore(he1, nhe1)

	nhe2 := &base.HalfEdge{Vertex: he1.Vertex, Edge: ne, Loop: he2.Loop}
	ne.He1 = nhe2
	nhe2.Edge = ne
	insertBefore(he2, nhe2)

	nhe1.Prev.Next = nhe2
	nhe2.Prev.Next = nhe1

	temp := nhe2.Prev
	nhe2.Prev = nhe1.Prev
	nhe1.Prev = temp

	he2.Loop.SetFirstHe(nhe2)
	nl.SetFirstHe(nhe1)

	return nf, ne, nil
}

// Lkev kills an edge and vertex (Low Kill Edge Vertex).
func Lkev(he *base.HalfEdge) error {
	if he == nil {
		return ErrNilHalfEdge
	}
	if he.Edge == nil {
		return ErrNilEdge
	}
	s := he.GetFace().Solid
	e := he.Edge
	mate := he.GetMate()

	killV := he.Vertex
	keepV := mate.Vertex

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
		if cur == mate.Next {
			break
		}
	}

	if he.Prev == mate {
		keepV.He = mate.Next.Next
	} else {
		keepV.He = mate.Next
	}
	if keepV.He != nil && keepV.He.Edge == nil {
		keepV.He = nil
	}

	he.Prev.Next = he.Next
	he.Next.Prev = he.Prev
	if he.Loop.HalfEdges == he {
		he.Loop.HalfEdges = he.Next
	}

	if mate != nil && mate != he {
		mate.Prev.Next = mate.Next
		mate.Next.Prev = mate.Prev
		if mate.Loop.HalfEdges == mate {
			mate.Loop.HalfEdges = mate.Next
		}
	}

	s.RemoveEdge(e)
	s.RemoveVertex(killV)
	return nil
}

// Lkef kills an edge and face (Low Kill Edge Face).
func Lkef(he *base.HalfEdge) error {
	if he == nil {
		return ErrNilHalfEdge
	}
	if he.Edge == nil {
		return ErrNilEdge
	}
	s := he.GetFace().Solid
	e := he.Edge
	mate := he.GetMate()
	keepF := he.GetFace()
	killF := mate.GetFace()

	if killF != keepF {
		for len(killF.Loops) > 0 {
			l := killF.Loops[0]
			killF.RemoveLoop(l)
			keepF.AddLoop(l, false)
		}
	}

	keepLoop := he.Loop
	killLoop := mate.Loop
	cur := mate
	start := mate
	for {
		cur.Loop = keepLoop
		cur = cur.Next
		if cur == start {
			break
		}
	}

	he.Prev.Next = mate.Next
	mate.Next.Prev = he.Prev
	mate.Prev.Next = he.Next
	he.Next.Prev = mate.Prev

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

	if keepF.LoopOut == killLoop {
		keepF.LoopOut = keepLoop
	}
	killLoop.HalfEdges = nil
	keepF.RemoveLoop(killLoop)

	s.RemoveEdge(e)
	if killF != keepF {
		s.RemoveFace(killF)
	}
	return nil
}

// Lkemr kills an edge, creating a new inner ring (Kill Edge Make Ring).
func Lkemr(he *base.HalfEdge) error {
	if he == nil {
		return ErrNilHalfEdge
	}
	if he.Edge == nil {
		return ErrNilEdge
	}
	s := he.GetFace().Solid
	e := he.Edge
	mate := he.GetMate()
	f := he.GetFace()

	nl := base.NewLoop(f, false)

	he.Prev.Next = mate.Next
	mate.Next.Prev = he.Prev

	mate.Prev.Next = he.Next
	he.Next.Prev = mate.Prev

	he.Loop.SetFirstHe(he.Prev)

	nl.HalfEdges = mate.Prev
	cur := mate.Prev
	start := cur
	for {
		cur.Loop = nl
		cur = cur.Next
		if cur == start {
			break
		}
	}

	s.RemoveEdge(e)
	return nil
}

// Lmekr makes an edge, killing an inner ring (Make Edge Kill Ring).
func Lmekr(he1, he2 *base.HalfEdge) (*base.Edge, error) {
	if he1 == nil || he2 == nil {
		return nil, ErrNilHalfEdge
	}
	s := he1.GetFace().Solid
	ne := base.NewEdge()
	s.AddEdge(ne)

	nhe1 := base.NewHalfEdge(he1.Loop, he2.Vertex)
	nhe2 := base.NewHalfEdge(he1.Loop, he1.Vertex)
	ne.He1 = nhe1
	ne.He2 = nhe2
	nhe1.Edge = ne
	nhe2.Edge = ne

	he1prev := he1.Prev
	he2prev := he2.Prev

	nhe1.Prev = he2prev
	nhe1.Next = he1
	he2prev.Next = nhe1
	he1.Prev = nhe1

	nhe2.Prev = he1prev
	nhe2.Next = he2
	he1prev.Next = nhe2
	he2.Prev = nhe2

	killLoop := he2.Loop
	cur := nhe1
	start := nhe1
	for {
		cur.Loop = he1.Loop
		cur = cur.Next
		if cur == start {
			break
		}
	}

	f := he1.GetFace()
	if killLoop != he1.Loop {
		f.RemoveLoop(killLoop)
	}

	he1.Loop.SetFirstHe(nhe1)
	return ne, nil
}

// Lmfkrh promotes an inner ring to a new face (Make Face Kill Ring Hole).
func Lmfkrh(innerLoop *base.Loop) (*base.Face, error) {
	if innerLoop == nil {
		return nil, ErrNilLoop
	}
	f := innerLoop.Face
	if f == nil {
		return nil, ErrNilFace
	}
	s := f.Solid
	nf := base.NewFace(s, f.GetColor(vec.White))
	s.AddFace(nf)

	f.RemoveLoop(innerLoop)
	nf.AddLoop(innerLoop, true)

	innerLoop.ForEachHe(func(he *base.HalfEdge) bool {
		he.Loop = innerLoop
		return true
	})

	return nf, nil
}

// Lkfmrh kills a face, creating an inner ring in another face (Kill Face Make Ring Hole).
func Lkfmrh(killFace, keepFace *base.Face) error {
	if killFace == nil || keepFace == nil {
		return ErrNilFace
	}
	if killFace.LoopOut == nil {
		return ErrNilLoop
	}
	s := killFace.Solid
	l := killFace.LoopOut
	killFace.RemoveLoop(l)
	keepFace.AddLoop(l, false)

	l.ForEachHe(func(he *base.HalfEdge) bool {
		he.Loop = l
		return true
	})

	s.RemoveFace(killFace)
	return nil
}

// Lringmv moves a loop from one face to another.
func Lringmv(s *base.Solid, l *base.Loop, toFace *base.Face, isOuter bool) error {
	if l == nil {
		return ErrNilLoop
	}
	if toFace == nil {
		return ErrNilFace
	}
	fromFace := l.Face
	fromFace.RemoveLoop(l)
	toFace.AddLoop(l, isOuter)
	return nil
}

// BuildFromIndexSet constructs a solid from vertex positions and face indices.
// indices uses -1 as a face separator (VRML convention).
func BuildFromIndexSet(positions []vec.SFVec3f, indices []int64, color vec.SFColor) (*base.Solid, error) {
	if len(positions) == 0 || len(indices) == 0 {
		return nil, errors.New("euler: empty positions or indices")
	}

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
		return nil, errors.New("euler: no valid faces in index set")
	}

	first := faces[0]
	s, v0, _ := Mvfs(positions[first[0]], color)
	verts := make([]*base.Vertex, len(positions))
	verts[first[0]] = v0

	prev := v0
	for i := 1; i < len(first); i++ {
		nv, _, err := Lmev(prev.He, positions[first[i]])
		if err != nil {
			return nil, err
		}
		verts[first[i]] = nv
		prev = nv
	}

	if len(first) >= 3 {
		if _, _, err := Lmef(v0.He, prev.He); err != nil {
			return nil, err
		}
	}

	for i, pos := range positions {
		if verts[i] == nil {
			nv, _, err := Lmev(v0.He, pos)
			if err != nil {
				return nil, err
			}
			verts[i] = nv
		}
	}

	for _, face := range faces[1:] {
		if err := buildFace(s, verts, face); err != nil {
			return nil, err
		}
	}

	s.CalcPlaneEquations()
	s.Renumber()
	return s, nil
}

func buildFace(s *base.Solid, verts []*base.Vertex, indices []int64) error {
	_ = s
	if len(indices) < 3 {
		return nil
	}
	v0 := verts[indices[0]]
	v1 := verts[indices[1]]
	if v0 == nil || v1 == nil {
		return ErrNilVertex
	}
	if v0.He == nil || v1.He == nil {
		return ErrNilHalfEdge
	}
	_, _, err := Lmef(v0.He, v1.He)
	return err
}

// MarkCreases marks edges as creases based on the crease angle.
func MarkCreases(s *base.Solid, creaseAngle float64) {
	for e := s.Edges; e != nil; e = e.Next {
		f1 := e.He1.GetFace()
		f2 := e.He2.GetFace()
		if f1 == nil || f2 == nil {
			continue
		}
		dot := f1.Normal.Dot(f2.Normal)
		if dot < 1.0-creaseAngle {
			e.Mark |= base.CREASE
		}
	}
}
