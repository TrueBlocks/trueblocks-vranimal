package base

import "github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"

// Vertex represents a point in 3D space with adjacency information.
type Vertex struct {
	Loc     vec.SFVec3f
	He      *HalfEdge // One half-edge starting at this vertex
	Index   uint64
	Mark    uint64
	Scratch float64
	Data    *ColorData
	Next    *Vertex
	Prev    *Vertex
}

// NewVertex creates a vertex at the given position.
func NewVertex(x, y, z float64) *Vertex {
	return &Vertex{Loc: vec.SFVec3f{X: x, Y: y, Z: z}}
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

// IsMarked returns true if the vertex mark equals m.
func (v *Vertex) IsMarked(m uint64) bool { return v.Mark == m }

// GetValence returns the number of edges incident on this vertex.
func (v *Vertex) GetValence() int {
	if v.He == nil {
		return 0
	}
	n := 0
	he := v.He
	for {
		n++
		m := he.GetMate()
		if m == nil {
			break
		}
		he = m.Next
		if he == v.He {
			break
		}
	}
	return n
}

// CalcNormal computes the vertex normal by averaging adjacent face normals,
// respecting crease edges. Ported from vraniml/src/solid/vertex.cpp.
func (v *Vertex) CalcNormal() {
	if v.He == nil {
		return
	}

	var normal vec.SFVec3f
	var nFaces int
	var crease, nocrease *HalfEdge

	start := v.He
	he := start
	for {
		if he.Edge != nil && he.Edge.Mark&CREASE != 0 {
			crease = he
		} else {
			nocrease = he
		}
		normal = normal.Add(he.GetFaceNormal())
		nFaces++
		if nFaces > 1000 {
			v.SetNormal(vec.YAxis)
			return
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

	if crease == nil {
		normal = normal.Scale(1.0 / float64(nFaces)).Normalize()
		v.SetNormal(normal)
		return
	}

	if nocrease == nil {
		return
	}

	if crease.GetMate() == nil {
		crease.SetNormal(crease.GetFaceNormal())
		return
	}

	heTable := make([]*HalfEdge, 0, 32)
	var accum vec.SFVec3f

	he = crease.GetMate().Next
	start = he
	for {
		heTable = append(heTable, he)
		accum = accum.Add(he.GetFaceNormal())

		if he.Edge != nil && he.Edge.Mark&CREASE != 0 {
			n := accum.Scale(1.0 / float64(len(heTable))).Normalize()
			for _, h := range heTable {
				h.SetNormal(n)
			}
			heTable = heTable[:0]
			accum = vec.SFVec3f{}
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

	if len(heTable) > 0 {
		n := accum.Scale(1.0 / float64(len(heTable))).Normalize()
		for _, h := range heTable {
			h.SetNormal(n)
		}
	}
}
