package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
)

func typeName(n node.Node) string {
	return fmt.Sprintf("%T", n)
}

func TestSubstituteIS_Basic(t *testing.T) {
	body := "scale IS size"
	vals := map[string]string{"size": "2 3 4"}
	got := substituteIS(body, vals)
	if got != "scale 2 3 4" {
		t.Errorf("got %q, want %q", got, "scale 2 3 4")
	}
}

func TestSubstituteIS_Multiple(t *testing.T) {
	body := `Transform {
	scale IS size
	children [
		Shape {
			appearance Appearance {
				material Material {
					diffuseColor IS color
				}
			}
		}
	]
}`
	vals := map[string]string{"size": "2 3 4", "color": "1 0 0"}
	got := substituteIS(body, vals)
	if strings.Contains(got, "IS") {
		t.Errorf("result still contains IS: %s", got)
	}
	if !strings.Contains(got, "scale 2 3 4") {
		t.Errorf("missing scale substitution: %s", got)
	}
	if !strings.Contains(got, "diffuseColor 1 0 0") {
		t.Errorf("missing color substitution: %s", got)
	}
}

func TestSubstituteIS_InString(t *testing.T) {
	body := `description "this IS cool" translation IS pos`
	vals := map[string]string{"pos": "1 2 3"}
	got := substituteIS(body, vals)
	if !strings.Contains(got, `"this IS cool"`) {
		t.Errorf("string was corrupted: %s", got)
	}
	if !strings.Contains(got, "translation 1 2 3") {
		t.Errorf("substitution failed: %s", got)
	}
}

func TestSubstituteIS_InComment(t *testing.T) {
	body := "scale IS size # IS comment\ntranslation IS pos"
	vals := map[string]string{"size": "2 3 4", "pos": "1 2 3"}
	got := substituteIS(body, vals)
	if !strings.Contains(got, "scale 2 3 4") {
		t.Errorf("scale substitution failed: %s", got)
	}
	if !strings.Contains(got, "translation 1 2 3") {
		t.Errorf("translation substitution failed: %s", got)
	}
}

func TestSubstituteIS_UnknownField(t *testing.T) {
	body := "diffuseColor IS unknownField"
	vals := map[string]string{"color": "1 0 0"}
	got := substituteIS(body, vals)
	if !strings.Contains(got, "IS") {
		t.Errorf("unknown IS was removed: %s", got)
	}
}

func TestParsePROTO_Declaration(t *testing.T) {
	input := `PROTO ColoredBox [
	field SFVec3f size 1 1 1
	exposedField SFColor color 0.8 0.2 0.2
	eventIn SFBool toggle
	eventOut SFTime clickTime
] {
	Shape {
		geometry Box { }
	}
}`
	p := NewParser(strings.NewReader(input))
	p.Parse()

	def, ok := p.ProtoTable()["ColoredBox"]
	if !ok {
		t.Fatal("PROTO ColoredBox not found in table")
	}
	if def.Name != "ColoredBox" {
		t.Errorf("name = %q, want ColoredBox", def.Name)
	}
	if len(def.Fields) != 4 {
		t.Fatalf("fields = %d, want 4", len(def.Fields))
	}
	if def.Fields[0].Kind != KindField {
		t.Errorf("field[0] kind = %d, want KindField", def.Fields[0].Kind)
	}
	if def.Fields[1].Kind != KindExposedField {
		t.Errorf("field[1] kind = %d, want KindExposedField", def.Fields[1].Kind)
	}
	if def.Fields[2].Kind != KindEventIn {
		t.Errorf("field[2] kind = %d, want KindEventIn", def.Fields[2].Kind)
	}
	if def.Fields[3].Kind != KindEventOut {
		t.Errorf("field[3] kind = %d, want KindEventOut", def.Fields[3].Kind)
	}
	if def.Fields[0].TypeName != "SFVec3f" || def.Fields[0].Name != "size" {
		t.Errorf("field[0] = %s %s, want SFVec3f size", def.Fields[0].TypeName, def.Fields[0].Name)
	}
	if def.Fields[2].Default != "" {
		t.Errorf("eventIn should have no default, got %q", def.Fields[2].Default)
	}
	if !strings.Contains(def.Body, "Shape") {
		t.Errorf("body missing Shape: %q", def.Body)
	}
}

func TestParsePROTO_SimpleInstance(t *testing.T) {
	input := `PROTO RedBox [
	field SFVec3f size 1 1 1
] {
	Shape {
		geometry Box {
			size IS size
		}
	}
}
RedBox { size 2 3 4 }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	if _, ok := nodes[0].(*node.Shape); !ok {
		t.Errorf("got %s, want *node.Shape", typeName(nodes[0]))
	}
}

func TestParsePROTO_DefaultValues(t *testing.T) {
	input := `PROTO MyBox [
	field SFVec3f size 5 5 5
] {
	Shape {
		geometry Box {
			size IS size
		}
	}
}
MyBox { }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	if _, ok := nodes[0].(*node.Shape); !ok {
		t.Errorf("got %s, want *node.Shape", typeName(nodes[0]))
	}
}

