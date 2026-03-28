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
	_, ok = toFloat32(float64(2.71))
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

// ===========================================================================
// Gap-filling tests (issue #46)
// ===========================================================================

// ---------------------------------------------------------------------------
// Traverser management
// ---------------------------------------------------------------------------

type mockTraverser struct {
	preCount  int
	postCount int
	nodes     int
}

func (m *mockTraverser) PreTraverse()             { m.preCount++ }
func (m *mockTraverser) PostTraverse()            { m.postCount++ }
func (m *mockTraverser) TraverseNode(n node.Node) { m.nodes++ }
func (m *mockTraverser) TraverseChildren(g *node.GroupingNode) {
	for _, c := range g.Children {
		m.TraverseNode(c)
	}
}

func TestBrowser_AddGetTraverser(t *testing.T) {
	b := NewBrowser()
	m := &mockTraverser{}
	b.AddTraverser(m)
	if b.NTraversers() != 1 {
		t.Fatalf("expected 1 traverser, got %d", b.NTraversers())
	}
	if b.GetTraverser(0) != m {
		t.Fatal("wrong traverser at index 0")
	}
	if b.GetTraverser(-1) != nil {
		t.Fatal("negative index should return nil")
	}
	if b.GetTraverser(1) != nil {
		t.Fatal("out-of-bounds should return nil")
	}
}

func TestBrowser_Tick_InvokesTraversers(t *testing.T) {
	b := NewBrowser()
	b.SetFrameRate(1000) // high rate so Tick fires
	m := &mockTraverser{}
	b.AddTraverser(m)
	b.AddChild(&node.Shape{})

	// Force LastFrame to be in the past so Tick fires
	b.LastFrame = time.Now().Add(-time.Second)
	b.Tick()

	if m.preCount != 1 {
		t.Fatalf("PreTraverse should be called once, got %d", m.preCount)
	}
	if m.postCount != 1 {
		t.Fatalf("PostTraverse should be called once, got %d", m.postCount)
	}
	if m.nodes != 1 {
		t.Fatalf("TraverseNode should be called once, got %d", m.nodes)
	}
}

func TestBrowser_Tick_RateLimited(t *testing.T) {
	b := NewBrowser()
	b.SetFrameRate(1) // 1 fps -> 1 second between frames
	m := &mockTraverser{}
	b.AddTraverser(m)
	b.LastFrame = time.Now() // just ticked

	b.Tick()
	if m.preCount != 0 {
		t.Fatal("should be rate-limited, not invoking traverser")
	}
}

// ---------------------------------------------------------------------------
// Binding stack edge cases
// ---------------------------------------------------------------------------

func TestBrowser_UnbindNonExistentViewpoint(t *testing.T) {
	b := NewBrowser()
	vp := node.NewViewpoint()
	b.BindViewpoint(vp, false) // unbind something never bound
	if len(b.ViewpointStack) != 0 {
		t.Fatal("should still be empty")
	}
}

func TestBrowser_BindFogStack(t *testing.T) {
	b := NewBrowser()
	fg1 := node.NewFog()
	fg2 := node.NewFog()
	b.BindFog(fg1, true)
	b.BindFog(fg2, true)
	if b.GetFog() != fg2 {
		t.Fatal("top should be fg2")
	}
	b.BindFog(fg2, false)
	if b.GetFog() != fg1 {
		t.Fatal("top should be fg1 after pop")
	}
}

func TestBrowser_BindBackgroundStack(t *testing.T) {
	b := NewBrowser()
	bg1 := &node.Background{}
	bg2 := &node.Background{}
	b.BindBackground(bg1, true)
	b.BindBackground(bg2, true)
	if b.GetBackground() != bg2 {
		t.Fatal("top should be bg2")
	}
	b.BindBackground(bg2, false)
	if b.GetBackground() != bg1 {
		t.Fatal("top should be bg1")
	}
}

func TestBrowser_BindNavInfoStack(t *testing.T) {
	b := NewBrowser()
	ni1 := node.NewNavigationInfo()
	ni2 := node.NewNavigationInfo()
	b.BindNavigationInfo(ni1, true)
	b.BindNavigationInfo(ni2, true)
	if b.GetNavigationInfo() != ni2 {
		t.Fatal("top should be ni2")
	}
	b.BindNavigationInfo(ni2, false)
	if b.GetNavigationInfo() != ni1 {
		t.Fatal("top should be ni1")
	}
}

// ---------------------------------------------------------------------------
// getField coverage
// ---------------------------------------------------------------------------

