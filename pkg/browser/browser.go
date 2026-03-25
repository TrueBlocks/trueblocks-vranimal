package browser

import (
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
}

// NewBrowser creates a browser with default settings.
func NewBrowser() *Browser {
	fps := 30
	return &Browser{
		Transform:    *node.NewTransform(),
		FramesPerSec: fps,
		FrameRate:    time.Second / time.Duration(fps),
		LastFrame:    time.Now(),
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
func (b *Browser) AddRoute(src node.Node, srcField int32, dst node.Node, dstField int32) *node.Route {
	r := node.NewRoute(src, srcField, dst, dstField)
	if bn, ok := src.(*node.BaseNode); ok {
		bn.AddRoute(r)
	}
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
}

// Tick advances the simulation by one frame, processing events and traversals.
func (b *Browser) Tick() {
	now := time.Now()
	if now.Sub(b.LastFrame) < b.FrameRate {
		return
	}
	b.LastFrame = now

	for _, t := range b.Traversers {
		t.PreTraverse()
		b.traverseScene(t)
		t.PostTraverse()
	}
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
