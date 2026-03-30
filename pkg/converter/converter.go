package converter

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/texture"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// NodeMap tracks correspondences between VRML nodes and their g3n counterparts.
// Used by the event engine to update g3n nodes when VRML fields change.
type NodeMap struct {
	Transforms map[*node.Transform]*core.Node
	Materials  map[*node.Material]*material.Standard
	LODs       map[*node.LOD][]*core.Node
	Switches   map[*node.Switch][]*core.Node
	Billboards map[*node.Billboard]*core.Node
	Anchors    map[*node.Anchor]*core.Node

	// CameraPos is the current camera world position, set before UpdateDynamic.
	CameraPos math32.Vector3
}

// NewNodeMap creates an empty node mapping.
func NewNodeMap() *NodeMap {
	return &NodeMap{
		Transforms: make(map[*node.Transform]*core.Node),
		Materials:  make(map[*node.Material]*material.Standard),
		LODs:       make(map[*node.LOD][]*core.Node),
		Switches:   make(map[*node.Switch][]*core.Node),
		Billboards: make(map[*node.Billboard]*core.Node),
		Anchors:    make(map[*node.Anchor]*core.Node),
	}
}

// Convert walks the VRML scene graph and adds corresponding g3n objects
// to the given g3n parent node. Returns a NodeMap for animation updates.
func Convert(vrmlNodes []node.Node, parent *core.Node, baseDir string) *NodeMap {
	nm := NewNodeMap()
	for _, n := range vrmlNodes {
		convertNode(n, parent, baseDir, nm)
	}
	return nm
}

func convertNode(n node.Node, parent *core.Node, baseDir string, nm *NodeMap) {
	switch v := n.(type) {
	case *node.Transform:
		convertTransform(v, parent, baseDir, nm)
	case *node.Group:
		convertGroup(&v.GroupingNode, parent, baseDir, nm)
	case *node.Shape:
		convertShape(v, parent, baseDir, nm)
	case *node.DirectionalLight:
		convertDirLight(v, parent)
	case *node.PointLight:
		convertPointLight(v, parent)
	case *node.SpotLight:
		convertSpotLight(v, parent)
	case *node.Anchor:
		convertAnchor(v, parent, baseDir, nm)
	case *node.Billboard:
		convertBillboard(v, parent, baseDir, nm)
	case *node.Collision:
		convertGroup(&v.GroupingNode, parent, baseDir, nm)
	case *node.Switch:
		convertSwitch(v, parent, baseDir, nm)
	case *node.LOD:
		convertLOD(v, parent, baseDir, nm)
	case *node.Inline:
		convertGroup(&v.GroupingNode, parent, baseDir, nm)
	case *node.Viewpoint:
		// handled by viewer main
	case *node.Background:
		convertBackground(v, parent, baseDir)
	case *node.Fog:
		// fog params extracted via GetFogParams()
	case *node.Sound:
		convertSound(v, parent, baseDir)
	default:
		// skip unsupported nodes
	}
}

// UpdateDynamic syncs changed VRML transform/material fields to their g3n counterparts.
func (nm *NodeMap) UpdateDynamic() {
	for vrmlT, g3nNode := range nm.Transforms {
		g3nNode.SetPosition(float32(vrmlT.Translation.X), float32(vrmlT.Translation.Y), float32(vrmlT.Translation.Z))
		if vrmlT.Rotation.W != 0 {
			axis := math32.Vector3{X: float32(vrmlT.Rotation.X), Y: float32(vrmlT.Rotation.Y), Z: float32(vrmlT.Rotation.Z)}
			axis.Normalize()
			q := math32.NewQuaternion(0, 0, 0, 1)
			q.SetFromAxisAngle(&axis, float32(vrmlT.Rotation.W))
			g3nNode.SetQuaternionQuat(q)
		}
		g3nNode.SetScale(float32(vrmlT.Scale.X), float32(vrmlT.Scale.Y), float32(vrmlT.Scale.Z))
	}
	for vrmlM, g3nMat := range nm.Materials {
		g3nMat.SetColor(toColor(&vrmlM.DiffuseColor))
		g3nMat.SetEmissiveColor(toColor(&vrmlM.EmissiveColor))
		if vrmlM.Transparency > 0 {
			g3nMat.SetOpacity(float32(1.0 - vrmlM.Transparency))
		}
	}
	for vrmlLOD, levels := range nm.LODs {
		active := vrmlLOD.ActiveLevel
		if active < 0 {
			active = 0
		}
		for i, wrapper := range levels {
			wrapper.SetVisible(int64(i) == active)
		}
	}
	for vrmlSW, choices := range nm.Switches {
		for i, wrapper := range choices {
			wrapper.SetVisible(int64(i) == vrmlSW.WhichChoice)
		}
	}
	for vrmlBB, g3nNode := range nm.Billboards {
		updateBillboardRotation(vrmlBB, g3nNode, nm.CameraPos)
	}
}

