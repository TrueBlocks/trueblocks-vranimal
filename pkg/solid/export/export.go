package export

import (
	"fmt"
	"io"
	"os"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func VRML(s *base.Solid, w io.Writer) error {
	if _, err := fmt.Fprintln(w, "#VRML V2.0 utf8"); err != nil {
		return err
	}
	return Shape(s, w, "")
}

func VRMLFile(s *base.Solid, path string) (retErr error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := f.Close(); cErr != nil && retErr == nil {
			retErr = cErr
		}
	}()
	return VRML(s, f)
}

func MultiVRML(w io.Writer, solids []*base.Solid, translations []vec.SFVec3f) error {
	if _, err := fmt.Fprintln(w, "#VRML V2.0 utf8"); err != nil {
		return err
	}
	for i, s := range solids {
		tx, ty, tz := float64(0), float64(0), float64(0)
		if translations != nil && i < len(translations) {
			tx = translations[i].X
			ty = translations[i].Y
			tz = translations[i].Z
		}
		if _, err := fmt.Fprintf(w, "Transform {\n  translation %g %g %g\n  children [\n", tx, ty, tz); err != nil {
			return err
		}
		if err := Shape(s, w, "    "); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "  ]\n}"); err != nil {
			return err
		}
	}
	return nil
}

func MultiVRMLFile(path string, solids []*base.Solid, translations []vec.SFVec3f) (retErr error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := f.Close(); cErr != nil && retErr == nil {
			retErr = cErr
		}
	}()
	return MultiVRML(f, solids, translations)
}

func Shape(s *base.Solid, w io.Writer, indent string) error {
	s.Renumber()

	type faceRec struct {
		verts []vec.SFVec3f
		color vec.SFColor
	}
	var faces []faceRec
	for f := s.Faces; f != nil; f = f.Next {
		var verts []vec.SFVec3f
		if f.LoopOut != nil {
			f.LoopOut.ForEachHe(func(he *base.HalfEdge) bool {
				verts = append(verts, he.Vertex.Loc)
				return true
			})
		}
		clr := vec.SFColor{R: 0.8, G: 0.8, B: 0.8, A: 1}
		if f.Data != nil && f.Data.HasColor() {
			clr = f.Data.GetColor()
		}
		faces = append(faces, faceRec{verts: verts, color: clr})
	}

	p := func(format string, a ...any) error {
		_, err := fmt.Fprintf(w, indent+format+"\n", a...)
		return err
	}

	if err := p("Shape {"); err != nil {
		return err
	}
	matR, matG, matB := float64(0.8), float64(0.8), float64(0.8)
	if len(faces) > 0 {
		matR, matG, matB = faces[0].color.R, faces[0].color.G, faces[0].color.B
	}
	if err := p("  appearance Appearance { material Material { diffuseColor %g %g %g } }", matR, matG, matB); err != nil {
		return err
	}
	if err := p("  geometry IndexedFaceSet {"); err != nil {
		return err
	}
	if err := p("    solid FALSE"); err != nil {
		return err
	}
	if err := p("    colorPerVertex FALSE"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%s    coordIndex [ ", indent); err != nil {
		return err
	}
	vi := 0
	for i, fr := range faces {
		for range fr.verts {
			if _, err := fmt.Fprintf(w, "%d, ", vi); err != nil {
				return err
			}
			vi++
		}
		if i < len(faces)-1 {
			if _, err := fmt.Fprint(w, "-1, "); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprint(w, "-1 "); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(w, "]"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%s    coord Coordinate { point [ ", indent); err != nil {
		return err
	}
	first := true
	for _, fr := range faces {
		for _, v := range fr.verts {
			if !first {
				if _, err := fmt.Fprint(w, ", "); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(w, "%g %g %g", v.X, v.Y, v.Z); err != nil {
				return err
			}
			first = false
		}
	}
	if _, err := fmt.Fprintln(w, " ] }"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%s    color Color { color [ ", indent); err != nil {
		return err
	}
	for i, fr := range faces {
		sep := ", "
		if i == len(faces)-1 {
			sep = " "
		}
		if _, err := fmt.Fprintf(w, "%g %g %g%s", fr.color.R, fr.color.G, fr.color.B, sep); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "] }"); err != nil {
		return err
	}

	if err := p("  }"); err != nil {
		return err
	}
	return p("}")
}

func Wireframe(s *base.Solid, w io.Writer, indent string, color vec.SFColor) error {
	s.Renumber()

	type vertKey struct{ x, y, z float64 }
	vmap := make(map[vertKey]int)
	var verts []vec.SFVec3f
	vertIdx := func(v vec.SFVec3f) int {
		k := vertKey{v.X, v.Y, v.Z}
		if idx, ok := vmap[k]; ok {
			return idx
		}
		idx := len(verts)
		vmap[k] = idx
		verts = append(verts, v)
		return idx
	}

	type edgeKey struct{ a, b int }
	seen := make(map[edgeKey]bool)
	var edges [][2]int
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut == nil {
			continue
		}
		f.LoopOut.ForEachHe(func(he *base.HalfEdge) bool {
			a := vertIdx(he.Vertex.Loc)
			b := vertIdx(he.Next.Vertex.Loc)
			lo, hi := a, b
			if lo > hi {
				lo, hi = hi, lo
			}
			k := edgeKey{lo, hi}
			if !seen[k] {
				seen[k] = true
				edges = append(edges, [2]int{a, b})
			}
			return true
		})
	}

	p := func(format string, a ...any) error {
		_, err := fmt.Fprintf(w, indent+format+"\n", a...)
		return err
	}

	if err := p("Shape {"); err != nil {
		return err
	}
	if err := p("  appearance Appearance { material Material { emissiveColor %g %g %g } }", color.R, color.G, color.B); err != nil {
		return err
	}
	if err := p("  geometry IndexedLineSet {"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%s    coordIndex [ ", indent); err != nil {
		return err
	}
	for i, e := range edges {
		sep := ", "
		if i == len(edges)-1 {
			sep = " "
		}
		if _, err := fmt.Fprintf(w, "%d, %d, -1%s", e[0], e[1], sep); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "]"); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%s    coord Coordinate { point [ ", indent); err != nil {
		return err
	}
	for i, v := range verts {
		sep := ", "
		if i == len(verts)-1 {
			sep = " "
		}
		if _, err := fmt.Fprintf(w, "%g %g %g%s", v.X, v.Y, v.Z, sep); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "] }"); err != nil {
		return err
	}

	if err := p("  }"); err != nil {
		return err
	}
	return p("}")
}
