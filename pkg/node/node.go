package node

import (
	"math"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/geo"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// NodeType enumerates node classification.
type NodeType int

const (
	NormalNodeType NodeType = iota
	ProtoGroup
	ProtoNode
)

// ---------------------------------------------------------------------------
// Node - base interface and concrete base for all scene graph nodes.
// ---------------------------------------------------------------------------

// Node is the interface implemented by all VRML node types.
type Node interface {
	GetName() string
	SetName(name string)
	GetNodeType() NodeType
	GetBounds() geo.BoundingBox
}

// BaseNode holds fields common to all VRML nodes.
type BaseNode struct {
	Name       string
	Bounds     geo.BoundingBox
	NodeType   NodeType
	Routes     []*Route
	ProtoGroup string
	IsMaps     []string
}

func (n *BaseNode) GetName() string            { return n.Name }
func (n *BaseNode) SetName(name string)        { n.Name = name }
func (n *BaseNode) GetNodeType() NodeType      { return n.NodeType }
func (n *BaseNode) GetBounds() geo.BoundingBox { return n.Bounds }

// AddRoute adds a route originating from this node.
func (n *BaseNode) AddRoute(r *Route) {
	n.Routes = append(n.Routes, r)
}

// RemoveRoute removes a route from this node.
func (n *BaseNode) RemoveRoute(r *Route) {
	for i, rr := range n.Routes {
		if rr == r {
			n.Routes = append(n.Routes[:i], n.Routes[i+1:]...)
			return
		}
	}
}

// HasRoutes returns true if the node has outgoing routes.
func (n *BaseNode) HasRoutes() bool { return len(n.Routes) > 0 }

// ---------------------------------------------------------------------------
// Route - VRML ROUTE connection between node fields.
// ---------------------------------------------------------------------------

// Route connects an output field on a source to an input on a destination.
type Route struct {
	Source      Node
	SrcField    string
	Destination Node
	DstField    string
	Internal    bool
	RouteID     int64
}

var nextRouteID int64

// NewRoute creates a route between two node fields.
func NewRoute(src Node, srcField string, dst Node, dstField string) *Route {
	nextRouteID++
	return &Route{
		Source:      src,
		SrcField:    srcField,
		Destination: dst,
		DstField:    dstField,
		RouteID:     nextRouteID,
	}
}

// ---------------------------------------------------------------------------
// Event - data sent along a route
// ---------------------------------------------------------------------------

// Event carries a field value from one node to another.
type Event struct {
	Destination Node
	FieldID     int64
	Value       any
}

// ---------------------------------------------------------------------------
// GroupingNode - container for child nodes in the scene graph.
// ---------------------------------------------------------------------------

// GroupingNode holds a list of children and optional bounding box hints.
type GroupingNode struct {
	BaseNode
	Children   []Node
	BboxCenter vec.SFVec3f
	BboxSize   vec.SFVec3f
	HasSensor  bool
	HasLight   bool
}

// NewGroupingNode creates a grouping node with default values.
func NewGroupingNode() *GroupingNode {
	return &GroupingNode{
		BboxSize: vec.SFVec3f{X: -1, Y: -1, Z: -1},
	}
}

// AddChild appends a child node.
func (g *GroupingNode) AddChild(child Node) {
	g.Children = append(g.Children, child)
}

// HasChildren returns true if the group has children.
func (g *GroupingNode) HasChildren() bool { return len(g.Children) > 0 }

// ---------------------------------------------------------------------------
// Appearance nodes
// ---------------------------------------------------------------------------

// Appearance combines material, texture, and texture transform.
type Appearance struct {
	BaseNode
	Material         *Material
	Texture          Node
	TextureTransform *TextureTransform
}

// Material defines surface shading properties.
type Material struct {
	BaseNode
	AmbientIntensity float64
	DiffuseColor     vec.SFColor
	EmissiveColor    vec.SFColor
	Shininess        float64
	SpecularColor    vec.SFColor
	Transparency     float64
}

// NewMaterial creates a material with VRML97 defaults.
func NewMaterial() *Material {
	return &Material{
		AmbientIntensity: 0.2,
		DiffuseColor:     vec.SFColor{R: 0.8, G: 0.8, B: 0.8, A: 1},
		EmissiveColor:    vec.Black,
		Shininess:        0.2,
		SpecularColor:    vec.Black,
		Transparency:     0.0,
	}
}

// AmbientColor returns ambientIntensity * diffuseColor.
func (m *Material) AmbientColor() vec.SFColor {
	return m.DiffuseColor.Scale(m.AmbientIntensity)
}

// ImageTexture references an external image file as a texture.
type ImageTexture struct {
	BaseNode
	URL     []string
	OrigURL []string
	RepeatS bool
	RepeatT bool
}

// NewImageTexture creates an image texture with defaults.
func NewImageTexture() *ImageTexture {
	return &ImageTexture{RepeatS: true, RepeatT: true}
}

// MovieTexture references an external movie file as an animated texture.
type MovieTexture struct {
	BaseNode
	URL       []string
	OrigURL   []string
	Loop      bool
	Speed     float64
	StartTime float64
	StopTime  float64
	RepeatS   bool
	RepeatT   bool
	Duration  float64
	IsActive  bool
}

// NewMovieTexture creates a movie texture with defaults.
func NewMovieTexture() *MovieTexture {
	return &MovieTexture{Speed: 1.0, RepeatS: true, RepeatT: true}
}

// PixelTexture defines a texture from raw pixel data.
type PixelTexture struct {
	BaseNode
	Image   vec.SFImage
	RepeatS bool
	RepeatT bool
}

// NewPixelTexture creates a pixel texture with defaults.
func NewPixelTexture() *PixelTexture {
	return &PixelTexture{RepeatS: true, RepeatT: true}
}

// TextureTransform applies a 2D transformation to texture coordinates.
type TextureTransform struct {
	BaseNode
	Center      vec.SFVec2f
	Rotation    float64
	Scale       vec.SFVec2f
	Translation vec.SFVec2f
}

// NewTextureTransform creates a texture transform with defaults.
func NewTextureTransform() *TextureTransform {
	return &TextureTransform{
		Scale: vec.SFVec2f{X: 1, Y: 1},
	}
}

// FontStyle controls text rendering appearance.
type FontStyle struct {
	BaseNode
	Family      string
	Horizontal  bool
	Justify     []string
	Language    string
	LeftToRight bool
	Size        float64
	Spacing     float64
	Style       string
	TopToBottom bool
}

// NewFontStyle creates a font style with VRML97 defaults.
func NewFontStyle() *FontStyle {
	return &FontStyle{
		Family:      "SERIF",
		Horizontal:  true,
		Justify:     []string{"BEGIN"},
		LeftToRight: true,
		Size:        1.0,
		Spacing:     1.0,
		Style:       "PLAIN",
		TopToBottom: true,
	}
}

// ---------------------------------------------------------------------------
// Bindable nodes
// ---------------------------------------------------------------------------

// Bindable is the base for nodes that maintain a binding stack.
type Bindable struct {
	BaseNode
	IsBound bool
	Bind    bool
}

// Background defines the scene skybox/ground colors and images.
type Background struct {
	Bindable
	GroundAngle []float64
	GroundColor []vec.SFColor
	SkyAngle    []float64
	SkyColor    []vec.SFColor
	BackURL     []string
	BottomURL   []string
	FrontURL    []string
	LeftURL     []string
	RightURL    []string
	TopURL      []string
}

// Fog applies distance-based fog to the scene.
type Fog struct {
	Bindable
	Color           vec.SFColor
	FogType         string
	VisibilityRange float64
}

// NewFog creates a fog node with defaults.
func NewFog() *Fog {
	return &Fog{
		Color:   vec.White,
		FogType: "LINEAR",
	}
}

// NavigationInfo configures navigation behavior.
type NavigationInfo struct {
	Bindable
	AvatarSize      []float64
	Headlight       bool
	Speed           float64
	Type            []string
	VisibilityLimit float64
}

// NewNavigationInfo creates a navigation info with VRML97 defaults.
func NewNavigationInfo() *NavigationInfo {
	return &NavigationInfo{
		AvatarSize: []float64{0.25},
		Headlight:  true,
		Speed:      1.0,
		Type:       []string{"WALK"},
	}
}

// Viewpoint defines a camera position and orientation.
type Viewpoint struct {
	Bindable
	Description string
	FieldOfView float64
	Jump        bool
	Orientation vec.SFRotation
	Position    vec.SFVec3f
	BindTime    float64
}

// NewViewpoint creates a viewpoint with VRML97 defaults.
func NewViewpoint() *Viewpoint {
	return &Viewpoint{
		FieldOfView: float64(math.Pi / 4),
		Jump:        true,
		Orientation: vec.SFRotation{X: 0, Y: 1, Z: 0, W: 0},
		Position:    vec.SFVec3f{X: 0, Y: 0, Z: 10},
	}
}

// ---------------------------------------------------------------------------
// Common nodes
// ---------------------------------------------------------------------------

// Shape pairs an appearance with geometry for rendering.
type Shape struct {
	BaseNode
	Appearance *Appearance
	Geometry   GeometryNode
}

// Light is the base for all light source types.
type Light struct {
	BaseNode
	On               bool
	Color            vec.SFColor
	Intensity        float64
	AmbientIntensity float64
	Attenuation      vec.SFVec3f
	LightID          int64
}

func newLight() Light {
	return Light{
		On:          true,
		Color:       vec.White,
		Intensity:   1.0,
		Attenuation: vec.SFVec3f{X: 1, Y: 0, Z: 0},
		LightID:     -1,
	}
}

// DirectionalLight emits parallel rays in a given direction.
type DirectionalLight struct {
	Light
	Direction vec.SFVec3f
}

// NewDirectionalLight creates a directional light with defaults.
func NewDirectionalLight() *DirectionalLight {
	return &DirectionalLight{
		Light:     newLight(),
		Direction: vec.SFVec3f{X: 0, Y: 0, Z: -1},
	}
}

// PointLight emits light in all directions from a point.
type PointLight struct {
	Light
	Location vec.SFVec3f
	Radius   float64
}

// NewPointLight creates a point light with defaults.
func NewPointLight() *PointLight {
	return &PointLight{
		Light:  newLight(),
		Radius: 100.0,
	}
}

// SpotLight emits a cone of light.
type SpotLight struct {
	Light
	BeamWidth   float64
	CutOffAngle float64
	Direction   vec.SFVec3f
	Location    vec.SFVec3f
	Radius      float64
}

// NewSpotLight creates a spot light with defaults.
func NewSpotLight() *SpotLight {
	return &SpotLight{
		Light:       newLight(),
		BeamWidth:   float64(math.Pi / 2),
		CutOffAngle: float64(math.Pi / 4),
		Direction:   vec.SFVec3f{X: 0, Y: 0, Z: -1},
		Radius:      100.0,
	}
}

// WorldInfo provides descriptive metadata about the world.
type WorldInfo struct {
	BaseNode
	Info  []string
	Title string
}

// Script contains application logic.
type Script struct {
	BaseNode
	URL          []string
	OrigURL      []string
	DirectOutput bool
	MustEvaluate bool
}

// Sound defines spatial sound.
type Sound struct {
	BaseNode
	Spatialize bool
	Direction  vec.SFVec3f
	Intensity  float64
	Location   vec.SFVec3f
	MaxBack    float64
	MaxFront   float64
	MinBack    float64
	MinFront   float64
	Priority   float64
	Source     *AudioClip
}

// NewSound creates a sound with defaults.
func NewSound() *Sound {
	return &Sound{
		Spatialize: true,
		Direction:  vec.SFVec3f{X: 0, Y: 0, Z: 1},
		Intensity:  1.0,
		MaxBack:    10.0,
		MaxFront:   10.0,
		MinBack:    1.0,
		MinFront:   1.0,
	}
}

// AudioClip references an audio file.
type AudioClip struct {
	BaseNode
	URL         []string
	OrigURL     []string
	Description string
	Loop        bool
	Pitch       float64
	StartTime   float64
	StopTime    float64
	Duration    float64
	IsActive    bool
}

// NewAudioClip creates an audio clip with defaults.
func NewAudioClip() *AudioClip {
	return &AudioClip{Pitch: 1.0}
}

// ---------------------------------------------------------------------------
// Grouping nodes
// ---------------------------------------------------------------------------

// Transform applies translation, rotation, and scale to children.
type Transform struct {
	GroupingNode
	Center           vec.SFVec3f
	Rotation         vec.SFRotation
	Scale            vec.SFVec3f
	ScaleOrientation vec.SFRotation
	Translation      vec.SFVec3f
}

// NewTransform creates a transform with VRML97 defaults.
func NewTransform() *Transform {
	t := &Transform{}
	t.BboxSize = vec.SFVec3f{X: -1, Y: -1, Z: -1}
	t.Scale = vec.SFVec3f{X: 1, Y: 1, Z: 1}
	t.Rotation = vec.SFRotation{X: 0, Y: 1, Z: 0, W: 0}
	t.ScaleOrientation = vec.SFRotation{X: 0, Y: 1, Z: 0, W: 0}
	return t
}

// GetLocalMatrix computes the local transformation matrix.
func (t *Transform) GetLocalMatrix() vec.Matrix {
	m := vec.TranslationMatrix(t.Translation.X, t.Translation.Y, t.Translation.Z)
	m = m.Mul(vec.TranslationMatrix(t.Center.X, t.Center.Y, t.Center.Z))
	m = m.Mul(vec.RotationMatrix(t.Rotation))
	m = m.Mul(vec.RotationMatrix(t.ScaleOrientation))
	m = m.Mul(vec.ScaleMatrix(t.Scale.X, t.Scale.Y, t.Scale.Z))
	invSO := vec.SFRotation{X: t.ScaleOrientation.X, Y: t.ScaleOrientation.Y, Z: t.ScaleOrientation.Z, W: -t.ScaleOrientation.W}
	m = m.Mul(vec.RotationMatrix(invSO))
	m = m.Mul(vec.TranslationMatrix(-t.Center.X, -t.Center.Y, -t.Center.Z))
	return m
}

// Group is a generic grouping node with no transformation.
type Group struct {
	GroupingNode
}

// Anchor is a group that navigates on activation.
type Anchor struct {
	GroupingNode
	URL         []string
	OrigURL     []string
	Description string
	Parameter   []string
}

// Billboard auto-rotates children to face the viewer.
type Billboard struct {
	GroupingNode
	AxisOfRotation vec.SFVec3f
}

// NewBillboard creates a billboard with defaults.
func NewBillboard() *Billboard {
	b := &Billboard{}
	b.BboxSize = vec.SFVec3f{X: -1, Y: -1, Z: -1}
	b.AxisOfRotation = vec.SFVec3f{X: 0, Y: 1, Z: 0}
	return b
}

// Collision enables collision detection for its children.
type Collision struct {
	GroupingNode
	Collide     bool
	CollideTime float64
	Proxy       Node
}

// NewCollision creates a collision node with defaults.
func NewCollision() *Collision {
	c := &Collision{Collide: true}
	c.BboxSize = vec.SFVec3f{X: -1, Y: -1, Z: -1}
	return c
}

// Inline loads and displays content from external VRML files.
type Inline struct {
	GroupingNode
	URL     []string
	OrigURL []string
}

// LOD (Level of Detail) switches children based on distance from viewer.
type LOD struct {
	BaseNode
	Center      vec.SFVec3f
	Range       []float64
	Level       []Node
	ActiveLevel int64
}

// NewLOD creates a LOD with defaults.
func NewLOD() *LOD {
	return &LOD{ActiveLevel: -1}
}

// Switch displays one child at a time based on whichChoice.
type Switch struct {
	BaseNode
	Choice      []Node
	WhichChoice int64
}

// NewSwitch creates a switch with defaults.
func NewSwitch() *Switch {
	return &Switch{WhichChoice: -1}
}

// ---------------------------------------------------------------------------
// Geometry nodes
// ---------------------------------------------------------------------------

// GeometryNode is the interface for all renderable geometry.
type GeometryNode interface {
	Node
	GetSolid() *base.Solid
}

// BaseGeometry holds fields common to geometry nodes.
type BaseGeometry struct {
	BaseNode
	Ccw         bool
	Convex      bool
	CreaseAngle float64
	IsSolid     bool
	Geom        *base.Solid
	Native      bool
}

func newBaseGeometry() BaseGeometry {
	return BaseGeometry{
		Ccw:     true,
		Convex:  true,
		IsSolid: true,
		Native:  true,
	}
}

func (g *BaseGeometry) GetSolid() *base.Solid { return g.Geom }

// Box represents a rectangular parallelepiped.
type Box struct {
	BaseGeometry
	Size vec.SFVec3f
}

// NewBox creates a box with VRML97 default size (2,2,2).
func NewBox() *Box {
	b := &Box{BaseGeometry: newBaseGeometry()}
	b.Size = vec.SFVec3f{X: 2, Y: 2, Z: 2}
	return b
}

// Sphere represents a sphere centered at the origin.
type Sphere struct {
	BaseGeometry
	Radius float64
	Slices int64
	Stacks int64
}

// NewSphere creates a sphere with defaults.
func NewSphere() *Sphere {
	return &Sphere{
		BaseGeometry: newBaseGeometry(),
		Radius:       1.0,
		Slices:       16,
		Stacks:       16,
	}
}

// Cone represents a cone.
type Cone struct {
	BaseGeometry
	BottomRadius float64
	Height       float64
	Side         bool
	Bottom       bool
}

// NewCone creates a cone with defaults.
func NewCone() *Cone {
	return &Cone{
		BaseGeometry: newBaseGeometry(),
		BottomRadius: 1.0,
		Height:       2.0,
		Side:         true,
		Bottom:       true,
	}
}

// Cylinder represents a capped cylinder.
type Cylinder struct {
	BaseGeometry
	Bottom bool
	Height float64
	Radius float64
	Side   bool
	Top    bool
}

// NewCylinder creates a cylinder with defaults.
func NewCylinder() *Cylinder {
	return &Cylinder{
		BaseGeometry: newBaseGeometry(),
		Bottom:       true,
		Height:       2.0,
		Radius:       1.0,
		Side:         true,
		Top:          true,
	}
}

// Extrusion creates geometry by sweeping a 2D cross-section along a spine.
type Extrusion struct {
	BaseGeometry
	BeginCap     bool
	EndCap       bool
	CrossSection []vec.SFVec2f
	Orientation  []vec.SFRotation
	Scale        []vec.SFVec2f
	Spine        []vec.SFVec3f
}

// NewExtrusion creates an extrusion with defaults.
func NewExtrusion() *Extrusion {
	return &Extrusion{
		BaseGeometry: newBaseGeometry(),
		BeginCap:     true,
		EndCap:       true,
		CrossSection: []vec.SFVec2f{{X: 1, Y: 1}},
		Orientation:  []vec.SFRotation{{X: 0, Y: 1, Z: 0, W: 0}},
		Scale:        []vec.SFVec2f{{X: 1, Y: 1}},
		Spine:        []vec.SFVec3f{{X: 0, Y: 0, Z: 0}},
	}
}

// Text renders 3D text.
type Text struct {
	BaseGeometry
	String    []string
	FontStyle *FontStyle
	Length    []float64
	MaxExtent float64
}

// ---------------------------------------------------------------------------
// DataSet-based geometry (indexed face/line/point sets, elevation grid)
// ---------------------------------------------------------------------------

// DataSet is the base for geometry that uses indexed vertex arrays.
type DataSet struct {
	BaseGeometry
	ColorPerVertex  bool
	NormalPerVertex bool
	ColorIndex      []int64
	CoordIndex      []int64
	NormalIndex     []int64
	TexCoordIndex   []int64
	Color           *ColorNode
	Coord           *Coordinate
	Normal          *NormalNode
	TexCoord        *TextureCoordinate
}

func newDataSet() DataSet {
	return DataSet{
		BaseGeometry:    newBaseGeometry(),
		ColorPerVertex:  true,
		NormalPerVertex: true,
	}
}

// IndexedFaceSet defines geometry via indexed vertices forming polygonal faces.
type IndexedFaceSet struct {
	DataSet
}

// NewIndexedFaceSet creates an indexed face set with defaults.
func NewIndexedFaceSet() *IndexedFaceSet {
	return &IndexedFaceSet{DataSet: newDataSet()}
}

// IndexedLineSet defines geometry via indexed vertices forming polylines.
type IndexedLineSet struct {
	DataSet
}

// NewIndexedLineSet creates an indexed line set with defaults.
func NewIndexedLineSet() *IndexedLineSet {
	return &IndexedLineSet{DataSet: newDataSet()}
}

// PointSet defines geometry as a set of points.
type PointSet struct {
	DataSet
}

// NewPointSet creates a point set with defaults.
func NewPointSet() *PointSet {
	return &PointSet{DataSet: newDataSet()}
}

// ElevationGrid builds terrain from a height field.
type ElevationGrid struct {
	DataSet
	Heights    []float64
	XDimension int64
	XSpacing   float64
	ZDimension int64
	ZSpacing   float64
}

// NewElevationGrid creates an elevation grid with defaults.
func NewElevationGrid() *ElevationGrid {
	return &ElevationGrid{DataSet: newDataSet()}
}

// ---------------------------------------------------------------------------
// Geometry data nodes
// ---------------------------------------------------------------------------

// ColorNode provides per-vertex or per-face color values.
type ColorNode struct {
	BaseNode
	Color []vec.SFColor
}

// Coordinate provides vertex positions.
type Coordinate struct {
	BaseNode
	Point []vec.SFVec3f
}

// NormalNode provides normal vectors.
type NormalNode struct {
	BaseNode
	Vector []vec.SFVec3f
}

// TextureCoordinate provides texture mapping coordinates.
type TextureCoordinate struct {
	BaseNode
	Point []vec.SFVec2f
}

// ---------------------------------------------------------------------------
// Interpolator nodes
// ---------------------------------------------------------------------------

// Interpolator is the base for animation interpolation.
type Interpolator struct {
	BaseNode
	Key []float64
}

// ColorInterpolator interpolates between color values.
type ColorInterpolator struct {
	Interpolator
	KeyValue []vec.SFColor
	Value    vec.SFColor
	Fraction float64
}

// PositionInterpolator interpolates between 3D positions.
type PositionInterpolator struct {
	Interpolator
	KeyValue []vec.SFVec3f
	Value    vec.SFVec3f
	Fraction float64
}

// OrientationInterpolator interpolates between rotations.
type OrientationInterpolator struct {
	Interpolator
	KeyValue []vec.SFRotation
	Value    vec.SFRotation
	Fraction float64
}

// ScalarInterpolator interpolates between float values.
type ScalarInterpolator struct {
	Interpolator
	KeyValue []float64
	Value    float64
	Fraction float64
}

// CoordinateInterpolator interpolates between sets of coordinates.
type CoordinateInterpolator struct {
	Interpolator
	KeyValue []vec.SFVec3f
	Value    []vec.SFVec3f
	Fraction float64
}

// NormalInterpolator interpolates between sets of normals.
type NormalInterpolator struct {
	Interpolator
	KeyValue []vec.SFVec3f
	Value    []vec.SFVec3f
	Fraction float64
}

// ---------------------------------------------------------------------------
// Sensor nodes
// ---------------------------------------------------------------------------

// Sensor is the base for all event-generating sensor nodes.
type Sensor struct {
	BaseNode
	Enabled  bool
	IsActive bool
}

func newSensor() Sensor {
	return Sensor{Enabled: true}
}

// PointingDeviceSensor is the base for mouse-interaction sensors.
type PointingDeviceSensor struct {
	Sensor
	FirstPoint vec.SFVec3f
	AutoOffset bool
	TrackPoint vec.SFVec3f
}

func newPointingDeviceSensor() PointingDeviceSensor {
	return PointingDeviceSensor{
		Sensor:     newSensor(),
		AutoOffset: true,
	}
}

// TouchSensor detects when the user clicks on geometry.
type TouchSensor struct {
	PointingDeviceSensor
	HitNormal   vec.SFVec3f
	HitPoint    vec.SFVec3f
	HitTexCoord vec.SFVec2f
	IsOver      bool
	TouchTime   float64
}

// NewTouchSensor creates a touch sensor with defaults.
func NewTouchSensor() *TouchSensor {
	return &TouchSensor{PointingDeviceSensor: newPointingDeviceSensor()}
}

// TimeSensor generates time-based events.
type TimeSensor struct {
	Sensor
	CycleInterval float64
	Loop          bool
	StartTime     float64
	StopTime      float64
	CycleTime     float64
	Fraction      float64
	Time          float64
}

// NewTimeSensor creates a time sensor with defaults.
func NewTimeSensor() *TimeSensor {
	return &TimeSensor{
		Sensor:        newSensor(),
		CycleInterval: 1.0,
	}
}

// ProximitySensor detects user position within a region.
type ProximitySensor struct {
	Sensor
	Center      vec.SFVec3f
	Size        vec.SFVec3f
	Position    vec.SFVec3f
	Orientation vec.SFRotation
	EnterTime   float64
	ExitTime    float64
}

// NewProximitySensor creates a proximity sensor with defaults.
func NewProximitySensor() *ProximitySensor {
	return &ProximitySensor{Sensor: newSensor()}
}

// CylinderSensor maps mouse motion to a rotation about an axis.
type CylinderSensor struct {
	PointingDeviceSensor
	Offset    float64
	DiskAngle float64
	MaxAngle  float64
	MinAngle  float64
	Rotation  vec.SFRotation
}

// NewCylinderSensor creates a cylinder sensor with defaults.
func NewCylinderSensor() *CylinderSensor {
	return &CylinderSensor{
		PointingDeviceSensor: newPointingDeviceSensor(),
		DiskAngle:            0.262,
		MaxAngle:             -1.0,
	}
}

// PlaneSensor maps mouse motion to a 2D translation.
type PlaneSensor struct {
	PointingDeviceSensor
	Offset      vec.SFVec3f
	MaxPosition vec.SFVec2f
	MinPosition vec.SFVec2f
	Translation vec.SFVec3f
}

// NewPlaneSensor creates a plane sensor with defaults.
func NewPlaneSensor() *PlaneSensor {
	return &PlaneSensor{
		PointingDeviceSensor: newPointingDeviceSensor(),
		MaxPosition:          vec.SFVec2f{X: -1, Y: -1},
	}
}

// SphereSensor maps mouse motion to a rotation.
type SphereSensor struct {
	PointingDeviceSensor
	Offset   vec.SFRotation
	Rotation vec.SFRotation
}

// NewSphereSensor creates a sphere sensor with defaults.
func NewSphereSensor() *SphereSensor {
	return &SphereSensor{
		PointingDeviceSensor: newPointingDeviceSensor(),
		Offset:               vec.SFRotation{X: 0, Y: 1, Z: 0, W: 0},
	}
}

// VisibilitySensor detects if a region is visible.
type VisibilitySensor struct {
	Sensor
	Center    vec.SFVec3f
	Size      vec.SFVec3f
	EnterTime float64
	ExitTime  float64
}

// NewVisibilitySensor creates a visibility sensor with defaults.
func NewVisibilitySensor() *VisibilitySensor {
	return &VisibilitySensor{Sensor: newSensor()}
}

// ensure interfaces are satisfied
var _ Node = (*BaseNode)(nil)
var _ GeometryNode = (*BaseGeometry)(nil)
