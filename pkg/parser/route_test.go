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
