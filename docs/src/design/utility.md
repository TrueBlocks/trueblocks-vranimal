# Utility Library (VRFC)

The utility library contains supporting types and functions used by all other components. In the original C++ SDK this was called the "Virtual Reality Foundation Classes" (VRFC), modeled after Microsoft's MFC.

In Go, these concepts map to simpler, idiomatic packages.

## Collection Types

**C++ originals**: `vrArray`, `vrArrayBase`, `vrList`, `vrRefCountList`, `vrStack`, `vrRefCountStack`, `vrIntrusiveList`, `vrRingList`

**Go equivalent**: Go slices and maps replace all collection classes. No custom container types are needed:

```go
// C++: MFVec3f was vrArray<SFVec3f>
// Go: just use a slice
type MFVec3f []SFVec3f

// C++: vrRefCountList with manual reference counting
// Go: garbage collector handles memory automatically
```

## Geometric Types

**C++ originals**: `SFVec2f`, `SFVec3f`, `SFColor`, `SFRotation`, `vrMatrix`, `vrBoundingBox`, `vrPlane`, `vrRay`, `vrRect2D`

**Go package**: `pkg/vec`

```go
type SFVec2f struct{ X, Y float32 }
type SFVec3f struct{ X, Y, Z float32 }
type SFColor struct{ R, G, B, A float32 }
type SFRotation struct{ X, Y, Z, Angle float32 }
type Matrix [4][4]float32
```

Each type supports arithmetic operations (add, subtract, scale, interpolate) so code can closely mirror the VRML specification. This makes the learning curve shallow — VRaniML code looks like VRML.

## Data Types

VRML97 defines single-valued (SF) and multi-valued (MF) field types:

| VRML Type | Go Type | Description |
|-----------|---------|-------------|
| `SFBool` | `bool` | Boolean |
| `SFInt32` | `int32` | 32-bit integer |
| `SFFloat` | `float32` | Single precision float |
| `SFTime` | `float64` | Double precision time |
| `SFVec2f` | `vec.SFVec2f` | 2D vector |
| `SFVec3f` | `vec.SFVec3f` | 3D vector |
| `SFColor` | `vec.SFColor` | RGBA color |
| `SFRotation` | `vec.SFRotation` | Axis-angle rotation |
| `SFString` | `string` | Text string |
| `SFNode` | `node.Node` | Node reference (interface) |
| `MFInt32` | `[]int32` | Array of integers |
| `MFFloat` | `[]float32` | Array of floats |
| `MFVec2f` | `[]vec.SFVec2f` | Array of 2D vectors |
| `MFVec3f` | `[]vec.SFVec3f` | Array of 3D vectors |
| `MFColor` | `[]vec.SFColor` | Array of colors |
| `MFRotation` | `[]vec.SFRotation` | Array of rotations |
| `MFString` | `[]string` | Array of strings |
| `MFNode` | `[]node.Node` | Array of nodes |

## Utility Functions

Common utility functions from the C++ library map naturally to Go:

| C++ Function | Go Equivalent | Description |
|-------------|---------------|-------------|
| `MIN(a, b)` | `min(a, b)` | Built-in since Go 1.21 |
| `MAX(a, b)` | `max(a, b)` | Built-in since Go 1.21 |
| `vrClamp(val, lo, hi)` | `max(lo, min(hi, val))` | Clamp to range |
| `vrInterpolate(from, to, fKey, tKey, t)` | Custom function | Linear interpolation |
| `vrDeg2Rad(val)` | `val * math.Pi / 180` | Degrees to radians |
| `vrRad2Deg(val)` | `val * 180 / math.Pi` | Radians to degrees |
| `vrEquals(a, b, tol)` | `math.Abs(a-b) < tol` | Floating point compare |
| `vrReadTextureImage(img, file)` | `image.Decode()` | Go stdlib handles this |
| `vrCacheFile(url)` | `http.Get()` + `os.Create()` | Go stdlib handles this |

## Runtime Typing

**C++ originals**: `DECLARE_NODE`, `IMPLEMENT_NODE`, `GETRUNTIME_CLASS`, `vrRuntimeClass`, `vrBaseNode`

**Go equivalent**: Go's interfaces and type assertions replace all runtime typing macros:

```go
// C++: DECLARE_NODE(vrBox)  +  IMPLEMENT_NODE(vrBox, vrGeometryNode, 1)
// Go: just implement the interface
type Box struct {
    Size vec.SFVec3f
}
func (b *Box) NodeType() string { return "Box" }

// C++: node->IsKindOf(GETRUNTIME_CLASS(vrBox))
// Go: type assertion
if box, ok := n.(*node.Box); ok { ... }
```

## Memory Management

**C++**: Manual reference counting with `vrDELETE` macro.

**Go**: The garbage collector handles all memory management automatically. No reference counting needed.
