package validator

import (
	"strings"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/parser"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func TestValidScene(t *testing.T) {
	scene := []node.Node{
		&node.Shape{
			Appearance: &node.Appearance{
				Material: node.NewMaterial(),
			},
			Geometry: node.NewBox(),
		},
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 0 {
		for _, f := range findings {
			t.Errorf("unexpected finding: %s", f)
		}
	}
}

func TestBoxZeroSize(t *testing.T) {
	scene := []node.Node{
		&node.Box{},
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != Error {
		t.Errorf("expected Error severity, got %s", findings[0].Severity)
	}
}

func TestSphereZeroRadius(t *testing.T) {
	scene := []node.Node{
		&node.Sphere{},
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestConeZero(t *testing.T) {
	scene := []node.Node{
		&node.Cone{},
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings (bottomRadius + height), got %d", len(findings))
	}
}

func TestCylinderZero(t *testing.T) {
	scene := []node.Node{
		&node.Cylinder{},
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings (radius + height), got %d", len(findings))
	}
}

func TestMaterialOutOfRange(t *testing.T) {
	scene := []node.Node{
		&node.Material{
			AmbientIntensity: 2.0,
			Shininess:        -0.5,
			Transparency:     1.5,
			DiffuseColor:     vec.SFColor{R: 2.0, G: 0, B: 0},
		},
	}
	v := New()
	findings := v.Validate(scene)
	// ambientIntensity, shininess, transparency, diffuseColor = 4 errors
	if len(findings) != 4 {
		for _, f := range findings {
			t.Logf("  %s", f)
		}
		t.Fatalf("expected 4 findings, got %d", len(findings))
	}
	for _, f := range findings {
		if f.Severity != Error {
			t.Errorf("expected Error severity, got %s for %s", f.Severity, f.Message)
		}
	}
}

func TestSpotLightValidation(t *testing.T) {
	scene := []node.Node{
		&node.SpotLight{
			Light:       node.Light{On: true, Intensity: 1.0},
			Radius:      -1,
			BeamWidth:   5.0,
			CutOffAngle: -0.1,
		},
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 3 {
		for _, f := range findings {
			t.Logf("  %s", f)
		}
		t.Fatalf("expected 3 findings (radius, beamWidth, cutOffAngle), got %d", len(findings))
	}
}

func TestInterpolatorKeyMismatch(t *testing.T) {
	scene := []node.Node{
		&node.ColorInterpolator{
			Interpolator: node.Interpolator{Key: []float64{0, 0.5, 1.0}},
			KeyValue:     []vec.SFColor{{R: 1}, {G: 1}},
		},
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != Error {
		t.Errorf("expected Error severity")
	}
}

func TestTransformZeroScale(t *testing.T) {
	scene := []node.Node{
		&node.Transform{
			GroupingNode: *node.NewGroupingNode(),
			Scale:        vec.SFVec3f{X: 1, Y: 0, Z: 1},
		},
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(findings))
	}
	if findings[0].Severity != Warning {
		t.Errorf("expected Warning severity, got %s", findings[0].Severity)
	}
}

func TestGroupTraversesChildren(t *testing.T) {
	scene := []node.Node{
		&node.Group{
			GroupingNode: node.GroupingNode{
				Children: []node.Node{
					&node.Sphere{},
					&node.Box{},
				},
			},
		},
	}
	v := New()
	findings := v.Validate(scene)
	// Sphere radius=0, Box size=0,0,0
	if len(findings) != 2 {
		for _, f := range findings {
			t.Logf("  %s", f)
		}
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestViewpointFieldOfView(t *testing.T) {
	scene := []node.Node{
		&node.Viewpoint{
			FieldOfView: 0,
		},
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestValidDefaults(t *testing.T) {
	// All nodes created with New* constructors should be valid
	scene := []node.Node{
		node.NewBox(),
		node.NewSphere(),
		node.NewCone(),
		node.NewCylinder(),
		node.NewMaterial(),
		node.NewViewpoint(),
		node.NewSpotLight(),
		node.NewTransform(),
	}
	v := New()
	findings := v.Validate(scene)
	if len(findings) != 0 {
		for _, f := range findings {
			t.Errorf("unexpected finding: %s", f)
		}
	}
}

// ===========================================================================
// Gap-filling tests (issue #51)
// ===========================================================================

// ---------------------------------------------------------------------------
// ElevationGrid validation
// ---------------------------------------------------------------------------

func TestElevationGrid_Valid(t *testing.T) {
	eg := node.NewElevationGrid()
	eg.XDimension = 3
	eg.ZDimension = 3
	eg.XSpacing = 1.0
	eg.ZSpacing = 1.0
	eg.Heights = []float64{0, 1, 0, 1, 2, 1, 0, 1, 0}

	v := New()
	findings := v.Validate([]node.Node{eg})
	if len(findings) != 0 {
		for _, f := range findings {
			t.Errorf("unexpected: %s", f)
		}
	}
}

func TestElevationGrid_ZeroDimensions(t *testing.T) {
	eg := node.NewElevationGrid()
	eg.XDimension = 0
	eg.ZDimension = 0

	v := New()
	findings := v.Validate([]node.Node{eg})
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	for _, f := range findings {
		if f.Severity != Error {
			t.Errorf("expected error, got %v: %s", f.Severity, f.Message)
		}
	}
}

func TestElevationGrid_NegativeSpacing(t *testing.T) {
	eg := node.NewElevationGrid()
	eg.XDimension = 2
	eg.ZDimension = 2
	eg.XSpacing = -1.0
	eg.ZSpacing = -1.0
	eg.Heights = []float64{0, 0, 0, 0}

	v := New()
	findings := v.Validate([]node.Node{eg})
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings (negative spacing), got %d", len(findings))
	}
}

func TestElevationGrid_InsufficientHeights(t *testing.T) {
	eg := node.NewElevationGrid()
	eg.XDimension = 3
	eg.ZDimension = 3
	eg.Heights = []float64{0, 1, 2} // need 9

	v := New()
	findings := v.Validate([]node.Node{eg})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for insufficient heights")
	}
}

// ---------------------------------------------------------------------------
// IndexedFaceSet — coordIndex validation
// ---------------------------------------------------------------------------

func TestIFS_Valid(t *testing.T) {
	ifs := node.NewIndexedFaceSet()
	ifs.Coord = &node.Coordinate{
		Point: []vec.SFVec3f{{X: 0}, {X: 1}, {Y: 1}, {X: 1, Y: 1}},
	}
	ifs.CoordIndex = []int64{0, 1, 2, -1, 1, 3, 2, -1}

	v := New()
	findings := v.Validate([]node.Node{ifs})
	if len(findings) != 0 {
		for _, f := range findings {
			t.Errorf("unexpected: %s", f)
		}
	}
}

func TestIFS_CoordIndexOutOfRange(t *testing.T) {
	ifs := node.NewIndexedFaceSet()
	ifs.Coord = &node.Coordinate{
		Point: []vec.SFVec3f{{X: 0}, {X: 1}, {Y: 1}},
	}
	ifs.CoordIndex = []int64{0, 1, 5, -1} // 5 is out of range [0,3)

	v := New()
	findings := v.Validate([]node.Node{ifs})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for out-of-range coordIndex")
	}
}

func TestIFS_NegativeCoordIndex(t *testing.T) {
	ifs := node.NewIndexedFaceSet()
	ifs.Coord = &node.Coordinate{
		Point: []vec.SFVec3f{{X: 0}, {X: 1}, {Y: 1}},
	}
	ifs.CoordIndex = []int64{0, -2, 2, -1} // -2 is invalid (only -1 is sentinel)

	v := New()
	findings := v.Validate([]node.Node{ifs})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for negative non-sentinel coordIndex")
	}
}

// ---------------------------------------------------------------------------
// IndexedFaceSet — color validation (per-face mode)
// ---------------------------------------------------------------------------

func TestIFS_ColorPerFace_Valid(t *testing.T) {
	ifs := node.NewIndexedFaceSet()
	ifs.ColorPerVertex = false
	ifs.Coord = &node.Coordinate{
		Point: []vec.SFVec3f{{X: 0}, {X: 1}, {Y: 1}, {X: 1, Y: 1}},
	}
	ifs.CoordIndex = []int64{0, 1, 2, -1, 1, 3, 2, -1} // 2 faces
	ifs.Color = &node.ColorNode{
		Color: []vec.SFColor{{R: 1}, {G: 1}}, // 2 colors for 2 faces
	}

	v := New()
	findings := v.Validate([]node.Node{ifs})
	if len(findings) != 0 {
		for _, f := range findings {
			t.Errorf("unexpected: %s", f)
		}
	}
}

func TestIFS_ColorPerFace_Insufficient(t *testing.T) {
	ifs := node.NewIndexedFaceSet()
	ifs.ColorPerVertex = false
	ifs.Coord = &node.Coordinate{
		Point: []vec.SFVec3f{{X: 0}, {X: 1}, {Y: 1}, {X: 1, Y: 1}},
	}
	ifs.CoordIndex = []int64{0, 1, 2, -1, 1, 3, 2, -1} // 2 faces
	ifs.Color = &node.ColorNode{
		Color: []vec.SFColor{{R: 1}}, // only 1 color for 2 faces
	}

	v := New()
	findings := v.Validate([]node.Node{ifs})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for insufficient per-face colors")
	}
}

func TestIFS_ColorPerFace_IndexOutOfRange(t *testing.T) {
	ifs := node.NewIndexedFaceSet()
	ifs.ColorPerVertex = false
	ifs.Coord = &node.Coordinate{
		Point: []vec.SFVec3f{{X: 0}, {X: 1}, {Y: 1}},
	}
	ifs.CoordIndex = []int64{0, 1, 2, -1}
	ifs.Color = &node.ColorNode{
		Color: []vec.SFColor{{R: 1}, {G: 1}},
	}
	ifs.ColorIndex = []int64{5} // out of range [0,2)

	v := New()
	findings := v.Validate([]node.Node{ifs})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for colorIndex out of range")
	}
}

// ---------------------------------------------------------------------------
// IndexedFaceSet — color validation (per-vertex mode)
// ---------------------------------------------------------------------------

func TestIFS_ColorPerVertex_InsufficientColors(t *testing.T) {
	ifs := node.NewIndexedFaceSet()
	ifs.ColorPerVertex = true
	ifs.Coord = &node.Coordinate{
		Point: []vec.SFVec3f{{X: 0}, {X: 1}, {Y: 1}, {X: 1, Y: 1}},
	}
	ifs.CoordIndex = []int64{0, 1, 2, -1, 1, 3, 2, -1}
	ifs.Color = &node.ColorNode{
		Color: []vec.SFColor{{R: 1}}, // only 1 color, but coordIndex refs up to 3
	}

	v := New()
	findings := v.Validate([]node.Node{ifs})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for insufficient per-vertex colors")
	}
}

func TestIFS_ColorPerVertex_IndexOutOfRange(t *testing.T) {
	ifs := node.NewIndexedFaceSet()
	ifs.ColorPerVertex = true
	ifs.Coord = &node.Coordinate{
		Point: []vec.SFVec3f{{X: 0}, {X: 1}, {Y: 1}},
	}
	ifs.CoordIndex = []int64{0, 1, 2, -1}
	ifs.Color = &node.ColorNode{
		Color: []vec.SFColor{{R: 1}, {G: 1}},
	}
	ifs.ColorIndex = []int64{0, 5, 0, -1} // 5 out of range

	v := New()
	findings := v.Validate([]node.Node{ifs})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for vertex colorIndex out of range")
	}
}

// ---------------------------------------------------------------------------
// IndexedLineSet — coord validation
// ---------------------------------------------------------------------------

func TestILS_CoordIndexOutOfRange(t *testing.T) {
	ils := node.NewIndexedLineSet()
	ils.Coord = &node.Coordinate{
		Point: []vec.SFVec3f{{X: 0}, {X: 1}},
	}
	ils.CoordIndex = []int64{0, 1, 10, -1} // 10 out of range

	v := New()
	findings := v.Validate([]node.Node{ils})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for IndexedLineSet out-of-range coordIndex")
	}
}

