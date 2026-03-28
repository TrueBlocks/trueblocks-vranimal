package traverser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ===========================================================================
// Base traverser
// ===========================================================================

func TestNewBase(t *testing.T) {
	b := NewBase()
	if len(b.MatrixStack) != 1 {
		t.Fatalf("expected 1-element stack, got %d", len(b.MatrixStack))
	}
	if b.TopMatrix() != vec.Identity() {
		t.Fatal("initial matrix should be identity")
	}
}

func TestPushPopMatrix(t *testing.T) {
	b := NewBase()
	tr := vec.TranslationMatrix(1, 2, 3)
	b.PushMatrix(tr)
	if len(b.MatrixStack) != 2 {
		t.Fatalf("expected 2 after push, got %d", len(b.MatrixStack))
	}
	// Top should be identity * translate = translate
	top := b.TopMatrix()
	if top != tr {
		t.Fatal("top should equal the pushed translation matrix")
	}

	b.PopMatrix()
	if len(b.MatrixStack) != 1 {
		t.Fatalf("expected 1 after pop, got %d", len(b.MatrixStack))
	}
	if b.TopMatrix() != vec.Identity() {
		t.Fatal("should revert to identity after pop")
	}
}

func TestPopMatrixGuard(t *testing.T) {
	b := NewBase()
	b.PopMatrix() // should not panic or go below 1
	b.PopMatrix()
	if len(b.MatrixStack) != 1 {
		t.Fatal("should never go below 1")
	}
}

func TestPushMatrixAccumulation(t *testing.T) {
	b := NewBase()
	t1 := vec.TranslationMatrix(10, 0, 0)
	t2 := vec.TranslationMatrix(0, 20, 0)
	b.PushMatrix(t1)
	b.PushMatrix(t2)
	if len(b.MatrixStack) != 3 {
		t.Fatalf("expected 3, got %d", len(b.MatrixStack))
	}
	// Top should be identity * t1 * t2 = translate(10, 20, 0)
	top := b.TopMatrix()
	combined := t1.Mul(t2)
	if top != combined {
		t.Fatal("nested push should accumulate")
	}
	b.PopMatrix()
	if b.TopMatrix() != t1 {
		t.Fatal("after pop should be t1")
	}
}

// ===========================================================================
// WriteTraverser
// ===========================================================================

func TestWriteTraverser_PreTraverse(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	wt.PreTraverse()
	if !strings.Contains(buf.String(), "#VRML V2.0 utf8") {
		t.Fatal("PreTraverse should write VRML header")
	}
}

func TestWriteTraverser_Box(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	wt.TraverseNode(&node.Box{Size: vec.SFVec3f{X: 2, Y: 3, Z: 4}})
	out := buf.String()
	if !strings.Contains(out, "Box") || !strings.Contains(out, "2") {
		t.Fatalf("expected Box output, got: %s", out)
	}
}

func TestWriteTraverser_Sphere(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	wt.TraverseNode(&node.Sphere{Radius: 5})
	if !strings.Contains(buf.String(), "Sphere") || !strings.Contains(buf.String(), "5") {
		t.Fatal("expected Sphere output")
	}
}

func TestWriteTraverser_Cone(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	wt.TraverseNode(&node.Cone{BottomRadius: 1.5, Height: 3})
	out := buf.String()
	if !strings.Contains(out, "Cone") || !strings.Contains(out, "1.5") {
		t.Fatalf("expected Cone output, got: %s", out)
	}
}

func TestWriteTraverser_Cylinder(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	wt.TraverseNode(&node.Cylinder{Radius: 0.5, Height: 2})
	out := buf.String()
	if !strings.Contains(out, "Cylinder") || !strings.Contains(out, "0.5") {
		t.Fatalf("expected Cylinder output, got: %s", out)
	}
}

