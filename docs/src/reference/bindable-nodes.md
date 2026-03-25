# Bindable Nodes

Only one node of each bindable type can be active at a time. The browser maintains a stack for each: Viewpoint, Background, NavigationInfo, Fog. When a new node is "bound", it becomes active and the previous one is pushed onto the stack.

## Viewpoint

Camera position and orientation.

```go
type Viewpoint struct {
    Position    SFVec3f     // default {0, 0, 10}
    Orientation SFRotation  // default {0, 0, 1, 0}
    FieldOfView float32     // radians, default 0.785398 (π/4)
    Description string
    Jump        bool        // default true
}
```

```vrml
Viewpoint {
    position 0 5 15
    orientation 1 0 0 -0.3
    description "Main camera"
}
```

## Background

Sky and ground colors plus optional skybox textures.

```go
type Background struct {
    SkyColor   []SFColor   // default [{0, 0, 0}]
    SkyAngle   []float32
    GroundColor []SFColor
    GroundAngle []float32
    BackURL    []string
    BottomURL  []string
    FrontURL   []string
    LeftURL    []string
    RightURL   []string
    TopURL     []string
}
```

Sky/ground colors are specified as gradients from zenith to horizon using angle arrays.

## NavigationInfo

Controls how the viewer navigates the scene.

```go
type NavigationInfo struct {
    Type            []string  // "WALK", "EXAMINE", "FLY", "NONE"
    Speed           float32   // meters/second, default 1.0
    AvatarSize      []float32 // [collision_radius, eye_height, step_height]
    VisibilityLimit float32   // max draw distance, 0 = infinite
    Headlight       bool      // default true
}
```

**Status**: [Issue #21](https://github.com/TrueBlocks/trueblocks-3d/issues/21)

## Fog

Distance-based atmospheric fog.

```go
type Fog struct {
    Color           SFColor   // default {1, 1, 1}
    FogType         string    // "LINEAR" or "EXPONENTIAL"
    VisibilityRange float32   // distance at which fog is opaque, 0 = no fog
}
```

**Status**: [Issue #19](https://github.com/TrueBlocks/trueblocks-3d/issues/19)
