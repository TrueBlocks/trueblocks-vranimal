package solid

import (
	"math"
	"testing"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------- helpers ----------

// edgeLengths returns min and max edge lengths in the solid.
func edgeLengths(s *Solid) (min, max float64) {
	min = math.MaxFloat64
	s.ForEachEdge(func(e *Edge) bool {
		l := e.Length()
		if l < min {
			min = l
		}
		if l > max {
			max = l
		}
		return true
	})
	return
}

// vertexSet collects all vertex locations.
func vertexSet(s *Solid) []vec.SFVec3f {
	var locs []vec.SFVec3f
	s.ForEachVertex(func(v *Vertex) bool {
		locs = append(locs, v.Loc)
		return true
	})
	return locs
}

// distinctCount returns how many unique points exist (within tolerance).
func distinctCount(pts []vec.SFVec3f, tol float64) int {
	unique := make([]vec.SFVec3f, 0, len(pts))
	for _, p := range pts {
		found := false
		for _, u := range unique {
			if p.Sub(u).Length() < tol {
				found = true
				break
			}
		}
		if !found {
			unique = append(unique, p)
		}
	}
	return len(unique)
}

// ---------- Test 1: Large Coordinate Precision ----------

func TestPrecision_LargeCoords(t *testing.T) {
	h := 5000.0 // halfSize
	s := MakeCube(h, vec.Red)
	if s == nil {
		t.Fatal("MakeCube returned nil")
	}

	// Verify basic topology
	errs := s.VerifyDetailed()
	for _, err := range errs {
		t.Errorf("verify: %v", err)
	}

	// Volume = (2h)^3 = 1e12
	expected := math.Pow(2*h, 3)
	vol := s.Volume()
	if math.Abs(vol-expected)/expected > 1e-12 {
		t.Errorf("volume = %.15g, want %.15g", vol, expected)
	}

	// Every vertex coordinate should be exactly ±5000
	s.ForEachVertex(func(v *Vertex) bool {
		for _, c := range []float64{v.Loc.X, v.Loc.Y, v.Loc.Z} {
			if math.Abs(c) != h {
				t.Errorf("vertex coord %.15g is not ±%.0f", c, h)
			}
		}
		return true
	})

	// 8 distinct vertices
	f, e, v := s.Stats()
	if v != 8 {
		t.Errorf("verts = %d, want 8", v)
	}
	if f != 6 {
		t.Errorf("faces = %d, want 6", f)
	}
	if e != 12 {
		t.Errorf("edges = %d, want 12", e)
	}
}

// ---------- Test 2: Small Coordinate Precision ----------

func TestPrecision_SmallCoords(t *testing.T) {
	h := 5e-5 // halfSize → side = 1e-4
	s := MakeCube(h, vec.Green)
	if s == nil {
		t.Fatal("MakeCube returned nil")
	}

	errs := s.VerifyDetailed()
	for _, err := range errs {
		t.Errorf("verify: %v", err)
	}

	// 8 distinct vertices — float32 could merge coincident points
	pts := vertexSet(s)
	if n := distinctCount(pts, 1e-15); n != 8 {
		t.Errorf("distinct vertices = %d, want 8", n)
	}

	// All edges positive length
	minL, _ := edgeLengths(s)
	if minL <= 0 {
		t.Errorf("min edge length = %g, want > 0", minL)
	}

	// Volume = (1e-4)^3 = 1e-12
	expected := math.Pow(2*h, 3)
	vol := s.Volume()
	if math.Abs(vol-expected) > 1e-24 {
		t.Errorf("volume = %.15g, want %.15g", vol, expected)
	}
}

// ---------- Test 3: Far-From-Origin Precision ----------

func TestPrecision_FarFromOrigin(t *testing.T) {
	s := MakeCube(0.5, vec.Blue) // side = 1.0
	if s == nil {
		t.Fatal("MakeCube returned nil")
	}

	// Translate to (1e6, 1e6, 1e6)
	m := vec.TranslationMatrix(1e6, 1e6, 1e6)
	s.TransformGeometry(m)

	errs := s.VerifyDetailed()
	for _, err := range errs {
		t.Errorf("verify: %v", err)
	}

	// Edge lengths should still be exactly 1.0
	minL, maxL := edgeLengths(s)
	if math.Abs(minL-1.0) > 1e-9 || math.Abs(maxL-1.0) > 1e-9 {
		t.Errorf("edge lengths: min=%.15g max=%.15g, want 1.0", minL, maxL)
	}

	// Volume should still be 1.0
	vol := s.Volume()
	if math.Abs(vol-1.0) > 1e-6 {
		t.Errorf("volume = %.15g, want 1.0", vol)
	}

	// Face normals should still be axis-aligned after translation
	s.CalcPlaneEquations()
	s.ForEachFace(func(f *Face) bool {
		n := f.Normal
		// One component should be ±1, others 0
		absX, absY, absZ := math.Abs(n.X), math.Abs(n.Y), math.Abs(n.Z)
		maxN := math.Max(absX, math.Max(absY, absZ))
		if math.Abs(maxN-1.0) > 1e-10 {
			t.Errorf("face %d normal (%.6g, %.6g, %.6g) not axis-aligned", f.Index, n.X, n.Y, n.Z)
		}
		return true
	})
}

// ---------- Test 4: Epsilon Separation ----------

func TestPrecision_EpsilonSeparation(t *testing.T) {
	s1 := MakeCube(0.5, vec.Red)
	s2 := MakeCube(0.5, vec.Green)
	if s1 == nil || s2 == nil {
		t.Fatal("MakeCube returned nil")
	}

	// Translate s2 by Z = 1.0 + 1e-10 — just barely separated
	m := vec.TranslationMatrix(0, 0, 1.0+1e-10)
	s2.TransformGeometry(m)

	// Merge into one solid (two disjoint shells)
	s1.Merge(s2)

	// Must have 16 distinct vertices (not 12 if faces were merged)
	pts := vertexSet(s1)
	if n := distinctCount(pts, 1e-12); n != 16 {
		t.Errorf("distinct vertices = %d, want 16 (cubes should remain separate)", n)
	}

	// Topology: 12 faces, 24 edges, 16 verts
	f, e, v := s1.Stats()
	if v != 16 {
		t.Errorf("verts = %d, want 16", v)
	}
	if f != 12 {
		t.Errorf("faces = %d, want 12", f)
	}
	if e != 24 {
		t.Errorf("edges = %d, want 24", e)
	}
}

// ---------- Test 5: Near-Coplanar Detection ----------

func TestPrecision_NearCoplanar(t *testing.T) {
	s1 := MakeCube(0.5, vec.Red)
	s2 := MakeCube(0.5, vec.Green)
	if s1 == nil || s2 == nil {
		t.Fatal("MakeCube returned nil")
	}

	// Rotate s2 by 0.001° around Z. The +X and +Y faces get their normals
	// rotated, while the +Z face normal stays (0,0,1). We test the +X face.
	angle := 0.001 * math.Pi / 180.0
	rot := vec.RotationMatrix(vec.SFRotation{X: 0, Y: 0, Z: 1, W: angle})
	s2.TransformGeometry(rot)
	s2.CalcPlaneEquations()

	// Find the +X face on each solid (normal.X > 0.99)
	var s1Face, s2Face *Face
	s1.ForEachFace(func(f *Face) bool {
		if f.Normal.X > 0.99 {
			s1Face = f
			return false
		}
		return true
	})
	s2.ForEachFace(func(f *Face) bool {
		if f.Normal.X > 0.99 {
			s2Face = f
			return false
		}
		return true
	})

	if s1Face == nil || s2Face == nil {
		t.Fatal("could not find +X faces")
	}

	// The normals differ by sin(0.001°) ≈ 1.75e-5, which is above the coplanar
	// threshold of 1e-7. IsCoplanar should return false.
	if s1Face.IsCoplanar(s2Face) {
		cross := s1Face.Normal.Cross(s2Face.Normal)
		t.Errorf("IsCoplanar returned true for 0.001°-rotated +X faces; cross magnitude = %g", cross.Length())
	}

	// Also verify that truly coplanar faces are still detected
	s3 := MakeCube(0.5, vec.Blue)
	var s3Face *Face
	s3.ForEachFace(func(f *Face) bool {
		if f.Normal.X > 0.99 {
			s3Face = f
			return false
		}
		return true
	})
	if s3Face != nil && !s1Face.IsCoplanar(s3Face) {
		t.Error("IsCoplanar returned false for identical faces; should be true")
	}
}

// ---------- Test 6: Accumulated Rotation Precision ----------

func TestPrecision_AccumulatedRotations(t *testing.T) {
	s := MakeCube(0.5, vec.Yellow)
	if s == nil {
		t.Fatal("MakeCube returned nil")
	}

	// Save original vertex positions
	original := make([]vec.SFVec3f, 0, 8)
	s.ForEachVertex(func(v *Vertex) bool {
		original = append(original, v.Loc)
		return true
	})

	// Apply 360 × 1° rotations around Y axis
	oneRot := vec.RotationMatrix(vec.SFRotation{X: 0, Y: 1, Z: 0, W: math.Pi / 180.0})
	for i := 0; i < 360; i++ {
		s.TransformGeometry(oneRot)
	}

	// Verify vertices match original within tolerance
	// With float64, accumulated error for 360 rotations should be < 1e-10
	i := 0
	s.ForEachVertex(func(v *Vertex) bool {
		if i >= len(original) {
			return false
		}
		diff := v.Loc.Sub(original[i]).Length()
		if diff > 1e-10 {
			t.Errorf("vertex %d drifted by %g after 360×1° rotations", i, diff)
		}
		i++
		return true
	})
}

// ---------- Test 7: Cross-Product Precision ----------

func TestPrecision_CrossProduct(t *testing.T) {
	// Nearly-parallel vectors — cross product magnitude is critical for coplanar tests
	a := vec.SFVec3f{X: 1.0, Y: 0.0, Z: 0.0}
	b := vec.SFVec3f{X: 1.0, Y: 1e-12, Z: 0.0}
	cross := a.Cross(b)

	// Z component should be exactly 1e-12
	// float32 would lose this — the smallest distinguishable float32 near 1.0 is ~6e-8
	if cross.Z == 0 {
		t.Fatal("cross product Z component is exactly zero — precision lost")
	}
	if math.Abs(cross.Z-1e-12) > 1e-24 {
		t.Errorf("cross.Z = %g, want 1e-12", cross.Z)
	}

	// X and Y should be exactly zero
	if cross.X != 0 || cross.Y != 0 {
		t.Errorf("cross = (%g, %g, %g), want (0, 0, 1e-12)", cross.X, cross.Y, cross.Z)
	}

	// Even more extreme: 1e-15 separation — only detectable with float64
	c := vec.SFVec3f{X: 1.0, Y: 1e-15, Z: 0.0}
	cross2 := a.Cross(c)
	if cross2.Z == 0 {
		t.Error("cross product lost 1e-15 separation — float64 should handle this")
	}
}

// ---------- Test 8: Volume Computation Precision ----------

func TestPrecision_VolumeConservation(t *testing.T) {
	// Build cubes of various sizes and verify Volume() returns exact expected values.
	// This exercises the divergence theorem computation at different scales.
	tests := []struct {
		name     string
		halfSize float64
		wantVol  float64
	}{
		{"unit", 0.5, 1.0},
		{"large", 500.0, 1e9},
		{"small", 5e-4, 1e-9},
		{"huge", 5000.0, 1e12},
		{"tiny", 5e-7, 1e-18},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MakeCube(tt.halfSize, vec.White)
			if s == nil {
				t.Fatal("MakeCube returned nil")
			}
			vol := s.Volume()
			relErr := math.Abs(vol-tt.wantVol) / tt.wantVol
			if relErr > 1e-12 {
				t.Errorf("Volume() = %.15g, want %.15g (relErr = %g)", vol, tt.wantVol, relErr)
			}
		})
	}

	// Also verify that translating doesn't change volume
	t.Run("translated", func(t *testing.T) {
		s := MakeCube(0.5, vec.White)
		s.TransformGeometry(vec.TranslationMatrix(1e5, -1e5, 1e5))
		vol := s.Volume()
		if math.Abs(vol-1.0) > 1e-6 {
			t.Errorf("Volume after translation = %.15g, want 1.0", vol)
		}
	})

	// Verify sphere volume is positive and scales with radius³
	t.Run("sphere_scaling", func(t *testing.T) {
		s1 := MakeSphere(1.0, 16, 16, vec.White)
		s2 := MakeSphere(2.0, 16, 16, vec.White)
		if s1 == nil || s2 == nil {
			t.Fatal("MakeSphere returned nil")
		}
		v1 := math.Abs(s1.Volume())
		v2 := math.Abs(s2.Volume())
		if v1 <= 0 {
			t.Fatal("sphere volume is zero or negative")
		}
		// Volume scales as r³: v2/v1 should be 8.0
		ratio := v2 / v1
		if math.Abs(ratio-8.0) > 0.01 {
			t.Errorf("volume ratio = %g, want 8.0 (r³ scaling)", ratio)
		}
	})
}
