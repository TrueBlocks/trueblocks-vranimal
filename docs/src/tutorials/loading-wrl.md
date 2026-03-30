# Tutorial 3: Loading .wrl Files

This tutorial builds a viewer that reads a `.wrl` file from disk and renders it using g3n. This is the Go equivalent of the original "MFC Application" tutorial.

## The Viewer

The `cmd/viewer/main.go` application is the primary example of this pattern:

```bash
./viewer examples/test_scene.wrl
```

## How It Works

```text
.wrl file → parser.NewParser() → p.Parse() → []node.Node → converter.Convert() → g3n scene → render loop
```

### Step 1: Parse the File

```go
import (
    "os"
    "path/filepath"

    "github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
)

f, err := os.Open("scene.wrl")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

p := parser.NewParser(f)
p.SetBaseDir(filepath.Dir("scene.wrl"))
nodes := p.Parse()
if errs := p.Errors(); len(errs) > 0 {
    log.Fatal(errs)
}
```

The parser reads the VRML97 text format and returns a slice of root-level nodes.

### Step 2: Convert to g3n

```go
import (
    "github.com/TrueBlocks/trueblocks-vranimal/pkg/converter"
    "github.com/g3n/engine/core"
)

g3nScene := core.NewNode()
nm := converter.Convert(nodes, g3nScene, baseDir)
```

The `Convert` function walks the VRML node tree and creates equivalent g3n objects under `g3nScene`. It returns a `*NodeMap` that maps between VRML and g3n nodes for dynamic updates.

This walks the VRML node tree and creates equivalent g3n objects:
- `Transform` → `core.Node` with position/rotation/scale
- `Shape` + geometry → `graphic.Mesh` with appropriate material
- Lights → g3n light objects
- `Switch` → only the selected child is added
- `LOD` → child selected by distance

### Step 3: Extract Camera from Viewpoint

```go
vp := converter.GetViewpoint(nodes)
if vp != nil {
    cam.SetPosition(vp.Position.X, vp.Position.Y, vp.Position.Z)
    // Apply orientation as axis-angle rotation
}
```

### Step 4: Apply Background

```go
bg := converter.GetBackground(nodes)
if bg != nil {
    // Set clear color from bg.SkyColor[0]
}
```

### Step 5: Render Loop

Same as Tutorial 2 — clear, update aspect ratio, render.

## Example .wrl File

```vrml
#VRML V2.0 utf8

DEF MyViewpoint Viewpoint {
    position 0 5 15
    orientation 1 0 0 -0.3
}

Background {
    skyColor [0.2 0.2 0.4]
}

DirectionalLight {
    direction -0.5 -1 -0.5
    intensity 0.8
}

DEF RedBox Transform {
    translation -2 0.5 0
    children [
        Shape {
            appearance Appearance {
                material Material {
                    diffuseColor 1 0 0
                }
            }
            geometry Box { size 1 1 1 }
        }
    ]
}

DEF GreenSphere Transform {
    translation 2 0.5 0
    children [
        Shape {
            appearance Appearance {
                material Material {
                    diffuseColor 0 0.8 0
                }
            }
            geometry Sphere { radius 0.7 }
        }
    ]
}

DEF Ground Shape {
    appearance Appearance {
        material Material {
            diffuseColor 0.5 0.5 0.5
        }
    }
    geometry IndexedFaceSet {
        coord Coordinate {
            point [-5 0 -5, 5 0 -5, 5 0 5, -5 0 5]
        }
        coordIndex [0 1 2 3 -1]
    }
}
```

## What You Learned

- The full pipeline from `.wrl` file to rendered 3D scene
- `parser.NewParser()` and `p.Parse()` handle the VRML text format
- `converter.Convert(nodes, parent, baseDir)` maps VRML nodes to g3n scene objects
- `converter.GetViewpoint()` / `GetBackground()` extract environment settings
- Any valid VRML97 file (using supported node types) can be loaded and viewed
