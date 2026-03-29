package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/algorithms"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/boolop"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/export"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/primitives"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

var (
	yellow    = vec.SFColor{R: 1.0, G: 0.9, B: 0.2, A: 1}
	lightBlue = vec.SFColor{R: 0.4, G: 0.6, B: 1.0, A: 1}
	vizGreen  = vec.SFColor{R: 0.2, G: 0.8, B: 0.3, A: 1}
)

type vizCase struct {
	name  string
	makeA func() *base.Solid
	makeB func() *base.Solid
}

func setColor(s *base.Solid, c vec.SFColor) {
	s.ForEachFace(func(f *base.Face) bool {
		f.SetColor(c)
		return true
	})
}

func doTranslate(s *base.Solid, x, y, z float64) {
	s.TransformGeometry(vec.TranslationMatrix(x, y, z))
}

func doScale(s *base.Solid, sx, sy, sz float64) {
	s.TransformGeometry(vec.ScaleMatrix(sx, sy, sz))
}

func doRotateCenter(s *base.Solid, degrees float64, axis vec.SFVec3f) {
	mn, mx := s.Extents()
	cx := (mn.X + mx.X) / 2
	cy := (mn.Y + mx.Y) / 2
	cz := (mn.Z + mx.Z) / 2
	doTranslate(s, -cx, -cy, -cz)
	radians := degrees * math.Pi / 180.0
	rot := vec.SFRotation{X: axis.X, Y: axis.Y, Z: axis.Z, W: radians}
	s.TransformGeometry(vec.RotationMatrix(rot))
	doTranslate(s, cx, cy, cz)
}

func makeHexVerts(radius float64) []vec.SFVec3f {
	verts := make([]vec.SFVec3f, 6)
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3.0
		verts[i] = vec.SFVec3f{
			X: radius * math.Cos(angle),
			Y: radius * math.Sin(angle),
			Z: 0,
		}
	}
	return verts
}

func makeSweptLamina(verts []vec.SFVec3f, dir vec.SFVec3f, color vec.SFColor) *base.Solid {
	s := primitives.MakeLamina(verts, color)
	algorithms.TranslationalSweep(s, s.GetFirstFace(), dir)
	s.CalcPlaneEquations()
	s.Renumber()
	return s
}

