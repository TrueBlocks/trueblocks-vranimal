package boolop

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

const (
	IN  = -1
	OUT = 1
)

const BAD_CROSSING = 12

const (
	SAME_FACE = 1
	OPP_FACE  = 2
)

type vfNeighborhood struct {
	sector     *base.HalfEdge
	cl         int
	intersects bool
}

type vfClassifyRecord struct {
	sp   base.Plane
	v    *base.Vertex
	f    *base.Face
	br   *BoolopRecord
	isB  bool
	op   int
	nbrs []vfNeighborhood
}

func (cr *vfClassifyRecord) reset(br *BoolopRecord) {
	cr.br = br
	cr.op = br.Op
	cr.v = nil
	cr.f = nil
	cr.nbrs = cr.nbrs[:0]
}

func (cr *vfClassifyRecord) setVertexFace(i int, isB bool) {
	cr.isB = isB
	if isB {
		cr.v = cr.br.VertsB[i].V
		cr.f = cr.br.VertsB[i].F
	} else {
		cr.v = cr.br.VertsA[i].V
		cr.f = cr.br.VertsA[i].F
	}
	cr.sp = base.Plane{Normal: cr.f.Normal, D: cr.f.D}
	cr.nbrs = cr.nbrs[:0]
}

func (cr *vfClassifyRecord) vfGetNeighborhood() {
	cr.nbrs = cr.nbrs[:0]
	he := cr.v.He
	for {
		vNext := he.Next.Vertex
		cl := base.FloatCompare(cr.sp.GetDistance(vNext.Loc))
		if he.IsWide(false) {
			pCl := cl
			bisector := he.Bisect()
			cl = base.FloatCompare(cr.sp.GetDistance(bisector))
			if pCl != 0 {
				cr.nbrs = append(cr.nbrs, vfNeighborhood{sector: he, cl: pCl, intersects: true})
			}
			if cl != 0 {
				cr.nbrs = append(cr.nbrs, vfNeighborhood{sector: he, cl: cl, intersects: true})
			}
		} else {
			cr.nbrs = append(cr.nbrs, vfNeighborhood{sector: he, cl: cl, intersects: true})
		}
		he = he.GetMate().Next
		if he == cr.v.He {
			break
		}
	}
}

func (cr *vfClassifyRecord) vfReclassifyOnSectors() {
	n := len(cr.nbrs)
	epsSq := float64(1e-5 * 1e-5)
	spNormal := cr.sp.Normal
	for i := 0; i < n; i++ {
		facNormal := cr.nbrs[i].sector.GetFaceNormal()
		c := facNormal.Cross(spNormal)
		d := c.Dot(c)
		if base.FloatCompare(d-epsSq) <= 0 {
			prv := (i + n - 1) % n
			dotN := facNormal.Dot(spNormal)
			var newCl int
			if base.FloatCompare(dotN) == 1 {
				if cr.isB {
					newCl = boolIf(cr.op == base.BoolUnion, IN, OUT)
				} else {
					newCl = boolIf(cr.op == base.BoolUnion, OUT, IN)
				}
			} else {
				if cr.isB {
					newCl = boolIf(cr.op == base.BoolUnion, IN, OUT)
				} else {
					newCl = boolIf(cr.op == base.BoolUnion, IN, OUT)
				}
			}
			cr.nbrs[i].cl = newCl
			cr.nbrs[prv].cl = newCl
		}
	}
}

func (cr *vfClassifyRecord) vfReclassifyOnEdges() {
	n := len(cr.nbrs)
	for i := 0; i < n; i++ {
		if cr.nbrs[i].cl == base.ON {
			prv := cr.nbrs[(n+i-1)%n].cl
			nxt := cr.nbrs[(i+1)%n].cl
			if prv == nxt {
				cr.nbrs[i].cl = prv
			} else {
				cr.nbrs[i].cl = IN
			}
		}
	}
}

func (cr *vfClassifyRecord) vfShiftDown() {
	n := len(cr.nbrs)
	if n == 0 {
		return
	}
	saved := cr.nbrs[0]
	for i := 0; i < n-1; i++ {
		cr.nbrs[i] = cr.nbrs[i+1]
	}
	cr.nbrs[n-1] = saved
}

