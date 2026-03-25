package node

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// BaseNode
// ---------------------------------------------------------------------------

func TestBaseNode_Name(t *testing.T) {
	n := &BaseNode{}
	n.SetName("foo")
	if n.GetName() != "foo" {
		t.Fatalf("got %q", n.GetName())
	}
}

func TestBaseNode_NodeType(t *testing.T) {
	n := &BaseNode{NodeType: ProtoNode}
	if n.GetNodeType() != ProtoNode {
		t.Fatal("wrong type")
	}
}

func TestBaseNode_Routes(t *testing.T) {
	n := &BaseNode{}
	if n.HasRoutes() {
		t.Fatal("should have no routes")
	}
	r := &Route{}
	n.AddRoute(r)
	if !n.HasRoutes() {
		t.Fatal("should have routes")
	}
	n.RemoveRoute(r)
	if n.HasRoutes() {
		t.Fatal("should have no routes after remove")
	}
}

func TestBaseNode_RemoveRouteNotFound(t *testing.T) {
	n := &BaseNode{}
	n.AddRoute(&Route{})
	n.RemoveRoute(&Route{}) // different pointer, no-op
	if len(n.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(n.Routes))
	}
}

// ---------------------------------------------------------------------------
// Route
// ---------------------------------------------------------------------------

func TestNewRoute(t *testing.T) {
	src := &BaseNode{}
	dst := &BaseNode{}
	r := NewRoute(src, "fraction_changed", dst, "set_fraction")
	if r.Source != src || r.Destination != dst {
		t.Fatal("source/dest wrong")
	}
	if r.SrcField != "fraction_changed" || r.DstField != "set_fraction" {
		t.Fatal("field wrong")
	}
	if r.RouteID == 0 {
		t.Fatal("expected non-zero route ID")
	}
}

// ---------------------------------------------------------------------------
// GroupingNode
// ---------------------------------------------------------------------------

func TestGroupingNode_Children(t *testing.T) {
	g := NewGroupingNode()
	if g.HasChildren() {
		t.Fatal("should have no children")
	}
	g.AddChild(&BaseNode{})
	g.AddChild(&BaseNode{})
	if !g.HasChildren() {
		t.Fatal("should have children")
	}
	if len(g.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(g.Children))
	}
}

func TestGroupingNode_DefaultBbox(t *testing.T) {
	g := NewGroupingNode()
	if g.BboxSize.X != -1 {
		t.Fatalf("expected -1, got %g", g.BboxSize.X)
	}
}

// ---------------------------------------------------------------------------
// Material
// ---------------------------------------------------------------------------

func TestNewMaterial_Defaults(t *testing.T) {
	m := NewMaterial()
	if m.AmbientIntensity != 0.2 {
		t.Fatalf("ambient intensity = %g", m.AmbientIntensity)
	}
	if m.DiffuseColor != (vec.SFColor{R: 0.8, G: 0.8, B: 0.8, A: 1}) {
		t.Fatalf("diffuse = %v", m.DiffuseColor)
	}
	if m.Shininess != 0.2 {
		t.Fatalf("shininess = %g", m.Shininess)
	}
	if m.Transparency != 0 {
		t.Fatalf("transparency = %g", m.Transparency)
	}
}

func TestMaterial_AmbientColor(t *testing.T) {
	m := NewMaterial()
	ac := m.AmbientColor()
	expected := m.DiffuseColor.Scale(m.AmbientIntensity)
	if !ac.Eq(expected) {
		t.Fatalf("got %v, want %v", ac, expected)
	}
}

// ---------------------------------------------------------------------------
// Transform
// ---------------------------------------------------------------------------

func TestNewTransform_Defaults(t *testing.T) {
	tr := NewTransform()
	if tr.Scale != (vec.SFVec3f{X: 1, Y: 1, Z: 1}) {
		t.Fatalf("scale = %v", tr.Scale)
	}
	if tr.Translation != (vec.SFVec3f{}) {
		t.Fatalf("translation = %v", tr.Translation)
	}
	if tr.Rotation != (vec.SFRotation{X: 0, Y: 1, Z: 0, W: 0}) {
		t.Fatalf("rotation = %v", tr.Rotation)
	}
}

func TestTransform_GetLocalMatrix_Identity(t *testing.T) {
	tr := NewTransform()
	m := tr.GetLocalMatrix()
	id := vec.Identity()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			diff := m[i][j] - id[i][j]
			if diff > 1e-5 || diff < -1e-5 {
				t.Fatalf("not identity at [%d][%d]: %g", i, j, m[i][j])
			}
		}
	}
}

func TestTransform_GetLocalMatrix_Translation(t *testing.T) {
	tr := NewTransform()
	tr.Translation = vec.SFVec3f{X: 1, Y: 2, Z: 3}
	m := tr.GetLocalMatrix()
	p := m.TransformPoint(vec.SFVec3f{})
	if p.X != 1 || p.Y != 2 || p.Z != 3 {
		t.Fatalf("expected (1,2,3), got %v", p)
	}
}

// ---------------------------------------------------------------------------
// Geometry constructors
// ---------------------------------------------------------------------------

func TestNewBox_Defaults(t *testing.T) {
	b := NewBox()
	if b.Size != (vec.SFVec3f{X: 2, Y: 2, Z: 2}) {
		t.Fatalf("box size = %v", b.Size)
	}
}