func main() {
	outDir := filepath.Join("examples", "bool_demos")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	cases := []vizCase{
		{
			name:  "partial_penetration",
			makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
			makeB: func() *base.Solid {
				s := primitives.MakeCube(0.5, lightBlue)
				doScale(s, 1, 1, 2)
				doTranslate(s, 0.25, 0.25, -0.25)
				return s
			},
		},
		{
			name:  "fully_contained",
			makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
			makeB: func() *base.Solid {
				s := primitives.MakeCube(0.5, lightBlue)
				doTranslate(s, 0.25, 0.25, 0.25)
				return s
			},
		},
		{
			name:  "through",
			makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
			makeB: func() *base.Solid {
				s := primitives.MakeCube(0.5, lightBlue)
				doScale(s, 1, 1, 4)
				doTranslate(s, 0.25, 0.25, -0.15)
				return s
			},
		},
		{
			name:  "edge_on_edge",
			makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
			makeB: func() *base.Solid {
				s := primitives.MakeCube(1.0, lightBlue)
				doTranslate(s, 0, -1, -1)
				return s
			},
		},
		{
			name:  "rotated_elongated",
			makeA: func() *base.Solid { return primitives.MakeCube(1.0, yellow) },
			makeB: func() *base.Solid {
				s := primitives.MakeCube(0.5, lightBlue)
				doScale(s, 1, 1, 4)
				doTranslate(s, -0.25, -0.25, -0.5)
				doTranslate(s, 0.5, 0, -0.25)
				doRotateCenter(s, 55.0, vec.XAxis)
				doTranslate(s, 0, -0.07, 0)
				return s
			},
		},
		{
			name: "hexagon_prism",
			makeA: func() *base.Solid {
				return makeSweptLamina(makeHexVerts(0.8), vec.SFVec3f{X: 0, Y: 0, Z: -1.5}, yellow)
			},
			makeB: func() *base.Solid {
				s := makeSweptLamina(makeHexVerts(0.8), vec.SFVec3f{X: 0, Y: 0, Z: -1.5}, lightBlue)
				doTranslate(s, 0.5, 0.3, -0.4)
				return s
			},
		},
		{
			name:  "sphere_vs_cube",
			makeA: func() *base.Solid { return primitives.MakeSphere(1.0, 10, 10, yellow) },
			makeB: func() *base.Solid {
				s := primitives.MakeCube(1.0, lightBlue)
				doTranslate(s, 0, 0, -1)
				return s
			},
		},
	}

	ops := []struct {
		code int
		name string
	}{
		{base.BoolUnion, "union"},
		{base.BoolIntersection, "intersection"},
		{base.BoolDifference, "difference"},
	}

	for _, c := range cases {
		for _, op := range ops {
			dispA := c.makeA()
			dispB := c.makeB()
			setColor(dispA, yellow)
			setColor(dispB, lightBlue)

			opA := c.makeA()
			opB := c.makeB()
			result, ok := boolop.BoolOp(opA, opB, op.code)

			mnA, mxA := dispA.Extents()
			mnB, mxB := dispB.Extents()
			xmin := math.Min(mnA.X, mnB.X)
			xmax := math.Max(mxA.X, mxB.X)
			span := xmax - xmin
			if span < 1 {
				span = 1
			}
			gap := span * 0.8
			leftX := -(span/2 + gap/2)
			rightX := span/2 + gap/2

			solids := []*base.Solid{dispA, dispB}
			translations := []vec.SFVec3f{
				{X: leftX, Y: 0, Z: 0},
				{X: leftX, Y: 0, Z: 0},
			}

			status := "PASS"
			prefix := "pass"
			if ok && result != nil {
				errs := algorithms.VerifyDetailed(result)
				if len(errs) > 0 {
					status = "FAIL (invalid topology)"
					prefix = "fail"
				}
				setColor(result, vizGreen)
				solids = append(solids, result)
				translations = append(translations, vec.SFVec3f{X: rightX, Y: 0, Z: 0})
			} else {
				status = "FAIL (no result)"
				prefix = "fail"
			}

			fileName := fmt.Sprintf("%s_%s_%s.wrl", prefix, c.name, op.name)
			outPath := filepath.Join(outDir, fileName)
			f, err := os.Create(outPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %s: %v\n", outPath, err)
				continue
			}

			w := func(format string, args ...any) {
				if _, err := fmt.Fprintf(f, format, args...); err != nil {
					fmt.Fprintf(os.Stderr, "write error: %v\n", err)
				}
			}

			w("#VRML V2.0 utf8\n")
			w("# Bool Visualization: %s %s\n", c.name, strings.ToUpper(op.name))
			w("# Status: %s\n", status)
			w("# Left (yellow): input solids A and B\n")
			if ok && result != nil {
				w("# Right (green): boolean result\n")
			} else {
				w("# Right: (no result - operation failed)\n")
			}
			w("\n")
			w("NavigationInfo { type \"EXAMINE\" }\n")
			w("Viewpoint { position 0 0 6 description \"Front\" }\n")
			w("\n")

			labels := []string{"Input A", "Input B", "Result"}
			wireColors := []vec.SFColor{yellow, lightBlue, vizGreen}
			for i, s := range solids {
				tx := translations[i]
				w("# %s\n", labels[i])
				w("Transform {\n  translation %g %g %g\n  children [\n", tx.X, tx.Y, tx.Z)
				if i == 0 {
					if err := export.Shape(s, f, "    "); err != nil {
						fmt.Fprintf(os.Stderr, "export error: %v\n", err)
					}
				} else {
					if err := export.Wireframe(s, f, "    ", wireColors[i]); err != nil {
						fmt.Fprintf(os.Stderr, "export error: %v\n", err)
					}
				}
				w("  ]\n}\n")
				w("\n")
			}

			if err := f.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "close error: %v\n", err)
			}
			fmt.Printf("  %-55s %s\n", fileName, status)
		}
	}
}
