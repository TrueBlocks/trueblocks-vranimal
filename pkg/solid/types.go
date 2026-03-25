package solid

// Package solid implements the half-edge boundary representation data structure
// for manifold 3D solid geometry incorporating vertices, edges, half-edges,
// loops, and faces.
// Ported from vraniml/src/solid/.

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Mark constants from decls.h
// ---------------------------------------------------------------------------

const (
	ON    = 0
	ABOVE = 1
	BELOW = -1

	UNKNOWN       = 1 << 1
	BRANDNEW      = 1 << 2
	VISITED       = 1 << 3
	ALTERED       = 1 << 4
	CREASE        = 1 << 12
	PARTIALORMAL  = 1 << 13
	INVISIBLE     = 1 << 15
	DELETED       = 0xdddddddd
	POINTSET      = 1 << 20
	LINESET       = 1 << 21
	FACESET       = POINTSET | LINESET
	CALCED        = 1 << 25

	SOLIDNONE = 0
	SOLIDA    = 11
	SOLIDB    = 12

	BoolUnion        = 1
	BoolIntersection = 2
	BoolDifference   = 3

	PerFace          = 1
	PerVertex        = 2
	PerVertexPerFace = 3
)

// ColorDataType flags
const (
	ColorFlag    = 0x1
	NormalFlag   = 0x2
	TexCoordFlag = 0x4
)

// ---------------------------------------------------------------------------
// ColorData - per-element attribute data
// ---------------------------------------------------------------------------

// ColorData holds per-vertex or per-face color, normal, and texture coordinate data.
type ColorData struct {
	Type     int32
	Color    vec.SFColor
	Normal   vec.SFVec3f
	TexCoord vec.SFVec2f
}

// SetColor sets the color and marks the color flag.
func (d *ColorData) SetColor(c vec.SFColor) {
	d.Type |= ColorFlag
	d.Color = c
}

// SetNormal sets the normal and marks the normal flag.
func (d *ColorData) SetNormal(n vec.SFVec3f) {
	d.Type |= NormalFlag
	d.Normal = n
}

// SetTexCoord sets the texture coordinate and marks the texcoord flag.
func (d *ColorData) SetTexCoord(t vec.SFVec2f) {
	d.Type |= TexCoordFlag
	d.TexCoord = t
}

// GetColor returns the color (caller should check HasColor).
func (d *ColorData) GetColor() vec.SFColor { return d.Color }

// GetNormal returns the normal (caller should check HasNormal).
func (d *ColorData) GetNormal() vec.SFVec3f { return d.Normal }

// GetTexCoord returns the texture coordinate (caller should check HasTexCoord).
func (d *ColorData) GetTexCoord() vec.SFVec2f { return d.TexCoord }

// HasColor returns true if color data is present.
func (d *ColorData) HasColor() bool { return d.Type&ColorFlag != 0 }

// HasNormal returns true if normal data is present.
func (d *ColorData) HasNormal() bool { return d.Type&NormalFlag != 0 }

// HasTexCoord returns true if texture coordinate data is present.
func (d *ColorData) HasTexCoord() bool { return d.Type&TexCoordFlag != 0 }

// ---------------------------------------------------------------------------
// HalfEdge - directed edge in a loop
// ---------------------------------------------------------------------------

