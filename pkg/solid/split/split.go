package split

import (
	"sort"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
)

type splitNeighborhood struct {
	sector *base.HalfEdge
	cl     int
}

type splitClassifyRecord struct {
	sp    base.Plane
	v     *base.Vertex
	nbrs  []splitNeighborhood
	spRec *SplitRecord
}

func (cr *splitClassifyRecord) reset(spR *SplitRecord) {
	cr.spRec = spR
	cr.sp = spR.sp
	cr.v = nil
	cr.nbrs = cr.nbrs[:0]
}

func (cr *splitClassifyRecord) setVertex(i int) {
	cr.v = cr.spRec.verts[i]
	cr.nbrs = cr.nbrs[:0]
}

func (cr *splitClassifyRecord) addNeighborhood(he *base.HalfEdge, cl int) {
	cr.nbrs = append(cr.nbrs, splitNeighborhood{sector: he, cl: cl})
}

func (cr *splitClassifyRecord) getNeighborhood() {
	cr.nbrs = cr.nbrs[:0]
	he := cr.v.He
	for {
		vNext := he.Next.Vertex
		cl := base.FloatCompare(cr.sp.GetDistance(vNext.Loc))
		cr.addNeighborhood(he, cl)
		he = he.GetMate().Next
		if he == cr.v.He {
			break
		}
	}
}

func (cr *splitClassifyRecord) reclassifyOnSectors() {
	n := len(cr.nbrs)
	epsSq := float64(1e-5 * 1e-5)
	spNormal := cr.sp.Normal

	for i := 0; i < n; i++ {
		facNormal := cr.nbrs[i].sector.GetFaceNormal()
		c := facNormal.Cross(spNormal)
		d := c.Dot(c)

		if d < epsSq {
			prv := (i + n - 1) % n
			dotN := facNormal.Dot(spNormal)
			if base.FloatCompare(dotN) == 1 {
				cr.nbrs[prv].cl = base.BELOW
				cr.nbrs[i].cl = base.BELOW
			} else {
				cr.nbrs[prv].cl = base.ABOVE
				cr.nbrs[i].cl = base.ABOVE
			}
		}
	}
}

func (cr *splitClassifyRecord) reclassifyOnEdges() {
	n := len(cr.nbrs)
	for i := 0; i < n; i++ {
		if cr.nbrs[i].cl == base.ON {
			prv := cr.nbrs[(n+i-1)%n].cl
			nxt := cr.nbrs[(i+1)%n].cl
			if prv == nxt {
				cr.nbrs[i].cl = prv
			} else {
				cr.nbrs[i].cl = base.ABOVE
			}
		}
	}
}

func (cr *splitClassifyRecord) insertNullEdges() {
	n := len(cr.nbrs)
	if n == 0 {
		return
	}

	cur := 0
	for cur < n {
		nxt := (cur + 1) % n
		if cr.nbrs[cur].cl == base.BELOW && cr.nbrs[nxt].cl == base.ABOVE {
			break
		}
		cur++
	}
	if cur == n {
		return
	}

	cur = (cur + 1) % n
	start := cur
	head := cr.nbrs[start].sector

	for {
		for {
			nxt := (cur + 1) % n
			if cr.nbrs[cur].cl == base.ABOVE && cr.nbrs[nxt].cl == base.BELOW {
				break
			}
			cur = (cur + 1) % n
		}

		tail := cr.nbrs[(cur+1)%n].sector

		_, e, _ := euler.Lmev2(head, tail, head.Vertex.Loc)
		cr.spRec.addEdge(e)

		for {
			nxt := (cur + 1) % n
			if cr.nbrs[cur].cl == base.BELOW && cr.nbrs[nxt].cl == base.ABOVE {
				break
			}
			cur = (cur + 1) % n
		}

		if (cur+1)%n == start {
			return
		}

		cur = (cur + 1) % n
		head = cr.nbrs[cur].sector
	}
}

type SplitRecord struct {
	s  *base.Solid
	sp base.Plane
	A  **base.Solid
	B  **base.Solid

	verts     []*base.Vertex
	edges     []*base.Edge
	faces     []*base.Face
	looseEnds []*base.HalfEdge
}

func NewSplitRecord() *SplitRecord {
	return &SplitRecord{}
}

func (sr *SplitRecord) reset(sp base.Plane, s *base.Solid, a, b **base.Solid) {
	sr.sp = sp
	sr.s = s
	sr.A = a
	sr.B = b
	*a = nil
	*b = nil
	sr.verts = sr.verts[:0]
	sr.edges = sr.edges[:0]
	sr.faces = sr.faces[:0]
	sr.looseEnds = sr.looseEnds[:0]
}

func (sr *SplitRecord) addVert(v *base.Vertex) {
	if v.IsMarked(base.ON) {
		return
	}
	v.Mark = base.ON
	sr.verts = append(sr.verts, v)
}

func (sr *SplitRecord) addEdge(e *base.Edge) {
	sr.edges = append(sr.edges, e)
}

func (sr *SplitRecord) addFace(f *base.Face) {
	sr.faces = append(sr.faces, f)
}

func (sr *SplitRecord) addLooseEnd(he *base.HalfEdge) {
	sr.looseEnds = append(sr.looseEnds, he)
}

