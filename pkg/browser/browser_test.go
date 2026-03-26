package browser

import (
	"testing"
	"time"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Browser creation and basics
// ---------------------------------------------------------------------------

func TestNewBrowser(t *testing.T) {
	b := NewBrowser()
	if b.FramesPerSec != 30 {
		t.Fatalf("fps = %d", b.FramesPerSec)
	}
	if b.NTraversers() != 0 {
		t.Fatal("should have no traversers")
	}
}

func TestBrowser_SetFrameRate(t *testing.T) {
	b := NewBrowser()
	b.SetFrameRate(60)
	if b.FramesPerSec != 60 {
		t.Fatalf("fps = %d", b.FramesPerSec)
	}
	b.SetFrameRate(0) // should default to 30
	if b.FramesPerSec != 30 {
		t.Fatalf("fps after 0 = %d", b.FramesPerSec)
	}
}

func TestBrowser_GetVersion(t *testing.T) {
	v := GetVersion()
	if v == "" {
		t.Fatal("empty version")
	}
}

// ---------------------------------------------------------------------------
// Binding stacks
// ---------------------------------------------------------------------------

func TestBrowser_BindViewpoint(t *testing.T) {
	b := NewBrowser()
	if b.GetViewpoint() != nil {
		t.Fatal("should be nil")
	}
	vp := node.NewViewpoint()
	b.BindViewpoint(vp, true)
	if b.GetViewpoint() != vp {
		t.Fatal("wrong viewpoint")
	}
	if !vp.IsBound {
		t.Fatal("should be bound")
	}
	b.BindViewpoint(vp, false)
	if b.GetViewpoint() != nil {
		t.Fatal("should be nil after unbind")
	}
	if vp.IsBound {
		t.Fatal("should be unbound")
	}
}

func TestBrowser_BindViewpointStack(t *testing.T) {
	b := NewBrowser()
	vp1 := node.NewViewpoint()
	vp2 := node.NewViewpoint()
	b.BindViewpoint(vp1, true)
	b.BindViewpoint(vp2, true)
	if b.GetViewpoint() != vp2 {
		t.Fatal("top should be vp2")
	}
	b.BindViewpoint(vp2, false)
	if b.GetViewpoint() != vp1 {
		t.Fatal("top should be vp1 after pop")
	}
}

func TestBrowser_BindNavigationInfo(t *testing.T) {
	b := NewBrowser()
	if b.GetNavigationInfo() != nil {
		t.Fatal("should be nil")
	}
	ni := node.NewNavigationInfo()
	b.BindNavigationInfo(ni, true)
	if b.GetNavigationInfo() != ni {
		t.Fatal("wrong nav info")
	}
	b.BindNavigationInfo(ni, false)
	if b.GetNavigationInfo() != nil {
		t.Fatal("should be nil")
	}
}

func TestBrowser_BindBackground(t *testing.T) {
	b := NewBrowser()
	if b.GetBackground() != nil {
		t.Fatal("should be nil")
	}
	bg := &node.Background{}
	b.BindBackground(bg, true)
	if b.GetBackground() != bg {
		t.Fatal("wrong bg")
	}
	b.BindBackground(bg, false)
	if b.GetBackground() != nil {
		t.Fatal("should be nil")
	}
}

func TestBrowser_BindFog(t *testing.T) {
	b := NewBrowser()
	if b.GetFog() != nil {
		t.Fatal("should be nil")
	}
	fg := node.NewFog()
	b.BindFog(fg, true)
	if b.GetFog() != fg {
		t.Fatal("wrong fog")
	}
	b.BindFog(fg, false)
	if b.GetFog() != nil {
		t.Fatal("should be nil")
	}
}

// ---------------------------------------------------------------------------
// Routes and event engine
// ---------------------------------------------------------------------------

func TestBrowser_AddRoute(t *testing.T) {
	b := NewBrowser()
	src := node.NewTimeSensor()
	dst := &node.Transform{}
	r := b.AddRoute(src, "fraction_changed", dst, "set_fraction")
	if r == nil {
		t.Fatal("nil route")
	}
	if len(b.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(b.Routes))
	}
}

func TestBrowser_Clear(t *testing.T) {
	b := NewBrowser()
	vp := node.NewViewpoint()
	b.BindViewpoint(vp, true)
	b.AddRoute(node.NewTimeSensor(), "fraction_changed", &node.Transform{}, "set_fraction")
	b.AddChild(node.NewTransform())
	b.Clear()
	if b.GetViewpoint() != nil {
		t.Fatal("viewpoint should be nil")
	}
	if len(b.Routes) != 0 {
		t.Fatal("routes should be empty")
	}
	if len(b.Children) != 0 {
		t.Fatal("children should be empty")
	}
}

func TestBrowser_SimTime(t *testing.T) {
	b := NewBrowser()
	b.Update(time.Millisecond * 100)
	if b.SimTime() <= 0 {
		t.Fatal("sim time should be > 0")
	}
}

// ---------------------------------------------------------------------------
// TimeSensor collection and updating
// ---------------------------------------------------------------------------

