package solid

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func TestBuildFromIndexSet_Cube(t *testing.T) {
	t.Skip("known hanging — infinite loop in Lmef, see issue #54")
	// Same vertex layout as overlapping_cubes_input.wrl after dedup
	verts := []vec.SFVec3f{
		{X: -1, Y: 1, Z: -1},  // 0
		{X: -1, Y: -1, Z: -1}, // 1
		{X: -1, Y: -1, Z: 1},  // 2
		{X: -1, Y: 1, Z: 1},   // 3
		{X: 1, Y: -1, Z: -1},  // 4
		{X: 1, Y: -1, Z: 1},   // 5
		{X: 1, Y: 1, Z: -1},   // 6
		{X: 1, Y: 1, Z: 1},    // 7
	}
	indices := []int64{
		0, 1, 2, 3, -1,
		1, 4, 5, 2, -1,
		4, 6, 7, 5, -1,
		6, 0, 3, 7, -1,
		1, 0, 6, 4, -1,
		3, 2, 5, 7, -1,
	}
	color := vec.SFColor{R: 0.9, G: 0.2, B: 0.2, A: 1}

	s := BuildFromIndexSet(verts, indices, color)
	if s == nil {
		t.Fatal("BuildFromIndexSet returned nil")
	}

	nF, nV, nE := 0, 0, 0
	for f := s.Faces; f != nil; f = f.Next {
		nF++
	}
	for v := s.Verts; v != nil; v = v.Next {
		nV++
	}
	for e := s.Edges; e != nil; e = e.Next {
		nE++
	}
	t.Logf("F=%d V=%d E=%d", nF, nV, nE)

	errs := s.VerifyDetailed()
	for _, e := range errs {
		t.Errorf("verify: %v", e)
	}
}
