package parser

import (
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
)

func TestRoutesParsed(t *testing.T) {
	vrml := `#VRML V2.0 utf8
DEF TIMER TimeSensor { cycleInterval 4.0 loop TRUE }
DEF BOUNCE PositionInterpolator { key [ 0 1 ] keyValue [ 0 0 0, 0 3 0 ] }
ROUTE TIMER.fraction_changed TO BOUNCE.set_fraction
`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	routes := p.GetRoutes()
	t.Logf("nodes=%d routes=%d errors=%v", len(nodes), len(routes), p.Errors())
	for i, r := range routes {
		t.Logf("route[%d]: %s.%s -> %s.%s", i, r.Source.GetName(), r.SrcField, r.Destination.GetName(), r.DstField)
	}
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if routes[0].SrcField != "fraction_changed" {
		t.Errorf("expected srcField=fraction_changed, got %s", routes[0].SrcField)
	}
	if routes[0].DstField != "set_fraction" {
		t.Errorf("expected dstField=set_fraction, got %s", routes[0].DstField)
	}
}

func TestTimeSensorFieldsParsed(t *testing.T) {
	vrml := `#VRML V2.0 utf8
DEF T TimeSensor { cycleInterval 5.0 loop TRUE startTime 1.0 stopTime 10.0 }
`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	ts, ok := nodes[0].(*node.TimeSensor)
	if !ok {
		t.Fatalf("expected TimeSensor, got %T", nodes[0])
	}
	if ts.CycleInterval != 5.0 {
		t.Errorf("cycleInterval: got %f, want 5.0", ts.CycleInterval)
	}
	if !ts.Loop {
		t.Error("loop should be true")
	}
	if ts.StartTime != 1.0 {
		t.Errorf("startTime: got %f, want 1.0", ts.StartTime)
	}
	if ts.StopTime != 10.0 {
		t.Errorf("stopTime: got %f, want 10.0", ts.StopTime)
	}
}

func TestPositionInterpolatorParsed(t *testing.T) {
	vrml := `#VRML V2.0 utf8
DEF PI PositionInterpolator {
  key [ 0.0, 0.5, 1.0 ]
  keyValue [ 0 0 0, 0 3 0, 0 0 0 ]
}
`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	pi, ok := nodes[0].(*node.PositionInterpolator)
	if !ok {
		t.Fatalf("expected PositionInterpolator, got %T", nodes[0])
	}
	if len(pi.Key) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(pi.Key))
	}
	if len(pi.KeyValue) != 3 {
		t.Fatalf("expected 3 keyValues, got %d", len(pi.KeyValue))
	}
	if pi.KeyValue[1].Y != 3.0 {
		t.Errorf("keyValue[1].Y: got %f, want 3.0", pi.KeyValue[1].Y)
	}
}

func TestOrientationInterpolatorParsed(t *testing.T) {
	vrml := `#VRML V2.0 utf8
DEF OI OrientationInterpolator {
  key [ 0.0, 1.0 ]
  keyValue [ 0 1 0 0, 0 1 0 3.14 ]
}
`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	oi, ok := nodes[0].(*node.OrientationInterpolator)
	if !ok {
		t.Fatalf("expected OrientationInterpolator, got %T", nodes[0])
	}
	if len(oi.Key) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(oi.Key))
	}
	if len(oi.KeyValue) != 2 {
		t.Fatalf("expected 2 keyValues, got %d", len(oi.KeyValue))
	}
	if oi.KeyValue[1].W < 3.13 || oi.KeyValue[1].W > 3.15 {
		t.Errorf("keyValue[1].W: got %f, want ~3.14", oi.KeyValue[1].W)
	}
}

func TestPixelTextureFieldsParsed(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  appearance Appearance {
    texture PixelTexture {
      image 2 2 3 0xFF0000 0x00FF00 0x0000FF 0xFFFF00
      repeatS FALSE
      repeatT TRUE
    }
  }
  geometry Box {}
}
`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	s, ok := nodes[0].(*node.Shape)
	if !ok {
		t.Fatalf("expected Shape, got %T", nodes[0])
	}
	pt, ok := s.Appearance.Texture.(*node.PixelTexture)
	if !ok {
		t.Fatalf("expected PixelTexture, got %T", s.Appearance.Texture)
	}
	if pt.Image.Width != 2 || pt.Image.Height != 2 || pt.Image.NumComponents != 3 {
		t.Errorf("image dims: got %dx%dx%d, want 2x2x3", pt.Image.Width, pt.Image.Height, pt.Image.NumComponents)
	}
	if len(pt.Image.Pixels) != 12 {
		t.Fatalf("expected 12 pixel bytes, got %d", len(pt.Image.Pixels))
	}
	// First pixel: 0xFF0000 -> R=255, G=0, B=0
	if pt.Image.Pixels[0] != 255 || pt.Image.Pixels[1] != 0 || pt.Image.Pixels[2] != 0 {
		t.Errorf("pixel[0]: got %d,%d,%d want 255,0,0", pt.Image.Pixels[0], pt.Image.Pixels[1], pt.Image.Pixels[2])
	}
	if pt.RepeatS {
		t.Error("repeatS should be false")
	}
	if !pt.RepeatT {
		t.Error("repeatT should be true")
	}
}

func TestTextureTransformFieldsParsed(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  appearance Appearance {
    texture ImageTexture { url "test.jpg" }
    textureTransform TextureTransform {
      center 0.5 0.5
      rotation 1.57
      scale 2.0 3.0
      translation 0.1 0.2
    }
  }
  geometry Box {}
}
`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	s := nodes[0].(*node.Shape)
	tt := s.Appearance.TextureTransform
	if tt == nil {
		t.Fatal("expected TextureTransform")
		return
	}
	if tt.Center.X != 0.5 || tt.Center.Y != 0.5 {
		t.Errorf("center: got %v, want {0.5, 0.5}", tt.Center)
	}
	if tt.Rotation < 1.56 || tt.Rotation > 1.58 {
		t.Errorf("rotation: got %f, want ~1.57", tt.Rotation)
	}
	if tt.Scale.X != 2.0 || tt.Scale.Y != 3.0 {
		t.Errorf("scale: got %v, want {2.0, 3.0}", tt.Scale)
	}
	if tt.Translation.X < 0.09 || tt.Translation.X > 0.11 {
		t.Errorf("translation.X: got %f, want ~0.1", tt.Translation.X)
	}
}

func TestMovieTextureFieldsParsed(t *testing.T) {
	vrml := `#VRML V2.0 utf8
Shape {
  appearance Appearance {
    texture MovieTexture {
      url [ "movie.mpg" ]
      loop TRUE
      speed 2.0
      repeatS FALSE
    }
  }
  geometry Sphere {}
}
`
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	s := nodes[0].(*node.Shape)
	mt, ok := s.Appearance.Texture.(*node.MovieTexture)
	if !ok {
		t.Fatalf("expected MovieTexture, got %T", s.Appearance.Texture)
	}
	if len(mt.URL) != 1 || mt.URL[0] != "movie.mpg" {
		t.Errorf("url: got %v, want [movie.mpg]", mt.URL)
	}
	if !mt.Loop {
		t.Error("loop should be true")
	}
	if mt.Speed != 2.0 {
		t.Errorf("speed: got %f, want 2.0", mt.Speed)
	}
	if mt.RepeatS {
		t.Error("repeatS should be false")
	}
	if !mt.RepeatT {
		t.Error("repeatT should be true (default)")
	}
}
