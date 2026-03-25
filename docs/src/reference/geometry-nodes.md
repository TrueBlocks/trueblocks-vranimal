# Geometry Nodes

All geometry nodes implement the `GeometryNode` interface:

```go
type GeometryNode interface {
    Node
    GetSolid() bool  // true = backface culling enabled
}
```

## Box

Axis-aligned rectangular box centered at the origin.

```go
type Box struct {
    Size SFVec3f  // default {2, 2, 2}
}
```

```vrml
geometry Box { size 1 2 3 }
```

## Sphere

Geodesic sphere centered at the origin.

```go
type Sphere struct {
    Radius float32  // default 1
}
```

## Cone

Cone with apex at top, base at bottom, centered at origin.

```go
type Cone struct {
    BottomRadius float32  // default 1
    Height       float32  // default 2
    Side         bool     // render side, default true
    Bottom       bool     // render bottom cap, default true
}
```

## Cylinder

Cylinder centered at the origin.

```go
type Cylinder struct {
    Radius float32  // default 1
    Height float32  // default 2
    Side   bool     // default true
    Top    bool     // default true
    Bottom bool     // default true
}
```

## IndexedFaceSet

Arbitrary polygonal mesh defined by vertex coordinates and face indices.

```go
type IndexedFaceSet struct {
    Coord           *Coordinate
    CoordIndex      []int32          // face indices, -1 = face separator
    Color           *Color
    ColorIndex      []int32
    ColorPerVertex  bool             // default true
    Normal          *Normal
    NormalIndex     []int32
    NormalPerVertex bool             // default true
    TexCoord        *TextureCoordinate
    TexCoordIndex   []int32
    Solid           bool             // backface culling, default true
    Ccw             bool             // counter-clockwise winding, default true
    Convex          bool             // all faces convex, default true
    CreaseAngle     float32          // auto-smooth angle in radians
}
```

```vrml
geometry IndexedFaceSet {
    coord Coordinate {
        point [-1 0 -1, 1 0 -1, 1 0 1, -1 0 1, 0 1.5 0]
    }
    coordIndex [0 1 4 -1, 1 2 4 -1, 2 3 4 -1, 3 0 4 -1, 3 2 1 0 -1]
}
```

## ElevationGrid

Height-field terrain on an X-Z grid.

```go
type ElevationGrid struct {
    XDimension     int32
    ZDimension     int32
    XSpacing       float32          // default 1.0
    ZSpacing       float32          // default 1.0
    Height         []float32        // xDimension * zDimension values
    Color          *Color
    ColorPerVertex bool
    Normal         *Normal
    NormalPerVertex bool
    TexCoord       *TextureCoordinate
    Solid          bool
    Ccw            bool
    CreaseAngle    float32
}
```

## Extrusion

Sweeps a 2D cross-section along a 3D spine.

```go
type Extrusion struct {
    CrossSection []SFVec2f     // 2D profile, default unit square
    Spine        []SFVec3f     // 3D path, default straight line
    Scale        []SFVec2f     // per-spine scaling
    Orientation  []SFRotation  // per-spine rotation
    BeginCap     bool          // default true
    EndCap       bool          // default true
    Solid        bool
    Ccw          bool
    Convex       bool
    CreaseAngle  float32
}
```

## IndexedLineSet

Polylines (not filled, wireframe only).

```go
type IndexedLineSet struct {
    Coord      *Coordinate
    CoordIndex []int32
    Color      *Color
    ColorIndex []int32
    ColorPerVertex bool
}
```

## PointSet

Point cloud.

```go
type PointSet struct {
    Coord *Coordinate
    Color *Color
}
```

## Text

3D text string.

```go
type Text struct {
    String    []string
    FontStyle *FontStyle
    MaxExtent float32
    Length    []float32
}
```

**Status**: [Issue #20](https://github.com/TrueBlocks/trueblocks-3d/issues/20)

## Supporting Nodes

### Coordinate

Shared vertex position data — `type Coordinate struct { Point []SFVec3f }`.

### Color

Shared color data — `type Color struct { Color []SFColor }`.

### Normal

Shared normal vector data — `type Normal struct { Vector []SFVec3f }`.

### TextureCoordinate

UV texture mapping coordinates — `type TextureCoordinate struct { Point []SFVec2f }`.