func TestWriteTraverser_Material(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	mat := &node.Material{
		DiffuseColor:     vec.SFColor{R: 1, G: 0, B: 0},
		AmbientIntensity: 0.5,
		Shininess:        0.8,
		Transparency:     0.3,
	}
	wt.TraverseNode(mat)
	out := buf.String()
	for _, word := range []string{"Material", "diffuseColor", "ambientIntensity", "shininess", "transparency"} {
		if !strings.Contains(out, word) {
			t.Fatalf("missing %q in: %s", word, out)
		}
	}
}

func TestWriteTraverser_MaterialDefaults(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	// Default material: ambientIntensity=0.2, shininess=0.2, transparency=0
	mat := &node.Material{
		DiffuseColor:     vec.SFColor{R: 0.8, G: 0.8, B: 0.8},
		AmbientIntensity: 0.2,
		Shininess:        0.2,
		Transparency:     0,
	}
	wt.TraverseNode(mat)
	out := buf.String()
	// Defaults should be omitted
	if strings.Contains(out, "ambientIntensity") {
		t.Fatal("default ambientIntensity should be omitted")
	}
	if strings.Contains(out, "transparency") {
		t.Fatal("default transparency should be omitted")
	}
}

func TestWriteTraverser_DirLight(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	dl := node.NewDirectionalLight()
	dl.Direction = vec.SFVec3f{X: 0, Y: -1, Z: 0}
	dl.Intensity = 0.8
	dl.Color = vec.SFColor{R: 1, G: 1, B: 1}
	wt.TraverseNode(dl)
	out := buf.String()
	if !strings.Contains(out, "DirectionalLight") || !strings.Contains(out, "direction") {
		t.Fatalf("expected DirectionalLight output, got: %s", out)
	}
}

func TestWriteTraverser_PointLight(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	pl := node.NewPointLight()
	pl.Location = vec.SFVec3f{X: 5, Y: 10, Z: 5}
	pl.Intensity = 1
	pl.Color = vec.SFColor{R: 1, G: 0.9, B: 0.8}
	wt.TraverseNode(pl)
	out := buf.String()
	if !strings.Contains(out, "PointLight") || !strings.Contains(out, "location") {
		t.Fatalf("expected PointLight output, got: %s", out)
	}
}

func TestWriteTraverser_Viewpoint(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	vp := &node.Viewpoint{
		Position:    vec.SFVec3f{X: 0, Y: 1.6, Z: 10},
		Orientation: vec.SFRotation{X: 0, Y: 1, Z: 0, W: 0},
		Description: "Front",
	}
	wt.TraverseNode(vp)
	out := buf.String()
	if !strings.Contains(out, "Viewpoint") || !strings.Contains(out, "Front") {
		t.Fatalf("expected Viewpoint with description, got: %s", out)
	}
}

func TestWriteTraverser_ViewpointNoDEF(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	vp := &node.Viewpoint{Position: vec.SFVec3f{X: 0, Y: 0, Z: 5}}
	wt.TraverseNode(vp)
	if strings.Contains(buf.String(), "DEF") {
		t.Fatal("unnamed viewpoint should not have DEF prefix")
	}
}

func TestWriteTraverser_ViewpointDEF(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	vp := &node.Viewpoint{Position: vec.SFVec3f{X: 0, Y: 0, Z: 5}}
	vp.SetName("MainView")
	wt.TraverseNode(vp)
	out := buf.String()
	if !strings.Contains(out, "DEF MainView Viewpoint") {
		t.Fatalf("expected DEF prefix, got: %s", out)
	}
}

func TestWriteTraverser_TransformNonDefault(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	tr := &node.Transform{
		Translation: vec.SFVec3f{X: 1, Y: 2, Z: 3},
		Scale:       vec.SFVec3f{X: 2, Y: 2, Z: 2},
	}
	wt.TraverseNode(tr)
	out := buf.String()
	if !strings.Contains(out, "translation") {
		t.Fatal("should write non-default translation")
	}
	if !strings.Contains(out, "scale") {
		t.Fatal("should write non-default scale")
	}
}

