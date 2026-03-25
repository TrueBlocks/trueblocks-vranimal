# Appearance Nodes

These nodes control the visual appearance of geometry.

## Appearance

Pairs a Material with optional Texture and TextureTransform.

```go
type Appearance struct {
    Material         *Material
    Texture          Node            // ImageTexture, PixelTexture, or MovieTexture
    TextureTransform *TextureTransform
}
```

```vrml
Shape {
    appearance Appearance {
        material Material { diffuseColor 1 0 0 }
        texture ImageTexture { url "brick.jpg" }
    }
    geometry Box {}
}
```

## Material

Defines surface color and lighting properties.

```go
type Material struct {
    AmbientIntensity float32    // 0.0–1.0, default 0.2
    DiffuseColor     SFColor    // default {0.8, 0.8, 0.8}
    EmissiveColor    SFColor    // default {0, 0, 0}
    Shininess        float32    // 0.0–1.0, default 0.2
    SpecularColor    SFColor    // default {0, 0, 0}
    Transparency     float32    // 0.0=opaque, 1.0=invisible, default 0.0
}
```

## ImageTexture

Loads a texture image from a URL or file path.

```go
type ImageTexture struct {
    URL     []string  // list of URLs to try
    RepeatS bool      // repeat horizontally (default true)
    RepeatT bool      // repeat vertically (default true)
}
```

**Status**: Image loading tracked in [Issue #17](https://github.com/TrueBlocks/trueblocks-3d/issues/17).

## PixelTexture

Defines a texture from inline pixel data.

```go
type PixelTexture struct {
    Image   SFImage   // width, height, components, pixel data
    RepeatS bool
    RepeatT bool
}
```

## MovieTexture

Animated video texture.

```go
type MovieTexture struct {
    URL       []string
    Loop      bool
    Speed     float32   // default 1.0
    StartTime float64
    StopTime  float64
    RepeatS   bool
    RepeatT   bool
}
```

## FontStyle

Controls font rendering for Text nodes.

```go
type FontStyle struct {
    Family     []string  // "SERIF", "SANS", "TYPEWRITER"
    Style      string    // "PLAIN", "BOLD", "ITALIC", "BOLDITALIC"
    Size       float32   // default 1.0
    Spacing    float32   // line spacing, default 1.0
    Justify    []string  // "BEGIN", "MIDDLE", "END"
    Language   string
    Horizontal bool      // default true
    LeftToRight bool     // default true
    TopToBottom bool     // default true
}
```

## TextureTransform

2D transformation of texture coordinates.

```go
type TextureTransform struct {
    Center      SFVec2f   // default {0, 0}
    Rotation    float32   // radians, default 0
    Scale       SFVec2f   // default {1, 1}
    Translation SFVec2f   // default {0, 0}
}
```