func (cr *vfClassifyRecord) vfPrepareNextCrossing() bool {
	n := len(cr.nbrs)
	for k := 0; k < n; k++ {
		prv := (k + n - 1) % n
		nxt := (k + 1) % n
		clPrev := cr.nbrs[prv].cl
		cl := cr.nbrs[k].cl
		clNext := cr.nbrs[nxt].cl
		cr.nbrs[k].intersects = true
		if clPrev == cl && cl == clNext {
			cr.nbrs[k].intersects = false
		}
	}
	j := 0
	for k := 0; k < n; k++ {
		if cr.nbrs[k].intersects {
			cr.nbrs[j] = cr.nbrs[k]
			j++
		}
	}
	cr.nbrs = cr.nbrs[:j]
	n = j
	if n > 1 {
		for attempts := 0; attempts < n; attempts++ {
			if cr.nbrs[0].cl == IN && cr.nbrs[1].cl == OUT {
				break
			}
			cr.vfShiftDown()
		}
	}
	return n > 0
}

func isOnOuterLoop(v *base.Vertex) bool {
	f := v.He.GetFace()
	if f == nil || f.LoopOut == nil {
		return true
	}
	found := false
	f.LoopOut.ForEachHe(func(he *base.HalfEdge) bool {
		if he.Vertex == v {
			found = true
			return false
		}
		return true
	})
	return found
}

func (cr *vfClassifyRecord) vfFindNextCrossing() (hInOut, hOutIn *base.HalfEdge) {
	n := len(cr.nbrs)
	if n < 2 {
		return nil, nil
	}
	nxt := (1 + 1) % n
	if n > 2 && cr.nbrs[nxt].cl == IN {
		hInOut = cr.nbrs[0].sector.GetMate().Next
		hOutIn = cr.nbrs[1].sector.GetMate().Next
		cr.nbrs[1].cl = IN
	} else if n > 2 {
		hInOut = cr.nbrs[0].sector.GetMate().Next
		hOutIn = cr.nbrs[nxt].sector.GetMate().Next
		if !isOnOuterLoop(cr.v) {
			hInOut, hOutIn = hOutIn, hInOut
		}
		cr.nbrs[1].cl = IN
		if nxt < n {
			cr.nbrs[nxt].cl = IN
		}
		cr.vfShiftDown()
	} else {
		hInOut = cr.nbrs[0].sector.GetMate().Next
		hOutIn = cr.nbrs[1].sector.GetMate().Next
		cr.nbrs[1].cl = IN
	}
	cr.vfShiftDown()
	cr.vfShiftDown()
	return hInOut, hOutIn
}

func (cr *vfClassifyRecord) vfSeperate(hInOut, hOutIn *base.HalfEdge) {
	outer := isOnOuterLoop(cr.v)
	if cr.br.NoVV {
		_, _, _ = euler.Lmev2(hInOut, hOutIn, hInOut.Vertex.Loc)
	} else {
		_, _, _ = euler.Lmev2(hOutIn, hInOut, hInOut.Vertex.Loc)
	}
	e := hInOut.Prev.Edge
	if !outer {
		e.SwapHes()
	}
	if cr.isB {
		cr.br.EdgesB = append(cr.br.EdgesB, e)
	} else {
		cr.br.EdgesA = append(cr.br.EdgesA, e)
	}
	e2 := makeRing(cr.f, hInOut.Vertex.Loc)
	if !outer {
		e2.SwapHes()
	}
	if !cr.isB {
		cr.br.EdgesB = append(cr.br.EdgesB, e2)
	} else {
		cr.br.EdgesA = append(cr.br.EdgesA, e2)
	}
}

func (cr *vfClassifyRecord) vfInsertNullEdges() {
	for cr.vfPrepareNextCrossing() {
		hInOut, hOutIn := cr.vfFindNextCrossing()
		if hInOut != nil && hOutIn != nil {
			cr.vfSeperate(hInOut, hOutIn)
		}
	}
}

func (cr *vfClassifyRecord) vfClassify(i int, isB bool) {
	cr.setVertexFace(i, isB)
	cr.vfGetNeighborhood()
	cr.vfReclassifyOnSectors()
	cr.vfReclassifyOnEdges()
	cr.vfInsertNullEdges()
}

type vvNeighborhood struct {
	he         *base.HalfEdge
	vNext      vec.SFVec3f
	vPrev      vec.SFVec3f
	vCross     vec.SFVec3f
	fromBisect int
}

type vvSector struct {
	indexA     int
	clPrevA    int
	clNextA    int
	indexB     int
	clPrevB    int
	clNextB    int
	intersects int
	onFace     int
}

type vvClassifyRecord struct {
	v       *base.Vertex
	v2      *base.Vertex
	br      *BoolopRecord
	op      int
	nhA     []vvNeighborhood
	nhB     []vvNeighborhood
	sectors []vvSector
}

