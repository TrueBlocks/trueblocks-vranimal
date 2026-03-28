package base

import "github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"

// HalfEdge is one half of a directed edge in the half-edge data structure.
type HalfEdge struct {
	Vertex *Vertex
	Edge   *Edge
	Loop   *Loop
	Mark   uint64
	Data   *ColorData
	Next   *HalfEdge
	Prev   *HalfEdge
}

// NewHalfEdge creates a half-edge pointing to the given vertex within a loop.
func NewHalfEdge(l *Loop, v *Vertex) *HalfEdge {
	he := &HalfEdge{
		Vertex: v,
		Loop:   l,
	}
	he.Next = he
	he.Prev = he
	return he
}

// GetMate returns the opposite half-edge (the other side of the same edge).
func (he *HalfEdge) GetMate() *HalfEdge {
	if he.Edge == nil {
		return nil
	}
	if he.Edge.He1 == he {
		return he.Edge.He2
	}
	return he.Edge.He1
}

// GetFace returns the face that this half-edge's loop belongs to.
func (he *HalfEdge) GetFace() *Face {
	if he.Loop == nil {
		return nil
	}
	return he.Loop.Face
}

// GetFaceNormal returns the normal of the face this half-edge belongs to.
func (he *HalfEdge) GetFaceNormal() vec.SFVec3f {
	f := he.GetFace()
	if f == nil {
		return vec.SFVec3f{}
	}
	return f.Normal
}

// GetMateFace returns the face on the other side of this edge.
func (he *HalfEdge) GetMateFace() *Face {
	m := he.GetMate()
	if m == nil {
		return nil
	}
	return m.GetFace()
}

// GetMateVertex returns the destination vertex of the mate half-edge.
func (he *HalfEdge) GetMateVertex() *Vertex {
	m := he.GetMate()
	if m == nil {
		return nil
	}
	return m.Vertex
}

// GetIndex returns the index of the destination vertex.
func (he *HalfEdge) GetIndex() uint64 {
	if he.Vertex == nil {
		return 0
	}
	return he.Vertex.Index
}

// IsNullEdge returns true if both half-edges point to the same vertex.
func (he *HalfEdge) IsNullEdge() bool {
	m := he.GetMate()
	if m == nil {
		return false
	}
	return he.Vertex == m.Vertex
}

// Bisect returns the midpoint between this vertex and the mate vertex.
func (he *HalfEdge) Bisect() vec.SFVec3f {
	m := he.GetMate()
	if m == nil {
		return he.Vertex.Loc
	}
	return he.Vertex.Loc.Add(m.Vertex.Loc).Scale(0.5)
}

// SetColor sets per-half-edge color data.
func (he *HalfEdge) SetColor(c vec.SFColor) {
	if he.Data == nil {
		he.Data = &ColorData{}
	}
	he.Data.SetColor(c)
}

// SetNormal sets per-half-edge normal data.
func (he *HalfEdge) SetNormal(n vec.SFVec3f) {
	if he.Data == nil {
		he.Data = &ColorData{}
	}
	he.Data.SetNormal(n)
}

// SetTexCoord sets per-half-edge texture coordinate data.
func (he *HalfEdge) SetTexCoord(t vec.SFVec2f) {
	if he.Data == nil {
		he.Data = &ColorData{}
	}
	he.Data.SetTexCoord(t)
}

// GetColor returns the half-edge color or the default.
func (he *HalfEdge) GetColor(def vec.SFColor) vec.SFColor {
	if he.Data != nil && he.Data.HasColor() {
		return he.Data.GetColor()
	}
	return def
}

// GetNormal returns the half-edge normal or the default.
func (he *HalfEdge) GetNormal(def vec.SFVec3f) vec.SFVec3f {
	if he.Data != nil && he.Data.HasNormal() {
		return he.Data.GetNormal()
	}
	return def
}

// GetTexCoord returns the half-edge texture coordinate or the default.
func (he *HalfEdge) GetTexCoord(def vec.SFVec2f) vec.SFVec2f {
	if he.Data != nil && he.Data.HasTexCoord() {
		return he.Data.GetTexCoord()
	}
	return def
}

// GetMateIndex returns the vertex index of the mate half-edge.
func (he *HalfEdge) GetMateIndex() uint64 {
	m := he.GetMate()
	if m == nil {
		return 0
	}
	return m.GetIndex()
}

// GetSolid returns the solid this half-edge belongs to.
func (he *HalfEdge) GetSolid() *Solid {
	f := he.GetFace()
	if f == nil {
		return nil
	}
	return f.Solid
}

// IsNullStrut returns true if this edge is a "strut" — the mate is adjacent in the loop.
func (he *HalfEdge) IsNullStrut() bool {
	m := he.GetMate()
	if m == nil {
		return false
	}
	return he == m.Next || he == m.Prev
}

// IsMate returns true if other is the mate of this half-edge.
func (he *HalfEdge) IsMate(other *HalfEdge) bool {
	return he.GetMate() == other
}

// IsNeighbor returns true if he and other share a face and are on opposite
// sides of their edges.
func (he *HalfEdge) IsNeighbor(other *HalfEdge) bool {
	if he.Edge == nil || other.Edge == nil {
		return false
	}
	if he.GetFace() != other.GetFace() {
		return false
	}
	return (he == he.Edge.He1 && other == other.Edge.He2) ||
		(he == he.Edge.He2 && other == other.Edge.He1)
}

// InsideVector returns the inward-pointing vector at this half-edge.
func (he *HalfEdge) InsideVector() vec.SFVec3f {
	dir := he.Next.Vertex.Loc.Sub(he.Vertex.Loc)
	return he.GetFaceNormal().Cross(dir)
}

// IsWide returns true if the interior angle at this vertex is greater than 180°.
// If inc180 is true, exactly 180° is also considered wide.
func (he *HalfEdge) IsWide(inc180 bool) bool {
	v2 := he.Next.Vertex.Loc.Sub(he.Vertex.Loc)
	v1 := he.Prev.Vertex.Loc.Sub(he.Vertex.Loc)
	cross := v1.Cross(v2)
	if cross.Length() < 1e-12 {
		return inc180
	}
	cross = cross.Normalize()
	norm := he.GetFaceNormal()
	return cross.Dot(norm) < 0
}

// Is180 returns true if the interior angle is exactly 180°.
func (he *HalfEdge) Is180() bool {
	return he.IsWide(true) && !he.IsWide(false)
}

// IsConvexEdge returns true if the dihedral angle across this edge is convex.
func (he *HalfEdge) IsConvexEdge() bool {
	m := he.GetMate()
	if m == nil {
		return true
	}
	return !he.IsWide(false) && !m.IsWide(false)
}

// SetMark sets the half-edge mark.
func (he *HalfEdge) SetMark(m uint64) { he.Mark = m }

// Marked returns true if the half-edge mark equals m.
func (he *HalfEdge) Marked(m uint64) bool { return he.Mark == m }
