package writer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func writeScene(nodes []node.Node) string {
	var buf bytes.Buffer
	w := New(&buf)
	w.WriteScene(nodes)
	return buf.String()
}

func parseVRML(vrml string) []node.Node {
	p := parser.NewParser(strings.NewReader(vrml))
	return p.Parse()
}

func TestWriteScene_Header(t *testing.T) {
	out := writeScene(nil)
	if !strings.HasPrefix(out, "#VRML V2.0 utf8") {
		t.Fatalf("bad header")
	}
}

func TestRoundTrip_Box(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nShape { geometry Box { size 2 3 4 } }"
	nodes := parseVRML(vrml)
	out1 := writeScene(nodes)
	nodes2 := parseVRML(out1)
	out2 := writeScene(nodes2)
	if out1 != out2 {
		t.Fatalf("round trip mismatch")
	}
}

func TestRoundTrip_Transform(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nTransform { translation 1 2 3 children [ Shape { geometry Sphere { radius 5 } } ] }"
	nodes := parseVRML(vrml)
	out1 := writeScene(nodes)
	nodes2 := parseVRML(out1)
	out2 := writeScene(nodes2)
	if out1 != out2 {
		t.Fatalf("round trip mismatch")
	}
}

func TestRoundTrip_Material(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nShape { appearance Appearance { material Material { diffuseColor 1 0 0 } } geometry Sphere {} }"
	nodes := parseVRML(vrml)
	out1 := writeScene(nodes)
	nodes2 := parseVRML(out1)
	out2 := writeScene(nodes2)
	if out1 != out2 {
		t.Fatalf("round trip mismatch")
	}
}

func TestRoundTrip_DEF_USE(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nDEF MyMat Material { diffuseColor 1 0 0 }\nShape { appearance Appearance { material USE MyMat } geometry Box {} }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "DEF MyMat") {
		t.Fatalf("missing DEF")
	}
	if !strings.Contains(out, "USE MyMat") {
		t.Fatalf("missing USE")
	}
}

func TestWrite_NavigationInfo(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nNavigationInfo { speed 2 headlight FALSE }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "NavigationInfo") {
		t.Fatalf("missing NavigationInfo")
	}
	if !strings.Contains(out, "speed 2") {
		t.Fatalf("missing speed")
	}
}

func TestWrite_DirectionalLight(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nDirectionalLight { direction 0 -1 0 intensity 0.8 }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "DirectionalLight") {
		t.Fatalf("missing node")
	}
}

func TestWrite_TimeSensor(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nTimeSensor { cycleInterval 5 loop TRUE }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "cycleInterval 5") {
		t.Fatalf("missing interval")
	}
	if !strings.Contains(out, "loop TRUE") {
		t.Fatalf("missing loop")
	}
}

func TestWrite_EmptyScene(t *testing.T) {
	out := writeScene([]node.Node{})
	if !strings.HasPrefix(out, "#VRML V2.0 utf8") {
		t.Fatalf("bad output")
	}
}

func TestWrite_ProgrammaticMaterial(t *testing.T) {
	m := node.NewMaterial()
	m.DiffuseColor = vec.SFColor{R: 1, G: 0, B: 0, A: 1}
	out := writeScene([]node.Node{m})
	if !strings.Contains(out, "Material") {
		t.Fatalf("missing Material")
	}
	if !strings.Contains(out, "diffuseColor 1 0 0") {
		t.Fatalf("missing color")
	}
}

func TestWrite_Switch(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nSwitch { whichChoice 1 choice [ Shape { geometry Box {} } Shape { geometry Sphere {} } ] }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "Switch") {
		t.Fatalf("missing Switch")
	}
	if !strings.Contains(out, "whichChoice 1") {
		t.Fatalf("missing whichChoice")
	}
}

func TestWrite_LOD(t *testing.T) {
	vrml := "#VRML V2.0 utf8\nLOD { range [50 100] level [ Shape { geometry Box {} } Shape { geometry Sphere {} } ] }"
	nodes := parseVRML(vrml)
	out := writeScene(nodes)
	if !strings.Contains(out, "LOD") {
		t.Fatalf("missing LOD")
	}
}