func convertTransform(t *node.Transform, parent *core.Node, baseDir string, nm *NodeMap) {
	gn := core.NewNode()
	gn.SetName(t.GetName())

	hasCenter := t.Center.X != 0 || t.Center.Y != 0 || t.Center.Z != 0

	if hasCenter {
		// VRML transform: T * C * R * S * C^-1
		gn.SetPosition(
			float32(t.Translation.X+t.Center.X),
			float32(t.Translation.Y+t.Center.Y),
			float32(t.Translation.Z+t.Center.Z),
		)
	} else {
		gn.SetPosition(float32(t.Translation.X), float32(t.Translation.Y), float32(t.Translation.Z))
	}

	if t.Rotation.W != 0 {
		axis := math32.Vector3{X: float32(t.Rotation.X), Y: float32(t.Rotation.Y), Z: float32(t.Rotation.Z)}
		axis.Normalize()
		q := math32.NewQuaternion(0, 0, 0, 1)
		q.SetFromAxisAngle(&axis, float32(t.Rotation.W))
		gn.SetQuaternionQuat(q)
	}

	gn.SetScale(float32(t.Scale.X), float32(t.Scale.Y), float32(t.Scale.Z))
	parent.Add(gn)

	// Register for dynamic updates
	nm.Transforms[t] = gn

	childParent := gn
	if hasCenter {
		inner := core.NewNode()
		inner.SetPosition(float32(-t.Center.X), float32(-t.Center.Y), float32(-t.Center.Z))
		gn.Add(inner)
		childParent = inner
	}

	// Tag g3n node with VRML children if this group contains pointing-device sensors.
	// This lets the Picker walk up from a hit mesh to find sibling sensors.
	if hasSensorChild(t.Children) {
		childParent.SetUserData(asNodeSlice(t.Children))
	}

	for _, child := range t.Children {
		convertNode(child, childParent, baseDir, nm)
	}
}

func convertGroup(g *node.GroupingNode, parent *core.Node, baseDir string, nm *NodeMap) {
	gn := core.NewNode()
	gn.SetName(g.GetName())
	parent.Add(gn)

	if hasSensorChild(g.Children) {
		gn.SetUserData(asNodeSlice(g.Children))
	}

	for _, child := range g.Children {
		convertNode(child, gn, baseDir, nm)
	}
}

func convertBillboard(bb *node.Billboard, parent *core.Node, baseDir string, nm *NodeMap) {
	gn := core.NewNode()
	gn.SetName(bb.GetName())
	parent.Add(gn)
	nm.Billboards[bb] = gn

	if hasSensorChild(bb.Children) {
		gn.SetUserData(asNodeSlice(bb.Children))
	}

	for _, child := range bb.Children {
		convertNode(child, gn, baseDir, nm)
	}
}

func convertAnchor(a *node.Anchor, parent *core.Node, baseDir string, nm *NodeMap) {
	gn := core.NewNode()
	gn.SetName(a.GetName())
	parent.Add(gn)
	nm.Anchors[a] = gn

	// Tag with the Anchor node so the Picker can detect it when walking parents
	gn.SetUserData(a)

	for _, child := range a.Children {
		convertNode(child, gn, baseDir, nm)
	}
}

