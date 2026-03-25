# Extending the Library

VRaniML is designed to be extended. Here are common extension patterns.

## Adding a Custom Node Type

To add a node type not in the VRML97 spec:

1. Define the Go struct implementing `node.Node`:

```go
package node

type MyCustomNode struct {
    BaseNode
    CustomField float32
    Children    []Node
}

func (n *MyCustomNode) NodeType() string { return "MyCustomNode" }
```

2. Register it in the parser if you want to read it from `.wrl` files (via PROTO).

3. Add a conversion case in `pkg/converter/converter.go` to map it to g3n objects.

## Adding Fields to Existing Nodes

Add new fields to the Go struct directly. Update the parser to read them from `.wrl` files and the converter to use them during rendering.

## Using the Node Library Without Rendering

The `pkg/node` and `pkg/parser` packages have no dependency on g3n or OpenGL. You can use them for:

- VRML file format validation
- Scene graph analysis and statistics
- Format conversion (parse VRML, output JSON/etc.)
- Programmatic scene generation

```go
import (
    "github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
)

nodes, err := parser.Parse("scene.wrl")
// Work with nodes without any rendering
```

## Using the Solid Library Standalone

The `pkg/solid` package is independent. You can use it for:

- Custom geometry processing
- Boolean operations (when ported)
- Mesh analysis

```go
import "github.com/TrueBlocks/trueblocks-vranimal/pkg/solid"

s := solid.Mvfs(vec.SFVec3f{0, 0, 0})
// Build geometry using Euler operators
```