func (cr *vvClassifyRecord) reset(br *BoolopRecord) {
	cr.br = br
	cr.op = br.Op
	cr.v = nil
	cr.v2 = nil
	cr.nhA = cr.nhA[:0]
	cr.nhB = cr.nhB[:0]
	cr.sectors = cr.sectors[:0]
}

func (cr *vvClassifyRecord) setVertexVertex(i int) {
	cr.v = cr.br.VertsV[i].Va
	cr.v2 = cr.br.VertsV[i].Vb
}

func (cr *vvClassifyRecord) preprocess(v *base.Vertex, isB bool) {
	he := v.He
	for {
		vPrev := he.Prev.Vertex.Loc.Sub(he.Vertex.Loc)
		vNext := he.Next.Vertex.Loc.Sub(he.Vertex.Loc)
		cr.addNeighbor(he, vPrev, vNext, isB, 0)
		he = he.GetMate().Next
		if he == v.He {
			break
		}
	}
}

func (cr *vvClassifyRecord) addNeighbor(he *base.HalfEdge, vP, vN vec.SFVec3f, isB bool, fromBisect int) {
	vC := vP.Cross(vN).Negate()
	nb := vvNeighborhood{
		he:         he,
		vNext:      vN,
		vPrev:      vP,
		vCross:     vC,
		fromBisect: fromBisect,
	}
	if isB {
		cr.nhB = append(cr.nhB, nb)
	} else {
		cr.nhA = append(cr.nhA, nb)
	}
}

func (cr *vvClassifyRecord) getNeighborhood() {
	for i := 0; i < len(cr.nhA); i++ {
		for j := 0; j < len(cr.nhB); j++ {
			cr.addSector(i, j)
		}
	}
}

func (cr *vvClassifyRecord) addSector(i, j int) {
	ha := cr.nhA[i].he
	hb := cr.nhB[j].he
	normalA := ha.GetFaceNormal()
	normalB := hb.GetFaceNormal()
	d1 := normalB.Dot(cr.nhA[i].vNext)
	d2 := normalB.Dot(cr.nhA[i].vPrev)
	d3 := normalA.Dot(cr.nhB[j].vNext)
	d4 := normalA.Dot(cr.nhB[j].vPrev)
	s := vvSector{
		indexA:     i,
		indexB:     j,
		clNextA:    base.FloatCompare(d1),
		clPrevA:    base.FloatCompare(d2),
		clNextB:    base.FloatCompare(d3),
		clPrevB:    base.FloatCompare(d4),
		intersects: 1,
	}
	cr.sectors = append(cr.sectors, s)
}

func checkIntersectFlag(s *vvSector) {
	if s.clNextA == s.clPrevA && s.clNextA != base.ON {
		s.intersects = 0
	}
	if s.clNextB == s.clPrevB && s.clNextB != base.ON {
		s.intersects = 0
	}
}

func (cr *vvClassifyRecord) reclassifyOnSectors() {
	for i := range cr.sectors {
		if cr.sectors[i].clNextA == base.ON && cr.sectors[i].clPrevA == base.ON &&
			cr.sectors[i].clNextB == base.ON && cr.sectors[i].clPrevB == base.ON {
			indexA := cr.sectors[i].indexA
			indexB := cr.sectors[i].indexB
			normA := cr.nhA[indexA].he.GetFaceNormal()
			normB := cr.nhB[indexB].he.GetFaceNormal()
			var newAcl, newBcl int
			if normA.Eq(normB) {
				newAcl = boolIf(cr.op == base.BoolUnion, OUT, IN)
				newBcl = boolIf(cr.op == base.BoolUnion, IN, OUT)
			} else {
				newAcl = boolIf(cr.op == base.BoolUnion, IN, OUT)
				newBcl = boolIf(cr.op == base.BoolUnion, IN, OUT)
			}
			for j := range cr.sectors {
				if cr.sectors[j].intersects != 0 {
					if cr.sectors[j].indexA == indexA {
						if cr.sectors[j].clPrevB == base.ON {
							cr.sectors[j].clPrevB = newBcl
						}
						if cr.sectors[j].clNextB == base.ON {
							cr.sectors[j].clNextB = newBcl
						}
					}
					if cr.sectors[j].indexB == indexB {
						if cr.sectors[j].clPrevA == base.ON {
							cr.sectors[j].clPrevA = newAcl
						}
						if cr.sectors[j].clNextA == base.ON {
							cr.sectors[j].clNextA = newAcl
						}
					}
					checkIntersectFlag(&cr.sectors[j])
				}
			}
		}
	}
}

