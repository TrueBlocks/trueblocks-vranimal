package base

import "github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"

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
func (l *Loop) Area() float64 {
	if l.HalfEdges == nil {
		return 0
	}
	he := l.HalfEdges
	var ax, ay, az float64
	for {
		next := he.Next
		v1 := he.Vertex.Loc
		v2 := next.Vertex.Loc
		ax += v1.Y*v2.Z - v1.Z*v2.Y
		ay += v1.Z*v2.X - v1.X*v2.Z
		az += v1.X*v2.Y - v1.Y*v2.X
		he = next
		if he == l.HalfEdges {
			break
		}
	}
	if l.Face != nil {
		n := l.Face.Normal
		return 0.5 * (ax*n.X + ay*n.Y + az*n.Z)
	}
	// fallback: magnitude
	lenSq := ax*ax + ay*ay + az*az
	if lenSq == 0 {
		return 0
	}
	return 0.5 * vec.SFVec3f{X: ax, Y: ay, Z: az}.Length()
}

// IsOuterLoop returns true if this loop is the outer boundary of its face.
func (l *Loop) IsOuterLoop() bool {
	if l.Face == nil {
		return false
	}
	return l.Face.LoopOut == l
}

// GetSolid returns the solid this loop belongs to.
func (l *Loop) GetSolid() *Solid {
	if l.Face == nil {
		return nil
	}
	return l.Face.Solid
}

// NHalfEdges returns the number of half-edges in the loop.
func (l *Loop) NHalfEdges() int {
	if l.HalfEdges == nil {
		return 0
	}
	n := 0
	l.ForEachHe(func(_ *HalfEdge) bool {
		n++
		return true
	})
	return n
}

// GetVertexLocations fills locs with vertex positions from this loop.
// Returns the number of vertices written.
func (l *Loop) GetVertexLocations(locs []vec.SFVec3f) int {
	if l.HalfEdges == nil {
		return 0
	}
	i := 0
	l.ForEachHe(func(he *HalfEdge) bool {
		if i >= len(locs) {
			return false
		}
		locs[i] = he.Vertex.Loc
		i++
		return true
	})
	return i
}