// ---------------------------------------------------------------------------
// PointSet — color warning
// ---------------------------------------------------------------------------

func TestPointSet_MoreColorsThanCoords(t *testing.T) {
	ps := node.NewPointSet()
	ps.Coord = &node.Coordinate{Point: []vec.SFVec3f{{X: 1}}}
	ps.Color = &node.ColorNode{Color: []vec.SFColor{{R: 1}, {G: 1}, {B: 1}}}

	v := New()
	findings := v.Validate([]node.Node{ps})
	warnings := countWarnings(findings)
	if warnings < 1 {
		t.Fatal("expected warning for more colors than coords")
	}
}

// ---------------------------------------------------------------------------
// Inline — no URL warning
// ---------------------------------------------------------------------------

func TestInline_NoURL(t *testing.T) {
	in := &node.Inline{}

	v := New()
	findings := v.Validate([]node.Node{in})
	warnings := countWarnings(findings)
	if warnings < 1 {
		t.Fatal("expected warning for Inline with no URL")
	}
}

func TestInline_WithURL(t *testing.T) {
	in := &node.Inline{URL: []string{"test.wrl"}}

	v := New()
	findings := v.Validate([]node.Node{in})
	if len(findings) != 0 {
		t.Errorf("unexpected finding for Inline with URL: %v", findings)
	}
}

