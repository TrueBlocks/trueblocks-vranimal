package solid

// Ported from vraniml/src/solid/algorithms/bool/boolclassify.cpp,
// boolvertexclassify.cpp, boolfaceclassify.cpp, boolsector.cpp.
//
// Phase 2 of the boolean pipeline: classify intersection neighborhoods
// as IN/OUT/ON relative to the other solid, then insert null edges at
// crossings.

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// IN and OUT classification constants for boolean ops.
// IN means inside the other solid, OUT means outside.
const (
	IN  = -1
	OUT = 1
)

// BAD_CROSSING marks a sector whose classification is inconsistent.
const BAD_CROSSING = 12

// ---------------------------------------------------------------------------
// MakeRing — create a degenerate ring (inner loop) at a vertex on a face.
// Ported from vrFace::MakeRing in face.cpp.
// ---------------------------------------------------------------------------

func (f *Face) MakeRing(v vec.SFVec3f) *Edge {
	head := f.LoopOut.GetFirstHe()
	s := f.Solid

	// Create first null vertex at head
	Lmev2(head, head, v)
	he1 := head.Prev

	// Create second null vertex
	Lmev2(he1, he1, v)
	he2 := he1.Prev

	// Kill edge, make ring → creates inner loop
	Lkemr(he1.GetMate())
	_ = s

	return he2.Edge
}

// ---------------------------------------------------------------------------
// VF Neighborhood — sector around a vertex classified against a face.
// ---------------------------------------------------------------------------

type vfNeighborhood struct {
	sector     *HalfEdge
	cl         int  // IN, OUT, or ON
	intersects bool // whether this sector contains an intersection
}

// ---------------------------------------------------------------------------
// vfClassifyRecord — per vertex-face classification state.
// ---------------------------------------------------------------------------

type vfClassifyRecord struct {
	sp   Plane
	v    *Vertex
	f    *Face
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
	cr.sp = Plane{Normal: cr.f.Normal, D: cr.f.D}
	cr.nbrs = cr.nbrs[:0]
}