func convertShape(s *node.Shape, parent *core.Node, baseDir string, nm *NodeMap) {
	if s.Geometry == nil {
		return
	}

	// Lines, points, and text use separate graphic types
	switch gn := s.Geometry.(type) {
	case *node.IndexedLineSet:
		convertIndexedLineSet(gn, s.Appearance, parent)
		return
	case *node.PointSet:
		convertPointSet(gn, s.Appearance, parent)
		return
	case *node.Text:
		convertText(gn, s.Appearance, parent, baseDir)
		return
	}

	mat := buildMaterial(s.Appearance, baseDir)
	geom := buildGeometry(s.Geometry)
	if geom == nil {
		return
	}
	mesh := graphic.NewMesh(geom, mat)
	mesh.SetName(s.GetName())
	parent.Add(mesh)

	// Register material for dynamic updates (color changes via routes)
	if s.Appearance != nil && s.Appearance.Material != nil {
		if stdMat, ok := mat.(*material.Standard); ok {
			nm.Materials[s.Appearance.Material] = stdMat
		}
	}
}

func buildMaterial(app *node.Appearance, baseDir string) material.IMaterial {
	if app == nil {
		return material.NewStandard(&math32.Color{R: 0.8, G: 0.8, B: 0.8})
	}

	// When there's a texture but no Material node, use white so the texture
	// colours come through unmodified.
	var mat *material.Standard
	if app.Material != nil {
		m := app.Material
		dc := toColor(&m.DiffuseColor)
		mat = material.NewStandard(dc)
		mat.SetEmissiveColor(toColor(&m.EmissiveColor))
		mat.SetSpecularColor(toColor(&m.SpecularColor))
		mat.SetShininess(float32(m.Shininess * 128.0))
		if m.Transparency > 0 {
			mat.SetOpacity(float32(1.0 - m.Transparency))
			mat.SetTransparent(true)
		}
	} else {
		mat = material.NewStandard(&math32.Color{R: 1, G: 1, B: 1})
	}

	if app.Texture != nil {
		var tex *texture.Texture2D
		switch t := app.Texture.(type) {
		case *node.ImageTexture:
			if len(t.URL) > 0 {
				tex = loadTexture(t.URL[0], baseDir)
				if tex != nil {
					setWrap(tex, t.RepeatS, t.RepeatT)
				}
			}
		case *node.PixelTexture:
			tex = pixelTextureToG3n(&t.Image)
			if tex != nil {
				setWrap(tex, t.RepeatS, t.RepeatT)
			}
		case *node.MovieTexture:
			// Treat MovieTexture as a static image (first frame from URL).
			if len(t.URL) > 0 {
				tex = loadTexture(t.URL[0], baseDir)
				if tex != nil {
					setWrap(tex, t.RepeatS, t.RepeatT)
				}
			}
		}
		if tex != nil {
			if app.TextureTransform != nil {
				applyTextureTransform(tex, app.TextureTransform)
			}
			mat.AddTexture(tex)
		}
	}

	return mat
}

func loadTexture(url string, baseDir string) *texture.Texture2D {
	path := url
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, url)
	}

	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot load texture %s: %v\n", url, err)
		return nil
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: close error for texture %s: %v\n", url, err)
		}
	}()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot decode texture %s: %v\n", url, err)
		return nil
	}

	return texture.NewTexture2DFromRGBA(toRGBA(img))
}

func toRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba
}

func setWrap(tex *texture.Texture2D, repeatS, repeatT bool) {
	if repeatS {
		tex.SetWrapS(gls.REPEAT)
	} else {
		tex.SetWrapS(gls.CLAMP_TO_EDGE)
	}
	if repeatT {
		tex.SetWrapT(gls.REPEAT)
	} else {
		tex.SetWrapT(gls.CLAMP_TO_EDGE)
	}
}

