# Traverser Classes

See the [Traversers design chapter](../design/traversers.md) for the full architecture.

## Go Packages: `pkg/traverser`, `pkg/converter`

| Traverser | Purpose | Go Status |
|-----------|---------|----------| 
| **Converter** | VRML → g3n scene objects | Working |
| **ActionTraverser** | Event generation per frame | [Issue #10](https://github.com/TrueBlocks/trueblocks-3d/issues/10) |
| **WriteTraverser** | Export .wrl (pretty-print) | [Issue #9](https://github.com/TrueBlocks/trueblocks-3d/issues/9) |
| **ValidateTraverser** | Scene graph validation | [Issue #8](https://github.com/TrueBlocks/trueblocks-3d/issues/8) |
| **SerializeTraverser** | Binary scene save/load | [Issue #7](https://github.com/TrueBlocks/trueblocks-3d/issues/7) |
| **OGLTraverser** | OpenGL rendering | g3n handles this ([Issue #6](https://github.com/TrueBlocks/trueblocks-3d/issues/6)) |

## Converter API

```go
package converter

// Convert transforms VRML nodes into a g3n scene graph.
func Convert(nodes []node.Node) *core.Node

// GetViewpoint returns the first Viewpoint node found, or nil.
func GetViewpoint(nodes []node.Node) *node.Viewpoint

// GetBackground returns the first Background node found, or nil.
func GetBackground(nodes []node.Node) *node.Background
```
