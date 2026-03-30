package main

import (
	"math"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/primitives"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

var (
	vrYellow  = vec.SFColor{R: 1.0, G: 1.0, B: 0.0, A: 1}
	vrBlue    = vec.SFColor{R: 0.0, G: 0.0, B: 1.0, A: 1}
	vrRed     = vec.SFColor{R: 1.0, G: 0.0, B: 0.0, A: 1}
	vrGreen   = vec.SFColor{R: 0.0, G: 0.5, B: 0.0, A: 1}
	vrDimGray = vec.SFColor{R: 0.41, G: 0.41, B: 0.41, A: 1}
	vrWhite   = vec.SFColor{R: 1.0, G: 1.0, B: 1.0, A: 1}
)

func solidTranslate(s *base.Solid, x, y, z float64) {
	s.TransformGeometry(vec.TranslationMatrix(x, y, z))
}

func solidScale(s *base.Solid, x, y, z float64) {
	s.TransformGeometry(vec.ScaleMatrix(x, y, z))
}

func solidRotateX(s *base.Solid, degrees float64) {
	r := vec.SFRotation{X: 1, Y: 0, Z: 0, W: degrees * math.Pi / 180}
	s.TransformGeometry(vec.RotationMatrix(r))
}

func solidRotateY(s *base.Solid, degrees float64) {
	r := vec.SFRotation{X: 0, Y: 1, Z: 0, W: degrees * math.Pi / 180}
	s.TransformGeometry(vec.RotationMatrix(r))
}

func solidRotateZ(s *base.Solid, degrees float64) {
	r := vec.SFRotation{X: 0, Y: 0, Z: 1, W: degrees * math.Pi / 180}
	s.TransformGeometry(vec.RotationMatrix(r))
}

func makePedroNose() *base.Solid {
	s := primitives.MakeCube(0.25, vrYellow)
	solidTranslate(s, -0.25, -0.25, 0)
	solidScale(s, 0.125, 0.75, 1)
	solidRotateX(s, -15)
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)
	return s
}

func makePedroLips() *base.Solid {
	verts := []vec.SFVec3f{
		{X: -1.0, Y: 0.8, Z: 0},
		{X: 1.0, Y: 0.8, Z: 0},
		{X: 2.0, Y: 1.0, Z: 0},
		{X: 1.0, Y: 0.0, Z: 0},
		{X: -1.0, Y: 0.0, Z: 0},
		{X: -2.0, Y: 1.0, Z: 0},
	}
	s := primitives.MakeLamina(verts, vrRed)
	primitives.TranslationalSweep(s, s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: 0.25})
	solidScale(s, 0.26, 0.125, 1)
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)
	return s
}

func makePedroHat() *base.Solid {
	verts := []vec.SFVec3f{
		{X: 1.9, Y: 0.0, Z: 0},
		{X: 2.0, Y: 0.2, Z: 0},
		{X: 2.6, Y: -3.2, Z: 0},
		{X: 4.6, Y: -3.2, Z: 0},
		{X: 5.0, Y: -2.8, Z: 0},
		{X: 4.6, Y: -3.3, Z: 0},
		{X: 2.5, Y: -3.3, Z: 0},
		{X: 1.9, Y: -0.1, Z: 0},
		{X: 0.1, Y: -0.1, Z: 0},
		{X: 0.1, Y: 0.0, Z: 0},
	}
	s := primitives.MakeLamina(verts, vrDimGray)
	solidRotateZ(s, 90)
	primitives.RotationalSweep(s, 30)
	solidRotateZ(s, -90)
	solidScale(s, 0.40, 0.40, 0.5)
	solidTranslate(s, 0, 2.0, -0.75)
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)
	return s
}

func MakePedro() *base.Solid {
	head := primitives.MakeSphere(1.0, 20, 20, vrYellow)
	head.CalcPlaneEquations()
	algorithms.Triangulate(head)
	solidScale(head, 1, 1.7, 1)

	eye1 := primitives.MakeSphere(0.18, 15, 15, vrBlue)
	eye1.CalcPlaneEquations()
	algorithms.Triangulate(eye1)
	solidTranslate(eye1, 0.3, 0.4, 0.8)

	eye2 := primitives.MakeSphere(0.18, 15, 15, vrBlue)
	eye2.CalcPlaneEquations()
	algorithms.Triangulate(eye2)
	solidTranslate(eye2, -0.3, 0.4, 0.8)

	lips := makePedroLips()
	solidTranslate(lips, 0, -0.4, 0.75)
	lips.SetColor(vrRed)

	nose := makePedroNose()
	solidTranslate(nose, 0, 0, 0.75)
	nose.SetColor(vrYellow)

	hat := makePedroHat()
	solidTranslate(hat, 0, 0, 0.70)
	hat.SetColor(vrDimGray)
	solidRotateX(hat, -14)
	solidScale(hat, 0.8, 0.8, 0.8)
	solidTranslate(hat, 0, 0.375, 0.09)

	eye1.Merge(eye2)
	eye1.Merge(lips)
	eye1.Merge(nose)
	eye1.Merge(hat)
	eye1.Merge(head)

	return eye1
}

