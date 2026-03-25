package browser

import (
	"math"
	"time"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// Traverser is the interface for scene graph visitors (rendering, serialization, etc.).
type Traverser interface {
	PreTraverse()
	PostTraverse()
	TraverseNode(n node.Node)
	TraverseChildren(g *node.GroupingNode)
}

// UserData is the interface for application-specific data attached to nodes.
type UserData interface {
	GetID() string
}

// UserDataItem stores a user data instance with a string key.
type UserDataItem struct {
	ID   string
	Data UserData
}

// ---------------------------------------------------------------------------
// Browser - the VRML scene root
// ---------------------------------------------------------------------------

// Browser is the top-level scene graph container that manages traversers,
// binding stacks, navigation, and event processing.
type Browser struct {
	node.Transform
	Traversers   []Traverser
	PickedObject *node.GroupingNode

	FramesPerSec int
	FrameRate    time.Duration
	LastFrame    time.Time

	// Binding stacks
	ViewpointStack  []*node.Viewpoint
	NavInfoStack    []*node.NavigationInfo
	BackgroundStack []*node.Background
	FogStack        []*node.Fog

	// Event engine
	Routes       []*node.Route
	TimeSensors  []*node.TimeSensor
	startTime    time.Time // wall-clock origin for VRML time
	simTime      float64   // current simulation time in seconds

	// Route change-detection: only fire when source value changes.
	// Keyed by route index. VRML97 routes are event-driven.
	routePrev []any
}

// NewBrowser creates a browser with default settings.
func NewBrowser() *Browser {
	fps := 30
	now := time.Now()
	return &Browser{
		Transform:    *node.NewTransform(),
		FramesPerSec: fps,
		FrameRate:    time.Second / time.Duration(fps),
		LastFrame:    now,
		startTime:    now,
	}
}

// SetFrameRate configures the target frames per second.
func (b *Browser) SetFrameRate(fps int) {
	if fps <= 0 {
		fps = 30
	}
	b.FramesPerSec = fps
	b.FrameRate = time.Second / time.Duration(fps)
}

// AddTraverser appends a traverser to the browser.
func (b *Browser) AddTraverser(t Traverser) {
	b.Traversers = append(b.Traversers, t)
}

// GetTraverser returns the i-th traverser, or nil.
func (b *Browser) GetTraverser(i int) Traverser {
	if i < 0 || i >= len(b.Traversers) {
		return nil
	}
	return b.Traversers[i]
}

// NTraversers returns the number of traversers.
func (b *Browser) NTraversers() int { return len(b.Traversers) }

// ---------------------------------------------------------------------------
// Binding stack operations
// ---------------------------------------------------------------------------

// BindViewpoint pushes/pops a viewpoint onto the binding stack.
func (b *Browser) BindViewpoint(vp *node.Viewpoint, bind bool) {
	if bind {
		b.ViewpointStack = append(b.ViewpointStack, vp)
		vp.IsBound = true
	} else {
		for i, v := range b.ViewpointStack {
			if v == vp {
				b.ViewpointStack = append(b.ViewpointStack[:i], b.ViewpointStack[i+1:]...)
				vp.IsBound = false
				return
			}
		}
	}
}

// GetViewpoint returns the currently bound viewpoint, or nil.
func (b *Browser) GetViewpoint() *node.Viewpoint {
	if len(b.ViewpointStack) == 0 {
		return nil
	}
	return b.ViewpointStack[len(b.ViewpointStack)-1]
}

// BindNavigationInfo pushes/pops a navigation info.
func (b *Browser) BindNavigationInfo(ni *node.NavigationInfo, bind bool) {
	if bind {
		b.NavInfoStack = append(b.NavInfoStack, ni)
		ni.IsBound = true
	} else {
		for i, n := range b.NavInfoStack {
			if n == ni {
				b.NavInfoStack = append(b.NavInfoStack[:i], b.NavInfoStack[i+1:]...)
				ni.IsBound = false
				return
			}
		}
	}
}

// GetNavigationInfo returns the currently bound navigation info, or nil.
func (b *Browser) GetNavigationInfo() *node.NavigationInfo {
	if len(b.NavInfoStack) == 0 {
		return nil
	}
	return b.NavInfoStack[len(b.NavInfoStack)-1]
}

// BindBackground pushes/pops a background node.
func (b *Browser) BindBackground(bg *node.Background, bind bool) {
	if bind {
		b.BackgroundStack = append(b.BackgroundStack, bg)
		bg.IsBound = true
	} else {
		for i, bb := range b.BackgroundStack {
			if bb == bg {
				b.BackgroundStack = append(b.BackgroundStack[:i], b.BackgroundStack[i+1:]...)
				bg.IsBound = false
				return
			}
		}
	}
}

// GetBackground returns the currently bound background, or nil.
func (b *Browser) GetBackground() *node.Background {
	if len(b.BackgroundStack) == 0 {
		return nil
	}
	return b.BackgroundStack[len(b.BackgroundStack)-1]
}

// BindFog pushes/pops a fog node.
func (b *Browser) BindFog(fg *node.Fog, bind bool) {
	if bind {
		b.FogStack = append(b.FogStack, fg)
		fg.IsBound = true
	} else {
		for i, f := range b.FogStack {
			if f == fg {
				b.FogStack = append(b.FogStack[:i], b.FogStack[i+1:]...)
				fg.IsBound = false
				return
			}
		}
	}
}

// GetFog returns the currently bound fog, or nil.
func (b *Browser) GetFog() *node.Fog {
	if len(b.FogStack) == 0 {
		return nil
	}
	return b.FogStack[len(b.FogStack)-1]
}

// ---------------------------------------------------------------------------
// Scene management
// ---------------------------------------------------------------------------

// AddRoute creates and registers a route between two nodes.
func (b *Browser) AddRoute(src node.Node, srcField string, dst node.Node, dstField string) *node.Route {
	r := node.NewRoute(src, srcField, dst, dstField)
	b.Routes = append(b.Routes, r)
	return r
}

// Clear removes all children and resets the binding stacks.
func (b *Browser) Clear() {
	b.Children = nil
	b.ViewpointStack = nil
	b.NavInfoStack = nil
	b.BackgroundStack = nil
	b.FogStack = nil
	b.PickedObject = nil
	b.Routes = nil
	b.TimeSensors = nil
}

// Update advances the simulation clock and processes all events for one frame.
// Call this from the render loop with the elapsed time since last frame.
func (b *Browser) Update(dt time.Duration) {
	b.simTime = time.Since(b.startTime).Seconds()
	b.updateTimeSensors()
	b.processRoutes()
}

// SimTime returns the current simulation time in seconds.
func (b *Browser) SimTime() float64 {
	return b.simTime
}

// Tick advances the simulation by one frame (rate-limited).
func (b *Browser) Tick() {
	now := time.Now()
	if now.Sub(b.LastFrame) < b.FrameRate {
		return
	}
	dt := now.Sub(b.LastFrame)
	b.LastFrame = now
	b.Update(dt)

	for _, t := range b.Traversers {
		t.PreTraverse()
		b.traverseScene(t)
		t.PostTraverse()
	}
}

// ---------------------------------------------------------------------------
// Event engine
// ---------------------------------------------------------------------------

// CollectTimeSensors walks the scene graph and gathers all TimeSensor nodes.
func (b *Browser) CollectTimeSensors() {
	b.TimeSensors = nil
	collectTimeSensors(b.Children, &b.TimeSensors)
}

func collectTimeSensors(nodes []node.Node, out *[]*node.TimeSensor) {
	for _, n := range nodes {
		switch v := n.(type) {
		case *node.TimeSensor:
			*out = append(*out, v)
		case *node.Transform:
			collectTimeSensors(v.Children, out)
		case *node.Group:
			collectTimeSensors(v.Children, out)
		case *node.GroupingNode:
			collectTimeSensors(v.Children, out)
		}
	}
}

// updateTimeSensors ticks all TimeSensors, computing fraction and isActive.
func (b *Browser) updateTimeSensors() {
	now := b.simTime
	for _, ts := range b.TimeSensors {
		if !ts.Enabled {
			continue
		}
		start := ts.StartTime
		stop := ts.StopTime
		interval := ts.CycleInterval
		if interval <= 0 {
			interval = 1.0
		}

		// Not yet started
		if now < start {
			continue
		}

		// Stopped
		if stop > start && now >= stop {
			if ts.IsActive {
				ts.IsActive = false
				ts.Fraction = 1.0
				ts.Time = now
			}
			continue
		}

		elapsed := now - start
		if ts.Loop {
			ts.Fraction = float32(elapsed/interval - math.Floor(elapsed/interval))
			ts.IsActive = true
			ts.CycleTime = start + math.Floor(elapsed/interval)*interval
		} else {
			if elapsed >= interval {
				if ts.IsActive {
					ts.Fraction = 1.0
					ts.IsActive = false
				}
				continue
			}
			ts.Fraction = float32(elapsed / interval)
			ts.IsActive = true
		}
		ts.Time = now
	}
}

// processRoutes dispatches all routes, propagating field values.
// VRML97 routes are event-driven: only fire when the source value changes.
func (b *Browser) processRoutes() {
	if len(b.routePrev) != len(b.Routes) {
		b.routePrev = make([]any, len(b.Routes))
	}
	for i, r := range b.Routes {
		val := getField(r.Source, r.SrcField)
		if val == nil {
			continue
		}
		if routeValueEqual(val, b.routePrev[i]) {
			continue // no change, skip
		}
		b.routePrev[i] = val
		setField(r.Destination, r.DstField, val)
	}
}

// routeValueEqual compares two route values for equality.
// Handles non-comparable types (slices) by always treating them as changed.
func routeValueEqual(a, b any) bool {
	defer func() { recover() }()
	return a == b
}

// getField reads a named field from a VRML node.
func getField(n node.Node, field string) any {
	switch field {
	case node.FractionStr: // "fraction_changed"
		if ts, ok := n.(*node.TimeSensor); ok {
			return ts.Fraction
		}
	case node.IsActiveStr: // "isActive"
		switch v := n.(type) {
		case *node.TimeSensor:
			return v.IsActive
		case *node.ProximitySensor:
			return v.IsActive
		case *node.TouchSensor:
			return v.IsActive
		case *node.PlaneSensor:
			return v.IsActive
		case *node.SphereSensor:
			return v.IsActive
		case *node.CylinderSensor:
			return v.IsActive
		}
	case node.ValueChangedStr:
		return getInterpolatorValue(n)
	case node.CycleTimeStr:
		if ts, ok := n.(*node.TimeSensor); ok {
			return ts.CycleTime
		}
	case node.TimeStr:
		if ts, ok := n.(*node.TimeSensor); ok {
			return ts.Time
		}
	case node.EnterTimeStr:
		if ps, ok := n.(*node.ProximitySensor); ok {
			return ps.EnterTime
		}
	case node.ExitTimeStr:
		if ps, ok := n.(*node.ProximitySensor); ok {
			return ps.ExitTime
		}
	case node.PositionChangedStr:
		if ps, ok := n.(*node.ProximitySensor); ok {
			return ps.Position
		}
	case node.OrientationChangedStr:
		if ps, ok := n.(*node.ProximitySensor); ok {
			return ps.Orientation
		}
	case node.IsOverStr:
		if ts, ok := n.(*node.TouchSensor); ok {
			return ts.IsOver
		}
	case node.TouchTimeStr:
		if ts, ok := n.(*node.TouchSensor); ok {
			return ts.TouchTime
		}
	case node.HitPointStr:
		if ts, ok := n.(*node.TouchSensor); ok {
			return ts.HitPoint
		}
	case node.HitNormalStr:
		if ts, ok := n.(*node.TouchSensor); ok {
			return ts.HitNormal
		}
	case node.RotationChangedStr:
		switch v := n.(type) {
		case *node.SphereSensor:
			return v.Rotation
		case *node.CylinderSensor:
			return v.Rotation
		}
	case node.TranslationChangedStr:
		if ps, ok := n.(*node.PlaneSensor); ok {
			return ps.Translation
		}
	}
	return nil
}

// getInterpolatorValue returns the current output of an interpolator.
func getInterpolatorValue(n node.Node) any {
	switch v := n.(type) {
	case *node.PositionInterpolator:
		return v.Value
	case *node.OrientationInterpolator:
		return v.Value
	case *node.ColorInterpolator:
		return v.Value
	case *node.ScalarInterpolator:
		return v.Value
	case *node.CoordinateInterpolator:
		return v.Value
	case *node.NormalInterpolator:
		return v.Value
	}
	return nil
}

// setField writes a value to a named field on a VRML node, triggering evaluation.
func setField(n node.Node, field string, val any) {
	switch field {
	case node.SetFractionStr: // "set_fraction"
		if frac, ok := toFloat32(val); ok {
			setInterpolatorFraction(n, frac)
		}
	case node.TranslationStr, node.SetTranslationStr:
		if t, ok := n.(*node.Transform); ok {
			if v, ok := val.(vec.SFVec3f); ok {
				t.Translation = v
			}
		}
	case node.RotationStr, node.SetRotationStr:
		if t, ok := n.(*node.Transform); ok {
			if v, ok := val.(vec.SFRotation); ok {
				t.Rotation = v
			}
		}
	case node.ScaleStr, node.SetScaleStr:
		if t, ok := n.(*node.Transform); ok {
			if v, ok := val.(vec.SFVec3f); ok {
				t.Scale = v
			}
		}
	case node.PositionStr, node.SetPositionStr:
		if vp, ok := n.(*node.Viewpoint); ok {
			if v, ok := val.(vec.SFVec3f); ok {
				vp.Position = v
			}
		}
	case node.OrientationStr, node.SetOrientationStr:
		if vp, ok := n.(*node.Viewpoint); ok {
			if v, ok := val.(vec.SFRotation); ok {
				vp.Orientation = v
			}
		}
	case node.DiffuseColorStr, node.SetDiffuseColorStr:
		if m, ok := n.(*node.Material); ok {
			if v, ok := val.(vec.SFColor); ok {
				m.DiffuseColor = v
			}
		}
	case node.TransparencyStr:
		if m, ok := n.(*node.Material); ok {
			if v, ok := toFloat32(val); ok {
				m.Transparency = v
			}
		}
	case node.StartTimeStr:
		if ts, ok := n.(*node.TimeSensor); ok {
			if v, ok := toFloat64(val); ok {
				if ts.IsActive && ts.Loop {
					// Toggle: stop a running looping timer
					ts.StopTime = v
				} else {
					ts.StartTime = v
					ts.StopTime = 0 // reset so timer can start
				}
			}
		}
	case node.EnabledStr, node.SetEnabledStr:
		if b, ok := val.(bool); ok {
			switch s := n.(type) {
			case *node.TimeSensor:
				s.Enabled = b
			case *node.TouchSensor:
				s.Enabled = b
			case *node.ProximitySensor:
				s.Enabled = b
			case *node.PlaneSensor:
				s.Enabled = b
			case *node.SphereSensor:
				s.Enabled = b
			case *node.CylinderSensor:
				s.Enabled = b
			}
		}
	}
}

// setInterpolatorFraction sets the fraction and evaluates an interpolator.
func setInterpolatorFraction(n node.Node, frac float32) {
	switch v := n.(type) {
	case *node.PositionInterpolator:
		v.Fraction = frac
		v.Value = evalPositionInterp(v.Key, v.KeyValue, frac)
	case *node.OrientationInterpolator:
		v.Fraction = frac
		v.Value = evalOrientationInterp(v.Key, v.KeyValue, frac)
	case *node.ColorInterpolator:
		v.Fraction = frac
		v.Value = evalColorInterp(v.Key, v.KeyValue, frac)
	case *node.ScalarInterpolator:
		v.Fraction = frac
		v.Value = evalScalarInterp(v.Key, v.KeyValue, frac)
	case *node.CoordinateInterpolator:
		v.Fraction = frac
		v.Value = evalCoordinateInterp(v.Key, v.KeyValue, frac)
	case *node.NormalInterpolator:
		v.Fraction = frac
		v.Value = evalCoordinateInterp(v.Key, v.KeyValue, frac)
	}
}

func toFloat32(val any) (float32, bool) {
	switch v := val.(type) {
	case float32:
		return v, true
	case float64:
		return float32(v), true
	case bool:
		if v {
			return 1.0, true
		}
		return 0.0, true
	}
	return 0, false
}

func toFloat64(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case bool:
		if v {
			return 1.0, true
		}
		return 0.0, true
	}
	return 0, false
}

