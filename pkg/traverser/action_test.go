package traverser

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func TestCollectSensorsFindsAll(t *testing.T) {
	ts := node.NewTouchSensor()
	ps := node.NewProximitySensor()
	cs := node.NewCylinderSensor()
	lod := node.NewLOD()
	lod.Level = []node.Node{&node.BaseNode{}}

	root := &node.Transform{}
	root.Children = []node.Node{ts, ps, cs, lod}

	at := NewActionTraverser()
	at.CollectSensors([]node.Node{root})

	if len(at.TouchSensors) != 1 {
		t.Errorf("expected 1 TouchSensor, got %d", len(at.TouchSensors))
	}
	if len(at.ProximitySensors) != 1 {
		t.Errorf("expected 1 ProximitySensor, got %d", len(at.ProximitySensors))
	}
	if len(at.LODs) != 1 {
		t.Errorf("expected 1 LOD, got %d", len(at.LODs))
	}
	if !at.HasSensor {
		t.Error("expected HasSensor to be true")
	}
}

func TestCollectSensorsNested(t *testing.T) {
	ts := node.NewTouchSensor()
	inner := &node.Group{}
	inner.Children = []node.Node{ts}
	outer := &node.Transform{}
	outer.Children = []node.Node{inner}

	at := NewActionTraverser()
	at.CollectSensors([]node.Node{outer})

	if len(at.TouchSensors) != 1 {
		t.Errorf("expected 1 TouchSensor, got %d", len(at.TouchSensors))
	}
	if !at.HasSensor {
		t.Error("expected HasSensor true")
	}
}

func TestProximitySensorEnterExit(t *testing.T) {
	ps := node.NewProximitySensor()
	ps.Center = vec.SFVec3f{X: 0, Y: 0, Z: 0}
	ps.Size = vec.SFVec3f{X: 10, Y: 10, Z: 10}

	at := NewActionTraverser()
	at.ProximitySensors = []*node.ProximitySensor{ps}

	at.Update(vec.SFVec3f{X: 20, Y: 0, Z: 0}, 1.0)
	if ps.IsActive {
		t.Error("expected inactive when viewer is outside")
	}

	at.Update(vec.SFVec3f{X: 3, Y: 0, Z: 0}, 2.0)
	if !ps.IsActive {
		t.Error("expected active when viewer is inside")
	}
	if ps.EnterTime != 2.0 {
		t.Errorf("enterTime: got %f, want 2.0", ps.EnterTime)
	}

	at.Update(vec.SFVec3f{X: 1, Y: 1, Z: 0}, 3.0)
	if !ps.IsActive {
		t.Error("expected still active")
	}
	if ps.Position.X != 1 || ps.Position.Y != 1 {
		t.Errorf("position: got %v, want {1,1,0}", ps.Position)
	}

	at.Update(vec.SFVec3f{X: 20, Y: 0, Z: 0}, 4.0)
	if ps.IsActive {
		t.Error("expected inactive after leaving")
	}
	if ps.ExitTime != 4.0 {
		t.Errorf("exitTime: got %f, want 4.0", ps.ExitTime)
	}
}

func TestLODDistanceSwitching(t *testing.T) {
	lod := node.NewLOD()
	lod.Center = vec.SFVec3f{X: 0, Y: 0, Z: 0}
	lod.Range = []float32{5, 15, 30}
	lod.Level = []node.Node{
		&node.BaseNode{},
		&node.BaseNode{},
		&node.BaseNode{},
		&node.BaseNode{},
	}

	at := NewActionTraverser()
	at.LODs = []*node.LOD{lod}

	at.Update(vec.SFVec3f{X: 3, Y: 0, Z: 0}, 0)
	if lod.ActiveLevel != 0 {
		t.Errorf("expected level 0 at dist=3, got %d", lod.ActiveLevel)
	}

	at.Update(vec.SFVec3f{X: 10, Y: 0, Z: 0}, 0)
	if lod.ActiveLevel != 1 {
		t.Errorf("expected level 1 at dist=10, got %d", lod.ActiveLevel)
	}

	at.Update(vec.SFVec3f{X: 20, Y: 0, Z: 0}, 0)
	if lod.ActiveLevel != 2 {
		t.Errorf("expected level 2 at dist=20, got %d", lod.ActiveLevel)
	}

	at.Update(vec.SFVec3f{X: 50, Y: 0, Z: 0}, 0)
	if lod.ActiveLevel != 3 {
		t.Errorf("expected level 3 at dist=50, got %d", lod.ActiveLevel)
	}
}

func TestDisabledProximitySensor(t *testing.T) {
	ps := node.NewProximitySensor()
	ps.Center = vec.SFVec3f{X: 0, Y: 0, Z: 0}
	ps.Size = vec.SFVec3f{X: 10, Y: 10, Z: 10}
	ps.Enabled = false

	at := NewActionTraverser()
	at.ProximitySensors = []*node.ProximitySensor{ps}

	at.Update(vec.SFVec3f{X: 0, Y: 0, Z: 0}, 1.0)
	if ps.IsActive {
		t.Error("disabled sensor should not activate")
	}
}
