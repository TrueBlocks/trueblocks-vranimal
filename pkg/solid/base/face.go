package base

import "github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"

// Face represents a polygon face defined by loops of half-edges.
type Face struct {
	Solid   *Solid
	LoopOut *Loop
	Loops   []*Loop
	Normal  vec.SFVec3f
	D       float64
	Index   uint64
	Mark1   uint64
	Mark2   uint64
	Data    *ColorData
	Next    *Face
	Prev    *Face
}

// NewFace creates a face belonging to the given solid.
func NewFace(s *Solid, color vec.SFColor) *Face {
	f := &Face{Solid: s}
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
	return sum.Scale(1 / float64(n))
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
	var nx, ny, nz float64
	he := l.HalfEdges
	for {
		next := he.Next
		v1 := he.Vertex.Loc
		v2 := next.Vertex.Loc
		nx += (v1.Y - v2.Y) * (v1.Z + v2.Z)
		ny += (v1.Z - v2.Z) * (v1.X + v2.X)
		nz += (v1.X - v2.X) * (v1.Y + v2.Y)
		he = next
		if he == l.HalfEdges {
			break
		}
	}
	normal := vec.SFVec3f{X: nx, Y: ny, Z: nz}.Normalize()
	f.Normal = normal
	f.D = -normal.Dot(l.HalfEdges.Vertex.Loc)
	return true
}

// GetDistance returns the signed distance from a point to this face's plane.
func (f *Face) GetDistance(pt vec.SFVec3f) float64 {
	return f.Normal.Dot(pt) + f.D
}

// InvertNormal reverses the face normal direction.
func (f *Face) InvertNormal() {
	f.Normal = f.Normal.Negate()
	f.D = -f.D
}

// Area returns the area of the face's outer loop.
func (f *Face) Area() float64 {
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
func (f *Face) Marked1(m uint64) bool { return f.Mark1 == m }

// Marked2 returns true if mark2 equals m.
func (f *Face) Marked2(m uint64) bool { return f.Mark2 == m }

// GetFirstLoop returns the first loop (outer boundary).
func (f *Face) GetFirstLoop() *Loop {
	if len(f.Loops) == 0 {
		return nil
	}
	return f.Loops[0]
}

// GetSecondLoop returns the second loop (first hole), or nil.
func (f *Face) GetSecondLoop() *Loop {
	if len(f.Loops) < 2 {
		return nil
	}
	return f.Loops[1]
}

// IsDegenerate returns true if the face has zero area or all vertices coincide.
func (f *Face) IsDegenerate() bool {
	if f.LoopOut == nil || f.LoopOut.HalfEdges == nil {
		return true
	}
	v := f.LoopOut.HalfEdges.Vertex.Loc
	allSame := true
	for _, l := range f.Loops {
		l.ForEachHe(func(he *HalfEdge) bool {
			if he.Vertex.Loc != v {
				allSame = false
				return false
			}
			return true
		})
		if !allSame {
			break
		}
	}
	if allSame {
		return true
	}
	a := f.Area()
	return a < 1e-12 && a > -1e-12
}

// IsPlanar returns true if all vertices lie on the face's plane.
func (f *Face) IsPlanar() bool {
	for _, l := range f.Loops {
		ok := true
		l.ForEachHe(func(he *HalfEdge) bool {
			d := f.GetDistance(he.Vertex.Loc)
			if d > 1e-5 || d < -1e-5 {
				ok = false
				return false
			}
			return true
		})
		if !ok {
			return false
		}
	}
	return true
}

// GetNormal returns the face normal (read-only alias).
func (f *Face) GetNormal() vec.SFVec3f { return f.Normal }

// GetD returns the plane distance.
func (f *Face) GetD() float64 { return f.D }