// ---------------------------------------------------------------------------
// Interpolation functions
// ---------------------------------------------------------------------------

// findKeySegment returns the segment index and local t for a given fraction.
func findKeySegment(keys []float32, frac float32) (int, float32) {
	if len(keys) == 0 {
		return 0, 0
	}
	if frac <= keys[0] {
		return 0, 0
	}
	if frac >= keys[len(keys)-1] {
		return len(keys) - 2, 1
	}
	for i := 1; i < len(keys); i++ {
		if frac <= keys[i] {
			span := keys[i] - keys[i-1]
			if span <= 0 {
				return i - 1, 0
			}
			t := (frac - keys[i-1]) / span
			return i - 1, t
		}
	}
	return len(keys) - 2, 1
}

func evalPositionInterp(keys []float32, values []vec.SFVec3f, frac float32) vec.SFVec3f {
	if len(values) == 0 {
		return vec.SFVec3f{}
	}
	if len(values) == 1 {
		return values[0]
	}
	seg, t := findKeySegment(keys, frac)
	if seg >= len(values)-1 {
		return values[len(values)-1]
	}
	a, b := values[seg], values[seg+1]
	return vec.SFVec3f{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
		Z: a.Z + (b.Z-a.Z)*t,
	}
}

func evalColorInterp(keys []float32, values []vec.SFColor, frac float32) vec.SFColor {
	if len(values) == 0 {
		return vec.SFColor{}
	}
	if len(values) == 1 {
		return values[0]
	}
	seg, t := findKeySegment(keys, frac)
	if seg >= len(values)-1 {
		return values[len(values)-1]
	}
	a, b := values[seg], values[seg+1]
	return vec.SFColor{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
	}
}

