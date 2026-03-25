# VRML97 Node Library

Each VRML97 node defined in the specification has a corresponding Go type. All nodes implement the `node.Node` interface.

## Go Package: `pkg/node`

```go
// Node is the interface implemented by all VRML97 nodes
type Node interface {
    NodeType() string
    GetName() string
    SetName(string)
}
```

## Node Categories

### Appearance Nodes

Control the visual look of geometry: materials, textures, and texture transforms.

- **Appearance** — pairs a Material with a Texture on a Shape
- **FontStyle** — font family, size, style for Text nodes
- **ImageTexture** — loads texture image from a URL
- **Material** — diffuse, specular, emissive colors, shininess, transparency
- **MovieTexture** — animated video texture
- **PixelTexture** — inline pixel data as texture
- **TextureTransform** — 2D texture coordinate transformation

### Bindable Nodes

Only one node of each bindable type can be active at a time. The browser maintains a stack for each type.

- **Background** — sky/ground colors and skybox textures
- **Fog** — linear or exponential distance fog
- **NavigationInfo** — navigation mode, speed, avatar size, headlight
- **Viewpoint** — camera position, orientation, field of view

### Common Nodes

General-purpose nodes for lighting, audio, scripting, and world metadata.

- **AudioClip** — loads and plays audio files
- **DirectionalLight** — infinite parallel light rays
- **PointLight** — omnidirectional point light source
- **Script** — embedded scripting for event processing
- **Shape** — pairs geometry with appearance
- **Sound** — spatialized 3D audio
- **SpotLight** — cone-shaped light with angular falloff
- **WorldInfo** — metadata (title, info strings)

### Geometry Nodes

Define 3D shapes. All geometry uses the solid modeling library (half-edge B-rep) as internal representation.

- **Box** — axis-aligned rectangular box
- **Color** — per-vertex/per-face color data
- **Cone** — cone with optional bottom cap
- **Coordinate** — shared vertex coordinate data
- **Cylinder** — cylinder with optional top/bottom caps
- **ElevationGrid** — height-field terrain
- **Extrusion** — 2D cross-section swept along a 3D spine
- **IndexedFaceSet** — arbitrary polygonal mesh
- **IndexedLineSet** — polylines
- **Normal** — per-vertex/per-face normal vectors
- **PointSet** — point cloud
- **Sphere** — geodesic sphere
- **Text** — 3D text string
- **TextureCoordinate** — UV texture mapping coordinates

### Grouping Nodes

Organize children into hierarchies with transforms.

- **Anchor** — hyperlink that loads a URL on click
- **Billboard** — auto-rotates children to face the camera
- **Collision** — enables/disables collision detection
- **Group** — groups children without transform
- **Inline** — loads external .wrl file by URL
- **LOD** — level of detail (picks child by distance)
- **Switch** — shows one child selected by `whichChoice`
- **Transform** — translation, rotation, scale of children

### Interpolator Nodes

Produce smooth value transitions driven by a `set_fraction` eventIn (0.0–1.0), typically connected to a TimeSensor via ROUTE.

- **ColorInterpolator** — interpolates between colors
- **CoordinateInterpolator** — interpolates vertex positions (morphing)
- **NormalInterpolator** — interpolates normal vectors
- **OrientationInterpolator** — interpolates rotations (slerp)
- **PositionInterpolator** — interpolates 3D positions
- **ScalarInterpolator** — interpolates a single float value

### Sensor Nodes

Detect user interaction and generate events.

- **CylinderSensor** — maps drag to rotation around an axis
- **PlaneSensor** — maps drag to 2D translation
- **ProximitySensor** — detects viewer entering/exiting a region
- **SphereSensor** — maps drag to spherical rotation
- **TimeSensor** — generates time-based fraction events (animation clock)
- **TouchSensor** — detects mouse click on geometry
- **VisibilitySensor** — detects when a region is visible

### Base Classes (non-VRML)

Abstract base types providing shared field access:

- **BaseNode** — name, reference count (Go: `Node` interface)
- **GroupingNode** — children field, bounding box
- **Bindable** — bind/unbind stack management
- **Light** — on, color, intensity, ambientIntensity
- **GeometryNode** — solid flag (for backface culling)
- **DataSet** — shared coordinate/color/normal/texCoord fields
- **Interpolator** — key, keyValue fields
- **Sensor** — enabled flag
