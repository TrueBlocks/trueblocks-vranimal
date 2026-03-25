package traverser

import (
	"math"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/experimental/collision"
	"github.com/g3n/engine/math32"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// Picker performs ray-casting mouse picking against the g3n scene and
// dispatches pointer events to VRML sensor nodes. This is the Go port
// of the C++ picking logic from browser/event.cpp.
type Picker struct {
	scene  *core.Node
	cam    *camera.Camera
	width  int
	height int

	// captured holds the sensor currently receiving drag events (mouse held).
	captured node.Node
	// capturedHit stores the initial hit point when drag started.
	capturedHit vec.SFVec3f

	// raycaster is reused across picks.
	raycaster *collision.Raycaster

	// SimTime is the current simulation time, set externally each frame.
	SimTime float64
}

// NewPicker creates a picker for the given g3n scene and camera.
func NewPicker(scene *core.Node, cam *camera.Camera) *Picker {
	origin := math32.Vector3{}
	dir := math32.Vector3{X: 0, Y: 0, Z: -1}
	return &Picker{
		scene:     scene,
		cam:       cam,
		raycaster: collision.NewRaycaster(&origin, &dir),
	}
}

// SetSize updates the viewport dimensions for coordinate normalization.
func (p *Picker) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// PointerAction indicates the type of mouse interaction.
type PointerAction int

const (
	PointerDown PointerAction = iota
	PointerUp
	PointerMove
)

// HandlePointer processes a mouse event at the given screen coordinates.
// Returns true if a sensor consumed the event.
func (p *Picker) HandlePointer(screenX, screenY float32, action PointerAction) bool {
	if p.width == 0 || p.height == 0 {
		return false
	}

	// Normalize screen coordinates to [-1, +1] range for g3n raycaster.
	// g3n's SetFromCamera expects normalized device coordinates.
	nx := (screenX/float32(p.width))*2 - 1
	ny := -((screenY/float32(p.height))*2 - 1) // Y is inverted

	// Set up raycaster from camera
	p.raycaster.SetFromCamera(p.cam, nx, ny)

	// If we have a captured sensor (drag in progress), handle it
	if p.captured != nil {
		if action == PointerUp {
			p.handleRelease()
			return true
		}
		if action == PointerMove {
			p.handleDrag(nx, ny)
			return true
		}
		return true
	}

	// Cast ray against scene
	intersects := p.raycaster.IntersectObject(p.scene, true)
	if len(intersects) == 0 {
		return false
	}

	// Find the VRML sensor group attached to the hit mesh
	hit := intersects[0]
	children := findSensorGroup(hit.Object)
	if children == nil {
		return false
	}

	hitPoint := vec.SFVec3f{X: hit.Point.X, Y: hit.Point.Y, Z: hit.Point.Z}

	// Compute a simple hit normal (approximate: direction from object center to hit)
	hitNormal := approximateNormal(hit)

	// Find sensors that are siblings of the hit geometry
	sensors := findSiblingSensors(children)
	if len(sensors) == 0 {
		return false
	}

	// Dispatch to each sensor
	consumed := false
	for _, sensor := range sensors {
		switch s := sensor.(type) {
		case *node.TouchSensor:
			consumed = p.handleTouchSensor(s, action, hitPoint, hitNormal) || consumed
		case *node.PlaneSensor:
			consumed = p.handlePlaneSensor(s, action, hitPoint) || consumed
		case *node.SphereSensor:
			consumed = p.handleSphereSensor(s, action, hitPoint) || consumed
		case *node.CylinderSensor:
			consumed = p.handleCylinderSensor(s, action, hitPoint) || consumed
		}
	}
	return consumed
}

// handleTouchSensor processes pointer events for a TouchSensor.
func (p *Picker) handleTouchSensor(ts *node.TouchSensor, action PointerAction, hitPoint, hitNormal vec.SFVec3f) bool {
	if !ts.Enabled {
		return false
	}

	ts.HitPoint = hitPoint
	ts.HitNormal = hitNormal

	switch action {
	case PointerDown:
		ts.IsActive = true
		ts.IsOver = true
		p.captured = ts
		p.capturedHit = hitPoint
	case PointerMove:
		ts.IsOver = true
	case PointerUp:
		ts.TouchTime = p.SimTime
		ts.IsActive = false
	}
	return true
}

// handleRelease ends any captured sensor drag.
func (p *Picker) handleRelease() {
	switch s := p.captured.(type) {
	case *node.TouchSensor:
		s.TouchTime = p.SimTime
		s.IsActive = false
		s.IsOver = false
	case *node.PlaneSensor:
		if s.AutoOffset {
			s.Offset = s.Translation
		}
		s.IsActive = false
	case *node.SphereSensor:
		if s.AutoOffset {
			s.Offset = s.Rotation
		}
		s.IsActive = false
	case *node.CylinderSensor:
		if s.AutoOffset {
			s.Offset = s.Rotation.W
		}
		s.IsActive = false
	}
	p.captured = nil
}

// handleDrag dispatches ongoing drag events to the captured sensor.
func (p *Picker) handleDrag(nx, ny float32) {
	switch s := p.captured.(type) {
	case *node.PlaneSensor:
		p.dragPlaneSensor(s, nx, ny)
	case *node.SphereSensor:
		p.dragSphereSensor(s, nx, ny)
	case *node.CylinderSensor:
		p.dragCylinderSensor(s, nx, ny)
	case *node.TouchSensor:
		// TouchSensor doesn't need drag updates
	}
}

// handlePlaneSensor processes initial pointer press for a PlaneSensor.
func (p *Picker) handlePlaneSensor(ps *node.PlaneSensor, action PointerAction, hitPoint vec.SFVec3f) bool {
	if !ps.Enabled || action != PointerDown {
		return false
	}
	ps.IsActive = true
	ps.FirstPoint = hitPoint
	ps.TrackPoint = hitPoint
	p.captured = ps
	p.capturedHit = hitPoint
	return true
}

// dragPlaneSensor updates translation during drag.
func (p *Picker) dragPlaneSensor(ps *node.PlaneSensor, nx, ny float32) {
	_, _ = nx, ny
	// Project current mouse pos onto the XY plane at the original hit depth
	ray := p.raycaster.Ray
	origin := ray.Origin()
	dir := ray.Direction()

	// Simple projection: use initial Z as the plane distance
	if dir.Z == 0 {
		return
	}
	t := (p.capturedHit.Z - origin.Z) / dir.Z
	if t < 0 {
		return
	}
	worldX := origin.X + dir.X*t
	worldY := origin.Y + dir.Y*t

	dx := worldX - ps.FirstPoint.X
	dy := worldY - ps.FirstPoint.Y

	trans := vec.SFVec3f{
		X: ps.Offset.X + dx,
		Y: ps.Offset.Y + dy,
		Z: ps.Offset.Z,
	}

	// Clamp to bounds if specified (negative means unconstrained)
	if ps.MinPosition.X <= ps.MaxPosition.X {
		trans.X = clampf(trans.X, ps.MinPosition.X, ps.MaxPosition.X)
	}
	if ps.MinPosition.Y <= ps.MaxPosition.Y {
		trans.Y = clampf(trans.Y, ps.MinPosition.Y, ps.MaxPosition.Y)
	}

	ps.Translation = trans
	ps.TrackPoint = vec.SFVec3f{X: worldX, Y: worldY, Z: p.capturedHit.Z}
}

// handleSphereSensor processes initial pointer press for a SphereSensor.
func (p *Picker) handleSphereSensor(ss *node.SphereSensor, action PointerAction, hitPoint vec.SFVec3f) bool {
	if !ss.Enabled || action != PointerDown {
		return false
	}
	ss.IsActive = true
	ss.FirstPoint = hitPoint
	ss.TrackPoint = hitPoint
	p.captured = ss
	p.capturedHit = hitPoint
	return true
}

// dragSphereSensor computes rotation from hemisphere projection.
func (p *Picker) dragSphereSensor(ss *node.SphereSensor, nx, ny float32) {
	// Project both initial and current points onto unit sphere
	first := projectOntoSphere(p.capturedHit)
	current := projectOntoSphere(vec.SFVec3f{X: nx, Y: ny, Z: 0})

	// Cross product gives rotation axis
	axis := vec.SFVec3f{
		X: first.Y*current.Z - first.Z*current.Y,
		Y: first.Z*current.X - first.X*current.Z,
		Z: first.X*current.Y - first.Y*current.X,
	}
	axisLen := float32(math.Sqrt(float64(axis.X*axis.X + axis.Y*axis.Y + axis.Z*axis.Z)))
	if axisLen < 1e-6 {
		return
	}
	axis.X /= axisLen
	axis.Y /= axisLen
	axis.Z /= axisLen

	// Dot product gives rotation angle
	dot := first.X*current.X + first.Y*current.Y + first.Z*current.Z
	dot = clampf(dot, -1, 1)
	angle := float32(math.Acos(float64(dot)))

	ss.Rotation = vec.SFRotation{X: axis.X, Y: axis.Y, Z: axis.Z, W: angle}
	ss.TrackPoint = vec.SFVec3f{X: nx, Y: ny, Z: 0}
}

// handleCylinderSensor processes initial pointer press for a CylinderSensor.
func (p *Picker) handleCylinderSensor(cs *node.CylinderSensor, action PointerAction, hitPoint vec.SFVec3f) bool {
	if !cs.Enabled || action != PointerDown {
		return false
	}
	cs.IsActive = true
	cs.FirstPoint = hitPoint
	cs.TrackPoint = hitPoint
	p.captured = cs
	p.capturedHit = hitPoint
	return true
}

// dragCylinderSensor computes rotation about Y axis from horizontal mouse motion.
func (p *Picker) dragCylinderSensor(cs *node.CylinderSensor, nx, ny float32) {
	// Map horizontal mouse displacement to Y-axis rotation
	dx := nx - p.capturedHit.X
	angle := cs.Offset + dx*float32(math.Pi)

	// Clamp if min < max
	if cs.MinAngle < cs.MaxAngle {
		angle = clampf(angle, cs.MinAngle, cs.MaxAngle)
	}

	cs.Rotation = vec.SFRotation{X: 0, Y: 1, Z: 0, W: angle}
	cs.TrackPoint = vec.SFVec3f{X: nx, Y: ny, Z: 0}
}

// findSensorGroup walks up the g3n node tree to find a parent whose UserData
// is a []node.Node (set by the converter for groups containing sensors).
func findSensorGroup(inode core.INode) []node.Node {
	n := inode.GetNode()
	for n != nil {
		if children, ok := n.UserData().([]node.Node); ok {
			return children
		}
		p := n.Parent()
		if p == nil {
			break
		}
		n = p.GetNode()
	}
	return nil
}

// findSiblingSensors returns all sensor nodes from a group's children.
func findSiblingSensors(children []node.Node) []node.Node {
	var sensors []node.Node
	for _, child := range children {
		switch child.(type) {
		case *node.TouchSensor, *node.PlaneSensor, *node.SphereSensor, *node.CylinderSensor:
			sensors = append(sensors, child)
		}
	}
	return sensors
}

// projectOntoSphere projects a point onto the unit sphere (trackball projection).
func projectOntoSphere(p vec.SFVec3f) vec.SFVec3f {
	d := p.X*p.X + p.Y*p.Y
	if d < 1.0 {
		return vec.SFVec3f{X: p.X, Y: p.Y, Z: float32(math.Sqrt(float64(1.0 - d)))}
	}
	s := float32(1.0 / math.Sqrt(float64(d)))
	return vec.SFVec3f{X: p.X * s, Y: p.Y * s, Z: 0}
}

// approximateNormal computes an approximate surface normal from the intersection.
func approximateNormal(hit collision.Intersect) vec.SFVec3f {
	n := hit.Object.GetNode()
	matWorld := n.MatrixWorld()
	var center math32.Vector3
	center.SetFromMatrixPosition(&matWorld)
	dir := math32.Vector3{
		X: hit.Point.X - center.X,
		Y: hit.Point.Y - center.Y,
		Z: hit.Point.Z - center.Z,
	}
	dir.Normalize()
	return vec.SFVec3f{X: dir.X, Y: dir.Y, Z: dir.Z}
}

func clampf(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
