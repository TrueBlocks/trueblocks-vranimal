# Architecture Overview

The VRaniML SDK is composed of five components. Each component is briefly described here and further defined on the following pages.

```text
┌─────────────────────────────────────────────────────┐
│                   Traversers                         │
│    (walk the scene graph to render, export, etc.)    │
├─────────────────────────────────────────────────────┤
│               Browser (VEM)                          │
│  (parser, event routing, PROTO, clock management)    │
├─────────────────────────────────────────────────────┤
│             VRML97 Node Library                      │
│      (54 node types matching the VRML spec)          │
├─────────────────────────────────────────────────────┤
│            Solid Modeling Library                     │
│     (half-edge B-rep, boolean ops, split planes)     │
├─────────────────────────────────────────────────────┤
│        Utility Library (VRFC)                        │
│  (math types, containers, geometry, runtime support) │
└─────────────────────────────────────────────────────┘
```

## Components

### [Utility Library (VRFC)](./utility.md)

A collection of supporting classes and routines: collection types (lists, arrays, stacks), geometry types (vec2, vec3, rotation, matrix), and utility support (timer, image loading, runtime typing).

**Go packages**: `pkg/vec`, `pkg/types`, `pkg/geo`

### [Browser Related Classes (VEM)](./browser.md)

The VRML Execution Model. Supports event ROUTEing, scene graph management, PROTO/EXTERNPROTO handling, the VRML parser (lex/yacc based), and the internal clock.

**Go packages**: `pkg/browser`, `pkg/parser`

### [VRML97 Node Library](./nodes.md)

One Go type for each of the 54 VRML97 nodes. These types provide access to the data fields of each node. The node types are organized into categories: appearance, bindable, common, geometry, grouping, interpolator, and sensor nodes.

**Go package**: `pkg/node`

### [Solid Modeling Library](./solid.md)

All geometry nodes use the solid modeling library as their underlying data representation. The half-edge data structure enables boolean operations, mesh decimation, splitting planes, progressive meshes, and automatic level-of-detail generation.

**Go package**: `pkg/solid`

### [Traversers](./traversers.md)

Classes that "walk" the scene graph to perform operations: rendering to screen (via g3n/OpenGL), exporting to .wrl files, validating scene graph integrity, and processing user interaction events.

**Go packages**: `pkg/traverser`, `pkg/converter`