// ---------------------------------------------------------------------------
// LOD — range/level mismatch warning
// ---------------------------------------------------------------------------

func TestLOD_RangeLevelMismatch(t *testing.T) {
	lod := &node.LOD{
		Level: []node.Node{node.NewBox(), node.NewSphere(), node.NewCone()}, // 3 levels
		Range: []float64{10},                                                // need 2 ranges
	}

	v := New()
	findings := v.Validate([]node.Node{lod})
	warnings := countWarnings(findings)
	if warnings < 1 {
		t.Fatal("expected warning for LOD range/level mismatch")
	}
}

func TestLOD_TraversesChildren(t *testing.T) {
	// LOD should validate children — put an invalid sphere in there
	s := node.NewSphere()
	s.Radius = 0
	lod := &node.LOD{Level: []node.Node{s}, Range: []float64{}}

	v := New()
	findings := v.Validate([]node.Node{lod})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error from LOD child validation")
	}
}

// ---------------------------------------------------------------------------
// Switch — whichChoice out of range
// ---------------------------------------------------------------------------

func TestSwitch_WhichChoiceOutOfRange(t *testing.T) {
	sw := &node.Switch{
		Choice:      []node.Node{node.NewBox()},
		WhichChoice: 5, // only 1 choice
	}

	v := New()
	findings := v.Validate([]node.Node{sw})
	warnings := countWarnings(findings)
	if warnings < 1 {
		t.Fatal("expected warning for whichChoice out of range")
	}
}

