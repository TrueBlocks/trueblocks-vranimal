package algorithms

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
)

func IsCoplanar(f1, f2 *base.Face) bool {
	n1 := f1.Normal
	n2 := f2.Normal
	cross := n1.Cross(n2)
	return cross.Length() < 1e-7
}

func HasCoplanarNeighbor(f *base.Face) (*base.Face, *base.HalfEdge, bool) {
	if f.LoopOut == nil {
		return nil, nil, false
	}
	he := f.LoopOut.GetFirstHe()
	if he == nil {
		return nil, nil, false
	}
	start := he
	for {
		mate := he.GetMate()
		if mate != nil {
			f2 := mate.GetFace()
			if f2 != nil && IsCoplanar(f, f2) && !he.IsNullStrut() {
				return f2, he, true
			}
		}
		he = he.Next
		if he == start {
			break
		}
	}
	return nil, nil, false
}

func RemoveCoplanarFace(f *base.Face, f2 *base.Face, he *base.HalfEdge) {
	if f2 == f {
		_ = euler.Lkemr(he.GetMate())
		AdjustOuterLoop(f)
	} else {
		_ = euler.Lkef(he)
	}

	f.CalcEquation()

	if nf, nhe, ok := HasCoplanarNeighbor(f); ok {
		RemoveCoplanarFace(f, nf, nhe)
	}
}

func RemoveCoplanarFaces(s *base.Solid) {
	for f := s.Faces; f != nil; f = f.Next {
		if f2, he, ok := HasCoplanarNeighbor(f); ok {
			RemoveCoplanarFace(f, f2, he)
		}
	}
}

func AdjustOuterLoop(f *base.Face) {
	if f.NLoops() == 1 {
		f.LoopOut = f.GetFirstLoop()
		return
	}

	drop := base.GetDominantComp(f.Normal.X, f.Normal.Y, f.Normal.Z)

	for changed := true; changed; {
		changed = false
		for _, l := range f.Loops {
			if l == f.LoopOut {
				continue
			}
			v := f.LoopOut.GetFirstHe().Vertex
			rec := NewIntersectRecord()
			if LoopContains(l, v, drop, &rec) {
				f.LoopOut = l
				changed = true
				break
			}
		}
	}
}

func RemoveCoplaneColine(s *base.Solid) {
	s.CalcPlaneEquations()

	for e := s.Edges; e != nil; {
		next := e.Next

		he1 := e.He1
		he2 := e.He2
		if he1 != nil && he2 != nil {
			f1 := he1.GetFace()
			f2 := he2.GetFace()
			if f1 != nil && f2 != nil && IsCoplanar(f1, f2) && f1.Normal.Dot(f2.Normal) > 0 {
				if f1 != f2 {
					_ = euler.Lkef(he1)
					f1.CalcEquation()
					RemoveColinearVerts(f1)
				} else if he1.IsNullStrut() {
					if he1.Next == he2 {
						_ = euler.Lkev(he2)
					} else {
						_ = euler.Lkev(he1)
					}
				}
			}
		}

		e = next
	}

	for f := s.Faces; f != nil; f = f.Next {
		RemoveColinearVerts(f)
	}
}

func LoopHasColinearVerts(l *base.Loop) *base.HalfEdge {
	if l.HalfEdges == nil {
		return nil
	}
	he := l.HalfEdges
	for {
		v1 := he.Prev.Vertex.Loc
		v := he.Vertex.Loc
		v2 := he.Next.Vertex.Loc
		if base.Collinear(v1.X, v1.Y, v1.Z, v.X, v.Y, v.Z, v2.X, v2.Y, v2.Z) {
			return he
		}
		he = he.Next
		if he == l.HalfEdges {
			break
		}
	}
	return nil
}

func FaceHasColinearVerts(f *base.Face) (*base.Loop, *base.HalfEdge) {
	for _, l := range f.Loops {
		if he := LoopHasColinearVerts(l); he != nil {
			return l, he
		}
	}
	return nil, nil
}

func SolidHasColinearVerts(s *base.Solid) bool {
	for f := s.Faces; f != nil; f = f.Next {
		if _, he := FaceHasColinearVerts(f); he != nil {
			return true
		}
	}
	return false
}

func HasDegenerateFaces(s *base.Solid) *base.Face {
	for f := s.Faces; f != nil; f = f.Next {
		if f.IsDegenerate() {
			return f
		}
	}
	return nil
}

func RemoveDegenerateFaces(s *base.Solid) {
	for f := s.Faces; f != nil; {
		next := f.Next
		if f.IsDegenerate() {
			CollapseFace(f)
		}
		f = next
	}
}
