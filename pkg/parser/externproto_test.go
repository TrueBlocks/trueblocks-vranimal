package parser

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
)

// mockFetcher returns a URL fetcher that serves content from a map.
func mockFetcher(files map[string]string) func(string) (io.ReadCloser, error) {
	return func(url string) (io.ReadCloser, error) {
		if content, ok := files[url]; ok {
			return io.NopCloser(strings.NewReader(content)), nil
		}
		return nil, fmt.Errorf("not found: %s", url)
	}
}

func TestParseURLList_Single(t *testing.T) {
	input := `EXTERNPROTO Foo [ ] "foo.wrl"`
	p := NewParser(strings.NewReader(input))
	p.lex.Next() // consume EXTERNPROTO
	p.lex.Next() // consume Foo
	p.lex.Next() // consume [
	p.lex.Next() // consume ]
	urls := p.parseURLList()
	if len(urls) != 1 || urls[0] != "foo.wrl" {
		t.Fatalf("expected [foo.wrl], got %v", urls)
	}
}

func TestParseURLList_Multiple(t *testing.T) {
	input := `EXTERNPROTO Foo [ ] [ "a.wrl" "b.wrl#Bar" "c.wrl" ]`
	p := NewParser(strings.NewReader(input))
	p.lex.Next() // EXTERNPROTO
	p.lex.Next() // Foo
	p.lex.Next() // [
	p.lex.Next() // ]
	urls := p.parseURLList()
	if len(urls) != 3 {
		t.Fatalf("expected 3 URLs, got %d: %v", len(urls), urls)
	}
	if urls[0] != "a.wrl" || urls[1] != "b.wrl#Bar" || urls[2] != "c.wrl" {
		t.Fatalf("unexpected URLs: %v", urls)
	}
}

func TestExternProto_ResolveSimple(t *testing.T) {
	// External file has a PROTO named "Lamp"
	extFile := `#VRML V2.0 utf8
PROTO Lamp [
  exposedField SFColor color 1 1 1
  field SFFloat intensity 0.8
] {
  PointLight {
    color IS color
    intensity IS intensity
  }
}
`
	input := `#VRML V2.0 utf8
EXTERNPROTO Lamp [
  exposedField SFColor color
  field SFFloat intensity
] [ "lamp.wrl" ]

Lamp { color 1 0 0 }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(mockFetcher(map[string]string{"lamp.wrl": extFile}))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	// Should be a PointLight with color overridden
	if _, ok := nodes[0].(*node.PointLight); !ok {
		t.Fatalf("expected *node.PointLight, got %T", nodes[0])
	}
	pl := nodes[0].(*node.PointLight)
	if pl.Color.R != 1 || pl.Color.G != 0 || pl.Color.B != 0 {
		t.Errorf("expected color 1 0 0, got %v", pl.Color)
	}
}

func TestExternProto_ResolveWithFragment(t *testing.T) {
	// External file has multiple PROTOs; fragment selects SpotLamp
	extFile := `#VRML V2.0 utf8
PROTO PointLamp [
  field SFFloat intensity 0.5
] {
  PointLight { intensity IS intensity }
}

PROTO SpotLamp [
  field SFFloat cutOffAngle 0.785
] {
  SpotLight { cutOffAngle IS cutOffAngle }
}
`
	input := `#VRML V2.0 utf8
EXTERNPROTO MySpot [
  field SFFloat cutOffAngle
] [ "lights.wrl#SpotLamp" ]

MySpot { cutOffAngle 1.2 }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(mockFetcher(map[string]string{"lights.wrl": extFile}))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if _, ok := nodes[0].(*node.SpotLight); !ok {
		t.Fatalf("expected *node.SpotLight, got %T", nodes[0])
	}
	sl := nodes[0].(*node.SpotLight)
	if sl.CutOffAngle < 1.19 || sl.CutOffAngle > 1.21 {
		t.Errorf("expected cutOffAngle ~1.2, got %f", sl.CutOffAngle)
	}
}

func TestExternProto_FallbackURL(t *testing.T) {
	// First URL fails, second succeeds
	extFile := `#VRML V2.0 utf8
PROTO Light [
  field SFFloat intensity 0.5
] {
  PointLight { intensity IS intensity }
}
`
	input := `#VRML V2.0 utf8
EXTERNPROTO Light [
  field SFFloat intensity
] [ "missing.wrl" "backup.wrl" ]

Light { }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(mockFetcher(map[string]string{"backup.wrl": extFile}))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if _, ok := nodes[0].(*node.PointLight); !ok {
		t.Fatalf("expected *node.PointLight, got %T", nodes[0])
	}
	// Should use default intensity 0.5 from the PROTO
	pl := nodes[0].(*node.PointLight)
	if pl.Intensity < 0.49 || pl.Intensity > 0.51 {
		t.Errorf("expected intensity ~0.5, got %f", pl.Intensity)
	}
}

func TestExternProto_AllURLsFail(t *testing.T) {
	input := `#VRML V2.0 utf8
EXTERNPROTO Mystery [
  field SFFloat val
] [ "a.wrl" "b.wrl" ]

Mystery { val 1 }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(mockFetcher(map[string]string{}))
	nodes := p.Parse()
	if len(p.Errors()) == 0 {
		t.Fatal("expected errors when all URLs fail")
	}
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(nodes))
	}
}