func TestSwitch_TraversesChildren(t *testing.T) {
	s := node.NewSphere()
	s.Radius = -1
	sw := &node.Switch{Choice: []node.Node{s}, WhichChoice: 0}

	v := New()
	findings := v.Validate([]node.Node{sw})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error from Switch child validation")
	}
}

// ---------------------------------------------------------------------------
// Interpolator variants
// ---------------------------------------------------------------------------

func TestCoordinateInterpolator_Mismatch(t *testing.T) {
	ci := &node.CoordinateInterpolator{
		Interpolator: node.Interpolator{Key: []float64{0, 1}},
		KeyValue:     []vec.SFVec3f{{X: 0}, {X: 1}, {X: 2}}, // 3 not multiple of 2
	}

	v := New()
	findings := v.Validate([]node.Node{ci})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for CoordinateInterpolator mismatch")
	}
}

func TestOrientationInterpolator_Mismatch(t *testing.T) {
	oi := &node.OrientationInterpolator{
		Interpolator: node.Interpolator{Key: []float64{0, 0.5, 1}},
		KeyValue:     []vec.SFRotation{{W: 0}, {W: 1}}, // 2 != 3
	}

	v := New()
	findings := v.Validate([]node.Node{oi})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for OrientationInterpolator mismatch")
	}
}

