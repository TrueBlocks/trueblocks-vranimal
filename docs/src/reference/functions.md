# Utility Functions

Common utility functions available across the library.

## Math Functions

Most C++ utility functions map to Go builtins or stdlib:

| Function | Go Equivalent | Description |
|----------|---------------|-------------|
| `Min(a, b)` | `min(a, b)` | Built-in since Go 1.21 |
| `Max(a, b)` | `max(a, b)` | Built-in since Go 1.21 |
| `Clamp(val, lo, hi)` | `max(lo, min(hi, val))` | Clamp to range |
| `InRange(val, lo, hi)` | `val >= lo && val <= hi` | Range check |
| `Deg2Rad(v)` | `v * math.Pi / 180` | Degrees to radians |
| `Rad2Deg(v)` | `v * 180 / math.Pi` | Radians to degrees |
| `Equals(a, b, tol)` | `math.Abs(a-b) < tol` | Tolerant float compare |

## Interpolation

```go
func Interpolate(from, to T, fromKey, toKey, t float32) T
```

For any type supporting `+`, `-`, `*` (all SF/MF types), returns linear interpolation at position `t` between key frames `fromKey` and `toKey`:

```
ratio = (t - fromKey) / (toKey - fromKey)
result = from + (to - from) * ratio
```

## Image Loading

```go
// Go stdlib handles all image formats:
import "image/jpeg"
import "image/png"

f, _ := os.Open("texture.jpg")
img, _, _ := image.Decode(f)
```

The C++ library had 55 files for BMP/JPEG/PNG codecs. Go's `image` stdlib package replaces all of them.

## File Fetching

```go
// C++: vrCacheFile(remoteURL)
// Go: standard HTTP client
resp, _ := http.Get(url)
defer resp.Body.Close()
data, _ := io.ReadAll(resp.Body)
```

## Scene Graph Traversal

```go
// Walk every node in the tree
func ForEvery(root node.Node, fn func(node.Node) bool)

// Find first node of a given type
func FindByType[T node.Node](root node.Node) T

// Find first node with a given DEF name
func FindByName(root node.Node, name string) node.Node
```