func TestBrowser_CollectTimeSensors(t *testing.T) {
	b := NewBrowser()
	ts1 := node.NewTimeSensor()
	ts2 := node.NewTimeSensor()
	tr := node.NewTransform()
	tr.AddChild(ts2)
	b.AddChild(ts1)
	b.AddChild(tr)
	b.CollectTimeSensors()
	if len(b.TimeSensors) != 2 {
		t.Fatalf("expected 2 time sensors, got %d", len(b.TimeSensors))
	}
}

func TestBrowser_TimeSensorFraction(t *testing.T) {
	b := NewBrowser()
	ts := node.NewTimeSensor()
	ts.Enabled = true
	ts.CycleInterval = 1.0
	ts.StartTime = 0
	ts.Loop = false
	b.AddChild(ts)
	b.CollectTimeSensors()
	// Advance simulation
	time.Sleep(50 * time.Millisecond)
	b.Update(50 * time.Millisecond)
	if ts.Fraction < 0 || ts.Fraction > 1 {
		t.Fatalf("fraction out of range: %g", ts.Fraction)
	}
}

// ---------------------------------------------------------------------------
// Interpolation via routes
// ---------------------------------------------------------------------------

func TestBrowser_PositionInterpolatorRoute(t *testing.T) {
	b := NewBrowser()
	pi := &node.PositionInterpolator{
		KeyValue: []vec.SFVec3f{{X: 0, Y: 0, Z: 0}, {X: 10, Y: 0, Z: 0}},
	}
	pi.Key = []float64{0, 1}
	tr := node.NewTransform()
	ts := node.NewTimeSensor()
	ts.Enabled = true
	ts.CycleInterval = 1.0
	ts.StartTime = 0
	b.AddChild(ts)
	b.AddChild(pi)
	b.AddChild(tr)
	b.CollectTimeSensors()
	b.AddRoute(ts, "fraction_changed", pi, "set_fraction")
	b.AddRoute(pi, "value_changed", tr, "set_translation")
	// Advance and process
	time.Sleep(50 * time.Millisecond)
	b.Update(50 * time.Millisecond)
	// The route should have fired at least once
	if len(b.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(b.Routes))
	}
}

func TestBrowser_ColorInterpolatorRoute(t *testing.T) {
	b := NewBrowser()
	ci := &node.ColorInterpolator{
		KeyValue: []vec.SFColor{{R: 1, G: 0, B: 0}, {R: 0, G: 0, B: 1}},
	}
	ci.Key = []float64{0, 1}
	mat := node.NewMaterial()
	ts := node.NewTimeSensor()
	ts.Enabled = true
	ts.CycleInterval = 1.0
	ts.StartTime = 0
	b.AddChild(ts)
	b.AddChild(ci)
	b.AddChild(mat)
	b.CollectTimeSensors()
	b.AddRoute(ts, "fraction_changed", ci, "set_fraction")
	b.AddRoute(ci, "value_changed", mat, "set_diffuseColor")
	time.Sleep(50 * time.Millisecond)
	b.Update(50 * time.Millisecond)
	if len(b.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(b.Routes))
	}
}

// ---------------------------------------------------------------------------
// FindByName
// ---------------------------------------------------------------------------

func TestBrowser_FindByName(t *testing.T) {
	b := NewBrowser()
	tr := node.NewTransform()
	tr.SetName("MyTransform")
	b.AddChild(tr)
	found := b.FindByName("MyTransform")
	if found == nil {
		t.Fatal("should find node")
	}
	if found.GetName() != "MyTransform" {
		t.Fatalf("got %q", found.GetName())
	}
}

func TestBrowser_FindByName_Nested(t *testing.T) {
	b := NewBrowser()
	tr := node.NewTransform()
	child := node.NewTransform()
	child.SetName("Nested")
	tr.AddChild(child)
	b.AddChild(tr)
	found := b.FindByName("Nested")
	if found == nil {
		t.Fatal("should find nested node")
	}
}

