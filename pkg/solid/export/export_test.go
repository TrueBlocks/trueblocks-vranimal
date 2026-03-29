package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

var yellow = vec.SFColor{R: 1, G: 0.85, B: 0.1, A: 1}

func makeTriangle() *base.Solid {
	positions := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	indices := []int64{0, 1, 2, -1}
	s, _ := euler.BuildFromIndexSet(positions, indices, yellow)
	return s
}

func makeQuad() *base.Solid {
	positions := []vec.SFVec3f{
		{X: -1, Y: -1, Z: 0},
		{X: 1, Y: -1, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: -1, Y: 1, Z: 0},
	}
	indices := []int64{0, 1, 2, 3, -1}
	s, _ := euler.BuildFromIndexSet(positions, indices, yellow)
	return s
}

func TestVRML_Header(t *testing.T) {
	s := makeTriangle()
	if s == nil {
		t.Fatal("makeTriangle returned nil")
	}
	var buf bytes.Buffer
	if err := VRML(s, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "#VRML V2.0 utf8\n") {
		t.Errorf("missing VRML header, got: %q", out[:min(len(out), 40)])
	}
	if !strings.Contains(out, "IndexedFaceSet") {
		t.Error("output should contain IndexedFaceSet")
	}
	if !strings.Contains(out, "coordIndex") {
		t.Error("output should contain coordIndex")
	}
}

func TestShape_IndentedOutput(t *testing.T) {
	s := makeTriangle()
	if s == nil {
		t.Fatal("makeTriangle returned nil")
	}
	var buf bytes.Buffer
	if err := Shape(s, &buf, "  "); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "  Shape {") {
		t.Errorf("Shape output should be indented, got: %q", out[:min(len(out), 20)])
	}
}

func TestMultiVRML_TwoSolids(t *testing.T) {
	s1 := makeTriangle()
	s2 := makeQuad()
	if s1 == nil || s2 == nil {
		t.Fatal("failed to build test solids")
	}
	translations := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 3, Y: 0, Z: 0},
	}
	var buf bytes.Buffer
	if err := MultiVRML(&buf, []*base.Solid{s1, s2}, translations); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "#VRML V2.0 utf8\n") {
		t.Error("missing VRML header")
	}
	if strings.Count(out, "Transform {") != 2 {
		t.Errorf("expected 2 Transform blocks, got %d", strings.Count(out, "Transform {"))
	}
	if !strings.Contains(out, "translation 3 0 0") {
		t.Error("second solid should have translation 3 0 0")
	}
}

func TestWireframe_Output(t *testing.T) {
	s := makeQuad()
	if s == nil {
		t.Fatal("makeQuad returned nil")
	}
	var buf bytes.Buffer
	white := vec.SFColor{R: 1, G: 1, B: 1, A: 1}
	if err := Wireframe(s, &buf, "", white); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "IndexedLineSet") {
		t.Error("wireframe output should contain IndexedLineSet")
	}
	if !strings.Contains(out, "emissiveColor 1 1 1") {
		t.Error("wireframe should use the given emissive color")
	}
}

func TestVRMLFile_WritesFile(t *testing.T) {
	s := makeTriangle()
	if s == nil {
		t.Fatal("makeTriangle returned nil")
	}
	path := t.TempDir() + "/test.wrl"
	if err := VRMLFile(s, path); err != nil {
		t.Fatal(err)
	}
}

func TestMultiVRMLFile_WritesFile(t *testing.T) {
	s := makeTriangle()
	if s == nil {
		t.Fatal("makeTriangle returned nil")
	}
	path := t.TempDir() + "/multi.wrl"
	if err := MultiVRMLFile(path, []*base.Solid{s}, nil); err != nil {
		t.Fatal(err)
	}
}

func TestVRML_FaceColor(t *testing.T) {
	s := makeTriangle()
	if s == nil {
		t.Fatal("makeTriangle returned nil")
	}
	var buf bytes.Buffer
	if err := VRML(s, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Color { color [") {
		t.Error("output should contain per-face color data")
	}
}
