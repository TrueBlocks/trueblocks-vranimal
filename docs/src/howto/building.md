# Building from Source

## Prerequisites

- **Go 1.21+** (uses `min`/`max` builtins)
- **C compiler** (for CGO — OpenGL, audio)
- **macOS**: Xcode command line tools
- **Linux**: `build-essential`, `libgl1-mesa-dev`, `libopenal-dev`, `libvorbis-dev`

## macOS (Homebrew)

```bash
# Install audio dependencies
brew install openal-soft libogg libvorbis

# Clone
git clone https://github.com/TrueBlocks/trueblocks-vranimal.git
cd trueblocks-vranimal

# Build the viewer
CGO_CFLAGS="-I/opt/homebrew/include" \
CGO_LDFLAGS="-L/opt/homebrew/lib" \
go build -o viewer ./cmd/viewer/

# Run
./viewer examples/test_scene.wrl
```

## Linux (apt)

```bash
# Install dependencies
sudo apt install build-essential libgl1-mesa-dev libopenal-dev libvorbis-dev

# Clone and build
git clone https://github.com/TrueBlocks/trueblocks-vranimal.git
cd trueblocks-vranimal
go build -o viewer ./cmd/viewer/
```

## Build Just the Libraries (no CGO)

The core packages (`pkg/node`, `pkg/parser`, `pkg/vec`, `pkg/solid`, `pkg/types`, `pkg/geo`) are pure Go:

```bash
go build ./pkg/...
go test ./pkg/...
```

Only `pkg/converter` and `cmd/viewer` require CGO (g3n engine → OpenGL).

## Running Tests

```bash
go test ./pkg/...
```

## Building the Documentation

```bash
cd docs
mdbook build
# Output in docs/book/
mdbook serve  # local preview at http://localhost:3000
```

## What Ships with Your Application

The viewer binary is self-contained (statically linked Go binary with CGO). Ship:

1. The `viewer` binary
2. Any `.wrl` files your application needs
3. Any texture images referenced by the `.wrl` files

No DLLs, shared libraries, or runtime dependencies beyond system OpenGL and OpenAL.
