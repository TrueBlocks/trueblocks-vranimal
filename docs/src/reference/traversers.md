# Traverser Classes

See the [Traversers design chapter](../design/traversers.md) for the full architecture.

## Go Packages: `pkg/traverser`, `pkg/converter`, `pkg/writer`, `pkg/validator`, `pkg/serializer`

| Traverser | Purpose | Go Package | Status |
|-----------|---------|------------|--------|
| **Converter** | VRML → g3n scene objects | `pkg/converter` | Done |
| **ActionTraverser** | Event generation per frame | `pkg/traverser` | Done |
| **PickingTraverser** | Mouse ray-cast hit testing | `pkg/traverser` | Done |
| **WriteTraverser** | Export .wrl (pretty-print) | `pkg/writer` | Done (`vrml-fmt` CLI) |
| **ValidateTraverser** | Scene graph validation | `pkg/validator` | Done (`vrml-validate` CLI) |
| **SerializeTraverser** | Binary scene save/load | `pkg/serializer` | Done (`vrml-serialize` CLI) |
| **OGLTraverser** | OpenGL rendering | — | Replaced by g3n engine |

## Converter API

```go
package converter

// Convert transforms VRML nodes into g3n scene objects under parent.
// Returns a NodeMap for bidirectional VRML↔g3n lookups.
func Convert(vrmlNodes []node.Node, parent *core.Node, baseDir string) *NodeMap

// GetViewpoint returns the first Viewpoint node found, or nil.
func GetViewpoint(nodes []node.Node) *node.Viewpoint

// GetBackground returns the first Background node found, or nil.
func GetBackground(nodes []node.Node) *node.Background
```

## Writer API

```go
package writer

// New creates a Writer that emits VRML97 text to w.
func New(w io.Writer) *Writer

// WriteScene serializes the scene graph as valid .wrl text.
func (wr *Writer) WriteScene(nodes []node.Node)
```

## Validator API

```go
package validator

// New creates a Validator.
func New() *Validator

// Validate checks the scene graph and returns findings.
func (v *Validator) Validate(nodes []node.Node) []Finding
```

## Serializer API

```go
package serializer

// Encode writes the scene graph as binary (gob format).
func Encode(w io.Writer, nodes []node.Node) error

// Decode reads a binary scene graph.
func Decode(r io.Reader) ([]node.Node, error)
```
