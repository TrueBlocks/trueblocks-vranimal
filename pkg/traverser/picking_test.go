package traverser

import (
	"testing"

	"github.com/g3n/engine/core"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func TestFindSiblingSensors(t *testing.T) {
	ts := node.NewTouchSensor()
	ps := node.NewPlaneSensor()
	shape := &node.Shape{}
	children := []node.Node{ts, shape, ps}

	sensors := findSiblingSensors(children)
	if len(sensors) != 2 {
		t.Fatalf("expected 2 sensors, got %d", len(sensors))
	}
	if sensors[0] != ts {
		t.Error("expected first sensor to be TouchSensor")
	}
	if sensors[1] != ps {
		t.Error("expected second sensor to be PlaneSensor")
	}
}

func TestFindSiblingSensorsNoSensors(t *testing.T) {
	children := []node.Node{&node.Shape{}, &node.Box{}}
	sensors := findSiblingSensors(children)
	if len(sensors) != 0 {
		t.Fatalf("expected 0 sensors, got %d", len(sensors))
	}
}

func TestFindSiblingSensorsAllTypes(t *testing.T) {
	children := []node.Node{
		node.NewTouchSensor(),
		node.NewPlaneSensor(),
		node.NewSphereSensor(),
		node.NewCylinderSensor(),
		&node.Shape{},
	}
	sensors := findSiblingSensors(children)
	if len(sensors) != 4 {
		t.Fatalf("expected 4 sensors, got %d", len(sensors))
	}
}

func TestProjectOntoSphereInside(t *testing.T) {
	p := projectOntoSphere(vec.SFVec3f{X: 0, Y: 0, Z: 0})
	if p.Z != 1.0 {
		t.Errorf("center project Z should be 1.0, got %f", p.Z)
	}
}

func TestProjectOntoSphereEdge(t *testing.T) {
	p := projectOntoSphere(vec.SFVec3f{X: 2, Y: 0, Z: 0})
	// Outside unit circle, should be on the equator (Z=0)
	if p.Z != 0 {
		t.Errorf("edge project Z should be 0, got %f", p.Z)
	}
}

func TestClampf(t *testing.T) {
	tests := []struct {
		v, min, max, want float32
	}{
		{5, 0, 10, 5},
		{-1, 0, 10, 0},
		{15, 0, 10, 10},
	}
	for _, tt := range tests {
		got := clampf(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clampf(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

func TestTouchSensorDispatch(t *testing.T) {
	ts := node.NewTouchSensor()

	// Simulate the same dispatch logic as Picker.handleTouchSensor
	// without needing g3n/OpenGL
	hitPoint := vec.SFVec3f{X: 1, Y: 2, Z: 3}
	hitNormal := vec.SFVec3f{X: 0, Y: 1, Z: 0}

	// Create a minimal picker (no g3n needed for these tests)
	p := &Picker{SimTime: 10.0}

	// PointerDown
	p.handleTouchSensor(ts, PointerDown, hitPoint, hitNormal)
	if !ts.IsActive {
		t.Error("expected IsActive=true after PointerDown")
	}
	if !ts.IsOver {
		t.Error("expected IsOver=true after PointerDown")
	}
	if ts.HitPoint != hitPoint {
		t.Error("expected HitPoint to be set")
	}
	if p.captured != ts {
		t.Error("expected sensor to be captured")
	}

	// PointerUp (via handleRelease since sensor is captured)
	p.handleRelease()
	if ts.IsActive {
		t.Error("expected IsActive=false after release")
	}
	if ts.TouchTime != 10.0 {
		t.Errorf("expected TouchTime=10.0, got %f", ts.TouchTime)
	}
	if p.captured != nil {
		t.Error("expected captured to be nil after release")
	}
}

func TestDisabledTouchSensor(t *testing.T) {
	ts := node.NewTouchSensor()
	ts.Enabled = false

	p := &Picker{SimTime: 5.0}
	hit := vec.SFVec3f{X: 1, Y: 0, Z: 0}
	consumed := p.handleTouchSensor(ts, PointerDown, hit, hit)
	if consumed {
		t.Error("disabled sensor should not consume event")
	}
	if ts.IsActive {
		t.Error("disabled sensor should not become active")
	}
}

func TestPlaneSensorCapture(t *testing.T) {
	ps := node.NewPlaneSensor()
	p := &Picker{SimTime: 1.0}

	hitPoint := vec.SFVec3f{X: 0.5, Y: 0.5, Z: 0}
	consumed := p.handlePlaneSensor(ps, PointerDown, hitPoint)
	if !consumed {
		t.Error("expected PlaneSensor to consume PointerDown")
	}
	if !ps.IsActive {
		t.Error("expected IsActive=true")
	}
	if p.captured != ps {
		t.Error("expected PlaneSensor to be captured")
	}

	// Release with autoOffset
	p.handleRelease()
	if ps.IsActive {
		t.Error("expected IsActive=false after release")
	}
}

func TestSphereSensorCapture(t *testing.T) {
	ss := node.NewSphereSensor()
	p := &Picker{SimTime: 1.0}

	hitPoint := vec.SFVec3f{X: 0.5, Y: 0.5, Z: 0}
	consumed := p.handleSphereSensor(ss, PointerDown, hitPoint)
	if !consumed {
		t.Error("expected SphereSensor to consume PointerDown")
	}
	if !ss.IsActive {
		t.Error("expected IsActive=true")
	}

	p.handleRelease()
	if ss.IsActive {
		t.Error("expected IsActive=false after release")
	}
}

func TestCylinderSensorCapture(t *testing.T) {
	cs := node.NewCylinderSensor()
	p := &Picker{SimTime: 1.0}

	hitPoint := vec.SFVec3f{X: 0.5, Y: 0, Z: 0}
	consumed := p.handleCylinderSensor(cs, PointerDown, hitPoint)
	if !consumed {
		t.Error("expected CylinderSensor to consume PointerDown")
	}
	if p.captured != cs {
		t.Error("expected capture")
	}

	p.handleRelease()
	if cs.IsActive {
		t.Error("expected IsActive=false after release")
	}
}

func TestFindAnchorParent(t *testing.T) {
	anchor := &node.Anchor{URL: []string{"http://example.com"}, Description: "test"}
	parent := core.NewNode()
	parent.SetUserData(anchor)

	child := core.NewNode()
	parent.Add(child)

	result := findAnchorParent(child)
	if result != anchor {
		t.Fatal("expected to find Anchor in parent chain")
	}
}

func TestFindAnchorParentNone(t *testing.T) {
	parent := core.NewNode()
	child := core.NewNode()
	parent.Add(child)

	result := findAnchorParent(child)
	if result != nil {
		t.Fatal("expected nil when no Anchor in parent chain")
	}
}
