package converter

import (
	"strings"
	"testing"

	"github.com/g3n/engine/core"

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
		t.Fatalf("expected 1 active choice, got %d", childCount(root))
	}
}

func TestConvert_LOD(t *testing.T) {
	root := parseAndConvert(t, "#VRML V2.0 utf8\nLOD { level [ Shape { geometry Box {} } Shape { geometry Sphere {} } ] }")
	if childCount(root) != 1 {
		t.Fatalf("expected 1 LOD level, got %d", childCount(root))
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
