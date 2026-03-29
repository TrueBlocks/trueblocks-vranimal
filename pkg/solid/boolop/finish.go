package boolop

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
)

func (br *BoolopRecord) Finish() {
	nFacesA := len(br.FacesA)
	nFacesB := len(br.FacesB)
	if nFacesA == 0 {
		switch br.Op {
		case base.BoolUnion:
			br.A.Merge(br.B)
			br.Result = br.A
		case base.BoolIntersection:
			br.Result = nil
		case base.BoolDifference:
			br.Result = br.A
		}
		return
	}
	mirrorsA := make([]*base.Face, nFacesA)
	mirrorsB := make([]*base.Face, nFacesB)
	for i := 0; i < nFacesA; i++ {
		fA := br.FacesA[i]
		if len(fA.Loops) >= 2 && fA.Loops[1] == fA.LoopOut {
			fA.LoopOut = fA.Loops[0]
		}
		sl := fA.GetSecondLoop()
		if sl != nil {
			mirrorsA[i], _ = euler.Lmfkrh(sl)
		}
		fB := br.FacesB[i]
		if len(fB.Loops) >= 2 && fB.Loops[1] == fB.LoopOut {
			fB.LoopOut = fB.Loops[0]
		}
		sl = fB.GetSecondLoop()
		if sl != nil {
			mirrorsB[i], _ = euler.Lmfkrh(sl)
		}
		// When there are no vertex-vertex hits, the face/mirror assignment
		// from Connect is inverted. For Union this only applies when the
		// input was perturbed (degenerate case); non-perturbed Union with
		// asymmetric VF (e.g. B passes through A) is already correct.
		if br.NoVV && (br.Op != base.BoolUnion || br.Perturbed) {
			br.FacesA[i], mirrorsA[i] = mirrorsA[i], br.FacesA[i]
			br.FacesB[i], mirrorsB[i] = mirrorsB[i], br.FacesB[i]
		}
	}
	br.Result = base.NewSolid()
	switch br.Op {
	case base.BoolUnion:
		br.A.ClearFaceMarks2()
		br.B.ClearFaceMarks2()
		for i := 0; i < nFacesA; i++ {
			algorithms.MoveFace(br.FacesA[i], br.Result)
			algorithms.MoveFace(br.FacesB[i], br.Result)
		}
		br.Result.Cleanup()
		for i := 0; i < nFacesA; i++ {
			_ = euler.Lkfmrh(br.FacesB[i], br.FacesA[i])
			algorithms.LoopGlue(br.FacesA[i])
		}
	case base.BoolIntersection:
		br.A.ClearFaceMarks2()
		br.B.ClearFaceMarks2()
		for i := 0; i < nFacesA; i++ {
			algorithms.MoveFace(mirrorsA[i], br.Result)
			algorithms.MoveFace(mirrorsB[i], br.Result)
		}
		br.Result.Cleanup()
		for i := 0; i < nFacesA; i++ {
			_ = euler.Lkfmrh(mirrorsB[i], mirrorsA[i])
			algorithms.LoopGlue(mirrorsA[i])
		}
	case base.BoolDifference:
		tmpSolid := base.NewSolid()
		br.A.ClearFaceMarks2()
		br.B.ClearFaceMarks2()
		for i := 0; i < nFacesA; i++ {
			algorithms.MoveFace(br.FacesA[i], br.Result)
			algorithms.MoveFace(mirrorsB[i], tmpSolid)
		}
		tmpSolid.Revert()
		br.Result.Merge(tmpSolid)
		br.Result.Cleanup()
		for i := 0; i < nFacesA; i++ {
			_ = euler.Lkfmrh(mirrorsB[i], br.FacesA[i])
			algorithms.LoopGlue(br.FacesA[i])
		}
	}
	br.A.Cleanup()
	br.B.Cleanup()
	if br.Result != nil {
		for f := br.Result.Faces; f != nil; f = f.Next {
			if f.LoopOut == nil && len(f.Loops) > 0 {
				f.LoopOut = f.Loops[0]
			}
		}
	}
}
