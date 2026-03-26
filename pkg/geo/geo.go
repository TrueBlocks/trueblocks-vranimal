// Package geo provides geometric primitive types: Ray, Plane, BoundingBox, Rect2D.
// Ported from vraniml/src/utils/geometry/.
package geo

import (
	"math"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Ray — directed line segment (origin + direction)
// ---------------------------------------------------------------------------

// Ray represents a directed line segment with an origin and direction.
type Ray struct {
	Loc vec.SFVec3f // Origin
	Dir vec.SFVec3f // Direction (and magnitude)
}

// NewRay creates a ray from an origin and direction.
func NewRay(loc, dir vec.SFVec3f) Ray { return Ray{Loc: loc, Dir: dir} }

// Evaluate returns the point at parameter t: Loc + t*Dir.
func (r Ray) Evaluate(t float64) vec.SFVec3f {
	return r.Loc.Add(r.Dir.Scale(t))
}

// GetDistance returns the distance from Loc to the point at parameter t.
func (r Ray) GetDistance(t float64) float64 {
	return r.Dir.Scale(t).Length()
}

// Extrapolate returns a ray from the current endpoint extending in Dir.
func (r Ray) Extrapolate() Ray {
	return Ray{Loc: r.Evaluate(1), Dir: r.Dir}
}

// Interpolate returns the midpoint of the ray segment.
func (r Ray) Interpolate() vec.SFVec3f {
	return r.Evaluate(0.5)
}

// ReflectRay reflects the ray direction about a plane normal.
func (r Ray) ReflectRay(normal vec.SFVec3f) Ray {
	d := r.Dir.Normalize()
	reflected := d.Sub(normal.Scale(2 * d.Dot(normal)))
	return Ray{Loc: r.Loc, Dir: reflected}
}

// ApplyTransform transforms location and direction by a matrix.
func (r Ray) ApplyTransform(m vec.Matrix) Ray {
	return Ray{
		Loc: m.TransformPoint(r.Loc),
		Dir: m.TransformDirection(r.Dir),
	}
}

// ---------------------------------------------------------------------------
// Plane — defined by normal vector and distance from origin
// ---------------------------------------------------------------------------

// Plane represents a 3D plane as a normal vector and signed distance.
type Plane struct {
	Normal vec.SFVec3f
	D      float64
}

// NewPlane creates a plane from a normal and distance.
func NewPlane(normal vec.SFVec3f, d float64) Plane {
	return Plane{Normal: normal, D: d}
}

// NewPlaneFromPoints creates a plane from three non-collinear points.
func NewPlaneFromPoints(a, b, c vec.SFVec3f) Plane {
	ab := b.Sub(a)
	ac := c.Sub(a)
	n := ab.Cross(ac).Normalize()
	return Plane{Normal: n, D: -n.Dot(a)}
}

// GetDistance returns the signed distance from a point to the plane.
func (p Plane) GetDistance(pt vec.SFVec3f) float64 {
	return p.Normal.Dot(pt) + p.D
}

// IntersectRay returns the parameter t where the ray intersects the plane,
// and whether the intersection is valid (not parallel).
func (p Plane) IntersectRay(r Ray) (t float64, ok bool) {
	denom := p.Normal.Dot(r.Dir)
	if denom == 0 {
		return 0, false
	}
	t = -(p.Normal.Dot(r.Loc) + p.D) / denom
	return t, true
}

// IntersectPlane returns the line of intersection of two planes as a ray,
// and whether the planes are not parallel.
func (p Plane) IntersectPlane(other Plane) (Ray, bool) {
	dir := p.Normal.Cross(other.Normal)
	if dir.Length() < 1e-7 {
		return Ray{}, false
	}
	dir = dir.Normalize()
	d1 := p.D
	d2 := other.D
	n1n2 := p.Normal.Dot(other.Normal)
	det := float64(1) - n1n2*n1n2
	if det == 0 {
		return Ray{}, false
	}
	c1 := (-d1 + d2*n1n2) / det
	c2 := (-d2 + d1*n1n2) / det
	pt := p.Normal.Scale(c1).Add(other.Normal.Scale(c2))
	return Ray{Loc: pt, Dir: dir}, true
}

// XIntercept returns the X coordinate where the plane crosses the X-axis.
func (p Plane) XIntercept() float64 {
	if p.Normal.X == 0 {
		return 0
	}
	return -p.D / p.Normal.X
}

// YIntercept returns the Y coordinate where the plane crosses the Y-axis.
func (p Plane) YIntercept() float64 {
	if p.Normal.Y == 0 {
		return 0
	}
	return -p.D / p.Normal.Y
}

// ZIntercept returns the Z coordinate where the plane crosses the Z-axis.
func (p Plane) ZIntercept() float64 {
	if p.Normal.Z == 0 {
		return 0
	}
	return -p.D / p.Normal.Z
}

// ---------------------------------------------------------------------------
// BoundingBox — axis-aligned bounding box
// ---------------------------------------------------------------------------

// BoundingBox is an axis-aligned bounding box defined by min/max corners.
type BoundingBox struct {
	Min, Max vec.SFVec3f
}

// DefaultBBox returns an "invalid" bounding box (min > max) used as initial state.
func DefaultBBox() BoundingBox {
	big := float64(math.MaxFloat32)
	return BoundingBox{
		Min: vec.SFVec3f{X: big, Y: big, Z: big},
		Max: vec.SFVec3f{X: -big, Y: -big, Z: -big},
	}
}

// NewBBox creates a bounding box from min and max corners.
func NewBBox(min, max vec.SFVec3f) BoundingBox {
	return BoundingBox{Min: min, Max: max}
}

// IsDefault returns true if the bounding box has not been expanded.
func (b BoundingBox) IsDefault() bool {
	return b.Min.X > b.Max.X
}

// Center returns the centroid of the bounding box.
func (b BoundingBox) Center() vec.SFVec3f {
	return b.Min.Add(b.Max).Scale(0.5)
}

// Size returns the dimensions of the bounding box.
func (b BoundingBox) Size() vec.SFVec3f {
	return b.Max.Sub(b.Min)
}

// Include expands the bounding box to contain the given point.
func (b *BoundingBox) Include(pt vec.SFVec3f) {
	if pt.X < b.Min.X {
		b.Min.X = pt.X
	}
	if pt.Y < b.Min.Y {
		b.Min.Y = pt.Y
	}
	if pt.Z < b.Min.Z {
		b.Min.Z = pt.Z
	}
	if pt.X > b.Max.X {
		b.Max.X = pt.X
	}
	if pt.Y > b.Max.Y {
		b.Max.Y = pt.Y
	}
	if pt.Z > b.Max.Z {
		b.Max.Z = pt.Z
	}
}

// IncludeBox expands the bounding box to contain another bounding box.
func (b *BoundingBox) IncludeBox(other BoundingBox) {
	b.Include(other.Min)
	b.Include(other.Max)
}

// IsInside returns true if a point is inside the box.
func (b BoundingBox) IsInside(pt vec.SFVec3f) bool {
	return pt.X >= b.Min.X && pt.X <= b.Max.X &&
		pt.Y >= b.Min.Y && pt.Y <= b.Max.Y &&
		pt.Z >= b.Min.Z && pt.Z <= b.Max.Z
}

// Intersect tests ray-box intersection, returning (t, hit).
func (b BoundingBox) Intersect(r Ray) (float64, bool) {
	var tmin, tmax float64
	if r.Dir.X != 0 {
		tx1 := (b.Min.X - r.Loc.X) / r.Dir.X
		tx2 := (b.Max.X - r.Loc.X) / r.Dir.X
		if tx1 > tx2 {
			tx1, tx2 = tx2, tx1
		}
		tmin = tx1
		tmax = tx2
	} else {
		if r.Loc.X < b.Min.X || r.Loc.X > b.Max.X {
			return 0, false
		}
		tmin = float64(-math.MaxFloat32)
		tmax = float64(math.MaxFloat32)
	}
	if r.Dir.Y != 0 {
		ty1 := (b.Min.Y - r.Loc.Y) / r.Dir.Y
		ty2 := (b.Max.Y - r.Loc.Y) / r.Dir.Y
		if ty1 > ty2 {
			ty1, ty2 = ty2, ty1
		}
		if ty1 > tmin {
			tmin = ty1
		}
		if ty2 < tmax {
			tmax = ty2
		}
	} else if r.Loc.Y < b.Min.Y || r.Loc.Y > b.Max.Y {
		return 0, false
	}
	if r.Dir.Z != 0 {
		tz1 := (b.Min.Z - r.Loc.Z) / r.Dir.Z
		tz2 := (b.Max.Z - r.Loc.Z) / r.Dir.Z
		if tz1 > tz2 {
			tz1, tz2 = tz2, tz1
		}
		if tz1 > tmin {
			tmin = tz1
		}
		if tz2 < tmax {
			tmax = tz2
		}
	} else if r.Loc.Z < b.Min.Z || r.Loc.Z > b.Max.Z {
		return 0, false
	}
	if tmin > tmax || tmax < 0 {
		return 0, false
	}
	if tmin < 0 {
		return tmax, true
	}
	return tmin, true
}

// Union returns the smallest box that contains both boxes.
func Union(a, b BoundingBox) BoundingBox {
	r := a
	r.IncludeBox(b)
	return r
}

// Overlap returns true if two bounding boxes overlap.
func Overlap(a, b BoundingBox) bool {
	return a.Min.X <= b.Max.X && a.Max.X >= b.Min.X &&
		a.Min.Y <= b.Max.Y && a.Max.Y >= b.Min.Y &&
		a.Min.Z <= b.Max.Z && a.Max.Z >= b.Min.Z
}

// TransformBox transforms a bounding box by a matrix.
func TransformBox(b BoundingBox, m vec.Matrix) BoundingBox {
	corners := [8]vec.SFVec3f{
		{X: b.Min.X, Y: b.Min.Y, Z: b.Min.Z},
		{X: b.Max.X, Y: b.Min.Y, Z: b.Min.Z},
		{X: b.Min.X, Y: b.Max.Y, Z: b.Min.Z},
		{X: b.Max.X, Y: b.Max.Y, Z: b.Min.Z},
		{X: b.Min.X, Y: b.Min.Y, Z: b.Max.Z},
		{X: b.Max.X, Y: b.Min.Y, Z: b.Max.Z},
		{X: b.Min.X, Y: b.Max.Y, Z: b.Max.Z},
		{X: b.Max.X, Y: b.Max.Y, Z: b.Max.Z},
	}
	result := DefaultBBox()
	for _, c := range corners {
		result.Include(m.TransformPoint(c))
	}
	return result
}

// SurfaceArea returns the surface area of the bounding box.
func (b BoundingBox) SurfaceArea() float64 {
	s := b.Size()
	return 2 * (s.X*s.Y + s.Y*s.Z + s.Z*s.X)
}

// ---------------------------------------------------------------------------
// Rect2D — 2D screen rectangle
// ---------------------------------------------------------------------------

// Rect2D represents a 2D screen rectangle.
type Rect2D struct {
	X, Y, W, H int64
}

// NewRect2D creates a 2D rectangle.
func NewRect2D(x, y, w, h int64) Rect2D { return Rect2D{x, y, w, h} }
