package solid

// Ported from vraniml/src/solid/algorithms/bool/
// Boolean operations on half-edge boundary representation solids.

// ---------------------------------------------------------------------------
// Intersection records — bookkeeping for the Generate phase.
// ---------------------------------------------------------------------------

// VertVertRecord records a vertex-to-vertex association created when an edge
// of one solid intersects an edge of the other, or when two vertices coincide.
type VertVertRecord struct {
	Va *Vertex // Vertex from solid A
	Vb *Vertex // Vertex from solid B
}

// VertFaceRecord records a vertex-on-face association created when an edge
// pierces a face (FaceHit) or when a vertex lies on a face plane.
type VertFaceRecord struct {
	V *Vertex
	F *Face
}

// ---------------------------------------------------------------------------
// BoolopRecord — mutable state for one boolean operation.
// ---------------------------------------------------------------------------

// BoolopRecord holds all working state for a boolean operation between two solids.
type BoolopRecord struct {
	A  *Solid // Solid A (left operand)
	B  *Solid // Solid B (right operand)
	Op int    // BoolUnion, BoolIntersection, or BoolDifference

	Result *Solid // The result solid (set by Finish or LastDitch)

	// Vertex-Vertex intersection records (edge crosses edge, or vertex-on-vertex)
	VertsV []VertVertRecord

	// Vertex-Face records: vertices of A lying on faces of B
	VertsA []VertFaceRecord

	// Vertex-Face records: vertices of B lying on faces of A
	VertsB []VertFaceRecord

	// Null edges created during Classify (for issue #33)
	EdgesA []*Edge
	EdgesB []*Edge

	// New faces created during Connect/Finish (for issue #33)
	FacesA []*Face
	FacesB []*Face

	// NoVV is true when no vertex-vertex intersections were found,
	// meaning all intersections are edge-pierces-face (VF only).
	NoVV bool

	// Quit signals that the operation should be aborted.
	Quit bool
}

// NewBoolopRecord creates an empty boolean operation record.
func NewBoolopRecord() *BoolopRecord {
	return &BoolopRecord{}
}

// Reset prepares the record for a new boolean operation.
func (br *BoolopRecord) Reset(a, b *Solid, op int) {
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
}

// ---------------------------------------------------------------------------
// AddVertFace — record a vertex lying on a face.
// ---------------------------------------------------------------------------

