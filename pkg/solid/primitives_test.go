package solid

import (
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

var yellow = vec.SFColor{R: 1, G: 0.85, B: 0.1, A: 1}

// ---------------------------------------------------------------------------
// Lamina
// ---------------------------------------------------------------------------

func TestMakeLamina_Triangle(t *testing.T) {
	verts := []vec.SFVec3f{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	s := MakeLamina(verts, yellow)
	if s == nil {
		t.Fatal("MakeLamina returned nil")
	}
	f, e, v := s.Stats()
	euler := f + v - e
	if euler != 2 {
		t.Errorf("triangle lamina: euler = %d (F=%d E=%d V=%d), want 2", euler, f, e, v)
	}
	if !s.IsLamina() {
		t.Errorf("triangle lamina: IsLamina() = false, want true")
	}
	if f != 2 {
		t.Errorf("triangle lamina: faces = %d, want 2", f)
	}
	errs := s.VerifyDetailed()
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("verify: %v", err)
		}
	}
}

func TestMakeLamina_Square(t *testing.T) {
	verts := []vec.SFVec3f{
		{X: -1, Y: -1, Z: 0},
		{X: 1, Y: -1, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: -1, Y: 1, Z: 0},
	}
	s := MakeLamina(verts, yellow)
	if s == nil {
		t.Fatal("MakeLamina returned nil")
	}
	f, e, v := s.Stats()
	euler := f + v - e
	if euler != 2 {
		t.Errorf("square lamina: euler = %d (F=%d E=%d V=%d), want 2", euler, f, e, v)
	}
	if f != 2 {
		t.Errorf("square lamina: faces = %d, want 2", f)
	}
}

func TestMakeLamina_Pentagon(t *testing.T) {
	n := 5
	verts := make([]vec.SFVec3f, n)
	for i := range verts {
		a := 2 * math.Pi * float64(i) / float64(n)
		verts[i] = vec.SFVec3f{X: float32(math.Cos(a)), Y: float32(math.Sin(a)), Z: 0}
	}
	s := MakeLamina(verts, yellow)
	if s == nil {
		t.Fatal("MakeLamina returned nil")
	}
	f, e, v := s.Stats()
	euler := f + v - e
	if euler != 2 {
		t.Errorf("pentagon lamina: euler = %d (F=%d E=%d V=%d), want 2", euler, f, e, v)
	}
	if v != n {
		t.Errorf("pentagon lamina: verts = %d, want %d", v, n)
	}
}

func TestMakeLamina_TooFewVertices(t *testing.T) {
	s := MakeLamina([]vec.SFVec3f{{}, {}}, yellow)
	if s != nil {
		t.Error("MakeLamina with 2 vertices should return nil")
	}
	s = MakeLamina(nil, yellow)
	if s != nil {
		t.Error("MakeLamina with nil should return nil")
	}
}

// ---------------------------------------------------------------------------
// Circle
// ---------------------------------------------------------------------------

func TestMakeCircle(t *testing.T) {
	s := MakeCircle(0, 0, 1.0, 0, 8, yellow)
	if s == nil {
		t.Fatal("MakeCircle returned nil")
	}
	f, e, v := s.Stats()
	euler := f + v - e
	if euler != 2 {
		t.Errorf("circle: euler = %d (F=%d E=%d V=%d), want 2", euler, f, e, v)
	}
	if !s.IsLamina() {
		t.Error("circle: IsLamina() = false, want true")
	}
	if v != 8 {
		t.Errorf("circle: verts = %d, want 8", v)
	}
}

func TestMakeCircle_TooFewSegs(t *testing.T) {
	s := MakeCircle(0, 0, 1.0, 0, 2, yellow)
	if s != nil {
		t.Error("MakeCircle with n=2 should return nil")
	}
}

// ---------------------------------------------------------------------------
// Cube
// ---------------------------------------------------------------------------

