# VRaniML — Go VRML97 Toolkit

VRaniML is a collection of Go packages that implement the [VRML97 specification](https://www.web3d.org/documents/specifications/14772/V2.0/index.html) for 3D scene graph manipulation and rendering. The toolkit uses the [g3n engine](https://github.com/g3n/engine) for OpenGL-based 3D rendering.

The library is organized into three layers that can be used independently or together:

| Layer | Packages | Description |
|-------|----------|-------------|
| **Utility** | `pkg/vec`, `pkg/types`, `pkg/geo` | Vector math, bounding boxes, base types |
| **Solid Modeling** | `pkg/solid` | Half-edge B-rep data structures, Euler operators |
| **VRML97** | `pkg/node`, `pkg/parser`, `pkg/browser`, `pkg/converter`, `pkg/traverser` | Full VRML97 node types, parser, scene graph conversion, rendering |

## Quick Start

```bash
# Install dependencies (macOS)
brew install openal-soft libogg libvorbis

# Build the viewer
CGO_CFLAGS="-I/opt/homebrew/include" \
CGO_LDFLAGS="-L/opt/homebrew/lib" \
go build -o viewer ./cmd/viewer/

# View a VRML file
./viewer examples/test_scene.wrl
```

## What is VRML97?

VRML97 (Virtual Reality Modeling Language, ISO/IEC 14772-1:1997) is a standard file format for representing 3D interactive vector graphics. A `.wrl` file describes:

- **Geometry**: Boxes, spheres, cones, cylinders, indexed face sets, elevation grids, extrusions
- **Appearance**: Materials (diffuse/specular/emissive colors), textures, texture transforms
- **Scene Graph**: Hierarchical transforms, groups, switches, level-of-detail
- **Interaction**: Sensors (touch, proximity, time), interpolators, event routing (ROUTEs)
- **Environment**: Lights (directional, point, spot), viewpoints, backgrounds, fog, audio

## Architecture

```text
┌─────────────────────────────────────────────────────┐
│                    cmd/viewer                        │
│              (g3n application loop)                   │
├─────────────────────────────────────────────────────┤
│                   pkg/converter                      │
│          (VRML node → g3n scene objects)              │
├──────────┬──────────┬──────────┬────────────────────┤
│ pkg/node │pkg/parser│pkg/browser│   pkg/traverser    │
│  (types) │ (lexer+  │ (event   │   (scene graph     │
│          │  parser) │  loop)   │    walkers)        │
├──────────┴──────────┴──────────┴────────────────────┤
│                    pkg/solid                          │
│          (half-edge B-rep, Euler ops)                 │
├─────────────────────────────────────────────────────┤
│              pkg/vec  ·  pkg/geo                     │
│           (math types · bounding box)                │
└─────────────────────────────────────────────────────┘
```

## Origins

This codebase is a Go port of the original [VRaniML C++ SDK](https://github.com/TrueBlocks/vraniml), a collection of three C++ class libraries for Windows that implemented VRML97 functionality. The original was developed by Great Hill Corporation circa 1997-1999. The Go port modernizes the API while preserving the architectural insights of the original design.
