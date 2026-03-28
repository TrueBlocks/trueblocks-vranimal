package node

import (
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

const eps = 1e-5

func approx(a, b float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < eps
}

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

// ===========================================================================
// Gap-filling tests (issue #47)
// ===========================================================================

// ---------------------------------------------------------------------------
// Transform.GetLocalMatrix — combined transforms
// ---------------------------------------------------------------------------

func TestTransform_GetLocalMatrix_Rotation(t *testing.T) {
	tr := NewTransform()
	tr.Rotation = vec.NewRotation(0, 1, 0, math.Pi/2)
	m := tr.GetLocalMatrix()
	// Rotate X-axis 90° around Y → should become -Z
	p := m.TransformPoint(vec.XAxis)
	if !approx(p.X, 0) || !approx(p.Y, 0) || !approx(p.Z, -1) {
		t.Fatalf("Y-rot 90° of X: got %v, want (0,0,-1)", p)
	}
}

func TestTransform_GetLocalMatrix_Scale(t *testing.T) {
	tr := NewTransform()
	tr.Scale = vec.SFVec3f{X: 2, Y: 3, Z: 4}
	m := tr.GetLocalMatrix()
	p := m.TransformPoint(vec.SFVec3f{X: 1, Y: 1, Z: 1})
	if !approx(p.X, 2) || !approx(p.Y, 3) || !approx(p.Z, 4) {
		t.Fatalf("scale (2,3,4) of (1,1,1): got %v", p)
	}
}

func TestTransform_GetLocalMatrix_TranslateAndScale(t *testing.T) {
	tr := NewTransform()
	tr.Translation = vec.SFVec3f{X: 10, Y: 0, Z: 0}
	tr.Scale = vec.SFVec3f{X: 2, Y: 2, Z: 2}
	m := tr.GetLocalMatrix()
	p := m.TransformPoint(vec.SFVec3f{X: 1, Y: 0, Z: 0})
	// Row-vector: p * T * S → (1,0,0) * T(10) = (11,0,0) * S(2) = (22,0,0)
	if !approx(p.X, 22) || !approx(p.Y, 0) || !approx(p.Z, 0) {
		t.Fatalf("translate+scale: got %v, want (22,0,0)", p)
	}
}

func TestTransform_GetLocalMatrix_TranslateRotateScale(t *testing.T) {
	tr := NewTransform()
	tr.Translation = vec.SFVec3f{X: 5, Y: 0, Z: 0}
	tr.Rotation = vec.NewRotation(0, 0, 1, math.Pi/2) // 90° around Z
	tr.Scale = vec.SFVec3f{X: 2, Y: 2, Z: 2}
	m := tr.GetLocalMatrix()
	// Row-vector: p * T * R * S
	// (1,0,0) * T(5) = (6,0,0) * R(90° Z) = (0,6,0) * S(2) = (0,12,0)
	p := m.TransformPoint(vec.SFVec3f{X: 1, Y: 0, Z: 0})
	if !approx(p.X, 0) || !approx(p.Y, 12) || !approx(p.Z, 0) {
		t.Fatalf("T+R+S: got %v, want (0,12,0)", p)
	}
}

func TestTransform_GetLocalMatrix_WithCenter(t *testing.T) {
	tr := NewTransform()
	tr.Center = vec.SFVec3f{X: 1, Y: 0, Z: 0}
	tr.Rotation = vec.NewRotation(0, 0, 1, math.Pi/2) // 90° around Z
	m := tr.GetLocalMatrix()
	// Row-vector: p * C * R * C^-1
	// (0,0,0) * C(1,0,0) = (1,0,0) * R(90° Z) = (0,1,0) * C^-1(-1,0,0) = (-1,1,0)
	p := m.TransformPoint(vec.Vec3fZero)
	if !approx(p.X, -1) || !approx(p.Y, 1) || !approx(p.Z, 0) {
		t.Fatalf("center rotation: got %v, want (-1,1,0)", p)
	}
}

func TestTransform_GetLocalMatrix_IdentityTransform(t *testing.T) {
	tr := NewTransform()
	m := tr.GetLocalMatrix()
	id := vec.Identity()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if !approx(m[i][j], id[i][j]) {
				t.Fatalf("default transform not identity at [%d][%d]: %g", i, j, m[i][j])
			}
		}
	}
}

// ---------------------------------------------------------------------------
// BaseGeometry.GetSolid
// ---------------------------------------------------------------------------

func TestBaseGeometry_GetSolid_Nil(t *testing.T) {
	b := &BaseGeometry{}
	if b.GetSolid() != nil {
		t.Fatal("GetSolid should return nil by default")
	}
}

// ---------------------------------------------------------------------------
// Untested constructors — Extrusion
// ---------------------------------------------------------------------------

func TestNewExtrusion_Defaults(t *testing.T) {
	e := NewExtrusion()
	if !e.BeginCap || !e.EndCap {
		t.Fatalf("caps: begin=%v end=%v", e.BeginCap, e.EndCap)
	}
	if len(e.CrossSection) != 1 {
		t.Fatalf("CrossSection len=%d, want 1", len(e.CrossSection))
	}
	if len(e.Spine) != 1 {
		t.Fatalf("Spine len=%d, want 1", len(e.Spine))
	}
	if !e.Ccw || !e.Convex || !e.IsSolid {
		t.Fatalf("BaseGeometry defaults wrong: ccw=%v convex=%v solid=%v", e.Ccw, e.Convex, e.IsSolid)
	}
}

