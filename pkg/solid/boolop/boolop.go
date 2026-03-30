package boolop

import (
	"math"

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
	A             *base.Solid
	B             *base.Solid
	Op            int
	Result        *base.Solid
	VertsV        []VertVertRecord
	VertsA        []VertFaceRecord
	VertsB        []VertFaceRecord
	EdgesA        []*base.Edge
	EdgesB        []*base.Edge
	FacesA        []*base.Face
	FacesB        []*base.Face
	NoVV          bool
	Quit          bool
	Perturbed     bool // input was perturbed to resolve degeneracy
	UsedLastDitch bool // result came from the LastDitch fallback
}

// lastDitchUsed is set by LastDitch() and cleared by BoolOp() at the start.
var lastDitchUsed bool

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
	var ok bool
	switch {
	case aContainsB && bContainsA:
		ok = br.lastDitchIdentical()
	case aContainsB:
		ok = br.lastDitchAContainsB()
	case bContainsA:
		ok = br.lastDitchBContainsA()
	default:
		ok = br.lastDitchDisjoint()
	}
	if ok {
		br.UsedLastDitch = true
		lastDitchUsed = true
	}
	return ok
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
	lastDitchUsed = false
	if !IsDegenerate(a, b) {
		return boolOpCore(a, b, op, false)
	}
	// For Difference, try targeted face-normal perturbation first.
	if op == base.BoolDifference {
		if res, ok := boolOpCoreTargeted(a, b, op); ok {
			return res, true
		}
	}
	// Fall back to uniform perturbation.
	return boolOpCore(a, b, op, true)
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
	if perturb && br.Result != nil {
		snapVertices(br.Result)
	}
	return br.Result, br.Result != nil
}

// snapVertices removes noise introduced by the degeneracy perturbation.
// For each vertex coordinate, if rounding to 4 decimal places changes the
// value by less than 2*eps, snap it. This collapses 1.5001 → 1.5,
// -0.4999 → -0.5, etc., while leaving legitimately precise coordinates alone.
func snapVertices(s *base.Solid) {
	const snapGrid = 1e-3 // coarser than eps=1e-4
	for v := s.Verts; v != nil; v = v.Next {
		v.Loc.X = math.Round(v.Loc.X/snapGrid) * snapGrid
		v.Loc.Y = math.Round(v.Loc.Y/snapGrid) * snapGrid
		v.Loc.Z = math.Round(v.Loc.Z/snapGrid) * snapGrid
	}
}

// boolOpCoreTargeted tries a Difference with face-normal perturbation:
// only B-faces that are coplanar with an A-face get pushed along the
// A-face normal. Returns (nil, false) if the result is degenerate.
func boolOpCoreTargeted(a, b *base.Solid, op int) (*base.Solid, bool) {
	workA := a.Copy()
	workB := b.Copy()
	workA.CalcPlaneEquations()
	workB.CalcPlaneEquations()
	perturbCoincidentFaces(workA, workB)
	// Re-check: if still degenerate, signal failure so caller falls back.
	if IsDegenerate(workA, workB) {
		return nil, false
	}
	workA.SetFaceMarks(base.UNKNOWN)
	workA.SetVertexMarks(base.UNKNOWN)
	workB.SetFaceMarks(base.UNKNOWN)
	workB.SetVertexMarks(base.UNKNOWN)
	br := NewBoolopRecord()
	br.Reset(workA, workB, op)
	br.Perturbed = true
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
	if br.Result != nil {
		snapVertices(br.Result)
	}
	return br.Result, br.Result != nil
}

// perturbCoincidentFaces pushes B vertices that lie on coplanar A-faces
// along the A-face normal. Each vertex accumulates contributions from
// every coplanar face pair it belongs to, then all offsets are applied
// at once. This avoids the problem of a vertex shared by multiple
// coplanar faces being only partially resolved.
func perturbCoincidentFaces(a, b *base.Solid) {
	const eps = 1e-4
	// Accumulate a displacement vector for each B vertex.
	accum := make(map[*base.Vertex]vec.SFVec3f)
	for fB := b.Faces; fB != nil; fB = fB.Next {
		for fA := a.Faces; fA != nil; fA = fA.Next {
			if !facesCoplanar(fA, fB) {
				continue
			}
			// Push every vertex of this B face along A's normal.
			push := fA.Normal.Scale(eps)
			he := fB.GetFirstHe()
			if he == nil {
				continue
			}
			start := he
			for {
				v := he.Vertex
				prev := accum[v]
				accum[v] = prev.Add(push)
				he = he.Next
				if he == start {
					break
				}
			}
		}
	}
	// Apply accumulated offsets.
	for v, offset := range accum {
		v.Loc = v.Loc.Add(offset)
	}
	// Recompute plane equations after moving vertices.
	b.CalcPlaneEquations()
}

// facesCoplanar returns true if fA and fB lie on the same plane
// (parallel normals, same signed distance for all fB vertices).
func facesCoplanar(fA, fB *base.Face) bool {
	// Check normals are parallel (dot product ≈ ±1).
	dot := fA.Normal.Dot(fB.Normal)
	if dot < 0.999 && dot > -0.999 {
		return false
	}
	// Check every vertex of fB lies on fA's plane.
	he := fB.GetFirstHe()
	if he == nil {
		return false
	}
	start := he
	for {
		d := fA.GetDistance(he.Vertex.Loc)
		if base.FloatCompare(d) != 0 {
			return false
		}
		he = he.Next
		if he == start {
			break
		}
	}
	return true
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

// BoolOpResult holds the result of a boolean operation with metadata.
type BoolOpResult struct {
	Solid         *base.Solid
	Ok            bool
	UsedLastDitch bool
}

// BoolOpEx is like BoolOp but returns additional metadata about the operation.
func BoolOpEx(a, b *base.Solid, op int) BoolOpResult {
	res, ok := BoolOp(a, b, op)
	return BoolOpResult{Solid: res, Ok: ok, UsedLastDitch: lastDitchUsed}
}

// ChainBoolOpEx applies a boolean operation across a slice of solids left-to-right:
// solids[0] op solids[1] → result, result op solids[2] → result, etc.
// Returns the accumulated result. Requires at least 2 solids.
func ChainBoolOpEx(solids []*base.Solid, op int) BoolOpResult {
	if len(solids) < 2 {
		if len(solids) == 1 {
			return BoolOpResult{Solid: solids[0], Ok: true}
		}
		return BoolOpResult{}
	}

	anyLastDitch := false
	acc := solids[0]
	for i := 1; i < len(solids); i++ {
		r := BoolOpEx(acc, solids[i], op)
		if !r.Ok || r.Solid == nil {
			return BoolOpResult{Solid: nil, Ok: false, UsedLastDitch: anyLastDitch || r.UsedLastDitch}
		}
		if r.UsedLastDitch {
			anyLastDitch = true
		}
		acc = r.Solid
		acc.CalcPlaneEquations()
	}
	return BoolOpResult{Solid: acc, Ok: true, UsedLastDitch: anyLastDitch}
}

func makeRing(f *base.Face, v vec.SFVec3f) *base.Edge {
	head := f.LoopOut.GetFirstHe()
	_, _, _ = euler.Lmev2(head, head, v)
	he1 := head.Prev
	_, _, _ = euler.Lmev2(he1, he1, v)
	_ = euler.Lkemr(he1.GetMate())
	return he1.Prev.Edge
}
