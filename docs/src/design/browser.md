# Browser & Execution Model (VEM)

The VRML Execution Model (VEM) is the runtime environment that drives VRML scene graph animation and interaction. It provides:

- **Parsing**: Reading `.wrl` files and building the scene graph
- **Event routing**: Propagating events along ROUTE connections
- **Clock management**: Maintaining a consistent frame rate
- **PROTO management**: Supporting user-defined node types
- **Bindable node stacks**: Managing Viewpoint, Background, NavigationInfo, Fog

## The Browser

The browser is the top-level container for the scene graph. It maintains:

1. **The scene graph** — a tree of VRML nodes rooted at one or more top-level Group/Transform nodes
2. **The VRML clock** — time management for sustained frame rate rendering
3. **A list of traversers** — supporting multi-pass rendering (e.g., action pass then render pass)
4. **Named node table** — DEF/USE name resolution
5. **ROUTE table** — event connections between nodes
6. **PROTO table** — user-defined node prototypes

### Go Package: `pkg/browser`

```go
type Browser struct {
    // Scene graph root nodes
    // Named node DEF/USE table
    // ROUTE connections
    // PROTO definitions
}
```

## Event Routing

VRML97's `ROUTE` statement connects an `eventOut` of one node to an `eventIn` of another:

```vrml
ROUTE TimeSensor1.fraction_changed TO Interpolator1.set_fraction
ROUTE Interpolator1.value_changed TO Transform1.set_translation
```

At each frame:
1. The **ActionTraverser** fires initial events (e.g., TimeSensor generates `fraction_changed`)
2. Events propagate along ROUTEs using the **Event** object
3. Receiving nodes update their fields and may generate cascading events
4. Cascading continues until no more events are pending
5. The **RenderTraverser** draws the updated scene

### Go Package: `pkg/parser` (parsing ROUTEs), `pkg/browser` (evaluating ROUTEs)

The parser currently reads ROUTE statements. Runtime evaluation is tracked in [Issue #13](https://github.com/TrueBlocks/trueblocks-3d/issues/13).

## The Event Class

An event carries:
- The **field** being changed
- The **new value** for that field
- The **timestamp** of the event

## The Route Class

A route stores:
- **Source node** and **eventOut field name**
- **Destination node** and **eventIn field name**

## The Parser

The VRML parser reads `.wrl` text files and constructs the in-memory scene graph.

### Go Package: `pkg/parser`

```go
func Parse(filename string) ([]node.Node, error)
```

The parser handles:
- Lexical scanning (numbers, strings, identifiers, brackets)
- Node instantiation for all 54 VRML97 node types
- `DEF` / `USE` name binding
- `ROUTE` statement parsing
- `PROTO` / `EXTERNPROTO` declaration recognition

**Current limitations**:
- PROTO bodies are recognized but not instantiated ([Issue #11](https://github.com/TrueBlocks/trueblocks-3d/issues/11))
- EXTERNPROTO URLs are not fetched ([Issue #12](https://github.com/TrueBlocks/trueblocks-3d/issues/12))
- ROUTEs are parsed but not evaluated at runtime ([Issue #13](https://github.com/TrueBlocks/trueblocks-3d/issues/13))

## UserMessage Abstraction

The original C++ library abstracted OS messages (mouse clicks, key presses) through a `UserMessage` class to ease cross-platform porting. In Go, the g3n engine provides platform-independent input handling via its `window` package, making a custom abstraction unnecessary.