func pixelTextureToG3n(si *vec.SFImage) *texture.Texture2D {
	w, h, nc := int(si.Width), int(si.Height), int(si.NumComponents)
	if w == 0 || h == 0 || nc == 0 {
		return nil
	}
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// VRML PixelTexture stores bottom-to-top; RGBA is top-to-bottom.
			srcY := h - 1 - y
			off := (srcY*w + x) * nc
			var r, g, b, a uint8
			switch nc {
			case 1:
				r, g, b, a = si.Pixels[off], si.Pixels[off], si.Pixels[off], 255
			case 2:
				r, g, b, a = si.Pixels[off], si.Pixels[off], si.Pixels[off], si.Pixels[off+1]
			case 3:
				r, g, b, a = si.Pixels[off], si.Pixels[off+1], si.Pixels[off+2], 255
			case 4:
				r, g, b, a = si.Pixels[off], si.Pixels[off+1], si.Pixels[off+2], si.Pixels[off+3]
			}
			rgba.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
	return texture.NewTexture2DFromRGBA(rgba)
}

func applyTextureTransform(tex *texture.Texture2D, tt *node.TextureTransform) {
	// VRML97 TextureTransform: Tc' = -C * S * R * C * T * Tc
	// g3n supports offset and repeat (scale) but not rotation.
	// We apply translation as offset and scale as repeat.
	tex.SetOffset(float32(tt.Translation.X-tt.Center.X*(tt.Scale.X-1)), float32(tt.Translation.Y-tt.Center.Y*(tt.Scale.Y-1)))
	tex.SetRepeat(float32(tt.Scale.X), float32(tt.Scale.Y))
}

func buildGeometry(gn node.GeometryNode) *geometry.Geometry {
	switch v := gn.(type) {
	case *node.Box:
		return geometry.NewBox(float32(v.Size.X), float32(v.Size.Y), float32(v.Size.Z))
	case *node.Sphere:
		return geometry.NewSphere(float64(v.Radius), int(v.Slices), int(v.Stacks))
	case *node.Cone:
		return geometry.NewCone(float64(v.BottomRadius), float64(v.Height), 32, 1, v.Bottom)
	case *node.Cylinder:
		return geometry.NewCylinder(float64(v.Radius), float64(v.Height), 32, 1, v.Top, v.Bottom)
	case *node.IndexedFaceSet:
		return buildIndexedFaceSet(v)
	case *node.ElevationGrid:
		return buildElevationGrid(v)
	case *node.Extrusion:
		return buildExtrusion(v)
	default:
		return nil
	}
}

func buildIndexedFaceSet(ifs *node.IndexedFaceSet) *geometry.Geometry {
	if ifs.Coord == nil || len(ifs.Coord.Point) == 0 {
		return nil
	}

	geom := geometry.NewGeometry()

	points := ifs.Coord.Point
	indices := ifs.CoordIndex

	// Build positions VBO
	positions := math32.NewArrayF32(0, len(points)*3)
	for _, p := range points {
		positions.Append(float32(p.X), float32(p.Y), float32(p.Z))
	}
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))

	// Normals
	if ifs.Normal != nil && len(ifs.Normal.Vector) > 0 {
		normals := math32.NewArrayF32(0, len(ifs.Normal.Vector)*3)
		for _, n := range ifs.Normal.Vector {
			normals.Append(float32(n.X), float32(n.Y), float32(n.Z))
		}
		geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	}

	// Colors
	if ifs.Color != nil && len(ifs.Color.Color) > 0 {
		colors := math32.NewArrayF32(0, len(ifs.Color.Color)*3)
		for _, c := range ifs.Color.Color {
			colors.Append(float32(c.R), float32(c.G), float32(c.B))
		}
		geom.AddVBO(gls.NewVBO(colors).AddAttrib(gls.VertexColor))
	}

	// Texture coords
	if ifs.TexCoord != nil && len(ifs.TexCoord.Point) > 0 {
		uvs := math32.NewArrayF32(0, len(ifs.TexCoord.Point)*2)
		for _, tc := range ifs.TexCoord.Point {
			uvs.Append(float32(tc.X), float32(tc.Y))
		}
		geom.AddVBO(gls.NewVBO(uvs).AddAttrib(gls.VertexTexcoord))
	}

	// Build triangle indices from VRML face indices (split by -1)
	triIndices := math32.NewArrayU32(0, 0)
	var face []uint32
	for _, idx := range indices {
		if idx == -1 {
			for i := 1; i+1 < len(face); i++ {
				if ifs.Ccw {
					triIndices.Append(face[0], face[i], face[i+1])
				} else {
					triIndices.Append(face[0], face[i+1], face[i])
				}
			}
			face = face[:0]
		} else {
			face = append(face, uint32(idx))
		}
	}
	if len(face) >= 3 {
		for i := 1; i+1 < len(face); i++ {
			if ifs.Ccw {
				triIndices.Append(face[0], face[i], face[i+1])
			} else {
				triIndices.Append(face[0], face[i+1], face[i])
			}
		}
	}
	geom.SetIndices(triIndices)

	// Auto-generate normals if not provided
	if ifs.Normal == nil || len(ifs.Normal.Vector) == 0 {
		computeNormals(geom, positions, triIndices)
	}

	return geom
}

