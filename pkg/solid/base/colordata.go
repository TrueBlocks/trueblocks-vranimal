package base

import "github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"

// ColorData holds per-vertex or per-face color, normal, and texture coordinate data.
type ColorData struct {
	Type     int64
	Color    vec.SFColor
	Normal   vec.SFVec3f
	TexCoord vec.SFVec2f
}

// SetColor sets the color and marks the color flag.
func (d *ColorData) SetColor(c vec.SFColor) {
	d.Type |= ColorFlag
	d.Color = c
}

// SetNormal sets the normal and marks the normal flag.
func (d *ColorData) SetNormal(n vec.SFVec3f) {
	d.Type |= NormalFlag
	d.Normal = n
}

// SetTexCoord sets the texture coordinate and marks the texcoord flag.
func (d *ColorData) SetTexCoord(t vec.SFVec2f) {
	d.Type |= TexCoordFlag
	d.TexCoord = t
}

// GetColor returns the color (caller should check HasColor).
func (d *ColorData) GetColor() vec.SFColor { return d.Color }

// GetNormal returns the normal (caller should check HasNormal).
func (d *ColorData) GetNormal() vec.SFVec3f { return d.Normal }

// GetTexCoord returns the texture coordinate (caller should check HasTexCoord).
func (d *ColorData) GetTexCoord() vec.SFVec2f { return d.TexCoord }

// HasColor returns true if color data is present.
func (d *ColorData) HasColor() bool { return d.Type&ColorFlag != 0 }

// HasNormal returns true if normal data is present.
func (d *ColorData) HasNormal() bool { return d.Type&NormalFlag != 0 }

// HasTexCoord returns true if texture coordinate data is present.
func (d *ColorData) HasTexCoord() bool { return d.Type&TexCoordFlag != 0 }
