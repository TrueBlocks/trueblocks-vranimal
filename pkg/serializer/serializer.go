// Package serializer provides binary serialization and deserialization of
// VRML97 scene graphs using encoding/gob. This enables fast save/load of
// parsed .wrl files as a cache format.
package serializer

import (
	"encoding/gob"
	"fmt"
	"io"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func init() {
	// Register all concrete types that may appear behind interfaces
	// (node.Node, node.GeometryNode) so gob can encode/decode them.
	gob.Register(&node.BaseNode{})
	gob.Register(&node.GroupingNode{})
	gob.Register(&node.Appearance{})
	gob.Register(&node.Material{})
	gob.Register(&node.ImageTexture{})
	gob.Register(&node.MovieTexture{})
	gob.Register(&node.PixelTexture{})
	gob.Register(&node.TextureTransform{})
	gob.Register(&node.FontStyle{})
	gob.Register(&node.Background{})
	gob.Register(&node.Fog{})
	gob.Register(&node.NavigationInfo{})
	gob.Register(&node.Viewpoint{})
	gob.Register(&node.Shape{})
	gob.Register(&node.DirectionalLight{})
	gob.Register(&node.PointLight{})
	gob.Register(&node.SpotLight{})
	gob.Register(&node.WorldInfo{})
	gob.Register(&node.Script{})
	gob.Register(&node.Sound{})
	gob.Register(&node.AudioClip{})
	gob.Register(&node.Transform{})
	gob.Register(&node.Group{})
	gob.Register(&node.Anchor{})
	gob.Register(&node.Billboard{})
	gob.Register(&node.Collision{})
	gob.Register(&node.Inline{})
	gob.Register(&node.LOD{})
	gob.Register(&node.Switch{})
	gob.Register(&node.Box{})
	gob.Register(&node.Cone{})
	gob.Register(&node.Cylinder{})
	gob.Register(&node.Sphere{})
	gob.Register(&node.Extrusion{})
	gob.Register(&node.Text{})
	gob.Register(&node.IndexedFaceSet{})
	gob.Register(&node.IndexedLineSet{})
	gob.Register(&node.PointSet{})
	gob.Register(&node.ElevationGrid{})
	gob.Register(&node.ColorNode{})
	gob.Register(&node.Coordinate{})
	gob.Register(&node.NormalNode{})
	gob.Register(&node.TextureCoordinate{})
	gob.Register(&node.ColorInterpolator{})
	gob.Register(&node.CoordinateInterpolator{})
	gob.Register(&node.NormalInterpolator{})
	gob.Register(&node.OrientationInterpolator{})
	gob.Register(&node.PositionInterpolator{})
	gob.Register(&node.ScalarInterpolator{})
	gob.Register(&node.TouchSensor{})
	gob.Register(&node.TimeSensor{})
	gob.Register(&node.ProximitySensor{})
	gob.Register(&node.CylinderSensor{})
	gob.Register(&node.PlaneSensor{})
	gob.Register(&node.SphereSensor{})
	gob.Register(&node.VisibilitySensor{})
	gob.Register(&node.BaseGeometry{})
	gob.Register(&node.DataSet{})
	gob.Register(&node.Interpolator{})
	gob.Register(&node.Sensor{})
	gob.Register(&node.PointingDeviceSensor{})
	gob.Register(&node.Bindable{})
	gob.Register(&node.Light{})

	// Register vec types that appear in slices behind interfaces
	gob.Register(vec.SFVec2f{})
	gob.Register(vec.SFVec3f{})
	gob.Register(vec.SFColor{})
	gob.Register(vec.SFRotation{})
	gob.Register(vec.SFImage{})
}

const magic = "VRA1" // file magic for version identification

// sceneEnvelope wraps a scene for gob encoding.
// Exported fields are required for gob.
type sceneEnvelope struct {
	Nodes []node.Node
}

// Encode serializes a scene graph to binary format.
func Encode(w io.Writer, nodes []node.Node) error {
	if _, err := w.Write([]byte(magic)); err != nil {
		return fmt.Errorf("writing magic: %w", err)
	}
	env := sceneEnvelope{Nodes: nodes}
	if err := gob.NewEncoder(w).Encode(&env); err != nil {
		return fmt.Errorf("encoding scene: %w", err)
	}
	return nil
}

// Decode deserializes a scene graph from binary format.
func Decode(r io.Reader) ([]node.Node, error) {
	mag := make([]byte, len(magic))
	if _, err := io.ReadFull(r, mag); err != nil {
		return nil, fmt.Errorf("reading magic: %w", err)
	}
	if string(mag) != magic {
		return nil, fmt.Errorf("invalid file magic: got %q, want %q", string(mag), magic)
	}
	var env sceneEnvelope
	if err := gob.NewDecoder(r).Decode(&env); err != nil {
		return nil, fmt.Errorf("decoding scene: %w", err)
	}
	return env.Nodes, nil
}
