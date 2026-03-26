package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

var (
	red    = vec.SFColor{R: 0.9, G: 0.2, B: 0.2, A: 1}
	green  = vec.SFColor{R: 0.2, G: 0.8, B: 0.3, A: 1}
	blue   = vec.SFColor{R: 0.2, G: 0.3, B: 0.9, A: 1}
	yellow = vec.SFColor{R: 0.9, G: 0.9, B: 0.2, A: 1}
	cyan   = vec.SFColor{R: 0.2, G: 0.8, B: 0.8, A: 1}
	purple = vec.SFColor{R: 0.7, G: 0.2, B: 0.8, A: 1}
)

func main() {
	os.MkdirAll("examples/bool_demos", 0755)

	genDisjointUnion()
	genContainedIntersection()
	genContainedDifference()
	genBContainsAUnion()
	genInputsOverlapping()
	genInputsRotated()
	genInputsCubeSphere()

	fmt.Println("Generated 7 demo files in examples/bool_demos/")
}

func genDisjointUnion() {
	a := solid.MakeCube(1.0, red)
	b := solid.MakeCube(1.0, green)
	b.TransformGeometry(vec.TranslationMatrix(4.0, 0, 0))

	result, ok := solid.Union(a, b)
	if !ok || result == nil {
		log.Fatal("disjoint union failed")
	}
	must(result.ExportVRMLFile("examples/bool_demos/disjoint_union.wrl"))
	fmt.Println("  disjoint_union.wrl")
}

func genContainedIntersection() {
	a := solid.MakeCube(2.0, red)
	b := solid.MakeCube(0.5, green)

	result, ok := solid.Intersection(a, b)
	if !ok || result == nil {
		log.Fatal("contained intersection failed")
	}
	must(result.ExportVRMLFile("examples/bool_demos/contained_intersection.wrl"))
	fmt.Println("  contained_intersection.wrl")
}

func genContainedDifference() {
	a := solid.MakeCube(2.0, red)
	b := solid.MakeCube(0.5, green)

	result, ok := solid.Difference(a, b)
	if !ok || result == nil {
		log.Fatal("contained difference failed")
	}
	must(result.ExportVRMLFile("examples/bool_demos/contained_difference.wrl"))
	fmt.Println("  contained_difference.wrl")
}

func genBContainsAUnion() {
	a := solid.MakeCube(0.5, red)
	b := solid.MakeCube(2.0, green)

	result, ok := solid.Union(a, b)
	if !ok || result == nil {
		log.Fatal("b-contains-a union failed")
	}
	must(result.ExportVRMLFile("examples/bool_demos/b_contains_a_union.wrl"))
	fmt.Println("  b_contains_a_union.wrl")
}

func genInputsOverlapping() {
	a := solid.MakeCube(1.0, red)
	b := solid.MakeCube(1.0, blue)
	b.TransformGeometry(vec.TranslationMatrix(1.0, 0, 0))

	must(solid.ExportMultiVRMLFile(
		"examples/bool_demos/overlapping_cubes_input.wrl",
		[]*solid.Solid{a, b},
		nil,
	))
	fmt.Println("  overlapping_cubes_input.wrl")
}

func genInputsRotated() {
	a := solid.MakeCube(1.0, yellow)
	b := solid.MakeCube(1.0, purple)
	radians := 45.0 * math.Pi / 180.0
	rot := vec.SFRotation{X: 0, Y: 1, Z: 0, W: radians}
	b.TransformGeometry(vec.RotationMatrix(rot))

	must(solid.ExportMultiVRMLFile(
		"examples/bool_demos/rotated_cubes_input.wrl",
		[]*solid.Solid{a, b},
		nil,
	))
	fmt.Println("  rotated_cubes_input.wrl")
}

func genInputsCubeSphere() {
	a := solid.MakeCube(1.0, cyan)
	b := solid.MakeSphere(1.2, 16, 16, red)

	must(solid.ExportMultiVRMLFile(
		"examples/bool_demos/cube_sphere_input.wrl",
		[]*solid.Solid{a, b},
		nil,
	))
	fmt.Println("  cube_sphere_input.wrl")
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