// ---------------------------------------------------------------------------
// Untested constructors — Sound / AudioClip
// ---------------------------------------------------------------------------

func TestNewSound_Defaults(t *testing.T) {
	s := NewSound()
	if !s.Spatialize {
		t.Fatal("Spatialize should be true")
	}
	if !approx(s.Intensity, 1.0) {
		t.Fatalf("Intensity=%g, want 1", s.Intensity)
	}
	if !approx(s.MaxBack, 10.0) || !approx(s.MaxFront, 10.0) {
		t.Fatalf("MaxBack=%g MaxFront=%g", s.MaxBack, s.MaxFront)
	}
	if !approx(s.MinBack, 1.0) || !approx(s.MinFront, 1.0) {
		t.Fatalf("MinBack=%g MinFront=%g", s.MinBack, s.MinFront)
	}
}

func TestNewAudioClip_Defaults(t *testing.T) {
	a := NewAudioClip()
	if !approx(a.Pitch, 1.0) {
		t.Fatalf("Pitch=%g, want 1", a.Pitch)
	}
}

// ---------------------------------------------------------------------------
// Untested constructors — MovieTexture
// ---------------------------------------------------------------------------

func TestNewMovieTexture_Defaults(t *testing.T) {
	m := NewMovieTexture()
	if !approx(m.Speed, 1.0) {
		t.Fatalf("Speed=%g, want 1", m.Speed)
	}
	if !m.RepeatS || !m.RepeatT {
		t.Fatalf("RepeatS=%v RepeatT=%v", m.RepeatS, m.RepeatT)
	}
}

// ---------------------------------------------------------------------------
// Untested constructors — DataSet-based geometry
// ---------------------------------------------------------------------------

func TestNewIndexedFaceSet_Defaults(t *testing.T) {
	ifs := NewIndexedFaceSet()
	if !ifs.ColorPerVertex {
		t.Fatal("ColorPerVertex should be true")
	}
	if !ifs.NormalPerVertex {
		t.Fatal("NormalPerVertex should be true")
	}
	if !ifs.Ccw || !ifs.Convex || !ifs.IsSolid {
		t.Fatalf("geometry defaults: ccw=%v convex=%v solid=%v", ifs.Ccw, ifs.Convex, ifs.IsSolid)
	}
}

func TestNewIndexedLineSet_Defaults(t *testing.T) {
	ils := NewIndexedLineSet()
	if !ils.ColorPerVertex || !ils.NormalPerVertex {
		t.Fatalf("ColorPerVertex=%v NormalPerVertex=%v", ils.ColorPerVertex, ils.NormalPerVertex)
	}
}

func TestNewPointSet_Defaults(t *testing.T) {
	ps := NewPointSet()
	if !ps.ColorPerVertex || !ps.NormalPerVertex {
		t.Fatalf("ColorPerVertex=%v NormalPerVertex=%v", ps.ColorPerVertex, ps.NormalPerVertex)
	}
}

func TestNewElevationGrid_Defaults(t *testing.T) {
	eg := NewElevationGrid()
	if !eg.ColorPerVertex || !eg.NormalPerVertex {
		t.Fatalf("ColorPerVertex=%v NormalPerVertex=%v", eg.ColorPerVertex, eg.NormalPerVertex)
	}
	if eg.XDimension != 0 || eg.ZDimension != 0 {
		t.Fatalf("dimensions: x=%d z=%d", eg.XDimension, eg.ZDimension)
	}
}

// ---------------------------------------------------------------------------
// Untested constructors — Sensors
// ---------------------------------------------------------------------------

func TestNewCylinderSensor_Defaults(t *testing.T) {
	cs := NewCylinderSensor()
	if !cs.Enabled {
		t.Fatal("Enabled should be true")
	}
	if !cs.AutoOffset {
		t.Fatal("AutoOffset should be true")
	}
	if !approx(cs.DiskAngle, 0.262) {
		t.Fatalf("DiskAngle=%g, want 0.262", cs.DiskAngle)
	}
	if !approx(cs.MaxAngle, -1.0) {
		t.Fatalf("MaxAngle=%g, want -1", cs.MaxAngle)
	}
}

func TestNewPlaneSensor_Defaults(t *testing.T) {
	ps := NewPlaneSensor()
	if !ps.Enabled {
		t.Fatal("Enabled should be true")
	}
	if !ps.AutoOffset {
		t.Fatal("AutoOffset should be true")
	}
	if !approx(ps.MaxPosition.X, -1) || !approx(ps.MaxPosition.Y, -1) {
		t.Fatalf("MaxPosition=%v, want (-1,-1)", ps.MaxPosition)
	}
}

func TestNewSphereSensor_Defaults(t *testing.T) {
	ss := NewSphereSensor()
	if !ss.Enabled {
		t.Fatal("Enabled should be true")
	}
	if !ss.AutoOffset {
		t.Fatal("AutoOffset should be true")
	}
	if !approx(ss.Offset.Y, 1) || !approx(ss.Offset.W, 0) {
		t.Fatalf("Offset=%v", ss.Offset)
	}
}

func TestNewVisibilitySensor_Defaults(t *testing.T) {
	vs := NewVisibilitySensor()
	if !vs.Enabled {
		t.Fatal("Enabled should be true")
	}
}
