package writer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func writeScene(nodes []node.Node) string {
	var buf bytes.Buffer
	w := New(&buf)
	w.WriteScene(nodes)
	return buf.String()
}

func parseVRML(vrml string) []node.Node {
	p := parser.NewParser(strings.NewReader(vrml))
	return p.Parse()
}

func TestWriteScene_Header(t *testing.T) {
	out := writeScene(nil)
	if !strings.HasPrefix(out, "#VRML V2.0 utf8") {
		t.Fatalf("bad header")
	}
}

func TestRoundTrip_Box(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nShape { geometry Box { size 2 3 4 } }"
	nodes := parseVRML(vrml)
	out1 := writeScene(nodes)
	nodes2 := parseVRML(out1)
	out2 := writeScene(nodes2)
	if out1 != out2 {
		t.Fatalf("round trip mismatch")
	}
}

func TestRoundTrip_Transform(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nTransform { translation 1 2 3 children [ Shape { geometry Sphere { radius 5 } } ] }"
	nodes := parseVRML(vrml)
	out1 := writeScene(nodes)
	nodes2 := parseVRML(out1)
	out2 := writeScene(nodes2)
	if out1 != out2 {
		t.Fatalf("round trip mismatch")
	}
}

func TestRoundTrip_Material(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nShape { appearance Appearance { material Material { diffuseColor 1 0 0 } } geometry Sphere {} }"
	nodes := parseVRML(vrml)
	out1 := writeScene(nodes)
	nodes2 := parseVRML(out1)
	out2 := writeScene(nodes2)
	if out1 != out2 {
		t.Fatalf("round trip mismatch")
	}
}

func TestRoundTrip_DEF_USE(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nDEF MyMat Material { diffuseColor 1 0 0 }\nShape { appearance Appearance { material USE MyMat } geometry Box {} }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "DEF MyMat") {
		t.Fatalf("missing DEF")
	}
	if !strings.Contains(out, "USE MyMat") {
		t.Fatalf("missing USE")
	}
}

func TestWrite_NavigationInfo(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nNavigationInfo { speed 2 headlight FALSE }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "NavigationInfo") {
		t.Fatalf("missing NavigationInfo")
	}
	if !strings.Contains(out, "speed 2") {
		t.Fatalf("missing speed")
	}
}

func TestWrite_DirectionalLight(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nDirectionalLight { direction 0 -1 0 intensity 0.8 }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "DirectionalLight") {
		t.Fatalf("missing node")
	}
}

func TestWrite_TimeSensor(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nTimeSensor { cycleInterval 5 loop TRUE }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "cycleInterval 5") {
		t.Fatalf("missing interval")
	}
	if !strings.Contains(out, "loop TRUE") {
		t.Fatalf("missing loop")
	}
}

func TestWrite_EmptyScene(t *testing.T) {
	out := writeScene([]node.Node{})
	if !strings.HasPrefix(out, "#VRML V2.0 utf8") {
		t.Fatalf("bad output")
	}
}

func TestWrite_ProgrammaticMaterial(t *testing.T) {
	m := node.NewMaterial()
	m.DiffuseColor = vec.SFColor{R: 1, G: 0, B: 0, A: 1}
	out := writeScene([]node.Node{m})
	if !strings.Contains(out, "Material") {
		t.Fatalf("missing Material")
	}
	if !strings.Contains(out, "diffuseColor 1 0 0") {
		t.Fatalf("missing color")
	}
}

func TestWrite_Switch(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nSwitch { whichChoice 1 choice [ Shape { geometry Box {} } Shape { geometry Sphere {} } ] }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "Switch") {
		t.Fatalf("missing Switch")
	}
	if !strings.Contains(out, "whichChoice 1") {
		t.Fatalf("missing whichChoice")
	}
}

func TestWrite_LOD(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nLOD { range [50 100] level [ Shape { geometry Box {} } Shape { geometry Sphere {} } ] }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "LOD") {
		t.Fatalf("missing LOD")
	}
}

// ===========================================================================
// Gap-filling tests (issue #43)
// ===========================================================================

// roundTrip parses VRML, writes it, re-parses, writes again and checks idempotency.
func roundTrip(t *testing.T, vrml string) string {
	t.Helper()
	nodes := parseVRML(vrml)
	out1 := writeScene(nodes)
	nodes2 := parseVRML(out1)
	out2 := writeScene(nodes2)
	if out1 != out2 {
		t.Fatalf("round-trip mismatch:\n--- first ---\n%s\n--- second ---\n%s", out1, out2)
	}
	return out1
}