func makeCarBody() *base.Solid {
	verts := []vec.SFVec3f{
		{X: 0.05, Y: 0.6, Z: 0},
		{X: 0.4, Y: 0.6, Z: 0},
		{X: 0.6, Y: 1.0, Z: 0},
		{X: 1.2, Y: 1.0, Z: 0},
		{X: 1.4, Y: 0.6, Z: 0},
		{X: 1.75, Y: 0.6, Z: 0},
		{X: 1.8, Y: 0.2, Z: 0},
		{X: 1.4, Y: 0.2, Z: 0},
		{X: 1.3, Y: 0.4, Z: 0},
		{X: 1.4, Y: 0.3, Z: 0},
		{X: 1.2, Y: 0.3, Z: 0},
		{X: 1.2, Y: 0.2, Z: 0},
		{X: 0.4, Y: 0.2, Z: 0},
		{X: 0.4, Y: 0.3, Z: 0},
		{X: 0.3, Y: 0.4, Z: 0},
		{X: 0.2, Y: 0.3, Z: 0},
		{X: 0.2, Y: 0.3, Z: 0},
		{X: 0.2, Y: 0.2, Z: 0},
		{X: 0.0, Y: 0.2, Z: 0},
	}
	s := primitives.MakeLamina(verts, vrRed)
	primitives.TranslationalSweep(s, s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: 0.9})
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)
	s.SetColor(vrRed)
	return s
}

func makeCarWindow() *base.Solid {
	verts := []vec.SFVec3f{
		{X: 0.45, Y: 0.6, Z: -0.02},
		{X: 0.65, Y: 0.95, Z: -0.02},
		{X: 1.15, Y: 0.95, Z: -0.02},
		{X: 1.35, Y: 0.6, Z: -0.02},
	}
	s := primitives.MakeLamina(verts, vrBlue)
	primitives.TranslationalSweep(s, s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: 0.96})
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)
	s.SetColor(vrBlue)
	return s
}

func makeCarWheel() *base.Solid {
	s := primitives.MakeCylinder(0.2, 0.2, 20, vrDimGray)
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)

	hub := primitives.MakeCylinder(0.1, 0.2, 20, vrWhite)
	hub.CalcPlaneEquations()
	algorithms.Triangulate(hub)
	solidTranslate(hub, 0, 0, -0.05)
	s.Merge(hub)
	return s
}

func MakeCar() *base.Solid {
	body := makeCarBody()
	win := makeCarWindow()
	body.Merge(win)

	w1 := makeCarWheel()
	solidTranslate(w1, 0.3, 0.2, -0.15)
	body.Merge(w1)

	w2 := makeCarWheel()
	solidRotateY(w2, 180)
	solidTranslate(w2, 0.3, 0.2, 1.05)
	body.Merge(w2)

	w3 := makeCarWheel()
	solidRotateY(w3, 180)
	solidTranslate(w3, 1.4, 0.2, 1.05)
	body.Merge(w3)

	w4 := makeCarWheel()
	solidTranslate(w4, 1.4, 0.2, -0.15)
	body.Merge(w4)

	solidTranslate(body, -1.0, -0.5, -0.5)
	solidScale(body, 2, 2, 2)
	return body
}