func TestExternProto_DefaultValues(t *testing.T) {
	// Instance uses no overrides — defaults from resolved PROTO should apply
	extFile := `#VRML V2.0 utf8
PROTO Bulb [
  field SFFloat intensity 0.3
  field SFColor color 0 1 0
] {
  PointLight {
    intensity IS intensity
    color IS color
  }
}
`
	input := `#VRML V2.0 utf8
EXTERNPROTO Bulb [
  field SFFloat intensity
  field SFColor color
] [ "bulb.wrl" ]

Bulb { }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(mockFetcher(map[string]string{"bulb.wrl": extFile}))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	pl := nodes[0].(*node.PointLight)
	if pl.Intensity < 0.29 || pl.Intensity > 0.31 {
		t.Errorf("expected intensity ~0.3, got %f", pl.Intensity)
	}
	if pl.Color.G < 0.99 {
		t.Errorf("expected green color, got %v", pl.Color)
	}
}

func TestExternProto_MultipleInstances(t *testing.T) {
	extFile := `#VRML V2.0 utf8
PROTO MyBox [
  field SFVec3f size 1 1 1
] {
  Shape { geometry Box { size IS size } }
}
`
	input := `#VRML V2.0 utf8
EXTERNPROTO MyBox [
  field SFVec3f size
] [ "box.wrl" ]

MyBox { size 2 3 4 }
MyBox { size 5 6 7 }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(mockFetcher(map[string]string{"box.wrl": extFile}))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	// Both should be Shape nodes
	for i, n := range nodes {
		if _, ok := n.(*node.Shape); !ok {
			t.Errorf("node %d: expected *node.Shape, got %T", i, n)
		}
	}
}

func TestExternProto_NestedExternProto(t *testing.T) {
	// extA.wrl has an EXTERNPROTO that references extB.wrl
	extB := `#VRML V2.0 utf8
PROTO Inner [
  field SFFloat val 0.5
] {
  PointLight { intensity IS val }
}
`
	extA := `#VRML V2.0 utf8
EXTERNPROTO Inner [
  field SFFloat val
] [ "extB.wrl" ]

PROTO Outer [
  field SFFloat brightness 0.7
] {
  Inner { val IS brightness }
}
`
	input := `#VRML V2.0 utf8
EXTERNPROTO Outer [
  field SFFloat brightness
] [ "extA.wrl" ]

Outer { brightness 0.9 }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(mockFetcher(map[string]string{
		"extA.wrl": extA,
		"extB.wrl": extB,
	}))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if _, ok := nodes[0].(*node.PointLight); !ok {
		t.Fatalf("expected *node.PointLight, got %T", nodes[0])
	}
}

func TestExternProto_URLsStoredInDef(t *testing.T) {
	input := `#VRML V2.0 utf8
EXTERNPROTO Foo [
  field SFFloat x
] [ "a.wrl" "b.wrl#Bar" ]
`
	p := NewParser(strings.NewReader(input))
	p.Parse()
	def, ok := p.ProtoTable()["Foo"]
	if !ok {
		t.Fatal("expected Foo in proto table")
	}
	if len(def.URLs) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(def.URLs))
	}
	if def.URLs[0] != "a.wrl" || def.URLs[1] != "b.wrl#Bar" {
		t.Fatalf("unexpected URLs: %v", def.URLs)
	}
}

func TestExternProto_SubsetInterface(t *testing.T) {
	// EXTERNPROTO exposes only 'intensity', but PROTO also has 'color'
	extFile := `#VRML V2.0 utf8
PROTO FullLight [
  field SFFloat intensity 0.5
  field SFColor color 1 0 0
] {
  PointLight {
    intensity IS intensity
    color IS color
  }
}
`
	input := `#VRML V2.0 utf8
EXTERNPROTO FullLight [
  field SFFloat intensity
] [ "fulllight.wrl" ]

FullLight { intensity 0.9 }
`
	p := NewParser(strings.NewReader(input))
	p.SetURLFetcher(mockFetcher(map[string]string{"fulllight.wrl": extFile}))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	pl := nodes[0].(*node.PointLight)
	// Intensity overridden
	if pl.Intensity < 0.89 || pl.Intensity > 0.91 {
		t.Errorf("expected intensity ~0.9, got %f", pl.Intensity)
	}
	// Color should use PROTO default (1 0 0)
	if pl.Color.R < 0.99 || pl.Color.G > 0.01 || pl.Color.B > 0.01 {
		t.Errorf("expected red color (PROTO default), got %v", pl.Color)
	}
}

func TestExternProto_LocalFileResolution(t *testing.T) {
	// Test with actual temp files
	tmpDir := t.TempDir()

	extContent := `#VRML V2.0 utf8
PROTO Widget [
  field SFFloat size 1
] {
  Shape { geometry Box { size IS size IS size IS size } }
}
`
	err := os.WriteFile(tmpDir+"/widget.wrl", []byte(extContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	input := `#VRML V2.0 utf8
EXTERNPROTO Widget [
  field SFFloat size
] [ "widget.wrl" ]

Widget { size 5 }
`
	p := NewParser(strings.NewReader(input))
	p.SetBaseDir(tmpDir)
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("errors: %v", p.Errors())
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if _, ok := nodes[0].(*node.Shape); !ok {
		t.Fatalf("expected Shape, got %T", nodes[0])
	}
}