func evalScalarInterp(keys []float32, values []float32, frac float32) float32 {
	if len(values) == 0 {
		return 0
	}
	if len(values) == 1 {
		return values[0]
	}
	seg, t := findKeySegment(keys, frac)
	if seg >= len(values)-1 {
		return values[len(values)-1]
	}
	return values[seg] + (values[seg+1]-values[seg])*t
}

func evalOrientationInterp(keys []float32, values []vec.SFRotation, frac float32) vec.SFRotation {
	if len(values) == 0 {
		return vec.SFRotation{}
	}
	if len(values) == 1 {
		return values[0]
	}
	seg, t := findKeySegment(keys, frac)
	if seg >= len(values)-1 {
		return values[len(values)-1]
	}
	return vec.SlerpRotation(values[seg], values[seg+1], t)
}

func evalCoordinateInterp(keys []float32, values []vec.SFVec3f, frac float32) []vec.SFVec3f {
	if len(values) == 0 || len(keys) == 0 {
		return nil
	}
	nPerKey := len(values) / len(keys)
	if nPerKey == 0 {
		return nil
	}
	seg, t := findKeySegment(keys, frac)
	if seg >= len(keys)-1 {
		seg = len(keys) - 2
	}
	out := make([]vec.SFVec3f, nPerKey)
	baseA := seg * nPerKey
	baseB := (seg + 1) * nPerKey
	for i := 0; i < nPerKey; i++ {
		if baseA+i < len(values) && baseB+i < len(values) {
			a, b := values[baseA+i], values[baseB+i]
			out[i] = vec.SFVec3f{
				X: a.X + (b.X-a.X)*t,
				Y: a.Y + (b.Y-a.Y)*t,
				Z: a.Z + (b.Z-a.Z)*t,
			}
		}
	}
	return out
}

