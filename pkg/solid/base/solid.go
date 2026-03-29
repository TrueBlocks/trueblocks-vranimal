package base

import "github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"

// Solid is the top-level container for a half-edge boundary representation.
type Solid struct {
	Faces *Face
	Edges *Edge
	Verts *Vertex
	NFace int
	NEdge int
	NVert int
}

// NewSolid creates an empty solid.
func NewSolid() *Solid {
	return &Solid{}
}

// NFaces returns the number of faces.
func (s *Solid) NFaces() int { return s.NFace }

// NEdges returns the number of edges.
func (s *Solid) NEdges() int { return s.NEdge }

// NVerts returns the number of vertices.
func (s *Solid) NVerts() int { return s.NVert }

// ---------------------------------------------------------------------------
// Intrusive list management
// ---------------------------------------------------------------------------

// AddFace adds a face to the front of the face list.
func (s *Solid) AddFace(f *Face) {
	f.Solid = s
	f.Next = s.Faces
	f.Prev = nil
	if s.Faces != nil {
		s.Faces.Prev = f
	}
	s.Faces = f
	s.NFace++
}

// AddEdge adds an edge to the front of the edge list.
func (s *Solid) AddEdge(e *Edge) {
	e.Next = s.Edges
	e.Prev = nil
	if s.Edges != nil {
		s.Edges.Prev = e
	}
	s.Edges = e
	s.NEdge++
}

// AddVertex adds a vertex to the front of the vertex list.
func (s *Solid) AddVertex(v *Vertex) {
	v.Next = s.Verts
	v.Prev = nil
	if s.Verts != nil {
		s.Verts.Prev = v
	}
	s.Verts = v
	s.NVert++
}

// RemoveFace removes a face from the list.
func (s *Solid) RemoveFace(f *Face) {
	if f.Prev != nil {
		f.Prev.Next = f.Next
	} else {
		s.Faces = f.Next
	}
	if f.Next != nil {
		f.Next.Prev = f.Prev
	}
	s.NFace--
}

// RemoveEdge removes an edge from the list.
func (s *Solid) RemoveEdge(e *Edge) {
	if e.Prev != nil {
		e.Prev.Next = e.Next
	} else {
		s.Edges = e.Next
	}
	if e.Next != nil {
		e.Next.Prev = e.Prev
	}
	s.NEdge--
}

// RemoveVertex removes a vertex from the list.
func (s *Solid) RemoveVertex(v *Vertex) {
	if v.Prev != nil {
		v.Prev.Next = v.Next
	} else {
		s.Verts = v.Next
	}
	if v.Next != nil {
		v.Next.Prev = v.Prev
	}
	s.NVert--
}

// ---------------------------------------------------------------------------
// Iterators
// ---------------------------------------------------------------------------

// ForEachFace calls fn for each face; stops if fn returns false.
func (s *Solid) ForEachFace(fn func(*Face) bool) {
	for f := s.Faces; f != nil; f = f.Next {
		if !fn(f) {
			return
		}
	}
}

// ForEachEdge calls fn for each edge; stops if fn returns false.
func (s *Solid) ForEachEdge(fn func(*Edge) bool) {
	for e := s.Edges; e != nil; e = e.Next {
		if !fn(e) {
			return
		}
	}
}

