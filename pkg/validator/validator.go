package validator

import (
	"fmt"
	"math"
	"strings"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// Severity indicates the importance of a validation finding.
type Severity int

const (
	Warning Severity = iota
	Error
)

func (s Severity) String() string {
	if s == Error {
		return "ERROR"
	}
	return "WARN"
}

// Finding represents a single validation issue found in the scene graph.
type Finding struct {
	Severity Severity
	NodeName string
	NodeType string
	Path     string
	Message  string
}

func (f Finding) String() string {
	parts := []string{f.Severity.String()}
	if f.Path != "" {
		parts = append(parts, f.Path)
	}
	if f.NodeType != "" {
		parts = append(parts, f.NodeType)
	}
	if f.NodeName != "" {
		parts = append(parts, fmt.Sprintf("(%s)", f.NodeName))
	}
	parts = append(parts, f.Message)
	return strings.Join(parts, ": ")
}

// Validator walks a VRML97 scene graph and checks for structural
// and value-range errors per the VRML97 specification.
type Validator struct {
	findings []Finding
	path     []string
}

// New creates a Validator.
func New() *Validator {
	return &Validator{}
}

// Validate checks all top-level nodes and returns any findings.
func (v *Validator) Validate(nodes []node.Node) []Finding {
	v.findings = nil
	for _, n := range nodes {
		v.validateNode(n)
	}
	return v.findings
}

// Findings returns the accumulated findings.
func (v *Validator) Findings() []Finding {
	return v.findings
}

func (v *Validator) currentPath() string {
	return strings.Join(v.path, "/")
}

func (v *Validator) pushPath(s string) {
	v.path = append(v.path, s)
}

func (v *Validator) popPath() {
	if len(v.path) > 0 {
		v.path = v.path[:len(v.path)-1]
	}
}

func (v *Validator) addError(nodeType, nodeName, msg string, args ...any) {
	v.findings = append(v.findings, Finding{
		Severity: Error,
		NodeType: nodeType,
		NodeName: nodeName,
		Path:     v.currentPath(),
		Message:  fmt.Sprintf(msg, args...),
	})
}

func (v *Validator) addWarning(nodeType, nodeName, msg string, args ...any) {
	v.findings = append(v.findings, Finding{
		Severity: Warning,
		NodeType: nodeType,
		NodeName: nodeName,
		Path:     v.currentPath(),
		Message:  fmt.Sprintf(msg, args...),
	})
}

func nodeLabel(n node.Node) (string, string) {
	switch n.(type) {
	case *node.Appearance:
		return "Appearance", n.GetName()
	case *node.Material:
		return "Material", n.GetName()
	case *node.ImageTexture:
		return "ImageTexture", n.GetName()
	case *node.MovieTexture:
		return "MovieTexture", n.GetName()
	case *node.PixelTexture:
		return "PixelTexture", n.GetName()
	case *node.TextureTransform:
		return "TextureTransform", n.GetName()
	case *node.FontStyle:
		return "FontStyle", n.GetName()
	case *node.Background:
		return "Background", n.GetName()
	case *node.Fog:
		return "Fog", n.GetName()
	case *node.NavigationInfo:
		return "NavigationInfo", n.GetName()
	case *node.Viewpoint:
		return "Viewpoint", n.GetName()
	case *node.Shape:
		return "Shape", n.GetName()
	case *node.DirectionalLight:
		return "DirectionalLight", n.GetName()
	case *node.PointLight:
		return "PointLight", n.GetName()
	case *node.SpotLight:
		return "SpotLight", n.GetName()
	case *node.WorldInfo:
		return "WorldInfo", n.GetName()
	case *node.Script:
		return "Script", n.GetName()
	case *node.Sound:
		return "Sound", n.GetName()
	case *node.AudioClip:
		return "AudioClip", n.GetName()
	case *node.Transform:
		return "Transform", n.GetName()
	case *node.Group:
		return "Group", n.GetName()
	case *node.Anchor:
		return "Anchor", n.GetName()
	case *node.Billboard:
		return "Billboard", n.GetName()
	case *node.Collision:
		return "Collision", n.GetName()
	case *node.Inline:
		return "Inline", n.GetName()
	case *node.LOD:
		return "LOD", n.GetName()
	case *node.Switch:
		return "Switch", n.GetName()
	case *node.Box:
		return "Box", n.GetName()
	case *node.Cone:
		return "Cone", n.GetName()
	case *node.Cylinder:
		return "Cylinder", n.GetName()
	case *node.Sphere:
		return "Sphere", n.GetName()
	case *node.Extrusion:
		return "Extrusion", n.GetName()
	case *node.Text:
		return "Text", n.GetName()
	case *node.IndexedFaceSet:
		return "IndexedFaceSet", n.GetName()
	case *node.IndexedLineSet:
		return "IndexedLineSet", n.GetName()
	case *node.PointSet:
		return "PointSet", n.GetName()
	case *node.ElevationGrid:
		return "ElevationGrid", n.GetName()
	case *node.ColorNode:
		return "Color", n.GetName()
	case *node.Coordinate:
		return "Coordinate", n.GetName()
	case *node.NormalNode:
		return "Normal", n.GetName()
	case *node.TextureCoordinate:
		return "TextureCoordinate", n.GetName()
	case *node.ColorInterpolator:
		return "ColorInterpolator", n.GetName()
	case *node.CoordinateInterpolator:
		return "CoordinateInterpolator", n.GetName()
	case *node.NormalInterpolator:
		return "NormalInterpolator", n.GetName()
	case *node.OrientationInterpolator:
		return "OrientationInterpolator", n.GetName()
	case *node.PositionInterpolator:
		return "PositionInterpolator", n.GetName()
	case *node.ScalarInterpolator:
		return "ScalarInterpolator", n.GetName()
	case *node.TouchSensor:
		return "TouchSensor", n.GetName()
	case *node.TimeSensor:
		return "TimeSensor", n.GetName()
	case *node.ProximitySensor:
		return "ProximitySensor", n.GetName()
	case *node.CylinderSensor:
		return "CylinderSensor", n.GetName()
	case *node.PlaneSensor:
		return "PlaneSensor", n.GetName()
	case *node.SphereSensor:
		return "SphereSensor", n.GetName()
	case *node.VisibilitySensor:
		return "VisibilitySensor", n.GetName()
	default:
		return fmt.Sprintf("%T", n), n.GetName()
	}
}

// validateNode dispatches to type-specific validation.
func (v *Validator) validateNode(n node.Node) {
	if n == nil {
		return
	}
	nType, nName := nodeLabel(n)
	label := nType
	if nName != "" {
		label = nName + ":" + nType
	}
	v.pushPath(label)
	defer v.popPath()

	switch nd := n.(type) {
	case *node.Appearance:
		v.validateAppearance(nd)
	case *node.Material:
		v.validateMaterial(nd)
	case *node.SpotLight:
		v.validateSpotLight(nd)
	case *node.Shape:
		v.validateShape(nd)
	case *node.Box:
		v.validateBox(nd)
	case *node.Cone:
		v.validateCone(nd)
	case *node.Cylinder:
		v.validateCylinder(nd)
	case *node.Sphere:
		v.validateSphere(nd)
	case *node.ElevationGrid:
		v.validateElevationGrid(nd)
	case *node.IndexedFaceSet:
		v.validateIndexedFaceSet(nd)
	case *node.IndexedLineSet:
		v.validateIndexedLineSet(nd)
	case *node.PointSet:
		v.validatePointSet(nd)
	case *node.ColorInterpolator:
		v.validateColorInterpolator(nd)
	case *node.CoordinateInterpolator:
		v.validateCoordinateInterpolator(nd)
	case *node.NormalInterpolator:
		v.validateNormalInterpolator(nd)
	case *node.OrientationInterpolator:
		v.validateOrientationInterpolator(nd)
	case *node.PositionInterpolator:
		v.validatePositionInterpolator(nd)
	case *node.ScalarInterpolator:
		v.validateScalarInterpolator(nd)
	case *node.Viewpoint:
		v.validateViewpoint(nd)
	case *node.Inline:
		v.validateInline(nd)
		v.validateChildren(nd.Children)
		return
	case *node.Transform:
		v.validateTransform(nd)
		v.validateChildren(nd.Children)
		return
	case *node.Group:
		v.validateChildren(nd.Children)
		return
	case *node.Anchor:
		v.validateChildren(nd.Children)
		return
	case *node.Billboard:
		v.validateChildren(nd.Children)
		return
	case *node.Collision:
		v.validateChildren(nd.Children)
		return
	case *node.LOD:
		v.validateLOD(nd)
		return
	case *node.Switch:
		v.validateSwitch(nd)
		return
	}
}

func (v *Validator) validateChildren(children []node.Node) {
	for _, child := range children {
		v.validateNode(child)
	}
}

// inRange checks if val is in [lo, hi].
func inRange(val, lo, hi float64) bool {
	return val >= lo && val <= hi
}

// inColorRange checks if all R,G,B components are in [0, 1].
func inColorRange(c vec.SFColor) bool {
	return inRange(c.R, 0, 1) && inRange(c.G, 0, 1) && inRange(c.B, 0, 1)
}

// --- Appearance ---

func (v *Validator) validateAppearance(n *node.Appearance) {
	if n.Material != nil {
		v.validateNode(n.Material)
	}
	if n.Texture != nil {
		v.validateNode(n.Texture)
	}
	if n.TextureTransform != nil {
		v.validateNode(n.TextureTransform)
	}
}

// --- Material ---

func (v *Validator) validateMaterial(n *node.Material) {
	nt, nn := "Material", n.GetName()
	if !inRange(n.AmbientIntensity, 0, 1) {
		v.addError(nt, nn, "ambientIntensity %g out of range [0, 1]", n.AmbientIntensity)
	}
	if !inRange(n.Shininess, 0, 1) {
		v.addError(nt, nn, "shininess %g out of range [0, 1]", n.Shininess)
	}
	if !inRange(n.Transparency, 0, 1) {
		v.addError(nt, nn, "transparency %g out of range [0, 1]", n.Transparency)
	}
	if !inColorRange(n.DiffuseColor) {
		v.addError(nt, nn, "diffuseColor components out of range [0, 1]")
	}
	if !inColorRange(n.EmissiveColor) {
		v.addError(nt, nn, "emissiveColor components out of range [0, 1]")
	}
	if !inColorRange(n.SpecularColor) {
		v.addError(nt, nn, "specularColor components out of range [0, 1]")
	}
}

// --- Shape ---

func (v *Validator) validateShape(n *node.Shape) {
	if n.Appearance != nil {
		v.validateNode(n.Appearance)
	}
	if n.Geometry != nil {
		v.validateNode(n.Geometry)
	}
}

// --- SpotLight ---

func (v *Validator) validateSpotLight(n *node.SpotLight) {
	nt, nn := "SpotLight", n.GetName()
	halfPi := float64(math.Pi / 2)
	if n.Radius < 0 {
		v.addError(nt, nn, "radius %g must be >= 0", n.Radius)
	}
	if n.BeamWidth < 0 || n.BeamWidth > halfPi {
		v.addError(nt, nn, "beamWidth %g out of range [0, pi/2]", n.BeamWidth)
	}
	if n.CutOffAngle < 0 || n.CutOffAngle > halfPi {
		v.addError(nt, nn, "cutOffAngle %g out of range [0, pi/2]", n.CutOffAngle)
	}
}

// --- Box ---

func (v *Validator) validateBox(n *node.Box) {
	nt, nn := "Box", n.GetName()
	if n.Size.X <= 0 || n.Size.Y <= 0 || n.Size.Z <= 0 {
		v.addError(nt, nn, "size (%g %g %g) must have all components > 0", n.Size.X, n.Size.Y, n.Size.Z)
	}
}

// --- Cone ---

func (v *Validator) validateCone(n *node.Cone) {
	nt, nn := "Cone", n.GetName()
	if n.BottomRadius <= 0 {
		v.addError(nt, nn, "bottomRadius %g must be > 0", n.BottomRadius)
	}
	if n.Height <= 0 {
		v.addError(nt, nn, "height %g must be > 0", n.Height)
	}
}

// --- Cylinder ---

func (v *Validator) validateCylinder(n *node.Cylinder) {
	nt, nn := "Cylinder", n.GetName()
	if n.Radius <= 0 {
		v.addError(nt, nn, "radius %g must be > 0", n.Radius)
	}
	if n.Height <= 0 {
		v.addError(nt, nn, "height %g must be > 0", n.Height)
	}
}

// --- Sphere ---

func (v *Validator) validateSphere(n *node.Sphere) {
	if n.Radius <= 0 {
		v.addError("Sphere", n.GetName(), "radius %g must be > 0", n.Radius)
	}
}

// --- Viewpoint ---

func (v *Validator) validateViewpoint(n *node.Viewpoint) {
	nt, nn := "Viewpoint", n.GetName()
	if n.FieldOfView <= 0 || n.FieldOfView >= float64(math.Pi) {
		v.addError(nt, nn, "fieldOfView %g out of range (0, pi)", n.FieldOfView)
	}
}

// --- Transform ---

func (v *Validator) validateTransform(n *node.Transform) {
	nt, nn := "Transform", n.GetName()
	if n.Scale.X == 0 || n.Scale.Y == 0 || n.Scale.Z == 0 {
		v.addWarning(nt, nn, "scale has zero component (%g %g %g) — geometry will collapse", n.Scale.X, n.Scale.Y, n.Scale.Z)
	}
}

// --- ElevationGrid ---

func (v *Validator) validateElevationGrid(n *node.ElevationGrid) {
	nt, nn := "ElevationGrid", n.GetName()
	if n.XDimension < 1 {
		v.addError(nt, nn, "xDimension %d must be >= 1", n.XDimension)
	}
	if n.ZDimension < 1 {
		v.addError(nt, nn, "zDimension %d must be >= 1", n.ZDimension)
	}
	if n.XSpacing < 0 {
		v.addError(nt, nn, "xSpacing %g must be >= 0", n.XSpacing)
	}
	if n.ZSpacing < 0 {
		v.addError(nt, nn, "zSpacing %g must be >= 0", n.ZSpacing)
	}
	nVerts := n.XDimension * n.ZDimension
	if int64(len(n.Heights)) < nVerts {
		v.addError(nt, nn, "height array has %d values, need at least %d (xDimension * zDimension)", len(n.Heights), nVerts)
	}
}

// --- IndexedFaceSet ---

func (v *Validator) validateIndexedFaceSet(n *node.IndexedFaceSet) {
	nt, nn := "IndexedFaceSet", n.GetName()
	v.validateDataSetCoords(nt, nn, n.Coord, n.CoordIndex)
	v.validateDataSetColors(nt, nn, n.Color, n.ColorIndex, n.CoordIndex, n.ColorPerVertex)
}

// --- IndexedLineSet ---

func (v *Validator) validateIndexedLineSet(n *node.IndexedLineSet) {
	nt, nn := "IndexedLineSet", n.GetName()
	v.validateDataSetCoords(nt, nn, n.Coord, n.CoordIndex)
	v.validateDataSetColors(nt, nn, n.Color, n.ColorIndex, n.CoordIndex, n.ColorPerVertex)
}

// --- PointSet ---

func (v *Validator) validatePointSet(n *node.PointSet) {
	nt, nn := "PointSet", n.GetName()
	if n.Coord != nil && n.Color != nil {
		nCoords := len(n.Coord.Point)
		nColors := len(n.Color.Color)
		if nCoords > 0 && nColors > 0 && nColors > nCoords {
			v.addWarning(nt, nn, "more colors (%d) than coordinates (%d)", nColors, nCoords)
		}
	}
}

// --- Shared dataset helpers ---

func (v *Validator) validateDataSetCoords(nt, nn string, coord *node.Coordinate, coordIndex []int64) {
	if coord == nil || len(coord.Point) == 0 {
		return
	}
	nCoords := int64(len(coord.Point))
	for i, idx := range coordIndex {
		if idx != -1 && (idx < 0 || idx >= nCoords) {
			v.addError(nt, nn, "coordIndex[%d] = %d out of range [0, %d)", i, idx, nCoords)
			break
		}
	}
}

func (v *Validator) validateDataSetColors(nt, nn string, color *node.ColorNode, colorIndex, coordIndex []int64, colorPerVertex bool) {
	if color == nil || len(color.Color) == 0 {
		return
	}
	nColors := int64(len(color.Color))

	if !colorPerVertex {
		// Color per face/polyline: count faces (separated by -1 sentinels)
		nFaces := countFaces(coordIndex)
		if len(colorIndex) > 0 {
			for i, idx := range colorIndex {
				if idx >= nColors {
					v.addError(nt, nn, "colorIndex[%d] = %d out of range [0, %d)", i, idx, nColors)
					break
				}
			}
		} else if nColors < int64(nFaces) {
			v.addError(nt, nn, "need at least %d colors for %d faces, have %d", nFaces, nFaces, nColors)
		}
	} else {
		// Color per vertex
		if len(colorIndex) > 0 {
			for i, idx := range colorIndex {
				if idx != -1 && idx >= nColors {
					v.addError(nt, nn, "colorIndex[%d] = %d out of range [0, %d)", i, idx, nColors)
					break
				}
			}
		} else {
			// When no colorIndex, largest coordIndex must be < nColors
			largest := int64(-1)
			for _, idx := range coordIndex {
				if idx > largest {
					largest = idx
				}
			}
			if largest >= nColors {
				v.addError(nt, nn, "largest coordIndex %d requires at least %d colors, have %d", largest, largest+1, nColors)
			}
		}
	}
}

func countFaces(coordIndex []int64) int {
	if len(coordIndex) == 0 {
		return 0
	}
	n := 0
	for _, idx := range coordIndex {
		if idx == -1 {
			n++
		}
	}
	// Last face may not end with -1
	if coordIndex[len(coordIndex)-1] != -1 {
		n++
	}
	return n
}

// --- Interpolators ---

func (v *Validator) validateColorInterpolator(n *node.ColorInterpolator) {
	nk := len(n.Key)
	nv := len(n.KeyValue)
	if nk > 0 && nv != nk {
		v.addError("ColorInterpolator", n.GetName(), "key count (%d) != keyValue count (%d)", nk, nv)
	}
}

func (v *Validator) validateCoordinateInterpolator(n *node.CoordinateInterpolator) {
	nk := len(n.Key)
	nv := len(n.KeyValue)
	if nk > 0 && nv%nk != 0 {
		v.addError("CoordinateInterpolator", n.GetName(), "keyValue count (%d) not a multiple of key count (%d)", nv, nk)
	}
}

func (v *Validator) validateNormalInterpolator(n *node.NormalInterpolator) {
	nk := len(n.Key)
	nv := len(n.KeyValue)
	if nk > 0 && nv%nk != 0 {
		v.addError("NormalInterpolator", n.GetName(), "keyValue count (%d) not a multiple of key count (%d)", nv, nk)
	}
}

func (v *Validator) validateOrientationInterpolator(n *node.OrientationInterpolator) {
	nk := len(n.Key)
	nv := len(n.KeyValue)
	if nk > 0 && nv != nk {
		v.addError("OrientationInterpolator", n.GetName(), "key count (%d) != keyValue count (%d)", nk, nv)
	}
}

func (v *Validator) validatePositionInterpolator(n *node.PositionInterpolator) {
	nk := len(n.Key)
	nv := len(n.KeyValue)
	if nk > 0 && nv != nk {
		v.addError("PositionInterpolator", n.GetName(), "key count (%d) != keyValue count (%d)", nk, nv)
	}
}

func (v *Validator) validateScalarInterpolator(n *node.ScalarInterpolator) {
	nk := len(n.Key)
	nv := len(n.KeyValue)
	if nk > 0 && nv != nk {
		v.addError("ScalarInterpolator", n.GetName(), "key count (%d) != keyValue count (%d)", nk, nv)
	}
}

// --- Inline ---

func (v *Validator) validateInline(n *node.Inline) {
	if len(n.URL) == 0 {
		v.addWarning("Inline", n.GetName(), "no URL specified")
	}
}

// --- LOD ---

func (v *Validator) validateLOD(n *node.LOD) {
	nt, nn := "LOD", n.GetName()
	if len(n.Level) > 0 && len(n.Range) > 0 {
		if len(n.Range) != len(n.Level)-1 {
			v.addWarning(nt, nn, "range count (%d) should be level count - 1 (%d)", len(n.Range), len(n.Level)-1)
		}
	}
	for _, child := range n.Level {
		v.validateNode(child)
	}
}

// --- Switch ---

func (v *Validator) validateSwitch(n *node.Switch) {
	nt, nn := "Switch", n.GetName()
	if n.WhichChoice >= int64(len(n.Choice)) && len(n.Choice) > 0 {
		v.addWarning(nt, nn, "whichChoice %d out of range [0, %d)", n.WhichChoice, len(n.Choice))
	}
	for _, child := range n.Choice {
		v.validateNode(child)
	}
}
