package algorithms

import (
	"math"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func Join(he1, he2 *base.HalfEdge, swap bool) {
	oldFace := he1.GetFace()

	var newFace *base.Face
	if he1.Loop == he2.Loop {
		if he1.Prev.Prev != he2 {
			if swap {
				newFace, _, _ = euler.Lmef(he2.Next, he1)
			} else {
				newFace, _, _ = euler.Lmef(he1, he2.Next)
			}
		}
	} else {
		_, _ = euler.Lmekr(he1, he2.Next)
	}

	if he1.Next.Next != he2 {
		_, _, _ = euler.Lmef(he2, he1.Next)
		if newFace != nil && oldFace.GetSecondLoop() != nil {
			for len(oldFace.Loops) > 1 {
				l := oldFace.Loops[len(oldFace.Loops)-1]
				if l == oldFace.LoopOut {
					break
				}
				_ = euler.Lringmv(l, newFace, false)
			}
		}
	}
}

func Cut(he *base.HalfEdge, setOp bool) *base.Face {
	if he.Edge == nil {
		return nil
	}

	he1 := he.Edge.He1
	he2 := he.Edge.He2

	if he1.Loop == he2.Loop {
		oldFace := he1.GetFace()
		if setOp {
			_ = euler.Lkemr(he2)
		} else {
			_ = euler.Lkemr(he1)
		}
		return oldFace
	}

	_ = euler.Lkef(he1)
	return nil
}

func MoveFace(f *base.Face, target *base.Solid) {
	if f == nil || f.Marked2(base.VISITED) {
		return
	}

	f.Mark2 = base.VISITED

	owner := f.Solid
	owner.RemoveFace(f)
	target.AddFace(f)

	for _, l := range f.Loops {
		l.ForEachHe(func(he *base.HalfEdge) bool {
			mf := he.GetMateFace()
			if mf != nil && mf.Solid != target {
				MoveFace(mf, target)
			}
			return true
		})
	}
}

func CollapseFace(f *base.Face) {
	if f == nil || f.GetSecondLoop() != nil {
		return
	}

	he := f.GetFirstHe()
	if he == nil {
		return
	}
	hen := he.Next
	hep := he.Prev

	_ = euler.Lkef(he.GetMate())

	he = hen
	for he != hep {
		hen = he.Next
		_ = euler.Lkev(he)
		he = hen
	}
	_ = euler.Lkev(he)
}

func RemoveColinearVerts(f *base.Face) {
	for _, l := range f.Loops {
		if l.HalfEdges == nil {
			continue
		}

		start := l.HalfEdges
		he := start
		prevV := he.Prev.Vertex.Loc
		v := he.Vertex.Loc

		for {
			nextV := he.Next.Vertex.Loc

			canRemove := he.GetMate() != nil &&
				he.Prev.GetMate() != nil &&
				he.Prev.GetMate() == he.GetMate().Next &&
				base.Collinear(prevV.X, prevV.Y, prevV.Z, v.X, v.Y, v.Z, nextV.X, nextV.Y, nextV.Z)

			if canRemove {
				n := he.Next
				if he == start {
					if he.GetMate() == he.Prev {
						start = he.Prev.Prev
					} else {
						start = he.Prev
					}
				}
				_ = euler.Lkev(he)
				v = nextV
				he = n
			} else {
				he = he.Next
				prevV = v
				v = nextV
			}

			if he == start {
				break
			}
		}
	}
}

func Triangulate(s *base.Solid) {
	for f := s.Faces; f != nil; f = f.Next {
		TriangulateFace(f)
	}
}

func TriangulateFace(f *base.Face) {
	if f.LoopOut == nil {
		return
	}

	for f.LoopOut.NHalfEdges() > 3 {
		he := f.LoopOut.GetFirstHe()
		if he == nil {
			return
		}
		target := he.Next.Next
		_, _, _ = euler.Lmef(he, target)
	}
}

func Twist(s *base.Solid, fn func(float64) float64) {
	for v := s.Verts; v != nil; v = v.Next {
		z := fn(v.Loc.Z)
		cosZ := math.Cos(z)
		sinZ := math.Sin(z)
		x := v.Loc.X
		y := v.Loc.Y
		v.Loc.X = x*cosZ - y*sinZ
		v.Loc.Y = x*sinZ + y*cosZ
	}
}

func Glue(s *base.Solid, s2 *base.Solid, f1, f2 *base.Face) {
	s.Merge(s2)
	_ = euler.Lkfmrh(f2, f1)
	LoopGlue(f1)
}

func LoopGlue(f *base.Face) {
	if f.NLoops() < 2 {
		return
	}

	h1 := f.GetFirstHe()
	h2Loop := f.GetSecondLoop()
	if h2Loop == nil {
		return
	}
	h2 := h2Loop.GetFirstHe()
	if h2 == nil {
		return
	}

	start := h2
	for !h1.Vertex.Loc.Eq(h2.Vertex.Loc) {
		h2 = h2.Next
		if h2 == start {
			return
		}
	}

	_, _ = euler.Lmekr(h1, h2)
	_ = euler.Lkev(h1.Prev)

	limit := 10000
	for h1.Next != h2 {
		h1next := h1.Next
		_, _, _ = euler.Lmef(h1.Next, h1.Prev)
		_ = euler.Lkev(h1.Next)
		_ = euler.Lkef(h1.GetMate())
		h1 = h1next
		limit--
		if limit <= 0 {
			return
		}
	}

	_ = euler.Lkef(h1.GetMate())
}

func Verify(s *base.Solid) bool {
	nLoops := 0
	for f := s.Faces; f != nil; f = f.Next {
		nLoops += f.NLoops()
	}
	nHoles := nLoops - s.NFaces()
	expected := s.NFaces() + s.NVerts() - 2
	got := s.NEdges() + nHoles
	return expected == got
}

func Arc(s *base.Solid, face *base.Face, startVertex *base.Vertex, cx, cy, radius, height, startAngle, endAngle float64, steps int) {
	angle := startAngle * math.Pi / 180.0
	stepSize := (endAngle - startAngle) * math.Pi / (180.0 * float64(steps))

	prevHe := startVertex.He
	for i := 0; i < steps; i++ {
		angle += stepSize
		x := cx + math.Cos(angle)*radius
		y := cy + math.Sin(angle)*radius
		_, _, _ = euler.Lmev(prevHe, vec.SFVec3f{X: x, Y: y, Z: height})
		prevHe = prevHe.Prev
	}
}

func ArcSweep(s *base.Solid, f *base.Face, dirs []vec.SFVec3f) {
	for _, d := range dirs {
		TranslationalSweep(s, f, d)
	}
}

func TranslationalSweep(s *base.Solid, f *base.Face, dir vec.SFVec3f) {
	for _, l := range f.Loops {
		first := l.GetFirstHe()
		if first == nil {
			continue
		}

		scan := first.Next
		v := scan.Vertex
		_, _, _ = euler.Lmev2(scan, scan, v.Loc.Add(dir))

		for scan != first {
			v = scan.Next.Vertex
			_, _, _ = euler.Lmev2(scan.Next, scan.Next, v.Loc.Add(dir))
			_, _, _ = euler.Lmef(scan.Prev, scan.Next.Next)
			scan = scan.Next.GetMate().Next
		}
		_, _, _ = euler.Lmef(scan.Prev, scan.Next.Next)
	}
}

func RemoveColinearVertsSolid(s *base.Solid) {
	for f := s.Faces; f != nil; f = f.Next {
		if f.IsDegenerate() {
			CollapseFace(f)
		} else {
			RemoveColinearVerts(f)
		}
	}
}