// ForEachVertex calls fn for each vertex; stops if fn returns false.
func (s *Solid) ForEachVertex(fn func(*Vertex) bool) {
	for v := s.Verts; v != nil; v = v.Next {
		if !fn(v) {
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Queries
// ---------------------------------------------------------------------------

// FindFace returns the face at the given index, or nil.
func (s *Solid) FindFace(index uint64) *Face {
	for f := s.Faces; f != nil; f = f.Next {
		if f.Index == index {
			return f
		}
	}
	return nil
}

// FindVertex returns the vertex at the given index, or nil.
func (s *Solid) FindVertex(index uint64) *Vertex {
	for v := s.Verts; v != nil; v = v.Next {
		if v.Index == index {
			return v
		}
	}
	return nil
}

// FindEdge returns the edge at the given index, or nil.
func (s *Solid) FindEdge(index uint64) *Edge {
	for e := s.Edges; e != nil; e = e.Next {
		if e.Index == index {
			return e
		}
	}
	return nil
}

// FindHalfEdge finds the half-edge going from vertex v1 to v2, or nil.
func (s *Solid) FindHalfEdge(v1, v2 *Vertex) *HalfEdge {
	for e := s.Edges; e != nil; e = e.Next {
		if e.He1.Vertex == v1 && e.He2.Vertex == v2 {
			return e.He1
		}
		if e.He2.Vertex == v1 && e.He1.Vertex == v2 {
			return e.He2
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Geometry operations
// ---------------------------------------------------------------------------

// Extents computes the axis-aligned bounding box of all vertices.
func (s *Solid) Extents() (min, max vec.SFVec3f) {
	first := true
	for v := s.Verts; v != nil; v = v.Next {
		if first {
			min = v.Loc
			max = v.Loc
			first = false
			continue
		}
		if v.Loc.X < min.X {
			min.X = v.Loc.X
		}
		if v.Loc.Y < min.Y {
			min.Y = v.Loc.Y
		}
		if v.Loc.Z < min.Z {
			min.Z = v.Loc.Z
		}
		if v.Loc.X > max.X {
			max.X = v.Loc.X
		}
		if v.Loc.Y > max.Y {
			max.Y = v.Loc.Y
		}
		if v.Loc.Z > max.Z {
			max.Z = v.Loc.Z
		}
	}
	return
}

// Stats returns face, edge, vertex counts.
func (s *Solid) Stats() (faces, edges, verts int) {
	return s.NFace, s.NEdge, s.NVert
}

// Volume computes the volume using the divergence theorem.
func (s *Solid) Volume() float64 {
	var vol float64
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut == nil {
			continue
		}
		he := f.LoopOut.HalfEdges
		if he == nil {
			continue
		}
		v0 := he.Vertex.Loc
		he = he.Next
		for he != f.LoopOut.HalfEdges {
			next := he.Next
			if next == f.LoopOut.HalfEdges {
				break
			}
			v1 := he.Vertex.Loc
			v2 := next.Vertex.Loc
			vol += v0.Dot(v1.Cross(v2))
			he = next
		}
	}
	return vol / 6.0
}

// SetColor sets every face to the given color.
func (s *Solid) SetColor(c vec.SFColor) {
	for f := s.Faces; f != nil; f = f.Next {
		f.SetColor(c)
	}
}

// SetFaceMarks sets Mark1 on all faces.
func (s *Solid) SetFaceMarks(m uint64) {
	for f := s.Faces; f != nil; f = f.Next {
		f.Mark1 = m
	}
}

// SetFaceMarks2 sets Mark2 on all faces.
func (s *Solid) SetFaceMarks2(m uint64) {
	for f := s.Faces; f != nil; f = f.Next {
		f.Mark2 = m
	}
}

// ClearFaceMarks2 resets Mark2 on all faces.
func (s *Solid) ClearFaceMarks2() {
	for f := s.Faces; f != nil; f = f.Next {
		f.Mark2 = 0
	}
}

// SetVertexMarks sets Mark on all vertices.
func (s *Solid) SetVertexMarks(m uint64) {
	for v := s.Verts; v != nil; v = v.Next {
		v.Mark = m
	}
}

// SetEdgeMarks sets Mark on all edges.
func (s *Solid) SetEdgeMarks(m uint64) {
	for e := s.Edges; e != nil; e = e.Next {
		e.Mark = m
	}
}

// CalcPlaneEquations recalculates plane equations for all faces.
func (s *Solid) CalcPlaneEquations() {
	for f := s.Faces; f != nil; f = f.Next {
		f.CalcEquation()
	}
}

// TransformGeometry applies a matrix to all vertex positions.
func (s *Solid) TransformGeometry(m vec.Matrix) {
	for v := s.Verts; v != nil; v = v.Next {
		v.Loc = m.TransformPoint(v.Loc)
	}
}

// Renumber assigns sequential indices to faces, edges, and vertices.
func (s *Solid) Renumber() {
	var i uint64
	for f := s.Faces; f != nil; f = f.Next {
		f.Index = i
		i++
	}
	i = 0
	for e := s.Edges; e != nil; e = e.Next {
		e.Index = i
		i++
	}
	i = 0
	for v := s.Verts; v != nil; v = v.Next {
		v.Index = i
		i++
	}
}

// Merge appends all faces, edges, and vertices from other into s.
func (s *Solid) Merge(other *Solid) {
	if other == nil {
		return
	}
	if other.Faces != nil {
		last := other.Faces
		for last.Next != nil {
			last.Solid = s
			last = last.Next
		}
		last.Solid = s
		last.Next = s.Faces
		if s.Faces != nil {
			s.Faces.Prev = last
		}
		s.Faces = other.Faces
		s.NFace += other.NFace
	}
	if other.Edges != nil {
		last := other.Edges
		for last.Next != nil {
			last = last.Next
		}
		last.Next = s.Edges
		if s.Edges != nil {
			s.Edges.Prev = last
		}
		s.Edges = other.Edges
		s.NEdge += other.NEdge
	}
	if other.Verts != nil {
		last := other.Verts
		for last.Next != nil {
			last = last.Next
		}
		last.Next = s.Verts
		if s.Verts != nil {
			s.Verts.Prev = last
		}
		s.Verts = other.Verts
		s.NVert += other.NVert
	}
}

// ---------------------------------------------------------------------------
// Accessors
// ---------------------------------------------------------------------------

// GetFirstFace returns the head of the face list.
func (s *Solid) GetFirstFace() *Face { return s.Faces }

// GetFirstEdge returns the head of the edge list.
func (s *Solid) GetFirstEdge() *Edge { return s.Edges }

// GetFirstVertex returns the head of the vertex list.
func (s *Solid) GetFirstVertex() *Vertex { return s.Verts }

// IsWire returns true if the solid has exactly one face (wire-frame).
func (s *Solid) IsWire() bool { return s.NFace == 1 }

// IsLamina returns true if the solid has exactly two faces.
func (s *Solid) IsLamina() bool { return s.NFace == 2 }

// Revert reverses the winding order of all faces.
func (s *Solid) Revert() {
	for f := s.Faces; f != nil; f = f.Next {
		f.Revert()
	}
}

// CalcVertexNormals computes vertex normals for all vertices.
func (s *Solid) CalcVertexNormals() {
	for v := s.Verts; v != nil; v = v.Next {
		v.CalcNormal()
	}
}

// AllocateColorData allocates ColorData for elements at the specified level.
func (s *Solid) AllocateColorData(where int) {
	switch where {
	case PerVertex:
		for v := s.Verts; v != nil; v = v.Next {
			if v.Data == nil {
				v.Data = &ColorData{}
			}
		}
	case PerVertexPerFace:
		for f := s.Faces; f != nil; f = f.Next {
			for _, l := range f.Loops {
				l.ForEachHe(func(he *HalfEdge) bool {
					if he.Data == nil {
						he.Data = &ColorData{}
					}
					return true
				})
			}
		}
	case PerFace:
		for f := s.Faces; f != nil; f = f.Next {
			if f.Data == nil {
				f.Data = &ColorData{}
			}
		}
	}
}

// Plane represents an oriented plane in 3D space: Normal · P + D = 0.
type Plane struct {
	Normal vec.SFVec3f
	D      float64
}

// GetDistance returns the signed distance from point p to the plane.
func (pl *Plane) GetDistance(p vec.SFVec3f) float64 {
	return pl.Normal.Dot(p) + pl.D
}

func (s *Solid) Cleanup() {
	s.Edges = nil
	s.NEdge = 0
	s.Verts = nil
	s.NVert = 0

	for f := s.Faces; f != nil; f = f.Next {
		for _, l := range f.Loops {
			l.ForEachHe(func(he *HalfEdge) bool {
				if he.Edge != nil && !he.Edge.Marked(VISITED) {
					he.Edge.Mark = VISITED
					s.AddEdge(he.Edge)
				}
				if he.Vertex != nil && !he.Vertex.IsMarked(VISITED) {
					he.Vertex.Mark = VISITED
					s.AddVertex(he.Vertex)
				}
				return true
			})
		}
	}

	for e := s.Edges; e != nil; e = e.Next {
		e.Mark = 0
	}
	for v := s.Verts; v != nil; v = v.Next {
		v.Mark = 0
	}
}

func (s *Solid) Copy() *Solid {
	ns := NewSolid()
	if s.Verts == nil {
		return ns
	}

	vmap := make(map[*Vertex]*Vertex)
	for v := s.Verts; v != nil; v = v.Next {
		nv := NewVertexVec(v.Loc)
		nv.Index = v.Index
		nv.Mark = v.Mark
		nv.Scratch = v.Scratch
		if v.Data != nil {
			d := *v.Data
			nv.Data = &d
		}
		ns.AddVertex(nv)
		vmap[v] = nv
	}

	hemap := make(map[*HalfEdge]*HalfEdge)

	for f := s.Faces; f != nil; f = f.Next {
		nf := &Face{
			Solid:  ns,
			Normal: f.Normal,
			D:      f.D,
			Index:  f.Index,
			Mark1:  f.Mark1,
			Mark2:  f.Mark2,
		}
		if f.Data != nil {
			d := *f.Data
			nf.Data = &d
		}
		ns.AddFace(nf)

		for _, l := range f.Loops {
			nl := &Loop{Face: nf}
			isOuter := l == f.LoopOut
			nf.AddLoop(nl, isOuter)

			l.ForEachHe(func(he *HalfEdge) bool {
				nv := vmap[he.Vertex]
				nhe := &HalfEdge{
					Vertex: nv,
					Loop:   nl,
					Mark:   he.Mark,
				}
				if he.Data != nil {
					d := *he.Data
					nhe.Data = &d
				}
				nhe.Next = nhe
				nhe.Prev = nhe
				nl.AddHalfEdge(nhe)
				if nv.He == nil {
					nv.He = nhe
				}
				hemap[he] = nhe
				return true
			})
		}
	}

	for e := s.Edges; e != nil; e = e.Next {
		ne := &Edge{
			Index: e.Index,
			Mark:  e.Mark,
		}
		if nhe1, ok := hemap[e.He1]; ok {
			ne.He1 = nhe1
			nhe1.Edge = ne
		}
		if nhe2, ok := hemap[e.He2]; ok {
			ne.He2 = nhe2
			nhe2.Edge = ne
		}
		ns.AddEdge(ne)
	}

	return ns
}
