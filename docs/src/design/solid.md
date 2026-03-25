# Solid Modeling Library

The solid modeling library implements a data structure called the **half-edge** (or **boundary representation** — B-rep). All geometry nodes (Box, Sphere, IndexedFaceSet, etc.) can use this data structure as their internal representation. This enables advanced geometric algorithms.

## Go Package: `pkg/solid`

## Why Half-Edge?

The half-edge data structure enables efficient traversal of mesh topology:

- Find all faces sharing a vertex — **O(degree)**
- Find all edges of a face — **O(face size)**
- Find neighboring faces — **O(1)** per edge
- Supports non-manifold geometry, faces with holes

This makes the following algorithms feasible:

| Algorithm | Description | Status |
|-----------|-------------|--------|
| **Boolean Operations** | Union, intersection, difference of two solids | [Issue #1](https://github.com/TrueBlocks/trueblocks-3d/issues/1) |
| **Splitting Planes** | Cut a solid along a plane into two halves | [Issue #2](https://github.com/TrueBlocks/trueblocks-3d/issues/2) |
| **Progressive Mesh** | Simplify geometry for transmission/LOD | Future |
| **Mesh Decimation** | Reduce polygon count automatically | Future |
| **Extrusions** | Sweep a 2D profile along a 3D spine | Ported |
| **Rotational Sweeps** | Revolve a 2D profile around an axis | Future |

## Data Structure

```text
Solid
├── Faces[]
│   ├── Loop (outer boundary)
│   │   └── HalfEdge → HalfEdge → HalfEdge → ...
│   └── Loop[] (holes, if any)
│       └── HalfEdge → ...
├── Edges[]
│   ├── HalfEdge (left)
│   └── HalfEdge (right)
└── Vertices[]
    └── position (SFVec3f)
```

### Solid

The main class. Represents a complete polygon-based boundary representation. Maintains lists of faces, edges, and vertices.

```go
type Solid struct {
    Faces    []*Face
    Edges    []*Edge
    Vertices []*Vertex
}
```

### Face

A polygon, possibly with holes. Contains a list of Loops (which are lists of HalfEdges pointing to vertices). May carry per-face color, texture coordinate, and normal data.

### Edge

Connects two vertices. Contains two HalfEdges (one for each adjacent face). Provides fast O(1) traversal between faces.

### HalfEdge

The fundamental element. Each half-edge belongs to one face and points to one vertex. Together, the half-edge pair for an edge provides bidirectional face-to-face traversal.

A HalfEdge can carry:
- Color data
- Texture coordinates
- Normal vector

### Loop

A cyclic list of HalfEdges forming the boundary of a face. The outer loop defines the face; inner loops define holes.

### Vertex

A point in 3D space. May also store color, texture coordinate, or normal data.

## Euler Operators

The half-edge data structure is modified through **Euler operators** — low-level operations that maintain topological consistency:

| Operator | Description |
|----------|-------------|
| `Mvfs` | Make vertex, face, solid — creates initial topology |
| `Lmev` | Loop, make edge, vertex — splits a vertex |
| `Lmef` | Loop, make edge, face — splits a face |
| `Lkev` | Loop, kill edge, vertex — merges vertices |
| `Lkef` | Loop, kill edge, face — merges faces |
| `Lkemr` | Loop, kill edge, make ring — creates a hole |
| `Lmekr` | Loop, make edge, kill ring — closes a hole |

These operators guarantee that the resulting data structure always represents a valid solid (satisfying Euler's formula: V - E + F = 2 for simple polyhedra).

## Boolean Operations

Boolean operations combine two solids:

- **Union**: The combined volume of both solids
- **Intersection**: Only the volume shared by both
- **Difference**: Volume of solid A minus solid B

The algorithm involves:
1. **Classification**: Classify edges/faces of each solid against the other
2. **Splitting**: Split faces that cross the boundary
3. **Connection**: Connect the split boundaries
4. **Merging**: Merge the two half-edge structures
5. **Re-classification**: Remove interior/exterior faces based on operation type

Implementation tracked in [Issue #1](https://github.com/TrueBlocks/trueblocks-3d/issues/1).

## Splitting Planes

Cuts a solid along a plane, producing two new solids. Uses the same classify/split/connect pattern as boolean operations but against an infinite plane rather than another solid.

Implementation tracked in [Issue #2](https://github.com/TrueBlocks/trueblocks-3d/issues/2).
