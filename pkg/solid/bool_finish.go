package solid

// Ported from vraniml/src/solid/algorithms/bool/boolfinish.cpp
//
// Phase 4 of the boolean pipeline: construct the result solid by selecting
// faces based on the boolean operation type.

// ---------------------------------------------------------------------------
// Finish — Phase 4 entry point.
// ---------------------------------------------------------------------------

// Finish builds the result solid from the paired faces produced by Connect.
// It creates mirror faces via Lmfkrh, then selects originals or mirrors
// depending on the operation type, merges them, and glues null faces.
func (br *BoolopRecord) Finish() {
	nFacesA := len(br.FacesA)
	nFacesB := len(br.FacesB)

	if nFacesA == 0 {
		// Degenerate: no paired faces, merge A and B directly
		br.A.Merge(br.B)
		br.Result = br.A
		return
	}

	// Create mirror faces: for each paired face, promote the second loop
	// into a new face via Lmfkrh.
	mirrorsA := make([]*Face, nFacesA)
	mirrorsB := make([]*Face, nFacesB)

	for i := 0; i < nFacesA; i++ {
		// Ensure GetSecondLoop is not the outer loop — if it is,
		// swap LoopOut to the first loop so Lmfkrh gets an inner loop.
		// (Matches C++ commented-out safety check in boolfinish.cpp.)
		fA := br.FacesA[i]
		if len(fA.Loops) >= 2 && fA.Loops[1] == fA.LoopOut {
			fA.LoopOut = fA.Loops[0]
		}
		sl := fA.GetSecondLoop()
		if sl != nil {
			mirrorsA[i] = Lmfkrh(sl)
		}

		fB := br.FacesB[i]
		if len(fB.Loops) >= 2 && fB.Loops[1] == fB.LoopOut {
			fB.LoopOut = fB.Loops[0]
		}
		sl = fB.GetSecondLoop()
		if sl != nil {
			mirrorsB[i] = Lmfkrh(sl)
		}

		// noVV swap: swap original and mirror faces
		// For Union, the swap is counterproductive because Go's Lkfmrh
		// kills its first parameter (opposite of C++), and the swap
		// puts mirrors where originals should be, causing extra faces.
		// Testing shows Intersection/Difference need the swap to pass.
		if br.NoVV && br.Op != BoolUnion {
			br.FacesA[i], mirrorsA[i] = mirrorsA[i], br.FacesA[i]
			br.FacesB[i], mirrorsB[i] = mirrorsB[i], br.FacesB[i]
		}
	}

	br.Result = NewSolid()

	switch br.Op {
	case BoolUnion:
		// Move original faces from both solids
		br.A.ClearFaceMarks2()
		br.B.ClearFaceMarks2()
		for i := 0; i < nFacesA; i++ {
			br.A.MoveFace(br.FacesA[i], br.Result)
			br.B.MoveFace(br.FacesB[i], br.Result)
		}
		br.Result.Cleanup()
		// Glue null faces: kill B, keep A (C++ lkfmrh keeps first arg)
		for i := 0; i < nFacesA; i++ {
			Lkfmrh(br.FacesB[i], br.FacesA[i])
			br.Result.LoopGlue(br.FacesA[i])
		}

	case BoolIntersection:
		// Move mirror faces from both solids
		br.A.ClearFaceMarks2()
		br.B.ClearFaceMarks2()
		for i := 0; i < nFacesA; i++ {
			br.A.MoveFace(mirrorsA[i], br.Result)
			br.B.MoveFace(mirrorsB[i], br.Result)
		}
		br.Result.Cleanup()
		// Glue null faces
		for i := 0; i < nFacesA; i++ {
			Lkfmrh(mirrorsA[i], mirrorsB[i])
			br.Result.LoopGlue(mirrorsA[i])
		}

	case BoolDifference:
		// Move A originals and B mirrors into separate solids
		tmpSolid := NewSolid()
		br.A.ClearFaceMarks2()
		br.B.ClearFaceMarks2()
		for i := 0; i < nFacesA; i++ {
			br.A.MoveFace(br.FacesA[i], br.Result)
			br.B.MoveFace(mirrorsB[i], tmpSolid)
		}
		// Revert B's topology (flip normals)
		tmpSolid.Revert()
		br.Result.Merge(tmpSolid)
		br.Result.Cleanup()
		// Glue null faces
		for i := 0; i < nFacesA; i++ {
			Lkfmrh(br.FacesA[i], mirrorsB[i])
			br.Result.LoopGlue(br.FacesA[i])
		}
	}

	br.A.Cleanup()
	br.B.Cleanup()

	// Ensure every face in the result has an outer loop.
	// LoopGlue/Lkef chains can leave LoopOut nil when the
	// original outer loop gets merged away.
	if br.Result != nil {
		for f := br.Result.Faces; f != nil; f = f.Next {
			if f.LoopOut == nil && len(f.Loops) > 0 {
				f.LoopOut = f.Loops[0]
			}
		}
	}
}

// ClearFaceMarks2 resets Mark2 on all faces (used before MoveFace).
func (s *Solid) ClearFaceMarks2() {
	for f := s.Faces; f != nil; f = f.Next {
		f.Mark2 = 0
	}
}