// mustContain checks that out contains all the given substrings.
func mustContain(t *testing.T, out string, substrs ...string) {
	t.Helper()
	for _, s := range substrs {
		if !strings.Contains(out, s) {
			t.Errorf("output missing %q:\n%s", s, out)
		}
	}
}

// ---------------------------------------------------------------------------
// Geometry round-trips
// ---------------------------------------------------------------------------

func TestRoundTrip_Cone(t *testing.T) {
	out := roundTrip(t, "#VRML V2.0 utf8\nShape { geometry Cone { bottomRadius 3 height 5 side FALSE } }")
	mustContain(t, out, "Cone", "bottomRadius 3", "height 5", "side FALSE")
}

func TestRoundTrip_Cylinder(t *testing.T) {
	out := roundTrip(t, "#VRML V2.0 utf8\nShape { geometry Cylinder { radius 2 height 4 top FALSE } }")
	mustContain(t, out, "Cylinder", "radius 2", "height 4", "top FALSE")
}

func TestRoundTrip_IndexedFaceSet(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  geometry IndexedFaceSet {
    coord Coordinate { point [0 0 0, 1 0 0, 0 1 0] }
    coordIndex [0 1 2 -1]
  }
}`
	out := roundTrip(t, vrml)
	mustContain(t, out, "IndexedFaceSet", "Coordinate", "coordIndex")
}

func TestRoundTrip_IndexedLineSet(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  geometry IndexedLineSet {
    coord Coordinate { point [0 0 0, 1 0 0, 1 1 0] }
    coordIndex [0 1 2 -1]
  }
}`
	out := roundTrip(t, vrml)
	mustContain(t, out, "IndexedLineSet", "coordIndex")
}

func TestRoundTrip_PointSet(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  geometry PointSet {
    coord Coordinate { point [0 0 0, 1 1 1] }
    color Color { color [1 0 0, 0 1 0] }
  }
}`
	out := roundTrip(t, vrml)
	mustContain(t, out, "PointSet", "Coordinate", "Color")
}

func TestRoundTrip_ElevationGrid(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  geometry ElevationGrid {
    xDimension 3
    zDimension 2
    xSpacing 2
    zSpacing 3
    height [0 1 0 1 2 1]
  }
}`
	out := roundTrip(t, vrml)
	mustContain(t, out, "ElevationGrid", "xDimension 3", "zDimension 2", "xSpacing 2", "zSpacing 3")
}

func TestRoundTrip_Extrusion(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  geometry Extrusion {
    crossSection [1 1, 1 -1, -1 -1, -1 1, 1 1]
    spine [0 0 0, 0 1 0]
  }
}`
	out := roundTrip(t, vrml)
	mustContain(t, out, "Extrusion", "crossSection", "spine")
}

func TestWrite_Text_Programmatic(t *testing.T) {
	fs := node.NewFontStyle()
	fs.Size = 2
	fs.Family = "SERIF"
	txt := &node.Text{String: []string{"Hello", "World"}, FontStyle: fs}
	out := writeScene([]node.Node{&node.Shape{Geometry: txt}})
	mustContain(t, out, "Text", `"Hello"`, `"World"`, "FontStyle", "size 2")
}

// ---------------------------------------------------------------------------
// Container round-trips
// ---------------------------------------------------------------------------

func TestRoundTrip_Group(t *testing.T) {
	out := roundTrip(t, "#VRML V2.0 utf8\nGroup { children [ Shape { geometry Box {} } ] }")
	mustContain(t, out, "Group", "children", "Box")
}

func TestRoundTrip_Anchor(t *testing.T) {
	out := roundTrip(t, `#VRML V2.0 utf8
Anchor { url "http://example.com" description "Click me" children [ Shape { geometry Sphere {} } ] }`)
	mustContain(t, out, "Anchor", `"http://example.com"`, `"Click me"`)
}

func TestRoundTrip_Billboard(t *testing.T) {
	out := roundTrip(t, "#VRML V2.0 utf8\nBillboard { axisOfRotation 0 1 0 children [ Shape { geometry Box {} } ] }")
	mustContain(t, out, "Billboard")
}

func TestRoundTrip_Collision(t *testing.T) {
	out := roundTrip(t, "#VRML V2.0 utf8\nCollision { collide FALSE children [ Shape { geometry Box {} } ] }")
	mustContain(t, out, "Collision", "collide FALSE")
}

func TestRoundTrip_Inline(t *testing.T) {
	out := roundTrip(t, `#VRML V2.0 utf8
