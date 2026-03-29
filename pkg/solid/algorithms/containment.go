package algorithms

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
)

const (
	NoHit           = 0
	FaceHit         = 1
	VertexHit       = 2
	EdgeHit         = 3
	InsideLoopNoHit = 4
)

type IntersectRecord struct {
	He   *base.HalfEdge
	Vert *base.Vertex
	Type int
}

func NewIntersectRecord() IntersectRecord {
	return IntersectRecord{Type: NoHit}
}

func (r *IntersectRecord) RecordHit(he *base.HalfEdge, v *base.Vertex, hitType int) {
	r.Type = hitType
	switch hitType {
	case FaceHit:
		r.He = nil
		r.Vert = nil
	case VertexHit:
		r.Vert = v
		r.He = nil
	case EdgeHit:
		r.He = he
		r.Vert = nil
	case InsideLoopNoHit:
		r.He = nil
		r.Vert = nil
	}
}

func (r *IntersectRecord) GetType() int { return r.Type }

func CheckForContainment(l *base.Loop, pt [3]float64, drop int) bool {
	if l.HalfEdges == nil {
		return false
	}

	var pgon [][2]float64
	l.ForEachHe(func(he *base.HalfEdge) bool {
		v := he.Vertex.Loc
		var p [2]float64
		switch drop {
		case 0:
			p[0] = v.Y
			p[1] = v.Z
		case 1:
			p[0] = v.X
			p[1] = v.Z
		default:
			p[0] = v.X
			p[1] = v.Y
		}
		pgon = append(pgon, p)
		return true
	})

	var pt2 [2]float64
	switch drop {
	case 0:
		pt2[0] = pt[1]
		pt2[1] = pt[2]
	case 1:
		pt2[0] = pt[0]
		pt2[1] = pt[2]
	default:
		pt2[0] = pt[0]
		pt2[1] = pt[1]
	}

	return base.CrossingsTest(pgon, pt2)
}

func BoundaryContains(l *base.Loop, v *base.Vertex, rec *IntersectRecord) bool {
	var found bool
	l.ForEachHe(func(he *base.HalfEdge) bool {
		if he.Vertex.Loc.Eq(v.Loc) {
			rec.RecordHit(nil, he.Vertex, VertexHit)
			found = true
			return false
		}
		return true
	})
	if found {
		return true
	}

	l.ForEachHe(func(he *base.HalfEdge) bool {
		vl := he.Vertex.Loc
		nl := he.Next.Vertex.Loc
		if base.ContainsOnEdge(vl.X, vl.Y, vl.Z, nl.X, nl.Y, nl.Z, v.Loc.X, v.Loc.Y, v.Loc.Z) {
			rec.RecordHit(he, nil, EdgeHit)
			found = true
			return false
		}
		return true
	})
	return found
}

func LoopContains(l *base.Loop, v *base.Vertex, drop int, rec *IntersectRecord) bool {
	if BoundaryContains(l, v, rec) {
		return true
	}
	return CheckForContainment(l, [3]float64{v.Loc.X, v.Loc.Y, v.Loc.Z}, drop)
}

func FaceContains(f *base.Face, v *base.Vertex, rec *IntersectRecord) bool {
	drop := base.GetDominantComp(f.Normal.X, f.Normal.Y, f.Normal.Z)

	if f.LoopOut == nil || !LoopContains(f.LoopOut, v, drop, rec) {
		return false
	}

	if rec.Type == VertexHit || rec.Type == EdgeHit {
		return true
	}

	for _, l := range f.Loops {
		if l == f.LoopOut {
			continue
		}
		if l.NHalfEdges() < 3 {
			continue
		}
		if LoopContains(l, v, drop, rec) {
			if rec.Type == VertexHit || rec.Type == EdgeHit {
				return true
			}
			rec.RecordHit(nil, nil, InsideLoopNoHit)
			return false
		}
	}

	rec.RecordHit(nil, nil, FaceHit)
	return true
}

func SolidContains(s *base.Solid, other *base.Solid) bool {
	for v := other.Verts; v != nil; v = v.Next {
		conclusive := true
		for f := s.Faces; f != nil; f = f.Next {
			d := f.GetDistance(v.Loc)
			if d > -1e-5 && d < 1e-5 {
				conclusive = false
			} else if d > 0.00001 {
				return false
			}
		}
		if conclusive {
			return true
		}
	}
	return true
}
