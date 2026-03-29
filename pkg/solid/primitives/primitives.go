package primitives

import (
	"math"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func MakeLamina(vertices []vec.SFVec3f, color vec.SFColor) *base.Solid {
	n := len(vertices)
	if n < 3 {
		return nil
	}

	s, startV, _ := euler.Mvfs(vertices[0], color)
	prev := startV
	for i := 1; i < n; i++ {
		nv, _, _ := euler.Lmev(prev.He, vertices[i])
		prev = nv
	}
	_, _, _ = euler.Lmef(prev.He, startV.He)
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func MakeCircle(cx, cy, radius, h float64, n int, color vec.SFColor) *base.Solid {
	if n < 3 {
		return nil
	}

	startLoc := vec.SFVec3f{X: cx + radius, Y: cy, Z: h}
	s, startV, _ := euler.Mvfs(startLoc, color)

	prev := startV
	for i := 1; i < n; i++ {
		a := float64(i) * 2 * math.Pi / float64(n)
		x := cx + radius*math.Cos(a)
		y := cy + radius*math.Sin(a)
		nv, _, _ := euler.Lmev(prev.He, vec.SFVec3f{X: x, Y: y, Z: h})
		prev = nv
	}

	_, _, _ = euler.Lmef(prev.He, startV.He)

	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func MakeTorus(majorR, minorR float64, majorSegs, minorSegs int, color vec.SFColor) *base.Solid {
	if majorSegs < 3 || minorSegs < 3 {
		return nil
	}

	s := MakeCircle(0, majorR, minorR, 0, minorSegs, color)
	if s == nil {
		return nil
	}

	RotationalSweep(s, majorSegs)
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func MakeCube(halfSize float64, color vec.SFColor) *base.Solid {
	h := halfSize
	s := MakeLamina([]vec.SFVec3f{
		{X: -h, Y: -h, Z: h},
		{X: h, Y: -h, Z: h},
		{X: h, Y: h, Z: h},
		{X: -h, Y: h, Z: h},
	}, color)
	if s == nil {
		return nil
	}
	TranslationalSweep(s, s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: -2 * h})
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func MakePrism(height float64, color vec.SFColor) *base.Solid {
	s := MakeLamina([]vec.SFVec3f{
		{X: 0, Y: 1, Z: height / 2},
		{X: 1, Y: -1, Z: height / 2},
		{X: -1, Y: -1, Z: height / 2},
	}, color)
	if s == nil {
		return nil
	}
	TranslationalSweep(s, s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: -height})
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func MakeCylinder(radius, height float64, n int, color vec.SFColor) *base.Solid {
	if n < 3 {
		return nil
	}

	s := MakeCircle(0, 0, radius, height/2, n, color)
	if s == nil {
		return nil
	}

	TranslationalSweep(s, s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: -height})
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func MakeSphere(radius float64, latSegs, lonSegs int, color vec.SFColor) *base.Solid {
	if latSegs < 3 || lonSegs < 3 {
		return nil
	}

	southPole := vec.SFVec3f{X: -radius, Y: 0, Z: 0}
	s, _, _ := euler.Mvfs(southPole, color)

	prev := s.Verts
	for i := 1; i < latSegs; i++ {
		a := float64(i) * math.Pi / float64(latSegs)
		loc := vec.SFVec3f{
			X: -radius * math.Cos(a),
			Y: radius * math.Sin(a),
			Z: 0,
		}
		nv, _, _ := euler.Lmev(prev.He, loc)
		prev = nv
	}
	northPole := vec.SFVec3f{X: radius, Y: 0, Z: 0}
	_, _, _ = euler.Lmev(prev.He, northPole)

	RotationalSweep(s, lonSegs)
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}