func computeNormals(geom *geometry.Geometry, positions math32.ArrayF32, indices math32.ArrayU32) {
	nVerts := positions.Len() / 3
	normals := math32.NewArrayF32(nVerts*3, nVerts*3)

	for i := 0; i+2 < indices.Len(); i += 3 {
		i0 := indices[i]
		i1 := indices[i+1]
		i2 := indices[i+2]

		v0 := math32.Vector3{X: positions[i0*3], Y: positions[i0*3+1], Z: positions[i0*3+2]}
		v1 := math32.Vector3{X: positions[i1*3], Y: positions[i1*3+1], Z: positions[i1*3+2]}
		v2 := math32.Vector3{X: positions[i2*3], Y: positions[i2*3+1], Z: positions[i2*3+2]}

		e1 := v1
		e1.Sub(&v0)
		e2 := v2
		e2.Sub(&v0)
		var n math32.Vector3
		n.CrossVectors(&e1, &e2)

		for _, idx := range []uint32{i0, i1, i2} {
			normals[idx*3] += n.X
			normals[idx*3+1] += n.Y
			normals[idx*3+2] += n.Z
		}
	}

	for i := 0; i < nVerts; i++ {
		nx := normals[i*3]
		ny := normals[i*3+1]
		nz := normals[i*3+2]
		l := float32(math.Sqrt(float64(nx*nx + ny*ny + nz*nz)))
		if l > 0 {
			normals[i*3] /= l
			normals[i*3+1] /= l
			normals[i*3+2] /= l
		}
	}

	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
}

func buildElevationGrid(eg *node.ElevationGrid) *geometry.Geometry {
	if eg.XDimension < 2 || eg.ZDimension < 2 || len(eg.Heights) == 0 {
		return nil
	}

	geom := geometry.NewGeometry()
	xDim := int(eg.XDimension)
	zDim := int(eg.ZDimension)

	positions := math32.NewArrayF32(0, xDim*zDim*3)
	for z := 0; z < zDim; z++ {
		for x := 0; x < xDim; x++ {
			idx := z*xDim + x
			h := float64(0)
			if idx < len(eg.Heights) {
				h = eg.Heights[idx]
			}
			positions.Append(float32(float64(x)*eg.XSpacing), float32(h), float32(float64(z)*eg.ZSpacing))
		}
	}
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))

	triIndices := math32.NewArrayU32(0, 0)
	for z := 0; z < zDim-1; z++ {
		for x := 0; x < xDim-1; x++ {
			i0 := uint32(z*xDim + x)
			i1 := i0 + 1
			i2 := uint32((z+1)*xDim + x)
			i3 := i2 + 1
			triIndices.Append(i0, i2, i1)
			triIndices.Append(i1, i2, i3)
		}
	}
	geom.SetIndices(triIndices)
	computeNormals(geom, positions, triIndices)

	return geom
}

