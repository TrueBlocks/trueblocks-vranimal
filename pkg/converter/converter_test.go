package converter

import (
	"strings"
	"testing"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/math32"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func parseAndConvert(t *testing.T, vrml string) *core.Node {
	t.Helper()
	p := parser.NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parse warning: %s", e)
		}
	}
	if len(nodes) == 0 {
		t.Fatal("no nodes parsed")
	}
	root := core.NewNode()
	Convert(nodes, root, "")
	return root
}

func childCount(n *core.Node) int {
	return len(n.Children())
}

func TestConvert_Box(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nShape { geometry Box { size 2 3 4 } }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestConvert_Transform(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nTransform { translation 1 2 3 children [ Shape { geometry Sphere { radius 1 } } ] }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestConvert_TransformCenter(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nTransform { center 1 0 0 translation 0 2 0 rotation 0 1 0 1.57 children [ Shape { geometry Box {} } ] }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
	// Transform with center creates outer + inner node
	transformNode := root.Children()[0]
	innerCount := len(transformNode.Children())
	if innerCount != 1 {
		t.Fatalf("expected 1 inner node for center, got %d", innerCount)
	}
}

func TestConvert_Lights(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nDirectionalLight { direction 0 -1 0 intensity 0.8 }\nPointLight { location 1 2 3 }\nSpotLight { location 0 5 0 direction 0 -1 0 cutOffAngle 0.5 }")
	if childCount(root) != 3 {
		t.Fatalf("expected 3 lights, got %d", childCount(root))
	}
}

func TestConvert_IndexedFaceSet(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nShape { geometry IndexedFaceSet { coord Coordinate { point [ 0 0 0, 1 0 0, 1 1 0, 0 1 0 ] } coordIndex [ 0 1 2 3 -1 ] } }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestConvert_IndexedLineSet(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nShape { geometry IndexedLineSet { coord Coordinate { point [ 0 0 0, 1 0 0, 1 1 0 ] } coordIndex [ 0 1 2 -1 ] } }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestConvert_PointSet(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nShape { geometry PointSet { coord Coordinate { point [ 0 0 0, 1 1 1, 2 2 2 ] } color Color { color [ 1 0 0, 0 1 0, 0 0 1 ] } } }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestConvert_InlineChildren(t *testing.T) {
	inl := &node.Inline{}
	inl.Children = []node.Node{
		&node.Shape{Geometry: &node.Box{Size: vec.SFVec3f{X: 1, Y: 1, Z: 1}}},
	}
	root := core.NewNode()
	Convert([]node.Node{inl}, root, "")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child from Inline, got %d", childCount(root))
	}
}

func TestConvert_Switch(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nSwitch { whichChoice 1 choice [ Shape { geometry Box {} } Shape { geometry Sphere {} } ] }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 container, got %d", childCount(root))
	}
	container := root.Children()[0].GetNode()
	// All choices rendered as wrapper nodes
	if len(container.Children()) != 2 {
		t.Fatalf("expected 2 wrapper children, got %d", len(container.Children()))
	}
	// whichChoice=1 → wrapper[0] hidden, wrapper[1] visible
	if container.Children()[0].GetNode().Visible() {
		t.Fatal("wrapper 0 should be hidden")
	}
	if !container.Children()[1].GetNode().Visible() {
		t.Fatal("wrapper 1 should be visible")
	}
}

func TestConvert_LOD(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nLOD { level [ Shape { geometry Box {} } Shape { geometry Sphere {} } ] }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 container, got %d", childCount(root))
	}
	container := root.Children()[0].GetNode()
	if len(container.Children()) != 2 {
		t.Fatalf("expected 2 wrapper children, got %d", len(container.Children()))
	}
	// Default active level 0 → wrapper[0] visible, wrapper[1] hidden
	if !container.Children()[0].GetNode().Visible() {
		t.Fatal("wrapper 0 should be visible")
	}
	if container.Children()[1].GetNode().Visible() {
		t.Fatal("wrapper 1 should be hidden")
	}
}

func TestConvert_ElevationGrid(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nShape { geometry ElevationGrid { xDimension 3 zDimension 3 xSpacing 1.0 zSpacing 1.0 height [ 0 1 0 1 2 1 0 1 0 ] } }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestGetViewpoint(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nViewpoint { position 1 2 3 fieldOfView 0.8 }"))
	nodes := p.Parse()
	vp := GetViewpoint(nodes)
	if vp == nil {
		t.Fatal("expected viewpoint")
		return
	}
	if vp.Position.X != 1 || vp.Position.Y != 2 || vp.Position.Z != 3 {
		t.Fatalf("wrong position: %v", vp.Position)
	}
}

func TestGetBackground(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nBackground { skyColor [ 0.2 0.3 0.8 ] }"))
	nodes := p.Parse()
	bg := GetBackground(nodes)
	if bg == nil {
		t.Fatal("expected background")
	}
}

func TestGetNavigationInfo(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nNavigationInfo { headlight FALSE speed 2.5 }"))
	nodes := p.Parse()
	ni := GetNavigationInfo(nodes)
	if ni == nil {
		t.Fatal("expected navigation info")
		return
	}
	if ni.Headlight {
		t.Fatal("headlight should be false")
	}
	if ni.Speed != 2.5 {
		t.Fatalf("wrong speed: %g", ni.Speed)
	}
}

func TestGetFog(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nFog { color 0.5 0.5 0.5 fogType \"LINEAR\" visibilityRange 100 }"))
	nodes := p.Parse()
	fg := GetFog(nodes)
	if fg == nil {
		t.Fatal("expected fog")
		return
	}
	if fg.FogType != "LINEAR" {
		t.Fatalf("wrong fog type: %s", fg.FogType)
	}
	if fg.VisibilityRange != 100 {
		t.Fatalf("wrong visibility range: %g", fg.VisibilityRange)
	}
}

func parseAndConvertNM(t *testing.T, vrml string) (*core.Node, *NodeMap) {
	t.Helper()
	p := parser.NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		for _, e := range errs {
			t.Logf("parse warning: %s", e)
		}
	}
	if len(nodes) == 0 {
		t.Fatal("no nodes parsed")
	}
	root := core.NewNode()
	nm := Convert(nodes, root, "")
	return root, nm
}

func TestSwitchDynamic(t *testing.T) {
	_, nm := parseAndConvertNM(t, "#VRML V2.0 utf8\nSwitch { whichChoice 0 choice [ Shape { geometry Box {} } Shape { geometry Sphere {} } Shape { geometry Cone {} } ] }")
	if len(nm.Switches) != 1 {
		t.Fatalf("expected 1 switch in NodeMap, got %d", len(nm.Switches))
	}
	for sw, wrappers := range nm.Switches {
		// Initially whichChoice=0
		if !wrappers[0].Visible() {
			t.Fatal("choice 0 should be visible initially")
		}
		if wrappers[1].Visible() || wrappers[2].Visible() {
			t.Fatal("choices 1,2 should be hidden initially")
		}
		// Change to choice 2
		sw.WhichChoice = 2
		nm.UpdateDynamic()
		if wrappers[0].Visible() || wrappers[1].Visible() {
			t.Fatal("choices 0,1 should be hidden after switch to 2")
		}
		if !wrappers[2].Visible() {
			t.Fatal("choice 2 should be visible after switch")
		}
		// -1 means none visible
		sw.WhichChoice = -1
		nm.UpdateDynamic()
		for i, w := range wrappers {
			if w.Visible() {
				t.Fatalf("choice %d should be hidden when whichChoice=-1", i)
			}
		}
	}
}

func TestLODDynamic(t *testing.T) {
	_, nm := parseAndConvertNM(t, "#VRML V2.0 utf8\nLOD { range [ 50 ] level [ Shape { geometry Box {} } Shape { geometry Sphere {} } ] }")
	if len(nm.LODs) != 1 {
		t.Fatalf("expected 1 LOD in NodeMap, got %d", len(nm.LODs))
	}
	for lod, wrappers := range nm.LODs {
		// Initially active=0 (default)
		if !wrappers[0].Visible() {
			t.Fatal("level 0 should be visible initially")
		}
		if wrappers[1].Visible() {
			t.Fatal("level 1 should be hidden initially")
		}
		// Switch to level 1
		lod.ActiveLevel = 1
		nm.UpdateDynamic()
		if wrappers[0].Visible() {
			t.Fatal("level 0 should be hidden after switch")
		}
		if !wrappers[1].Visible() {
			t.Fatal("level 1 should be visible after switch")
		}
	}
}

func TestConvert_Billboard(t *testing.T) {
	_, nm := parseAndConvertNM(t, "#VRML V2.0 utf8\nBillboard { axisOfRotation 0 1 0 children [ Shape { geometry Box {} } ] }")
	if len(nm.Billboards) != 1 {
		t.Fatalf("expected 1 billboard in NodeMap, got %d", len(nm.Billboards))
	}
}

func TestBillboardRotation(t *testing.T) {
	_, nm := parseAndConvertNM(t, "#VRML V2.0 utf8\nBillboard { axisOfRotation 0 1 0 children [ Shape { geometry Box {} } ] }")
	// Place camera along +X, billboard should rotate to face it
	nm.CameraPos = math32.Vector3{X: 10, Y: 0, Z: 0}
	nm.UpdateDynamic()
	for _, g3nNode := range nm.Billboards {
		rot := g3nNode.Rotation()
		// Y rotation should be approximately π/2 (facing +X)
		if rot.Y < 1.0 || rot.Y > 2.0 {
			t.Fatalf("expected Y rotation ~π/2 from +X camera, got %v", rot)
		}
	}
}

func TestConvert_Anchor(t *testing.T) {
	_, nm := parseAndConvertNM(t, "#VRML V2.0 utf8\nAnchor { url \"http://example.com\" children [ Shape { geometry Sphere {} } ] }")
	if len(nm.Anchors) != 1 {
		t.Fatalf("expected 1 anchor in NodeMap, got %d", len(nm.Anchors))
	}
	for a := range nm.Anchors {
		if len(a.URL) == 0 || a.URL[0] != "http://example.com" {
			t.Fatalf("expected anchor URL 'http://example.com', got %v", a.URL)
		}
	}
}

func TestConvert_Collision(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nCollision { children [ Shape { geometry Box {} } Shape { geometry Sphere {} } ] }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 container, got %d", childCount(root))
	}
	// Collision renders children as a group
	container := root.Children()[0].GetNode()
	if len(container.Children()) != 2 {
		t.Fatalf("expected 2 children in Collision group, got %d", len(container.Children()))
	}
}

