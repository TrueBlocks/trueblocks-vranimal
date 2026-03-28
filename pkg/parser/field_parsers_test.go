package parser

import (
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/writer"
)

// helper: parse VRML, return first node
func parseFirst(t *testing.T, vrml string) node.Node {
	t.Helper()
	p := NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if len(nodes) < 1 {
		t.Fatalf("expected at least 1 node, got %d", len(nodes))
	}
	return nodes[0]
}

func approx(a, b, eps float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < eps
}

// ===========================================================================
// Background
// ===========================================================================

func TestParse_Background(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
Background {
  skyColor [0 0 1, 0.5 0.5 1]
  skyAngle [1.57]
  groundColor [0.3 0.3 0.3]
  groundAngle [1.57]
  backUrl "back.jpg"
  bottomUrl "bottom.jpg"
  frontUrl "front.jpg"
  leftUrl "left.jpg"
  rightUrl "right.jpg"
  topUrl "top.jpg"
}`)
	bg, ok := n.(*node.Background)
	if !ok {
		t.Fatalf("expected *Background, got %T", n)
	}
	if len(bg.SkyColor) != 2 {
		t.Errorf("skyColor: got %d, want 2", len(bg.SkyColor))
	}
	if len(bg.SkyAngle) != 1 || !approx(bg.SkyAngle[0], 1.57, 0.01) {
		t.Errorf("skyAngle: got %v, want [1.57]", bg.SkyAngle)
	}
	if len(bg.GroundColor) != 1 {
		t.Errorf("groundColor: got %d, want 1", len(bg.GroundColor))
	}
	if len(bg.GroundAngle) != 1 {
		t.Errorf("groundAngle: got %d, want 1", len(bg.GroundAngle))
	}
	if len(bg.BackURL) != 1 || bg.BackURL[0] != "back.jpg" {
		t.Errorf("backUrl: got %v", bg.BackURL)
	}
	if len(bg.FrontURL) != 1 || bg.FrontURL[0] != "front.jpg" {
		t.Errorf("frontUrl: got %v", bg.FrontURL)
	}
	if len(bg.TopURL) != 1 || bg.TopURL[0] != "top.jpg" {
		t.Errorf("topUrl: got %v", bg.TopURL)
	}
}

// ===========================================================================
// WorldInfo
// ===========================================================================

func TestParse_WorldInfo(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
WorldInfo { title "Test Scene" info ["Author: jrush" "Version: 1"] }`)
	wi, ok := n.(*node.WorldInfo)
	if !ok {
		t.Fatalf("expected *WorldInfo, got %T", n)
	}
	if wi.Title != "Test Scene" {
		t.Errorf("title: got %q, want %q", wi.Title, "Test Scene")
	}
	if len(wi.Info) != 2 || wi.Info[0] != "Author: jrush" {
		t.Errorf("info: got %v", wi.Info)
	}
}

// ===========================================================================
// Script
// ===========================================================================

