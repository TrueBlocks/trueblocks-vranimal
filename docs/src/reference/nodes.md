# Node Reference

All VRML97 nodes are defined in `pkg/node`. Each node type implements the `Node` interface:

```go
type Node interface {
    NodeType() string
    GetName() string
    SetName(string)
}
```

Nodes are organized into the following categories:

- [Appearance Nodes](./appearance-nodes.md) — visual styling
- [Bindable Nodes](./bindable-nodes.md) — environment settings  
- [Common Nodes](./common-nodes.md) — lights, audio, shapes
- [Geometry Nodes](./geometry-nodes.md) — 3D shapes
- [Grouping Nodes](./grouping-nodes.md) — scene hierarchy
- [Interpolator Nodes](./interpolator-nodes.md) — animation values
- [Sensor Nodes](./sensor-nodes.md) — interaction detection

## Quick Reference

| Node | Category | Fields |
|------|----------|--------|
| Anchor | Grouping | `url`, `description`, `children` |
| Appearance | Appearance | `material`, `texture`, `textureTransform` |
| AudioClip | Common | `url`, `loop`, `startTime`, `stopTime` |
| Background | Bindable | `skyColor`, `groundColor`, `*Url` (6 faces) |
| Billboard | Grouping | `axisOfRotation`, `children` |
| Box | Geometry | `size` |
| Collision | Grouping | `collide`, `proxy`, `children` |
| Color | Geometry | `color` |
| ColorInterpolator | Interpolator | `key`, `keyValue` |
| Cone | Geometry | `bottomRadius`, `height`, `side`, `bottom` |
| Coordinate | Geometry | `point` |
| CoordinateInterpolator | Interpolator | `key`, `keyValue` |
| Cylinder | Geometry | `radius`, `height`, `side`, `top`, `bottom` |
| CylinderSensor | Sensor | `minAngle`, `maxAngle` |
| DirectionalLight | Common | `direction`, `color`, `intensity` |
| ElevationGrid | Geometry | `xDimension`, `zDimension`, `xSpacing`, `zSpacing`, `height` |
| Extrusion | Geometry | `crossSection`, `spine`, `scale`, `orientation` |
| Fog | Bindable | `color`, `fogType`, `visibilityRange` |
| FontStyle | Appearance | `family`, `style`, `size` |
| Group | Grouping | `children` |
| ImageTexture | Appearance | `url`, `repeatS`, `repeatT` |
| IndexedFaceSet | Geometry | `coord`, `coordIndex`, `color`, `normal`, `texCoord` |
| IndexedLineSet | Geometry | `coord`, `coordIndex`, `color` |
| Inline | Grouping | `url`, `bboxCenter`, `bboxSize` |
| LOD | Grouping | `level`, `center`, `range` |
| Material | Appearance | `diffuseColor`, `specularColor`, `emissiveColor`, `shininess`, `transparency` |
| MovieTexture | Appearance | `url`, `loop`, `speed` |
| NavigationInfo | Bindable | `type`, `speed`, `avatarSize`, `headlight` |
| Normal | Geometry | `vector` |
| NormalInterpolator | Interpolator | `key`, `keyValue` |
| OrientationInterpolator | Interpolator | `key`, `keyValue` |
| PixelTexture | Appearance | `image`, `repeatS`, `repeatT` |
| PlaneSensor | Sensor | `minPosition`, `maxPosition` |
| PointLight | Common | `location`, `color`, `intensity`, `radius` |
| PointSet | Geometry | `coord`, `color` |
| PositionInterpolator | Interpolator | `key`, `keyValue` |
| ProximitySensor | Sensor | `center`, `size` |
| ScalarInterpolator | Interpolator | `key`, `keyValue` |
| Script | Common | `url`, fields |
| Shape | Common | `appearance`, `geometry` |
| Sound | Common | `source`, `location`, `direction`, `intensity` |
| Sphere | Geometry | `radius` |
| SphereSensor | Sensor | `offset` |
| SpotLight | Common | `location`, `direction`, `cutOffAngle`, `beamWidth` |
| Switch | Grouping | `choice`, `whichChoice` |
| Text | Geometry | `string`, `fontStyle`, `maxExtent`, `length` |
| TextureCoordinate | Geometry | `point` |
| TextureTransform | Appearance | `center`, `rotation`, `scale`, `translation` |
| TimeSensor | Sensor | `cycleInterval`, `loop`, `startTime`, `stopTime` |
| TouchSensor | Sensor | (isOver, isActive events) |
| Transform | Grouping | `translation`, `rotation`, `scale`, `children` |
| Viewpoint | Bindable | `position`, `orientation`, `fieldOfView` |
| VisibilitySensor | Sensor | `center`, `size` |
| WorldInfo | Common | `title`, `info` |