Inline { url "other.wrl" }`)
	mustContain(t, out, "Inline", `"other.wrl"`)
}

func TestRoundTrip_Viewpoint(t *testing.T) {
	out := roundTrip(t, `#VRML V2.0 utf8
Viewpoint { position 0 5 10 fieldOfView 1.2 description "Main" }`)
	mustContain(t, out, "Viewpoint", "position 0 5 10", "fieldOfView 1.2", `"Main"`)
}

// ---------------------------------------------------------------------------
// Light round-trips
// ---------------------------------------------------------------------------

func TestRoundTrip_PointLight(t *testing.T) {
	out := roundTrip(t, "#VRML V2.0 utf8\nPointLight { location 1 2 3 radius 50 }")
	mustContain(t, out, "PointLight", "location 1 2 3", "radius 50")
}

func TestRoundTrip_SpotLight(t *testing.T) {
	out := roundTrip(t, "#VRML V2.0 utf8\nSpotLight { location 0 5 0 direction 0 -1 0 beamWidth 0.5 cutOffAngle 1 }")
	mustContain(t, out, "SpotLight", "location 0 5 0", "beamWidth 0.5", "cutOffAngle 1")
}

// ---------------------------------------------------------------------------
// Texture round-trips
// ---------------------------------------------------------------------------

func TestRoundTrip_ImageTexture(t *testing.T) {
	out := roundTrip(t, `#VRML V2.0 utf8
Shape { appearance Appearance { texture ImageTexture { url "tex.png" repeatS FALSE } } geometry Box {} }`)
	mustContain(t, out, "ImageTexture", `"tex.png"`, "repeatS FALSE")
}

func TestRoundTrip_MovieTexture(t *testing.T) {
	out := roundTrip(t, `#VRML V2.0 utf8
Shape { appearance Appearance { texture MovieTexture { url "movie.mpg" loop TRUE speed 2 } } geometry Box {} }`)
	mustContain(t, out, "MovieTexture", `"movie.mpg"`, "loop TRUE", "speed 2")
}

func TestRoundTrip_TextureTransform(t *testing.T) {
	out := roundTrip(t, `#VRML V2.0 utf8