func TestParse_Script(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
Script { url "script.js" directOutput TRUE mustEvaluate TRUE }`)
	sc, ok := n.(*node.Script)
	if !ok {
		t.Fatalf("expected *Script, got %T", n)
	}
	if len(sc.URL) != 1 || sc.URL[0] != "script.js" {
		t.Errorf("url: got %v", sc.URL)
	}
	if !sc.DirectOutput {
		t.Error("directOutput should be true")
	}
	if !sc.MustEvaluate {
		t.Error("mustEvaluate should be true")
	}
}

// ===========================================================================
// Sound + AudioClip
// ===========================================================================

func TestParse_Sound_AudioClip(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
Sound {
  direction 1 0 0
  intensity 0.8
  location 0 2 0
  maxBack 20
  maxFront 30
  minBack 5
  minFront 10
  priority 0.5
  spatialize FALSE
  source AudioClip {
    url "beep.wav"
    description "Beep"
    loop TRUE
    pitch 1.5
    startTime 2
    stopTime 10
  }
}`)
	snd, ok := n.(*node.Sound)
	if !ok {
		t.Fatalf("expected *Sound, got %T", n)
	}
	if snd.Direction != (vec.SFVec3f{X: 1}) {
		t.Errorf("direction: got %v", snd.Direction)
	}
	if !approx(snd.Intensity, 0.8, 0.001) {
		t.Errorf("intensity: got %f", snd.Intensity)
	}
	if snd.Location != (vec.SFVec3f{Y: 2}) {
		t.Errorf("location: got %v", snd.Location)
	}
	if !approx(snd.MaxBack, 20, 0.001) {
		t.Errorf("maxBack: got %f", snd.MaxBack)
	}
	if !approx(snd.MaxFront, 30, 0.001) {
		t.Errorf("maxFront: got %f", snd.MaxFront)
	}
	if !approx(snd.MinBack, 5, 0.001) {
		t.Errorf("minBack: got %f", snd.MinBack)
	}
	if !approx(snd.MinFront, 10, 0.001) {
		t.Errorf("minFront: got %f", snd.MinFront)
	}
	if !approx(snd.Priority, 0.5, 0.001) {
		t.Errorf("priority: got %f", snd.Priority)
	}
	if snd.Spatialize {
		t.Error("spatialize should be false")
	}
	if snd.Source == nil {
		t.Fatal("source is nil")
	}
	ac := snd.Source
	if len(ac.URL) != 1 || ac.URL[0] != "beep.wav" {
		t.Errorf("audioClip url: got %v", ac.URL)
	}
	if ac.Description != "Beep" {
		t.Errorf("description: got %q", ac.Description)
	}
	if !ac.Loop {
		t.Error("loop should be true")
	}
	if !approx(ac.Pitch, 1.5, 0.001) {
		t.Errorf("pitch: got %f", ac.Pitch)
	}
	if !approx(ac.StartTime, 2, 0.001) {
		t.Errorf("startTime: got %f", ac.StartTime)
	}
	if !approx(ac.StopTime, 10, 0.001) {
		t.Errorf("stopTime: got %f", ac.StopTime)
	}
}

// ===========================================================================
// Text + FontStyle
// ===========================================================================

func TestParse_Text(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
Shape {
  geometry Text {
    string ["Hello" "World"]
    maxExtent 5
    length [2 3]
    fontStyle FontStyle {
      family "SERIF"
      size 2
      style "BOLD"
      horizontal FALSE
      leftToRight FALSE
      topToBottom FALSE
      spacing 1.5
      justify ["MIDDLE" "BEGIN"]
    }
  }
}`)
	sh, ok := n.(*node.Shape)
	if !ok {
		t.Fatalf("expected *Shape, got %T", n)
	}
	txt, ok := sh.Geometry.(*node.Text)
	if !ok {
		t.Fatalf("expected *Text, got %T", sh.Geometry)
	}
	if len(txt.String) != 2 || txt.String[0] != "Hello" || txt.String[1] != "World" {
		t.Errorf("string: got %v", txt.String)
	}
	if !approx(txt.MaxExtent, 5, 0.001) {
		t.Errorf("maxExtent: got %f", txt.MaxExtent)
	}
	if len(txt.Length) != 2 {
		t.Errorf("length: got %v", txt.Length)
	}
	if txt.FontStyle == nil {
		t.Fatal("fontStyle is nil")
	}
	fs := txt.FontStyle
	if fs.Family != "SERIF" {
		t.Errorf("family: got %q", fs.Family)
	}
	if !approx(fs.Size, 2, 0.001) {
		t.Errorf("size: got %f", fs.Size)
	}
	if fs.Style != "BOLD" {
		t.Errorf("style: got %q", fs.Style)
	}
	if fs.Horizontal {
		t.Error("horizontal should be false")
	}
	if fs.LeftToRight {
		t.Error("leftToRight should be false")
	}
	if fs.TopToBottom {
		t.Error("topToBottom should be false")
	}
	if !approx(fs.Spacing, 1.5, 0.001) {
		t.Errorf("spacing: got %f", fs.Spacing)
	}
	if len(fs.Justify) != 2 || fs.Justify[0] != "MIDDLE" {
		t.Errorf("justify: got %v", fs.Justify)
	}
}

// ===========================================================================
// Extrusion
// ===========================================================================

func TestParse_Extrusion(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
Shape {
  geometry Extrusion {
    crossSection [1 1, 1 -1, -1 -1, -1 1, 1 1]
    spine [0 0 0, 0 1 0, 0 2 0]
    scale [1 1, 0.5 0.5]
    orientation [0 0 1 0, 0 0 1 1.57]
    beginCap FALSE
    endCap FALSE
    creaseAngle 0.5
  }
}`)
	sh := n.(*node.Shape)
	ext, ok := sh.Geometry.(*node.Extrusion)
	if !ok {
		t.Fatalf("expected *Extrusion, got %T", sh.Geometry)
	}
	if len(ext.CrossSection) != 5 {
		t.Errorf("crossSection: got %d, want 5", len(ext.CrossSection))
	}
	if len(ext.Spine) != 3 {
		t.Errorf("spine: got %d, want 3", len(ext.Spine))
	}
	if len(ext.Scale) != 2 {
		t.Errorf("scale: got %d, want 2", len(ext.Scale))
	}
	if len(ext.Orientation) != 2 {
		t.Errorf("orientation: got %d, want 2", len(ext.Orientation))
	}
	if ext.BeginCap {
		t.Error("beginCap should be false")
	}
	if ext.EndCap {
		t.Error("endCap should be false")
	}
	if !approx(ext.CreaseAngle, 0.5, 0.001) {
		t.Errorf("creaseAngle: got %f", ext.CreaseAngle)
	}
}