func TestWriteTraverser_TransformDefault(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	tr := &node.Transform{
		Translation: vec.SFVec3f{},
		Scale:       vec.SFVec3f{X: 1, Y: 1, Z: 1},
		Rotation:    vec.SFRotation{X: 0, Y: 1, Z: 0, W: 0},
	}
	wt.TraverseNode(tr)
	out := buf.String()
	if strings.Contains(out, "translation") {
		t.Fatal("default translation should be omitted")
	}
	if strings.Contains(out, "scale") {
		t.Fatal("default scale should be omitted")
	}
	if strings.Contains(out, "rotation") {
		t.Fatal("default rotation should be omitted")
	}
}

func TestWriteTraverser_TransformWithChildren(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	tr := &node.Transform{Translation: vec.SFVec3f{X: 1}}
	tr.Children = []node.Node{&node.Shape{Geometry: &node.Box{Size: vec.SFVec3f{X: 1, Y: 1, Z: 1}}}}
	wt.TraverseNode(tr)
	out := buf.String()
	if !strings.Contains(out, "children") {
		t.Fatal("should write children block")
	}
	if !strings.Contains(out, "Box") {
		t.Fatal("should contain Box child")
	}
}

func TestWriteTraverser_Group(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	g := &node.Group{}
	g.Children = []node.Node{&node.Shape{Geometry: &node.Sphere{Radius: 1}}}
	wt.TraverseNode(g)
	out := buf.String()
	if !strings.Contains(out, "Group") || !strings.Contains(out, "Sphere") {
		t.Fatalf("expected Group with Sphere, got: %s", out)
	}
}

func TestWriteTraverser_GroupEmpty(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	g := &node.Group{}
	wt.TraverseNode(g)
	out := buf.String()
	if strings.Contains(out, "children") {
		t.Fatal("empty group should omit children")
	}
}

func TestWriteTraverser_Shape(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	s := &node.Shape{
		Appearance: &node.Appearance{
			Material: &node.Material{DiffuseColor: vec.SFColor{R: 1}},
		},
		Geometry: &node.Box{Size: vec.SFVec3f{X: 1, Y: 1, Z: 1}},
	}
	wt.TraverseNode(s)
	out := buf.String()
	for _, word := range []string{"Shape", "appearance", "Appearance", "material", "Material", "geometry", "Box"} {
		if !strings.Contains(out, word) {
			t.Fatalf("missing %q in: %s", word, out)
		}
	}
}

func TestWriteTraverser_IFS(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	ifs := node.NewIndexedFaceSet()
	ifs.Coord = &node.Coordinate{Point: []vec.SFVec3f{{X: 0}, {X: 1}, {Y: 1}}}
	ifs.CoordIndex = []int64{0, 1, 2, -1}
	ifs.Normal = &node.NormalNode{Vector: []vec.SFVec3f{{Z: 1}, {Z: 1}, {Z: 1}}}
	ifs.Color = &node.ColorNode{Color: []vec.SFColor{{R: 1}, {G: 1}, {B: 1}}}
	wt.TraverseNode(ifs)
	out := buf.String()
	for _, word := range []string{"IndexedFaceSet", "coord", "Coordinate", "coordIndex", "normal", "Normal", "color", "Color"} {
		if !strings.Contains(out, word) {
			t.Fatalf("missing %q in: %s", word, out)
		}
	}
}

func TestWriteTraverser_UnknownNode(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	// An unknown node type should be silently skipped
	wt.TraverseNode(&node.Anchor{})
	if buf.Len() != 0 {
		t.Fatal("unknown node should produce no output")
	}
}

func TestWriteTraverser_FullScene(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	wt.PreTraverse()

	// Build a small scene: transform with shape + light
	tr := &node.Transform{Translation: vec.SFVec3f{X: 5}}
	tr.Children = []node.Node{
		&node.Shape{
			Appearance: &node.Appearance{
				Material: &node.Material{DiffuseColor: vec.SFColor{R: 0, G: 1, B: 0}},
			},
			Geometry: &node.Sphere{Radius: 2},
		},
		func() node.Node {
			dl := node.NewDirectionalLight()
			dl.Direction = vec.SFVec3f{Y: -1}
			return dl
		}(),
	}
	wt.TraverseNode(tr)
	wt.PostTraverse()

	out := buf.String()
	if !strings.HasPrefix(out, "#VRML V2.0 utf8") {
		t.Fatal("should start with VRML header")
	}
	if !strings.Contains(out, "Transform") {
		t.Fatal("should contain Transform")
	}
	if !strings.Contains(out, "Sphere") {
		t.Fatal("should contain Sphere")
	}
	if !strings.Contains(out, "DirectionalLight") {
		t.Fatal("should contain DirectionalLight")
	}
}

