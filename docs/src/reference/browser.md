# Browser & Event Classes

See the [Browser & VEM design chapter](../design/browser.md) for the full architecture.

## Go Package: `pkg/browser`

### Browser

Top-level scene container managing the scene graph, clock, routes, and PROTOs.

```go
type Browser struct {
    // Internal fields
}

func (b *Browser) Parse(filename string) ([]node.Node, error)
func (b *Browser) AddRoute(route Route)
func (b *Browser) Tick(deltaTime float64)
```

**Status**: Stub implementation. Full event loop in [Issue #14](https://github.com/TrueBlocks/trueblocks-3d/issues/14).

### Route

Connects an eventOut to an eventIn.

```go
type Route struct {
    FromNode  node.Node
    FromField string
    ToNode    node.Node
    ToField   string
}
```

**Status**: Parsed but not evaluated. [Issue #13](https://github.com/TrueBlocks/trueblocks-3d/issues/13).

### Event

Carries a field change between nodes.

```go
type Event struct {
    Field     string
    Value     interface{}
    Timestamp float64
}
```