func TestPositionInterpolator_Mismatch(t *testing.T) {
	pi := &node.PositionInterpolator{
		Interpolator: node.Interpolator{Key: []float64{0, 1}},
		KeyValue:     []vec.SFVec3f{{X: 0}}, // 1 != 2
	}

	v := New()
	findings := v.Validate([]node.Node{pi})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for PositionInterpolator mismatch")
	}
}

func TestScalarInterpolator_Mismatch(t *testing.T) {
	si := &node.ScalarInterpolator{
		Interpolator: node.Interpolator{Key: []float64{0, 1}},
		KeyValue:     []float64{0, 50, 100}, // 3 != 2
	}

	v := New()
	findings := v.Validate([]node.Node{si})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for ScalarInterpolator mismatch")
	}
}

func TestNormalInterpolator_Mismatch(t *testing.T) {
	ni := &node.NormalInterpolator{
		Interpolator: node.Interpolator{Key: []float64{0, 1}},
		KeyValue:     []vec.SFVec3f{{Y: 1}, {Y: 1}, {Y: 1}}, // 3 not multiple of 2
	}

	v := New()
	findings := v.Validate([]node.Node{ni})
	errs := countErrors(findings)
	if errs < 1 {
		t.Fatal("expected error for NormalInterpolator mismatch")
	}
}

// ---------------------------------------------------------------------------
// Finding.String coverage
// ---------------------------------------------------------------------------

func TestFinding_String(t *testing.T) {
	f := Finding{
		Severity: Error,
		NodeType: "Box",
		NodeName: "myBox",
		Path:     "root/Shape",
		Message:  "size must be > 0",
	}
	s := f.String()
	if s == "" {
		t.Fatal("Finding.String() returned empty")
	}
	// Should contain all parts
	for _, want := range []string{"ERROR", "Box", "myBox", "root/Shape", "size must be > 0"} {
		if !contains(s, want) {
			t.Errorf("Finding.String() = %q, missing %q", s, want)
		}
	}
}

func TestFinding_String_Warning(t *testing.T) {
	f := Finding{Severity: Warning, Message: "test warning"}
	s := f.String()
	if !contains(s, "WARN") {
		t.Errorf("expected WARN in %q", s)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func countErrors(findings []Finding) int {
	n := 0
	for _, f := range findings {
		if f.Severity == Error {
			n++
		}
	}
	return n
}

func countWarnings(findings []Finding) int {
	n := 0
	for _, f := range findings {
		if f.Severity == Warning {
			n++
		}
	}
	return n
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ===========================================================================
// Gap-filling tests — validation pipeline (issue #42)
// ===========================================================================

// ---------------------------------------------------------------------------
// End-to-end: parse VRML → validate
// ---------------------------------------------------------------------------

func TestPipeline_ValidBox(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Shape {
  appearance Appearance { material Material { diffuseColor 1 0 0 } }
  geometry Box { size 2 2 2 }
}`)
	v := New()
	findings := v.Validate(nodes)
	if len(findings) != 0 {
		for _, f := range findings {
			t.Errorf("unexpected: %s", f)
		}
	}
}

func TestPipeline_InvalidSphereRadius(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Shape { geometry Sphere { radius 0 } }`)
	v := New()
	findings := v.Validate(nodes)
	if countErrors(findings) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", countErrors(findings), findings)
	}
}

func TestPipeline_NestedTransformWithBadGeometry(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Transform {
  children [
    Transform {
      scale 0 1 1
      children [
        Shape { geometry Box { size -1 1 1 } }
      ]
    }
  ]
}`)
	v := New()
	findings := v.Validate(nodes)
	// Should find: zero scale warning + bad box size error
	if countWarnings(findings) < 1 {
		t.Errorf("expected >= 1 warning for zero scale, got %d", countWarnings(findings))
	}
	if countErrors(findings) < 1 {
		t.Errorf("expected >= 1 error for bad box, got %d", countErrors(findings))
	}
}

func TestPipeline_MultipleShapes(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Shape { geometry Sphere { radius -1 } }
Shape { geometry Cone { bottomRadius 0 height 0 } }
Shape { geometry Cylinder { radius -2 height 0 } }`)
	v := New()
	findings := v.Validate(nodes)
	// Sphere: 1 error, Cone: 2 errors, Cylinder: 2 errors = 5 total
	if countErrors(findings) != 5 {
		t.Fatalf("expected 5 errors, got %d: %v", countErrors(findings), findings)
	}
}

