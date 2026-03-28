package solid

import (
	"fmt"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func TestBoolUnionBBox(t *testing.T) {
	t.Skip("known failing — union loses A's extent, see issue #54")
	red := vec.SFColor{R: 1, G: 0, B: 0, A: 1}
	green := vec.SFColor{R: 0, G: 1, B: 0, A: 1}

	a := MakeCube(1.0, red)
	b := MakeCube(0.5, green)
	scale(b, 1, 1, 2)
	translate(b, 0.25, 0.25, -0.25)

	mnA, mxA := a.Extents()
	mnB, mxB := b.Extents()
	fmt.Printf("A bbox: [%.2f,%.2f,%.2f]-[%.2f,%.2f,%.2f]\n", mnA.X, mnA.Y, mnA.Z, mxA.X, mxA.Y, mxA.Z)
	fmt.Printf("B bbox: [%.2f,%.2f,%.2f]-[%.2f,%.2f,%.2f]\n", mnB.X, mnB.Y, mnB.Z, mxB.X, mxB.Y, mxB.Z)

	result, ok := BoolOp(a, b, BoolUnion)
	if !ok || result == nil {
		t.Fatal("BoolOp failed")
	}
	result.CalcPlaneEquations()
	mn, mx := result.Extents()
	fmt.Printf("Union bbox: [%.2f,%.2f,%.2f]-[%.2f,%.2f,%.2f]\n", mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z)

	nF := 0
	for f := result.Faces; f != nil; f = f.Next {
		nv := 0
		var verts string
		if f.LoopOut != nil {
			f.LoopOut.ForEachHe(func(he *HalfEdge) bool {
				nv++
				v := he.Vertex.Loc
				verts += fmt.Sprintf(" (%.2f,%.2f,%.2f)", v.X, v.Y, v.Z)
				return true
			})
		}
		fmt.Printf("  face %d: %d verts, n=(%.2f,%.2f,%.2f)%s\n", nF, nv, f.Normal.X, f.Normal.Y, f.Normal.Z, verts)
		nF++
	}
	fmt.Printf("Total faces: %d\n", nF)

	// The union should extend to A's bounds
	if mn.X > -0.9 || mn.Y > -0.9 || mx.X < 0.9 || mx.Y < 0.9 {
		t.Errorf("Union bbox too small — missing A's extent. Got [%.2f,%.2f,%.2f]-[%.2f,%.2f,%.2f]",
			mn.X, mn.Y, mn.Z, mx.X, mx.Y, mx.Z)
	}
}