func TestParsePROTO_MultipleInstances(t *testing.T) {
	input := `PROTO Marker [
	field SFVec3f pos 0 0 0
] {
	Transform {
		translation IS pos
		children [
			Shape {
				geometry Sphere { radius 0.1 }
			}
		]
	}
}
Marker { pos 1 0 0 }
Marker { pos 0 1 0 }
Marker { pos 0 0 1 }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(nodes) != 3 {
		t.Fatalf("got %d nodes, want 3", len(nodes))
	}
	for i, n := range nodes {
		if _, ok := n.(*node.Transform); !ok {
			t.Errorf("node[%d] = %s, want *node.Transform", i, typeName(n))
		}
	}
}

func TestParsePROTO_MultiBodyNodes(t *testing.T) {
	input := `PROTO TwoLights [
	field SFColor c1 1 1 1
	field SFColor c2 0.5 0.5 0.5
] {
	DirectionalLight { color IS c1 }
	DirectionalLight { color IS c2 }
}
TwoLights { c1 1 0 0 c2 0 1 0 }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1 (Group wrapper)", len(nodes))
	}
	if _, ok := nodes[0].(*node.Group); !ok {
		t.Errorf("got %s, want *node.Group", typeName(nodes[0]))
	}
}

func TestParsePROTO_NestedProto(t *testing.T) {
	input := `PROTO Inner [
	field SFFloat r 1
] {
	Shape {
		geometry Sphere { radius IS r }
	}
}
PROTO Outer [
	field SFVec3f pos 0 0 0
] {
	Transform {
		translation IS pos
		children [
			Inner { r 0.5 }
		]
	}
}
Outer { pos 1 2 3 }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	if _, ok := nodes[0].(*node.Transform); !ok {
		t.Errorf("got %s, want *node.Transform", typeName(nodes[0]))
	}
}

func TestParsePROTO_BoolAndString(t *testing.T) {
	input := `PROTO MySwitch [
	field SFBool lit TRUE
	field SFString label "default"
] {
	WorldInfo { title IS label }
}
MySwitch { label "hello" }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	if _, ok := nodes[0].(*node.WorldInfo); !ok {
		t.Errorf("got %s, want *node.WorldInfo", typeName(nodes[0]))
	}
}

func TestParsePROTO_DEFInsideBody(t *testing.T) {
	input := `PROTO Named [
	field SFVec3f pos 0 0 0
] {
	DEF MyTransform Transform {
		translation IS pos
	}
}
Named { pos 5 5 5 }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	if _, ok := nodes[0].(*node.Transform); !ok {
		t.Errorf("got %s, want *node.Transform", typeName(nodes[0]))
	}
}

func TestParseEXTERNPROTO_Declaration(t *testing.T) {
	input := `EXTERNPROTO Brick [
	field SFVec3f brickSize
	exposedField SFColor brickColor
] [ "brick.wrl" "backup.wrl" ]
`
	p := NewParser(strings.NewReader(input))
	p.Parse()
	def, ok := p.ProtoTable()["Brick"]
	if !ok {
		t.Fatal("EXTERNPROTO Brick not found in table")
	}
	if len(def.Fields) != 2 {
		t.Fatalf("fields = %d, want 2", len(def.Fields))
	}
	if def.Body != "" {
		t.Errorf("EXTERNPROTO body should be empty, got %q", def.Body)
	}
}

func TestParsePROTO_MFStringDefault(t *testing.T) {
	input := `PROTO MyInline [
	field MFString url [ "a.wrl" "b.wrl" ]
] {
	WorldInfo { }
}
MyInline { }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	def := p.ProtoTable()["MyInline"]
	if def == nil {
		t.Fatal("PROTO MyInline not found")
		return
	}
	if !strings.Contains(def.Fields[0].Default, "a.wrl") {
		t.Errorf("MFString default = %q, expected a.wrl", def.Fields[0].Default)
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
}

func TestParsePROTO_EmptyBody(t *testing.T) {
	input := `EXTERNPROTO Widget [
	field SFFloat size
] "widget.wrl"
Widget { size 2 }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes from EXTERNPROTO instance, got %d", len(nodes))
	}
}

func TestParsePROTO_FieldValueVerification(t *testing.T) {
	input := `PROTO SizedBox [
	field SFVec3f size 1 1 1
] {
	Shape {
		geometry Box {
			size IS size
		}
	}
}
SizedBox { size 5 10 15 }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(nodes))
	}
	shape, ok := nodes[0].(*node.Shape)
	if !ok {
		t.Fatalf("got %s, want *node.Shape", typeName(nodes[0]))
	}
	box, ok := shape.Geometry.(*node.Box)
	if !ok {
		t.Fatalf("geometry = %T, want *node.Box", shape.Geometry)
	}
	if box.Size.X != 5 || box.Size.Y != 10 || box.Size.Z != 15 {
		t.Errorf("box size = %v, want {5 10 15}", box.Size)
	}
}
