# Interpolator Nodes

Interpolators produce smooth value transitions. Each takes a `set_fraction` eventIn (0.0–1.0), typically driven by a TimeSensor via ROUTE, and outputs `value_changed`.

All interpolators share a common pattern:

```go
type XxxInterpolator struct {
    Key      []float32  // fraction values (0.0–1.0), sorted ascending
    KeyValue []T        // corresponding output values
}
```

Given a fraction `f`, the interpolator finds the two keys bracketing `f` and linearly interpolates the corresponding values.

## ColorInterpolator

Interpolates between RGB colors.

```go
type ColorInterpolator struct {
    Key      []float32
    KeyValue []SFColor
}
```

## CoordinateInterpolator

Interpolates vertex positions for mesh morphing.

```go
type CoordinateInterpolator struct {
    Key      []float32
    KeyValue []SFVec3f  // len = len(Key) * numVertices
}
```

## NormalInterpolator

Interpolates normal vectors.

```go
type NormalInterpolator struct {
    Key      []float32
    KeyValue []SFVec3f
}
```

## OrientationInterpolator

Interpolates rotations using spherical linear interpolation (slerp).

```go
type OrientationInterpolator struct {
    Key      []float32
    KeyValue []SFRotation
}
```

## PositionInterpolator

Interpolates 3D positions.

```go
type PositionInterpolator struct {
    Key      []float32
    KeyValue []SFVec3f
}
```

## ScalarInterpolator

Interpolates a single float value.

```go
type ScalarInterpolator struct {
    Key      []float32
    KeyValue []float32
}
```

## Example: Bouncing Ball

```vrml
DEF Timer TimeSensor { cycleInterval 2 loop TRUE }
DEF Mover PositionInterpolator {
    key [0, 0.5, 1]
    keyValue [0 0 0, 0 3 0, 0 0 0]
}
DEF Ball Transform {
    children [ Shape { geometry Sphere { radius 0.5 } } ]
}
ROUTE Timer.fraction_changed TO Mover.set_fraction
ROUTE Mover.value_changed TO Ball.set_translation
```

**Status**: [Issue #16](https://github.com/TrueBlocks/trueblocks-3d/issues/16)