func TestWriteTraverser_Indentation(t *testing.T) {
	var buf bytes.Buffer
	wt := NewWriteTraverser(&buf)
	tr := &node.Transform{Translation: vec.SFVec3f{X: 1}}
	tr.Children = []node.Node{&node.Shape{Geometry: &node.Box{Size: vec.SFVec3f{X: 1, Y: 1, Z: 1}}}}
	wt.TraverseNode(tr)
	lines := strings.Split(buf.String(), "\n")
	// The "children [" line should be indented
	foundIndented := false
	for _, line := range lines {
		if strings.Contains(line, "children") && strings.HasPrefix(line, "  ") {
			foundIndented = true
		}
	}
	if !foundIndented {
		t.Fatalf("expected indented children line, got:\n%s", buf.String())
	}
}

// ===========================================================================
// ValidateTraverser
// ===========================================================================

func TestValidateTraverser_ShapeNoGeometry(t *testing.T) {
	vt := NewValidateTraverser()
	vt.PreTraverse()
	vt.TraverseNode(&node.Shape{})
	if len(vt.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(vt.Warnings))
	}
	if !strings.Contains(vt.Warnings[0], "no geometry") {
		t.Fatalf("expected 'no geometry' warning, got: %s", vt.Warnings[0])
	}
}

func TestValidateTraverser_ShapeWithGeometry(t *testing.T) {
	vt := NewValidateTraverser()
	vt.PreTraverse()
	vt.TraverseNode(&node.Shape{Geometry: &node.Box{Size: vec.SFVec3f{X: 1, Y: 1, Z: 1}}})
	if len(vt.Warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d: %v", len(vt.Warnings), vt.Warnings)
	}
}

func TestValidateTraverser_NestedShapes(t *testing.T) {
	vt := NewValidateTraverser()
	vt.PreTraverse()
	g := &node.Group{}
	g.Children = []node.Node{
		&node.Shape{Geometry: &node.Box{Size: vec.SFVec3f{X: 1, Y: 1, Z: 1}}},
		&node.Shape{}, // no geometry
	}
	vt.TraverseNode(g)
	if len(vt.Warnings) != 1 {
		t.Fatalf("expected 1 warning from nested shape, got %d", len(vt.Warnings))
	}
}

func TestValidateTraverser_TransformChildren(t *testing.T) {
	vt := NewValidateTraverser()
	vt.PreTraverse()
	tr := &node.Transform{}
	tr.Children = []node.Node{&node.Shape{}, &node.Shape{}}
	vt.TraverseNode(tr)
	if len(vt.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(vt.Warnings))
	}
}

func TestValidateTraverser_PreTraverseResets(t *testing.T) {
	vt := NewValidateTraverser()
	vt.PreTraverse()
	vt.TraverseNode(&node.Shape{})
	if len(vt.Warnings) != 1 {
		t.Fatal("expected 1 warning")
	}
	vt.PreTraverse() // should reset
	if len(vt.Warnings) != 0 {
		t.Fatal("PreTraverse should reset warnings")
	}
}

func TestValidateTraverser_DeeplyNested(t *testing.T) {
	vt := NewValidateTraverser()
	vt.PreTraverse()
	inner := &node.Group{}
	inner.Children = []node.Node{&node.Shape{}} // no geom
	outer := &node.Transform{}
	outer.Children = []node.Node{inner}
	vt.TraverseNode(outer)
	if len(vt.Warnings) != 1 {
		t.Fatalf("expected 1 warning from deep nesting, got %d", len(vt.Warnings))
	}
}