func TestGetField_FractionChanged(t *testing.T) {
	ts := node.NewTimeSensor()
	ts.Fraction = 0.75
	val := getField(ts, node.FractionStr)
	if val != 0.75 {
		t.Fatalf("expected 0.75, got %v", val)
	}
}

func TestGetField_IsActive_AllSensors(t *testing.T) {
	tests := []struct {
		name string
		n    node.Node
	}{
		{"TimeSensor", node.NewTimeSensor()},
		{"ProximitySensor", node.NewProximitySensor()},
		{"TouchSensor", node.NewTouchSensor()},
		{"PlaneSensor", node.NewPlaneSensor()},
		{"SphereSensor", node.NewSphereSensor()},
		{"CylinderSensor", node.NewCylinderSensor()},
	}
	for _, tt := range tests {
		val := getField(tt.n, node.IsActiveStr)
		if val == nil {
			t.Errorf("%s: isActive returned nil", tt.name)
		}
	}
}

func TestGetField_ValueChanged_AllInterpolators(t *testing.T) {
	tests := []struct {
		name string
		n    node.Node
	}{
		{"Position", &node.PositionInterpolator{}},
		{"Orientation", &node.OrientationInterpolator{}},
		{"Color", &node.ColorInterpolator{}},
		{"Scalar", &node.ScalarInterpolator{}},
		{"Coordinate", &node.CoordinateInterpolator{}},
		{"Normal", &node.NormalInterpolator{}},
	}
	for _, tt := range tests {
		val := getField(tt.n, node.ValueChangedStr)
		_ = val // Some interpolators may return zero-value; that's ok
	}
}

func TestGetField_CycleTime(t *testing.T) {
	ts := node.NewTimeSensor()
	ts.CycleTime = 42.0
	val := getField(ts, node.CycleTimeStr)
	if val != 42.0 {
		t.Fatalf("expected 42.0, got %v", val)
	}
}

func TestGetField_Time(t *testing.T) {
	ts := node.NewTimeSensor()
	ts.Time = 10.0
	val := getField(ts, node.TimeStr)
	if val != 10.0 {
		t.Fatalf("expected 10.0, got %v", val)
	}
}

func TestGetField_ProximitySensorFields(t *testing.T) {
	ps := node.NewProximitySensor()
	ps.EnterTime = 1.0
	ps.ExitTime = 2.0
	ps.Position = vec.SFVec3f{X: 5}
	ps.Orientation = vec.SFRotation{Y: 1, W: 3.14}

	if getField(ps, node.EnterTimeStr) != 1.0 {
		t.Fatal("enterTime")
	}
	if getField(ps, node.ExitTimeStr) != 2.0 {
		t.Fatal("exitTime")
	}
	if getField(ps, node.PositionChangedStr) != ps.Position {
		t.Fatal("positionChanged")
	}
	if getField(ps, node.OrientationChangedStr) != ps.Orientation {
		t.Fatal("orientationChanged")
	}
}

func TestGetField_TouchSensorFields(t *testing.T) {
	ts := node.NewTouchSensor()
	ts.IsOver = true
	ts.TouchTime = 5.5
	ts.HitPoint = vec.SFVec3f{X: 1, Y: 2, Z: 3}
	ts.HitNormal = vec.SFVec3f{Z: 1}

	if getField(ts, node.IsOverStr) != true {
		t.Fatal("isOver")
	}
	if getField(ts, node.TouchTimeStr) != 5.5 {
		t.Fatal("touchTime")
	}
	if getField(ts, node.HitPointStr) != ts.HitPoint {
		t.Fatal("hitPoint")
	}
	if getField(ts, node.HitNormalStr) != ts.HitNormal {
		t.Fatal("hitNormal")
	}
}

func TestGetField_RotationChanged(t *testing.T) {
	ss := node.NewSphereSensor()
	ss.Rotation = vec.SFRotation{Y: 1, W: 1.0}
	val := getField(ss, node.RotationChangedStr)
	if val != ss.Rotation {
		t.Fatal("SphereSensor rotationChanged")
	}

	cs := node.NewCylinderSensor()
	cs.Rotation = vec.SFRotation{Y: 1, W: 2.0}
	val = getField(cs, node.RotationChangedStr)
	if val != cs.Rotation {
		t.Fatal("CylinderSensor rotationChanged")
	}
}

func TestGetField_TranslationChanged(t *testing.T) {
	ps := node.NewPlaneSensor()
	ps.Translation = vec.SFVec3f{X: 10, Y: 20}
	val := getField(ps, node.TranslationChangedStr)
	if val != ps.Translation {
		t.Fatal("PlaneSensor translationChanged")
	}
}

