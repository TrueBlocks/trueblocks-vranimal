package boolop

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

type VertVertRecord struct {
	Va *base.Vertex
	Vb *base.Vertex
}

type VertFaceRecord struct {
	V *base.Vertex
	F *base.Face
}

type BoolopRecord struct {
	A         *base.Solid
	B         *base.Solid
	Op        int
	Result    *base.Solid
	VertsV    []VertVertRecord
	VertsA    []VertFaceRecord
	VertsB    []VertFaceRecord
	EdgesA    []*base.Edge
	EdgesB    []*base.Edge
	FacesA    []*base.Face
	FacesB    []*base.Face
	NoVV      bool
	Quit      bool
	Perturbed bool // input was perturbed to resolve degeneracy
}

func NewBoolopRecord() *BoolopRecord {
	return &BoolopRecord{}
}

func (br *BoolopRecord) Reset(a, b *base.Solid, op int) {
	br.A = a
	br.B = b
	br.Op = op
	br.Result = nil
	br.VertsV = br.VertsV[:0]
	br.VertsA = br.VertsA[:0]
	br.VertsB = br.VertsB[:0]
	br.EdgesA = br.EdgesA[:0]
	br.EdgesB = br.EdgesB[:0]
	br.FacesA = br.FacesA[:0]
	br.FacesB = br.FacesB[:0]
	br.NoVV = true
	br.Quit = false
	br.Perturbed = false
}

func (br *BoolopRecord) AddVertFace(v *base.Vertex, f *base.Face, bVsA bool) {
	v.Mark = base.ON
	f.Mark1 = base.ON
	if bVsA {
		for i := range br.VertsB {
			if br.VertsB[i].V == v && br.VertsB[i].F == f {
				return
			}
		}
		br.VertsB = append(br.VertsB, VertFaceRecord{V: v, F: f})
	} else {
		for i := range br.VertsA {
			if br.VertsA[i].V == v && br.VertsA[i].F == f {
				return
			}
		}
		br.VertsA = append(br.VertsA, VertFaceRecord{V: v, F: f})
	}
}

func (br *BoolopRecord) AddVertVert(v1, v2 *base.Vertex, bVsA bool) {
	v1.Mark = base.ON
	v2.Mark = base.ON
	br.NoVV = false
	if bVsA {
		for i := range br.VertsV {
			if br.VertsV[i].Vb == v1 && br.VertsV[i].Va == v2 {
				return
			}
		}
		br.VertsV = append(br.VertsV, VertVertRecord{Va: v2, Vb: v1})
	} else {
		for i := range br.VertsV {
			if br.VertsV[i].Vb == v2 && br.VertsV[i].Va == v1 {
				return
			}
		}
		br.VertsV = append(br.VertsV, VertVertRecord{Va: v1, Vb: v2})
	}
}

func (br *BoolopRecord) Generate() {
	edgesA := snapshotEdges(br.A)
	for _, e := range edgesA {
		e.Mark = base.UNKNOWN
		br.processEdge(br.B, e)
	}
	edgesB := snapshotEdges(br.B)
	for _, e := range edgesB {
		e.Mark = base.UNKNOWN
		br.processEdge(br.A, e)
	}
}

func snapshotEdges(s *base.Solid) []*base.Edge {
	var edges []*base.Edge
	for e := s.Edges; e != nil; e = e.Next {
		edges = append(edges, e)
	}
	return edges
}

func (br *BoolopRecord) processEdge(s *base.Solid, e *base.Edge) {
	for f := s.Faces; f != nil; f = f.Next {
		br.doGenerate(e, f)
	}
}

