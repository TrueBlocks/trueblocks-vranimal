package solid

// Ported from vraniml/src/solid/algorithms/bool/boolconnect.cpp
//
// Phase 3 of the boolean pipeline: sort null edges, match loose ends,
// join and cut to build the connectivity between the two solids.

// ---------------------------------------------------------------------------
// nullEdgeInfo — metadata for each null edge pair used in SpecialCase.
// ---------------------------------------------------------------------------

type nullEdgeInfo struct {
	hes [4]*HalfEdge
	cl  [4]bool // whether the he's loop is the outer loop
}

// ---------------------------------------------------------------------------
// Connect — Phase 3 entry point.
// ---------------------------------------------------------------------------

// Connect matches null edges produced by Classify, joining and cutting them
// to produce the paired faces needed by Finish.
func (br *BoolopRecord) Connect() {
	br.FacesA = br.FacesA[:0]
	br.FacesB = br.FacesB[:0]

	nEndsA := 0
	nEndsB := 0
	endsA := make([]*HalfEdge, 0, len(br.EdgesA))
	endsB := make([]*HalfEdge, 0, len(br.EdgesA))

	br.sortNullEdges()

	// Build metadata table for SpecialCase
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

	// getThingCl looks up the metadata for SpecialCase.
	getThingCl := func(he *HalfEdge) bool {
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

	// specialCase determines the swap flag for Join.
	specialCase := func(he1, he2 *HalfEdge) bool {
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

	// canJoin searches the loose ends for a matching pair.
	canJoin := func(hea, heb *HalfEdge) (*HalfEdge, *HalfEdge) {
		hea.SetMark(NOT_LOOSE)
		heb.SetMark(NOT_LOOSE)

		for i := 0; i < nEndsA; i++ {
			if hea.IsNeighbor(endsA[i]) && heb.IsNeighbor(endsB[i]) {
				retA := endsA[i]
				retB := endsB[i]

				// Compact: shift remaining entries down
				for j := i + 1; j < nEndsA; j++ {
					endsA[j-1] = endsA[j]
					endsB[j-1] = endsB[j]
				}
				nEndsA--
				nEndsB--
				endsA = endsA[:nEndsA]
				endsB = endsB[:nEndsB]

				retA.SetMark(NOT_LOOSE)
				retB.SetMark(NOT_LOOSE)
				return retA, retB
			}
		}

		// No match — add as loose ends
		hea.SetMark(LOOSE)
		heb.SetMark(LOOSE)
		endsA = append(endsA, hea)
		endsB = append(endsB, heb)
		nEndsA++
		nEndsB++
		return nil, nil
	}

	// Process each null edge pair sequentially.
	for idx := 0; idx < len(br.EdgesA); idx++ {
		edgeA := br.EdgesA[idx]
		edgeB := br.EdgesB[idx]

		outA := edgeA.He1
		inB := edgeB.He2
		inA := edgeA.He2
		outB := edgeB.He1

		var joinInA, joinOutA, joinInB, joinOutB *HalfEdge

		joinInA, joinOutB = canJoin(outA, inB)
		if joinInA != nil {
			br.A.Join(joinInA, outA, specialCase(joinInA, outA))
			if !joinInA.GetMate().Marked(LOOSE) {
				f := br.A.Cut(joinInA, true)
				if f != nil {
					br.FacesA = append(br.FacesA, f)
				}
			}

			br.B.Join(joinOutB, inB, false)
			if !joinOutB.GetMate().Marked(LOOSE) {
				f := br.B.Cut(joinOutB, true)
				if f != nil {
					br.FacesB = append(br.FacesB, f)
				}
			}
		}

		joinOutA, joinInB = canJoin(inA, outB)
		if joinOutA != nil {
			br.A.Join(joinOutA, inA, false)
			if !joinOutA.GetMate().Marked(LOOSE) {
				f := br.A.Cut(joinOutA, true)
				if f != nil {
					br.FacesA = append(br.FacesA, f)
				}
			}

			br.B.Join(joinInB, outB, specialCase(joinInB, outB))
			if !joinInB.GetMate().Marked(LOOSE) {
				f := br.B.Cut(joinInB, true)
				if f != nil {
					br.FacesB = append(br.FacesB, f)
				}
			}
		}

		if joinInA != nil && joinInB != nil && joinOutA != nil && joinOutB != nil {
			f := br.A.Cut(outA, true)
			if f != nil {
				br.FacesA = append(br.FacesA, f)
			}
			f = br.B.Cut(outB, true)
			if f != nil {
				br.FacesB = append(br.FacesB, f)
			}
		}
	}

	// Clear edge arrays
	br.EdgesA = br.EdgesA[:0]
	br.EdgesB = br.EdgesB[:0]
}

// sortNullEdges sorts EdgesA and EdgesB in parallel by vertex coordinate
// (lexicographic: x, then y, then z with bigEps tolerance).
func (br *BoolopRecord) sortNullEdges() {
	n := len(br.EdgesA)
	// Bubble sort (matching C++ implementation)
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

	// Assign indices
	for i := 0; i < n; i++ {
		br.EdgesA[i].Index = uint64(i)
		br.EdgesB[i].Index = uint64(i)
	}
}

// vertexGreaterThan returns true if v1 > v2 in lexicographic order (x, y, z)
// using bigEps tolerance.
func vertexGreaterThan(v1, v2 *Vertex) bool {
	res := floatCompare(v1.Loc.X, v2.Loc.X)
	if res == -1 {
		return false
	}
	if res == 0 {
		res = floatCompare(v1.Loc.Y, v2.Loc.Y)
		if res == -1 {
			return false
		}
		if res == 0 {
			res = floatCompare(v1.Loc.Z, v2.Loc.Z)
			if res == -1 {
				return false
			}
		}
	}
	return true
}