Shape { appearance Appearance { material Material {} textureTransform TextureTransform { scale 2 2 rotation 1.57 } } geometry Box {} }`)
	mustContain(t, out, "TextureTransform", "scale 2 2", "rotation 1.57")
}

func TestRoundTrip_PixelTexture(t *testing.T) {
	// 2x2 grayscale image: pixel values 10, 20, 30, 40
	vrml := "#VRML V2.0 utf8\nShape { appearance Appearance { texture PixelTexture { image 2 2 1 10 20 30 40 repeatS FALSE } } geometry Box {} }"
	out := roundTrip(t, vrml)
	mustContain(t, out, "PixelTexture", "image 2 2 1", "repeatS FALSE")
}

func TestRoundTrip_PixelTexture_RGB(t *testing.T) {
	// 1x2 RGB: red=(255<<16)=16711680, green=(255<<8)=65280
	vrml := "#VRML V2.0 utf8\nShape { appearance Appearance { texture PixelTexture { image 1 2 3 16711680 65280 } } geometry Box {} }"
	out := roundTrip(t, vrml)
	mustContain(t, out, "PixelTexture", "image 1 2 3 16711680 65280")
}

func TestWrite_PixelTexture_Programmatic(t *testing.T) {
	pt := node.NewPixelTexture()
	pt.Image = vec.SFImage{Width: 2, Height: 1, NumComponents: 3, Pixels: []uint8{255, 0, 0, 0, 255, 0}}
	out := writeScene([]node.Node{pt})
	// Red = (255<<16) = 16711680, Green = (255<<8) = 65280
	mustContain(t, out, "image 2 1 3 16711680 65280")
}

func TestWrite_PixelTexture_RGBA(t *testing.T) {
	pt := node.NewPixelTexture()
	pt.Image = vec.SFImage{Width: 1, Height: 1, NumComponents: 4, Pixels: []uint8{255, 128, 64, 32}}
	out := writeScene([]node.Node{pt})
	// (255<<24)|(128<<16)|(64<<8)|32 = 4286595104
	mustContain(t, out, "image 1 1 4 4286595104")
}

func TestWrite_PixelTexture_Empty(t *testing.T) {
	pt := node.NewPixelTexture()
	out := writeScene([]node.Node{pt})
	// Empty image → no image field, defaults → minimal output
	if strings.Contains(out, "image") {
		t.Fatal("should not write image for empty PixelTexture")
	}
}

// ---------------------------------------------------------------------------
// Interpolator round-trips
// ---------------------------------------------------------------------------

func TestRoundTrip_ColorInterpolator(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nColorInterpolator { key [0 1] keyValue [1 0 0, 0 0 1] }"
	out := roundTrip(t, vrml)
	mustContain(t, out, "ColorInterpolator", "key", "keyValue")
}

func TestRoundTrip_CoordinateInterpolator(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nCoordinateInterpolator { key [0 1] keyValue [0 0 0, 1 1 1] }"
	out := roundTrip(t, vrml)
	mustContain(t, out, "CoordinateInterpolator", "key", "keyValue")
}

func TestRoundTrip_NormalInterpolator(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nNormalInterpolator { key [0 1] keyValue [0 1 0, 1 0 0] }"
	out := roundTrip(t, vrml)
	mustContain(t, out, "NormalInterpolator", "key", "keyValue")
}

func TestRoundTrip_OrientationInterpolator(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nOrientationInterpolator { key [0 1] keyValue [0 1 0 0, 0 1 0 3.14] }"
	out := roundTrip(t, vrml)
	mustContain(t, out, "OrientationInterpolator", "key", "keyValue")
}

func TestRoundTrip_PositionInterpolator(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nPositionInterpolator { key [0 0.5 1] keyValue [0 0 0, 5 5 5, 10 0 0] }"
	out := roundTrip(t, vrml)
	mustContain(t, out, "PositionInterpolator", "key", "keyValue")
}

func TestRoundTrip_ScalarInterpolator(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nScalarInterpolator { key [0 1] keyValue [0 100] }"
	out := roundTrip(t, vrml)
	mustContain(t, out, "ScalarInterpolator", "key", "keyValue")
}

// ---------------------------------------------------------------------------
// Sensor round-trips
// ---------------------------------------------------------------------------

func TestRoundTrip_TouchSensor(t *testing.T) {
	out := roundTrip(t, "#VRML V2.0 utf8\nTransform { children [ TouchSensor { enabled FALSE } Shape { geometry Box {} } ] }")
	mustContain(t, out, "TouchSensor", "enabled FALSE")
}

func TestRoundTrip_ProximitySensor(t *testing.T) {
	out := roundTrip(t, "#VRML V2.0 utf8\nProximitySensor { center 1 2 3 size 10 10 10 }")
	mustContain(t, out, "ProximitySensor", "center 1 2 3", "size 10 10 10")
}

func TestWrite_CylinderSensor_Programmatic(t *testing.T) {
	cs := node.NewCylinderSensor()
	cs.MaxAngle = 3.14
	cs.MinAngle = -3.14
	out := writeScene([]node.Node{cs})
	mustContain(t, out, "CylinderSensor", "maxAngle 3.14", "minAngle -3.14")
}

func TestWrite_PlaneSensor_Programmatic(t *testing.T) {
	ps := node.NewPlaneSensor()
	ps.MaxPosition = vec.SFVec2f{X: 10, Y: 10}
	out := writeScene([]node.Node{ps})
	mustContain(t, out, "PlaneSensor", "maxPosition 10 10")
}

func TestWrite_SphereSensor_Programmatic(t *testing.T) {
	ss := node.NewSphereSensor()
	ss.AutoOffset = false
	out := writeScene([]node.Node{ss})
	mustContain(t, out, "SphereSensor", "autoOffset FALSE")
}

func TestWrite_VisibilitySensor_Programmatic(t *testing.T) {
	vs := node.NewVisibilitySensor()
	vs.Center = vec.SFVec3f{X: 0, Y: 5, Z: 0}
	vs.Size = vec.SFVec3f{X: 10, Y: 10, Z: 10}
	out := writeScene([]node.Node{vs})
	mustContain(t, out, "VisibilitySensor", "center 0 5 0", "size 10 10 10")
}

// ---------------------------------------------------------------------------
// Miscellaneous round-trips
// ---------------------------------------------------------------------------

func TestWrite_WorldInfo_Programmatic(t *testing.T) {
	wi := &node.WorldInfo{Title: "Test Scene", Info: []string{"Author: test", "Version: 1"}}
	out := writeScene([]node.Node{wi})
	mustContain(t, out, "WorldInfo", `"Test Scene"`, `"Author: test"`)
}

func TestWrite_Background_Programmatic(t *testing.T) {
	bg := &node.Background{
		SkyColor:    []vec.SFColor{{R: 0, G: 0, B: 1}},
		GroundColor: []vec.SFColor{{R: 0.3, G: 0.3, B: 0.3}},
		SkyAngle:    []float64{1.57},
		GroundAngle: []float64{1.57},
	}
	out := writeScene([]node.Node{bg})
	mustContain(t, out, "Background", "skyColor", "groundColor")
}

func TestRoundTrip_Fog(t *testing.T) {
	out := roundTrip(t, `#VRML V2.0 utf8