func buildExtrusion(ex *node.Extrusion) *geometry.Geometry {
	if len(ex.CrossSection) < 3 || len(ex.Spine) < 2 {
		return nil
	}

	nSpine := len(ex.Spine)
	nCross := len(ex.CrossSection)

	geom := geometry.NewGeometry()
	positions := math32.NewArrayF32(0, nSpine*nCross*3)

	for i, sp := range ex.Spine {
		scl := vec.SFVec2f{X: 1, Y: 1}
		if i < len(ex.Scale) {
			scl = ex.Scale[i]
		}
		for _, cs := range ex.CrossSection {
			positions.Append(
				float32(sp.X+cs.X*scl.X),
				float32(sp.Y),
				float32(sp.Z+cs.Y*scl.Y),
			)
		}
	}
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))

	triIndices := math32.NewArrayU32(0, 0)
	for i := 0; i < nSpine-1; i++ {
		for j := 0; j < nCross-1; j++ {
			i0 := uint32(i*nCross + j)
			i1 := i0 + 1
			i2 := uint32((i+1)*nCross + j)
			i3 := i2 + 1
			triIndices.Append(i0, i2, i1)
			triIndices.Append(i1, i2, i3)
		}
	}
	geom.SetIndices(triIndices)
	computeNormals(geom, positions, triIndices)

	return geom
}

func convertDirLight(dl *node.DirectionalLight, parent *core.Node) {
	if !dl.On {
		return
	}
	l := light.NewDirectional(toColor(&dl.Color), float32(dl.Intensity))
	l.SetPosition(float32(dl.Direction.X), float32(dl.Direction.Y), float32(dl.Direction.Z))
	l.SetName(dl.GetName())
	parent.Add(l)
}

func convertPointLight(pl *node.PointLight, parent *core.Node) {
	if !pl.On {
		return
	}
	l := light.NewPoint(toColor(&pl.Color), float32(pl.Intensity))
	l.SetPosition(float32(pl.Location.X), float32(pl.Location.Y), float32(pl.Location.Z))
	l.SetLinearDecay(float32(pl.Attenuation.Y))
	l.SetQuadraticDecay(float32(pl.Attenuation.Z))
	l.SetName(pl.GetName())
	parent.Add(l)
}

func convertSpotLight(sl *node.SpotLight, parent *core.Node) {
	if !sl.On {
		return
	}
	l := light.NewSpot(toColor(&sl.Color), float32(sl.Intensity))
	l.SetPosition(float32(sl.Location.X), float32(sl.Location.Y), float32(sl.Location.Z))
	l.SetDirection(float32(sl.Direction.X), float32(sl.Direction.Y), float32(sl.Direction.Z))
	l.SetCutoffAngle(float32(sl.CutOffAngle * 180.0 / math.Pi))
	l.SetLinearDecay(float32(sl.Attenuation.Y))
	l.SetQuadraticDecay(float32(sl.Attenuation.Z))
	l.SetName(sl.GetName())
	parent.Add(l)
}

func convertSwitch(sw *node.Switch, parent *core.Node, baseDir string, nm *NodeMap) {
	if len(sw.Choice) == 0 {
		return
	}
	container := core.NewNode()
	container.SetName(sw.GetName())
	parent.Add(container)

	levels := make([]*core.Node, len(sw.Choice))
	for i, child := range sw.Choice {
		wrapper := core.NewNode()
		wrapper.SetVisible(i == int(sw.WhichChoice))
		container.Add(wrapper)
		convertNode(child, wrapper, baseDir, nm)
		levels[i] = wrapper
	}
	nm.Switches[sw] = levels
}