func (cr *vvClassifyRecord) reclassifyDoublyOnEdges(i int) {
	s := &cr.sectors[i]
	if (s.clNextA == base.ON && s.clNextB == base.ON) || (s.clPrevA == base.ON && s.clPrevB == base.ON) {
		newAcl := boolIf(cr.op == base.BoolUnion, OUT, IN)
		newBcl := boolIf(cr.op == base.BoolUnion, IN, OUT)
		indexA := s.indexA
		indexB := s.indexB
		for j := range cr.sectors {
			if cr.sectors[j].intersects != 0 {
				if cr.sectors[j].indexA == indexA {
					if cr.sectors[j].clPrevB == base.ON {
						cr.sectors[j].clPrevB = newBcl
					}
					if cr.sectors[j].clNextB == base.ON {
						cr.sectors[j].clNextB = newBcl
					}
				}
				if cr.sectors[j].indexB == indexB {
					if cr.sectors[j].clPrevA == base.ON {
						cr.sectors[j].clPrevA = newAcl
					}
					if cr.sectors[j].clNextA == base.ON {
						cr.sectors[j].clNextA = newAcl
					}
				}
				checkIntersectFlag(&cr.sectors[j])
			}
		}
	}
}

func (cr *vvClassifyRecord) reclassifySinglyOnEdges(i int) {
	s := &cr.sectors[i]
	indexA := s.indexA
	indexB := s.indexB
	newAcl := boolIf(cr.op == base.BoolUnion, OUT, IN)
	newBcl := boolIf(cr.op == base.BoolUnion, OUT, IN)
	if s.clPrevA == base.ON || s.clNextA == base.ON || s.clPrevB == base.ON || s.clNextB == base.ON {
		for j := range cr.sectors {
			if cr.sectors[j].intersects != 0 {
				if cr.sectors[j].indexA == indexA {
					if cr.sectors[j].clPrevB == base.ON {
						cr.sectors[j].clPrevB = newBcl
					}
					if cr.sectors[j].clNextB == base.ON {
						cr.sectors[j].clNextB = newBcl
					}
				}
				if cr.sectors[j].indexB == indexB {
					if cr.sectors[j].clPrevA == base.ON {
						cr.sectors[j].clPrevA = newAcl
					}
					if cr.sectors[j].clNextA == base.ON {
						cr.sectors[j].clNextA = newAcl
					}
				}
				checkIntersectFlag(&cr.sectors[j])
			}
		}
	}
}

func (cr *vvClassifyRecord) reclassifyOnEdges() {
	for i := range cr.sectors {
		if cr.sectors[i].intersects != 0 {
			cr.reclassifyDoublyOnEdges(i)
			cr.reclassifySinglyOnEdges(i)
		}
	}
}

func (cr *vvClassifyRecord) vvShiftDown() {
	n := len(cr.sectors)
	if n == 0 {
		return
	}
	saved := cr.sectors[0]
	for i := 0; i < n-1; i++ {
		cr.sectors[i] = cr.sectors[i+1]
	}
	cr.sectors[n-1] = saved
}

func separate1(from, to *base.HalfEdge) *base.Edge {
	v := to.Vertex.Loc
	_, _, _ = euler.Lmev2(from, to, v)
	return from.Prev.Edge
}

func separate2(he *base.HalfEdge) *base.Edge {
	v := he.Vertex.Loc
	_, _, _ = euler.Lmev2(he, he, v)
	return he.Prev.Edge
}

func (cr *vvClassifyRecord) prepareNextCrossing() bool {
	n := 0
	for k := 0; k < len(cr.sectors); k++ {
		if cr.sectors[k].intersects != 0 {
			if cr.sectors[k].clPrevA == cr.sectors[k].clPrevB {
				cr.sectors[k].intersects = BAD_CROSSING
			}
			cr.sectors[n] = cr.sectors[k]
			n++
		}
	}
	cr.sectors = cr.sectors[:n]
	if n > 1 && cr.op == base.BoolUnion {
		j := 0
		for k := 0; k < n; k++ {
			if cr.sectors[k].intersects != BAD_CROSSING {
				cr.sectors[j] = cr.sectors[k]
				j++
			}
		}
		cr.sectors = cr.sectors[:j]
		n = j
	}
	if n == 0 {
		return false
	}
	if len(cr.sectors) > 0 && cr.sectors[0].clNextA != OUT {
		cr.vvShiftDown()
	}
	if n == 1 {
		s := &cr.sectors[0]
		normA := cr.nhA[s.indexA].he.GetMate().GetFaceNormal()
		normB := cr.nhB[s.indexB].he.GetMate().GetFaceNormal()
		s.onFace = 0
		if normA.Eq(normB) {
			s.onFace = SAME_FACE
		} else if normA.Eq(normB.Negate()) {
			s.onFace = OPP_FACE
			if cr.op != base.BoolUnion && (s.intersects == 2 || s.intersects == 4) {
				cr.sectors = cr.sectors[:0]
				return false
			}
		}
	} else if n > 1 {
		j := 0
		for k := 0; k < n; k++ {
			if cr.sectors[k].intersects != BAD_CROSSING {
				cr.sectors[j] = cr.sectors[k]
				j++
			}
		}
		cr.sectors = cr.sectors[:j]
		if len(cr.sectors) > 0 && cr.sectors[0].clNextA != OUT {
			cr.vvShiftDown()
		}
	}
	return len(cr.sectors) > 0
}

