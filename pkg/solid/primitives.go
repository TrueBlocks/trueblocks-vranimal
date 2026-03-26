package solid

import (
	"math"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// MakeLamina constructs a flat polygonal sheet (lamina) from the given vertices.
// A lamina has exactly 2 faces: front and back. It requires at least 3 vertices.
func MakeLamina(vertices []vec.SFVec3f, color vec.SFColor) *Solid {
	n := len(vertices)
	if n < 3 {
		return nil
	}

	s, startV, _ := Mvfs(vertices[0], color)
	prev := startV
	for i := 1; i < n; i++ {
		nv, _ := Mev(s, prev, vertices[i])
		prev = nv
	}
	Mef(s, prev, startV)
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

// MakeCircle constructs a circular lamina (disc) centered at (cx, cy) with the
// given radius at height h, approximated by n segments. n must be >= 3.
// This matches the C++ vrSolidCircle constructor.
func MakeCircle(cx, cy, radius, h float64, n int, color vec.SFColor) *Solid {
	if n < 3 {
		return nil
	}

	// First vertex at angle=0: (cx+radius, cy, h)
	startLoc := vec.SFVec3f{X: cx + radius, Y: cy, Z: h}
	s, startV, _ := Mvfs(startLoc, color)

	// Create n-1 additional vertices around the circle, chaining from each previous.
	prev := startV
	for i := 1; i < n; i++ {
		a := float64(i) * 2 * math.Pi / float64(n)
		x := cx + radius*float64(math.Cos(a))
		y := cy + radius*float64(math.Sin(a))
		nv, _ := Mev(s, prev, vec.SFVec3f{X: x, Y: y, Z: h})
		prev = nv
	}

	// Close the polygon: connect last vertex back to first.
	Mef(s, prev, startV)

	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

// MakeTorus constructs a torus (donut) by revolving a circular cross-section
// around the X axis. majorR is the distance from the center of the tube to the
// center of the torus. minorR is the radius of the tube. majorSegs and minorSegs
// control the tessellation. Both must be >= 3.
func MakeTorus(majorR, minorR float64, majorSegs, minorSegs int, color vec.SFColor) *Solid {
	if majorSegs < 3 || minorSegs < 3 {
		return nil
	}

	// Build circular cross-section as a wire/lamina in the YZ plane at X=0.
	// The circle is centered at (0, majorR) in the YZ plane with radius minorR.
	// SolidCircle(cx=0, cy=majorR, rad=minorR, h=0, n=minorSegs)
	// Starting vertex is at (0, majorR+minorR, 0) — note: Arc uses (x, y, height)
	// which maps to (X=height, Y=x, Z=y) for the rotation sweep around X.
	//
	// Actually, the C++ code puts the circle in the XY plane at height h=0.
	// SolidCircle(0, rd1, rd2, 0, nf2) means cx=0, cy=rd1 (major radius),
	// rad=rd2 (minor radius), h=0, n=nf2. The circle lies in Y with Z=h=0.
	// Then RotationalSweep rotates around X axis.
	//
	// But our Arc uses (cx, cy) as 2D center and height as the Z coord of the
	// 3D vertex {X: x, Y: y, Z: height}. The sweep rotates around X.
	// So we need to put the circle in the XY plane: vertices at
	// {X: cx + cos(a)*r, Y: cy + sin(a)*r, Z: 0}.
	// The sweep rotates Y→Z, which makes the torus.

	// Build circle cross-section (this creates a lamina)
	s := MakeCircle(0, majorR, minorR, 0, minorSegs, color)
	if s == nil {
		return nil
	}

	// Revolve around X axis to form the torus
	s.RotationalSweep(majorSegs)
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

// MakeCube constructs a cube centered at the origin with the given half-size.
func MakeCube(halfSize float64, color vec.SFColor) *Solid {
	h := halfSize
	positions := []vec.SFVec3f{
		{X: -h, Y: h, Z: h},
		{X: h, Y: h, Z: h},
		{X: h, Y: -h, Z: h},
		{X: -h, Y: -h, Z: h},
	}
	indices := []int64{0, 1, 2, 3, -1}
	s := BuildFromIndexSet(positions, indices, color)
	if s == nil {
		return nil
	}
	s.TranslationalSweep(s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: -2 * h})
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

// MakePrism constructs a triangular prism by sweeping a triangle along Z.
func MakePrism(height float64, color vec.SFColor) *Solid {
	positions := []vec.SFVec3f{
		{X: 0, Y: 1, Z: height / 2},
		{X: 1, Y: -1, Z: height / 2},
		{X: -1, Y: -1, Z: height / 2},
	}
	indices := []int64{0, 1, 2, -1}
	s := BuildFromIndexSet(positions, indices, color)
	if s == nil {
		return nil
	}
	s.TranslationalSweep(s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: -height})
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

// MakeCylinder constructs a cylinder of given radius and height, centered at origin.
// The cylinder axis is along Z. n controls the number of circumferential segments.
func MakeCylinder(radius, height float64, n int, color vec.SFColor) *Solid {
	if n < 3 {
		return nil
	}

	// Build circle cross-section at z = +height/2 (normal points +Z).
	// Sweep in -Z so side faces point outward.
	s := MakeCircle(0, 0, radius, height/2, n, color)
	if s == nil {
		return nil
	}

	s.TranslationalSweep(s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: -height})
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

// MakeSphere constructs an approximated sphere using rotational sweep of a
// semicircular arc. latSegs and lonSegs control tessellation. Both must be >= 3.
func MakeSphere(radius float64, latSegs, lonSegs int, color vec.SFColor) *Solid {
	if latSegs < 3 || lonSegs < 3 {
		return nil
	}

	// Build an open semicircular wire in the XY plane from south pole to north pole.
	// RotationalSweep rotates around X: Y → Y*cos - Z*sin, Z → Y*sin + Z*cos.
	// Poles must lie on the X axis so they collapse to single vertices after sweep.
	// South pole at (-radius, 0, 0), north pole at (radius, 0, 0).
	// Intermediate vertices trace a semicircle: X = -radius*cos(a), Y = radius*sin(a)
	// for a going from 0 to pi in latSegs steps.

	southPole := vec.SFVec3f{X: -radius, Y: 0, Z: 0}
	s, _, _ := Mvfs(southPole, color)

	prev := s.Verts
	for i := 1; i < latSegs; i++ {
		a := float64(i) * math.Pi / float64(latSegs)
		loc := vec.SFVec3f{
			X: -radius * float64(math.Cos(a)),
			Y: radius * float64(math.Sin(a)),
			Z: 0,
		}
		nv, _ := Lmev(prev.He, loc)
		prev = nv
	}
	// North pole
	northPole := vec.SFVec3f{X: radius, Y: 0, Z: 0}
	Lmev(prev.He, northPole)

	s.RotationalSweep(lonSegs)
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}
