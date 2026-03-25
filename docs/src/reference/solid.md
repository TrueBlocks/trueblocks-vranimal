# Solid Modeling Classes

The solid modeling library implements the half-edge boundary representation (B-rep). All geometry nodes use this as their internal data structure.

See the [Solid Modeling design chapter](../design/solid.md) for the full architecture.

## Go Package: `pkg/solid`

### Solid

```go
type Solid struct {
    Faces    []*Face
    Edges    []*Edge
    Vertices []*Vertex
}
```

### Face

```go
type Face struct {
    OuterLoop *Loop
    Loops     []*Loop   // outer + holes
}
```

### Edge

```go
type Edge struct {
    He1, He2 *HalfEdge  // two half-edges, one per adjacent face
}
```

### HalfEdge

```go
type HalfEdge struct {
    Vertex *Vertex
    Edge   *Edge
    Loop   *Loop
    Next   *HalfEdge
    Prev   *HalfEdge
    Mate   *HalfEdge  // opposite half-edge on same edge
}
```

### Loop

```go
type Loop struct {
    HalfEdge *HalfEdge  // start of cyclic list
    Face     *Face
}
```

### Vertex

```go
type Vertex struct {
    Position vec.SFVec3f
    HalfEdge *HalfEdge  // one incident half-edge
}
```

## Euler Operators

| Function | Description |
|----------|-------------|
| `Mvfs(pos)` | Make vertex, face, solid — bootstrap |
| `Lmev(he, pos)` | Split vertex, create edge |
| `Lmef(he1, he2)` | Split face, create edge |
| `Lkev(he)` | Kill edge, merge vertices |
| `Lkef(he)` | Kill edge, merge faces |
| `Lkemr(he)` | Kill edge, make hole ring |
| `Lmekr(he1, he2)` | Make edge, kill hole ring |