func convertLOD(lod *node.LOD, parent *core.Node, baseDir string, nm *NodeMap) {
	if len(lod.Level) == 0 {
		return
	}
	container := core.NewNode()
	container.SetName(lod.GetName())
	parent.Add(container)

	active := lod.ActiveLevel
	if active < 0 {
		active = 0
	}
	levels := make([]*core.Node, len(lod.Level))
	for i, child := range lod.Level {
		wrapper := core.NewNode()
		wrapper.SetVisible(int64(i) == active)
		container.Add(wrapper)
		convertNode(child, wrapper, baseDir, nm)
		levels[i] = wrapper
	}
	nm.LODs[lod] = levels
}

// updateBillboardRotation rotates a billboard's g3n node to face the camera.
// Uses the VRML97 axisOfRotation constraint: if non-zero, rotate only around
// that axis; if zero (0,0,0), rotate freely to face the viewer.
func updateBillboardRotation(bb *node.Billboard, g3nNode *core.Node, camPos math32.Vector3) {
	matWorld := g3nNode.MatrixWorld()
	var bbPos math32.Vector3
	bbPos.SetFromMatrixPosition(&matWorld)

	dir := math32.Vector3{
		X: camPos.X - bbPos.X,
		Y: camPos.Y - bbPos.Y,
		Z: camPos.Z - bbPos.Z,
	}

	axis := bb.AxisOfRotation
	if axis.X == 0 && axis.Y == 0 && axis.Z == 0 {
		// Free rotation: face the camera fully
		dir.Normalize()
		q := math32.NewQuaternion(0, 0, 0, 1)
		// LookAt: default forward is -Z in g3n
		var up math32.Vector3
		up.Set(0, 1, 0)
		var rotMat math32.Matrix4
		rotMat.LookAt(&bbPos, &camPos, &up)
		q.SetFromRotationMatrix(&rotMat)
		g3nNode.SetQuaternionQuat(q)
	} else {
		// Constrained rotation around axisOfRotation (typically Y axis)
		angle := float64(math.Atan2(float64(dir.X), float64(dir.Z)))
		g3nAxis := math32.Vector3{X: float32(axis.X), Y: float32(axis.Y), Z: float32(axis.Z)}
		g3nAxis.Normalize()
		q := math32.NewQuaternion(0, 0, 0, 1)
		q.SetFromAxisAngle(&g3nAxis, float32(angle))
		g3nNode.SetQuaternionQuat(q)
	}
}

func toColor(c *vec.SFColor) *math32.Color {
	return &math32.Color{R: float32(c.R), G: float32(c.G), B: float32(c.B)}
}

// GetViewpoint searches for the first Viewpoint node.
func GetViewpoint(nodes []node.Node) *node.Viewpoint {
	for _, n := range nodes {
		if vp, ok := n.(*node.Viewpoint); ok {
			return vp
		}
		switch v := n.(type) {
		case *node.Transform:
			if vp := GetViewpoint(v.Children); vp != nil {
				return vp
			}
		case *node.Group:
			if vp := GetViewpoint(v.Children); vp != nil {
				return vp
			}
		}
	}
	return nil
}

// GetBackground searches for the first Background node.
func GetBackground(nodes []node.Node) *node.Background {
	for _, n := range nodes {
		if bg, ok := n.(*node.Background); ok {
			return bg
		}
		switch v := n.(type) {
		case *node.Transform:
			if bg := GetBackground(v.Children); bg != nil {
				return bg
			}
		case *node.Group:
			if bg := GetBackground(v.Children); bg != nil {
				return bg
			}
		}
	}
	return nil
}

// GetNavigationInfo searches for the first NavigationInfo node.
func GetNavigationInfo(nodes []node.Node) *node.NavigationInfo {
	for _, n := range nodes {
		if ni, ok := n.(*node.NavigationInfo); ok {
			return ni
		}
		switch v := n.(type) {
		case *node.Transform:
			if ni := GetNavigationInfo(v.Children); ni != nil {
				return ni
			}
		case *node.Group:
			if ni := GetNavigationInfo(v.Children); ni != nil {
				return ni
			}
		}
	}
	return nil
}