func TestPipeline_IFS_BadCoordIndex(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Shape {
  geometry IndexedFaceSet {
    coord Coordinate { point [0 0 0, 1 0 0, 0 1 0] }
    coordIndex [0 1 99 -1]
  }
}`)
	v := New()
	findings := v.Validate(nodes)
	if countErrors(findings) != 1 {
		t.Fatalf("expected 1 error for OOB coordIndex, got %d", countErrors(findings))
	}
}

func TestPipeline_MaterialAllBad(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Shape {
  appearance Appearance {
    material Material {
      ambientIntensity -1
      shininess 2
      transparency -0.5
      diffuseColor 2 -1 0
    }
  }
  geometry Box { size 1 1 1 }
}`)
	v := New()
	findings := v.Validate(nodes)
	// ambientIntensity, shininess, transparency, diffuseColor = 4 errors
	if countErrors(findings) != 4 {
		t.Fatalf("expected 4 material errors, got %d: %v", countErrors(findings), findings)
	}
}

func TestPipeline_SpotLight(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
SpotLight { radius -1 beamWidth 2 cutOffAngle -1 }`)
	v := New()
	findings := v.Validate(nodes)
	if countErrors(findings) != 3 {
		t.Fatalf("expected 3 errors, got %d: %v", countErrors(findings), findings)
	}
}

func TestPipeline_ElevationGrid(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Shape {
  geometry ElevationGrid {
    xDimension 3
    zDimension 2
    xSpacing 1
    zSpacing 1
    height [0 1 2 3 4 5]
  }
}`)
	v := New()
	findings := v.Validate(nodes)
	if len(findings) != 0 {
		for _, f := range findings {
			t.Errorf("unexpected: %s", f)
		}
	}
}

func TestPipeline_Viewpoint(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Viewpoint { fieldOfView 0.785 position 0 0 10 }`)
	v := New()
	findings := v.Validate(nodes)
	if len(findings) != 0 {
		t.Fatalf("valid viewpoint should have 0 findings, got %d", len(findings))
	}
}

func TestPipeline_ViewpointBadFOV(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Viewpoint { fieldOfView 0 }`)
	v := New()
	findings := v.Validate(nodes)
	if countErrors(findings) != 1 {
		t.Fatalf("expected 1 error for bad FOV, got %d", countErrors(findings))
	}
}