func (b *Browser) traverseScene(t Traverser) {
	for _, child := range b.Children {
		t.TraverseNode(child)
	}
}

// FindByName searches the scene graph for a node with the given DEF name.
func (b *Browser) FindByName(name string) node.Node {
	return findByNameInChildren(b.Children, name)
}

func findByNameInChildren(children []node.Node, name string) node.Node {
	for _, child := range children {
		if child.GetName() == name {
			return child
		}
		if g, ok := child.(*node.GroupingNode); ok {
			if found := findByNameInChildren(g.Children, name); found != nil {
				return found
			}
		}
		if t, ok := child.(*node.Transform); ok {
			if found := findByNameInChildren(t.Children, name); found != nil {
				return found
			}
		}
		if grp, ok := child.(*node.Group); ok {
			if found := findByNameInChildren(grp.Children, name); found != nil {
				return found
			}
		}
	}
	return nil
}

// GetSelection returns the currently picked/selected group.
func (b *Browser) GetSelection() *node.GroupingNode {
	return b.PickedObject
}

// SetSelection sets the picked/selected group.
func (b *Browser) SetSelection(g *node.GroupingNode) {
	b.PickedObject = g
}

// GetVersion returns the VRML specification version string.
func GetVersion() string {
	return "VRML 2.0 / VRML97"
}

// GetWorldURL returns the URL of the currently loaded world (placeholder).
func (b *Browser) GetWorldURL() string {
	return ""
}

// suppress unused import warning
var _ = vec.Identity