// GetFog searches for the first Fog node.
func GetFog(nodes []node.Node) *node.Fog {
	for _, n := range nodes {
		if fg, ok := n.(*node.Fog); ok {
			return fg
		}
		switch v := n.(type) {
		case *node.Transform:
			if fg := GetFog(v.Children); fg != nil {
				return fg
			}
		case *node.Group:
			if fg := GetFog(v.Children); fg != nil {
				return fg
			}
		}
	}
	return nil
}

func convertIndexedLineSet(ils *node.IndexedLineSet, app *node.Appearance, parent *core.Node) {
	if ils.Coord == nil || len(ils.Coord.Point) == 0 {
		return
	}

	geom := geometry.NewGeometry()
	nPts := len(ils.Coord.Point)

	positions := math32.NewArrayF32(0, nPts*3)
	for _, p := range ils.Coord.Point {
		positions.Append(float32(p.X), float32(p.Y), float32(p.Z))
	}
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))

	// Basic shader requires per-vertex colors
	colors := math32.NewArrayF32(0, nPts*3)
	if ils.Color != nil && len(ils.Color.Color) > 0 {
		nColors := len(ils.Color.Color)
		for i := 0; i < nPts; i++ {
			ci := i
			if ci >= nColors {
				ci = nColors - 1
			}
			c := ils.Color.Color[ci]
			colors.Append(float32(c.R), float32(c.G), float32(c.B))
		}
	} else {
		var r, g, b float64 = 1, 1, 1
		if app != nil && app.Material != nil {
			ec := app.Material.EmissiveColor
			r, g, b = ec.R, ec.G, ec.B
		}
		for i := 0; i < nPts; i++ {
			colors.Append(float32(r), float32(g), float32(b))
		}
	}
	geom.AddVBO(gls.NewVBO(colors).AddAttrib(gls.VertexColor))

	// Build line pair indices from polylines separated by -1
	lineIndices := math32.NewArrayU32(0, 0)
	nCoords := uint32(nPts)
	var strip []uint32
	for _, idx := range ils.CoordIndex {
		if idx == -1 {
			for i := 0; i+1 < len(strip); i++ {
				lineIndices.Append(strip[i], strip[i+1])
			}
			strip = strip[:0]
		} else if uint32(idx) < nCoords {
			strip = append(strip, uint32(idx))
		}
	}
	for i := 0; i+1 < len(strip); i++ {
		lineIndices.Append(strip[i], strip[i+1])
	}
	geom.SetIndices(lineIndices)

	mat := material.NewBasic()
	lines := graphic.NewLines(geom, mat)
	parent.Add(lines)
}

func convertPointSet(ps *node.PointSet, app *node.Appearance, parent *core.Node) {
	if ps.Coord == nil || len(ps.Coord.Point) == 0 {
		return
	}

	geom := geometry.NewGeometry()

	positions := math32.NewArrayF32(0, len(ps.Coord.Point)*3)
	for _, p := range ps.Coord.Point {
		positions.Append(float32(p.X), float32(p.Y), float32(p.Z))
	}
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))

	ptColor := &math32.Color{R: 1, G: 1, B: 1}
	if ps.Color != nil && len(ps.Color.Color) > 0 {
		ptColor = toColor(&ps.Color.Color[0])
	} else if app != nil && app.Material != nil {
		ptColor = toColor(&app.Material.EmissiveColor)
	}
	mat := material.NewPoint(ptColor)
	mat.SetSize(2.0)

	pts := graphic.NewPoints(geom, mat)
	parent.Add(pts)
}

// hasSensorChild returns true if children contains a pointing-device sensor.
func hasSensorChild(children []node.Node) bool {
	for _, c := range children {
		switch c.(type) {
		case *node.TouchSensor, *node.PlaneSensor, *node.SphereSensor, *node.CylinderSensor:
			return true
		}
	}
	return false
}

// asNodeSlice converts a []node.Node for storage as UserData.
func asNodeSlice(children []node.Node) []node.Node {
	out := make([]node.Node, len(children))
	copy(out, children)
	return out
}
