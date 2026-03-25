package vec

import (
	"math"
	"testing"
)

const eps = 1e-5

func approx(a, b float32) bool {
	return float64(math.Abs(float64(a-b))) < eps
}

// ---------------------------------------------------------------------------
// SFVec2f
// ---------------------------------------------------------------------------

func TestVec2f_New(t *testing.T) {
	v := NewVec2f(3, 4)
	if v.X != 3 || v.Y != 4 {
		t.Fatalf("got %v", v)
	}
}

func TestVec2f_Add(t *testing.T) {
	r := NewVec2f(1, 2).Add(NewVec2f(3, 4))
	if r != (SFVec2f{4, 6}) {
		t.Fatalf("got %v", r)
	}
}

func TestVec2f_Sub(t *testing.T) {
	r := NewVec2f(5, 7).Sub(NewVec2f(2, 3))
	if r != (SFVec2f{3, 4}) {
		t.Fatalf("got %v", r)
	}
}

func TestVec2f_Scale(t *testing.T) {
	r := NewVec2f(2, 3).Scale(2)
	if r != (SFVec2f{4, 6}) {
		t.Fatalf("got %v", r)
	}
}

func TestVec2f_Dot(t *testing.T) {
	d := NewVec2f(1, 0).Dot(NewVec2f(0, 1))
	if d != 0 {
		t.Fatalf("expected 0, got %g", d)
	}
	d = NewVec2f(3, 4).Dot(NewVec2f(3, 4))
	if d != 25 {
		t.Fatalf("expected 25, got %g", d)
	}
}

func TestVec2f_Length(t *testing.T) {
	l := NewVec2f(3, 4).Length()
	if !approx(l, 5) {
		t.Fatalf("expected 5, got %g", l)
	}
}

func TestVec2f_Normalize(t *testing.T) {
	n := NewVec2f(0, 5).Normalize()
	if !approx(n.X, 0) || !approx(n.Y, 1) {
		t.Fatalf("got %v", n)
	}
}

func TestVec2f_NormalizeZero(t *testing.T) {
	n := SFVec2f{}.Normalize()
	if n != (SFVec2f{}) {
		t.Fatalf("expected zero vec, got %v", n)
	}
}

func TestVec2f_Eq(t *testing.T) {
	a := NewVec2f(1, 2)
	b := NewVec2f(1, 2)
	if !a.Eq(b) {
		t.Fatal("expected equal")
	}
	if a.Eq(NewVec2f(1, 3)) {
		t.Fatal("expected not equal")
	}
}

func TestVec2f_Negate(t *testing.T) {
	r := NewVec2f(1, -2).Negate()
	if r != (SFVec2f{-1, 2}) {
		t.Fatalf("got %v", r)
	}
}

func TestVec2f_Index(t *testing.T) {
	v := NewVec2f(10, 20)
	if v.Index(0) != 10 || v.Index(1) != 20 {
		t.Fatalf("got %g, %g", v.Index(0), v.Index(1))
	}
}

func TestVec2f_String(t *testing.T) {
	s := NewVec2f(1, 2).String()
	if s != "1 2" {
		t.Fatalf("got %q", s)
	}
}

// ---------------------------------------------------------------------------
// SFVec3f
// ---------------------------------------------------------------------------

func TestVec3f_New(t *testing.T) {
	v := NewVec3f(1, 2, 3)
	if v.X != 1 || v.Y != 2 || v.Z != 3 {
		t.Fatalf("got %v", v)
	}
}

func TestVec3f_Add(t *testing.T) {
	r := NewVec3f(1, 2, 3).Add(NewVec3f(4, 5, 6))
	if r != (SFVec3f{5, 7, 9}) {
		t.Fatalf("got %v", r)
	}
}

func TestVec3f_Sub(t *testing.T) {
	r := NewVec3f(5, 7, 9).Sub(NewVec3f(1, 2, 3))
	if r != (SFVec3f{4, 5, 6}) {
		t.Fatalf("got %v", r)
	}
}

func TestVec3f_Scale(t *testing.T) {
	r := NewVec3f(1, 2, 3).Scale(3)
	if r != (SFVec3f{3, 6, 9}) {
		t.Fatalf("got %v", r)
	}
}

func TestVec3f_Dot(t *testing.T) {
	d := XAxis.Dot(YAxis)
	if d != 0 {
		t.Fatalf("expected 0, got %g", d)
	}
}

