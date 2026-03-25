# Tutorial 2: Programmatic Scene (Viewer)

This tutorial builds a 3D viewer that constructs a scene graph programmatically (without reading a `.wrl` file) and renders it using the g3n engine. This is the Go equivalent of the original "Win32 Application" tutorial.

## Step 1: Create the Project

```bash
mkdir -p cmd/demo
```

## Step 2: Build a Scene Programmatically

Create `cmd/demo/main.go`:

```go
package main

import (
    "github.com/TrueBlocks/trueblocks-vranimal/pkg/converter"
    "github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
    "github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"

    "github.com/g3n/engine/app"
    "github.com/g3n/engine/camera"
    "github.com/g3n/engine/gls"
    "github.com/g3n/engine/light"
    "github.com/g3n/engine/math32"
    "github.com/g3n/engine/renderer"
    "github.com/g3n/engine/window"
)

func main() {
    // Build VRML scene graph programmatically
    scene := buildScene()

    // Convert VRML nodes to g3n scene objects
    a := app.App()
    g3nScene := converter.Convert(scene)

    // Add lights
    l1 := light.NewDirectional(&math32.Color{1, 1, 1}, 0.8)
    l1.SetPosition(1, 1, 1)
    g3nScene.Add(l1)
    l2 := light.NewAmbient(&math32.Color{1, 1, 1}, 0.3)
    g3nScene.Add(l2)

    // Camera
    cam := camera.New(1)
    cam.SetPosition(0, 2, 5)
    cam.LookAt(&math32.Vector3{0, 0, 0}, &math32.Vector3{0, 1, 0})
    g3nScene.Add(cam)
    camera.NewOrbitControl(cam)

    // Render loop
    a.Run(func(rend *renderer.Renderer, deltaTime float64) {
        a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
        w, h := a.GetFramebufferSize()
        cam.SetAspect(float32(w) / float32(h))
        rend.Render(g3nScene, cam)
    })
}

func buildScene() []node.Node {
    // Red box at the origin
    redBox := &node.Transform{
        Children: []node.Node{
            &node.Shape{
                Appearance: &node.Appearance{
                    Material: &node.Material{
                        DiffuseColor:     vec.SFColor{R: 1, G: 0, B: 0, A: 1},
                        AmbientIntensity: 0.2,
                        Shininess:        0.2,
                    },
                },
                Geometry: &node.Box{
                    Size: vec.SFVec3f{X: 1, Y: 1, Z: 1},
                },
            },
        },
    }
    redBox.SetName("RedBox")

    // Green sphere offset to the right
    greenSphere := &node.Transform{
        Translation: vec.SFVec3f{X: 2, Y: 0, Z: 0},
        Children: []node.Node{
            &node.Shape{
                Appearance: &node.Appearance{
                    Material: &node.Material{
                        DiffuseColor:     vec.SFColor{R: 0, G: 0.8, B: 0, A: 1},
                        AmbientIntensity: 0.2,
                        Shininess:        0.5,
                    },
                },
                Geometry: &node.Sphere{Radius: 0.5},
            },
        },
    }
    greenSphere.SetName("GreenSphere")

    return []node.Node{redBox, greenSphere}
}
```

## Step 3: Build and Run

```bash
CGO_CFLAGS="-I/opt/homebrew/include" \
CGO_LDFLAGS="-L/opt/homebrew/lib" \
go build -o demo ./cmd/demo/

./demo
```

A window opens showing a red box at the origin and a green sphere to its right, lit by a white directional light.

## Step 4: The Render Loop

The render loop runs every frame:

1. Clear the framebuffer
2. Get the window size (using `GetFramebufferSize()` for Retina/HiDPI)
3. Update camera aspect ratio
4. Render the scene

## Step 5: Making It Move

To animate, modify node fields in the render loop:

```go
angle := float32(0)
a.Run(func(rend *renderer.Renderer, deltaTime float64) {
    angle += float32(deltaTime)
    // Find the red box in the g3n scene and rotate it
    // (animation via event routing is the VRML way — see Issue #13)

    a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
    w, h := a.GetFramebufferSize()
    cam.SetAspect(float32(w) / float32(h))
    rend.Render(g3nScene, cam)
})
```

The proper VRML way to animate is via TimeSensor → Interpolator → ROUTE connections. This will be supported when event routing is implemented ([Issue #13](https://github.com/TrueBlocks/trueblocks-3d/issues/13)).

## What You Learned

- Create VRML node trees programmatically in Go
- Convert them to g3n with `converter.Convert()`
- Set up a g3n render loop with camera and lights
- Use `GetFramebufferSize()` for Retina/HiDPI displays
