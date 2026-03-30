# Node Inventory (Port Status)

Status of each VRML97 node in the Go port.

**Legend**: Done = fully working, Partial = parsed but not all features, Deferred = intentionally deferred to a later phase.

| Node | Parser | Converter | Runtime Events | Notes |
|------|--------|-----------|----------------|-------|
| **Appearance Nodes** | | | | |
| Appearance | Done | Done | N/A | |
| FontStyle | Done | Done | N/A | Size, family, justify, style |
| ImageTexture | Done | Done | N/A | |
| Material | Done | Done | N/A | |
| MovieTexture | Done | Done | N/A | |
| PixelTexture | Done | Done | N/A | |
| TextureTransform | Done | Done | N/A | |
| **Bindable Nodes** | | | | |
| Background | Done | Done | Done | Binding stack managed by Browser |
| Fog | Done | Done | Done | Params exposed via GetFogParams() |
| NavigationInfo | Done | Done | Done | Binding stack managed by Browser |
| Viewpoint | Done | Done | Done | Binding stack managed by Browser |
| **Common Nodes** | | | | |
| AudioClip | Done | Done | Done | WAV/OGG via g3n OpenAL |
| DirectionalLight | Done | Done | N/A | |
| PointLight | Done | Done | N/A | |
| Script | Done | N/A | Done | Go callback handlers |
| Shape | Done | Done | N/A | |
| Sound | Done | Done | Done | Spatial audio via g3n OpenAL |
| SpotLight | Done | Done | N/A | |
| WorldInfo | Done | N/A | N/A | |
| **Geometry Nodes** | | | | |
| Box | Done | Done | N/A | |
| Color | Done | Done | N/A | |
| Cone | Done | Done | N/A | |
| Coordinate | Done | Done | N/A | |
| Cylinder | Done | Done | N/A | |
| ElevationGrid | Done | Done | N/A | |
| Extrusion | Done | Done | N/A | |
| IndexedFaceSet | Done | Done | N/A | |
| IndexedLineSet | Done | Done | N/A | |
| Normal | Done | Done | N/A | |
| PointSet | Done | Done | N/A | |
| Sphere | Done | Done | N/A | |
| Text | Done | Done | N/A | TrueType font rendering to texture |
| TextureCoordinate | Done | Done | N/A | |
| **Grouping Nodes** | | | | |
| Anchor | Done | Done | Done | |
| Billboard | Done | Done | Done | Dynamic rotation in converter |
| Collision | Done | Done | Done | |
| Group | Done | Done | N/A | |
| Inline | Done | Done | N/A | URL-based .wrl loading |
| LOD | Done | Done | Done | Distance-based child selection |
| Switch | Done | Done | Done | whichChoice-based selection |
| Transform | Done | Done | N/A | |
| **Interpolator Nodes** | | | | |
| ColorInterpolator | Done | N/A | Done | |
| CoordinateInterpolator | Done | N/A | Done | |
| NormalInterpolator | Done | N/A | Done | |
| OrientationInterpolator | Done | N/A | Done | |
| PositionInterpolator | Done | N/A | Done | |
| ScalarInterpolator | Done | N/A | Done | |
| **Sensor Nodes** | | | | |
| CylinderSensor | Done | N/A | Done | |
| PlaneSensor | Done | N/A | Done | |
| ProximitySensor | Done | N/A | Done | |
| SphereSensor | Done | N/A | Done | |
| TimeSensor | Done | N/A | Done | |
| TouchSensor | Done | N/A | Done | |
| VisibilitySensor | Done | N/A | Done | |
