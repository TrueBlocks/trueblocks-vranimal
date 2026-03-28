package base

import "github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"

// Edge connects two half-edges (He1 and He2) on opposite sides.
type Edge struct {
	He1   *HalfEdge
	He2   *HalfEdge
	Index uint64
	Mark  uint64
	Next  *Edge
	Prev  *Edge
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
func (e *Edge) Marked(m uint64) bool { return e.Mark == m }

// GetSolid returns the solid this edge belongs to (via He1's face).
func (e *Edge) GetSolid() *Solid {
	if e.He1 != nil {
		return e.He1.GetSolid()
	}
	return nil
}

// Length returns the length of this edge.
func (e *Edge) Length() float64 {
	if e.He1 == nil || e.He2 == nil {
		return 0
	}
	return e.He1.Vertex.Loc.Sub(e.He2.Vertex.Loc).Length()
}

// Midpoint returns the midpoint of this edge.
func (e *Edge) Midpoint() vec.SFVec3f {
	if e.He1 == nil || e.He2 == nil {
		return vec.SFVec3f{}
	}
	return e.He1.Vertex.Loc.Add(e.He2.Vertex.Loc).Scale(0.5)
}