func TestMakeCube(t *testing.T) {
	s := MakeCube(1.0, yellow)
	if s == nil {
		t.Fatal("MakeCube returned nil")
	}
	f, e, v := s.Stats()
	euler := f + v - e
	if euler != 2 {
		t.Errorf("cube: euler = %d (F=%d E=%d V=%d), want 2", euler, f, e, v)
	}
	if f != 6 || e != 12 || v != 8 {
		t.Errorf("cube: F=%d E=%d V=%d, want F=6 E=12 V=8", f, e, v)
	}
	vol := s.Volume()
	if math.Abs(float64(vol)-8.0) > 0.01 {
		t.Errorf("cube: volume = %.3f, want 8.0", vol)
	}
	errs := s.VerifyDetailed()
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("verify: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Prism
// ---------------------------------------------------------------------------

func TestMakePrism(t *testing.T) {
	s := MakePrism(2.0, yellow)
	if s == nil {
		t.Fatal("MakePrism returned nil")
	}
	f, e, v := s.Stats()
	euler := f + v - e
	if euler != 2 {
		t.Errorf("prism: euler = %d (F=%d E=%d V=%d), want 2", euler, f, e, v)
	}
	if f != 5 || e != 9 || v != 6 {
		t.Errorf("prism: F=%d E=%d V=%d, want F=5 E=9 V=6", f, e, v)
	}
}

// ---------------------------------------------------------------------------
// Cylinder
// ---------------------------------------------------------------------------

func TestMakeCylinder(t *testing.T) {
	s := MakeCylinder(1.0, 2.0, 8, yellow)
	if s == nil {
		t.Fatal("MakeCylinder returned nil")
	}
	f, e, v := s.Stats()
	euler := f + v - e
	if euler != 2 {
		t.Errorf("cylinder: euler = %d (F=%d E=%d V=%d), want 2", euler, f, e, v)
	}
	// Cylinder with n segments: 2 caps + n sides = n+2 faces, 3n edges, 2n verts
	if f != 10 || e != 24 || v != 16 {
		t.Errorf("cylinder n=8: F=%d E=%d V=%d, want F=10 E=24 V=16", f, e, v)
	}
	errs := s.VerifyDetailed()
	if len(errs) > 0 {
		for _, err := range errs {
			t.Errorf("verify: %v", err)
		}
	}
}

func TestMakeCylinder_TooFewSegs(t *testing.T) {
	s := MakeCylinder(1.0, 2.0, 2, yellow)
	if s != nil {
		t.Error("MakeCylinder with n=2 should return nil")
	}
}

// ---------------------------------------------------------------------------
// Torus
// ---------------------------------------------------------------------------

func TestMakeTorus(t *testing.T) {
	s := MakeTorus(2.0, 0.5, 12, 8, yellow)
	if s == nil {
		t.Fatal("MakeTorus returned nil")
	}
	f, e, v := s.Stats()
	euler := f + v - e
	if euler != 0 {
		t.Errorf("torus: euler = %d (F=%d E=%d V=%d), want 0 (genus 1)", euler, f, e, v)
	}
	// Torus with major=m, minor=n: m*n faces, 2*m*n edges, m*n verts → euler=0
	wantF := 12 * 8
	wantE := 2 * 12 * 8
	wantV := 12 * 8
	if f != wantF || e != wantE || v != wantV {
		t.Errorf("torus 12x8: F=%d E=%d V=%d, want F=%d E=%d V=%d", f, e, v, wantF, wantE, wantV)
	}
}

func TestMakeTorus_TooFewSegs(t *testing.T) {
	if s := MakeTorus(2.0, 0.5, 2, 8, yellow); s != nil {
		t.Error("MakeTorus with majorSegs=2 should return nil")
	}
	if s := MakeTorus(2.0, 0.5, 8, 2, yellow); s != nil {
		t.Error("MakeTorus with minorSegs=2 should return nil")
	}
}

// ---------------------------------------------------------------------------
// Sphere
// ---------------------------------------------------------------------------

func TestMakeSphere(t *testing.T) {
	s := MakeSphere(1.0, 8, 12, yellow)
	if s == nil {
		t.Fatal("MakeSphere returned nil")
	}
	f, e, v := s.Stats()
	euler := f + v - e
	if euler != 2 {
		t.Errorf("sphere: euler = %d (F=%d E=%d V=%d), want 2", euler, f, e, v)
	}
}

func TestMakeSphere_TooFewSegs(t *testing.T) {
	if s := MakeSphere(1.0, 2, 8, yellow); s != nil {
		t.Error("MakeSphere with latSegs=2 should return nil")
	}
	if s := MakeSphere(1.0, 8, 2, yellow); s != nil {
		t.Error("MakeSphere with lonSegs=2 should return nil")
	}
}
