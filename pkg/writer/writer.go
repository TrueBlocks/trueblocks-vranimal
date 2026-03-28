// Package writer serializes a VRML97 scene graph back to valid .wrl text.
package writer

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// Writer serializes VRML97 scene graphs to an io.Writer.
type Writer struct {
	w       io.Writer
	indent  int
	tab     string
	defined map[string]bool
}

// New creates a Writer that outputs to w.
func New(w io.Writer) *Writer {
	return &Writer{
		w:       w,
		tab:     "  ",
		defined: make(map[string]bool),
	}
}

// WriteScene writes the VRML header and all top-level nodes.
func (wr *Writer) WriteScene(nodes []node.Node) {
	wr.printf("#VRML V2.0 utf8\n\n")
	for _, n := range nodes {
		wr.writeNode(n)
		wr.printf("\n")
	}
}

func (wr *Writer) printf(format string, args ...any) {
	fmt.Fprintf(wr.w, format, args...)
}

func (wr *Writer) indentStr() string {
	return strings.Repeat(wr.tab, wr.indent)
}

func (wr *Writer) line(format string, args ...any) {
	wr.printf("%s", wr.indentStr())
	wr.printf(format, args...)
	wr.printf("\n")
}

func isNilNode(n node.Node) bool {
	if n == nil {
		return true
	}
	v := reflect.ValueOf(n)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

func (wr *Writer) writeNode(n node.Node) {
	if isNilNode(n) {
		return
	}
	name := n.GetName()
	if name != "" {
		if wr.defined[name] {
			wr.line("USE %s", name)
			return
		}
		wr.defined[name] = true
		wr.printf("%sDEF %s ", wr.indentStr(), name)
	} else {
		wr.printf("%s", wr.indentStr())
	}
	wr.writeNodeBody(n)
}

func (wr *Writer) writeNodeBody(n node.Node) {
	switch v := n.(type) {
	case *node.Appearance:
		wr.writeAppearance(v)
	case *node.Material:
		wr.writeMaterial(v)
	case *node.ImageTexture:
		wr.writeImageTexture(v)
	case *node.MovieTexture:
		wr.writeMovieTexture(v)
	case *node.PixelTexture:
		wr.writePixelTexture(v)
	case *node.TextureTransform:
		wr.writeTextureTransform(v)
	case *node.FontStyle:
		wr.writeFontStyle(v)
	case *node.Background:
		wr.writeBackground(v)
	case *node.Fog:
		wr.writeFog(v)
	case *node.NavigationInfo:
		wr.writeNavigationInfo(v)
	case *node.Viewpoint:
		wr.writeViewpoint(v)
	case *node.Shape:
		wr.writeShape(v)
	case *node.DirectionalLight:
		wr.writeDirectionalLight(v)
	case *node.PointLight:
		wr.writePointLight(v)
	case *node.SpotLight:
		wr.writeSpotLight(v)
	case *node.WorldInfo:
		wr.writeWorldInfo(v)
	case *node.Script:
		wr.writeScript(v)
	case *node.Sound:
		wr.writeSound(v)
	case *node.AudioClip:
		wr.writeAudioClip(v)
	case *node.Transform:
		wr.writeTransform(v)
	case *node.Anchor:
		wr.writeAnchor(v)
	case *node.Billboard:
		wr.writeBillboard(v)
	case *node.Collision:
		wr.writeCollision(v)
	case *node.Inline:
		wr.writeInline(v)
	case *node.Group:
		wr.writeGroup(v)
	case *node.LOD:
		wr.writeLOD(v)
	case *node.Switch:
		wr.writeSwitch(v)
	case *node.Box:
		wr.writeBox(v)
	case *node.Cone:
		wr.writeCone(v)
	case *node.Cylinder:
		wr.writeCylinder(v)
	case *node.Sphere:
		wr.writeSphere(v)
	case *node.Extrusion:
		wr.writeExtrusion(v)
	case *node.Text:
		wr.writeText(v)
	case *node.IndexedFaceSet:
		wr.writeIndexedFaceSet(v)
	case *node.IndexedLineSet:
		wr.writeIndexedLineSet(v)
	case *node.PointSet:
		wr.writePointSet(v)
	case *node.ElevationGrid:
		wr.writeElevationGrid(v)
	case *node.ColorNode:
		wr.writeColorNode(v)
	case *node.Coordinate:
		wr.writeCoordinate(v)
	case *node.NormalNode:
		wr.writeNormalNode(v)
	case *node.TextureCoordinate:
		wr.writeTextureCoordinate(v)
	case *node.ColorInterpolator:
		wr.writeColorInterpolator(v)
	case *node.CoordinateInterpolator:
		wr.writeCoordinateInterpolator(v)
	case *node.NormalInterpolator:
		wr.writeNormalInterpolator(v)
	case *node.OrientationInterpolator:
		wr.writeOrientationInterpolator(v)
	case *node.PositionInterpolator:
		wr.writePositionInterpolator(v)
	case *node.ScalarInterpolator:
		wr.writeScalarInterpolator(v)
	case *node.TouchSensor:
		wr.writeTouchSensor(v)
	case *node.TimeSensor:
		wr.writeTimeSensor(v)
	case *node.ProximitySensor:
		wr.writeProximitySensor(v)
	case *node.CylinderSensor:
		wr.writeCylinderSensor(v)
	case *node.PlaneSensor:
		wr.writePlaneSensor(v)
	case *node.SphereSensor:
		wr.writeSphereSensor(v)
	case *node.VisibilitySensor:
		wr.writeVisibilitySensor(v)
	default:
		wr.printf("# unknown node type %T\n", n)
	}
}

func fmtColor(c vec.SFColor) string {
	return fmt.Sprintf("%g %g %g", c.R, c.G, c.B)
}

func fmtVec3(v vec.SFVec3f) string {
	return fmt.Sprintf("%g %g %g", v.X, v.Y, v.Z)
}

func fmtVec2(v vec.SFVec2f) string {
	return fmt.Sprintf("%g %g", v.X, v.Y)
}

func fmtRot(r vec.SFRotation) string {
	return fmt.Sprintf("%g %g %g %g", r.X, r.Y, r.Z, r.W)
}

func fmtBool(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

func (wr *Writer) writeChildren(children []node.Node) {
	if len(children) == 0 {
		return
	}
	wr.line("children [")
	wr.indent++
	for _, child := range children {
		wr.writeNode(child)
	}
	wr.indent--
	wr.line("]")
}

func (wr *Writer) writeSFNode(fieldName string, n node.Node) {
	if isNilNode(n) {
		return
	}
	wr.printf("%s%s ", wr.indentStr(), fieldName)
	wr.writeNodeInline(n)
}

// writeNodeInline writes a node without leading indent (used after field names).
func (wr *Writer) writeNodeInline(n node.Node) {
	if isNilNode(n) {
		return
	}
	name := n.GetName()
	if name != "" {
		if wr.defined[name] {
			wr.printf("USE %s\n", name)
			return
		}
		wr.defined[name] = true
		wr.printf("DEF %s ", name)
	}
	wr.writeNodeBody(n)
}

func (wr *Writer) writeMFString(name string, vals []string) {
	if len(vals) == 0 {
		return
	}
	if len(vals) == 1 {
		wr.line("%s \"%s\"", name, vals[0])
		return
	}
	wr.line("%s [", name)
	wr.indent++
	for _, s := range vals {
		wr.line("\"%s\"", s)
	}
	wr.indent--
	wr.line("]")
}

func (wr *Writer) writeMFFloat(name string, vals []float64) {
	if len(vals) == 0 {
		return
	}
	wr.line("%s [", name)
	wr.indent++
	s := wr.indentStr()
	for i, v := range vals {
		if i%8 == 0 && i > 0 {
			wr.printf("\n%s", s)
		}
		if i > 0 {
			wr.printf(" ")
		} else {
			wr.printf("%s", s)
		}
		wr.printf("%g", v)
		if i < len(vals)-1 {
			wr.printf(",")
		}
	}
	wr.printf("\n")
	wr.indent--
	wr.line("]")
}

func (wr *Writer) writeMFInt32(name string, vals []int64) {
	if len(vals) == 0 {
		return
	}
	wr.line("%s [", name)
	wr.indent++
	s := wr.indentStr()
	wr.printf("%s", s)
	col := 0
	for i, v := range vals {
		wr.printf("%d", v)
		last := i == len(vals)-1
		if !last {
			wr.printf(",")
		}
		if v == -1 {
			wr.printf("\n")
			if !last {
				wr.printf("%s", s)
			}
			col = 0
		} else {
			col++
			if col >= 20 {
				wr.printf("\n")
				if !last {
					wr.printf("%s", s)
				}
				col = 0
			} else if !last {
				wr.printf(" ")
			}
		}
	}
	if col > 0 {
		wr.printf("\n")
	}
	wr.indent--
	wr.line("]")
}

func (wr *Writer) writeMFColor(name string, vals []vec.SFColor) {
	if len(vals) == 0 {
		return
	}
	wr.line("%s [", name)
	wr.indent++
	for _, c := range vals {
		wr.line("%s,", fmtColor(c))
	}
	wr.indent--
	wr.line("]")
}

func (wr *Writer) writeMFVec3f(name string, vals []vec.SFVec3f) {
	if len(vals) == 0 {
		return
	}
	wr.line("%s [", name)
	wr.indent++
	for _, v := range vals {
		wr.line("%s,", fmtVec3(v))
	}
	wr.indent--
	wr.line("]")
}

func (wr *Writer) writeMFVec2f(name string, vals []vec.SFVec2f) {
	if len(vals) == 0 {
		return
	}
	wr.line("%s [", name)
	wr.indent++
	for _, v := range vals {
		wr.line("%s,", fmtVec2(v))
	}
	wr.indent--
	wr.line("]")
}

func (wr *Writer) writeMFRotation(name string, vals []vec.SFRotation) {
	if len(vals) == 0 {
		return
	}
	wr.line("%s [", name)
	wr.indent++
	for _, r := range vals {
		wr.line("%s,", fmtRot(r))
	}
	wr.indent--
	wr.line("]")
}

func (wr *Writer) writeMFNode(name string, vals []node.Node) {
	if len(vals) == 0 {
		return
	}
	wr.line("%s [", name)
	wr.indent++
	for _, n := range vals {
		wr.writeNode(n)
	}
	wr.indent--
	wr.line("]")
}

func (wr *Writer) writeAppearance(n *node.Appearance) {
	wr.printf("Appearance {\n")
	wr.indent++
	wr.writeSFNode("material", n.Material)
	wr.writeSFNode("texture", n.Texture)
	wr.writeSFNode("textureTransform", n.TextureTransform)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeMaterial(n *node.Material) {
	def := node.NewMaterial()
	wr.printf("Material {\n")
	wr.indent++
	if n.AmbientIntensity != def.AmbientIntensity {
		wr.line("ambientIntensity %g", n.AmbientIntensity)
	}
	if !n.DiffuseColor.Eq(def.DiffuseColor) {
		wr.line("diffuseColor %s", fmtColor(n.DiffuseColor))
	}
	if !n.EmissiveColor.Eq(def.EmissiveColor) {
		wr.line("emissiveColor %s", fmtColor(n.EmissiveColor))
	}
	if n.Shininess != def.Shininess {
		wr.line("shininess %g", n.Shininess)
	}
	if !n.SpecularColor.Eq(def.SpecularColor) {
		wr.line("specularColor %s", fmtColor(n.SpecularColor))
	}
	if n.Transparency != def.Transparency {
		wr.line("transparency %g", n.Transparency)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeImageTexture(n *node.ImageTexture) {
	def := node.NewImageTexture()
	wr.printf("ImageTexture {\n")
	wr.indent++
	wr.writeMFString("url", n.URL)
	if n.RepeatS != def.RepeatS {
		wr.line("repeatS %s", fmtBool(n.RepeatS))
	}
	if n.RepeatT != def.RepeatT {
		wr.line("repeatT %s", fmtBool(n.RepeatT))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeMovieTexture(n *node.MovieTexture) {
	def := node.NewMovieTexture()
	wr.printf("MovieTexture {\n")
	wr.indent++
	wr.writeMFString("url", n.URL)
	if n.Loop != def.Loop {
		wr.line("loop %s", fmtBool(n.Loop))
	}
	if n.Speed != def.Speed {
		wr.line("speed %g", n.Speed)
	}
	if n.StartTime != def.StartTime {
		wr.line("startTime %g", n.StartTime)
	}
	if n.StopTime != def.StopTime {
		wr.line("stopTime %g", n.StopTime)
	}
	if n.RepeatS != def.RepeatS {
		wr.line("repeatS %s", fmtBool(n.RepeatS))
	}
	if n.RepeatT != def.RepeatT {
		wr.line("repeatT %s", fmtBool(n.RepeatT))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writePixelTexture(n *node.PixelTexture) {
	def := node.NewPixelTexture()
	wr.printf("PixelTexture {\n")
	wr.indent++
	img := n.Image
	if img.Width > 0 || img.Height > 0 || len(img.Pixels) > 0 {
		wr.printf("%simage %d %d %d", wr.indentStr(), img.Width, img.Height, img.NumComponents)
		nPixels := int(img.Width) * int(img.Height)
		nc := int(img.NumComponents)
		for i := 0; i < nPixels; i++ {
			off := i * nc
			var pixel uint32
			switch nc {
			case 1:
				pixel = uint32(img.Pixels[off])
			case 2:
				pixel = uint32(img.Pixels[off])<<8 | uint32(img.Pixels[off+1])
			case 3:
				pixel = uint32(img.Pixels[off])<<16 | uint32(img.Pixels[off+1])<<8 | uint32(img.Pixels[off+2])
			case 4:
				pixel = uint32(img.Pixels[off])<<24 | uint32(img.Pixels[off+1])<<16 | uint32(img.Pixels[off+2])<<8 | uint32(img.Pixels[off+3])
			}
			wr.printf(" %d", pixel)
		}
		wr.printf("\n")
	}
	if n.RepeatS != def.RepeatS {
		wr.line("repeatS %s", fmtBool(n.RepeatS))
	}
	if n.RepeatT != def.RepeatT {
		wr.line("repeatT %s", fmtBool(n.RepeatT))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeTextureTransform(n *node.TextureTransform) {
	def := node.NewTextureTransform()
	zero2 := vec.SFVec2f{}
	wr.printf("TextureTransform {\n")
	wr.indent++
	if n.Center != zero2 {
		wr.line("center %s", fmtVec2(n.Center))
	}
	if n.Rotation != def.Rotation {
		wr.line("rotation %g", n.Rotation)
	}
	if n.Scale != def.Scale {
		wr.line("scale %s", fmtVec2(n.Scale))
	}
	if n.Translation != zero2 {
		wr.line("translation %s", fmtVec2(n.Translation))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeFontStyle(n *node.FontStyle) {
	def := node.NewFontStyle()
	wr.printf("FontStyle {\n")
	wr.indent++
	if n.Family != def.Family {
		wr.line("family \"%s\"", n.Family)
	}
	if n.Horizontal != def.Horizontal {
		wr.line("horizontal %s", fmtBool(n.Horizontal))
	}
	if len(n.Justify) > 0 {
		wr.writeMFString("justify", n.Justify)
	}
	if n.Language != def.Language {
		wr.line("language \"%s\"", n.Language)
	}
	if n.LeftToRight != def.LeftToRight {
		wr.line("leftToRight %s", fmtBool(n.LeftToRight))
	}
	if n.Size != def.Size {
		wr.line("size %g", n.Size)
	}
	if n.Spacing != def.Spacing {
		wr.line("spacing %g", n.Spacing)
	}
	if n.Style != def.Style {
		wr.line("style \"%s\"", n.Style)
	}
	if n.TopToBottom != def.TopToBottom {
		wr.line("topToBottom %s", fmtBool(n.TopToBottom))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeBackground(n *node.Background) {
	wr.printf("Background {\n")
	wr.indent++
	wr.writeMFFloat("groundAngle", n.GroundAngle)
	wr.writeMFColor("groundColor", n.GroundColor)
	wr.writeMFFloat("skyAngle", n.SkyAngle)
	wr.writeMFColor("skyColor", n.SkyColor)
	wr.writeMFString("backUrl", n.BackURL)
	wr.writeMFString("bottomUrl", n.BottomURL)
	wr.writeMFString("frontUrl", n.FrontURL)
	wr.writeMFString("leftUrl", n.LeftURL)
	wr.writeMFString("rightUrl", n.RightURL)
	wr.writeMFString("topUrl", n.TopURL)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeFog(n *node.Fog) {
	def := node.NewFog()
	wr.printf("Fog {\n")
	wr.indent++
	if !n.Color.Eq(def.Color) {
		wr.line("color %s", fmtColor(n.Color))
	}
	if n.FogType != def.FogType {
		wr.line("fogType \"%s\"", n.FogType)
	}
	if n.VisibilityRange != def.VisibilityRange {
		wr.line("visibilityRange %g", n.VisibilityRange)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeNavigationInfo(n *node.NavigationInfo) {
	def := node.NewNavigationInfo()
	wr.printf("NavigationInfo {\n")
	wr.indent++
	wr.writeMFFloat("avatarSize", n.AvatarSize)
	if n.Headlight != def.Headlight {
		wr.line("headlight %s", fmtBool(n.Headlight))
	}
	if n.Speed != def.Speed {
		wr.line("speed %g", n.Speed)
	}
	wr.writeMFString("type", n.Type)
	if n.VisibilityLimit != def.VisibilityLimit {
		wr.line("visibilityLimit %g", n.VisibilityLimit)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeViewpoint(n *node.Viewpoint) {
	def := node.NewViewpoint()
	wr.printf("Viewpoint {\n")
	wr.indent++
	if n.FieldOfView != def.FieldOfView {
		wr.line("fieldOfView %g", n.FieldOfView)
	}
	if n.Jump != def.Jump {
		wr.line("jump %s", fmtBool(n.Jump))
	}
	if n.Orientation != def.Orientation {
		wr.line("orientation %s", fmtRot(n.Orientation))
	}
	if n.Position != def.Position {
		wr.line("position %s", fmtVec3(n.Position))
	}
	if n.Description != "" {
		wr.line("description \"%s\"", n.Description)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeShape(n *node.Shape) {
	wr.printf("Shape {\n")
	wr.indent++
	wr.writeSFNode("appearance", n.Appearance)
	if n.Geometry != nil {
		wr.writeSFNode("geometry", n.Geometry)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeLightFields(n *node.Light, def *node.Light) {
	if n.On != def.On {
		wr.line("on %s", fmtBool(n.On))
	}
	if !n.Color.Eq(def.Color) {
		wr.line("color %s", fmtColor(n.Color))
	}
	if n.Intensity != def.Intensity {
		wr.line("intensity %g", n.Intensity)
	}
	if n.AmbientIntensity != def.AmbientIntensity {
		wr.line("ambientIntensity %g", n.AmbientIntensity)
	}
	if n.Attenuation != def.Attenuation {
		wr.line("attenuation %s", fmtVec3(n.Attenuation))
	}
}

func (wr *Writer) writeDirectionalLight(n *node.DirectionalLight) {
	def := node.NewDirectionalLight()
	wr.printf("DirectionalLight {\n")
	wr.indent++
	wr.writeLightFields(&n.Light, &def.Light)
	if n.Direction != def.Direction {
		wr.line("direction %s", fmtVec3(n.Direction))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writePointLight(n *node.PointLight) {
	def := node.NewPointLight()
	wr.printf("PointLight {\n")
	wr.indent++
	wr.writeLightFields(&n.Light, &def.Light)
	if n.Location != def.Location {
		wr.line("location %s", fmtVec3(n.Location))
	}
	if n.Radius != def.Radius {
		wr.line("radius %g", n.Radius)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeSpotLight(n *node.SpotLight) {
	def := node.NewSpotLight()
	wr.printf("SpotLight {\n")
	wr.indent++
	wr.writeLightFields(&n.Light, &def.Light)
	if n.BeamWidth != def.BeamWidth {
		wr.line("beamWidth %g", n.BeamWidth)
	}
	if n.CutOffAngle != def.CutOffAngle {
		wr.line("cutOffAngle %g", n.CutOffAngle)
	}
	if n.Direction != def.Direction {
		wr.line("direction %s", fmtVec3(n.Direction))
	}
	if n.Location != def.Location {
		wr.line("location %s", fmtVec3(n.Location))
	}
	if n.Radius != def.Radius {
		wr.line("radius %g", n.Radius)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeWorldInfo(n *node.WorldInfo) {
	wr.printf("WorldInfo {\n")
	wr.indent++
	wr.writeMFString("info", n.Info)
	if n.Title != "" {
		wr.line("title \"%s\"", n.Title)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeScript(n *node.Script) {
	wr.printf("Script {\n")
	wr.indent++
	wr.writeMFString("url", n.URL)
	if n.DirectOutput {
		wr.line("directOutput TRUE")
	}
	if n.MustEvaluate {
		wr.line("mustEvaluate TRUE")
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeSound(n *node.Sound) {
	def := node.NewSound()
	wr.printf("Sound {\n")
	wr.indent++
	if n.Direction != def.Direction {
		wr.line("direction %s", fmtVec3(n.Direction))
	}
	if n.Intensity != def.Intensity {
		wr.line("intensity %g", n.Intensity)
	}
	if n.Location != def.Location {
		wr.line("location %s", fmtVec3(n.Location))
	}
	if n.MaxBack != def.MaxBack {
		wr.line("maxBack %g", n.MaxBack)
	}
	if n.MaxFront != def.MaxFront {
		wr.line("maxFront %g", n.MaxFront)
	}
	if n.MinBack != def.MinBack {
		wr.line("minBack %g", n.MinBack)
	}
	if n.MinFront != def.MinFront {
		wr.line("minFront %g", n.MinFront)
	}
	if n.Priority != def.Priority {
		wr.line("priority %g", n.Priority)
	}
	wr.writeSFNode("source", n.Source)
	if n.Spatialize != def.Spatialize {
		wr.line("spatialize %s", fmtBool(n.Spatialize))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeAudioClip(n *node.AudioClip) {
	def := node.NewAudioClip()
	wr.printf("AudioClip {\n")
	wr.indent++
	if n.Description != "" {
		wr.line("description \"%s\"", n.Description)
	}
	if n.Loop != def.Loop {
		wr.line("loop %s", fmtBool(n.Loop))
	}
	if n.Pitch != def.Pitch {
		wr.line("pitch %g", n.Pitch)
	}
	if n.StartTime != def.StartTime {
		wr.line("startTime %g", n.StartTime)
	}
	if n.StopTime != def.StopTime {
		wr.line("stopTime %g", n.StopTime)
	}
	wr.writeMFString("url", n.URL)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeTransform(n *node.Transform) {
	def := node.NewTransform()
	zero3 := vec.SFVec3f{}
	wr.printf("Transform {\n")
	wr.indent++
	if n.Center != zero3 {
		wr.line("center %s", fmtVec3(n.Center))
	}
	if n.Rotation != def.Rotation {
		wr.line("rotation %s", fmtRot(n.Rotation))
	}
	if n.Scale != def.Scale {
		wr.line("scale %s", fmtVec3(n.Scale))
	}
	if n.ScaleOrientation != def.ScaleOrientation {
		wr.line("scaleOrientation %s", fmtRot(n.ScaleOrientation))
	}
	if n.Translation != zero3 {
		wr.line("translation %s", fmtVec3(n.Translation))
	}
	wr.writeChildren(n.Children)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeGroup(n *node.Group) {
	wr.printf("Group {\n")
	wr.indent++
	wr.writeChildren(n.Children)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeAnchor(n *node.Anchor) {
	wr.printf("Anchor {\n")
	wr.indent++
	if n.Description != "" {
		wr.line("description \"%s\"", n.Description)
	}
	wr.writeMFString("parameter", n.Parameter)
	wr.writeMFString("url", n.URL)
	wr.writeChildren(n.Children)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeBillboard(n *node.Billboard) {
	def := node.NewBillboard()
	wr.printf("Billboard {\n")
	wr.indent++
	if n.AxisOfRotation != def.AxisOfRotation {
		wr.line("axisOfRotation %s", fmtVec3(n.AxisOfRotation))
	}
	wr.writeChildren(n.Children)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeCollision(n *node.Collision) {
	def := node.NewCollision()
	wr.printf("Collision {\n")
	wr.indent++
	if n.Collide != def.Collide {
		wr.line("collide %s", fmtBool(n.Collide))
	}
	wr.writeSFNode("proxy", n.Proxy)
	wr.writeChildren(n.Children)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeInline(n *node.Inline) {
	wr.printf("Inline {\n")
	wr.indent++
	wr.writeMFString("url", n.URL)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeLOD(n *node.LOD) {
	zero3 := vec.SFVec3f{}
	wr.printf("LOD {\n")
	wr.indent++
	if n.Center != zero3 {
		wr.line("center %s", fmtVec3(n.Center))
	}
	wr.writeMFFloat("range", n.Range)
	wr.writeMFNode("level", n.Level)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeSwitch(n *node.Switch) {
	def := node.NewSwitch()
	wr.printf("Switch {\n")
	wr.indent++
	if n.WhichChoice != def.WhichChoice {
		wr.line("whichChoice %d", n.WhichChoice)
	}
	wr.writeMFNode("choice", n.Choice)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeBox(n *node.Box) {
	def := node.NewBox()
	wr.printf("Box {\n")
	wr.indent++
	if n.Size != def.Size {
		wr.line("size %s", fmtVec3(n.Size))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeCone(n *node.Cone) {
	def := node.NewCone()
	wr.printf("Cone {\n")
	wr.indent++
	if n.BottomRadius != def.BottomRadius {
		wr.line("bottomRadius %g", n.BottomRadius)
	}
	if n.Height != def.Height {
		wr.line("height %g", n.Height)
	}
	if n.Side != def.Side {
		wr.line("side %s", fmtBool(n.Side))
	}
	if n.Bottom != def.Bottom {
		wr.line("bottom %s", fmtBool(n.Bottom))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeCylinder(n *node.Cylinder) {
	def := node.NewCylinder()
	wr.printf("Cylinder {\n")
	wr.indent++
	if n.Bottom != def.Bottom {
		wr.line("bottom %s", fmtBool(n.Bottom))
	}
	if n.Height != def.Height {
		wr.line("height %g", n.Height)
	}
	if n.Radius != def.Radius {
		wr.line("radius %g", n.Radius)
	}
	if n.Side != def.Side {
		wr.line("side %s", fmtBool(n.Side))
	}
	if n.Top != def.Top {
		wr.line("top %s", fmtBool(n.Top))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeSphere(n *node.Sphere) {
	def := node.NewSphere()
	wr.printf("Sphere {\n")
	wr.indent++
	if n.Radius != def.Radius {
		wr.line("radius %g", n.Radius)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeGeomFields(n *node.BaseGeometry, def *node.BaseGeometry) {
	if n.Ccw != def.Ccw {
		wr.line("ccw %s", fmtBool(n.Ccw))
	}
	if n.Convex != def.Convex {
		wr.line("convex %s", fmtBool(n.Convex))
	}
	if n.CreaseAngle != def.CreaseAngle {
		wr.line("creaseAngle %g", n.CreaseAngle)
	}
	if n.IsSolid != def.IsSolid {
		wr.line("solid %s", fmtBool(n.IsSolid))
	}
}

func (wr *Writer) writeExtrusion(n *node.Extrusion) {
	def := node.NewExtrusion()
	wr.printf("Extrusion {\n")
	wr.indent++
	wr.writeGeomFields(&n.BaseGeometry, &def.BaseGeometry)
	if n.BeginCap != def.BeginCap {
		wr.line("beginCap %s", fmtBool(n.BeginCap))
	}
	wr.writeMFVec2f("crossSection", n.CrossSection)
	if n.EndCap != def.EndCap {
		wr.line("endCap %s", fmtBool(n.EndCap))
	}
	wr.writeMFRotation("orientation", n.Orientation)
	wr.writeMFVec2f("scale", n.Scale)
	wr.writeMFVec3f("spine", n.Spine)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeText(n *node.Text) {
	wr.printf("Text {\n")
	wr.indent++
	wr.writeMFString("string", n.String)
	wr.writeSFNode("fontStyle", n.FontStyle)
	wr.writeMFFloat("length", n.Length)
	if n.MaxExtent != 0 {
		wr.line("maxExtent %g", n.MaxExtent)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeDataSetFields(n *node.DataSet, def *node.DataSet) {
	wr.writeGeomFields(&n.BaseGeometry, &def.BaseGeometry)
	if n.ColorPerVertex != def.ColorPerVertex {
		wr.line("colorPerVertex %s", fmtBool(n.ColorPerVertex))
	}
	if n.NormalPerVertex != def.NormalPerVertex {
		wr.line("normalPerVertex %s", fmtBool(n.NormalPerVertex))
	}
	wr.writeSFNode("color", n.Color)
	wr.writeSFNode("coord", n.Coord)
	wr.writeSFNode("normal", n.Normal)
	wr.writeSFNode("texCoord", n.TexCoord)
}

func (wr *Writer) writeIndexedFaceSet(n *node.IndexedFaceSet) {
	def := node.NewIndexedFaceSet()
	wr.printf("IndexedFaceSet {\n")
	wr.indent++
	wr.writeDataSetFields(&n.DataSet, &def.DataSet)
	wr.writeMFInt32("colorIndex", n.ColorIndex)
	wr.writeMFInt32("coordIndex", n.CoordIndex)
	wr.writeMFInt32("normalIndex", n.NormalIndex)
	wr.writeMFInt32("texCoordIndex", n.TexCoordIndex)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeIndexedLineSet(n *node.IndexedLineSet) {
	def := node.NewIndexedLineSet()
	wr.printf("IndexedLineSet {\n")
	wr.indent++
	wr.writeDataSetFields(&n.DataSet, &def.DataSet)
	wr.writeMFInt32("colorIndex", n.ColorIndex)
	wr.writeMFInt32("coordIndex", n.CoordIndex)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writePointSet(n *node.PointSet) {
	def := node.NewPointSet()
	wr.printf("PointSet {\n")
	wr.indent++
	wr.writeDataSetFields(&n.DataSet, &def.DataSet)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeElevationGrid(n *node.ElevationGrid) {
	def := node.NewElevationGrid()
	wr.printf("ElevationGrid {\n")
	wr.indent++
	wr.writeDataSetFields(&n.DataSet, &def.DataSet)
	wr.writeMFFloat("height", n.Heights)
	if n.XDimension != def.XDimension {
		wr.line("xDimension %d", n.XDimension)
	}
	if n.XSpacing != def.XSpacing {
		wr.line("xSpacing %g", n.XSpacing)
	}
	if n.ZDimension != def.ZDimension {
		wr.line("zDimension %d", n.ZDimension)
	}
	if n.ZSpacing != def.ZSpacing {
		wr.line("zSpacing %g", n.ZSpacing)
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeColorNode(n *node.ColorNode) {
	wr.printf("Color {\n")
	wr.indent++
	wr.writeMFColor("color", n.Color)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeCoordinate(n *node.Coordinate) {
	wr.printf("Coordinate {\n")
	wr.indent++
	wr.writeMFVec3f("point", n.Point)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeNormalNode(n *node.NormalNode) {
	wr.printf("Normal {\n")
	wr.indent++
	wr.writeMFVec3f("vector", n.Vector)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeTextureCoordinate(n *node.TextureCoordinate) {
	wr.printf("TextureCoordinate {\n")
	wr.indent++
	wr.writeMFVec2f("point", n.Point)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeColorInterpolator(n *node.ColorInterpolator) {
	wr.printf("ColorInterpolator {\n")
	wr.indent++
	wr.writeMFFloat("key", n.Key)
	wr.writeMFColor("keyValue", n.KeyValue)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeCoordinateInterpolator(n *node.CoordinateInterpolator) {
	wr.printf("CoordinateInterpolator {\n")
	wr.indent++
	wr.writeMFFloat("key", n.Key)
	wr.writeMFVec3f("keyValue", n.KeyValue)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeNormalInterpolator(n *node.NormalInterpolator) {
	wr.printf("NormalInterpolator {\n")
	wr.indent++
	wr.writeMFFloat("key", n.Key)
	wr.writeMFVec3f("keyValue", n.KeyValue)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeOrientationInterpolator(n *node.OrientationInterpolator) {
	wr.printf("OrientationInterpolator {\n")
	wr.indent++
	wr.writeMFFloat("key", n.Key)
	wr.writeMFRotation("keyValue", n.KeyValue)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writePositionInterpolator(n *node.PositionInterpolator) {
	wr.printf("PositionInterpolator {\n")
	wr.indent++
	wr.writeMFFloat("key", n.Key)
	wr.writeMFVec3f("keyValue", n.KeyValue)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeScalarInterpolator(n *node.ScalarInterpolator) {
	wr.printf("ScalarInterpolator {\n")
	wr.indent++
	wr.writeMFFloat("key", n.Key)
	wr.writeMFFloat("keyValue", n.KeyValue)
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeTouchSensor(n *node.TouchSensor) {
	wr.printf("TouchSensor {\n")
	wr.indent++
	if !n.Enabled {
		wr.line("enabled FALSE")
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeTimeSensor(n *node.TimeSensor) {
	def := node.NewTimeSensor()
	wr.printf("TimeSensor {\n")
	wr.indent++
	if n.CycleInterval != def.CycleInterval {
		wr.line("cycleInterval %g", n.CycleInterval)
	}
	if n.Loop != def.Loop {
		wr.line("loop %s", fmtBool(n.Loop))
	}
	if n.StartTime != def.StartTime {
		wr.line("startTime %g", n.StartTime)
	}
	if n.StopTime != def.StopTime {
		wr.line("stopTime %g", n.StopTime)
	}
	if n.Enabled != def.Enabled {
		wr.line("enabled %s", fmtBool(n.Enabled))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeProximitySensor(n *node.ProximitySensor) {
	zero3 := vec.SFVec3f{}
	wr.printf("ProximitySensor {\n")
	wr.indent++
	if n.Center != zero3 {
		wr.line("center %s", fmtVec3(n.Center))
	}
	if n.Size != zero3 {
		wr.line("size %s", fmtVec3(n.Size))
	}
	if !n.Enabled {
		wr.line("enabled FALSE")
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeCylinderSensor(n *node.CylinderSensor) {
	def := node.NewCylinderSensor()
	wr.printf("CylinderSensor {\n")
	wr.indent++
	if n.AutoOffset != def.AutoOffset {
		wr.line("autoOffset %s", fmtBool(n.AutoOffset))
	}
	if n.DiskAngle != def.DiskAngle {
		wr.line("diskAngle %g", n.DiskAngle)
	}
	if n.MaxAngle != def.MaxAngle {
		wr.line("maxAngle %g", n.MaxAngle)
	}
	if n.MinAngle != def.MinAngle {
		wr.line("minAngle %g", n.MinAngle)
	}
	if n.Offset != def.Offset {
		wr.line("offset %g", n.Offset)
	}
	if n.Enabled != def.Enabled {
		wr.line("enabled %s", fmtBool(n.Enabled))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writePlaneSensor(n *node.PlaneSensor) {
	def := node.NewPlaneSensor()
	zero3 := vec.SFVec3f{}
	wr.printf("PlaneSensor {\n")
	wr.indent++
	if n.AutoOffset != def.AutoOffset {
		wr.line("autoOffset %s", fmtBool(n.AutoOffset))
	}
	if n.MaxPosition != def.MaxPosition {
		wr.line("maxPosition %s", fmtVec2(n.MaxPosition))
	}
	if n.MinPosition != def.MinPosition {
		wr.line("minPosition %s", fmtVec2(n.MinPosition))
	}
	if n.Offset != zero3 {
		wr.line("offset %s", fmtVec3(n.Offset))
	}
	if n.Enabled != def.Enabled {
		wr.line("enabled %s", fmtBool(n.Enabled))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeSphereSensor(n *node.SphereSensor) {
	def := node.NewSphereSensor()
	wr.printf("SphereSensor {\n")
	wr.indent++
	if n.AutoOffset != def.AutoOffset {
		wr.line("autoOffset %s", fmtBool(n.AutoOffset))
	}
	if n.Offset != def.Offset {
		wr.line("offset %s", fmtRot(n.Offset))
	}
	if n.Enabled != def.Enabled {
		wr.line("enabled %s", fmtBool(n.Enabled))
	}
	wr.indent--
	wr.line("}")
}

func (wr *Writer) writeVisibilitySensor(n *node.VisibilitySensor) {
	zero3 := vec.SFVec3f{}
	wr.printf("VisibilitySensor {\n")
	wr.indent++
	if n.Center != zero3 {
		wr.line("center %s", fmtVec3(n.Center))
	}
	if n.Size != zero3 {
		wr.line("size %s", fmtVec3(n.Size))
	}
	if !n.Enabled {
		wr.line("enabled FALSE")
	}
	wr.indent--
	wr.line("}")
}