// HalfEdge is one half of a directed edge in the half-edge data structure.
type HalfEdge struct {
	Vertex *Vertex
	Edge   *Edge
	Loop   *Loop
	Mark   uint32
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
func (he *HalfEdge) GetIndex() uint32 {
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

// ---------------------------------------------------------------------------
// Loop - ordered ring of half-edges forming a polygon boundary
// ---------------------------------------------------------------------------

// Loop is a circular ring of half-edges that form a polygon boundary or hole.
type Loop struct {
	Face      *Face
	HalfEdges *HalfEdge // Head of the circular list
}

// NewLoop creates a loop belonging to the given face.
func NewLoop(f *Face, isOuter bool) *Loop {
	l := &Loop{Face: f}
	f.AddLoop(l, isOuter)
	return l
}

// GetFirstHe returns the first half-edge in the loop.
func (l *Loop) GetFirstHe() *HalfEdge {
	return l.HalfEdges
}

// SetFirstHe sets the head of the half-edge ring.
func (l *Loop) SetFirstHe(he *HalfEdge) {
	l.HalfEdges = he
}

// AddHalfEdge appends a half-edge after the current tail.
func (l *Loop) AddHalfEdge(he *HalfEdge) {
	he.Loop = l
	if l.HalfEdges == nil {
		he.Next = he
		he.Prev = he
		l.HalfEdges = he
		return
	}
	tail := l.HalfEdges.Prev
	he.Next = l.HalfEdges
	he.Prev = tail
	tail.Next = he
	l.HalfEdges.Prev = he
}

// InsertHalfEdge inserts he after the before half-edge in the ring.
func (l *Loop) InsertHalfEdge(before, he *HalfEdge) {
	he.Loop = l
	he.Next = before.Next
	he.Prev = before
	before.Next.Prev = he
	before.Next = he
}

// RemoveHe removes a half-edge from the ring.
func (l *Loop) RemoveHe(he *HalfEdge) {
	if he.Next == he {
		l.HalfEdges = nil
		return
	}
	he.Prev.Next = he.Next
	he.Next.Prev = he.Prev
	if l.HalfEdges == he {
		l.HalfEdges = he.Next
	}
}

// ForEachHe iterates over all half-edges in the loop, calling fn for each.
// Stops if fn returns false.
func (l *Loop) ForEachHe(fn func(*HalfEdge) bool) {
	if l.HalfEdges == nil {
		return
	}
	he := l.HalfEdges
	for {
		if !fn(he) {
			return
		}
		he = he.Next
		if he == l.HalfEdges {
			return
		}
	}
}

// Area calculates the signed area of the loop polygon.
func (l *Loop) Area() float32 {
	if l.HalfEdges == nil {
		return 0
	}
	he := l.HalfEdges
	var area vec.SFVec3f
	for {
		next := he.Next
		cross := he.Vertex.Loc.Cross(next.Vertex.Loc)
		area = area.Add(cross)
		he = next
		if he == l.HalfEdges {
			break
		}
	}
	if l.Face != nil {
		return 0.5 * area.Dot(l.Face.Normal)
	}
	return 0.5 * area.Length()
}

// IsOuterLoop returns true if this loop is the outer boundary of its face.
func (l *Loop) IsOuterLoop() bool {
	if l.Face == nil {
		return false
	}
	return l.Face.LoopOut == l
}

// ---------------------------------------------------------------------------
// Vertex - point in 3D space within the solid
// ---------------------------------------------------------------------------

// Vertex represents a point in 3D space with adjacency information.
type Vertex struct {
	Loc     vec.SFVec3f
	He      *HalfEdge // One half-edge starting at this vertex
	Index   uint32
	Mark    uint32
	Scratch float32
	Data    *ColorData
	// Intrusive list pointers
	Next *Vertex
	Prev *Vertex
}

// NewVertex creates a vertex at the given position.
func NewVertex(x, y, z float32) *Vertex {
	return &Vertex{
		Loc: vec.SFVec3f{X: x, Y: y, Z: z},
	}
}

// NewVertexVec creates a vertex from a vector.
func NewVertexVec(v vec.SFVec3f) *Vertex {
	return &Vertex{Loc: v}
}

// SetColor sets per-vertex color.
func (v *Vertex) SetColor(c vec.SFColor) {
	if v.Data == nil {
		v.Data = &ColorData{}
	}
	v.Data.SetColor(c)
}

// SetNormal sets per-vertex normal.
func (v *Vertex) SetNormal(n vec.SFVec3f) {
	if v.Data == nil {
		v.Data = &ColorData{}
	}
	v.Data.SetNormal(n)
}

// SetTexCoord sets per-vertex texture coordinate.
func (v *Vertex) SetTexCoord(t vec.SFVec2f) {
	if v.Data == nil {
		v.Data = &ColorData{}
	}
	v.Data.SetTexCoord(t)
}

// GetColor returns the vertex color or the default.
func (v *Vertex) GetColor(def vec.SFColor) vec.SFColor {
	if v.Data != nil && v.Data.HasColor() {
		return v.Data.GetColor()
	}
	return def
}

// GetNormal returns the vertex normal or the default.
func (v *Vertex) GetNormal(def vec.SFVec3f) vec.SFVec3f {
	if v.Data != nil && v.Data.HasNormal() {
		return v.Data.GetNormal()
	}
	return def
}

// GetTexCoord returns the vertex texture coordinate or the default.
func (v *Vertex) GetTexCoord(def vec.SFVec2f) vec.SFVec2f {
	if v.Data != nil && v.Data.HasTexCoord() {
		return v.Data.GetTexCoord()
	}
	return def
}

// ---------------------------------------------------------------------------
// Edge - undirected edge connecting two half-edges
// ---------------------------------------------------------------------------

// Edge connects two half-edges (He1 and He2) on opposite sides.
type Edge struct {
	He1   *HalfEdge
	He2   *HalfEdge
	Index uint32
	Mark  uint32
	// Intrusive list pointers
	Next *Edge
	Prev *Edge
}

// NewEdge creates an empty edge.
func NewEdge() *Edge {
	return &Edge{}
}

// SwapHes swaps the two half-edges.
func (e *Edge) SwapHes() {
	e.He1, e.He2 = e.He2, e.He1
}

// GetVertex returns the vertex of the specified half-edge (0=He1, 1=He2).
func (e *Edge) GetVertex(which int) *Vertex {
	if which == 0 {
		return e.He1.Vertex
	}
	return e.He2.Vertex
}

// GetLoop returns the loop of the specified half-edge.
func (e *Edge) GetLoop(which int) *Loop {
	if which == 0 {
		return e.He1.Loop
	}
	return e.He2.Loop
}

// Marked returns true if the edge mark equals m.
func (e *Edge) Marked(m uint32) bool { return e.Mark == m }

// ---------------------------------------------------------------------------
// Face - polygon face with outer boundary and optional holes
// ---------------------------------------------------------------------------

// Face represents a polygon face defined by loops of half-edges.
type Face struct {
	Solid   *Solid
	LoopOut *Loop     // Outer boundary loop
	Loops   []*Loop   // All loops (outer + holes)
	Normal  vec.SFVec3f
	D       float32   // Plane distance
	Index   uint32
	Mark1   uint32
	Mark2   uint32
	Data    *ColorData
	// Intrusive list pointers
	Next *Face
	Prev *Face
}

// NewFace creates a face belonging to the given solid.
func NewFace(s *Solid, color vec.SFColor) *Face {
	f := &Face{
		Solid: s,
	}
	f.Data = &ColorData{}
	f.Data.SetColor(color)
	return f
}

// AddLoop adds a loop to this face; if isOuter, sets it as the outer boundary.
func (f *Face) AddLoop(l *Loop, isOuter bool) {
	l.Face = f
	f.Loops = append(f.Loops, l)
	if isOuter {
		f.LoopOut = l
	}
}

// RemoveLoop removes a loop from this face's loop list.
func (f *Face) RemoveLoop(l *Loop) {
	for i, ll := range f.Loops {
		if ll == l {
			f.Loops = append(f.Loops[:i], f.Loops[i+1:]...)
			if f.LoopOut == l {
				f.LoopOut = nil
			}
			return
		}
	}
}

// GetFirstHe returns the first half-edge of the outer loop.
func (f *Face) GetFirstHe() *HalfEdge {
	if f.LoopOut == nil {
		return nil
	}
	return f.LoopOut.GetFirstHe()
}

// GetFirstVertex returns the first vertex of the outer loop.
func (f *Face) GetFirstVertex() *Vertex {
	he := f.GetFirstHe()
	if he == nil {
		return nil
	}
	return he.Vertex
}

// NLoops returns the number of loops.
func (f *Face) NLoops() int { return len(f.Loops) }

// GetColor returns the face color or the default.
func (f *Face) GetColor(def vec.SFColor) vec.SFColor {
	if f.Data != nil && f.Data.HasColor() {
		return f.Data.GetColor()
	}
	return def
}

// SetColor sets the face color.
func (f *Face) SetColor(c vec.SFColor) {
	if f.Data == nil {
		f.Data = &ColorData{}
	}
	f.Data.SetColor(c)
}

// GetCenter returns the centroid of the outer loop.
func (f *Face) GetCenter() vec.SFVec3f {
	if f.LoopOut == nil {
		return vec.SFVec3f{}
	}
	var sum vec.SFVec3f
	n := 0
	f.LoopOut.ForEachHe(func(he *HalfEdge) bool {
		sum = sum.Add(he.Vertex.Loc)
		n++
		return true
	})
	if n == 0 {
		return vec.SFVec3f{}
	}
	return sum.Scale(1 / float32(n))
}

// CalcEquation calculates the plane equation (normal + D) using Newell's method.
func (f *Face) CalcEquation() bool {
	if f.LoopOut == nil {
		return false
	}
	return f.CalcEquationFromLoop(f.LoopOut)
}

// CalcEquationFromLoop calculates the plane equation from a specific loop.
func (f *Face) CalcEquationFromLoop(l *Loop) bool {
	if l.HalfEdges == nil {
		return false
	}
	var normal vec.SFVec3f
	he := l.HalfEdges
	for {
		next := he.Next
		v1 := he.Vertex.Loc
		v2 := next.Vertex.Loc
		normal.X += (v1.Y - v2.Y) * (v1.Z + v2.Z)
		normal.Y += (v1.Z - v2.Z) * (v1.X + v2.X)
		normal.Z += (v1.X - v2.X) * (v1.Y + v2.Y)
		he = next
		if he == l.HalfEdges {
			break
		}
	}
	normal = normal.Normalize()
	f.Normal = normal
	f.D = -normal.Dot(l.HalfEdges.Vertex.Loc)
	return true
}

// GetDistance returns the signed distance from a point to this face's plane.
func (f *Face) GetDistance(pt vec.SFVec3f) float32 {
	return f.Normal.Dot(pt) + f.D
}

// InvertNormal reverses the face normal direction.
func (f *Face) InvertNormal() {
	f.Normal = f.Normal.Negate()
	f.D = -f.D
}

// Area returns the area of the face's outer loop.
func (f *Face) Area() float32 {
	if f.LoopOut == nil {
		return 0
	}
	a := f.LoopOut.Area()
	if a < 0 {
		return -a
	}
	return a
}

// Revert reverses the winding order of all loops.
func (f *Face) Revert() {
	for _, l := range f.Loops {
		if l.HalfEdges == nil {
			continue
		}
		he := l.HalfEdges
		for {
			he.Next, he.Prev = he.Prev, he.Next
			he = he.Prev // was Next before swap
			if he == l.HalfEdges {
				break
			}
		}
	}
}

// Marked1 returns true if mark1 equals m.
func (f *Face) Marked1(m uint32) bool { return f.Mark1 == m }

// Marked2 returns true if mark2 equals m.
func (f *Face) Marked2(m uint32) bool { return f.Mark2 == m }
