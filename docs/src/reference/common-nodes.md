# Common Nodes

## Shape

Pairs geometry with appearance. This is the only way to make geometry visible.

```go
type Shape struct {
    Appearance *Appearance
    Geometry   Node  // any geometry node
}
```

```vrml
Shape {
    appearance Appearance {
        material Material { diffuseColor 1 0 0 }
    }
    geometry Box { size 2 2 2 }
}
```

## DirectionalLight

Infinite parallel light rays (like the sun). Illuminates all geometry in the same group.

```go
type DirectionalLight struct {
    On               bool      // default true
    Color            SFColor   // default {1, 1, 1}
    Intensity        float32   // 0–1, default 1
    AmbientIntensity float32   // 0–1, default 0
    Direction        SFVec3f   // default {0, 0, -1}
}
```

## PointLight

Omnidirectional point light source. Intensity attenuates with distance.

```go
type PointLight struct {
    On               bool
    Color            SFColor
    Intensity        float32
    AmbientIntensity float32
    Location         SFVec3f    // default {0, 0, 0}
    Radius           float32    // max range, default 100
    Attenuation      SFVec3f    // {constant, linear, quadratic}, default {1, 0, 0}
}
```

## SpotLight

Cone-shaped light with direction and angular falloff.

```go
type SpotLight struct {
    On               bool
    Color            SFColor
    Intensity        float32
    AmbientIntensity float32
    Location         SFVec3f
    Direction        SFVec3f    // default {0, 0, -1}
    CutOffAngle      float32    // outer cone angle, default π/4
    BeamWidth        float32    // inner cone angle, default π/2
    Radius           float32    // default 100
    Attenuation      SFVec3f
}
```

## AudioClip

Loads and plays audio from a URL.

```go
type AudioClip struct {
    URL       []string
    Loop      bool
    Pitch     float32    // playback speed, default 1.0
    StartTime float64
    StopTime  float64
    Description string
}
```

**Status**: [Issue #18](https://github.com/TrueBlocks/trueblocks-3d/issues/18)

## Sound

Spatializes audio in 3D space.

```go
type Sound struct {
    Source    Node       // AudioClip or MovieTexture
    Location SFVec3f
    Direction SFVec3f   // default {0, 0, 1}
    Intensity float32   // default 1.0
    MinFront  float32
    MaxFront  float32   // default 10
    MinBack   float32
    MaxBack   float32   // default 10
    Spatialize bool     // default true
    Priority  float32
}
```

## Script

Embedded scripting for event processing.

```go
type Script struct {
    URL             []string
    DirectOutput    bool
    MustEvaluate    bool
    // User-defined fields for eventIn/eventOut
}
```

**Status**: [Issue #27](https://github.com/TrueBlocks/trueblocks-3d/issues/27)

## WorldInfo

Scene metadata (not rendered).

```go
type WorldInfo struct {
    Title string
    Info  []string
}
```
