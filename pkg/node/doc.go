// Package node defines the complete VRML97 node type hierarchy used to
// represent parsed scene graphs. Every VRML97 node type that the parser can
// produce has a corresponding Go struct here, along with constructors that
// set the VRML97 default field values.
//
// The hierarchy is organized as:
//
//   - Node interface and BaseNode — common identity, naming (DEF/USE), and
//     user-data attachment points.
//   - GroupingNode — shared children list and bounding-box for Group,
//     Transform, Anchor, Billboard, Collision, Inline, LOD, and Switch.
//   - GeometryNode interface and BaseGeometry — shared flag for Box, Sphere,
//     Cone, Cylinder, Extrusion, IndexedFaceSet, IndexedLineSet, PointSet,
//     ElevationGrid, and Text.
//   - Light — base for DirectionalLight, PointLight, SpotLight.
//   - Interpolator — base for Color, Position, Orientation, Scalar,
//     Coordinate, and Normal interpolators.
//   - Sensor — base for TimeSensor, TouchSensor, ProximitySensor,
//     CylinderSensor, PlaneSensor, SphereSensor.
//   - Appearance, Material, textures, Background, Fog, NavigationInfo,
//     Viewpoint, WorldInfo, Script, Sound, AudioClip.
//
// Route and Event capture the VRML97 event-routing model.
//
// This package has no dependencies on rendering or file I/O; it is a pure
// data-model layer.
package node
