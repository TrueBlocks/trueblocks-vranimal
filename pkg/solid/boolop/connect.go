package boolop

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
)

type nullEdgeInfo struct {
	hes [4]*base.HalfEdge
	cl  [4]bool
}

func (br *BoolopRecord) Connect() {
	br.FacesA = br.FacesA[:0]
	br.FacesB = br.FacesB[:0]
	nEndsA := 0
	nEndsB := 0
	endsA := make([]*base.HalfEdge, 0, len(br.EdgesA))
	endsB := make([]*base.HalfEdge, 0, len(br.EdgesA))
	br.sortNullEdges()
	things := make([]nullEdgeInfo, len(br.EdgesA))
	for i := 0; i < len(br.EdgesA); i++ {
		e := br.EdgesA[i]
		things[i].hes[0] = e.He1
		things[i].cl[0] = e.He1.Loop.IsOuterLoop()
		things[i].hes[1] = e.He2
		things[i].cl[1] = e.He2.Loop.IsOuterLoop()
		e2 := br.EdgesB[i]
		things[i].hes[2] = e2.He1
		things[i].cl[2] = e2.He1.Loop.IsOuterLoop()
		things[i].hes[3] = e2.He2
		things[i].cl[3] = e2.He2.Loop.IsOuterLoop()
	}
	getThingCl := func(he *base.HalfEdge) bool {
		if he.Edge == nil {
			return false
		}
		idx := he.Edge.Index
		if int(idx) >= len(things) {
			return false
		}
		mate := he.GetMate()
		for i := 0; i < 4; i++ {
			if things[idx].hes[i] == mate {
				return things[idx].cl[i]
			}
		}
		return false
	}
	specialCase := func(he1, he2 *base.HalfEdge) bool {
		l := he1.Loop
		if l == he2.Loop {
			if he1.Prev.Prev != he2 {
				if !l.IsOuterLoop() {
					cl1 := getThingCl(he1)
					cl2 := getThingCl(he2)
					if cl1 && cl2 {
						return false
					}
					return true
				}
			}
		}
		return false
	}
	canJoin := func(hea, heb *base.HalfEdge) (*base.HalfEdge, *base.HalfEdge) {
		hea.SetMark(base.NOT_LOOSE)
		heb.SetMark(base.NOT_LOOSE)
		for i := 0; i < nEndsA; i++ {
			if hea.IsNeighbor(endsA[i]) && heb.IsNeighbor(endsB[i]) {
				retA := endsA[i]
				retB := endsB[i]
				for j := i + 1; j < nEndsA; j++ {
					endsA[j-1] = endsA[j]
					endsB[j-1] = endsB[j]
				}
				nEndsA--
				nEndsB--
				endsA = endsA[:nEndsA]
				endsB = endsB[:nEndsB]
				retA.SetMark(base.NOT_LOOSE)
				retB.SetMark(base.NOT_LOOSE)
				return retA, retB
			}
		}
		hea.SetMark(base.LOOSE)
		heb.SetMark(base.LOOSE)
		endsA = append(endsA, hea)
		endsB = append(endsB, heb)
		nEndsA++
		nEndsB++
		return nil, nil
	}
	for idx := 0; idx < len(br.EdgesA); idx++ {
		edgeA := br.EdgesA[idx]
		edgeB := br.EdgesB[idx]
		outA := edgeA.He1
		inB := edgeB.He2
		inA := edgeA.He2
		outB := edgeB.He1
		var joinInA, joinOutA, joinInB, joinOutB *base.HalfEdge
		joinInA, joinOutB = canJoin(outA, inB)
		if joinInA != nil {
			algorithms.Join(joinInA, outA, specialCase(joinInA, outA))
			if !joinInA.GetMate().Marked(base.LOOSE) {
				f := algorithms.Cut(joinInA, true)
				if f != nil {
					br.FacesA = append(br.FacesA, f)
				}
			}
			algorithms.Join(joinOutB, inB, false)
			if !joinOutB.GetMate().Marked(base.LOOSE) {
				f := algorithms.Cut(joinOutB, true)
				if f != nil {
					br.FacesB = append(br.FacesB, f)
				}
			}
		}
		joinOutA, joinInB = canJoin(inA, outB)
		if joinOutA != nil {
			algorithms.Join(joinOutA, inA, false)
			if !joinOutA.GetMate().Marked(base.LOOSE) {
				f := algorithms.Cut(joinOutA, true)
				if f != nil {
					br.FacesA = append(br.FacesA, f)
				}
			}
			algorithms.Join(joinInB, outB, specialCase(joinInB, outB))
			if !joinInB.GetMate().Marked(base.LOOSE) {
				f := algorithms.Cut(joinInB, true)
				if f != nil {
					br.FacesB = append(br.FacesB, f)
				}
			}
		}
		if joinInA != nil && joinInB != nil && joinOutA != nil && joinOutB != nil {
			f := algorithms.Cut(outA, true)
			if f != nil {
				br.FacesA = append(br.FacesA, f)
			}
			f = algorithms.Cut(outB, true)
			if f != nil {
				br.FacesB = append(br.FacesB, f)
			}
		}
	}
	br.EdgesA = br.EdgesA[:0]
	br.EdgesB = br.EdgesB[:0]
}

func (br *BoolopRecord) sortNullEdges() {
	n := len(br.EdgesA)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			v1 := br.EdgesA[i].He1.Vertex
			v2 := br.EdgesA[j].He1.Vertex
			if vertexGreaterThan(v1, v2) {
				br.EdgesA[i], br.EdgesA[j] = br.EdgesA[j], br.EdgesA[i]
				br.EdgesB[i], br.EdgesB[j] = br.EdgesB[j], br.EdgesB[i]
			}
		}
	}
	for i := 0; i < n; i++ {
		br.EdgesA[i].Index = uint64(i)
		br.EdgesB[i].Index = uint64(i)
	}
}

func vertexGreaterThan(v1, v2 *base.Vertex) bool {
	res := base.FloatCompare(v1.Loc.X - v2.Loc.X)
	if res == -1 {
		return false
	}
	if res == 0 {
		res = base.FloatCompare(v1.Loc.Y - v2.Loc.Y)
		if res == -1 {
			return false
		}
		if res == 0 {
			res = base.FloatCompare(v1.Loc.Z - v2.Loc.Z)
			if res == -1 {
				return false
			}
		}
	}
	return true
}