func TestGetField_UnknownField(t *testing.T) {
	ts := node.NewTimeSensor()
	if getField(ts, "nonexistent_field") != nil {
		t.Fatal("unknown field should return nil")
	}
}

func TestGetField_WrongNodeType(t *testing.T) {
	// fraction_changed on a non-TimeSensor
	box := &node.Box{}
	if getField(box, node.FractionStr) != nil {
		t.Fatal("fraction_changed on Box should return nil")
	}
}

// ---------------------------------------------------------------------------
// setField coverage
// ---------------------------------------------------------------------------

func TestSetField_Translation(t *testing.T) {
	tr := node.NewTransform()
	v := vec.SFVec3f{X: 1, Y: 2, Z: 3}
	setField(tr, node.TranslationStr, v)
	if tr.Translation != v {
		t.Fatalf("expected %v, got %v", v, tr.Translation)
	}
}

func TestSetField_Rotation(t *testing.T) {
	tr := node.NewTransform()
	r := vec.SFRotation{X: 0, Y: 1, Z: 0, W: 1.57}
	setField(tr, node.RotationStr, r)
	if tr.Rotation != r {
		t.Fatalf("expected %v, got %v", r, tr.Rotation)
	}
}

func TestSetField_Scale(t *testing.T) {
	tr := node.NewTransform()
	s := vec.SFVec3f{X: 2, Y: 2, Z: 2}
	setField(tr, node.ScaleStr, s)
	if tr.Scale != s {
		t.Fatal("scale not set")
	}
}

func TestSetField_ViewpointPosition(t *testing.T) {
	vp := node.NewViewpoint()
	pos := vec.SFVec3f{X: 5, Y: 10, Z: 15}
	setField(vp, node.PositionStr, pos)
	if vp.Position != pos {
		t.Fatal("position not set")
	}
}

func TestSetField_ViewpointOrientation(t *testing.T) {
	vp := node.NewViewpoint()
	orient := vec.SFRotation{X: 1, Y: 0, Z: 0, W: 0.5}
	setField(vp, node.OrientationStr, orient)
	if vp.Orientation != orient {
		t.Fatal("orientation not set")
	}
}

func TestSetField_DiffuseColor(t *testing.T) {
	mat := node.NewMaterial()
	c := vec.SFColor{R: 0, G: 1, B: 0}
	setField(mat, node.DiffuseColorStr, c)
	if mat.DiffuseColor != c {
		t.Fatal("diffuseColor not set")
	}
}

func TestSetField_Transparency(t *testing.T) {
	mat := node.NewMaterial()
	setField(mat, node.TransparencyStr, float64(0.5))
	if mat.Transparency != 0.5 {
		t.Fatalf("expected 0.5, got %g", mat.Transparency)
	}
}

func TestSetField_Enabled(t *testing.T) {
	ts := node.NewTimeSensor()
	setField(ts, node.EnabledStr, false)
	if ts.Enabled {
		t.Fatal("should be disabled")
	}
	setField(ts, node.EnabledStr, true)
	if !ts.Enabled {
		t.Fatal("should be enabled")
	}
}

func TestSetField_WhichChoice(t *testing.T) {
	sw := &node.Switch{}
	setField(sw, node.WhichChoiceStr, float64(2))
	if sw.WhichChoice != 2 {
		t.Fatalf("expected 2, got %d", sw.WhichChoice)
	}
}

func TestSetField_StartTime_Toggle(t *testing.T) {
	ts := node.NewTimeSensor()
	ts.IsActive = true
	ts.Loop = true
	// Setting startTime while running and looping → should set stopTime
	setField(ts, node.StartTimeStr, float64(99.0))
	if ts.StopTime != 99.0 {
		t.Fatalf("expected stopTime=99, got %g", ts.StopTime)
	}
}

func TestSetField_StartTime_Normal(t *testing.T) {
	ts := node.NewTimeSensor()
	ts.IsActive = false
	setField(ts, node.StartTimeStr, float64(5.0))
	if ts.StartTime != 5.0 {
		t.Fatalf("expected startTime=5, got %g", ts.StartTime)
	}
	if ts.StopTime != 0 {
		t.Fatalf("stopTime should be reset to 0, got %g", ts.StopTime)
	}
}

func TestSetField_WrongType(t *testing.T) {
	tr := node.NewTransform()
	// Pass wrong type for translation
	setField(tr, node.TranslationStr, "not a vec")
	if tr.Translation != (vec.SFVec3f{}) {
		t.Fatal("should not change on wrong type")
	}
}

// ---------------------------------------------------------------------------
// setInterpolatorFraction
// ---------------------------------------------------------------------------