func MakeBevel() *base.Solid {
	verts := []vec.SFVec3f{
		{X: 0.000, Y: 0.000, Z: 0},
		{X: 0.300, Y: 0.000, Z: 0},
		{X: 0.200, Y: 0.050, Z: 0},
		{X: 0.200, Y: 0.075, Z: 0},
		{X: 0.225, Y: 0.100, Z: 0},
		{X: 0.075, Y: 0.100, Z: 0},
		{X: 0.100, Y: 0.075, Z: 0},
		{X: 0.100, Y: 0.050, Z: 0},
	}
	s := primitives.MakeLamina(verts, vrBlue)
	f1 := s.FindFace(1)
	if f1 == nil {
		f1 = s.Faces
	}
	primitives.TranslationalSweep(s, f1, vec.SFVec3f{X: 0, Y: 0, Z: 0.4})
	solidTranslate(s, -0.150, -0.050, 0)
	solidScale(s, 3, 3, 3)
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)
	s.Revert()
	s.SetColor(vrBlue)
	return s
}

func MakeRotSweep() *base.Solid {
	verts := []vec.SFVec3f{
		{X: 0, Y: 0.5, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 2, Z: 0},
		{X: 1, Y: 2, Z: 0},
		{X: 1, Y: 2.5, Z: 0},
		{X: 0.5, Y: 2.5, Z: 0},
		{X: 0.5, Y: 3, Z: 0},
		{X: 1.4, Y: 2.7, Z: 0},
		{X: 2, Y: 0.5, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 0.5, Z: 0},
	}
	s := primitives.MakeLamina(verts, vrGreen)
	primitives.RotationalSweep(s, 30)
	s.SetColor(vrGreen)
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)
	return s
}

func makeRoomWall() *base.Solid {
	s := primitives.MakeCube(0.5, vrWhite)
	solidScale(s, 1, 1, 0.1)
	solidRotateY(s, 90)
	solidTranslate(s, 1, 0, -0.5)
	solidScale(s, 1, 2, 2)
	return s
}

func makeRoomWalls() *base.Solid {
	s1 := makeRoomWall()

	s2 := makeRoomWall()
	solidScale(s2, 1, 1, 0.9)
	solidRotateY(s2, 90)

	s3 := makeRoomWall()
	solidRotateY(s3, 180)

	s4 := makeRoomWall()
	solidScale(s4, 1, 1, 0.9)
	solidRotateY(s4, -90)

	s1.Merge(s2)
	s1.Merge(s3)
	s1.Merge(s4)
	return s1
}

func makeRoomRoofSection() *base.Solid {
	s := primitives.MakeCircle(0, 0, 1.0, 0, 3, vrRed)
	solidRotateX(s, -90)
	solidTranslate(s, 0.5, 0, 0)
	solidRotateZ(s, 15)
	solidTranslate(s, -1.4488904476, 0, 0)
	return s
}

func makeRoomRoof() *base.Solid {
	left := makeRoomRoofSection()
	solidRotateY(left, -90)

	back := makeRoomRoofSection()
	solidRotateY(back, 90)

	right := makeRoomRoofSection()
	solidRotateY(right, 180)

	front := makeRoomRoofSection()
	solidRotateY(front, -90)

	left.Merge(back)
	left.Merge(right)
	left.Merge(front)
	return left
}

func makeRoomFloor() *base.Solid {
	s := primitives.MakeCube(0.5, vrWhite)
	solidScale(s, 1.9, 0.1, 1.9)
	solidTranslate(s, -0.95, 0, -0.95)
	return s
}

func MakeRoom() *base.Solid {
	roof := makeRoomRoof()
	solidTranslate(roof, 0, 2.0, 0)
	walls := makeRoomWalls()
	roof.Merge(walls)
	floor := makeRoomFloor()
	solidTranslate(floor, 0, 0.1, 0)
	roof.Merge(floor)
	solidTranslate(roof, 0, -0.75, 0)
	solidScale(roof, 1, 0.4, 1)
	return roof
}