// vfGetNeighborhood walks around vertex v collecting edge classifications
// relative to the face's plane.
func (cr *vfClassifyRecord) vfGetNeighborhood() {
	cr.nbrs = cr.nbrs[:0]
	he := cr.v.He
	for {
		vNext := he.Next.Vertex
		cl := floatCompare(cr.sp.GetDistance(vNext.Loc), 0)

		if he.IsWide(false) {
			pCl := cl
			bisector := he.Bisect()
			cl = floatCompare(cr.sp.GetDistance(bisector), 0)
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

// vfReclassifyOnSectors handles coplanar sectors.
func (cr *vfClassifyRecord) vfReclassifyOnSectors() {
	n := len(cr.nbrs)
	epsSq := float64(1e-5 * 1e-5)
	spNormal := cr.sp.Normal

	for i := 0; i < n; i++ {
		facNormal := cr.nbrs[i].sector.GetFaceNormal()
		c := facNormal.Cross(spNormal)
		d := c.Dot(c)

		if floatCompare(d, epsSq) <= 0 {
			prv := (i + n - 1) % n
			dotN := facNormal.Dot(spNormal)

			var newCl int
			if floatCompare(dotN, 0) == 1 {
				// Same direction normals
				if cr.isB {
					newCl = boolIf(cr.op == BoolUnion, IN, OUT)
				} else {
					newCl = boolIf(cr.op == BoolUnion, OUT, IN)
				}
			} else {
				// Opposite direction normals
				if cr.isB {
					newCl = boolIf(cr.op == BoolUnion, IN, OUT)
				} else {
					newCl = boolIf(cr.op == BoolUnion, IN, OUT)
				}
			}
			cr.nbrs[i].cl = newCl
			cr.nbrs[prv].cl = newCl
		}
	}
}

// vfReclassifyOnEdges resolves remaining ON neighborhoods.
func (cr *vfClassifyRecord) vfReclassifyOnEdges() {
	n := len(cr.nbrs)
	for i := 0; i < n; i++ {
		if cr.nbrs[i].cl == ON {
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

// vfShiftDown rotates the neighborhood array one position.
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

// vfPrepareNextCrossing finds the next IN→OUT crossing.
func (cr *vfClassifyRecord) vfPrepareNextCrossing() bool {
	n := len(cr.nbrs)
	// Mark entries as intersecting based on local context
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

	// Remove non-intersecting entries
	j := 0
	for k := 0; k < n; k++ {
		if cr.nbrs[k].intersects {
			cr.nbrs[j] = cr.nbrs[k]
			j++
		}
	}
	cr.nbrs = cr.nbrs[:j]
	n = j

	// Rotate to first IN→OUT crossing
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

// isOnOuterLoop checks if a vertex is on the outer loop of its face.
func isOnOuterLoop(v *Vertex) bool {
	f := v.He.GetFace()
	if f == nil || f.LoopOut == nil {
		return true
	}
	found := false
	f.LoopOut.ForEachHe(func(he *HalfEdge) bool {
		if he.Vertex == v {
			found = true
			return false
		}
		return true
	})
	return found
}

// vfFindNextCrossing finds the next pair of half-edges at an IN→OUT crossing.
func (cr *vfClassifyRecord) vfFindNextCrossing() (hInOut, hOutIn *HalfEdge) {
	n := len(cr.nbrs)
	if n < 2 {
		return nil, nil
	}

	nxt := (1 + 1) % n

	if n > 2 && cr.nbrs[nxt].cl == IN {
		// IN-OUT-IN case
		hInOut = cr.nbrs[0].sector.GetMate().Next
		hOutIn = cr.nbrs[1].sector.GetMate().Next
		cr.nbrs[1].cl = IN
	} else if n > 2 {
		// IN-OUT-OUT case
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
		// Simple 2-entry case
		hInOut = cr.nbrs[0].sector.GetMate().Next
		hOutIn = cr.nbrs[1].sector.GetMate().Next
		cr.nbrs[1].cl = IN
	}

	cr.vfShiftDown()
	cr.vfShiftDown()

	return hInOut, hOutIn
}

// vfSeperate inserts null edges at a VF crossing.
func (cr *vfClassifyRecord) vfSeperate(hInOut, hOutIn *HalfEdge) {
	s := hInOut.GetSolid()
	outer := isOnOuterLoop(cr.v)

	// Create null edge on the vertex's solid
	if cr.br.NoVV {
		Lmev2(hInOut, hOutIn, hInOut.Vertex.Loc)
	} else {
		Lmev2(hOutIn, hInOut, hInOut.Vertex.Loc)
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
	_ = s

	// Create ring on the face (from the other solid)
	e2 := cr.f.MakeRing(hInOut.Vertex.Loc)
	if !outer {
		e2.SwapHes()
	}

	if !cr.isB {
		cr.br.EdgesB = append(cr.br.EdgesB, e2)
	} else {
		cr.br.EdgesA = append(cr.br.EdgesA, e2)
	}
}

// vfInsertNullEdges loops through crossings inserting null edges.
func (cr *vfClassifyRecord) vfInsertNullEdges() {
	for cr.vfPrepareNextCrossing() {
		hInOut, hOutIn := cr.vfFindNextCrossing()
		if hInOut != nil && hOutIn != nil {
			cr.vfSeperate(hInOut, hOutIn)
		}
	}
}

// vfClassify performs vertex-face classification for record i.
func (cr *vfClassifyRecord) vfClassify(i int, isB bool) {
	cr.setVertexFace(i, isB)
	cr.vfGetNeighborhood()
	cr.vfReclassifyOnSectors()
	cr.vfReclassifyOnEdges()
	cr.vfInsertNullEdges()
}

// ---------------------------------------------------------------------------
// VV Neighborhood — sector for vertex-vertex classification.
// ---------------------------------------------------------------------------

type vvNeighborhood struct {
	he         *HalfEdge
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
	intersects int // TRUE=1, FALSE=0, BAD_CROSSING=12
	onFace     int // 0=none, 1=SAME, 2=OPP
}

const (
	SAME_FACE = 1
	OPP_FACE  = 2
)

// ---------------------------------------------------------------------------
// vvClassifyRecord — per vertex-vertex classification state.
// ---------------------------------------------------------------------------

type vvClassifyRecord struct {
	sp      Plane
	v       *Vertex // from solid A
	v2      *Vertex // from solid B
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

// preprocess builds the neighborhood for a vertex.
func (cr *vvClassifyRecord) preprocess(v *Vertex, isB bool) {
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

func (cr *vvClassifyRecord) addNeighbor(he *HalfEdge, vP, vN vec.SFVec3f, isB bool, fromBisect int) {
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

// getNeighborhood creates sectors for all pairs of A and B neighborhoods.
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
		clNextA:    floatCompare(d1, 0),
		clPrevA:    floatCompare(d2, 0),
		clNextB:    floatCompare(d3, 0),
		clPrevB:    floatCompare(d4, 0),
		intersects: 1, // TRUE
	}
	cr.sectors = append(cr.sectors, s)
}

// checkIntersectFlag marks a sector as non-intersecting if classifications are equal and non-ON.
func checkIntersectFlag(s *vvSector) {
	if s.clNextA == s.clPrevA && s.clNextA != ON {
		s.intersects = 0
	}
	if s.clNextB == s.clPrevB && s.clNextB != ON {
		s.intersects = 0
	}
}

// reclassifyOnSectors handles coplanar VV sectors (all ON).
func (cr *vvClassifyRecord) reclassifyOnSectors() {
	for i := range cr.sectors {
		if cr.sectors[i].clNextA == ON && cr.sectors[i].clPrevA == ON &&
			cr.sectors[i].clNextB == ON && cr.sectors[i].clPrevB == ON {

			indexA := cr.sectors[i].indexA
			indexB := cr.sectors[i].indexB

			normA := cr.nhA[indexA].he.GetFaceNormal()
			normB := cr.nhB[indexB].he.GetFaceNormal()

			var newAcl, newBcl int
			if normA.Eq(normB) {
				newAcl = boolIf(cr.op == BoolUnion, OUT, IN)
				newBcl = boolIf(cr.op == BoolUnion, IN, OUT)
			} else {
				newAcl = boolIf(cr.op == BoolUnion, IN, OUT)
				newBcl = boolIf(cr.op == BoolUnion, IN, OUT)
			}

			for j := range cr.sectors {
				if cr.sectors[j].intersects != 0 {
					if cr.sectors[j].indexA == indexA {
						if cr.sectors[j].clPrevB == ON {
							cr.sectors[j].clPrevB = newBcl
						}
						if cr.sectors[j].clNextB == ON {
							cr.sectors[j].clNextB = newBcl
						}
					}
					if cr.sectors[j].indexB == indexB {
						if cr.sectors[j].clPrevA == ON {
							cr.sectors[j].clPrevA = newAcl
						}
						if cr.sectors[j].clNextA == ON {
							cr.sectors[j].clNextA = newAcl
						}
					}
					checkIntersectFlag(&cr.sectors[j])
				}
			}
		}
	}
}

// reclassifyDoublyOnEdges handles doubly-ON edge sectors.
func (cr *vvClassifyRecord) reclassifyDoublyOnEdges(i int) {
	s := &cr.sectors[i]
	if (s.clNextA == ON && s.clNextB == ON) || (s.clPrevA == ON && s.clPrevB == ON) {
		newAcl := boolIf(cr.op == BoolUnion, OUT, IN)
		newBcl := boolIf(cr.op == BoolUnion, IN, OUT)
		indexA := s.indexA
		indexB := s.indexB

		for j := range cr.sectors {
			if cr.sectors[j].intersects != 0 {
				if cr.sectors[j].indexA == indexA {
					if cr.sectors[j].clPrevB == ON {
						cr.sectors[j].clPrevB = newBcl
					}
					if cr.sectors[j].clNextB == ON {
						cr.sectors[j].clNextB = newBcl
					}
				}
				if cr.sectors[j].indexB == indexB {
					if cr.sectors[j].clPrevA == ON {
						cr.sectors[j].clPrevA = newAcl
					}
					if cr.sectors[j].clNextA == ON {
						cr.sectors[j].clNextA = newAcl
					}
				}
				checkIntersectFlag(&cr.sectors[j])
			}
		}
	}
}

// reclassifySinglyOnEdges handles singly-ON edge sectors.
func (cr *vvClassifyRecord) reclassifySinglyOnEdges(i int) {
	s := &cr.sectors[i]
	indexA := s.indexA
	indexB := s.indexB

	newAcl := boolIf(cr.op == BoolUnion, OUT, IN)
	newBcl := boolIf(cr.op == BoolUnion, OUT, IN)

	if s.clPrevA == ON || s.clNextA == ON || s.clPrevB == ON || s.clNextB == ON {
		for j := range cr.sectors {
			if cr.sectors[j].intersects != 0 {
				if cr.sectors[j].indexA == indexA {
					if cr.sectors[j].clPrevB == ON {
						cr.sectors[j].clPrevB = newBcl
					}
					if cr.sectors[j].clNextB == ON {
						cr.sectors[j].clNextB = newBcl
					}
				}
				if cr.sectors[j].indexB == indexB {
					if cr.sectors[j].clPrevA == ON {
						cr.sectors[j].clPrevA = newAcl
					}
					if cr.sectors[j].clNextA == ON {
						cr.sectors[j].clNextA = newAcl
					}
				}
				checkIntersectFlag(&cr.sectors[j])
			}
		}
	}
}

// reclassifyOnEdges resolves remaining ON classifications.
func (cr *vvClassifyRecord) reclassifyOnEdges() {
	for i := range cr.sectors {
		if cr.sectors[i].intersects != 0 {
			cr.reclassifyDoublyOnEdges(i)
			cr.reclassifySinglyOnEdges(i)
		}
	}
}

// vvShiftDown rotates the sectors array one position.
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

// separate1 creates a null edge spanning from→to.
func separate1(from, to *HalfEdge) *Edge {
	s := to.GetSolid()
	v := to.Vertex.Loc
	Lmev2(from, to, v)
	_ = s
	return from.Prev.Edge
}

// separate2 creates a null self-loop edge at he.
func separate2(he *HalfEdge) *Edge {
	s := he.GetSolid()
	v := he.Vertex.Loc
	Lmev2(he, he, v)
	_ = s
	return he.Prev.Edge
}

// prepareNextCrossing finds the next valid VV crossing.
func (cr *vvClassifyRecord) prepareNextCrossing() bool {
	// Remove non-intersecting sectors
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

	// For UNION, remove BAD_CROSSINGs when there are multiple sectors
	if n > 1 && cr.op == BoolUnion {
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

	// Ensure first sector has clNextA == OUT
	if len(cr.sectors) > 0 && cr.sectors[0].clNextA != OUT {
		cr.vvShiftDown()
	}

	// Handle single-sector case
	if n == 1 {
		s := &cr.sectors[0]
		normA := cr.nhA[s.indexA].he.GetMate().GetFaceNormal()
		normB := cr.nhB[s.indexB].he.GetMate().GetFaceNormal()

		s.onFace = 0

		if normA.Eq(normB) {
			s.onFace = SAME_FACE
		} else if normA.Eq(normB.Negate()) {
			s.onFace = OPP_FACE
			if cr.op != BoolUnion && (s.intersects == 2 || s.intersects == 4) {
				cr.sectors = cr.sectors[:0]
				return false
			}
		}
	} else if n > 1 {
		// Remove BAD_CROSSINGs
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

// findNextCrossing picks half-edges for the next VV crossing.
func (cr *vvClassifyRecord) findNextCrossing() (aHead, aTail, bHead, bTail *HalfEdge) {
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

		// Single sector — determine tail based on 180° edges
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

// vvSeperate inserts null edges for a VV crossing.
func (cr *vvClassifyRecord) vvSeperate(aHead, aTail, bHead, bTail *HalfEdge) {
	var e1, e2 *Edge

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

// insertNullEdges loops through VV crossings.
func (cr *vvClassifyRecord) insertNullEdges() {
	for cr.prepareNextCrossing() {
		aHead, aTail, bHead, bTail := cr.findNextCrossing()
		if aHead != nil && aTail != nil && bHead != nil && bTail != nil {
			cr.vvSeperate(aHead, aTail, bHead, bTail)
		}
	}
}

// vvClassify performs vertex-vertex classification for record i.
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

// ---------------------------------------------------------------------------
// Classify — Phase 2 entry point.
// ---------------------------------------------------------------------------

// Classify classifies all intersection records from Generate, inserting null
// edges at crossings. Returns true if null edges were created (>2 means
// a real boolean intersection exists).
func (br *BoolopRecord) Classify() bool {
	br.EdgesA = br.EdgesA[:0]
	br.EdgesB = br.EdgesB[:0]

	br.NoVV = len(br.VertsV) == 0

	// Vertex-Face classification
	fRec := &vfClassifyRecord{}
	fRec.reset(br)
	for i := 0; i < len(br.VertsA); i++ {
		fRec.vfClassify(i, false)
	}
	for i := 0; i < len(br.VertsB); i++ {
		fRec.vfClassify(i, true)
	}

	// Vertex-Vertex classification
	vRec := &vvClassifyRecord{}
	vRec.reset(br)
	for i := 0; i < len(br.VertsV); i++ {
		vRec.vvClassify(i)
	}

	// Clear intersection records
	br.VertsA = br.VertsA[:0]
	br.VertsB = br.VertsB[:0]
	br.VertsV = br.VertsV[:0]

	return len(br.EdgesA) > 2
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func boolIf(cond bool, trueVal, falseVal int) int {
	if cond {
		return trueVal
	}
	return falseVal
}