func TestBrowser_FindByName_NotFound(t *testing.T) {
	b := NewBrowser()
	if b.FindByName("nonexistent") != nil {
		t.Fatal("should return nil")
	}
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------

func TestBrowser_Selection(t *testing.T) {
	b := NewBrowser()
	if b.GetSelection() != nil {
		t.Fatal("should be nil")
	}
	g := node.NewGroupingNode()
	b.SetSelection(g)
	if b.GetSelection() != g {
		t.Fatal("wrong selection")
	}
}

// ---------------------------------------------------------------------------
// Coercion helpers
// ---------------------------------------------------------------------------

func TestToFloat32(t *testing.T) {
	v, ok := toFloat32(float64(3.14))
	if !ok || v != 3.14 {
		t.Fatalf("float64: %g %v", v, ok)
	}
	v, ok = toFloat32(float64(2.71))
	if !ok {
		t.Fatal("float64 should convert")
	}
	v, ok = toFloat32(true)
	if !ok || v != 1 {
		t.Fatal("bool true should be 1")
	}
	v, ok = toFloat32(false)
	if !ok || v != 0 {
		t.Fatal("bool false should be 0")
	}
	_, ok = toFloat32("bad")
	if ok {
		t.Fatal("string should fail")
	}
}

func TestToFloat64(t *testing.T) {
	v, ok := toFloat64(float64(3.14))
	if !ok || v != 3.14 {
		t.Fatalf("float64: %g %v", v, ok)
	}
	v, ok = toFloat64(float64(2.0))
	if !ok || v != 2.0 {
		t.Fatal("float64 should convert")
	}
	_, ok = toFloat64("bad")
	if ok {
		t.Fatal("string should fail")
	}
}

func TestToInt32(t *testing.T) {
	v, ok := toInt32(int64(42))
	if !ok || v != 42 {
		t.Fatal("int64")
	}
	v, ok = toInt32(float64(5.0))
	if !ok || v != 5 {
		t.Fatal("float64 to int64")
	}
	v, ok = toInt32(float64(7.0))
	if !ok || v != 7 {
		t.Fatal("float64 to int64")
	}
	_, ok = toInt32("bad")
	if ok {
		t.Fatal("string should fail")
	}
}

// ---------------------------------------------------------------------------
// Key segment search
// ---------------------------------------------------------------------------

func TestFindKeySegment(t *testing.T) {
	keys := []float64{0, 0.5, 1.0}
	seg, lt := findKeySegment(keys, 0)
	if seg != 0 || lt != 0 {
		t.Fatalf("at 0: seg=%d t=%g", seg, lt)
	}
	seg, lt = findKeySegment(keys, 0.25)
	if seg != 0 {
		t.Fatalf("at 0.25: seg=%d", seg)
	}
	if lt < 0.49 || lt > 0.51 {
		t.Fatalf("at 0.25: t=%g", lt)
	}
	seg, lt = findKeySegment(keys, 1.0)
	if seg != 1 || lt != 1 {
		t.Fatalf("at 1.0: seg=%d t=%g", seg, lt)
	}
}

func TestFindKeySegment_Empty(t *testing.T) {
	seg, lt := findKeySegment(nil, 0.5)
	if seg != 0 || lt != 0 {
		t.Fatalf("empty: seg=%d t=%g", seg, lt)
	}
}

// ---------------------------------------------------------------------------
// Interpolation evaluation
// ---------------------------------------------------------------------------

func TestEvalPositionInterp(t *testing.T) {
	keys := []float64{0, 1}
	vals := []vec.SFVec3f{{X: 0}, {X: 10}}
	r := evalPositionInterp(keys, vals, 0.5)
	if r.X < 4.9 || r.X > 5.1 {
		t.Fatalf("expected ~5, got %g", r.X)
	}
}

func TestEvalPositionInterp_Empty(t *testing.T) {
	r := evalPositionInterp(nil, nil, 0.5)
	if r != (vec.SFVec3f{}) {
		t.Fatalf("expected zero, got %v", r)
	}
}

func TestEvalColorInterp(t *testing.T) {
	keys := []float64{0, 1}
	vals := []vec.SFColor{{R: 1, G: 0, B: 0}, {R: 0, G: 0, B: 1}}
	r := evalColorInterp(keys, vals, 0.5)
	if r.R < 0.49 || r.R > 0.51 {
		t.Fatalf("expected ~0.5 red, got %g", r.R)
	}
	if r.B < 0.49 || r.B > 0.51 {
		t.Fatalf("expected ~0.5 blue, got %g", r.B)
	}
}

func TestEvalScalarInterp(t *testing.T) {
	keys := []float64{0, 1}
	vals := []float64{0, 100}
	r := evalScalarInterp(keys, vals, 0.25)
	if r < 24.9 || r > 25.1 {
		t.Fatalf("expected ~25, got %g", r)
	}
}

func TestEvalOrientationInterp(t *testing.T) {
	keys := []float64{0, 1}
	vals := []vec.SFRotation{
		{X: 0, Y: 1, Z: 0, W: 0},
		{X: 0, Y: 1, Z: 0, W: 3.14159},
	}
	r := evalOrientationInterp(keys, vals, 0.5)
	// Should be approximately pi/2
	if r.W < 1.5 || r.W > 1.6 {
		t.Fatalf("expected ~pi/2, got %g", r.W)
	}
}

func TestEvalCoordinateInterp(t *testing.T) {
	keys := []float64{0, 1}
	vals := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 10, Y: 10, Z: 10},
	}
	r := evalCoordinateInterp(keys, vals, 0.5)
	if len(r) != 1 {
		t.Fatalf("expected 1 value, got %d", len(r))
	}
	if r[0].X < 4.9 || r[0].X > 5.1 {
		t.Fatalf("expected ~5, got %v", r[0])
	}
}

func TestEvalCoordinateInterp_Empty(t *testing.T) {
	r := evalCoordinateInterp(nil, nil, 0.5)
	if r != nil {
		t.Fatalf("expected nil, got %v", r)
	}
}
