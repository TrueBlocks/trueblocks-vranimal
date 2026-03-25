package parser

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
)

func TestInline_URLParsing(t *testing.T) {
	input := `#VRML V2.0 utf8
Inline { url "scene.wrl" }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	inl, ok := nodes[0].(*node.Inline)
	if !ok {
		t.Fatalf("expected *node.Inline, got %T", nodes[0])
	}
	if len(inl.URL) != 1 || inl.URL[0] != "scene.wrl" {
		t.Fatalf("expected URL [scene.wrl], got %v", inl.URL)
	}
}

func TestInline_MultipleURLs(t *testing.T) {
	input := `#VRML V2.0 utf8
Inline { url [ "a.wrl" "b.wrl" "c.wrl" ] }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	inl := nodes[0].(*node.Inline)
	if len(inl.URL) != 3 {
		t.Fatalf("expected 3 URLs, got %d", len(inl.URL))
	}
}

func TestInline_BboxFields(t *testing.T) {
	input := `#VRML V2.0 utf8
Inline {
  url "scene.wrl"
  bboxCenter 1 2 3
  bboxSize 4 5 6
}
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	inl := nodes[0].(*node.Inline)
	if inl.BboxCenter.X != 1 || inl.BboxCenter.Y != 2 || inl.BboxCenter.Z != 3 {
		t.Errorf("expected bboxCenter 1 2 3, got %v", inl.BboxCenter)
	}
	if inl.BboxSize.X != 4 || inl.BboxSize.Y != 5 || inl.BboxSize.Z != 6 {
		t.Errorf("expected bboxSize 4 5 6, got %v", inl.BboxSize)
	}
}

func TestInline_ResolvesExternalFile(t *testing.T) {
	extFile := `#VRML V2.0 utf8
Shape {
  geometry Box { size 2 2 2 }
}
`
	input := `#VRML V2.0 utf8
Inline { url "shapes.wrl" }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(func(url string) (io.ReadCloser, error) {
		if url == "shapes.wrl" {
			return io.NopCloser(strings.NewReader(extFile)), nil
		}
		return nil, fmt.Errorf("not found: %s", url)
	})
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	inl := nodes[0].(*node.Inline)
	if len(inl.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(inl.Children))
	}
	if _, ok := inl.Children[0].(*node.Shape); !ok {
		t.Fatalf("expected *node.Shape child, got %T", inl.Children[0])
	}
}

func TestInline_FallbackURL(t *testing.T) {
	extFile := `#VRML V2.0 utf8
PointLight { intensity 0.5 }
`
	input := `#VRML V2.0 utf8
Inline { url [ "missing.wrl" "backup.wrl" ] }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(func(url string) (io.ReadCloser, error) {
		if url == "backup.wrl" {
			return io.NopCloser(strings.NewReader(extFile)), nil
		}
		return nil, fmt.Errorf("not found: %s", url)
	})
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	inl := nodes[0].(*node.Inline)
	if len(inl.Children) != 1 {
		t.Fatalf("expected 1 child from fallback, got %d", len(inl.Children))
	}
	if _, ok := inl.Children[0].(*node.PointLight); !ok {
		t.Fatalf("expected *node.PointLight, got %T", inl.Children[0])
	}
}

func TestInline_AllURLsFail(t *testing.T) {
	input := `#VRML V2.0 utf8
Inline { url [ "a.wrl" "b.wrl" ] }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(func(url string) (io.ReadCloser, error) {
		return nil, fmt.Errorf("not found: %s", url)
	})
	nodes := p.Parse()
	// Should parse without errors — just no children resolved
	if len(p.Errors()) > 0 {
		t.Fatalf("unexpected errors: %v", p.Errors())
	}
	inl := nodes[0].(*node.Inline)
	if len(inl.Children) != 0 {
		t.Fatalf("expected 0 children, got %d", len(inl.Children))
	}
}

func TestInline_MultipleChildren(t *testing.T) {
	extFile := `#VRML V2.0 utf8
Shape { geometry Box { } }
Shape { geometry Sphere { } }
DirectionalLight { }
`
	input := `#VRML V2.0 utf8
Inline { url "multi.wrl" }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(func(url string) (io.ReadCloser, error) {
		if url == "multi.wrl" {
			return io.NopCloser(strings.NewReader(extFile)), nil
		}
		return nil, fmt.Errorf("not found: %s", url)
	})
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	inl := nodes[0].(*node.Inline)
	if len(inl.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(inl.Children))
	}
}

func TestInline_DEFPropagation(t *testing.T) {
	extFile := `#VRML V2.0 utf8
DEF MyBox Shape { geometry Box { } }
`
	input := `#VRML V2.0 utf8
Inline { url "defs.wrl" }
USE MyBox
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(func(url string) (io.ReadCloser, error) {
		if url == "defs.wrl" {
			return io.NopCloser(strings.NewReader(extFile)), nil
		}
		return nil, fmt.Errorf("not found: %s", url)
	})
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes (Inline + USE), got %d", len(nodes))
	}
	// Second node should be the same Shape via USE
	if _, ok := nodes[1].(*node.Shape); !ok {
		t.Fatalf("expected USE to return *node.Shape, got %T", nodes[1])
	}
}

func TestInline_ProtoPropagation(t *testing.T) {
	extFile := `#VRML V2.0 utf8
PROTO Lamp [
  field SFFloat intensity 0.5
] {
  PointLight { intensity IS intensity }
}
`
	input := `#VRML V2.0 utf8
Inline { url "protos.wrl" }
Lamp { intensity 0.9 }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(func(url string) (io.ReadCloser, error) {
		if url == "protos.wrl" {
			return io.NopCloser(strings.NewReader(extFile)), nil
		}
		return nil, fmt.Errorf("not found: %s", url)
	})
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	// Inline (empty body, just PROTO defs) + Lamp instance
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	// Second should be PointLight from Lamp PROTO
	if _, ok := nodes[1].(*node.PointLight); !ok {
		t.Fatalf("expected *node.PointLight from Lamp, got %T", nodes[1])
	}
}

func TestInline_LocalFileResolution(t *testing.T) {
	tmpDir := t.TempDir()

	extContent := `#VRML V2.0 utf8
Shape { geometry Sphere { } }
`
	err := os.WriteFile(tmpDir+"/sphere.wrl", []byte(extContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	input := `#VRML V2.0 utf8
Inline { url "sphere.wrl" }
`
	p := NewParser(strings.NewReader(input))
	p.SetBaseDir(tmpDir)
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	inl := nodes[0].(*node.Inline)
	if len(inl.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(inl.Children))
	}
	if _, ok := inl.Children[0].(*node.Shape); !ok {
		t.Fatalf("expected Shape, got %T", inl.Children[0])
	}
}

func TestInline_NoURL(t *testing.T) {
	input := `#VRML V2.0 utf8
Inline { }
`
	p := NewParser(strings.NewReader(input))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	inl := nodes[0].(*node.Inline)
	if len(inl.Children) != 0 {
		t.Fatalf("expected 0 children, got %d", len(inl.Children))
	}
}
