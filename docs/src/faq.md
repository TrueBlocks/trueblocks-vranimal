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

## What VRML features are not yet implemented?

- PROTO / EXTERNPROTO instantiation (Issues [#11](https://github.com/TrueBlocks/trueblocks-3d/issues/11), [#12](https://github.com/TrueBlocks/trueblocks-3d/issues/12))
- ROUTE evaluation at runtime ([#13](https://github.com/TrueBlocks/trueblocks-3d/issues/13))
- Animation (interpolators + time sensors) ([#16](https://github.com/TrueBlocks/trueblocks-3d/issues/16))
- Mouse picking / sensors ([#15](https://github.com/TrueBlocks/trueblocks-3d/issues/15))
- Texture loading ([#17](https://github.com/TrueBlocks/trueblocks-3d/issues/17))
- Audio ([#18](https://github.com/TrueBlocks/trueblocks-3d/issues/18))
- Text rendering ([#20](https://github.com/TrueBlocks/trueblocks-3d/issues/20))
- Boolean operations ([#1](https://github.com/TrueBlocks/trueblocks-3d/issues/1))

See the [Node Inventory](./reference/inventory.md) for complete status.

## What .wrl files work today?

Any `.wrl` file using these nodes will render correctly: Transform, Group, Shape, Appearance, Material, Box, Sphere, Cone, Cylinder, IndexedFaceSet, ElevationGrid, Extrusion, DirectionalLight, PointLight, SpotLight, Viewpoint, Background, Switch, LOD.

## How do I report bugs?

File an issue at [github.com/TrueBlocks/trueblocks-3d/issues](https://github.com/TrueBlocks/trueblocks-3d/issues).
