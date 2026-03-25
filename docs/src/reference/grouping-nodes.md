# Grouping Nodes

Grouping nodes organize children into hierarchies. All grouping nodes have a `Children` field.

## Transform

The most common grouping node. Applies translation, rotation, and scale to its children.

```go
type Transform struct {
    Translation      SFVec3f     // default {0, 0, 0}
    Rotation         SFRotation  // default {0, 0, 1, 0}
    Scale            SFVec3f     // default {1, 1, 1}
    ScaleOrientation SFRotation  // default {0, 0, 1, 0}
    Center           SFVec3f     // default {0, 0, 0}
    Children         []Node
    BboxCenter       SFVec3f
    BboxSize         SFVec3f
}
```

```vrml
Transform {
    translation 2 0 0
    rotation 0 1 0 1.57
    scale 1.5 1.5 1.5
    children [ Shape { ... } ]
}
```

## Group

Groups children without applying any transform.

```go
type Group struct {
    Children   []Node
    BboxCenter SFVec3f
    BboxSize   SFVec3f
}
```

## Anchor

Hyperlink node. Loads a URL when any child geometry is clicked.

```go
type Anchor struct {
    URL         []string
    Description string
    Parameter   []string
    Children    []Node
    BboxCenter  SFVec3f
    BboxSize    SFVec3f
}
```

**Status**: [Issue #29](https://github.com/TrueBlocks/trueblocks-3d/issues/29)

## Billboard

Auto-rotates children to always face the camera.

```go
type Billboard struct {
    AxisOfRotation SFVec3f  // default {0, 1, 0}; {0,0,0} = full camera-facing
    Children       []Node
    BboxCenter     SFVec3f
    BboxSize       SFVec3f
}
```

**Status**: [Issue #29](https://github.com/TrueBlocks/trueblocks-3d/issues/29)

## Collision

Enables or disables collision detection for its children.

```go
type Collision struct {
    Collide    bool   // default true
    Proxy      Node   // simplified collision geometry
    Children   []Node
    BboxCenter SFVec3f
    BboxSize   SFVec3f
}
```

**Status**: [Issue #21](https://github.com/TrueBlocks/trueblocks-3d/issues/21)

## Inline

Loads an external `.wrl` file and inserts it into the scene graph.

```go
type Inline struct {
    URL        []string
    BboxCenter SFVec3f
    BboxSize   SFVec3f
}
```

**Status**: [Issue #26](https://github.com/TrueBlocks/trueblocks-3d/issues/26)

## LOD (Level of Detail)

Selects one child based on distance from the viewer. Use for performance optimization.

```go
type LOD struct {
    Level  []Node
    Center SFVec3f
    Range  []float32  // distance thresholds
}
```

**Status**: [Issue #28](https://github.com/TrueBlocks/trueblocks-3d/issues/28)

## Switch

Shows exactly one child selected by `WhichChoice`. Use for toggling geometry or implementing state machines.

```go
type Switch struct {
    Choice      []Node
    WhichChoice int32  // -1 = none visible, default -1
}
```

**Status**: [Issue #28](https://github.com/TrueBlocks/trueblocks-3d/issues/28)