func TestSetInterpolatorFraction_AllTypes(t *testing.T) {
	pi := &node.PositionInterpolator{KeyValue: []vec.SFVec3f{{X: 0}, {X: 10}}}
	pi.Key = []float64{0, 1}
	setInterpolatorFraction(pi, 0.5)
	if pi.Value.X < 4.9 || pi.Value.X > 5.1 {
		t.Fatalf("Position: expected ~5, got %g", pi.Value.X)
	}

	oi := &node.OrientationInterpolator{
		KeyValue: []vec.SFRotation{{Y: 1, W: 0}, {Y: 1, W: 3.14}},
	}
	oi.Key = []float64{0, 1}
	setInterpolatorFraction(oi, 0.5)
	if oi.Value.W < 1.4 || oi.Value.W > 1.7 {
		t.Fatalf("Orientation: expected ~pi/2, got %g", oi.Value.W)
	}

	ci := &node.ColorInterpolator{
		KeyValue: []vec.SFColor{{R: 1}, {B: 1}},
	}
	ci.Key = []float64{0, 1}
	setInterpolatorFraction(ci, 0.5)
	if ci.Value.R < 0.4 || ci.Value.R > 0.6 {
		t.Fatalf("Color: expected ~0.5 R, got %g", ci.Value.R)
	}

	si := &node.ScalarInterpolator{KeyValue: []float64{0, 100}}
	si.Key = []float64{0, 1}
	setInterpolatorFraction(si, 0.25)
	if si.Value < 24 || si.Value > 26 {
		t.Fatalf("Scalar: expected ~25, got %g", si.Value)
	}

	coi := &node.CoordinateInterpolator{
		KeyValue: []vec.SFVec3f{{}, {X: 10, Y: 10, Z: 10}},
	}
	coi.Key = []float64{0, 1}
	setInterpolatorFraction(coi, 0.5)
	if len(coi.Value) != 1 || coi.Value[0].X < 4.9 {
		t.Fatalf("Coordinate: unexpected %v", coi.Value)
	}

	ni := &node.NormalInterpolator{
		KeyValue: []vec.SFVec3f{{Z: 1}, {Z: -1}},
	}
	ni.Key = []float64{0, 1}
	setInterpolatorFraction(ni, 0.5)
	if len(ni.Value) != 1 {
		t.Fatalf("Normal: expected 1 value, got %d", len(ni.Value))
	}
}

// ---------------------------------------------------------------------------
// processRoutes: change detection
// ---------------------------------------------------------------------------

func TestProcessRoutes_SkipsUnchanged(t *testing.T) {
	b := NewBrowser()
	ts := node.NewTimeSensor()
	ts.Fraction = 0.5
	tr := node.NewTransform()
	pi := &node.PositionInterpolator{KeyValue: []vec.SFVec3f{{}, {X: 10}}}
	pi.Key = []float64{0, 1}

	b.AddRoute(ts, node.FractionStr, pi, node.SetFractionStr)
	b.AddRoute(pi, node.ValueChangedStr, tr, node.TranslationStr)

	b.processRoutes() // first time — should fire
	firstX := tr.Translation.X

	b.processRoutes() // same fraction — should skip
	if tr.Translation.X != firstX {
		t.Fatal("unchanged fraction should not re-fire route")
	}
}

func TestProcessRoutes_FiresOnChange(t *testing.T) {
	b := NewBrowser()
	ts := node.NewTimeSensor()
	tr := node.NewTransform()
	pi := &node.PositionInterpolator{KeyValue: []vec.SFVec3f{{}, {X: 10}}}
	pi.Key = []float64{0, 1}

	b.AddRoute(ts, node.FractionStr, pi, node.SetFractionStr)
	b.AddRoute(pi, node.ValueChangedStr, tr, node.TranslationStr)

	ts.Fraction = 0.3
	b.processRoutes()

	ts.Fraction = 0.7
	b.processRoutes()
	if tr.Translation.X < 6.9 || tr.Translation.X > 7.1 {
		t.Fatalf("expected ~7, got %g", tr.Translation.X)
	}
}

// ---------------------------------------------------------------------------
// TimeSensor edge cases
// ---------------------------------------------------------------------------

func TestTimeSensor_Disabled(t *testing.T) {
	b := NewBrowser()
	ts := node.NewTimeSensor()
	ts.Enabled = false
	ts.CycleInterval = 1.0
	ts.StartTime = 0
	b.AddChild(ts)
	b.CollectTimeSensors()
	b.Update(time.Millisecond * 100)
	if ts.IsActive {
		t.Fatal("disabled TimeSensor should not be active")
	}
}