func TestGetNavigationInfo_Defaults(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nNavigationInfo {}"))
	nodes := p.Parse()
	ni := GetNavigationInfo(nodes)
	if ni == nil {
		t.Fatal("expected navigation info")
		return
	}
	if !ni.Headlight {
		t.Fatal("default headlight should be true")
	}
	if ni.Speed != 1.0 {
		t.Fatalf("default speed should be 1.0, got %g", ni.Speed)
	}
	if len(ni.Type) == 0 || ni.Type[0] != "WALK" {
		t.Fatalf("default type should be WALK, got %v", ni.Type)
	}
}

func TestGetNavigationInfo_Custom(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nNavigationInfo { headlight FALSE speed 5.0 type \"EXAMINE\" visibilityLimit 100 }"))
	nodes := p.Parse()
	ni := GetNavigationInfo(nodes)
	if ni == nil {
		t.Fatal("expected navigation info")
		return
	}
	if ni.Headlight {
		t.Fatal("headlight should be false")
	}
	if ni.Speed != 5.0 {
		t.Fatalf("speed should be 5.0, got %g", ni.Speed)
	}
	if len(ni.Type) == 0 || ni.Type[0] != "EXAMINE" {
		t.Fatalf("type should be EXAMINE, got %v", ni.Type)
	}
	if ni.VisibilityLimit != 100 {
		t.Fatalf("visibilityLimit should be 100, got %g", ni.VisibilityLimit)
	}
}