func TestVec3f_Cross(t *testing.T) {
	r := XAxis.Cross(YAxis)
	if r != ZAxis {
		t.Fatalf("expected Z axis, got %v", r)
	}
	r = YAxis.Cross(XAxis)
	if r != ZAxis.Negate() {
		t.Fatalf("expected -Z axis, got %v", r)
	}
}

func TestVec3f_Length(t *testing.T) {
	l := NewVec3f(2, 3, 6).Length()
	if !approx(l, 7) {
		t.Fatalf("expected 7, got %g", l)
	}
}

func TestVec3f_Normalize(t *testing.T) {
	n := NewVec3f(0, 0, 5).Normalize()
	if !approx(n.Z, 1) {
		t.Fatalf("got %v", n)
	}
}

func TestVec3f_NormalizeZero(t *testing.T) {
	n := Vec3fZero.Normalize()
	if n != Vec3fZero {
		t.Fatalf("expected zero, got %v", n)
	}
}

func TestVec3f_Negate(t *testing.T) {
	r := NewVec3f(1, -2, 3).Negate()
	if r != (SFVec3f{-1, 2, -3}) {
		t.Fatalf("got %v", r)
	}
}

func TestVec3f_Eq(t *testing.T) {
	if !XAxis.Eq(SFVec3f{1, 0, 0}) {
		t.Fatal("expected equal")
	}
}

func TestVec3f_Index(t *testing.T) {
	v := NewVec3f(10, 20, 30)
	if v.Index(0) != 10 || v.Index(1) != 20 || v.Index(2) != 30 {
		t.Fatal("index mismatch")
	}
}

func TestVec3f_String(t *testing.T) {
	s := NewVec3f(1, 2, 3).String()
	if s != "1 2 3" {
		t.Fatalf("got %q", s)
	}
}

func TestVec3f_Lerp(t *testing.T) {
	a := NewVec3f(0, 0, 0)
	b := NewVec3f(10, 20, 30)
	mid := a.Lerp(b, 0.5)
	if !approx(mid.X, 5) || !approx(mid.Y, 10) || !approx(mid.Z, 15) {
		t.Fatalf("got %v", mid)
	}
	if a.Lerp(b, 0) != a {
		t.Fatal("lerp 0 should equal a")
	}
	if a.Lerp(b, 1) != b {
		t.Fatal("lerp 1 should equal b")
	}
}

func TestVec3f_Constants(t *testing.T) {
	if Vec3fZero != (SFVec3f{}) {
		t.Fatal("zero wrong")
	}
	if XAxis != (SFVec3f{1, 0, 0}) {
		t.Fatal("X wrong")
	}
	if YAxis != (SFVec3f{0, 1, 0}) {
		t.Fatal("Y wrong")
	}
	if ZAxis != (SFVec3f{0, 0, 1}) {
		t.Fatal("Z wrong")
	}
}

// ---------------------------------------------------------------------------
// SFVec4f
// ---------------------------------------------------------------------------

func TestVec4f_New(t *testing.T) {
	v := NewVec4f(1, 2, 3, 4)
	if v.X != 1 || v.Y != 2 || v.Z != 3 || v.W != 4 {
		t.Fatalf("got %v", v)
	}
}

// ---------------------------------------------------------------------------
// SFColor
// ---------------------------------------------------------------------------

func TestColor_New(t *testing.T) {
	c := NewColor(1, 0, 0)
	if c.A != 1 {
		t.Fatal("alpha should be 1")
	}
	if c.R != 1 || c.G != 0 || c.B != 0 {
		t.Fatalf("got %v", c)
	}
}

func TestColor_NewA(t *testing.T) {
	c := NewColorA(1, 1, 1, 0.5)
	if c.A != 0.5 {
		t.Fatal("alpha should be 0.5")
	}
}

func TestColor_Add(t *testing.T) {
	r := Red.Add(Green)
	if r != (SFColor{1, 1, 0, 2}) {
		t.Fatalf("got %v", r)
	}
}

func TestColor_Sub(t *testing.T) {
	r := White.Sub(Red)
	if r != (SFColor{0, 1, 1, 0}) {
		t.Fatalf("got %v", r)
	}
}

func TestColor_Scale(t *testing.T) {
	r := Red.Scale(0.5)
	if !approx(r.R, 0.5) || r.G != 0 {
		t.Fatalf("got %v", r)
	}
}

