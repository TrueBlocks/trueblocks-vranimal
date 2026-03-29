package algorithms

import (
	"fmt"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
)

type VerifyError struct {
	Element string
	Index   uint64
	Message string
}

func (e *VerifyError) Error() string {
	return fmt.Sprintf("%s[%d]: %s", e.Element, e.Index, e.Message)
}

func VerifyDetailed(s *base.Solid) []error {
	var errs []error

	nLoops := 0
	for f := s.Faces; f != nil; f = f.Next {
		nLoops += f.NLoops()
	}
	nHoles := nLoops - s.NFaces()
	lhs := s.NFaces() + s.NVerts() - 2
	rhs := s.NEdges() + nHoles
	if lhs != rhs {
		errs = append(errs, &VerifyError{
			Element: "Solid",
			Index:   0,
			Message: fmt.Sprintf("Euler formula failed: F(%d)+V(%d)-2=%d, E(%d)+H(%d)=%d",
				s.NFaces(), s.NVerts(), lhs, s.NEdges(), nHoles, rhs),
		})
	}

	for f := s.Faces; f != nil; f = f.Next {
		if f.Solid != s {
			errs = append(errs, &VerifyError{"Face", f.Index, "does not point back to solid"})
		}
		if f.LoopOut == nil {
			errs = append(errs, &VerifyError{"Face", f.Index, "has nil outer loop"})
		}
		foundOut := false
		for _, l := range f.Loops {
			if l.Face != f {
				errs = append(errs, &VerifyError{"Face", f.Index, "loop does not point back to face"})
			}
			if l == f.LoopOut {
				foundOut = true
			}
			if l.HalfEdges == nil {
				errs = append(errs, &VerifyError{"Face", f.Index, "loop has no half-edges"})
				continue
			}
			count := 0
			he := l.HalfEdges
			for {
				count++
				if count > 100000 {
					errs = append(errs, &VerifyError{"Face", f.Index, "loop has circular reference bug"})
					break
				}
				if he.Loop != l {
					errs = append(errs, &VerifyError{"Face", f.Index,
						fmt.Sprintf("half-edge %d does not point back to loop", he.GetIndex())})
				}
				if he.Vertex == nil {
					errs = append(errs, &VerifyError{"Face", f.Index,
						fmt.Sprintf("half-edge %d has nil vertex", he.GetIndex())})
				}
				if he.Edge != nil {
					if he.Edge.He1 != he && he.Edge.He2 != he {
						errs = append(errs, &VerifyError{"Face", f.Index,
							fmt.Sprintf("half-edge %d not referenced by its edge", he.GetIndex())})
					}
				}
				he = he.Next
				if he == l.HalfEdges {
					break
				}
			}
		}
		if f.LoopOut != nil && !foundOut {
			errs = append(errs, &VerifyError{"Face", f.Index, "outer loop not in face's loop list"})
		}
	}

	for e := s.Edges; e != nil; e = e.Next {
		if e.He1 != nil {
			if e.He1.Edge != e {
				errs = append(errs, &VerifyError{"Edge", e.Index, "He1 does not point back to edge"})
			}
		}
		if e.He2 != nil {
			if e.He2.Edge != e {
				errs = append(errs, &VerifyError{"Edge", e.Index, "He2 does not point back to edge"})
			}
		}
	}

	for v := s.Verts; v != nil; v = v.Next {
		if v.He != nil && v.He.Vertex != v {
			errs = append(errs, &VerifyError{"Vertex", v.Index,
				"half-edge does not point back to vertex"})
		}
	}

	return errs
}

func VertexFaces(v *base.Vertex) []*base.Face {
	if v.He == nil {
		return nil
	}
	seen := map[*base.Face]bool{}
	var faces []*base.Face
	start := v.He
	he := start
	for {
		f := he.GetFace()
		if f != nil && !seen[f] {
			seen[f] = true
			faces = append(faces, f)
		}
		m := he.GetMate()
		if m == nil {
			break
		}
		he = m.Next
		if he == start {
			break
		}
	}
	return faces
}

func VertexEdges(v *base.Vertex) []*base.Edge {
	if v.He == nil {
		return nil
	}
	seen := map[*base.Edge]bool{}
	var edges []*base.Edge
	start := v.He
	he := start
	for {
		if he.Edge != nil && !seen[he.Edge] {
			seen[he.Edge] = true
			edges = append(edges, he.Edge)
		}
		m := he.GetMate()
		if m == nil {
			break
		}
		he = m.Next
		if he == start {
			break
		}
	}
	return edges
}

func VertexNeighbors(v *base.Vertex) []*base.Vertex {
	if v.He == nil {
		return nil
	}
	seen := map[*base.Vertex]bool{}
	var nbrs []*base.Vertex
	start := v.He
	he := start
	for {
		mv := he.GetMateVertex()
		if mv != nil && !seen[mv] {
			seen[mv] = true
			nbrs = append(nbrs, mv)
		}
		m := he.GetMate()
		if m == nil {
			break
		}
		he = m.Next
		if he == start {
			break
		}
	}
	return nbrs
}

func FaceNeighbors(f *base.Face) []*base.Face {
	seen := map[*base.Face]bool{}
	var nbrs []*base.Face
	for _, l := range f.Loops {
		l.ForEachHe(func(he *base.HalfEdge) bool {
			mf := he.GetMateFace()
			if mf != nil && mf != f && !seen[mf] {
				seen[mf] = true
				nbrs = append(nbrs, mf)
			}
			return true
		})
	}
	return nbrs
}