func (sr *SplitRecord) generate() {
	sr.verts = sr.verts[:0]

	for v := sr.s.Verts; v != nil; v = v.Next {
		v.Scratch = sr.sp.GetDistance(v.Loc)
	}

	var edgeList []*base.Edge
	for e := sr.s.Edges; e != nil; e = e.Next {
		edgeList = append(edgeList, e)
	}

	for _, e := range edgeList {
		e.Mark = base.UNKNOWN

		if e.He1 == nil || e.He2 == nil {
			continue
		}

		v1 := e.He1.Vertex
		v2 := e.He2.Vertex
		d1 := v1.Scratch
		d2 := v2.Scratch
		s1 := base.FloatCompare(d1)
		s2 := base.FloatCompare(d2)

		if (s1 == -1 && s2 == 1) || (s1 == 1 && s2 == -1) {
			t := d1 / (d1 - d2)
			diff := v2.Loc.Sub(v1.Loc)
			intersectionPt := v1.Loc.Add(diff.Scale(t))

			he := e.He2.Next
			newV, _, _ := euler.Lmev2(e.He1, he, intersectionPt)
			newV.Scratch = 0
			newV.Mark = base.UNKNOWN
			sr.addVert(newV)
		} else {
			if s1 == 0 {
				sr.addVert(v1)
			}
			if s2 == 0 {
				sr.addVert(v2)
			}
		}
	}
}

func (sr *SplitRecord) classify() bool {
	sr.edges = sr.edges[:0]

	cr := &splitClassifyRecord{}
	cr.reset(sr)

	for i := 0; i < len(sr.verts); i++ {
		cr.setVertex(i)
		cr.getNeighborhood()
		cr.reclassifyOnSectors()
		cr.reclassifyOnEdges()
		cr.insertNullEdges()
	}

	sr.verts = sr.verts[:0]
	return len(sr.edges) != 0
}

func (sr *SplitRecord) connect() {
	sr.faces = sr.faces[:0]

	sort.Slice(sr.edges, func(i, j int) bool {
		v1 := sr.edges[i].He1.Vertex.Loc
		v2 := sr.edges[j].He1.Vertex.Loc
		if c := base.FloatCompare(v1.X - v2.X); c != 0 {
			return c < 0
		}
		if c := base.FloatCompare(v1.Y - v2.Y); c != 0 {
			return c < 0
		}
		return base.FloatCompare(v1.Z-v2.Z) < 0
	})

	for idx := 0; idx < len(sr.edges); idx++ {
		nextEdge := sr.edges[idx]

		he1 := sr.canJoin(nextEdge.He1)
		he2 := sr.canJoin(nextEdge.He2)

		if he1 != nil {
			algorithms.Join(he1, nextEdge.He1, false)
			if he1.GetMate().Mark != base.LOOSE {
				if ff := algorithms.Cut(he1, false); ff != nil {
					sr.addFace(ff)
				}
			}
		}

		if he2 != nil {
			algorithms.Join(he2, nextEdge.He2, false)
			if he2.GetMate().Mark != base.LOOSE {
				if ff := algorithms.Cut(he2, false); ff != nil {
					sr.addFace(ff)
				}
			}
		}

		if he1 != nil && he2 != nil {
			if ff := algorithms.Cut(nextEdge.He1, false); ff != nil {
				sr.addFace(ff)
			}
		}
	}

	sr.edges = sr.edges[:0]
}

func (sr *SplitRecord) canJoin(he *base.HalfEdge) *base.HalfEdge {
	he.Mark = base.NOT_LOOSE

	for i, loose := range sr.looseEnds {
		if he.IsNeighbor(loose) {
			loose.Mark = base.NOT_LOOSE
			sr.looseEnds = append(sr.looseEnds[:i], sr.looseEnds[i+1:]...)
			return loose
		}
	}

	he.Mark = base.LOOSE
	sr.addLooseEnd(he)
	return nil
}

func (sr *SplitRecord) finish() {
	nf := len(sr.faces)
	for i := 0; i < nf; i++ {
		f := sr.faces[i]
		if f.GetFirstLoop() == f.LoopOut {
			f.LoopOut = f.GetSecondLoop()
		}
		companion, _ := euler.Lmfkrh(f.GetFirstLoop())
		sr.faces = append(sr.faces, companion)
	}

	above := base.NewSolid()
	below := base.NewSolid()
	*sr.A = above
	*sr.B = below

	sr.s.SetFaceMarks2(0)

	sr.finishClassify(nf)

	above.Cleanup()
	below.Cleanup()

	sr.s.Cleanup()
}

func (sr *SplitRecord) finishClassify(nf int) {
	for i := 0; i < nf; i++ {
		algorithms.MoveFace(sr.faces[i], *sr.A)
		algorithms.MoveFace(sr.faces[nf+i], *sr.B)
	}

	for sr.s.GetFirstFace() != nil {
		f := sr.s.GetFirstFace()
		sr.s.SetFaceMarks2(base.UNKNOWN)

		side := 0
		for _, l := range f.Loops {
			done := false
			l.ForEachHe(func(he *base.HalfEdge) bool {
				d := sr.sp.GetDistance(he.Vertex.Loc)
				s := base.FloatCompare(d)
				if s != 0 {
					side = s
					done = true
					return false
				}
				return true
			})
			if done {
				break
			}
		}

		if side == base.BELOW {
			algorithms.MoveFace(f, *sr.A)
		} else {
			algorithms.MoveFace(f, *sr.B)
		}
	}
}

func (sr *SplitRecord) complete() {
	(*sr.A).Renumber()
	(*sr.B).Renumber()
	algorithms.RemoveCoplaneColine(*sr.A)
	algorithms.RemoveCoplaneColine(*sr.B)
}

func Split(s *base.Solid, sp base.Plane) (above, below *base.Solid, ok bool) {
	s.CalcPlaneEquations()

	rec := NewSplitRecord()
	rec.reset(sp, s, &above, &below)

	s.SetFaceMarks(base.UNKNOWN)
	s.SetVertexMarks(base.UNKNOWN)

	rec.generate()

	if !rec.classify() {
		return nil, nil, false
	}

	rec.connect()
	rec.finish()
	rec.complete()

	return above, below, true
}