func MakePlant() *base.Solid {
	verts := []vec.SFVec3f{
		{X: 0.05, Y: 0.6, Z: 0},
		{X: 0.4, Y: 0.6, Z: 0},
		{X: 0.6, Y: 1.0, Z: 0},
		{X: 1.2, Y: 1.0, Z: 0},
		{X: 1.4, Y: 0.6, Z: 0},
		{X: 1.75, Y: 0.6, Z: 0},
		{X: 1.8, Y: 0.2, Z: 0},
		{X: 1.4, Y: 0.2, Z: 0},
		{X: 1.3, Y: 0.4, Z: 0},
		{X: 1.4, Y: 0.3, Z: 0},
		{X: 1.2, Y: 0.3, Z: 0},
		{X: 1.2, Y: 0.2, Z: 0},
		{X: 0.4, Y: 0.2, Z: 0},
		{X: 0.4, Y: 0.3, Z: 0},
		{X: 0.3, Y: 0.4, Z: 0},
		{X: 0.2, Y: 0.3, Z: 0},
		{X: 0.2, Y: 0.3, Z: 0},
		{X: 0.2, Y: 0.2, Z: 0},
		{X: 0.0, Y: 0.2, Z: 0},
	}

	makeArm := func(yRot float64) *base.Solid {
		arm := primitives.MakeCircle(0, 0.2, 0.125, 0, 10, vrGreen)
		solidRotateY(arm, 90)
		algorithms.ArcSweep(arm, arm.Faces, verts)
		if yRot != 0 {
			solidRotateY(arm, yRot)
		}
		return arm
	}

	s := makeArm(0)
	s1 := makeArm(35)
	s2 := makeArm(70)
	s3 := makeArm(105)
	s4 := makeArm(145)
	s.Merge(s1)
	s.Merge(s2)
	s.Merge(s3)
	s.Merge(s4)
	s.SetColor(vrGreen)
	solidScale(s, 0.25, 0.75, 0.25)
	solidTranslate(s, 0, -2.5, 0)
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)
	s.SetColor(vrGreen)
	return s
}

func makeLetterPart(verts []vec.SFVec3f, depth float64, color vec.SFColor) *base.Solid {
	s := primitives.MakeLamina(verts, color)
	if s == nil {
		return nil
	}
	primitives.TranslationalSweep(s, s.Faces, vec.SFVec3f{X: 0, Y: 0, Z: depth})
	s.SetColor(color)
	s.CalcPlaneEquations()
	algorithms.Triangulate(s)
	return s
}

func makeLetterV(color vec.SFColor) *base.Solid {
	parts := [][]vec.SFVec3f{
		{{X: 0.00, Y: 1.00}, {X: 0.20, Y: 1.00}, {X: 0.60, Y: 0.00}, {X: 0.40, Y: 0.00}},
		{{X: 0.85, Y: 1.00}, {X: 1.05, Y: 1.00}, {X: 0.66, Y: 0.00}, {X: 0.56, Y: 0.25}},
	}
	var result *base.Solid
	for _, p := range parts {
		s := makeLetterPart(p, 0.15, color)
		if s == nil {
			continue
		}
		if result == nil {
			result = s
		} else {
			result.Merge(s)
		}
	}
	return result
}

func makeLetterR(color vec.SFColor) *base.Solid {
	parts := [][]vec.SFVec3f{
		{{X: .00, Y: 1}, {X: .20, Y: 1}, {X: .20, Y: 0}, {X: .00, Y: 0}},
		{{X: .25, Y: 1}, {X: .55, Y: 1}, {X: .50, Y: .8}, {X: .25, Y: .8}},
		{{X: .50, Y: .8}, {X: .55, Y: 1}, {X: .75, Y: .7}, {X: .70, Y: .7}},
		{{X: .25, Y: .6}, {X: .50, Y: .6}, {X: .55, Y: .4}, {X: .25, Y: .4}},
		{{X: .55, Y: .4}, {X: .50, Y: .6}, {X: .75, Y: .7}, {X: .70, Y: .65}},
		{{X: .25, Y: .35}, {X: .44, Y: .35}, {X: .79, Y: 0}, {X: .55, Y: 0}},
	}
	var result *base.Solid
	for _, p := range parts {
		s := makeLetterPart(p, 0.15, color)
		if s == nil {
			continue
		}
		if result == nil {
			result = s
		} else {
			result.Merge(s)
		}
	}
	return result
}

func makeLetterSmallA(color vec.SFColor) *base.Solid {
	parts := [][]vec.SFVec3f{
		{{X: .00, Y: .55}, {X: .35, Y: .6}, {X: .485, Y: .535}, {X: .3, Y: .47}},
		{{X: .35, Y: .45}, {X: .5, Y: .5}, {X: .5, Y: .15}, {X: .35, Y: .0}},
		{{X: .4, Y: -.0125}, {X: .5, Y: .12}, {X: .6, Y: .00}, {X: .5, Y: -.10}},
		{{X: .3, Y: .02}, {X: .25, Y: .02}, {X: .05, Y: .3}, {X: .15, Y: .3}, {X: .3, Y: .3}},
	}
	var result *base.Solid
	for _, p := range parts {
		s := makeLetterPart(p, 0.15, color)
		if s == nil {
			continue
		}
		if result == nil {
			result = s
		} else {
			result.Merge(s)
		}
	}
	return result
}

