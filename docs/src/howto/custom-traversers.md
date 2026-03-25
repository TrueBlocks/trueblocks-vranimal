# Writing Custom Traversers

A traverser walks the scene graph performing an operation at each node. You can write your own.

## Pattern

```go
func MyTraverser(root node.Node) {
    walk(root, 0)
}

func walk(n node.Node, depth int) {
    // Pre-visit: do something with this node
    process(n, depth)

    // Visit children
    switch v := n.(type) {
    case *node.Transform:
        for _, child := range v.Children {
            walk(child, depth+1)
        }
    case *node.Group:
        for _, child := range v.Children {
            walk(child, depth+1)
        }
    case *node.Shape:
        if v.Appearance != nil {
            walk(v.Appearance, depth+1)
        }
        if v.Geometry != nil {
            walk(v.Geometry, depth+1)
        }
    case *node.Switch:
        if v.WhichChoice >= 0 && int(v.WhichChoice) < len(v.Choice) {
            walk(v.Choice[v.WhichChoice], depth+1)
        }
    // ... other grouping nodes
    }

    // Post-visit: pop state if needed
}
```

## Example: Node Counter

```go
func CountNodes(roots []node.Node) map[string]int {
    counts := make(map[string]int)
    for _, root := range roots {
        countWalk(root, counts)
    }
    return counts
}

func countWalk(n node.Node, counts map[string]int) {
    counts[n.NodeType()]++
    // recurse into children...
}
```

## Example: Bounding Box Calculator

```go
func ComputeBounds(roots []node.Node) geo.BoundingBox {
    var bb geo.BoundingBox
    // Walk tree, accumulate transforms, expand bounding box
    // for each geometry node encountered
    return bb
}
```

## Transform Stack

When traversing, maintain a transform stack to track the current coordinate space:

```go
type TransformStack struct {
    stack []vec.Matrix
}

func (s *TransformStack) Push(m vec.Matrix) { s.stack = append(s.stack, m) }
func (s *TransformStack) Pop()              { s.stack = s.stack[:len(s.stack)-1] }
func (s *TransformStack) Current() vec.Matrix { return s.stack[len(s.stack)-1] }
```
