# Browser & Event Classes

See the [Browser & VEM design chapter](../design/browser.md) for the full architecture.

## Go Package: `pkg/browser`

### Browser

Top-level scene container managing traversers, binding stacks (viewpoint, background, fog, navigation info), routes, and the frame clock.

```go
type Browser struct {
    // Internal fields
}

func NewBrowser() *Browser
func (b *Browser) SetFrameRate(fps int)
func (b *Browser) AddTraverser(t Traverser)
func (b *Browser) GetTraverser(i int) Traverser
func (b *Browser) NTraversers() int
func (b *Browser) BindViewpoint(vp *node.Viewpoint, bind bool)
func (b *Browser) GetViewpoint() *node.Viewpoint
func (b *Browser) BindNavigationInfo(ni *node.NavigationInfo, bind bool)
func (b *Browser) GetNavigationInfo() *node.NavigationInfo
func (b *Browser) BindBackground(bg *node.Background, bind bool)
func (b *Browser) GetBackground() *node.Background
func (b *Browser) BindFog(fg *node.Fog, bind bool)
func (b *Browser) GetFog() *node.Fog
func (b *Browser) AddRoute(src node.Node, srcField string, dst node.Node, dstField string) *node.Route
func (b *Browser) Clear()
func (b *Browser) Update(dt time.Duration)
```

### Route

Connects an eventOut to an eventIn. Defined in `pkg/node`.

```go
type Route struct {
    Source      Node
    SrcField    string
    Destination Node
    DstField    string
    Internal    bool
    RouteID     int64
}
```

### Event

Carries a field change between nodes.

```go
type Event struct {
    Field     string
    Value     interface{}
    Timestamp float64
}
```