func (br *BoolopRecord) doGenerate(e *base.Edge, f *base.Face) {
	edgeSolid := e.GetSolid()
	faceSolid := f.Solid
	if faceSolid == edgeSolid {
		return
	}
	v1 := e.He1.Vertex
	v2 := e.He2.Vertex
	d1 := f.GetDistance(v1.Loc)
	d2 := f.GetDistance(v2.Loc)
	s1 := base.FloatCompare(d1)
	s2 := base.FloatCompare(d2)
	bVsA := faceSolid == br.A
	if (s1 == -1 && s2 == 1) || (s1 == 1 && s2 == -1) {
		t := d1 / (d1 - d2)
		diff := v2.Loc.Sub(v1.Loc)
		pt := v1.Loc.Add(diff.Scale(t))
		testVert := &base.Vertex{Loc: pt}
		rec := algorithms.NewIntersectRecord()
		if algorithms.FaceContains(f, testVert, &rec) {
			_, _, _ = euler.Lmev2(e.He1, e.He2.Next, pt)
			switch rec.GetType() {
			case algorithms.FaceHit:
				br.AddVertFace(e.He1.Vertex, f, bVsA)
			case algorithms.EdgeHit:
				if rec.He != nil {
					_, _, _ = euler.Lmev2(rec.He, rec.He.GetMate().Next, pt)
					br.AddVertVert(e.He1.Vertex, rec.He.Vertex, bVsA)
				}
			case algorithms.VertexHit:
				if rec.Vert != nil {
					br.AddVertVert(e.He1.Vertex, rec.Vert, bVsA)
				}
			}
			newEdge := e.He1.Prev.Edge
			if newEdge != e {
				br.processEdge(faceSolid, newEdge)
			}
		}
		return
	}
	if s1 == 0 {
		br.vertexOnFace(v1, f)
	}
	if s2 == 0 {
		br.vertexOnFace(v2, f)
	}
}

func (br *BoolopRecord) vertexOnFace(v *base.Vertex, f *base.Face) {
	faceSolid := f.Solid
	bVsA := faceSolid == br.A
	rec := algorithms.NewIntersectRecord()
	if !algorithms.FaceContains(f, v, &rec) {
		return
	}
	switch rec.GetType() {
	case algorithms.FaceHit:
		br.AddVertFace(v, f, bVsA)
	case algorithms.EdgeHit:
		if rec.He != nil {
			_, _, _ = euler.Lmev2(rec.He, rec.He.GetMate().Next, v.Loc)
			br.AddVertVert(v, rec.He.Vertex, bVsA)
		}
	case algorithms.VertexHit:
		if rec.Vert != nil {
			br.AddVertVert(v, rec.Vert, bVsA)
		}
	}
}

func (br *BoolopRecord) LastDitch() bool {
	aContainsB := algorithms.SolidContains(br.A, br.B)
	bContainsA := algorithms.SolidContains(br.B, br.A)
	switch {
	case aContainsB && bContainsA:
		return br.lastDitchIdentical()
	case aContainsB:
		return br.lastDitchAContainsB()
	case bContainsA:
		return br.lastDitchBContainsA()
	default:
		return br.lastDitchDisjoint()
	}
}

func (br *BoolopRecord) lastDitchIdentical() bool {
	switch br.Op {
	case base.BoolUnion:
		br.Result = br.A.Copy()
		return true
	case base.BoolIntersection:
		br.Result = br.B.Copy()
		return true
	case base.BoolDifference:
		br.Result = nil
		return false
	}
	return false
}

func (br *BoolopRecord) lastDitchAContainsB() bool {
	switch br.Op {
	case base.BoolUnion:
		br.Result = br.A.Copy()
		return true
	case base.BoolIntersection:
		br.Result = br.B.Copy()
		return true
	case base.BoolDifference:
		br.Result = br.A.Copy()
		return true
	}
	return false
}

func (br *BoolopRecord) lastDitchBContainsA() bool {
	switch br.Op {
	case base.BoolUnion:
		br.Result = br.B.Copy()
		return true
	case base.BoolIntersection:
		br.Result = br.A.Copy()
		return true
	case base.BoolDifference:
		br.Result = nil
		return false
	}
	return false
}

func (br *BoolopRecord) lastDitchDisjoint() bool {
	switch br.Op {
	case base.BoolUnion:
		br.Result = br.A.Copy()
		br.Result.Merge(br.B.Copy())
		return true
	case base.BoolIntersection:
		br.Result = nil
		return false
	case base.BoolDifference:
		br.Result = br.A.Copy()
		return true
	}
	return false
}