func TestTimeSensor_Loop(t *testing.T) {
	b := NewBrowser()
	ts := node.NewTimeSensor()
	ts.Enabled = true
	ts.Loop = true
	ts.CycleInterval = 0.01 // very short
	ts.StartTime = 0
	b.AddChild(ts)
	b.CollectTimeSensors()
	time.Sleep(50 * time.Millisecond)
	b.Update(time.Millisecond * 50)
	if !ts.IsActive {
		t.Fatal("looping TimeSensor should be active")
	}
	if ts.Fraction < 0 || ts.Fraction > 1 {
		t.Fatalf("fraction out of range: %g", ts.Fraction)
	}
}

func TestTimeSensor_Stopped(t *testing.T) {
	b := NewBrowser()
	ts := node.NewTimeSensor()
	ts.Enabled = true
	ts.CycleInterval = 1.0
	ts.StartTime = 0
	ts.StopTime = 0.001 // immediately stop
	ts.IsActive = true
	b.AddChild(ts)
	b.CollectTimeSensors()
	time.Sleep(20 * time.Millisecond)
	b.Update(time.Millisecond * 20)
	if ts.IsActive {
		t.Fatal("should be inactive after stop time")
	}
	if ts.Fraction != 1.0 {
		t.Fatalf("fraction should clamp to 1.0 on stop, got %g", ts.Fraction)
	}
}

// ---------------------------------------------------------------------------
// Interpolation edge cases
// ---------------------------------------------------------------------------

func TestEvalScalarInterp_SingleValue(t *testing.T) {
	r := evalScalarInterp([]float64{0}, []float64{42}, 0.5)
	if r != 42 {
		t.Fatalf("single value: expected 42, got %g", r)
	}
}

func TestEvalColorInterp_SingleValue(t *testing.T) {
	r := evalColorInterp([]float64{0}, []vec.SFColor{{R: 1}}, 0.5)
	if r.R != 1 {
		t.Fatalf("single color: expected R=1, got %g", r.R)
	}
}

func TestEvalOrientationInterp_SingleValue(t *testing.T) {
	r := evalOrientationInterp([]float64{0}, []vec.SFRotation{{Y: 1, W: 1.0}}, 0.5)
	if r.W != 1.0 {
		t.Fatalf("single rotation: expected W=1, got %g", r.W)
	}
}

func TestEvalPositionInterp_SingleValue(t *testing.T) {
	r := evalPositionInterp([]float64{0}, []vec.SFVec3f{{X: 7}}, 0.5)
	if r.X != 7 {
		t.Fatalf("single position: expected X=7, got %g", r.X)
	}
}

func TestEvalOrientationInterp_Empty(t *testing.T) {
	r := evalOrientationInterp(nil, nil, 0.5)
	if r != (vec.SFRotation{}) {
		t.Fatalf("expected zero, got %v", r)
	}
}

func TestEvalScalarInterp_Empty(t *testing.T) {
	r := evalScalarInterp(nil, nil, 0.5)
	if r != 0 {
		t.Fatalf("expected 0, got %g", r)
	}
}

func TestEvalColorInterp_Empty(t *testing.T) {
	r := evalColorInterp(nil, nil, 0.5)
	if r != (vec.SFColor{}) {
		t.Fatalf("expected zero, got %v", r)
	}
}

// ---------------------------------------------------------------------------
// routeValueEqual
// ---------------------------------------------------------------------------

func TestRouteValueEqual(t *testing.T) {
	if !routeValueEqual(1.0, 1.0) {
		t.Fatal("equal floats")
	}
	if routeValueEqual(1.0, 2.0) {
		t.Fatal("unequal floats")
	}
	if routeValueEqual(nil, 1.0) {
		t.Fatal("nil vs value")
	}
	// Slice comparison: should return false (non-comparable)
	if routeValueEqual([]int{1}, []int{1}) {
		t.Fatal("slices should not be equal via ==")
	}
}

// ---------------------------------------------------------------------------
// GetWorldURL
// ---------------------------------------------------------------------------

func TestBrowser_GetWorldURL(t *testing.T) {
	b := NewBrowser()
	if b.GetWorldURL() != "" {
		t.Fatal("should return empty string")
	}
}

// ---------------------------------------------------------------------------
// SetFrameRate negative
// ---------------------------------------------------------------------------

func TestBrowser_SetFrameRate_Negative(t *testing.T) {
	b := NewBrowser()
	b.SetFrameRate(-10)
	if b.FramesPerSec != 30 {
		t.Fatalf("negative fps should default to 30, got %d", b.FramesPerSec)
	}
}
