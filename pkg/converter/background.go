package converter

import (
	"math"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func convertBackground(bg *node.Background, parent *core.Node, baseDir string) {
	if bg == nil {
		return
	}

	if len(bg.BackURL) > 0 || len(bg.FrontURL) > 0 || len(bg.LeftURL) > 0 ||
		len(bg.RightURL) > 0 || len(bg.TopURL) > 0 || len(bg.BottomURL) > 0 {
		convertSkybox(bg, parent, baseDir)
	}

	if len(bg.SkyColor) > 1 || len(bg.GroundColor) > 0 {
		convertSkyGradient(bg, parent)
	}
}

func firstURL(urls []string) string {
	if len(urls) > 0 {
		return urls[0]
	}
	return ""
}

func convertSkybox(bg *node.Background, parent *core.Node, baseDir string) {
	urls := [6]string{
		firstURL(bg.RightURL),
		firstURL(bg.LeftURL),
		firstURL(bg.TopURL),
		firstURL(bg.BottomURL),
		firstURL(bg.FrontURL),
		firstURL(bg.BackURL),
	}

	geom := geometry.NewCube(50)
	skyNode := core.NewNode()
	skyNode.SetName("__skybox__")

	for i, url := range urls {
		if url == "" {
			continue
		}
		tex := loadTexture(url, baseDir)
		if tex == nil {
			continue
		}
		mat := material.NewStandard(math32.NewColor("white"))
		mat.AddTexture(tex)
		mat.SetSide(material.SideBack)
		mat.SetUseLights(material.UseLightNone)
		mat.SetDepthMask(false)

		mesh := graphic.NewMesh(geom, nil)
		mesh.AddGroupMaterial(mat, i)
		skyNode.Add(mesh)
	}

	parent.Add(skyNode)
}

func convertSkyGradient(bg *node.Background, parent *core.Node) {
	const segments = 32
	const rings = 16
	radius := float32(45)

	nVerts := (rings + 1) * (segments + 1)
	positions := math32.NewArrayF32(0, nVerts*3)
	colors := math32.NewArrayF32(0, nVerts*3)
	indices := math32.NewArrayU32(0, rings*segments*6)

	for r := 0; r <= rings; r++ {
		theta := float64(r) * math.Pi / float64(rings)
		sinT := float32(math.Sin(theta))
		cosT := float32(math.Cos(theta))

		c := skyColorAtAngle(bg, theta)

		for s := 0; s <= segments; s++ {
			phi := float64(s) * 2.0 * math.Pi / float64(segments)
			x := radius * sinT * float32(math.Cos(phi))
			y := radius * cosT
			z := radius * sinT * float32(math.Sin(phi))
			positions.Append(x, y, z)
			colors.Append(c.R, c.G, c.B)
		}
	}

	for r := 0; r < rings; r++ {
		for s := 0; s < segments; s++ {
			a := uint32(r*(segments+1) + s)
			b := uint32(a + uint32(segments+1))
			indices.Append(a, b, a+1)
			indices.Append(b, b+1, a+1)
		}
	}

	geom := geometry.NewGeometry()
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(colors).AddAttrib(gls.VertexColor))
	geom.SetIndices(indices)

	mat := material.NewBasic()
	mat.SetSide(material.SideBack)
	mat.SetUseLights(material.UseLightNone)
	mat.SetDepthMask(false)

	mesh := graphic.NewMesh(geom, mat)
	mesh.SetName("__sky_gradient__")
	mesh.SetRenderOrder(-1000)
	parent.Add(mesh)
}

func skyColorAtAngle(bg *node.Background, theta float64) math32.Color {
	if theta <= math.Pi/2 {
		angle := theta
		return interpolateColorBand(bg.SkyColor, bg.SkyAngle, angle)
	}
	if len(bg.GroundColor) > 0 {
		angle := math.Pi - theta
		return interpolateColorBand(bg.GroundColor, bg.GroundAngle, angle)
	}
	if len(bg.SkyColor) > 0 {
		return interpolateColorBand(bg.SkyColor, bg.SkyAngle, theta)
	}
	return math32.Color{R: 0, G: 0, B: 0}
}

func interpolateColorBand(colors []vec.SFColor, angles []float64, angle float64) math32.Color {
	if len(colors) == 0 {
		return math32.Color{R: 0, G: 0, B: 0}
	}
	if len(colors) == 1 {
		return math32.Color{R: float32(colors[0].R), G: float32(colors[0].G), B: float32(colors[0].B)}
	}

	for i := 0; i < len(angles); i++ {
		if angle <= angles[i] {
			prevAngle := 0.0
			if i > 0 {
				prevAngle = angles[i-1]
			}
			t := float32(0)
			span := angles[i] - prevAngle
			if span > 0 {
				t = float32((angle - prevAngle) / span)
			}
			c0 := colors[i]
			c1 := colors[i+1]
			return math32.Color{
				R: float32(c0.R)*(1-t) + float32(c1.R)*t,
				G: float32(c0.G)*(1-t) + float32(c1.G)*t,
				B: float32(c0.B)*(1-t) + float32(c1.B)*t,
			}
		}
	}

	last := colors[len(colors)-1]
	return math32.Color{R: float32(last.R), G: float32(last.G), B: float32(last.B)}
}



// ApplyFog applies VRML fog effect by returning fog parameters for the viewer
// to use with distance-based alpha blending. The g3n engine does not have
// built-in fog, so we expose the parameters for the viewer's render loop.
type FogParams struct {
	Enabled         bool
	Color           math32.Color
	VisibilityRange float32
	Exponential     bool
}

func GetFogParams(nodes []node.Node) *FogParams {
	fg := GetFog(nodes)
	if fg == nil || fg.VisibilityRange <= 0 {
		return nil
	}
	return &FogParams{
		Enabled:         true,
		Color:           math32.Color{R: float32(fg.Color.R), G: float32(fg.Color.G), B: float32(fg.Color.B)},
		VisibilityRange: float32(fg.VisibilityRange),
		Exponential:     fg.FogType == "EXPONENTIAL",
	}
}