func TestColor_Eq(t *testing.T) {
	if !Red.Eq(SFColor{1, 0, 0, 1}) {
		t.Fatal("should be equal")
	}
	if Red.Eq(Blue) {
		t.Fatal("should not be equal")
	}
}

func TestColor_NormalizeColor(t *testing.T) {
	c := SFColor{-0.5, 1.5, 0.5, 2.0}
	n := c.NormalizeColor()
	if n.R != 0 || n.G != 1 || n.B != 0.5 || n.A != 1 {
		t.Fatalf("got %v", n)
	}
}

func TestColor_Vec3f(t *testing.T) {
	v := Red.Vec3f()
	if v != (SFVec3f{1, 0, 0}) {
		t.Fatalf("got %v", v)
	}
}

func TestColor_String(t *testing.T) {
	s := Red.String()
	if s != "1 0 0 1" {
		t.Fatalf("got %q", s)
	}
}

func TestColor_Predefined(t *testing.T) {
	if Black != (SFColor{0, 0, 0, 1}) {
		t.Fatal("Black wrong")
	}
	if White != (SFColor{1, 1, 1, 1}) {
		t.Fatal("White wrong")
	}
	if Green != (SFColor{0, 1, 0, 1}) {
		t.Fatal("Green wrong")
	}
	if Blue != (SFColor{0, 0, 1, 1}) {
		t.Fatal("Blue wrong")
	}
	if Yellow != (SFColor{1, 1, 0, 1}) {
		t.Fatal("Yellow wrong")
	}
	if Cyan != (SFColor{0, 1, 1, 1}) {
		t.Fatal("Cyan wrong")
	}
	if Magenta != (SFColor{1, 0, 1, 1}) {
		t.Fatal("Magenta wrong")
	}
	if !approx(Grey.R, 0.5) {
		t.Fatal("Grey wrong")
	}
}

// ---------------------------------------------------------------------------
// SFRotation
// ---------------------------------------------------------------------------

func TestRotation_New(t *testing.T) {
	r := NewRotation(0, 1, 0, 1.57)
	if r.X != 0 || r.Y != 1 || r.Z != 0 || r.W != 1.57 {
		t.Fatalf("got %v", r)
	}
}

func TestRotation_Axis(t *testing.T) {
	r := NewRotation(0, 0, 5, 1.0)
	a := r.Axis()
	if !approx(a.Z, 1) || !approx(a.X, 0) {
		t.Fatalf("expected Z axis, got %v", a)
	}
}

func TestRotation_Angle(t *testing.T) {
	r := NewRotation(0, 1, 0, 3.14)
	if r.Angle() != 3.14 {
		t.Fatal("angle mismatch")
	}
}

func TestRotation_Eq(t *testing.T) {
	a := NewRotation(0, 1, 0, 0)
	b := NewRotation(0, 1, 0, 0)
	if !a.Eq(b) {
		t.Fatal("should be equal")
	}
}

func TestRotation_String(t *testing.T) {
	s := NewRotation(0, 1, 0, 1.57).String()
	if s != "0 1 0 1.57" {
		t.Fatalf("got %q", s)
	}
}

func TestSlerpRotation_Identity(t *testing.T) {
	a := NewRotation(0, 1, 0, 0)
	b := NewRotation(0, 1, 0, float32(math.Pi))
	mid := SlerpRotation(a, b, 0.5)
	if !approx(mid.W, float32(math.Pi/2)) {
		t.Fatalf("expected pi/2, got %g", mid.W)
	}
}

func TestSlerpRotation_Endpoints(t *testing.T) {
	a := NewRotation(0, 1, 0, 0.5)
	b := NewRotation(0, 1, 0, 1.5)
	r0 := SlerpRotation(a, b, 0)
	if !approx(r0.W, a.W) {
		t.Fatalf("t=0: expected %g, got %g", a.W, r0.W)
	}
	r1 := SlerpRotation(a, b, 1)
	if !approx(r1.W, b.W) {
		t.Fatalf("t=1: expected %g, got %g", b.W, r1.W)
	}
}

// ---------------------------------------------------------------------------
// SFImage
// ---------------------------------------------------------------------------

func TestImage_New(t *testing.T) {
	img := NewImage(4, 4, 3)
	if img.Width != 4 || img.Height != 4 || img.NumComponents != 3 {
		t.Fatalf("got %v", img)
	}
	if len(img.Pixels) != 48 {
		t.Fatalf("expected 48 pixels, got %d", len(img.Pixels))
	}
}