// ===========================================================================
// Gap-filling tests (issue #41)
// ===========================================================================

// ---------------------------------------------------------------------------
// Extrusion conversion
// ---------------------------------------------------------------------------

func TestConvert_Extrusion(t *testing.T) {
	root := parseAndConvert(t, `#VRML V2.0 utf8
Shape {
  geometry Extrusion {
    crossSection [1 1, 1 -1, -1 -1, -1 1, 1 1]
    spine [0 0 0, 0 1 0, 0 2 0]
    scale [1 1, 0.5 0.5, 0.25 0.25]
  }
}`)
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestConvert_ExtrusionMinimal(t *testing.T) {
	// Edge case: only 3 cross-section + 2 spine (minimum)
	root := parseAndConvert(t, `#VRML V2.0 utf8
Shape {
  geometry Extrusion {
    crossSection [1 0, 0 1, -1 0]
    spine [0 0 0, 0 1 0]
  }
}`)
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestConvert_ExtrusionDegenerate(t *testing.T) {
	// Degenerate: too few cross-section or spine → should not crash
	root := core.NewNode()
	ex := node.NewExtrusion()
	ex.CrossSection = []vec.SFVec2f{{X: 1}}  // too few
	ex.Spine = []vec.SFVec3f{{Y: 0}, {Y: 1}} // ok
	Convert([]node.Node{&node.Shape{Geometry: ex}}, root, "")
	// Should not panic; may produce 0 or 1 child depending on guard
}

// ---------------------------------------------------------------------------
// IFS with auto-computed normals (no normals provided)
// ---------------------------------------------------------------------------

func TestConvert_IFS_AutoNormals(t *testing.T) {
	root := parseAndConvert(t, `#VRML V2.0 utf8
Shape {
  geometry IndexedFaceSet {
    coord Coordinate { point [0 0 0, 1 0 0, 0 1 0, 1 1 0] }
    coordIndex [0 1 3 2 -1]
  }
}`)
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

// ---------------------------------------------------------------------------
// IFS with colors and normals
// ---------------------------------------------------------------------------

func TestConvert_IFS_WithColors(t *testing.T) {
	root := parseAndConvert(t, `#VRML V2.0 utf8
Shape {
  geometry IndexedFaceSet {
    coord Coordinate { point [0 0 0, 1 0 0, 0 1 0] }
    coordIndex [0 1 2 -1]
    color Color { color [1 0 0, 0 1 0, 0 0 1] }
  }
}`)
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestConvert_IFS_WithNormalsAndTexCoords(t *testing.T) {
	root := parseAndConvert(t, `#VRML V2.0 utf8
Shape {
  geometry IndexedFaceSet {
    coord Coordinate { point [0 0 0, 1 0 0, 0 1 0] }
    coordIndex [0 1 2 -1]
    normal Normal { vector [0 0 1, 0 0 1, 0 0 1] }
    texCoord TextureCoordinate { point [0 0, 1 0, 0 1] }
  }
}`)
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

// ---------------------------------------------------------------------------
// Material/Appearance tests
// ---------------------------------------------------------------------------

func TestConvert_MaterialTransparency(t *testing.T) {
	_, nm := parseAndConvertNM(t, `#VRML V2.0 utf8
Shape {
  appearance Appearance { material Material { diffuseColor 1 0 0 transparency 0.5 } }
  geometry Sphere {}
}`)
	if len(nm.Materials) != 1 {
		t.Fatalf("expected 1 material in NodeMap, got %d", len(nm.Materials))
	}
}

func TestConvert_MaterialEmissive(t *testing.T) {
	root := parseAndConvert(t, `#VRML V2.0 utf8
Shape {
  appearance Appearance { material Material { emissiveColor 0 1 0 shininess 0.8 } }
  geometry Box {}
}`)
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

func TestConvert_MaterialDynamic(t *testing.T) {
	_, nm := parseAndConvertNM(t, `#VRML V2.0 utf8
Shape {
  appearance Appearance { material Material { diffuseColor 1 0 0 } }
  geometry Sphere {}
}`)
	// Change material color and verify UpdateDynamic syncs
	for mat := range nm.Materials {
		mat.DiffuseColor = vec.SFColor{R: 0, G: 0, B: 1, A: 1}
	}
	nm.UpdateDynamic()
	for _, g3nMat := range nm.Materials {
		c := g3nMat.AmbientColor()
		if c.R > 0.5 {
			t.Fatal("material should have updated to blue")
		}
	}
}

// ---------------------------------------------------------------------------
// Get* searches nested in groups/transforms
// ---------------------------------------------------------------------------

func TestGetViewpoint_Nested(t *testing.T) {
	p := parser.NewParser(strings.NewReader(`#VRML V2.0 utf8
Transform { children [
  Group { children [
    Viewpoint { position 5 5 5 }
  ] }
] }`))
	nodes := p.Parse()
	vp := GetViewpoint(nodes)
	if vp == nil {
		t.Fatal("should find Viewpoint nested in Transform>Group")
		return
	}
	if vp.Position.X != 5 {
		t.Fatalf("wrong position: %v", vp.Position)
	}
}

func TestGetBackground_Nested(t *testing.T) {
	p := parser.NewParser(strings.NewReader(`#VRML V2.0 utf8
Group { children [ Background { skyColor [0 0 1] } ] }`))
	nodes := p.Parse()
	bg := GetBackground(nodes)
	if bg == nil {
		t.Fatal("should find Background nested in Group")
		return
	}
	if len(bg.SkyColor) != 1 {
		t.Fatalf("expected 1 skyColor, got %d", len(bg.SkyColor))
	}
}

func TestGetFog_Nested(t *testing.T) {
	p := parser.NewParser(strings.NewReader(`#VRML V2.0 utf8
Transform { children [ Fog { visibilityRange 200 } ] }`))
	nodes := p.Parse()
	fg := GetFog(nodes)
	if fg == nil {
		t.Fatal("should find Fog nested in Transform")
		return
	}
	if fg.VisibilityRange != 200 {
		t.Fatalf("wrong range: %g", fg.VisibilityRange)
	}
}

func TestGetViewpoint_NotFound(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nShape { geometry Box {} }"))
	nodes := p.Parse()
	if GetViewpoint(nodes) != nil {
		t.Fatal("should return nil when no Viewpoint")
	}
}

func TestGetBackground_NotFound(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nShape { geometry Box {} }"))
	nodes := p.Parse()
	if GetBackground(nodes) != nil {
		t.Fatal("should return nil when no Background")
	}
}

func TestGetFog_NotFound(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nShape { geometry Box {} }"))
	nodes := p.Parse()
	if GetFog(nodes) != nil {
		t.Fatal("should return nil when no Fog")
	}
}

func TestGetNavigationInfo_NotFound(t *testing.T) {
	p := parser.NewParser(strings.NewReader("#VRML V2.0 utf8\nShape { geometry Box {} }"))
	nodes := p.Parse()
	if GetNavigationInfo(nodes) != nil {
		t.Fatal("should return nil when no NavigationInfo")
	}
}

// ---------------------------------------------------------------------------
// Switch/LOD edge cases
// ---------------------------------------------------------------------------

func TestConvert_Switch_NegativeChoice(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nSwitch { whichChoice -1 choice [ Shape { geometry Box {} } ] }")
	container := root.Children()[0].GetNode()
	// -1 means no choice visible
	for _, c := range container.Children() {
		if c.GetNode().Visible() {
			t.Fatal("all choices should be hidden when whichChoice=-1")
		}
	}
}

func TestConvert_Switch_OutOfBounds(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nSwitch { whichChoice 99 choice [ Shape { geometry Box {} } ] }")
	container := root.Children()[0].GetNode()
	// Out of bounds → none visible
	for _, c := range container.Children() {
		if c.GetNode().Visible() {
			t.Fatal("all choices should be hidden when whichChoice out of bounds")
		}
	}
}

func TestConvert_LOD_EmptyLevels(t *testing.T) {
	// LOD with no levels → should not crash
	root := core.NewNode()
	lod := &node.LOD{}
	Convert([]node.Node{lod}, root, "")
	// Just verify no panic
}

// ---------------------------------------------------------------------------
// Transform dynamic updates
// ---------------------------------------------------------------------------

func TestTransformDynamic(t *testing.T) {
	_, nm := parseAndConvertNM(t, `#VRML V2.0 utf8
Transform { translation 1 2 3 children [ Shape { geometry Box {} } ] }`)
	if len(nm.Transforms) != 1 {
		t.Fatalf("expected 1 transform in NodeMap, got %d", len(nm.Transforms))
	}
	for tr := range nm.Transforms {
		tr.Translation = vec.SFVec3f{X: 10, Y: 20, Z: 30}
	}
	nm.UpdateDynamic()
	for _, g3nNode := range nm.Transforms {
		pos := g3nNode.Position()
		if pos.X != 10 || pos.Y != 20 || pos.Z != 30 {
			t.Fatalf("transform position not updated: %v", pos)
		}
	}
}

// ---------------------------------------------------------------------------
// Empty / nil scenes
// ---------------------------------------------------------------------------

func TestConvert_EmptyScene(t *testing.T) {
	root := core.NewNode()
	nm := Convert(nil, root, "")
	if nm == nil {
		t.Fatal("should return non-nil NodeMap for empty scene")
	}
	if childCount(root) != 0 {
		t.Fatalf("expected 0 children, got %d", childCount(root))
	}
}

func TestConvert_NilNodes(t *testing.T) {
	root := core.NewNode()
	nm := Convert([]node.Node{nil}, root, "")
	if nm == nil {
		t.Fatal("should return non-nil NodeMap")
	}
}

// ---------------------------------------------------------------------------
// IndexedLineSet with colors
// ---------------------------------------------------------------------------

func TestConvert_IndexedLineSet_WithColors(t *testing.T) {
	root := parseAndConvert(t, `#VRML V2.0 utf8
Shape {
  geometry IndexedLineSet {
    coord Coordinate { point [0 0 0, 1 0 0, 1 1 0, 0 1 0] }
    coordIndex [0 1 2 3 -1]
    color Color { color [1 0 0, 0 1 0, 0 0 1, 1 1 0] }
  }
}`)
	if childCount(root) != 1 {
		t.Fatalf("expected 1 child, got %d", childCount(root))
	}
}

// ---------------------------------------------------------------------------
// Multiple shapes in one scene
// ---------------------------------------------------------------------------

func TestConvert_MultipleShapes(t *testing.T) {
	root := parseAndConvert(t, `#VRML V2.0 utf8
Shape { geometry Box {} }
Shape { geometry Sphere { radius 2 } }
Shape { geometry Cone { bottomRadius 1 height 3 } }
Shape { geometry Cylinder { radius 0.5 height 2 } }`)
	if childCount(root) != 4 {
		t.Fatalf("expected 4 children, got %d", childCount(root))
	}
}

// ---------------------------------------------------------------------------
// Group nested structure
// ---------------------------------------------------------------------------

func TestConvert_NestedGroups(t *testing.T) {
	root := parseAndConvert(t, `#VRML V2.0 utf8
Group {
  children [
    Group { children [ Shape { geometry Box {} } ] }
    Group { children [ Shape { geometry Sphere {} } ] }
  ]
}`)
	if childCount(root) != 1 {
		t.Fatalf("expected 1 top-level group, got %d", childCount(root))
	}
	outerGroup := root.Children()[0].GetNode()
	if len(outerGroup.Children()) != 2 {
		t.Fatalf("expected 2 inner groups, got %d", len(outerGroup.Children()))
	}
}
