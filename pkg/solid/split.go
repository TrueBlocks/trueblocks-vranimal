package solid

// Ported from vraniml/src/solid/algorithms/split/
// Split a solid along a plane into two resulting solids (Above and Below).

import (
	"sort"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Plane — splitting plane defined by normal and distance.
// ---------------------------------------------------------------------------

// Plane represents an oriented plane in 3D space: Normal · P + D = 0.
type Plane struct {
	Normal vec.SFVec3f
	D      float32
}

// GetDistance returns the signed distance from point p to the plane.
func (pl *Plane) GetDistance(p vec.SFVec3f) float32 {
	return pl.Normal.Dot(p) + pl.D
}

// ---------------------------------------------------------------------------
// HalfEdge mark helpers (LOOSE / NOT_LOOSE for split connect phase)
// ---------------------------------------------------------------------------

const (
	LOOSE     = 1 << 16
	NOT_LOOSE = 0
)

// SetMark sets the half-edge mark.
func (he *HalfEdge) SetMark(m uint32) { he.Mark = m }

// Marked returns true if the half-edge mark equals m.
func (he *HalfEdge) Marked(m uint32) bool { return he.Mark == m }

// ---------------------------------------------------------------------------
// splitNeighborhood — one sector around a vertex being classified.
// ---------------------------------------------------------------------------

type splitNeighborhood struct {
	sector *HalfEdge
	cl     int // ABOVE, BELOW, or ON
}

// ---------------------------------------------------------------------------
// splitClassifyRecord — per-vertex classification state.
// ---------------------------------------------------------------------------

type splitClassifyRecord struct {
	sp    Plane
	v     *Vertex
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

func (cr *splitClassifyRecord) addNeighborhood(he *HalfEdge, cl int) {
	cr.nbrs = append(cr.nbrs, splitNeighborhood{sector: he, cl: cl})
}

// getNeighborhood walks around vertex v collecting edge classifications.
func (cr *splitClassifyRecord) getNeighborhood() {
	cr.nbrs = cr.nbrs[:0]
	he := cr.v.He
	for {
		vNext := he.Next.Vertex
		cl := floatCompare(cr.sp.GetDistance(vNext.Loc), 0)
		cr.addNeighborhood(he, cl)
		he = he.GetMate().Next
		if he == cr.v.He {
			break
		}
	}
}

// reclassifyOnSectors handles coplanar sectors — when a face is coplanar
// with the splitting plane, both ends are reclassified based on the
// dot product of the face normal with the plane normal.
func (cr *splitClassifyRecord) reclassifyOnSectors() {
	n := len(cr.nbrs)
	epsSq := float32(1e-5 * 1e-5)
	spNormal := cr.sp.Normal

	for i := 0; i < n; i++ {
		facNormal := cr.nbrs[i].sector.GetFaceNormal()
		c := facNormal.Cross(spNormal)
		d := c.Dot(c)

		if d < epsSq {
			prv := (i + n - 1) % n
			dotN := facNormal.Dot(spNormal)
			if floatCompare(dotN, 0) == 1 {
				cr.nbrs[prv].cl = BELOW
				cr.nbrs[i].cl = BELOW
			} else {
				cr.nbrs[prv].cl = ABOVE
				cr.nbrs[i].cl = ABOVE
			}
		}
	}
}

// reclassifyOnEdges resolves remaining ON neighborhoods.
func (cr *splitClassifyRecord) reclassifyOnEdges() {
	n := len(cr.nbrs)
	for i := 0; i < n; i++ {
		if cr.nbrs[i].cl == ON {
			prv := cr.nbrs[(n+i-1)%n].cl
			nxt := cr.nbrs[(i+1)%n].cl
			if prv == nxt {
				cr.nbrs[i].cl = prv
			} else {
				cr.nbrs[i].cl = ABOVE
			}
		}
	}
}

// insertNullEdges scans for BELOW→ABOVE transitions, inserts null edges.
func (cr *splitClassifyRecord) insertNullEdges() {
	n := len(cr.nbrs)
	if n == 0 {
		return
	}

	// Find first BELOW→ABOVE transition.
	cur := 0
	for cur < n {
		nxt := (cur + 1) % n
		if cr.nbrs[cur].cl == BELOW && cr.nbrs[nxt].cl == ABOVE {
			break
		}
		cur++
	}
	if cur == n {
		return
	}

	cur = (cur + 1) % n // move to the ABOVE side
	start := cur
	head := cr.nbrs[start].sector

	for {
		// Find next ABOVE→BELOW transition.
		for {
			nxt := (cur + 1) % n
			if cr.nbrs[cur].cl == ABOVE && cr.nbrs[nxt].cl == BELOW {
				break
			}
			cur = (cur + 1) % n
		}

		tail := cr.nbrs[(cur+1)%n].sector

		// Insert null edge.
		sol := tail.GetSolid()
		v, e := Lmev2(head, tail, head.Vertex.Loc)
		_ = v
		cr.spRec.addEdge(e)

		// Find next BELOW→ABOVE transition.
		for {
			nxt := (cur + 1) % n
			if cr.nbrs[cur].cl == BELOW && cr.nbrs[nxt].cl == ABOVE {
				break
			}
			cur = (cur + 1) % n
			_ = sol
		}

		if (cur+1)%n == start {
			return
		}

		cur = (cur + 1) % n
		head = cr.nbrs[cur].sector
	}
}

// ---------------------------------------------------------------------------
// SplitRecord — state for one split operation.
// ---------------------------------------------------------------------------

// SplitRecord holds the working state for splitting a solid along a plane.
type SplitRecord struct {
	s  *Solid
	sp Plane
	A  **Solid
	B  **Solid

	verts     []*Vertex
	edges     []*Edge
	faces     []*Face
	looseEnds []*HalfEdge

	// DebugLog, if non-nil, receives trace messages during split phases.
	DebugLog func(string)
}

// NewSplitRecord creates a new empty split record.
func NewSplitRecord() *SplitRecord {
	return &SplitRecord{}
}

func (sr *SplitRecord) reset(sp Plane, s *Solid, a, b **Solid) {
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

func (sr *SplitRecord) addVert(v *Vertex) {
	if v.IsMarked(ON) {
		return
	}
	v.Mark = ON
	sr.verts = append(sr.verts, v)
}

func (sr *SplitRecord) addEdge(e *Edge) {
	sr.edges = append(sr.edges, e)
}

func (sr *SplitRecord) addFace(f *Face) {
	sr.faces = append(sr.faces, f)
}

func (sr *SplitRecord) addLooseEnd(he *HalfEdge) {
	sr.looseEnds = append(sr.looseEnds, he)
}

// ---------------------------------------------------------------------------
// Phase 1: Generate — find intersection vertices.
// ---------------------------------------------------------------------------

func (sr *SplitRecord) generate() {
	sr.verts = sr.verts[:0]

	// Precalculate distance from every vertex to the splitting plane.
	for v := sr.s.Verts; v != nil; v = v.Next {
		v.Scratch = sr.sp.GetDistance(v.Loc)
	}

	// Snapshot the edge list before modifying topology.
	var edgeList []*Edge
	for e := sr.s.Edges; e != nil; e = e.Next {
		edgeList = append(edgeList, e)
	}

	// Look for edges whose vertices straddle the plane.
	for _, e := range edgeList {
		e.Mark = UNKNOWN

		v1 := e.He1.Vertex
		v2 := e.He2.Vertex
		d1 := v1.Scratch
		d2 := v2.Scratch
		s1 := floatCompare(d1, 0)
		s2 := floatCompare(d2, 0)

		if (s1 == -1 && s2 == 1) || (s1 == 1 && s2 == -1) {
			// Edge straddles — compute intersection point.
			t := d1 / (d1 - d2)
			diff := v2.Loc.Sub(v1.Loc)
			intersectionPt := v1.Loc.Add(diff.Scale(t))

			he := e.He2.Next
			newV, _ := Lmev2(e.He1, he, intersectionPt)
			newV.Scratch = 0
			newV.Mark = UNKNOWN // ensure addVert doesn't skip (ON == 0 == default)
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

// ---------------------------------------------------------------------------
// Phase 2: Classify — insert null edges at on-plane vertices.
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Phase 3: Connect — link null edges to close the split.
// ---------------------------------------------------------------------------

func (sr *SplitRecord) connect() {
	sr.faces = sr.faces[:0]

	// Sort null edges by vertex position (x, y, z).
	sort.Slice(sr.edges, func(i, j int) bool {
		v1 := sr.edges[i].He1.Vertex.Loc
		v2 := sr.edges[j].He1.Vertex.Loc
		if c := floatCompare(v1.X, v2.X); c != 0 {
			return c < 0
		}
		if c := floatCompare(v1.Y, v2.Y); c != 0 {
			return c < 0
		}
		return floatCompare(v1.Z, v2.Z) < 0
	})

	for idx := 0; idx < len(sr.edges); idx++ {
		nextEdge := sr.edges[idx]

		he1 := sr.canJoin(nextEdge.He1)
		he2 := sr.canJoin(nextEdge.He2)

		if he1 != nil {
			sr.s.Join(he1, nextEdge.He1, false)
			if !he1.GetMate().Marked(LOOSE) {
				if ff := sr.s.Cut(he1, false); ff != nil {
					sr.addFace(ff)
				}
			}
		}

		if he2 != nil {
			sr.s.Join(he2, nextEdge.He2, false)
			if !he2.GetMate().Marked(LOOSE) {
				if ff := sr.s.Cut(he2, false); ff != nil {
					sr.addFace(ff)
				}
			}
		}

		if he1 != nil && he2 != nil {
			if ff := sr.s.Cut(nextEdge.He1, false); ff != nil {
				sr.addFace(ff)
			}
		}
	}

	sr.edges = sr.edges[:0]
}

// canJoin checks if he can be joined to a loose end. Returns the matching
// loose end, or nil.
func (sr *SplitRecord) canJoin(he *HalfEdge) *HalfEdge {
	he.SetMark(NOT_LOOSE)

	for i, loose := range sr.looseEnds {
		if he.IsNeighbor(loose) {
			loose.SetMark(NOT_LOOSE)
			// Remove from looseEnds.
			sr.looseEnds = append(sr.looseEnds[:i], sr.looseEnds[i+1:]...)
			return loose
		}
	}

	he.SetMark(LOOSE)
	sr.addLooseEnd(he)
	return nil
}

// ---------------------------------------------------------------------------
// Phase 4: Finish — separate the solid into Above and Below.
// ---------------------------------------------------------------------------

func (sr *SplitRecord) finish() {
	nf := len(sr.faces)
	// For each face created during Connect, fix outer loop orientation
	// and create the companion face via Lmfkrh.
	for i := 0; i < nf; i++ {
		f := sr.faces[i]
		if f.GetFirstLoop() == f.LoopOut {
			f.LoopOut = f.GetSecondLoop()
		}
		companion := Lmfkrh(f.GetFirstLoop())
		sr.faces = append(sr.faces, companion)
	}

	above := NewSolid()
	below := NewSolid()
	*sr.A = above
	*sr.B = below

	// Reset Mark2 on all faces so MoveFace can traverse correctly.
	// Without this, faces from a previous split retain Mark2=VISITED,
	// causing MoveFace to skip them and mispartition the solid.
	sr.s.SetFaceMarks2(0)

	sr.finishClassify(nf)

	above.Cleanup()
	below.Cleanup()

	sr.s.Cleanup()
}

func (sr *SplitRecord) finishClassify(nf int) {
	// Move face pairs: faces[i] → Above, faces[nf+i] → Below.
	for i := 0; i < nf; i++ {
		sr.s.MoveFace(sr.faces[i], *sr.A)
		sr.s.MoveFace(sr.faces[nf+i], *sr.B)
	}

	// Move remaining un-touched faces based on vertex distance.
	for sr.s.GetFirstFace() != nil {
		f := sr.s.GetFirstFace()
		// Reset visit marks to allow MoveFace recursion.
		sr.s.SetFaceMarks2(UNKNOWN)

		// Find a vertex that is definitively ABOVE or BELOW the plane.
		// The first vertex may lie exactly ON the plane (e.g. an
		// intersection vertex from a previous split), giving no signal.
		side := 0
		for _, l := range f.Loops {
			done := false
			l.ForEachHe(func(he *HalfEdge) bool {
				d := sr.sp.GetDistance(he.Vertex.Loc)
				s := floatCompare(d, 0)
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

		if side == BELOW {
			sr.s.MoveFace(f, *sr.A)
		} else {
			sr.s.MoveFace(f, *sr.B)
		}
	}
}

// complete runs post-split cleanup on both result solids.
func (sr *SplitRecord) complete() {
	(*sr.A).Renumber()
	(*sr.B).Renumber()
	(*sr.A).RemoveCoplaneColine()
	(*sr.B).RemoveCoplaneColine()
}

// ---------------------------------------------------------------------------
// Split — top-level entry point.
// ---------------------------------------------------------------------------

// Split divides solid s along the given plane. Returns true if the solid
// was actually split (edges crossed the plane). On success, *above and
// *below point to the two resulting solids. The original solid s should
// not be used after a successful split.
func (s *Solid) Split(sp Plane) (above, below *Solid, ok bool) {
	s.CalcPlaneEquations()

	rec := NewSplitRecord()
	rec.reset(sp, s, &above, &below)

	s.SetFaceMarks(UNKNOWN)
	s.SetVertexMarks(UNKNOWN)

	rec.generate()

	if !rec.classify() {
		return nil, nil, false
	}

	rec.connect()
	rec.finish()
	rec.complete()

	return above, below, true
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const bigEps = 0.00001

// floatCompare returns -1, 0, or 1 comparing f to 0 within tolerance.
func floatCompare(f, zero float32) int {
	if f < -bigEps {
		return -1
	}
	if f > bigEps {
		return 1
	}
	return 0
}

// SetFaceMarks2 sets Mark2 on all faces.
func (s *Solid) SetFaceMarks2(m uint32) {
	for f := s.Faces; f != nil; f = f.Next {
		f.Mark2 = m
	}
}