func (br *BoolopRecord) Complete() {
	if br.Result != nil {
		br.Result.Renumber()
	}
}

// IsDegenerate checks whether two solids have edges lying exactly on the
// other solid's face planes — the configuration that causes dual-Is180
// null-edge degeneracies in the boolean pipeline. This is a read-only
// check that does not modify either solid.
func IsDegenerate(a, b *base.Solid) bool {
	a.CalcPlaneEquations()
	b.CalcPlaneEquations()
	// Check edges of A against faces of B
	for e := a.Edges; e != nil; e = e.Next {
		v1 := e.He1.Vertex
		v2 := e.He2.Vertex
		for f := b.Faces; f != nil; f = f.Next {
			d1 := f.GetDistance(v1.Loc)
			d2 := f.GetDistance(v2.Loc)
			if base.FloatCompare(d1) == 0 && base.FloatCompare(d2) == 0 {
				return true
			}
		}
	}
	// Check edges of B against faces of A
	for e := b.Edges; e != nil; e = e.Next {
		v1 := e.He1.Vertex
		v2 := e.He2.Vertex
		for f := a.Faces; f != nil; f = f.Next {
			d1 := f.GetDistance(v1.Loc)
			d2 := f.GetDistance(v2.Loc)
			if base.FloatCompare(d1) == 0 && base.FloatCompare(d2) == 0 {
				return true
			}
		}
	}
	return false
}

func BoolOp(a, b *base.Solid, op int) (*base.Solid, bool) {
	perturb := IsDegenerate(a, b)
	return boolOpCore(a, b, op, perturb)
}

func boolOpCore(a, b *base.Solid, op int, perturb bool) (*base.Solid, bool) {
	workA := a.Copy()
	workB := b.Copy()
	if perturb {
		const eps = 1e-4 // > BigEps (1e-5) to escape FloatCompare tolerance
		workB.TransformGeometry(vec.TranslationMatrix(eps, eps, eps))
	}
	workA.CalcPlaneEquations()
	workB.CalcPlaneEquations()
	workA.SetFaceMarks(base.UNKNOWN)
	workA.SetVertexMarks(base.UNKNOWN)
	workB.SetFaceMarks(base.UNKNOWN)
	workB.SetVertexMarks(base.UNKNOWN)
	br := NewBoolopRecord()
	br.Reset(workA, workB, op)
	br.Perturbed = perturb
	br.Generate()
	if br.Quit {
		return nil, false
	}
	if !br.Classify() {
		if br.Quit {
			return nil, false
		}
		ok := br.LastDitch()
		if ok {
			br.Complete()
		}
		return br.Result, ok
	}
	br.Connect()
	if br.Quit {
		return nil, false
	}
	br.Finish()
	if br.Quit {
		return nil, false
	}
	br.Complete()
	return br.Result, br.Result != nil
}

func countLoopVerts(f *base.Face) int {
	if f.NLoops() == 0 || f.LoopOut == nil {
		return 0
	}
	he := f.LoopOut.GetFirstHe()
	if he == nil {
		return 0
	}
	cnt := 0
	start := he
	for {
		cnt++
		he = he.Next
		if he == start {
			break
		}
	}
	return cnt
}

func Union(a, b *base.Solid) (*base.Solid, bool) {
	return BoolOp(a, b, base.BoolUnion)
}

func Intersection(a, b *base.Solid) (*base.Solid, bool) {
	return BoolOp(a, b, base.BoolIntersection)
}

func Difference(a, b *base.Solid) (*base.Solid, bool) {
	return BoolOp(a, b, base.BoolDifference)
}

func makeRing(f *base.Face, v vec.SFVec3f) *base.Edge {
	head := f.LoopOut.GetFirstHe()
	_, _, _ = euler.Lmev2(head, head, v)
	he1 := head.Prev
	_, _, _ = euler.Lmev2(he1, he1, v)
	_ = euler.Lkemr(he1.GetMate())
	return he1.Prev.Edge
}