// ===========================================================================
// CylinderSensor
// ===========================================================================

func TestParse_CylinderSensor(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
Transform { children [
  CylinderSensor {
    autoOffset FALSE
    diskAngle 0.5
    maxAngle 3.14
    minAngle -3.14
    offset 1.0
    enabled FALSE
  }
  Shape { geometry Box {} }
] }`)
	tr := n.(*node.Transform)
	cs, ok := tr.Children[0].(*node.CylinderSensor)
	if !ok {
		t.Fatalf("expected *CylinderSensor, got %T", tr.Children[0])
	}
	if cs.AutoOffset {
		t.Error("autoOffset should be false")
	}
	if !approx(cs.DiskAngle, 0.5, 0.001) {
		t.Errorf("diskAngle: got %f", cs.DiskAngle)
	}
	if !approx(cs.MaxAngle, 3.14, 0.001) {
		t.Errorf("maxAngle: got %f", cs.MaxAngle)
	}
	if !approx(cs.MinAngle, -3.14, 0.001) {
		t.Errorf("minAngle: got %f", cs.MinAngle)
	}
	if !approx(cs.Offset, 1.0, 0.001) {
		t.Errorf("offset: got %f", cs.Offset)
	}
	if cs.Enabled {
		t.Error("enabled should be false")
	}
}

// ===========================================================================
// PlaneSensor
// ===========================================================================

func TestParse_PlaneSensor(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
Transform { children [
  PlaneSensor {
    autoOffset FALSE
    maxPosition 10 20
    minPosition -10 -20
    offset 1 2 3
    enabled FALSE
  }
  Shape { geometry Box {} }
] }`)
	tr := n.(*node.Transform)
	ps, ok := tr.Children[0].(*node.PlaneSensor)
	if !ok {
		t.Fatalf("expected *PlaneSensor, got %T", tr.Children[0])
	}
	if ps.AutoOffset {
		t.Error("autoOffset should be false")
	}
	if ps.MaxPosition != (vec.SFVec2f{X: 10, Y: 20}) {
		t.Errorf("maxPosition: got %v", ps.MaxPosition)
	}
	if ps.MinPosition != (vec.SFVec2f{X: -10, Y: -20}) {
		t.Errorf("minPosition: got %v", ps.MinPosition)
	}
	if ps.Offset != (vec.SFVec3f{X: 1, Y: 2, Z: 3}) {
		t.Errorf("offset: got %v", ps.Offset)
	}
	if ps.Enabled {
		t.Error("enabled should be false")
	}
}

// ===========================================================================
// SphereSensor
// ===========================================================================

