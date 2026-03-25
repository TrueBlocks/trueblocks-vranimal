# Traversers

Traversers are objects that "walk" the scene graph, maintaining state as they go. Each traverser performs a different operation on the scene graph. The pattern is:

```text
For each node in the scene graph:
    1. Pre-visit: push state (transforms, materials, etc.)
    2. Process: perform the traverser's specific operation
    3. Post-visit: pop state
```

## Go Packages: `pkg/traverser`, `pkg/converter`

In the Go port, g3n handles rendering internally, so the original OpenGL traverser is replaced by `pkg/converter` which converts the VRML scene graph into g3n scene objects.

## Traverser Types

### ActionTraverser

Fundamental to the execution model. At every frame, the ActionTraverser is the first traverser invoked:

1. Fires TimeSensor `fraction_changed` events
2. Converts mouse input to sensor events (TouchSensor, PlaneSensor, etc.)
3. Triggers event cascading along ROUTEs

The ActionTraverser is what drives animation. Without it, the scene is static.

**Status**: [Issue #10](https://github.com/TrueBlocks/trueblocks-3d/issues/10)

### RenderTraverser (base)

Base class for rendering traversers. Handles nodes common to all rendering (e.g., Switch child selection, LOD distance calculation). Specific rendering APIs derive from this.

### OGLTraverser (OpenGL)

Renders the scene graph to an OpenGL window. Each OGLTraverser has its own bound Viewpoint, supporting multiple views of the same scene.

In the Go port, g3n handles OpenGL rendering. The `pkg/converter` package translates VRML nodes to g3n objects, then g3n's internal renderer handles the draw calls.

**Status**: [Issue #6](https://github.com/TrueBlocks/trueblocks-3d/issues/6)

### D3DTraverser (Direct3D)

Placeholder for a Direct3D renderer. Was never implemented in the C++ version and is not needed in the Go port (g3n uses cross-platform OpenGL).

### WriteTraverser

Exports the scene graph back to a `.wrl` file. Ensures only non-default fields are written, reducing file size. Can serve as a pretty-printer.

**Status**: [Issue #9](https://github.com/TrueBlocks/trueblocks-3d/issues/9)

### ValidateTraverser

Walks the scene graph checking for invalid data: out-of-range field values, missing required fields, structural problems. Useful for debugging.

**Status**: [Issue #8](https://github.com/TrueBlocks/trueblocks-3d/issues/8)

### SerializeTraverser

Writes/reads the scene graph in a proprietary binary format for fast loading or content protection. Not implemented in C++.

**Status**: [Issue #7](https://github.com/TrueBlocks/trueblocks-3d/issues/7)

## The Converter (Go-specific)

The `pkg/converter` package replaces the OGLTraverser for the Go port. It walks the VRML node tree and creates equivalent g3n scene objects:

```go
func Convert(nodes []node.Node) *core.Node
```

Supported conversions:

| VRML Node | g3n Object |
|-----------|------------|
| Transform | `core.Node` with position/rotation/scale |
| Group | `core.Node` container |
| Shape + Box | `graphic.Mesh` with `geometry.NewBox` |
| Shape + Sphere | `graphic.Mesh` with `geometry.NewSphere` |
| Shape + Cone | `graphic.Mesh` with `geometry.NewCone` |
| Shape + Cylinder | `graphic.Mesh` with `geometry.NewCylinder` |
| Shape + IndexedFaceSet | `graphic.Mesh` with custom `geometry.Geometry` |
| Shape + ElevationGrid | `graphic.Mesh` with custom `geometry.Geometry` |
| Shape + Extrusion | `graphic.Mesh` with custom `geometry.Geometry` |
| Material | `material.Standard` |
| DirectionalLight | `light.Directional` |
| PointLight | `light.Point` |
| SpotLight | `light.Spot` |
| Switch | Selected child only |
| LOD | Distance-based child selection |

## ForEvery and FindBy

The scene graph supports two traversal utilities:

### ForEvery

Applies a user-defined function to every node in the scene graph. In Go:

```go
func ForEvery(root node.Node, fn func(node.Node) bool)
```

### FindByType / FindByName

Find the first node matching a type or name:

```go
func FindByType[T node.Node](root node.Node) T
func FindByName(root node.Node, name string) node.Node
```