func makeLetterSmallN(color vec.SFColor) *base.Solid {
	parts := [][]vec.SFVec3f{
		{{X: .00, Y: .6}, {X: .20, Y: .6}, {X: .20, Y: 0}, {X: .00, Y: 0}},
		{{X: .25, Y: .6}, {X: .52, Y: .6}, {X: .25, Y: .4}},
		{{X: .35, Y: .39}, {X: .55, Y: .54}, {X: .55, Y: .0}, {X: .35, Y: .0}},
	}
	var result *base.Solid
	for _, p := range parts {
		s := makeLetterPart(p, 0.15, color)
		if s == nil {
			continue
		}
		if result == nil {
			result = s
		} else {
			result.Merge(s)
		}
	}
	return result
}

func makeLetterSmallI(color vec.SFColor) *base.Solid {
	verts := []vec.SFVec3f{{X: .00, Y: .6}, {X: .20, Y: .6}, {X: .20, Y: 0}, {X: .00, Y: 0}}
	return makeLetterPart(verts, 0.15, color)
}

func makeLetterM(color vec.SFColor) *base.Solid {
	parts := [][]vec.SFVec3f{
		{{X: .00, Y: 1}, {X: .20, Y: 1}, {X: .20, Y: 0}, {X: .00, Y: 0}},
		{{X: .25, Y: 1}, {X: .75, Y: .225}, {X: .60, Y: .1}, {X: .40, Y: .1}, {X: .25, Y: .225}},
		{{X: .75, Y: 1}, {X: .75, Y: .225}, {X: .50, Y: .6}},
		{{X: .80, Y: 1}, {X: 1.0, Y: 1}, {X: 1.0, Y: 0}, {X: .80, Y: 0}},
	}
	var result *base.Solid
	for _, p := range parts {
		s := makeLetterPart(p, 0.15, color)
		if s == nil {
			continue
		}
		if result == nil {
			result = s
		} else {
			result.Merge(s)
		}
	}
	return result
}

func makeLetterL(color vec.SFColor) *base.Solid {
	parts := [][]vec.SFVec3f{
		{{X: .00, Y: 1}, {X: .20, Y: 1}, {X: .20, Y: 0}, {X: .00, Y: 0}},
		{{X: .25, Y: .1}, {X: .70, Y: .1}, {X: .70, Y: 0}, {X: .25, Y: 0}},
	}
	var result *base.Solid
	for _, p := range parts {
		s := makeLetterPart(p, 0.15, color)
		if s == nil {
			continue
		}
		if result == nil {
			result = s
		} else {
			result.Merge(s)
		}
	}
	return result
}

type logoDef struct {
	make func() *base.Solid
	xOff float64
	axis vec.SFVec3f
}

var logoColor = vec.SFColor{R: 0.75, G: 0.5, B: 1.0, A: 1}

var logoDefs = []logoDef{
	{func() *base.Solid { return makeLetterV(logoColor) }, -3.0, vec.SFVec3f{X: 0.3, Y: 0.7, Z: 0.5}},
	{func() *base.Solid { return makeLetterR(logoColor) }, -1.9, vec.SFVec3f{X: 0.8, Y: 0.2, Z: 0.4}},
	{func() *base.Solid { return makeLetterSmallA(logoColor) }, -1.1, vec.SFVec3f{X: 0.1, Y: 0.9, Z: 0.3}},
	{func() *base.Solid { return makeLetterSmallN(logoColor) }, -0.4, vec.SFVec3f{X: 0.6, Y: 0.4, Z: 0.7}},
	{func() *base.Solid { return makeLetterSmallI(logoColor) }, 0.3, vec.SFVec3f{X: 0.5, Y: 0.5, Z: 0.2}},
	{func() *base.Solid { return makeLetterM(logoColor) }, 0.6, vec.SFVec3f{X: 0.2, Y: 0.8, Z: 0.6}},
	{func() *base.Solid { return makeLetterL(logoColor) }, 1.7, vec.SFVec3f{X: 0.7, Y: 0.3, Z: 0.8}},
}

type solidDef struct {
	name string
	make func() *base.Solid
}

var solidDefs = []solidDef{
	{"Pedro", MakePedro},
	{"Car", MakeCar},
	{"Bevel", MakeBevel},
	{"RotSweep", MakeRotSweep},
	{"Room", MakeRoom},
	{"Plant", MakePlant},
}
