package solid

// Ported from vraniml/src/solid/algorithms/coplane.cpp, colinear.cpp,
// degenerate.cpp, and verify.cpp. These algorithms support boolean and
// split operations by detecting/removing coplanar faces, collinear vertices,
// and degenerate faces, and by verifying structural integrity.

import (
	"fmt"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Coplanarity detection and removal
// ---------------------------------------------------------------------------

// IsCoplanar returns true if two faces are coplanar (their normals'
// cross product is near zero).
func (f *Face) IsCoplanar(f2 *Face) bool {
	n1 := f.Normal
	n2 := f2.Normal
	cross := n1.Cross(n2)
	return cross.Length() < 1e-7
}

// HasCoplanarNeighbor scans the outer loop and returns the first
// neighbor face that is coplanar with f (and not a null strut).
// Returns the neighbor face, the half-edge across which coplanarity
// was found, and whether one was found.
func (f *Face) HasCoplanarNeighbor() (*Face, *HalfEdge, bool) {
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
			if f2 != nil && f.IsCoplanar(f2) && !he.IsNullStrut() {
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

// RemoveCoplanarFace removes a single coplanar neighbor f2, reached
// via half-edge he, by either killing the edge (making a ring) or
// killing the face. It recalculates the plane equation and recurses
// if additional coplanar neighbors remain.
func (f *Face) RemoveCoplanarFace(f2 *Face, he *HalfEdge) {
	if f2 == f {
		// Same face — kill the edge, making an inner ring.
		Lkemr(he.GetMate())
		f.AdjustOuterLoop()
	} else {
		// Different face — kill the face by merging across the edge.
		Lkef(he)
	}

	f.CalcEquation()

	// Recurse if more coplanar neighbors exist.
	if nf, nhe, ok := f.HasCoplanarNeighbor(); ok {
		f.RemoveCoplanarFace(nf, nhe)
	}
}

// RemoveCoplanarFaces scans all faces and removes coplanar adjacencies.
func (s *Solid) RemoveCoplanarFaces() {
	for f := s.Faces; f != nil; f = f.Next {
		if f2, he, ok := f.HasCoplanarNeighbor(); ok {
			f.RemoveCoplanarFace(f2, he)
		}
	}
}

// AdjustOuterLoop ensures that the outer loop (LoopOut) is the geometrically
// outermost loop of the face. After removing coplanar faces, an inner loop
// may actually enclose the former outer loop.
func (f *Face) AdjustOuterLoop() {
	if f.NLoops() == 1 {
		f.LoopOut = f.GetFirstLoop()
		return
	}

	drop := GetDominantComp(f.Normal)

	for changed := true; changed; {
		changed = false
		for _, l := range f.Loops {
			if l == f.LoopOut {
				continue
			}
			// If l contains a vertex of the current outer loop, l is bigger.
			v := f.LoopOut.GetFirstHe().Vertex
			rec := NewIntersectRecord()
			if l.LoopContains(v, drop, &rec) {
				f.LoopOut = l
				changed = true
				break
			}
		}
	}
}

// RemoveCoplaneColine performs a full cleanup pass: removes coplanar faces,
// then removes collinear vertices introduced by the merges.
func (s *Solid) RemoveCoplaneColine() {
	// Recalculate all plane equations first.
	s.CalcPlaneEquations()

	// Remove edges between coplanar faces.
	for e := s.Edges; e != nil; {
		next := e.Next

		he1 := e.He1
		he2 := e.He2
		if he1 != nil && he2 != nil {
			f1 := he1.GetFace()
			f2 := he2.GetFace()
			if f1 != nil && f2 != nil && f1.IsCoplanar(f2) && f1.Normal.Dot(f2.Normal) > 0 {
				if f1 != f2 {
					Lkef(he1)
					f1.CalcEquation()
					f1.RemoveColinearVerts()
				} else if he1.IsNullStrut() {
					if he1.Next == he2 {
						Lkev(he2)
					} else {
						Lkev(he1)
					}
				}
			}
		}

		e = next
	}

	// Final pass: remove any remaining collinear vertices.
	for f := s.Faces; f != nil; f = f.Next {
		f.RemoveColinearVerts()
	}
}

// ---------------------------------------------------------------------------
// Collinear vertex detection (loop / face / solid level)
// ---------------------------------------------------------------------------

// HasColinearVerts returns the first half-edge in the loop whose vertex
// is collinear with its neighbors, or nil if none.
func (l *Loop) HasColinearVerts() *HalfEdge {
	if l.HalfEdges == nil {
		return nil
	}
	he := l.HalfEdges
	for {
		v1 := he.Prev.Vertex.Loc
		v := he.Vertex.Loc
		v2 := he.Next.Vertex.Loc
		if Collinear(v1, v, v2) {
			return he
		}
		he = he.Next
		if he == l.HalfEdges {
			break
		}
	}
	return nil
}

// HasColinearVerts returns the first loop/half-edge pair with a collinear
// vertex, or nil, nil if none.
func (f *Face) HasColinearVerts() (*Loop, *HalfEdge) {
	for _, l := range f.Loops {
		if he := l.HasColinearVerts(); he != nil {
			return l, he
		}
	}
	return nil, nil
}

// HasColinearVerts returns true if any face contains collinear vertices.
func (s *Solid) HasColinearVerts() bool {
	for f := s.Faces; f != nil; f = f.Next {
		if _, he := f.HasColinearVerts(); he != nil {
			return true
		}
	}
	return false
}

// RemoveColinearVerts removes collinear vertices and degenerate faces
// across the entire solid.
func (s *Solid) RemoveColinearVerts() {
	for f := s.Faces; f != nil; f = f.Next {
		if f.IsDegenerate() {
			s.CollapseFace(f)
		} else {
			f.RemoveColinearVerts()
		}
	}
}

// ---------------------------------------------------------------------------
// Degenerate face detection and removal
// ---------------------------------------------------------------------------

// HasDegenerateFaces returns the first degenerate face, or nil.
func (s *Solid) HasDegenerateFaces() *Face {
	for f := s.Faces; f != nil; f = f.Next {
		if f.IsDegenerate() {
			return f
		}
	}
	return nil
}

// RemoveDegenerateFaces collapses all degenerate faces.
func (s *Solid) RemoveDegenerateFaces() {
	for f := s.Faces; f != nil; {
		next := f.Next
		if f.IsDegenerate() {
			s.CollapseFace(f)
		}
		f = next
	}
}

// ---------------------------------------------------------------------------
// Enhanced per-element verification
// ---------------------------------------------------------------------------

// VerifyError describes a structural integrity problem.
type VerifyError struct {
	Element string
	Index   uint64
	Message string
}

func (e *VerifyError) Error() string {
	return fmt.Sprintf("%s[%d]: %s", e.Element, e.Index, e.Message)
}

// VerifyDetailed checks Euler formula and per-element structural
// integrity. Returns a slice of errors (empty if valid).
func (s *Solid) VerifyDetailed() []error {
	var errs []error

	// Euler formula: F + V - 2 = E + H
	nLoops := 0
	for f := s.Faces; f != nil; f = f.Next {
		nLoops += f.NLoops()
	}
	nHoles := nLoops - s.nFaces
	lhs := s.nFaces + s.nVerts - 2
	rhs := s.nEdges + nHoles
	if lhs != rhs {
		errs = append(errs, &VerifyError{
			Element: "Solid",
			Index:   0,
			Message: fmt.Sprintf("Euler formula failed: F(%d)+V(%d)-2=%d, E(%d)+H(%d)=%d",
				s.nFaces, s.nVerts, lhs, s.nEdges, nHoles, rhs),
		})
	}

	// Verify faces
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
			// Verify loop half-edges
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

	// Verify edges
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

	// Verify vertices
	for v := s.Verts; v != nil; v = v.Next {
		if v.He != nil && v.He.Vertex != v {
			errs = append(errs, &VerifyError{"Vertex", v.Index,
				"half-edge does not point back to vertex"})
		}
	}

	return errs
}

// ---------------------------------------------------------------------------
// Neighborhood queries
// ---------------------------------------------------------------------------

// VertexFaces returns all faces adjacent to vertex v.
func (v *Vertex) VertexFaces() []*Face {
	if v.He == nil {
		return nil
	}
	seen := map[*Face]bool{}
	var faces []*Face
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

// VertexEdges returns all edges incident to vertex v.
func (v *Vertex) VertexEdges() []*Edge {
	if v.He == nil {
		return nil
	}
	seen := map[*Edge]bool{}
	var edges []*Edge
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

// VertexNeighbors returns all vertices connected to v by an edge.
func (v *Vertex) VertexNeighbors() []*Vertex {
	if v.He == nil {
		return nil
	}
	seen := map[*Vertex]bool{}
	var nbrs []*Vertex
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

// FaceNeighbors returns all faces that share an edge with f.
func (f *Face) FaceNeighbors() []*Face {
	seen := map[*Face]bool{}
	var nbrs []*Face
	for _, l := range f.Loops {
		l.ForEachHe(func(he *HalfEdge) bool {
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

// GetNormal computes the face normal using Newell's method (read-only alias).
func (f *Face) GetNormal() vec.SFVec3f {
	return f.Normal
}

// GetD returns the plane distance.
func (f *Face) GetD() float64 {
	return f.D
}
