package serializer

import (
	"bytes"
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func TestRoundTripEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := Encode(&buf, nil); err != nil {
		t.Fatal(err)
	}
	nodes, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes, got %d", len(nodes))
	}
}

func TestRoundTripSimpleScene(t *testing.T) {
	scene := []node.Node{
		&node.Shape{
			Appearance: &node.Appearance{
				Material: &node.Material{
					AmbientIntensity: 0.5,
					DiffuseColor:     vec.SFColor{R: 1, G: 0, B: 0, A: 1},
					Shininess:        0.8,
				},
			},
			Geometry: &node.Box{
				BaseGeometry: node.BaseGeometry{
					BaseNode: node.BaseNode{Name: "myBox"},
				},
				Size: vec.SFVec3f{X: 3, Y: 4, Z: 5},
			},
		},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, scene); err != nil {
		t.Fatal("encode:", err)
	}

	nodes, err := Decode(&buf)
	if err != nil {
		t.Fatal("decode:", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	shape, ok := nodes[0].(*node.Shape)
	if !ok {
		t.Fatal("expected *Shape")
	}
	if shape.Appearance == nil || shape.Appearance.Material == nil {
		t.Fatal("missing appearance/material")
	}
	m := shape.Appearance.Material
	if m.AmbientIntensity != 0.5 {
		t.Errorf("ambientIntensity: got %g, want 0.5", m.AmbientIntensity)
	}
	if m.DiffuseColor.R != 1 || m.DiffuseColor.G != 0 || m.DiffuseColor.B != 0 {
		t.Errorf("diffuseColor: got %v", m.DiffuseColor)
	}
	if m.Shininess != 0.8 {
		t.Errorf("shininess: got %g, want 0.8", m.Shininess)
	}

	box, ok := shape.Geometry.(*node.Box)
	if !ok {
		t.Fatal("expected *Box geometry")
	}
	if box.GetName() != "myBox" {
		t.Errorf("name: got %q, want %q", box.GetName(), "myBox")
	}
	if box.Size.X != 3 || box.Size.Y != 4 || box.Size.Z != 5 {
		t.Errorf("size: got %v", box.Size)
	}
}

func TestRoundTripTransformHierarchy(t *testing.T) {
	scene := []node.Node{
		&node.Transform{
			GroupingNode: node.GroupingNode{
				BaseNode: node.BaseNode{Name: "root"},
				Children: []node.Node{
					&node.Shape{
						Geometry: node.NewSphere(),
					},
					&node.Transform{
						GroupingNode: node.GroupingNode{
							Children: []node.Node{
								&node.Shape{Geometry: node.NewCone()},
							},
						},
						Translation: vec.SFVec3f{X: 5, Y: 0, Z: 0},
						Scale:       vec.SFVec3f{X: 1, Y: 1, Z: 1},
					},
				},
			},
			Translation: vec.SFVec3f{X: 0, Y: 2, Z: 0},
			Scale:       vec.SFVec3f{X: 1, Y: 1, Z: 1},
		},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, scene); err != nil {
		t.Fatal(err)
	}

	nodes, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	xf, ok := nodes[0].(*node.Transform)
	if !ok {
		t.Fatal("expected *Transform")
	}
	if xf.GetName() != "root" {
		t.Errorf("name: got %q", xf.GetName())
	}
	if xf.Translation.Y != 2 {
		t.Errorf("translation.Y: got %g", xf.Translation.Y)
	}
	if len(xf.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(xf.Children))
	}

	// First child is Shape with Sphere
	s1, ok := xf.Children[0].(*node.Shape)
	if !ok {
		t.Fatal("child 0: expected *Shape")
	}
	if _, ok := s1.Geometry.(*node.Sphere); !ok {
		t.Error("child 0: expected Sphere geometry")
	}

	// Second child is Transform with Cone
	xf2, ok := xf.Children[1].(*node.Transform)
	if !ok {
		t.Fatal("child 1: expected *Transform")
	}
	if xf2.Translation.X != 5 {
		t.Errorf("child 1 translation.X: got %g", xf2.Translation.X)
	}
	if len(xf2.Children) != 1 {
		t.Fatalf("child 1: expected 1 child, got %d", len(xf2.Children))
	}
	s2, ok := xf2.Children[0].(*node.Shape)
	if !ok {
		t.Fatal("child 1.0: expected *Shape")
	}
	if _, ok := s2.Geometry.(*node.Cone); !ok {
		t.Error("child 1.0: expected Cone geometry")
	}
}

func TestRoundTripIndexedFaceSet(t *testing.T) {
	scene := []node.Node{
		&node.IndexedFaceSet{
			DataSet: node.DataSet{
				BaseGeometry: node.BaseGeometry{
					Ccw:     true,
					IsSolid: true,
				},
				ColorPerVertex:  true,
				NormalPerVertex: true,
				CoordIndex:      []int64{0, 1, 2, -1, 2, 3, 0, -1},
				Coord: &node.Coordinate{
					Point: []vec.SFVec3f{
						{X: 0, Y: 0, Z: 0},
						{X: 1, Y: 0, Z: 0},
						{X: 1, Y: 1, Z: 0},
						{X: 0, Y: 1, Z: 0},
					},
				},
				Color: &node.ColorNode{
					Color: []vec.SFColor{
						{R: 1, G: 0, B: 0},
						{R: 0, G: 1, B: 0},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, scene); err != nil {
		t.Fatal(err)
	}

	nodes, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	ifs, ok := nodes[0].(*node.IndexedFaceSet)
	if !ok {
		t.Fatal("expected *IndexedFaceSet")
	}
	if len(ifs.CoordIndex) != 8 {
		t.Errorf("coordIndex: got %d, want 8", len(ifs.CoordIndex))
	}
	if ifs.CoordIndex[3] != -1 {
		t.Errorf("coordIndex[3]: got %d, want -1", ifs.CoordIndex[3])
	}
	if ifs.Coord == nil || len(ifs.Coord.Point) != 4 {
		t.Fatal("expected 4 coord points")
	}
	if ifs.Color == nil || len(ifs.Color.Color) != 2 {
		t.Fatal("expected 2 colors")
	}
	if ifs.Color.Color[0].R != 1 {
		t.Error("expected red first color")
	}
}

func TestRoundTripInterpolators(t *testing.T) {
	scene := []node.Node{
		&node.ColorInterpolator{
			Interpolator: node.Interpolator{Key: []float64{0, 0.5, 1.0}},
			KeyValue: []vec.SFColor{
				{R: 1, G: 0, B: 0},
				{R: 0, G: 1, B: 0},
				{R: 0, G: 0, B: 1},
			},
		},
		&node.PositionInterpolator{
			Interpolator: node.Interpolator{Key: []float64{0, 1}},
			KeyValue: []vec.SFVec3f{
				{X: 0, Y: 0, Z: 0},
				{X: 10, Y: 20, Z: 30},
			},
		},
		&node.ScalarInterpolator{
			Interpolator: node.Interpolator{Key: []float64{0, 1}},
			KeyValue:     []float64{0, 100},
		},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, scene); err != nil {
		t.Fatal(err)
	}

	nodes, err := Decode(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	ci, ok := nodes[0].(*node.ColorInterpolator)
	if !ok {
		t.Fatal("expected *ColorInterpolator")
	}
	if len(ci.Key) != 3 || len(ci.KeyValue) != 3 {
		t.Error("ColorInterpolator key/keyValue count mismatch")
	}

	pi, ok := nodes[1].(*node.PositionInterpolator)
	if !ok {
		t.Fatal("expected *PositionInterpolator")
	}
	if pi.KeyValue[1].X != 10 || pi.KeyValue[1].Y != 20 {
		t.Error("PositionInterpolator keyValue mismatch")
	}

	si, ok := nodes[2].(*node.ScalarInterpolator)
	if !ok {
		t.Fatal("expected *ScalarInterpolator")
	}
	if si.KeyValue[1] != 100 {
		t.Errorf("ScalarInterpolator keyValue: got %g", si.KeyValue[1])
	}
}

func TestRoundTripAllNodeTypes(t *testing.T) {
	// Build a scene with every node type to ensure all can round-trip.
	scene := []node.Node{
		node.NewBox(),
		node.NewSphere(),
		node.NewCone(),
		node.NewCylinder(),
		node.NewMaterial(),
		node.NewViewpoint(),
		node.NewSpotLight(),
		node.NewDirectionalLight(),
		node.NewPointLight(),
		node.NewTransform(),
		node.NewFog(),
		node.NewNavigationInfo(),
		node.NewFontStyle(),
		node.NewBillboard(),
		node.NewCollision(),
		node.NewImageTexture(),
		node.NewMovieTexture(),
		node.NewPixelTexture(),
		node.NewTextureTransform(),
		node.NewExtrusion(),
		node.NewTimeSensor(),
		node.NewTouchSensor(),
		node.NewProximitySensor(),
		node.NewCylinderSensor(),
		node.NewPlaneSensor(),
		node.NewSphereSensor(),
		node.NewVisibilitySensor(),
		node.NewAudioClip(),
		node.NewSound(),
		node.NewSwitch(),
		node.NewLOD(),
		node.NewIndexedFaceSet(),
		node.NewIndexedLineSet(),
		node.NewPointSet(),
		node.NewElevationGrid(),
		&node.Group{},
		&node.Anchor{},
		&node.Inline{},
		&node.WorldInfo{Title: "test"},
		&node.Script{URL: []string{"test.js"}},
		&node.Background{SkyColor: []vec.SFColor{{R: 0.5, G: 0.5, B: 1}}},
		&node.Text{String: []string{"Hello"}},
		&node.ColorNode{Color: []vec.SFColor{{R: 1}}},
		&node.Coordinate{Point: []vec.SFVec3f{{X: 1}}},
		&node.NormalNode{Vector: []vec.SFVec3f{{Y: 1}}},
		&node.TextureCoordinate{Point: []vec.SFVec2f{{X: 0.5, Y: 0.5}}},
		&node.Appearance{Material: node.NewMaterial()},
		&node.Shape{Geometry: node.NewBox()},
		&node.OrientationInterpolator{
			Interpolator: node.Interpolator{Key: []float64{0, 1}},
			KeyValue:     []vec.SFRotation{{X: 0, Y: 1, Z: 0, W: 0}, {X: 0, Y: 1, Z: 0, W: float64(math.Pi)}},
		},
		&node.CoordinateInterpolator{
			Interpolator: node.Interpolator{Key: []float64{0, 1}},
			KeyValue:     []vec.SFVec3f{{X: 0}, {X: 1}},
		},
		&node.NormalInterpolator{
			Interpolator: node.Interpolator{Key: []float64{0}},
			KeyValue:     []vec.SFVec3f{{Y: 1}},
		},
	}

	var buf bytes.Buffer
	if err := Encode(&buf, scene); err != nil {
		t.Fatal("encode:", err)
	}
	t.Logf("encoded %d nodes in %d bytes", len(scene), buf.Len())

	nodes, err := Decode(&buf)
	if err != nil {
		t.Fatal("decode:", err)
	}
	if len(nodes) != len(scene) {
		t.Fatalf("expected %d nodes, got %d", len(scene), len(nodes))
	}

	// Spot-check a few by type rather than fragile indices
	foundWorldInfo := false
	foundBackground := false
	for _, n := range nodes {
		switch v := n.(type) {
		case *node.WorldInfo:
			if v.Title != "test" {
				t.Errorf("WorldInfo.Title: got %q, want %q", v.Title, "test")
			}
			foundWorldInfo = true
		case *node.Background:
			if len(v.SkyColor) != 1 {
				t.Errorf("Background.SkyColor: got %d, want 1", len(v.SkyColor))
			}
			foundBackground = true
		}
	}
	if !foundWorldInfo {
		t.Error("WorldInfo not found in decoded nodes")
	}
	if !foundBackground {
		t.Error("Background not found in decoded nodes")
	}
}

func TestBadMagic(t *testing.T) {
	buf := bytes.NewBufferString("XXXX")
	_, err := Decode(buf)
	if err == nil {
		t.Fatal("expected error for bad magic")
	}
}

func TestTruncatedInput(t *testing.T) {
	_, err := Decode(bytes.NewReader([]byte("VR")))
	if err == nil {
		t.Fatal("expected error for truncated magic")
	}
}