func TestPipeline_LODRangeMismatch(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
LOD {
  range [10 20]
  level [
    Shape { geometry Box { size 1 1 1 } }
    Shape { geometry Sphere { radius 1 } }
  ]
}`)
	v := New()
	findings := v.Validate(nodes)
	// range has 2 entries, levels has 2 entries, should be levels-1=1 → warning
	if countWarnings(findings) != 1 {
		t.Fatalf("expected 1 warning for range/level mismatch, got %d: %v", countWarnings(findings), findings)
	}
}

func TestPipeline_SwitchOOB(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Switch {
  whichChoice 5
  choice [ Shape { geometry Box { size 1 1 1 } } ]
}`)
	v := New()
	findings := v.Validate(nodes)
	if countWarnings(findings) != 1 {
		t.Fatalf("expected 1 warning for OOB whichChoice, got %d", countWarnings(findings))
	}
}

func TestPipeline_InlineNoURL(t *testing.T) {
	nodes := parseVRML(t, `#VRML V2.0 utf8
Inline { }`)
	v := New()
	findings := v.Validate(nodes)
	if countWarnings(findings) != 1 {
		t.Fatalf("expected 1 warning for missing URL, got %d", countWarnings(findings))
	}
}

// ---------------------------------------------------------------------------
// Traversal through container types
// ---------------------------------------------------------------------------

func TestValidate_AnchorTraversesChildren(t *testing.T) {
	anchor := &node.Anchor{URL: []string{"http://example.com"}}
	anchor.Children = []node.Node{&node.Box{}} // zero-size box → error
	v := New()
	findings := v.Validate([]node.Node{anchor})
	if countErrors(findings) != 1 {
		t.Fatalf("expected 1 error from Anchor child, got %d", countErrors(findings))
	}
}

func TestValidate_BillboardTraversesChildren(t *testing.T) {
	bb := &node.Billboard{}
	bb.Children = []node.Node{&node.Sphere{}} // zero radius → error
	v := New()
	findings := v.Validate([]node.Node{bb})
	if countErrors(findings) != 1 {
		t.Fatalf("expected 1 error from Billboard child, got %d", countErrors(findings))
	}
}

func TestValidate_CollisionTraversesChildren(t *testing.T) {
	col := &node.Collision{}
	col.Children = []node.Node{&node.Cylinder{}} // zero radius/height → 2 errors
	v := New()
	findings := v.Validate([]node.Node{col})
	if countErrors(findings) != 2 {
		t.Fatalf("expected 2 errors from Collision child, got %d", countErrors(findings))
	}
}

func TestValidate_InlineTraversesChildren(t *testing.T) {
	inl := &node.Inline{URL: []string{"test.wrl"}}
	inl.Children = []node.Node{&node.Box{}} // zero-size → error
	v := New()
	findings := v.Validate([]node.Node{inl})
	if countErrors(findings) != 1 {
		t.Fatalf("expected 1 error from Inline child, got %d", countErrors(findings))
	}
}

// ---------------------------------------------------------------------------
// Nil / empty edge cases
// ---------------------------------------------------------------------------

func TestValidate_NilNodes(t *testing.T) {
	v := New()
	findings := v.Validate(nil)
	if len(findings) != 0 {
		t.Fatal("nil input should produce 0 findings")
	}
}

func TestValidate_NilNodeInSlice(t *testing.T) {
	v := New()
	findings := v.Validate([]node.Node{nil})
	if len(findings) != 0 {
		t.Fatal("nil node should produce 0 findings")
	}
}

func TestValidate_EmptyGroup(t *testing.T) {
	v := New()
	findings := v.Validate([]node.Node{&node.Group{}})
	if len(findings) != 0 {
		t.Fatal("empty group should produce 0 findings")
	}
}

// ---------------------------------------------------------------------------
// Findings() accessor
// ---------------------------------------------------------------------------

func TestFindings_Accessor(t *testing.T) {
	v := New()
	v.Validate([]node.Node{&node.Box{}}) // zero-size error
	f := v.Findings()
	if len(f) != 1 {
		t.Fatalf("expected 1 finding via Findings(), got %d", len(f))
	}
}

func TestFindings_ResetOnRevalidate(t *testing.T) {
	v := New()
	v.Validate([]node.Node{&node.Box{}})
	if len(v.Findings()) != 1 {
		t.Fatal("should have 1 finding")
	}
	// Second Validate should reset
	v.Validate([]node.Node{node.NewBox()})
	if len(v.Findings()) != 0 {
		t.Fatal("should have 0 findings after valid second pass")
	}
}

