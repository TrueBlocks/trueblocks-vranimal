# Tutorial 1: VRML Pretty Printer (CLI)

This tutorial builds a command-line tool that reads a `.wrl` file and prints a summary of its scene graph. This is equivalent to the original C++ "Console Application — Pretty Printer" tutorial.

## Step 1: Create the Project

```bash
mkdir -p cmd/prettyprint
```

## Step 2: Write the Code

Create `cmd/prettyprint/main.go`:

```go
package main

import (
    "fmt"
    "os"
    "strings"

    "github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
    "github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintf(os.Stderr, "Usage: prettyprint <file.wrl>\n")
        os.Exit(1)
    }

    nodes, err := parser.Parse(os.Args[1])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    for _, n := range nodes {
        printNode(n, 0)
    }
}

func printNode(n node.Node, depth int) {
    indent := strings.Repeat("  ", depth)
    name := n.GetName()
    if name != "" {
        fmt.Printf("%sDEF %s %s\n", indent, name, n.NodeType())
    } else {
        fmt.Printf("%s%s\n", indent, n.NodeType())
    }

    // Print children for grouping nodes
    switch v := n.(type) {
    case *node.Transform:
        for _, child := range v.Children {
            printNode(child, depth+1)
        }
    case *node.Group:
        for _, child := range v.Children {
            printNode(child, depth+1)
        }
    case *node.Shape:
        if v.Appearance != nil {
            printNode(v.Appearance, depth+1)
        }
        if v.Geometry != nil {
            printNode(v.Geometry, depth+1)
        }
    }
}
```

## Step 3: Build and Run

```bash
go build -o prettyprint ./cmd/prettyprint/
./prettyprint examples/test_scene.wrl
```

Output:

```
Viewpoint
Background
DirectionalLight
PointLight
DEF RedBox Transform
  Shape
    Appearance
      Material
    Box
DEF GreenSphere Transform
  Shape
    Appearance
      Material
    Sphere
...
```

## Step 4: Export Options

The C++ WriteTraverser supported options like:
- Writing only non-default fields (smaller output)
- Indentation control
- DEF/USE preservation

When the WriteTraverser is ported ([Issue #9](https://github.com/TrueBlocks/trueblocks-3d/issues/9)), the pretty printer can produce valid `.wrl` output that round-trips through parse → write.

## What You Learned

- How to use `parser.Parse()` to load a `.wrl` file
- The scene graph is a tree of `node.Node` values
- Type assertions (`switch v := n.(type)`) access node-specific fields
- Grouping nodes (Transform, Group) contain `Children` slices
