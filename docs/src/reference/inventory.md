# Node Inventory (Port Status)

Status of each VRML97 node in the Go port.

**Legend**: Done = fully working, Partial = parsed but not all features, Stub = type exists but no runtime behavior, Not Started = no code yet.

| Node | Parser | Converter | Runtime Events | Notes |
|------|--------|-----------|----------------|-------|
| **Appearance Nodes** | | | | |
| Appearance | Done | Done | N/A | |
| FontStyle | Done | Not Started | N/A | [#20](https://github.com/TrueBlocks/trueblocks-3d/issues/20) |
| ImageTexture | Done | Not Started | N/A | [#17](https://github.com/TrueBlocks/trueblocks-3d/issues/17) |
| Material | Done | Done | N/A | |
| MovieTexture | Done | Not Started | N/A | |
| PixelTexture | Done | Not Started | N/A | |
| TextureTransform | Done | Not Started | N/A | |
| **Bindable Nodes** | | | | |
| Background | Done | Partial | N/A | Sky color only |
| Fog | Done | Not Started | N/A | [#19](https://github.com/TrueBlocks/trueblocks-3d/issues/19) |
| NavigationInfo | Done | Not Started | N/A | [#21](https://github.com/TrueBlocks/trueblocks-3d/issues/21) |
| Viewpoint | Done | Done | N/A | |
| **Common Nodes** | | | | |
| AudioClip | Done | Not Started | Not Started | [#18](https://github.com/TrueBlocks/trueblocks-3d/issues/18) |
| DirectionalLight | Done | Done | N/A | |
| PointLight | Done | Done | N/A | |
| Script | Stub | Not Started | Not Started | [#27](https://github.com/TrueBlocks/trueblocks-3d/issues/27) |
| Shape | Done | Done | N/A | |
| Sound | Done | Not Started | Not Started | [#18](https://github.com/TrueBlocks/trueblocks-3d/issues/18) |
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
| IndexedLineSet | Done | Not Started | N/A | |
| Normal | Done | Done | N/A | |
| PointSet | Done | Not Started | N/A | |
| Sphere | Done | Done | N/A | |
| Text | Done | Not Started | N/A | [#20](https://github.com/TrueBlocks/trueblocks-3d/issues/20) |
| TextureCoordinate | Done | Not Started | N/A | |
| **Grouping Nodes** | | | | |
| Anchor | Done | Partial | Not Started | [#29](https://github.com/TrueBlocks/trueblocks-3d/issues/29) |
| Billboard | Done | Not Started | Not Started | [#29](https://github.com/TrueBlocks/trueblocks-3d/issues/29) |
| Collision | Done | Not Started | Not Started | [#21](https://github.com/TrueBlocks/trueblocks-3d/issues/21) |
| Group | Done | Done | N/A | |
| Inline | Done | Not Started | N/A | [#26](https://github.com/TrueBlocks/trueblocks-3d/issues/26) |
| LOD | Done | Partial | Not Started | [#28](https://github.com/TrueBlocks/trueblocks-3d/issues/28) |
| Switch | Done | Done | Not Started | [#28](https://github.com/TrueBlocks/trueblocks-3d/issues/28) |
| Transform | Done | Done | N/A | |
| **Interpolator Nodes** | | | | |
| ColorInterpolator | Done | N/A | Not Started | [#16](https://github.com/TrueBlocks/trueblocks-3d/issues/16) |
| CoordinateInterpolator | Done | N/A | Not Started | [#16](https://github.com/TrueBlocks/trueblocks-3d/issues/16) |
| NormalInterpolator | Done | N/A | Not Started | [#16](https://github.com/TrueBlocks/trueblocks-3d/issues/16) |
| OrientationInterpolator | Done | N/A | Not Started | [#16](https://github.com/TrueBlocks/trueblocks-3d/issues/16) |
| PositionInterpolator | Done | N/A | Not Started | [#16](https://github.com/TrueBlocks/trueblocks-3d/issues/16) |
| ScalarInterpolator | Done | N/A | Not Started | [#16](https://github.com/TrueBlocks/trueblocks-3d/issues/16) |
| **Sensor Nodes** | | | | |
| CylinderSensor | Done | N/A | Not Started | [#15](https://github.com/TrueBlocks/trueblocks-3d/issues/15) |
| PlaneSensor | Done | N/A | Not Started | [#15](https://github.com/TrueBlocks/trueblocks-3d/issues/15) |
| ProximitySensor | Done | N/A | Not Started | [#15](https://github.com/TrueBlocks/trueblocks-3d/issues/15) |
| SphereSensor | Done | N/A | Not Started | [#15](https://github.com/TrueBlocks/trueblocks-3d/issues/15) |
| TimeSensor | Done | N/A | Not Started | [#16](https://github.com/TrueBlocks/trueblocks-3d/issues/16) |
| TouchSensor | Done | N/A | Not Started | [#15](https://github.com/TrueBlocks/trueblocks-3d/issues/15) |
| VisibilitySensor | Done | N/A | Not Started | [#15](https://github.com/TrueBlocks/trueblocks-3d/issues/15) |
