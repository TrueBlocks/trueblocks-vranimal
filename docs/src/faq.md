# Frequently Asked Questions

## What is VRaniML?

VRaniML is a Go toolkit for reading, writing, and rendering VRML97 3D scene files. It's a port of the original C++ VRaniML SDK to idiomatic Go, using the g3n engine for rendering.

## What is VRML97?

VRML97 (Virtual Reality Modeling Language) is an ISO standard (ISO/IEC 14772-1:1997) for describing interactive 3D scenes. Files use the `.wrl` extension. It defines 54 node types covering geometry, appearance, animation, interaction, and environment.

## Why Go instead of C++?

- Cross-platform without build system complexity
- Garbage collection eliminates memory management bugs
- Go's interfaces replace C++ virtual functions and MFC macros
- Simpler dependency management via Go modules
- Faster compile times

## Do I need to understand the solid modeling library?

No. The half-edge data structure is completely internal. You work with high-level node types (Box, Sphere, IndexedFaceSet) and never touch the solid modeling layer directly.

## Can I use VRaniML without rendering?

Yes. The `pkg/node`, `pkg/parser`, `pkg/vec`, and `pkg/solid` packages have no rendering dependencies. Use them for parsing `.wrl` files, analyzing scene graphs, or generating VRML output.

## What VRML features are implemented?

The Go port covers the vast majority of VRML97, including:

- **Parsing**: Full VRML97 lexer/parser with DEF/USE, PROTO, EXTERNPROTO, Inline, and ROUTE support
- **Scene graph**: All 54 VRML97 node types defined with correct default values
- **Rendering**: g3n-based OpenGL viewer with geometry, materials, textures, lights, transforms
- **Solid modeling**: Half-edge B-rep data structure, Euler operators, primitive construction
- **Boolean operations**: Union, intersection, and difference on solid geometry
- **Split operations**: Plane-based solid splitting
- **Animation**: Interpolator evaluation (color, position, orientation, scalar, coordinate, normal)
- **Interaction**: Mouse picking, touch/proximity/cylinder/plane/sphere sensors
- **Event routing**: ROUTE evaluation, TimeSensor-driven animation loops
- **Traversers**: Write (pretty-print), validate, serialize, action, and picking traversers
- **Browser**: Event loop with frame timing, binding stacks, and traverser management
- **Texture loading**: ImageTexture, PixelTexture, and MovieTexture support
- **LOD & Switch**: Dynamic child selection based on distance and whichChoice

- **Audio**: Spatial audio playback via g3n's OpenAL backend (WAV, OGG)
- **Text rendering**: VRML Text node with FontStyle (TrueType font rendering)
- **Fog**: Linear and exponential fog parameters exposed for rendering
- **Background**: Sky/ground color gradients and skybox cubemaps
- **Script node**: Go callback handlers for event processing

## What VRML features are deferred?

All VRML97 node types are now implemented. No features remain deferred.

See the [Node Inventory](./reference/inventory.md) for complete status.

## What .wrl files work today?

Any `.wrl` file using these nodes will render correctly: Transform, Group, Shape, Appearance, Material, Box, Sphere, Cone, Cylinder, IndexedFaceSet, IndexedLineSet, PointSet, ElevationGrid, Extrusion, DirectionalLight, PointLight, SpotLight, Viewpoint, Background, Switch, LOD, Billboard, Anchor, Collision, Inline, ImageTexture, PixelTexture, TextureTransform, TimeSensor, TouchSensor, ProximitySensor, CylinderSensor, PlaneSensor, SphereSensor, ColorInterpolator, PositionInterpolator, OrientationInterpolator, ScalarInterpolator, CoordinateInterpolator, NormalInterpolator.

## How do I report bugs?

File an issue at [github.com/TrueBlocks/trueblocks-3d/issues](https://github.com/TrueBlocks/trueblocks-3d/issues).
