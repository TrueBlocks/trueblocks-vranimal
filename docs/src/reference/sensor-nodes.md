# Sensor Nodes

Sensors detect user interaction and environmental conditions, generating events that drive animation and interaction via ROUTEs.

All sensors have an `Enabled` field (default `true`).

## TimeSensor

The animation clock. Generates `fraction_changed` events (0.0–1.0) over its cycle interval.

```go
type TimeSensor struct {
    CycleInterval float64  // seconds, default 1
    Loop          bool     // default false
    StartTime     float64
    StopTime      float64
    Enabled       bool     // default true
}
```

```vrml
DEF Clock TimeSensor { cycleInterval 4 loop TRUE }
ROUTE Clock.fraction_changed TO Interp.set_fraction
```

## TouchSensor

Detects mouse click/hover on sibling geometry. Generates `isOver`, `isActive`, `touchTime`, `hitPoint_changed` events.

```go
type TouchSensor struct {
    Enabled bool
}
```

## PlaneSensor

Maps mouse drag to 2D translation in a plane.

```go
type PlaneSensor struct {
    MinPosition SFVec2f  // default {0, 0}
    MaxPosition SFVec2f  // default {-1, -1} (no clamping)
    Offset      SFVec3f
    AutoOffset  bool     // default true
    Enabled     bool
}
```

## CylinderSensor

Maps mouse drag to rotation around a Y axis.

```go
type CylinderSensor struct {
    MinAngle   float32  // default 0
    MaxAngle   float32  // default -1 (no clamping)
    DiskAngle  float32  // default π/12
    Offset     float32
    AutoOffset bool     // default true
    Enabled    bool
}
```

## SphereSensor

Maps mouse drag to free spherical rotation.

```go
type SphereSensor struct {
    Offset     SFRotation
    AutoOffset bool  // default true
    Enabled    bool
}
```

## ProximitySensor

Detects when the viewer enters or exits a box-shaped region.

```go
type ProximitySensor struct {
    Center  SFVec3f
    Size    SFVec3f
    Enabled bool
}
```

Generates: `enterTime`, `exitTime`, `position_changed`, `orientation_changed`.

## VisibilitySensor

Detects when a box-shaped region becomes visible or hidden.

```go
type VisibilitySensor struct {
    Center  SFVec3f
    Size    SFVec3f
    Enabled bool
}
```

Generates: `enterTime`, `exitTime`.

**Status**: Mouse picking for sensors in [Issue #15](https://github.com/TrueBlocks/trueblocks-3d/issues/15)
