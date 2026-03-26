package traverser

import (
	"math"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ActionTraverser walks the scene graph each frame to process sensors,
// LOD distance switching, viewpoint binding, and other event-generating
// actions. This is the Go port of the C++ vrActionTraverser.
type ActionTraverser struct {
	Base

	// HasSensor is true after a traversal if any pointing-device sensors
	// were found in the scene graph. Used to decide whether mouse picking
	// is needed.
	HasSensor bool

	// ViewerPos is the current camera/viewpoint position in world space.
	// Set this before calling Traverse.
	ViewerPos vec.SFVec3f

	// SimTime is the current simulation time (seconds since start).
	SimTime float64

	// ProximitySensors found during scene collection.
	ProximitySensors []*node.ProximitySensor

	// TouchSensors found during collection.
	TouchSensors []*node.TouchSensor

	// LODs found during collection.
	LODs []*node.LOD
}

// NewActionTraverser creates an action traverser.
func NewActionTraverser() *ActionTraverser {
	return &ActionTraverser{Base: NewBase()}
}

// CollectSensors walks the scene graph once and gathers all action nodes.
func (at *ActionTraverser) CollectSensors(roots []node.Node) {
	at.ProximitySensors = nil
	at.TouchSensors = nil
	at.LODs = nil
	at.HasSensor = false
	at.collectFrom(roots)
}

func (at *ActionTraverser) collectFrom(nodes []node.Node) {
	for _, n := range nodes {
		switch v := n.(type) {
		case *node.ProximitySensor:
			at.ProximitySensors = append(at.ProximitySensors, v)
		case *node.TouchSensor:
			at.TouchSensors = append(at.TouchSensors, v)
			at.HasSensor = true
		case *node.CylinderSensor:
			at.HasSensor = true
		case *node.PlaneSensor:
			at.HasSensor = true
		case *node.SphereSensor:
			at.HasSensor = true
		case *node.Anchor:
			at.HasSensor = true
			at.collectFrom(v.Children)
		case *node.Transform:
			at.collectFrom(v.Children)
		case *node.Group:
			at.collectFrom(v.Children)
		case *node.Billboard:
			at.collectFrom(v.Children)
		case *node.Collision:
			at.collectFrom(v.Children)
		case *node.Inline:
			at.collectFrom(v.Children)
		case *node.Switch:
			for _, c := range v.Choice {
				at.collectFrom([]node.Node{c})
			}
		case *node.LOD:
			at.LODs = append(at.LODs, v)
			for _, l := range v.Level {
				at.collectFrom([]node.Node{l})
			}
		}
	}
}

// Update processes one frame of action traversal.
// Call this each frame before route processing.
func (at *ActionTraverser) Update(viewerPos vec.SFVec3f, simTime float64) {
	at.ViewerPos = viewerPos
	at.SimTime = simTime
	at.updateProximitySensors()
	at.updateLODs()
}

// updateProximitySensors checks each ProximitySensor against the viewer position.
func (at *ActionTraverser) updateProximitySensors() {
	for _, ps := range at.ProximitySensors {
		if !ps.Enabled {
			continue
		}
		center := ps.Center
		size := ps.Size
		diff := vec.SFVec3f{
			X: at.ViewerPos.X - center.X,
			Y: at.ViewerPos.Y - center.Y,
			Z: at.ViewerPos.Z - center.Z,
		}

		// VRML97 ProximitySensor: user is inside if within the axis-aligned box
		// defined by center ± size/2.
		isInside := math.Abs(float64(diff.X)) < float64(size.X)/2 &&
			math.Abs(float64(diff.Y)) < float64(size.Y)/2 &&
			math.Abs(float64(diff.Z)) < float64(size.Z)/2

		wasInside := ps.IsActive

		if wasInside && !isInside {
			ps.IsActive = false
			ps.ExitTime = at.SimTime
		} else if !wasInside && isInside {
			ps.IsActive = true
			ps.EnterTime = at.SimTime
			ps.Position = at.ViewerPos
		} else if isInside {
			// Continuously update position while inside.
			ps.Position = at.ViewerPos
		}
	}
}

// updateLODs selects the active level for each LOD based on distance.
func (at *ActionTraverser) updateLODs() {
	for _, lod := range at.LODs {
		if len(lod.Level) == 0 {
			continue
		}
		diff := vec.SFVec3f{
			X: at.ViewerPos.X - lod.Center.X,
			Y: at.ViewerPos.Y - lod.Center.Y,
			Z: at.ViewerPos.Z - lod.Center.Z,
		}
		dist := float64(math.Sqrt(float64(diff.X*diff.X + diff.Y*diff.Y + diff.Z*diff.Z)))

		selected := int64(len(lod.Level) - 1) // default: coarsest
		for i, r := range lod.Range {
			if dist < r {
				selected = int64(i)
				if selected >= int64(len(lod.Level)) {
					selected = int64(len(lod.Level) - 1)
				}
				break
			}
		}
		lod.ActiveLevel = selected
	}
}