// ---------------------------------------------------------------------------
// Matrix
// ---------------------------------------------------------------------------

func TestMatrix_Identity(t *testing.T) {
	m := Identity()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			want := float32(0)
			if i == j {
				want = 1
			}
			if m[i][j] != want {
				t.Fatalf("m[%d][%d] = %g, want %g", i, j, m[i][j], want)
			}
		}
	}
}

func TestMatrix_Scale(t *testing.T) {
	m := ScaleMatrix(2, 3, 4)
	p := m.TransformPoint(SFVec3f{1, 1, 1})
	if !approx(p.X, 2) || !approx(p.Y, 3) || !approx(p.Z, 4) {
		t.Fatalf("got %v", p)
	}
}

func TestMatrix_Translation(t *testing.T) {
	m := TranslationMatrix(5, 6, 7)
	p := m.TransformPoint(SFVec3f{1, 1, 1})
	if !approx(p.X, 6) || !approx(p.Y, 7) || !approx(p.Z, 8) {
		t.Fatalf("got %v", p)
	}
}

func TestMatrix_TranslationDirection(t *testing.T) {
	m := TranslationMatrix(5, 6, 7)
	d := m.TransformDirection(SFVec3f{1, 0, 0})
	if !approx(d.X, 1) || !approx(d.Y, 0) || !approx(d.Z, 0) {
		t.Fatalf("direction should be unaffected by translation, got %v", d)
	}
}

func TestMatrix_Rotation(t *testing.T) {
	m := RotationMatrix(NewRotation(0, 0, 1, float32(math.Pi/2)))
	p := m.TransformPoint(SFVec3f{1, 0, 0})
	if !approx(p.X, 0) || !approx(p.Y, 1) || !approx(p.Z, 0) {
		t.Fatalf("got %v", p)
	}
}

func TestMatrix_Mul_Identity(t *testing.T) {
	m := TranslationMatrix(1, 2, 3)
	r := m.Mul(Identity())
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if !approx(r[i][j], m[i][j]) {
				t.Fatalf("mismatch at [%d][%d]", i, j)
			}
		}
	}
}

func TestMatrix_Mul_ScaleTranslation(t *testing.T) {
	s := ScaleMatrix(2, 2, 2)
	tr := TranslationMatrix(1, 0, 0)
	m := tr.Mul(s)
	p := m.TransformPoint(SFVec3f{1, 0, 0})
	if !approx(p.X, 4) {
		t.Fatalf("expected 4, got %g", p.X)
	}
}

func TestMatrix_Transpose(t *testing.T) {
	m := TranslationMatrix(1, 2, 3)
	mt := m.Transpose()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if mt[i][j] != m[j][i] {
				t.Fatalf("transpose mismatch at [%d][%d]", i, j)
			}
		}
	}
}

func TestMatrix_Invert_Identity(t *testing.T) {
	inv := Identity().Invert()
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			want := float32(0)
			if i == j {
				want = 1
			}
			if !approx(inv[i][j], want) {
				t.Fatalf("inv[%d][%d] = %g, want %g", i, j, inv[i][j], want)
			}
		}
	}
}

func TestMatrix_Invert_Translation(t *testing.T) {
	m := TranslationMatrix(3, 4, 5)
	inv := m.Invert()
	p := inv.TransformPoint(SFVec3f{3, 4, 5})
	if !approx(p.X, 0) || !approx(p.Y, 0) || !approx(p.Z, 0) {
		t.Fatalf("expected origin, got %v", p)
	}
}

func TestMatrix_Invert_Roundtrip(t *testing.T) {
	m := RotationMatrix(NewRotation(1, 1, 0, 0.7))
	inv := m.Invert()
	prod := m.Mul(inv)
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			want := float32(0)
			if i == j {
				want = 1
			}
			if !approx(prod[i][j], want) {
				t.Fatalf("prod[%d][%d] = %g, want %g", i, j, prod[i][j], want)
			}
		}
	}
}

func TestMatrix_Determinant_Identity(t *testing.T) {
	d := Identity().Determinant()
	if !approx(d, 1) {
		t.Fatalf("expected 1, got %g", d)
	}
}

func TestMatrix_Determinant_Scale(t *testing.T) {
	d := ScaleMatrix(2, 3, 4).Determinant()
	if !approx(d, 24) {
		t.Fatalf("expected 24, got %g", d)
	}
}
