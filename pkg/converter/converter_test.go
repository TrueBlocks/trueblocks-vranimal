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