Fog { color 0.5 0.5 0.5 fogType "EXPONENTIAL" visibilityRange 100 }`)
	mustContain(t, out, "Fog", "color 0.5 0.5 0.5", `"EXPONENTIAL"`, "visibilityRange 100")
}

func TestWrite_Sound_AudioClip_Programmatic(t *testing.T) {
	ac := node.NewAudioClip()
	ac.URL = []string{"sound.wav"}
	ac.Loop = true
	ac.Pitch = 1.5
	s := node.NewSound()
	s.Direction = vec.SFVec3f{X: 1, Y: 0, Z: 0} // non-default
	s.Intensity = 0.8
	s.Source = ac
	out := writeScene([]node.Node{s})
	mustContain(t, out, "Sound", "direction 1 0 0", "intensity 0.8", "AudioClip", `"sound.wav"`, "loop TRUE", "pitch 1.5")
}

func TestWrite_Script_Programmatic(t *testing.T) {
	sc := &node.Script{URL: []string{"script.js"}, DirectOutput: true, MustEvaluate: true}
	out := writeScene([]node.Node{sc})
	mustContain(t, out, "Script", `"script.js"`, "directOutput TRUE", "mustEvaluate TRUE")
}

// ---------------------------------------------------------------------------
// IFS with colors and normals (full DataSet round-trip)
// ---------------------------------------------------------------------------

func TestRoundTrip_IFS_WithColors(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  geometry IndexedFaceSet {
    colorPerVertex FALSE
    coord Coordinate { point [0 0 0, 1 0 0, 0 1 0, 1 1 0] }
    coordIndex [0 1 2 -1, 1 3 2 -1]
    color Color { color [1 0 0, 0 1 0] }
    colorIndex [0, 1]
  }
}`
	out := roundTrip(t, vrml)
	mustContain(t, out, "colorPerVertex FALSE", "coordIndex", "colorIndex", "Color")
}

func TestRoundTrip_IFS_WithNormals(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  geometry IndexedFaceSet {
    coord Coordinate { point [0 0 0, 1 0 0, 0 1 0] }
    coordIndex [0 1 2 -1]
    normal Normal { vector [0 0 1] }
    texCoord TextureCoordinate { point [0 0, 1 0, 0 1] }
  }
}`
	out := roundTrip(t, vrml)
	mustContain(t, out, "Normal", "vector", "TextureCoordinate", "point")
}

// ---------------------------------------------------------------------------
// isNilNode edge case
// ---------------------------------------------------------------------------

func TestWrite_NilNode(t *testing.T) {
	out := writeScene([]node.Node{nil})
	// Should not panic, should produce just header
	if !strings.HasPrefix(out, "#VRML V2.0 utf8") {
		t.Fatal("bad output for nil node")
	}
}

// ---------------------------------------------------------------------------
// Unknown node type
// ---------------------------------------------------------------------------

type unknownNode struct{ node.BaseNode }

func TestWrite_UnknownNode(t *testing.T) {
	out := writeScene([]node.Node{&unknownNode{}})
	mustContain(t, out, "# unknown node type")
}
