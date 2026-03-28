package traverser

import (
	"fmt"
	"io"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Base Traverser
// ---------------------------------------------------------------------------

// Base provides common traversal logic for the visitor pattern.
type Base struct {
	MatrixStack []vec.Matrix
}

// NewBase creates a base traverser with an identity matrix on the stack.
func NewBase() Base {
	return Base{
		MatrixStack: []vec.Matrix{vec.Identity()},
	}
}

// PushMatrix pushes a matrix onto the stack (multiplied with current top).
func (b *Base) PushMatrix(m vec.Matrix) {
	top := b.TopMatrix()
	b.MatrixStack = append(b.MatrixStack, top.Mul(m))
}

// PopMatrix pops the top matrix from the stack.
func (b *Base) PopMatrix() {
	if len(b.MatrixStack) > 1 {
		b.MatrixStack = b.MatrixStack[:len(b.MatrixStack)-1]
	}
}

// TopMatrix returns the current top of the matrix stack.
func (b *Base) TopMatrix() vec.Matrix {
	return b.MatrixStack[len(b.MatrixStack)-1]
}

// ---------------------------------------------------------------------------
// WriteTraverser - serializes the scene graph back to VRML text
// ---------------------------------------------------------------------------

// WriteTraverser writes a VRML scene graph to an io.Writer.
type WriteTraverser struct {
	Base
	w     io.Writer
	depth int
}

// NewWriteTraverser creates a write traverser targeting the given writer.
func NewWriteTraverser(w io.Writer) *WriteTraverser {
	return &WriteTraverser{
		Base: NewBase(),
		w:    w,
	}
}

// PreTraverse is called before traversal begins.
func (wt *WriteTraverser) PreTraverse() {
	_, _ = fmt.Fprintln(wt.w, "#VRML V2.0 utf8")
}

// PostTraverse is called after traversal ends.
func (wt *WriteTraverser) PostTraverse() {}

// TraverseNode dispatches to the appropriate write method.
func (wt *WriteTraverser) TraverseNode(n node.Node) {
	switch v := n.(type) {
	case *node.Transform:
		wt.writeTransform(v)
	case *node.Group:
		wt.writeGroup(v)
	case *node.Shape:
		wt.writeShape(v)
	case *node.Appearance:
		wt.writeAppearance(v)
	case *node.Material:
		wt.writeMaterial(v)
	case *node.Box:
		wt.writeBox(v)
	case *node.Sphere:
		wt.writeSphere(v)
	case *node.Cone:
		wt.writeCone(v)
	case *node.Cylinder:
		wt.writeCylinder(v)
	case *node.IndexedFaceSet:
		wt.writeIndexedFaceSet(v)
	case *node.Coordinate:
		wt.writeCoordinate(v)
	case *node.NormalNode:
		wt.writeNormal(v)
	case *node.ColorNode:
		wt.writeColorNode(v)
	case *node.DirectionalLight:
		wt.writeDirLight(v)
	case *node.PointLight:
		wt.writePointLight(v)
	case *node.Viewpoint:
		wt.writeViewpoint(v)
	default:
		// Skip unknown node types
	}
}

// TraverseChildren visits all children of a grouping node.
func (wt *WriteTraverser) TraverseChildren(g *node.GroupingNode) {
	for _, child := range g.Children {
		wt.TraverseNode(child)
	}
}

func (wt *WriteTraverser) indent() string {
	s := ""
	for i := 0; i < wt.depth; i++ {
		s += "  "
	}
	return s
}

func (wt *WriteTraverser) defPrefix(n node.Node) string {
	if n.GetName() != "" {
		return "DEF " + n.GetName() + " "
	}
	return ""
}

func (wt *WriteTraverser) wprintf(format string, args ...any) {
	_, _ = fmt.Fprintf(wt.w, format, args...)
}

func (wt *WriteTraverser) writeTransform(t *node.Transform) {
	ind := wt.indent()
	wt.wprintf("%s%sTransform {\n", ind, wt.defPrefix(t))
	wt.depth++
	if t.Translation != (vec.SFVec3f{}) {
		wt.wprintf("%stranslation %g %g %g\n", wt.indent(), t.Translation.X, t.Translation.Y, t.Translation.Z)
	}
	if t.Rotation != (vec.SFRotation{X: 0, Y: 1, Z: 0, W: 0}) {
		wt.wprintf("%srotation %g %g %g %g\n", wt.indent(), t.Rotation.X, t.Rotation.Y, t.Rotation.Z, t.Rotation.W)
	}
	if t.Scale != (vec.SFVec3f{X: 1, Y: 1, Z: 1}) {
		wt.wprintf("%sscale %g %g %g\n", wt.indent(), t.Scale.X, t.Scale.Y, t.Scale.Z)
	}
	if t.HasChildren() {
		wt.wprintf("%schildren [\n", wt.indent())
		wt.depth++
		wt.TraverseChildren(&t.GroupingNode)
		wt.depth--
		wt.wprintf("%s]\n", wt.indent())
	}
	wt.depth--
	wt.wprintf("%s}\n", ind)
}

func (wt *WriteTraverser) writeGroup(g *node.Group) {
	ind := wt.indent()
	wt.wprintf("%s%sGroup {\n", ind, wt.defPrefix(g))
	wt.depth++
	if g.HasChildren() {
		wt.wprintf("%schildren [\n", wt.indent())
		wt.depth++
		wt.TraverseChildren(&g.GroupingNode)
		wt.depth--
		wt.wprintf("%s]\n", wt.indent())
	}
	wt.depth--
	wt.wprintf("%s}\n", ind)
}

func (wt *WriteTraverser) writeShape(s *node.Shape) {
	ind := wt.indent()
	wt.wprintf("%s%sShape {\n", ind, wt.defPrefix(s))
	wt.depth++
	if s.Appearance != nil {
		wt.TraverseNode(s.Appearance)
	}
	if s.Geometry != nil {
		wt.wprintf("%sgeometry ", wt.indent())
		wt.TraverseNode(s.Geometry)
	}
	wt.depth--
	wt.wprintf("%s}\n", ind)
}

func (wt *WriteTraverser) writeAppearance(a *node.Appearance) {
	ind := wt.indent()
	wt.wprintf("%sappearance Appearance {\n", ind)
	wt.depth++
	if a.Material != nil {
		wt.TraverseNode(a.Material)
	}
	wt.depth--
	wt.wprintf("%s}\n", ind)
}

func (wt *WriteTraverser) writeMaterial(m *node.Material) {
	ind := wt.indent()
	wt.wprintf("%smaterial Material {\n", ind)
	wt.depth++
	wt.wprintf("%sdiffuseColor %g %g %g\n", wt.indent(), m.DiffuseColor.R, m.DiffuseColor.G, m.DiffuseColor.B)
	if m.AmbientIntensity != 0.2 {
		wt.wprintf("%sambientIntensity %g\n", wt.indent(), m.AmbientIntensity)
	}
	if m.Shininess != 0.2 {
		wt.wprintf("%sshininess %g\n", wt.indent(), m.Shininess)
	}
	if m.Transparency != 0 {
		wt.wprintf("%stransparency %g\n", wt.indent(), m.Transparency)
	}
	wt.depth--
	wt.wprintf("%s}\n", ind)
}

func (wt *WriteTraverser) writeBox(b *node.Box) {
	wt.wprintf("Box { size %g %g %g }\n", b.Size.X, b.Size.Y, b.Size.Z)
}

func (wt *WriteTraverser) writeSphere(s *node.Sphere) {
	wt.wprintf("Sphere { radius %g }\n", s.Radius)
}

func (wt *WriteTraverser) writeCone(c *node.Cone) {
	wt.wprintf("Cone { bottomRadius %g height %g }\n", c.BottomRadius, c.Height)
}

func (wt *WriteTraverser) writeCylinder(c *node.Cylinder) {
	wt.wprintf("Cylinder { radius %g height %g }\n", c.Radius, c.Height)
}

func (wt *WriteTraverser) writeIndexedFaceSet(ifs *node.IndexedFaceSet) {
	ind := wt.indent()
	wt.wprintf("IndexedFaceSet {\n")
	wt.depth++
	if ifs.Coord != nil {
		wt.wprintf("%scoord ", wt.indent())
		wt.TraverseNode(ifs.Coord)
	}
	if len(ifs.CoordIndex) > 0 {
		wt.wprintf("%scoordIndex [ ", wt.indent())
		for _, idx := range ifs.CoordIndex {
			wt.wprintf("%d, ", idx)
		}
		wt.wprintf("]\n")
	}
	if ifs.Normal != nil {
		wt.wprintf("%snormal ", wt.indent())
		wt.TraverseNode(ifs.Normal)
	}
	if ifs.Color != nil {
		wt.wprintf("%scolor ", wt.indent())
		wt.TraverseNode(ifs.Color)
	}
	wt.depth--
	wt.wprintf("%s}\n", ind)
}

func (wt *WriteTraverser) writeCoordinate(c *node.Coordinate) {
	wt.wprintf("Coordinate { point [ ")
	for _, p := range c.Point {
		wt.wprintf("%g %g %g, ", p.X, p.Y, p.Z)
	}
	wt.wprintf("]\n}")
	wt.wprintf("\n")
}

func (wt *WriteTraverser) writeNormal(n *node.NormalNode) {
	wt.wprintf("Normal { vector [ ")
	for _, v := range n.Vector {
		wt.wprintf("%g %g %g, ", v.X, v.Y, v.Z)
	}
	wt.wprintf("]\n}")
	wt.wprintf("\n")
}

func (wt *WriteTraverser) writeColorNode(c *node.ColorNode) {
	wt.wprintf("Color { color [ ")
	for _, col := range c.Color {
		wt.wprintf("%g %g %g, ", col.R, col.G, col.B)
	}
	wt.wprintf("]\n}")
	wt.wprintf("\n")
}

func (wt *WriteTraverser) writeDirLight(dl *node.DirectionalLight) {
	ind := wt.indent()
	wt.wprintf("%sDirectionalLight {\n", ind)
	wt.depth++
	wt.wprintf("%sdirection %g %g %g\n", wt.indent(), dl.Direction.X, dl.Direction.Y, dl.Direction.Z)
	wt.wprintf("%sintensity %g\n", wt.indent(), dl.Intensity)
	wt.wprintf("%scolor %g %g %g\n", wt.indent(), dl.Color.R, dl.Color.G, dl.Color.B)
	wt.depth--
	wt.wprintf("%s}\n", ind)
}

func (wt *WriteTraverser) writePointLight(pl *node.PointLight) {
	ind := wt.indent()
	wt.wprintf("%sPointLight {\n", ind)
	wt.depth++
	wt.wprintf("%slocation %g %g %g\n", wt.indent(), pl.Location.X, pl.Location.Y, pl.Location.Z)
	wt.wprintf("%sintensity %g\n", wt.indent(), pl.Intensity)
	wt.wprintf("%scolor %g %g %g\n", wt.indent(), pl.Color.R, pl.Color.G, pl.Color.B)
	wt.depth--
	wt.wprintf("%s}\n", ind)
}

func (wt *WriteTraverser) writeViewpoint(vp *node.Viewpoint) {
	ind := wt.indent()
	wt.wprintf("%s%sViewpoint {\n", ind, wt.defPrefix(vp))
	wt.depth++
	if vp.Description != "" {
		wt.wprintf("%sdescription \"%s\"\n", wt.indent(), vp.Description)
	}
	wt.wprintf("%sposition %g %g %g\n", wt.indent(), vp.Position.X, vp.Position.Y, vp.Position.Z)
	wt.wprintf("%sorientation %g %g %g %g\n", wt.indent(), vp.Orientation.X, vp.Orientation.Y, vp.Orientation.Z, vp.Orientation.W)
	wt.depth--
	wt.wprintf("%s}\n", ind)
}

// ---------------------------------------------------------------------------
// ValidateTraverser - checks scene graph consistency
// ---------------------------------------------------------------------------

// ValidateTraverser walks the scene graph and collects warnings.
type ValidateTraverser struct {
	Base
	Warnings []string
}

// NewValidateTraverser creates a new validation traverser.
func NewValidateTraverser() *ValidateTraverser {
	return &ValidateTraverser{Base: NewBase()}
}

// PreTraverse resets warnings before a new traversal.
func (vt *ValidateTraverser) PreTraverse() {
	vt.Warnings = nil
}

// PostTraverse is a no-op.
func (vt *ValidateTraverser) PostTraverse() {}

// TraverseNode checks a node for common issues.
func (vt *ValidateTraverser) TraverseNode(n node.Node) {
	switch v := n.(type) {
	case *node.Shape:
		if v.Geometry == nil {
			vt.Warnings = append(vt.Warnings, "Shape has no geometry")
		}
	case *node.Transform:
		vt.TraverseChildren(&v.GroupingNode)
	case *node.Group:
		vt.TraverseChildren(&v.GroupingNode)
	}
}

// TraverseChildren visits all children.
func (vt *ValidateTraverser) TraverseChildren(g *node.GroupingNode) {
	for _, child := range g.Children {
		vt.TraverseNode(child)
	}
}