// AddVertFace records that vertex v lies on face f. The bVsA flag indicates
// whether the face's solid is A (the vertex is from B hitting a face of A):
//
//	bVsA == true  → face belongs to solid A → store in VertsB (B's vertex on A's face)
//	bVsA == false → face belongs to solid B → store in VertsA (A's vertex on B's face)
func (br *BoolopRecord) AddVertFace(v *Vertex, f *Face, bVsA bool) {
	v.Mark = ON
	f.Mark1 = ON

	if bVsA {
		// Check for duplicate
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

// ---------------------------------------------------------------------------
// AddVertVert — record a vertex-vertex association.
// ---------------------------------------------------------------------------

// AddVertVert records that two vertices (one from each solid) coincide.
// The bVsA flag determines which vertex is "A's" and which is "B's":
//
//	bVsA == true  → v1 is from B, v2 is from A → store (va=v2, vb=v1)
//	bVsA == false → v1 is from A, v2 is from B → store (va=v1, vb=v2)
func (br *BoolopRecord) AddVertVert(v1, v2 *Vertex, bVsA bool) {
	v1.Mark = ON
	v2.Mark = ON
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

// ---------------------------------------------------------------------------
// Generate — Phase 1: find all edge-face intersections.
// ---------------------------------------------------------------------------

// Generate iterates all edges of each solid against all faces of the other,
// detecting intersections and splitting edges/faces via Euler operators.
func (br *BoolopRecord) Generate() {
	// Check each edge of solid A against all faces of solid B.
	// Snapshot edges first since ProcessEdge may split edges (modifying the list).
	edgesA := br.snapshotEdges(br.A)
	for _, e := range edgesA {
		e.Mark = UNKNOWN
		br.processEdge(br.B, e)
	}

	// Check each edge of solid B against all faces of solid A.
	edgesB := br.snapshotEdges(br.B)
	for _, e := range edgesB {
		e.Mark = UNKNOWN
		br.processEdge(br.A, e)
	}
}

// snapshotEdges collects all current edges of solid s into a slice.
// This is necessary because topology changes during Generate modify the edge list.
func (br *BoolopRecord) snapshotEdges(s *Solid) []*Edge {
	var edges []*Edge
	for e := s.Edges; e != nil; e = e.Next {
		edges = append(edges, e)
	}
	return edges
}

// processEdge checks one edge against all faces of the given solid.
func (br *BoolopRecord) processEdge(s *Solid, e *Edge) {
	for f := s.Faces; f != nil; f = f.Next {
		br.doGenerate(e, f)
	}
}

// doGenerate is the core intersection detection routine. It tests whether
// edge e (from one solid) intersects face f (from the other solid).
//
// Three cases:
//  1. Edge straddles face plane → compute intersection point, split edge,
//     classify hit type (FaceHit / EdgeHit / VertexHit), record association.
//  2. One or both endpoints lie ON the face plane → call vertexOnFace.
//  3. Both endpoints on same side → no intersection.
func (br *BoolopRecord) doGenerate(e *Edge, f *Face) {
	edgeSolid := e.GetSolid()
	faceSolid := f.Solid
	if faceSolid == edgeSolid {
		return // sanity: never test an edge against its own solid
	}

	v1 := e.He1.Vertex
	v2 := e.He2.Vertex
	d1 := f.GetDistance(v1.Loc)
	d2 := f.GetDistance(v2.Loc)
	s1 := floatCompare(d1, 0)
	s2 := floatCompare(d2, 0)

	// bVsA: is the face's solid the A solid? (determines record direction)
	bVsA := faceSolid == br.A

	// Case 1: edge straddles the face plane
	if (s1 == -1 && s2 == 1) || (s1 == 1 && s2 == -1) {
		// Compute intersection point via linear interpolation
		t := d1 / (d1 - d2)
		diff := v2.Loc.Sub(v1.Loc)
		pt := v1.Loc.Add(diff.Scale(t))

		// Check what the intersection hits on the face
		testVert := &Vertex{Loc: pt}
		rec := NewIntersectRecord()
		if f.FaceContains(testVert, &rec) {
			// Split the edge at the intersection point
			Lmev2(e.He1, e.He2.Next, pt)

			switch rec.GetType() {
			case FaceHit:
				br.AddVertFace(e.He1.Vertex, f, bVsA)

			case EdgeHit:
				// Also split the face's edge at the intersection point
				if rec.He != nil {
					Lmev2(rec.He, rec.He.GetMate().Next, pt)
					br.AddVertVert(e.He1.Vertex, rec.He.Vertex, bVsA)
				}

			case VertexHit:
				if rec.Vert != nil {
					br.AddVertVert(e.He1.Vertex, rec.Vert, bVsA)
				}
			}

			// The newly created edge (the "remainder" after splitting)
			// may cross additional faces. Process it recursively.
			newEdge := e.He1.Prev.Edge
			if newEdge != e {
				br.processEdge(faceSolid, newEdge)
			}
		}
		// If !Contains: intersection is outside the face polygon (miss) — do nothing.
		return
	}

	// Case 2: one or both vertices lie ON the face plane
	if s1 == 0 {
		br.vertexOnFace(v1, f)
	}
	if s2 == 0 {
		br.vertexOnFace(v2, f)
	}
}

// vertexOnFace handles the case where a vertex from one solid lies exactly
// on the plane of a face from the other solid. It checks whether the vertex
// actually lies within the face polygon and records the appropriate association.
func (br *BoolopRecord) vertexOnFace(v *Vertex, f *Face) {
	faceSolid := f.Solid
	bVsA := faceSolid == br.A

	rec := NewIntersectRecord()
	if !f.FaceContains(v, &rec) {
		return
	}

	switch rec.GetType() {
	case FaceHit:
		br.AddVertFace(v, f, bVsA)

	case EdgeHit:
		if rec.He != nil {
			Lmev2(rec.He, rec.He.GetMate().Next, v.Loc)
			br.AddVertVert(v, rec.He.Vertex, bVsA)
		}

	case VertexHit:
		if rec.Vert != nil {
			br.AddVertVert(v, rec.Vert, bVsA)
		}
	}
}

// ---------------------------------------------------------------------------
// LastDitch — handle degenerate cases (no intersections found).
// ---------------------------------------------------------------------------

// LastDitch handles cases where Classify finds no intersections: containment,
// identity, or disjointness. Returns true if a result was produced.
func (br *BoolopRecord) LastDitch() bool {
	aContainsB := br.A.SolidContains(br.B)
	bContainsA := br.B.SolidContains(br.A)

	switch {
	case aContainsB && bContainsA:
		// Identical (or nearly so)
		return br.lastDitchIdentical()
	case aContainsB:
		// A properly contains B
		return br.lastDitchAContainsB()
	case bContainsA:
		// B properly contains A
		return br.lastDitchBContainsA()
	default:
		// Completely disjoint
		return br.lastDitchDisjoint()
	}
}

func (br *BoolopRecord) lastDitchIdentical() bool {
	switch br.Op {
	case BoolUnion:
		br.Result = br.A.Copy()
		return true
	case BoolIntersection:
		br.Result = br.B.Copy()
		return true
	case BoolDifference:
		br.Result = nil
		return false
	}
	return false
}

func (br *BoolopRecord) lastDitchAContainsB() bool {
	switch br.Op {
	case BoolUnion:
		br.Result = br.A.Copy()
		return true
	case BoolIntersection:
		br.Result = br.B.Copy()
		return true
	case BoolDifference:
		br.Result = br.A.Copy()
		return true
	}
	return false
}

func (br *BoolopRecord) lastDitchBContainsA() bool {
	switch br.Op {
	case BoolUnion:
		br.Result = br.B.Copy()
		return true
	case BoolIntersection:
		br.Result = br.A.Copy()
		return true
	case BoolDifference:
		br.Result = nil
		return false
	}
	return false
}

func (br *BoolopRecord) lastDitchDisjoint() bool {
	switch br.Op {
	case BoolUnion:
		br.Result = br.A.Copy()
		br.Result.Merge(br.B.Copy())
		return true
	case BoolIntersection:
		br.Result = nil
		return false
	case BoolDifference:
		br.Result = br.A.Copy()
		return true
	}
	return false
}

// Classify, Connect, and Finish are implemented in:
//   bool_classify.go  — Phase 2: classify regions as IN/OUT/ON, insert null edges.
//   bool_connect.go   — Phase 3: sort null edges, match loose ends, join/cut.
//   bool_finish.go    — Phase 4: build result solid from paired faces.

// Complete runs post-operation cleanup: renumber indices and remove
// coplanar/collinear degeneracies.
func (br *BoolopRecord) Complete() {
	if br.Result != nil {
		br.Result.Renumber()
		// Note: C++ RemoveCoplaneColine is a no-op stub (just returns).
		// The Go implementation can cause iterator invalidation on
		// boolean results.  Skip it here to match C++ behavior.
	}
}

// ---------------------------------------------------------------------------
// BoolOp — top-level entry point.
// ---------------------------------------------------------------------------

// BoolOp performs a boolean operation (BoolUnion, BoolIntersection, or
// BoolDifference) on solids A and B. It works on copies to preserve the
// originals. Returns the result solid (nil if the result is empty) and
// true if the operation succeeded.
func BoolOp(a, b *Solid, op int) (*Solid, bool) {
	// Work on copies so originals are untouched.
	workA := a.Copy()
	workB := b.Copy()

	// Prepare plane equations and marks.
	workA.CalcPlaneEquations()
	workB.CalcPlaneEquations()
	workA.SetFaceMarks(UNKNOWN)
	workA.SetVertexMarks(UNKNOWN)
	workB.SetFaceMarks(UNKNOWN)
	workB.SetVertexMarks(UNKNOWN)

	br := NewBoolopRecord()
	br.Reset(workA, workB, op)

	// Phase 1: Generate intersections.
	br.Generate()
	if br.Quit {
		return nil, false
	}

	// Phase 2: Classify regions.
	if !br.Classify() {
		if br.Quit {
			return nil, false
		}
		// No crossings found — try containment/disjoint handling.
		ok := br.LastDitch()
		if ok {
			br.Complete()
		}
		return br.Result, ok
	}

	// Phase 3: Connect crossings.
	br.Connect()
	if br.Quit {
		return nil, false
	}

	// Phase 4: Build result.
	br.Finish()
	if br.Quit {
		return nil, false
	}

	// Phase 5: Cleanup.
	br.Complete()

	return br.Result, br.Result != nil
}

// ---------------------------------------------------------------------------
// Convenience wrappers
// ---------------------------------------------------------------------------

// Union returns A ∪ B.
func Union(a, b *Solid) (*Solid, bool) {
	return BoolOp(a, b, BoolUnion)
}

// Intersection returns A ∩ B.
func Intersection(a, b *Solid) (*Solid, bool) {
	return BoolOp(a, b, BoolIntersection)
}

// Difference returns A − B.
func Difference(a, b *Solid) (*Solid, bool) {
	return BoolOp(a, b, BoolDifference)
}

// ---------------------------------------------------------------------------
// Helper: SetEdgeMarks
// ---------------------------------------------------------------------------

// SetEdgeMarks sets Mark on all edges.
func (s *Solid) SetEdgeMarks(m uint64) {
	for e := s.Edges; e != nil; e = e.Next {
		e.Mark = m
	}
}