func (cr *vvClassifyRecord) findNextCrossing() (aHead, aTail, bHead, bTail *base.HalfEdge) {
	n := len(cr.sectors)
	if n == 0 {
		return nil, nil, nil, nil
	}
	if n > 1 {
		aHead = cr.nhA[cr.sectors[0].indexA].he
		bHead = cr.nhB[cr.sectors[0].indexB].he
		aTail = cr.nhA[cr.sectors[1].indexA].he
		bTail = cr.nhB[cr.sectors[1].indexB].he
		cr.sectors[0].intersects = 0
		cr.sectors[1].intersects = 0
	} else {
		s := &cr.sectors[0]
		aHead = cr.nhA[s.indexA].he
		bHead = cr.nhB[s.indexB].he
		aIs180 := aHead.Is180()
		bIs180 := bHead.Is180()
		if aIs180 {
			aTail = aHead.GetMate().Next
		} else {
			aTail = aHead
		}
		if bIs180 {
			bTail = bHead.GetMate().Next
		} else {
			bTail = bHead
		}
		s.intersects = 0
	}
	return aHead, aTail, bHead, bTail
}

func (cr *vvClassifyRecord) vvSeperate(aHead, aTail, bHead, bTail *base.HalfEdge) {
	var e1, e2 *base.Edge
	if aHead == aTail {
		e1 = separate2(aHead)
		cr.br.EdgesA = append(cr.br.EdgesA, e1)
		e2 = separate1(bHead, bTail)
		cr.br.EdgesB = append(cr.br.EdgesB, e2)
	} else if bHead == bTail {
		e2 = separate1(aTail, aHead)
		cr.br.EdgesA = append(cr.br.EdgesA, e2)
		e1 = separate2(bHead)
		cr.br.EdgesB = append(cr.br.EdgesB, e1)
	} else {
		e1 = separate1(aTail, aHead)
		cr.br.EdgesA = append(cr.br.EdgesA, e1)
		e2 = separate1(bHead, bTail)
		cr.br.EdgesB = append(cr.br.EdgesB, e2)
	}
}

func (cr *vvClassifyRecord) insertNullEdges() {
	for cr.prepareNextCrossing() {
		aHead, aTail, bHead, bTail := cr.findNextCrossing()
		if aHead != nil && aTail != nil && bHead != nil && bTail != nil {
			cr.vvSeperate(aHead, aTail, bHead, bTail)
		}
	}
}

func (cr *vvClassifyRecord) vvClassify(i int) {
	cr.nhA = cr.nhA[:0]
	cr.nhB = cr.nhB[:0]
	cr.sectors = cr.sectors[:0]
	cr.setVertexVertex(i)
	cr.preprocess(cr.v, false)
	cr.preprocess(cr.v2, true)
	cr.getNeighborhood()
	cr.reclassifyOnSectors()
	cr.reclassifyOnEdges()
	if len(cr.sectors) > 0 {
		cr.insertNullEdges()
	}
}

func (br *BoolopRecord) Classify() bool {
	br.EdgesA = br.EdgesA[:0]
	br.EdgesB = br.EdgesB[:0]
	br.NoVV = len(br.VertsV) == 0
	fRec := &vfClassifyRecord{}
	fRec.reset(br)
	for i := 0; i < len(br.VertsA); i++ {
		fRec.vfClassify(i, false)
	}
	for i := 0; i < len(br.VertsB); i++ {
		fRec.vfClassify(i, true)
	}
	vRec := &vvClassifyRecord{}
	vRec.reset(br)
	for i := 0; i < len(br.VertsV); i++ {
		vRec.vvClassify(i)
	}
	br.VertsA = br.VertsA[:0]
	br.VertsB = br.VertsB[:0]
	br.VertsV = br.VertsV[:0]
	return len(br.EdgesA) > 2
}

func boolIf(cond bool, trueVal, falseVal int) int {
	if cond {
		return trueVal
	}
	return falseVal
}
