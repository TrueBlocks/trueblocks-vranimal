package validator

import (
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
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
			Interpolator: node.Interpolator{Key: []float32{0, 0.5, 1.0}},
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