func TestNewSphere_Defaults(t *testing.T) {
	s := NewSphere()
	if s.Radius != 1 {
		t.Fatalf("radius = %g", s.Radius)
	}
}

func TestNewCone_Defaults(t *testing.T) {
	c := NewCone()
	if c.BottomRadius != 1 || c.Height != 2 {
		t.Fatalf("cone = radius %g, height %g", c.BottomRadius, c.Height)
	}
}

func TestNewCylinder_Defaults(t *testing.T) {
	c := NewCylinder()
	if c.Radius != 1 || c.Height != 2 {
		t.Fatalf("cylinder = radius %g, height %g", c.Radius, c.Height)
	}
}

// ---------------------------------------------------------------------------
// Light constructors
// ---------------------------------------------------------------------------

func TestNewDirectionalLight_Defaults(t *testing.T) {
	l := NewDirectionalLight()
	if l.Intensity != 1 {
		t.Fatalf("intensity = %g", l.Intensity)
	}
	if l.Direction != (vec.SFVec3f{X: 0, Y: 0, Z: -1}) {
		t.Fatalf("direction = %v", l.Direction)
	}
}

func TestNewPointLight_Defaults(t *testing.T) {
	l := NewPointLight()
	if l.Radius != 100 {
		t.Fatalf("radius = %g", l.Radius)
	}
}

func TestNewSpotLight_Defaults(t *testing.T) {
	l := NewSpotLight()
	if l.CutOffAngle == 0 {
		t.Fatal("cutoff should not be 0")
	}
}

// ---------------------------------------------------------------------------
// Sensor constructors
// ---------------------------------------------------------------------------

func TestNewTimeSensor_Defaults(t *testing.T) {
	ts := NewTimeSensor()
	if !ts.Enabled {
		t.Fatal("should be enabled")
	}
	if ts.CycleInterval != 1 {
		t.Fatalf("expected 1, got %g", ts.CycleInterval)
	}
}

func TestNewTouchSensor_Defaults(t *testing.T) {
	s := NewTouchSensor()
	if !s.Enabled {
		t.Fatal("should be enabled")
	}
}

func TestNewProximitySensor_Defaults(t *testing.T) {
	s := NewProximitySensor()
	if !s.Enabled {
		t.Fatal("should be enabled")
	}
}

// ---------------------------------------------------------------------------
// Bindable nodes
// ---------------------------------------------------------------------------

func TestNewViewpoint_Defaults(t *testing.T) {
	vp := NewViewpoint()
	if vp.FieldOfView == 0 {
		t.Fatal("FOV should not be 0")
	}
	if vp.Position != (vec.SFVec3f{X: 0, Y: 0, Z: 10}) {
		t.Fatalf("position = %v", vp.Position)
	}
}

func TestNewNavigationInfo_Defaults(t *testing.T) {
	ni := NewNavigationInfo()
	if ni.Speed != 1 {
		t.Fatalf("speed = %g", ni.Speed)
	}
	if !ni.Headlight {
		t.Fatal("headlight should be true")
	}
}

func TestNewFog_Defaults(t *testing.T) {
	f := NewFog()
	if f.VisibilityRange != 0 {
		t.Fatalf("visibility = %g", f.VisibilityRange)
	}
	if f.FogType != "LINEAR" {
		t.Fatalf("type = %q", f.FogType)
	}
}

// ---------------------------------------------------------------------------
// LOD, Switch, Billboard, Collision
// ---------------------------------------------------------------------------

func TestNewLOD_Defaults(t *testing.T) {
	l := NewLOD()
	if l.ActiveLevel != -1 {
		t.Fatalf("active level = %d", l.ActiveLevel)
	}
}

func TestNewSwitch_Defaults(t *testing.T) {
	s := NewSwitch()
	if s.WhichChoice != -1 {
		t.Fatalf("which choice = %d", s.WhichChoice)
	}
}

func TestNewBillboard_Defaults(t *testing.T) {
	b := NewBillboard()
	if b.AxisOfRotation != (vec.SFVec3f{X: 0, Y: 1, Z: 0}) {
		t.Fatalf("axis = %v", b.AxisOfRotation)
	}
}

func TestNewCollision_Defaults(t *testing.T) {
	c := NewCollision()
	if !c.Collide {
		t.Fatal("collide should be true")
	}
}

// ---------------------------------------------------------------------------
// Texture constructors
// ---------------------------------------------------------------------------

func TestNewImageTexture_Defaults(t *testing.T) {
	it := NewImageTexture()
	if !it.RepeatS || !it.RepeatT {
		t.Fatal("repeat should be true")
	}
}

func TestNewPixelTexture_Defaults(t *testing.T) {
	pt := NewPixelTexture()
	if !pt.RepeatS || !pt.RepeatT {
		t.Fatal("repeat should be true")
	}
}

func TestNewTextureTransform_Defaults(t *testing.T) {
	tt := NewTextureTransform()
	if tt.Scale != (vec.SFVec2f{X: 1, Y: 1}) {
		t.Fatalf("scale = %v", tt.Scale)
	}
}

// ---------------------------------------------------------------------------
// FontStyle
// ---------------------------------------------------------------------------

func TestNewFontStyle_Defaults(t *testing.T) {
	fs := NewFontStyle()
	if fs.Size != 1 {
		t.Fatalf("size = %g", fs.Size)
	}
}
