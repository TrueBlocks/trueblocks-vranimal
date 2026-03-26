package solid

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// TestGenerateExamples creates VRML example files for each solid primitive.
// Run with: go test ./pkg/solid/ -run TestGenerateExamples -v
func TestGenerateExamples(t *testing.T) {
	dir := filepath.Join("..", "..", "examples")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		build func() *Solid
	}{
		{"lamina_triangle", func() *Solid {
			return MakeLamina([]vec.SFVec3f{
				{X: 0, Y: 2, Z: 0},
				{X: -1.5, Y: -1, Z: 0},
				{X: 1.5, Y: -1, Z: 0},
			}, vec.SFColor{R: 0.8, G: 0.2, B: 0.2, A: 1})
		}},
		{"lamina_pentagon", func() *Solid {
			return MakeLamina([]vec.SFVec3f{
				{X: 0, Y: 2, Z: 0},
				{X: -1.9, Y: 0.6, Z: 0},
				{X: -1.2, Y: -1.6, Z: 0},
				{X: 1.2, Y: -1.6, Z: 0},
				{X: 1.9, Y: 0.6, Z: 0},
			}, vec.SFColor{R: 0.2, G: 0.6, B: 0.8, A: 1})
		}},
		{"circle", func() *Solid {
			return MakeCircle(0, 0, 2, 0, 16, vec.SFColor{R: 0.2, G: 0.8, B: 0.3, A: 1})
		}},
		{"cube", func() *Solid {
			return MakeCube(1.5, vec.SFColor{R: 0.9, G: 0.6, B: 0.1, A: 1})
		}},
		{"prism", func() *Solid {
			return MakePrism(3, vec.SFColor{R: 0.5, G: 0.2, B: 0.8, A: 1})
		}},
		{"cylinder", func() *Solid {
			return MakeCylinder(1, 3, 12, vec.SFColor{R: 0.1, G: 0.5, B: 0.7, A: 1})
		}},
		{"torus", func() *Solid {
			return MakeTorus(2, 0.5, 12, 8, vec.SFColor{R: 0.8, G: 0.3, B: 0.5, A: 1})
		}},
		{"sphere", func() *Solid {
			return MakeSphere(2, 8, 12, vec.SFColor{R: 0.3, G: 0.7, B: 0.9, A: 1})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.build()
			if s == nil {
				t.Fatalf("failed to build %s", tt.name)
			}
			path := filepath.Join(dir, tt.name+".wrl")
			if err := s.ExportVRMLFile(path); err != nil {
				t.Fatalf("failed to export %s: %v", tt.name, err)
			}
			t.Logf("wrote %s", path)
		})
	}

	// Also create a combined showcase file with all primitives.
	t.Run("showcase", func(t *testing.T) {
		solids := make([]*Solid, 0, len(tests))
		translations := make([]vec.SFVec3f, 0, len(tests))
		xOffset := float32(-12)
		for _, tt := range tests {
			s := tt.build()
			if s == nil {
				t.Fatalf("failed to build %s for showcase", tt.name)
			}
			solids = append(solids, s)
			translations = append(translations, vec.SFVec3f{X: xOffset, Y: 0, Z: 0})
			xOffset += 4
		}
		path := filepath.Join(dir, "primitives_showcase.wrl")
		if err := ExportMultiVRMLFile(path, solids, translations); err != nil {
			t.Fatalf("failed to export showcase: %v", err)
		}
		t.Logf("wrote %s", path)
	})
}