func TestParse_SphereSensor(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
Transform { children [
  SphereSensor {
    autoOffset FALSE
    offset 0 1 0 1.57
    enabled FALSE
  }
  Shape { geometry Box {} }
] }`)
	tr := n.(*node.Transform)
	ss, ok := tr.Children[0].(*node.SphereSensor)
	if !ok {
		t.Fatalf("expected *SphereSensor, got %T", tr.Children[0])
	}
	if ss.AutoOffset {
		t.Error("autoOffset should be false")
	}
	if !approx(ss.Offset.Y, 1, 0.001) || !approx(ss.Offset.W, 1.57, 0.001) {
		t.Errorf("offset: got %v", ss.Offset)
	}
	if ss.Enabled {
		t.Error("enabled should be false")
	}
}

// ===========================================================================
// VisibilitySensor
// ===========================================================================

func TestParse_VisibilitySensor(t *testing.T) {
	n := parseFirst(t, `#VRML V2.0 utf8
VisibilitySensor { center 1 2 3 size 10 20 30 enabled FALSE }`)
	vs, ok := n.(*node.VisibilitySensor)
	if !ok {
		t.Fatalf("expected *VisibilitySensor, got %T", n)
	}
	if vs.Center != (vec.SFVec3f{X: 1, Y: 2, Z: 3}) {
		t.Errorf("center: got %v", vs.Center)
	}
	if vs.Size != (vec.SFVec3f{X: 10, Y: 20, Z: 30}) {
		t.Errorf("size: got %v", vs.Size)
	}
	if vs.Enabled {
		t.Error("enabled should be false")
	}
}

// ===========================================================================
// Lexer coverage
// ===========================================================================

func TestLexer_Tokens(t *testing.T) {
	input := `{ } [ ] . ,
DEF USE ROUTE TO TRUE FALSE NULL
PROTO IS EXTERNPROTO
field exposedField eventIn eventOut
42 -7 3.14
"hello world"
MyIdentifier
`
	lex := NewLexer(strings.NewReader(input))

	expected := []Token{
		TokOpenBrace, TokCloseBrace, TokOpenBracket, TokCloseBracket, TokPeriod, TokComma,
		TokDEF, TokUSE, TokROUTE, TokTO, TokTRUE, TokFALSE, TokNULL,
		TokPROTO, TokIS, TokEXTERNPROTO,
		TokField, TokExposedField, TokEventIn, TokEventOut,
		TokInt, TokInt, TokFloat,
		TokString,
		TokIdentifier,
		TokEOF,
	}

	for i, want := range expected {
		got := lex.Next()
		if got != want {
			t.Errorf("token[%d]: got %v, want %v (strVal=%q)", i, got, want, lex.StrVal())
		}
	}
}

func TestLexer_IntValues(t *testing.T) {
	lex := NewLexer(strings.NewReader("42 -7 0xFF"))
	lex.Next()
	if lex.IntVal() != 42 {
		t.Errorf("got %d, want 42", lex.IntVal())
	}
	lex.Next()
	if lex.IntVal() != -7 {
		t.Errorf("got %d, want -7", lex.IntVal())
	}
	lex.Next()
	if lex.IntVal() != 255 {
		t.Errorf("got %d, want 255 (0xFF)", lex.IntVal())
	}
}

func TestLexer_FloatValues(t *testing.T) {
	lex := NewLexer(strings.NewReader("3.14 .5 1e4 1.5e+2"))
	for _, want := range []float64{3.14, 0.5, 10000, 150} {
		lex.Next()
		if !approx(lex.FloatVal(), want, 0.001) {
			t.Errorf("got %f, want %f", lex.FloatVal(), want)
		}
	}
}

func TestLexer_StringEscapes(t *testing.T) {
	lex := NewLexer(strings.NewReader(`"hello\nworld" "tab\there" "quote\"inside"`))
	lex.Next()
	if lex.StrVal() != "hello\nworld" {
		t.Errorf("got %q", lex.StrVal())
	}
	lex.Next()
	if lex.StrVal() != "tab\there" {
		t.Errorf("got %q", lex.StrVal())
	}
	lex.Next()
	if lex.StrVal() != `quote"inside` {
		t.Errorf("got %q", lex.StrVal())
	}
}

