package parser

import (
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
)

// ---------------------------------------------------------------------------
// Empty / minimal input
// ---------------------------------------------------------------------------

func TestParse_EmptyInput(t *testing.T) {
	p := NewParser(strings.NewReader(""))
	nodes := p.Parse()
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(nodes))
	}
	if len(p.Errors()) != 0 {
		t.Fatalf("expected no errors, got %v", p.Errors())
	}
}

func TestParse_OnlyWhitespace(t *testing.T) {
	p := NewParser(strings.NewReader("   \n\t\n  "))
	nodes := p.Parse()
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(nodes))
	}
}

func TestParse_CommentOnly(t *testing.T) {
	vrml := "#VRML V2.0 utf8\n# This file has no nodes\n"
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes from comment-only file, got %d", len(nodes))
	}
}

func TestParse_HeaderOnly(t *testing.T) {
	p := NewParser(strings.NewReader("#VRML V2.0 utf8\n"))
	nodes := p.Parse()
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(nodes))
	}
}

// ---------------------------------------------------------------------------
// DEF/USE edge cases
// ---------------------------------------------------------------------------

func TestParse_DEFWithoutName(t *testing.T) {
	vrml := `#VRML V2.0 utf8
DEF { }`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(p.Errors()) == 0 {
		t.Fatal("expected error for DEF without name")
	}
	_ = nodes
}

func TestParse_USEUndefined(t *testing.T) {
	vrml := `#VRML V2.0 utf8
USE NoSuchNode`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(p.Errors()) == 0 {
		t.Fatal("expected error for USE of undefined name")
	}
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(nodes))
	}
}

func TestParse_DEFThenUSE(t *testing.T) {
	vrml := `#VRML V2.0 utf8
DEF MyBox Box { size 2 2 2 }
USE MyBox`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	// Both nodes should be the same object
	if nodes[0] != nodes[1] {
		t.Fatal("USE should return same node reference")
	}
}

func TestParse_DEFShadowing(t *testing.T) {
	vrml := `#VRML V2.0 utf8
DEF MyShape Box { size 1 1 1 }
DEF MyShape Sphere { radius 5 }
USE MyShape`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	// After shadowing, USE should reference the latest DEF
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	_, isSphere := nodes[2].(*node.Sphere)
	if !isSphere {
		t.Fatalf("USE should resolve to Sphere after shadowing, got %T", nodes[2])
	}
}

// ---------------------------------------------------------------------------
// Unknown node types
// ---------------------------------------------------------------------------

func TestParse_UnknownNodeType(t *testing.T) {
	vrml := `#VRML V2.0 utf8
FooBar { baz 42 }
Box { size 3 3 3 }`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	// Unknown node should be silently skipped, Box should parse
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node (Box), got %d", len(nodes))
	}
	box, ok := nodes[0].(*node.Box)
	if !ok {
		t.Fatalf("expected Box, got %T", nodes[0])
	}
	if box.Size.X != 3 {
		t.Fatalf("box size X=%g, want 3", box.Size.X)
	}
}

func TestParse_UnknownNodeNested(t *testing.T) {
	vrml := `#VRML V2.0 utf8
UnknownOuter { inner UnknownInner { deep { } } val 1 }
Sphere { radius 2 }`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node (Sphere), got %d", len(nodes))
	}
}

// ---------------------------------------------------------------------------
// Unclosed braces / malformed syntax
// ---------------------------------------------------------------------------

func TestParse_UnclosedBrace(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Transform { children [`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	// Should not panic; may return partial result or empty
	_ = nodes
}

func TestParse_ExtraCloseBrace(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Box { size 1 1 1 } }`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	// Should parse the Box, extra brace consumed by default case
	if len(nodes) < 1 {
		t.Fatal("expected at least 1 node")
	}
}

// ---------------------------------------------------------------------------
// Multiple root nodes
// ---------------------------------------------------------------------------

func TestParse_MultipleRootNodes(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Box { size 1 1 1 }
Sphere { radius 2 }
Cylinder { radius 1 height 3 }`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(nodes) != 3 {
		t.Fatalf("expected 3 root nodes, got %d", len(nodes))
	}
}

// ---------------------------------------------------------------------------
// ROUTE edge cases
// ---------------------------------------------------------------------------

func TestParse_ROUTEMissingNodes(t *testing.T) {
	vrml := `#VRML V2.0 utf8
ROUTE NoNode.field TO AlsoMissing.field`
	p := NewParser(strings.NewReader(vrml))
	_ = p.Parse()
	routes := p.GetRoutes()
	if len(routes) != 0 {
		t.Fatalf("expected 0 routes for undefined nodes, got %d", len(routes))
	}
}

func TestParse_ROUTEPartiallyDefined(t *testing.T) {
	vrml := `#VRML V2.0 utf8
DEF T TimeSensor { loop TRUE }
ROUTE T.fraction_changed TO Missing.set_fraction`
	p := NewParser(strings.NewReader(vrml))
	_ = p.Parse()
	routes := p.GetRoutes()
	if len(routes) != 0 {
		t.Fatalf("expected 0 routes when dst undefined, got %d", len(routes))
	}
}

// ---------------------------------------------------------------------------
// Nested transforms
// ---------------------------------------------------------------------------

func TestParse_NestedTransforms(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Transform {
  translation 1 0 0
  children [
    Transform {
      translation 0 2 0
      children [
        Box { size 1 1 1 }
      ]
    }
  ]
}`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(nodes))
	}
	outer, ok := nodes[0].(*node.Transform)
	if !ok {
		t.Fatalf("expected Transform, got %T", nodes[0])
	}
	if outer.Translation.X != 1 {
		t.Fatalf("outer translation X=%g, want 1", outer.Translation.X)
	}
	if len(outer.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(outer.Children))
	}
	inner, ok := outer.Children[0].(*node.Transform)
	if !ok {
		t.Fatalf("expected inner Transform, got %T", outer.Children[0])
	}
	if inner.Translation.Y != 2 {
		t.Fatalf("inner translation Y=%g, want 2", inner.Translation.Y)
	}
	if len(inner.Children) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(inner.Children))
	}
}

// ---------------------------------------------------------------------------
// Complex scene round-trip
// ---------------------------------------------------------------------------

func TestParse_ComplexScene(t *testing.T) {
	vrml := `#VRML V2.0 utf8
DEF GROUND Transform {
  children [
    Shape {
      appearance Appearance { material Material { diffuseColor 0.2 0.8 0.2 } }
      geometry Box { size 10 0.1 10 }
    }
  ]
}
DEF BALL Transform {
  translation 0 2 0
  children [
    Shape {
      appearance Appearance { material Material { diffuseColor 1 0 0 } }
      geometry Sphere { radius 0.5 }
    }
  ]
}
DEF TIMER TimeSensor { cycleInterval 2 loop TRUE }
DEF MOVER PositionInterpolator { key [ 0 0.5 1 ] keyValue [ 0 2 0, 0 4 0, 0 2 0 ] }
ROUTE TIMER.fraction_changed TO MOVER.set_fraction
ROUTE MOVER.value_changed TO BALL.set_translation`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	routes := p.GetRoutes()
	if len(p.Errors()) != 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 4 {
		t.Fatalf("expected 4 nodes (2 Transform + TimeSensor + PI), got %d", len(nodes))
	}
	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}
}
