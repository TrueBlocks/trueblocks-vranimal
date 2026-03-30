package converter

import (
	"image"
	"image/draw"
	"strings"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/text"
	"github.com/g3n/engine/texture"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
)

func convertText(t *node.Text, app *node.Appearance, parent *core.Node, baseDir string) {
	if len(t.String) == 0 {
		return
	}

	size := 1.0
	spacing := 1.0
	family := "SERIF"
	style := "PLAIN"
	justify := "BEGIN"
	horizontal := true

	if t.FontStyle != nil {
		size = t.FontStyle.Size
		spacing = t.FontStyle.Spacing
		family = t.FontStyle.Family
		style = t.FontStyle.Style
		horizontal = t.FontStyle.Horizontal
		if len(t.FontStyle.Justify) > 0 {
			justify = t.FontStyle.Justify[0]
		}
	}

	fontPath := resolveFontPath(family, style)
	fnt, err := text.NewFont(fontPath)
	if err != nil {
		fnt = nil
	}

	if fnt == nil {
		convertTextFallback(t, app, parent, baseDir, size)
		return
	}

	pointSize := size * 72
	if pointSize < 8 {
		pointSize = 8
	}
	fnt.SetPointSize(pointSize)
	fnt.SetLineSpacing(spacing)
	fnt.SetColor(&math32.Color4{R: 1, G: 1, B: 1, A: 1})

	joined := strings.Join(t.String, "\n")

	imgW, imgH := fnt.MeasureText(joined)
	if imgW <= 0 || imgH <= 0 {
		return
	}

	pad := 4
	imgW += pad * 2
	imgH += pad * 2

	bgImg := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	draw.Draw(bgImg, bgImg.Bounds(), image.Transparent, image.Point{}, draw.Src)

	fnt.DrawTextOnImage(joined, pad, pad, bgImg)

	tex := texture.NewTexture2DFromRGBA(bgImg)
	tex.SetMagFilter(gls.LINEAR)
	tex.SetMinFilter(gls.LINEAR)

	mat := material.NewStandard(math32.NewColor("white"))
	mat.AddTexture(tex)
	mat.SetTransparent(true)
	mat.SetSide(material.SideFront)

	if app != nil && app.Material != nil {
		c := app.Material.DiffuseColor
		mat.SetColor(toColor(&c))
	}

	worldW := size * float64(imgW) / pointSize
	worldH := size * float64(imgH) / pointSize

	var geom *geometry.Geometry
	if horizontal {
		geom = buildTextQuad(float32(worldW), float32(worldH))
	} else {
		geom = buildTextQuadVertical(float32(worldW), float32(worldH))
	}

	mesh := graphic.NewMesh(geom, mat)
	mesh.SetName(t.GetName())

	applyTextJustify(mesh, justify, float32(worldW), float32(worldH), horizontal)

	parent.Add(mesh)
}

func buildTextQuad(w, h float32) *geometry.Geometry {
	positions := math32.NewArrayF32(0, 12)
	positions.Append(0, 0, 0)
	positions.Append(w, 0, 0)
	positions.Append(w, h, 0)
	positions.Append(0, h, 0)

	uvs := math32.NewArrayF32(0, 8)
	uvs.Append(0, 1)
	uvs.Append(1, 1)
	uvs.Append(1, 0)
	uvs.Append(0, 0)

	indices := math32.NewArrayU32(0, 6)
	indices.Append(0, 1, 2, 0, 2, 3)

	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(uvs).AddAttrib(gls.VertexTexcoord))
	geom.SetIndices(indices)
	return geom
}

func buildTextQuadVertical(w, h float32) *geometry.Geometry {
	positions := math32.NewArrayF32(0, 12)
	positions.Append(0, 0, 0)
	positions.Append(0, 0, -h)
	positions.Append(w, 0, -h)
	positions.Append(w, 0, 0)

	uvs := math32.NewArrayF32(0, 8)
	uvs.Append(0, 0)
	uvs.Append(0, 1)
	uvs.Append(1, 1)
	uvs.Append(1, 0)

	indices := math32.NewArrayU32(0, 6)
	indices.Append(0, 1, 2, 0, 2, 3)

	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(uvs).AddAttrib(gls.VertexTexcoord))
	geom.SetIndices(indices)
	return geom
}

func applyTextJustify(mesh *graphic.Mesh, justify string, w, h float32, horizontal bool) {
	switch justify {
	case "MIDDLE":
		if horizontal {
			mesh.SetPositionX(-w / 2)
		} else {
			mesh.SetPositionZ(h / 2)
		}
	case "END":
		if horizontal {
			mesh.SetPositionX(-w)
		} else {
			mesh.SetPositionZ(h)
		}
	}
}

func convertTextFallback(t *node.Text, app *node.Appearance, parent *core.Node, baseDir string, size float64) {
	joined := strings.Join(t.String, "\n")
	lines := strings.Split(joined, "\n")
	charW := float32(size * 0.6)
	charH := float32(size)

	for i, line := range lines {
		w := charW * float32(len(line))
		h := charH
		y := -float32(i) * charH * 1.2

		positions := math32.NewArrayF32(0, 12)
		positions.Append(0, y, 0)
		positions.Append(w, y, 0)
		positions.Append(w, y+h, 0)
		positions.Append(0, y+h, 0)

		indices := math32.NewArrayU32(0, 6)
		indices.Append(0, 1, 2, 0, 2, 3)

		geom := geometry.NewGeometry()
		geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
		geom.SetIndices(indices)

		c := math32.NewColor("white")
		if app != nil && app.Material != nil {
			c = toColor(&app.Material.DiffuseColor)
		}
		mat := material.NewStandard(c)
		mat.SetSide(material.SideFront)
		mat.SetOpacity(0.5)

		mesh := graphic.NewMesh(geom, mat)
		mesh.SetName(t.GetName())
		parent.Add(mesh)
	}
}

func resolveFontPath(family, style string) string {
	base := "/System/Library/Fonts/"
	switch strings.ToUpper(family) {
	case "SANS", "SANS-SERIF":
		switch strings.ToUpper(style) {
		case "BOLD":
			return base + "Helvetica.ttc"
		case "ITALIC":
			return base + "Helvetica.ttc"
		default:
			return base + "Helvetica.ttc"
		}
	case "TYPEWRITER", "MONOSPACE":
		return base + "Courier.dfont"
	default:
		switch strings.ToUpper(style) {
		case "BOLD":
			return base + "Times.ttc"
		case "ITALIC":
			return base + "Times.ttc"
		default:
			return base + "Times.ttc"
		}
	}
}