func TestLexer_PeekDoesNotConsume(t *testing.T) {
	lex := NewLexer(strings.NewReader("42 hello"))
	tok := lex.Peek()
	if tok != TokInt {
		t.Errorf("peek: got %v", tok)
	}
	// Peek again should return same
	tok2 := lex.Peek()
	if tok2 != TokInt {
		t.Errorf("second peek: got %v", tok2)
	}
	// Next should consume it
	lex.Next()
	if lex.IntVal() != 42 {
		t.Errorf("got %d after peek", lex.IntVal())
	}
	// Next identifier
	lex.Next()
	if lex.StrVal() != "hello" {
		t.Errorf("got %q", lex.StrVal())
	}
}

func TestLexer_Comments(t *testing.T) {
	lex := NewLexer(strings.NewReader("# comment\n42 # inline\n99"))
	lex.Next()
	if lex.IntVal() != 42 {
		t.Errorf("got %d, want 42", lex.IntVal())
	}
	lex.Next()
	if lex.IntVal() != 99 {
		t.Errorf("got %d, want 99", lex.IntVal())
	}
}

func TestLexer_LineTracking(t *testing.T) {
	lex := NewLexer(strings.NewReader("#VRML V2.0 utf8\n\n42"))
	lex.Next() // header (line 1)
	lex.Next() // 42 (line 3)
	if lex.Line() < 3 {
		t.Errorf("line: got %d, want >= 3", lex.Line())
	}
}

// ===========================================================================
// Round-trip tests for newly-parseable nodes (parser + writer)
// These verify the parser fixes by doing parse → write → parse → compare
// ===========================================================================

func TestRoundTrip_Background(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
Background { skyColor [0 0 1] groundColor [0.3 0.3 0.3] skyAngle [1.57] groundAngle [1.57] }`)
}

func TestRoundTrip_WorldInfo(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
WorldInfo { title "Test" info ["A" "B"] }`)
}

func TestRoundTrip_Script(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
Script { url "s.js" directOutput TRUE mustEvaluate TRUE }`)
}

func TestRoundTrip_Sound_AudioClip(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
Sound { direction 1 0 0 intensity 0.5 source AudioClip { url "a.wav" loop TRUE pitch 2 } }`)
}

func TestRoundTrip_Text_FontStyle(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
Shape { geometry Text { string ["Hi"] fontStyle FontStyle { family "SERIF" size 3 } } }`)
}

func TestRoundTrip_Extrusion(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
Shape { geometry Extrusion { crossSection [1 1, -1 1, -1 -1, 1 -1, 1 1] spine [0 0 0, 0 1 0] beginCap FALSE } }`)
}

func TestRoundTrip_CylinderSensor(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
Transform { children [ CylinderSensor { maxAngle 3 minAngle -3 } Shape { geometry Box {} } ] }`)
}

func TestRoundTrip_PlaneSensor(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
Transform { children [ PlaneSensor { maxPosition 10 10 } Shape { geometry Box {} } ] }`)
}

func TestRoundTrip_SphereSensor(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
Transform { children [ SphereSensor { autoOffset FALSE } Shape { geometry Box {} } ] }`)
}

func TestRoundTrip_VisibilitySensor(t *testing.T) {
	roundTrip(t, `#VRML V2.0 utf8
VisibilitySensor { center 1 2 3 size 10 20 30 }`)
}

// roundTrip helper: parse → write → parse → write, compare outputs
func roundTrip(t *testing.T, vrml string) {
	t.Helper()
	nodes1 := NewParser(strings.NewReader(vrml)).Parse()
	out1 := writeNodes(nodes1)
	nodes2 := NewParser(strings.NewReader(out1)).Parse()
	out2 := writeNodes(nodes2)
	if out1 != out2 {
		t.Fatalf("round-trip mismatch:\n--- first ---\n%s\n--- second ---\n%s", out1, out2)
	}
}

func writeNodes(nodes []node.Node) string {
	var buf strings.Builder
	w := writer.New(&buf)
	w.WriteScene(nodes)
	return buf.String()
}
