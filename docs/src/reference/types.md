# Data Types

VRML97 defines field types for node data. These are mapped to Go types as follows.

## Single-Valued Types (SF)

| VRML Type | Go Type | Description |
|-----------|---------|-------------|
| `SFBool` | `bool` | `TRUE` or `FALSE` |
| `SFInt32` | `int32` | 32-bit signed integer |
| `SFFloat` | `float32` | IEEE 754 single precision |
| `SFTime` | `float64` | Double precision seconds since epoch |
| `SFVec2f` | `vec.SFVec2f` | 2D vector `{X, Y float32}` |
| `SFVec3f` | `vec.SFVec3f` | 3D vector `{X, Y, Z float32}` |
| `SFColor` | `vec.SFColor` | RGBA color `{R, G, B, A float32}` (0–1 range) |
| `SFRotation` | `vec.SFRotation` | Axis-angle `{X, Y, Z, Angle float32}` (radians) |
| `SFString` | `string` | UTF-8 string |
| `SFImage` | `image.Image` | Go standard library image |
| `SFNode` | `node.Node` | Reference to any VRML node |

### SFVec2f

```go
type SFVec2f struct {
    X, Y float32
}
```

Supports: `Add`, `Sub`, `Scale`, `Length`, `Normalize`, `Dot`.

### SFVec3f

```go
type SFVec3f struct {
    X, Y, Z float32
}
```

Supports: `Add`, `Sub`, `Scale`, `Length`, `Normalize`, `Dot`, `Cross`.

### SFColor

```go
type SFColor struct {
    R, G, B, A float32
}
```

Colors use 0–1 range. `A` (alpha) is 1.0 for opaque, 0.0 for fully transparent. Predefined colors: `vec.Black`, `vec.White`, `vec.Red`, `vec.Green`, `vec.Blue`.

### SFRotation

```go
type SFRotation struct {
    X, Y, Z, Angle float32
}
```

Axis-angle rotation. The axis `(X, Y, Z)` should be normalized. `Angle` is in radians. Identity rotation: `{0, 0, 1, 0}`.

## Multi-Valued Types (MF)

Multi-valued types are Go slices of single-valued types:

| VRML Type | Go Type | Description |
|-----------|---------|-------------|
| `MFInt32` | `[]int32` | Array of integers |
| `MFFloat` | `[]float32` | Array of floats |
| `MFVec2f` | `[]vec.SFVec2f` | Array of 2D vectors |
| `MFVec3f` | `[]vec.SFVec3f` | Array of 3D vectors |
| `MFColor` | `[]vec.SFColor` | Array of colors |
| `MFRotation` | `[]vec.SFRotation` | Array of rotations |
| `MFString` | `[]string` | Array of strings |
| `MFNode` | `[]node.Node` | Array of nodes |

In VRML files, multi-valued fields use bracket notation:

```vrml
coord Coordinate {
    point [1 0 0, 0 1 0, -1 0 0]
}
coordIndex [0 1 2 -1]
```

## Matrix

```go
type Matrix [4][4]float32
```

4x4 transformation matrix. Supports: `Identity`, `Multiply`, `Translate`, `RotateAxis`, `Scale`, `Inverse`.