// ---------------------------------------------------------------------------
// Path tracking with DEF names
// ---------------------------------------------------------------------------

func TestPath_IncludesNodeType(t *testing.T) {
	v := New()
	tr := node.NewTransform()
	tr.Children = []node.Node{&node.Box{}} // zero-size error
	v.Validate([]node.Node{tr})
	f := v.Findings()
	if len(f) != 1 {
		t.Fatal("expected 1 finding")
	}
	s := f[0].String()
	if !contains(s, "Box") {
		t.Fatalf("path should mention Box, got: %s", s)
	}
}

func TestPath_IncludesDefName(t *testing.T) {
	v := New()
	tr := node.NewTransform()
	tr.SetName("Root")
	box := &node.Box{}
	box.SetName("MyBox")
	tr.Children = []node.Node{box}
	v.Validate([]node.Node{tr})
	f := v.Findings()
	if len(f) != 1 {
		t.Fatal("expected 1 finding")
	}
	if f[0].NodeName != "MyBox" {
		t.Fatalf("expected NodeName='MyBox', got %q", f[0].NodeName)
	}
	if !contains(f[0].Path, "Root") {
		t.Fatalf("expected path to include 'Root', got: %s", f[0].Path)
	}
}

// ---------------------------------------------------------------------------
// Appearance sub-tree traversal
// ---------------------------------------------------------------------------

func TestValidate_AppearanceTraversesMaterial(t *testing.T) {
	mat := node.NewMaterial()
	mat.AmbientIntensity = -1 // out of range
	scene := []node.Node{
		&node.Shape{
			Appearance: &node.Appearance{Material: mat},
			Geometry:   node.NewBox(),
		},
	}
	v := New()
	findings := v.Validate(scene)
	if countErrors(findings) != 1 {
		t.Fatalf("expected 1 material error, got %d", countErrors(findings))
	}
}

// ---------------------------------------------------------------------------
// countFaces edge cases
// ---------------------------------------------------------------------------

func TestCountFaces_Empty(t *testing.T) {
	if countFaces(nil) != 0 {
		t.Fatal("nil should return 0")
	}
	if countFaces([]int64{}) != 0 {
		t.Fatal("empty should return 0")
	}
}

func TestCountFaces_NoTrailingSentinel(t *testing.T) {
	// 0 1 2 (no trailing -1) → 1 face
	n := countFaces([]int64{0, 1, 2})
	if n != 1 {
		t.Fatalf("expected 1 face, got %d", n)
	}
}

func TestCountFaces_Multiple(t *testing.T) {
	// Two faces with sentinels
	n := countFaces([]int64{0, 1, 2, -1, 3, 4, 5, -1})
	if n != 2 {
		t.Fatalf("expected 2 faces, got %d", n)
	}
}

func TestCountFaces_MixedTrailing(t *testing.T) {
	// Three faces, last without -1
	n := countFaces([]int64{0, 1, 2, -1, 3, 4, 5, -1, 6, 7, 8})
	if n != 3 {
		t.Fatalf("expected 3 faces, got %d", n)
	}
}

// ---------------------------------------------------------------------------
// Severity.String
// ---------------------------------------------------------------------------

func TestSeverity_String(t *testing.T) {
	if Error.String() != "ERROR" {
		t.Fatalf("expected ERROR, got %s", Error.String())
	}
	if Warning.String() != "WARN" {
		t.Fatalf("expected WARN, got %s", Warning.String())
	}
}

// ---------------------------------------------------------------------------
// Helper: parse VRML text into nodes
// ---------------------------------------------------------------------------

func parseVRML(t *testing.T, vrml string) []node.Node {
	t.Helper()
	p := parser.NewParser(strings.NewReader(vrml))
	nodes := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	return nodes
}
