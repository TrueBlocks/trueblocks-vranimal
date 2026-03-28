package base

import "math"

// FloatCompare returns -1, 0, or 1 comparing f to 0 within BigEps tolerance.
func FloatCompare(f float64) int {
	if f < -BigEps {
		return -1
	}
	if f > BigEps {
		return 1
	}
	return 0
}

// GetDominantComp returns the index (0=X, 1=Y, 2=Z) of the largest
// absolute component of a normal vector. Used to project 3D→2D.
func GetDominantComp(x, y, z float64) int {
	ax := math.Abs(x)
	ay := math.Abs(y)
	az := math.Abs(z)
	if ax >= ay && ax >= az {
		return 0
	}
	if ay >= az {
		return 1
	}
	return 2
}

// Collinear returns true if three points (given as x,y,z triples) are collinear.
func Collinear(ax, ay, az, bx, by, bz, cx, cy, cz float64) bool {
	abx, aby, abz := bx-ax, by-ay, bz-az
	acx, acy, acz := cx-ax, cy-ay, cz-az
	// cross product
	crossX := aby*acz - abz*acy
	crossY := abz*acx - abx*acz
	crossZ := abx*acy - aby*acx
	lenSq := crossX*crossX + crossY*crossY + crossZ*crossZ
	return lenSq < 1e-14 // 1e-7 squared
}

// IntersectsEdgeVertex tests if point (px,py,pz) lies on the line through
// (v1x,v1y,v1z)→(v2x,v2y,v2z). Returns the parametric t value and ok.
func IntersectsEdgeVertex(v1x, v1y, v1z, v2x, v2y, v2z, px, py, pz float64) (float64, bool) {
	rx, ry, rz := v2x-v1x, v2y-v1y, v2z-v1z
	lenSq := rx*rx + ry*ry + rz*rz
	if lenSq == 0 {
		if v1x == px && v1y == py && v1z == pz {
			return 0, true
		}
		return 0, false
	}
	dx, dy, dz := px-v1x, py-v1y, pz-v1z
	t := (rx*dx + ry*dy + rz*dz) / lenSq
	// reconstruct point at t and compare
	tx, ty, tz := v1x+rx*t, v1y+ry*t, v1z+rz*t
	if VecEq(tx, ty, tz, px, py, pz) {
		return t, true
	}
	return t, false
}

// ContainsOnEdge tests if point p lies on the segment v1→v2 (parametric t in [0,1]).
func ContainsOnEdge(v1x, v1y, v1z, v2x, v2y, v2z, px, py, pz float64) bool {
	t, ok := IntersectsEdgeVertex(v1x, v1y, v1z, v2x, v2y, v2z, px, py, pz)
	if !ok {
		return false
	}
	return t >= -0.00001 && t <= 1.00001
}

// IntersectsEdgeEdge tests if two 2D-projected edges intersect.
// Projects to 2D by dropping the coordinate at index "drop".
// Returns parametric values (t1, t2) and whether they intersect.
func IntersectsEdgeEdge(
	v1x, v1y, v1z, v2x, v2y, v2z,
	v3x, v3y, v3z, v4x, v4y, v4z float64,
	drop int,
) (t1, t2 float64, ok bool) {
	var a1, a2, b1, b2, c1, c2 float64
	switch drop {
	case 0:
		a1 = v2y - v1y
		a2 = v2z - v1z
		b1 = v3y - v4y
		b2 = v3z - v4z
		c1 = v1y - v3y
		c2 = v1z - v3z
	case 1:
		a1 = v2x - v1x
		a2 = v2z - v1z
		b1 = v3x - v4x
		b2 = v3z - v4z
		c1 = v1x - v3x
		c2 = v1z - v3z
	default:
		a1 = v2x - v1x
		a2 = v2y - v1y
		b1 = v3x - v4x
		b2 = v3y - v4y
		c1 = v1x - v3x
		c2 = v1y - v3y
	}
	d := a1*b2 - a2*b1
	if d > -1e-7 && d < 1e-7 {
		return 0, 0, false
	}
	t1 = (c2*b1 - c1*b2) / d
	t2 = (a2*c1 - a1*c2) / d
	return t1, t2, true
}

// CrossingsTest implements the Graphics Gems IV crossing number algorithm.
// Tests if a 2D point is inside a 2D polygon. Returns true if inside.
func CrossingsTest(pgon [][2]float64, point [2]float64) bool {
	numverts := len(pgon)
	if numverts == 0 {
		return false
	}
	tx := point[0]
	ty := point[1]
	vtx0 := pgon[numverts-1]
	yflag0 := vtx0[1] >= ty
	inside := false
	for j := 0; j < numverts; j++ {
		vtx1 := pgon[j]
		yflag1 := vtx1[1] >= ty
		if yflag0 != yflag1 {
			xflag0 := vtx0[0] >= tx
			if xflag0 == (vtx1[0] >= tx) {
				if xflag0 {
					inside = !inside
				}
			} else {
				crossX := vtx1[0] - (vtx1[1]-ty)*(vtx0[0]-vtx1[0])/(vtx0[1]-vtx1[1])
				if crossX >= tx {
					inside = !inside
				}
			}
		}
		yflag0 = yflag1
		vtx0 = vtx1
	}
	return inside
}

// VecEq returns true if two 3D points are approximately equal (within 1e-5).
func VecEq(ax, ay, az, bx, by, bz float64) bool {
	const eps = 1e-5
	dx, dy, dz := ax-bx, ay-by, az-bz
	return dx > -eps && dx < eps && dy > -eps && dy < eps && dz > -eps && dz < eps
}
