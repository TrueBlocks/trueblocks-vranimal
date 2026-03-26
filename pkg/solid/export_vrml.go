package solid

import (
	"fmt"
	"io"
	"os"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ExportVRML writes the solid as a VRML97 Shape with an IndexedFaceSet to w.
// Vertices are duplicated per face and explicit normals provided for flat shading.
func (s *Solid) ExportVRML(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "#VRML V2.0 utf8"); err != nil {
		return err
	}
	return s.exportVRMLShape(w, "")
}

// ExportVRMLFile writes the solid as a VRML97 file at the given path.
func (s *Solid) ExportVRMLFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.ExportVRML(f)
}

// ExportMultiVRML writes multiple solids into one VRML97 file.
// Each solid becomes a separate Transform/Shape, offset by the given
// translations (nil means origin for that solid).
func ExportMultiVRML(w io.Writer, solids []*Solid, translations []vec.SFVec3f) error {
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
		// Write shape body (skip the #VRML header from ExportVRML by writing inline)
		if err := s.exportVRMLShape(w, "    "); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, "  ]\n}"); err != nil {
			return err
		}
	}
	return nil
}

// ExportMultiVRMLFile writes multiple solids to a VRML97 file with translations.
func ExportMultiVRMLFile(path string, solids []*Solid, translations []vec.SFVec3f) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return ExportMultiVRML(f, solids, translations)
}

// ExportVRMLShape writes just the Shape { geometry IndexedFaceSet { ... } }
// block to w, prefixed by indent on each line. Useful for embedding a solid
// inside a hand-crafted VRML scene with animation nodes.
func (s *Solid) ExportVRMLShape(w io.Writer, indent string) error {
	return s.exportVRMLShape(w, indent)
}

// exportVRMLShape writes just the Shape { ... } block with a given indent prefix.
// Vertices are duplicated per face (not shared) to guarantee flat shading.
func (s *Solid) exportVRMLShape(w io.Writer, indent string) error {
	s.Renumber()

	// Build per-face data with duplicated vertices for flat shading.
	type faceRec struct {
		verts []vec.SFVec3f
		color vec.SFColor
	}
	var faces []faceRec
	for f := s.Faces; f != nil; f = f.Next {
		var verts []vec.SFVec3f
		if f.LoopOut != nil {
			f.LoopOut.ForEachHe(func(he *HalfEdge) bool {
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

	// coordIndex — sequential indices, each face gets its own vertices.
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

	// coord — duplicated per face.
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

	// color — one per face.
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

// ExportVRMLWireframe writes the solid as a VRML97 Shape with an IndexedLineSet.
// Each edge is drawn once. The color comes from the first face.
func (s *Solid) ExportVRMLWireframe(w io.Writer, indent string, color vec.SFColor) error {
	s.Renumber()

	// Collect unique vertices and build index map.
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

	// Collect unique edges (unordered pairs of vertex indices).
	type edgeKey struct{ a, b int }
	seen := make(map[edgeKey]bool)
	var edges [][2]int
	for f := s.Faces; f != nil; f = f.Next {
		if f.LoopOut == nil {
			continue
		}
		f.LoopOut.ForEachHe(func(he *HalfEdge) bool {
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

	// coordIndex
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

	// coord
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
