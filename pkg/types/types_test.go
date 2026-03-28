package types

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// Clamp
// ---------------------------------------------------------------------------

func TestClamp_Float64_Middle(t *testing.T) {
	if got := Clamp(0.5, 0.0, 1.0); got != 0.5 {
		t.Errorf("Clamp(0.5, 0, 1) = %v, want 0.5", got)
	}
}

func TestClamp_Float64_BelowLo(t *testing.T) {
	if got := Clamp(-1.0, 0.0, 1.0); got != 0.0 {
		t.Errorf("Clamp(-1, 0, 1) = %v, want 0", got)
	}
}

func TestClamp_Float64_AboveHi(t *testing.T) {
	if got := Clamp(5.0, 0.0, 1.0); got != 1.0 {
		t.Errorf("Clamp(5, 0, 1) = %v, want 1", got)
	}
}

func TestClamp_Float64_AtLo(t *testing.T) {
	if got := Clamp(0.0, 0.0, 1.0); got != 0.0 {
		t.Errorf("Clamp(0, 0, 1) = %v, want 0", got)
	}
}

func TestClamp_Float64_AtHi(t *testing.T) {
	if got := Clamp(1.0, 0.0, 1.0); got != 1.0 {
		t.Errorf("Clamp(1, 0, 1) = %v, want 1", got)
	}
}

func TestClamp_Int64_BelowLo(t *testing.T) {
	if got := Clamp(int64(-5), int64(0), int64(10)); got != 0 {
		t.Errorf("Clamp(-5, 0, 10) = %v, want 0", got)
	}
}

func TestClamp_Int64_AboveHi(t *testing.T) {
	if got := Clamp(int64(20), int64(0), int64(10)); got != 10 {
		t.Errorf("Clamp(20, 0, 10) = %v, want 10", got)
	}
}

func TestClamp_Int64_Middle(t *testing.T) {
	if got := Clamp(int64(5), int64(0), int64(10)); got != 5 {
		t.Errorf("Clamp(5, 0, 10) = %v, want 5", got)
	}
}

func TestClamp_Negative_Range(t *testing.T) {
	if got := Clamp(-3.0, -5.0, -1.0); got != -3.0 {
		t.Errorf("Clamp(-3, -5, -1) = %v, want -3", got)
	}
}

// ---------------------------------------------------------------------------
// InRange
// ---------------------------------------------------------------------------

func TestInRange_Inside(t *testing.T) {
	if !InRange(0.5, 0.0, 1.0) {
		t.Error("InRange(0.5, 0, 1) = false, want true")
	}
}

func TestInRange_AtLo(t *testing.T) {
	if !InRange(0.0, 0.0, 1.0) {
		t.Error("InRange(0, 0, 1) = false, want true (inclusive)")
	}
}

func TestInRange_AtHi(t *testing.T) {
	if !InRange(1.0, 0.0, 1.0) {
		t.Error("InRange(1, 0, 1) = false, want true (inclusive)")
	}
}

func TestInRange_BelowLo(t *testing.T) {
	if InRange(-0.1, 0.0, 1.0) {
		t.Error("InRange(-0.1, 0, 1) = true, want false")
	}
}

func TestInRange_AboveHi(t *testing.T) {
	if InRange(1.1, 0.0, 1.0) {
		t.Error("InRange(1.1, 0, 1) = true, want false")
	}
}

func TestInRange_Int64(t *testing.T) {
	if !InRange(int64(5), int64(0), int64(10)) {
		t.Error("InRange(5, 0, 10) = false, want true")
	}
	if InRange(int64(11), int64(0), int64(10)) {
		t.Error("InRange(11, 0, 10) = true, want false")
	}
}

// ---------------------------------------------------------------------------
// Interpolate
// ---------------------------------------------------------------------------

func TestInterpolate_AtZero(t *testing.T) {
	if got := Interpolate(1.0, 3.0, 0.0); got != 1.0 {
		t.Errorf("Interpolate(1, 3, 0) = %v, want 1", got)
	}
}

func TestInterpolate_AtOne(t *testing.T) {
	if got := Interpolate(1.0, 3.0, 1.0); got != 3.0 {
		t.Errorf("Interpolate(1, 3, 1) = %v, want 3", got)
	}
}

func TestInterpolate_AtHalf(t *testing.T) {
	if got := Interpolate(0.0, 10.0, 0.5); got != 5.0 {
		t.Errorf("Interpolate(0, 10, 0.5) = %v, want 5", got)
	}
}

func TestInterpolate_NegativeValues(t *testing.T) {
	if got := Interpolate(-2.0, 2.0, 0.5); got != 0.0 {
		t.Errorf("Interpolate(-2, 2, 0.5) = %v, want 0", got)
	}
}

func TestInterpolate_Extrapolation(t *testing.T) {
	if got := Interpolate(0.0, 10.0, 2.0); got != 20.0 {
		t.Errorf("Interpolate(0, 10, 2) = %v, want 20", got)
	}
}

func TestInterpolate_SameValues(t *testing.T) {
	if got := Interpolate(5.0, 5.0, 0.7); got != 5.0 {
		t.Errorf("Interpolate(5, 5, 0.7) = %v, want 5", got)
	}
}

// ---------------------------------------------------------------------------
// Deg2Rad / Rad2Deg
// ---------------------------------------------------------------------------

func TestDeg2Rad_Zero(t *testing.T) {
	if got := Deg2Rad(0); got != 0 {
		t.Errorf("Deg2Rad(0) = %v, want 0", got)
	}
}

func TestDeg2Rad_90(t *testing.T) {
	if got := Deg2Rad(90); !Equals(got, math.Pi/2, 1e-12) {
		t.Errorf("Deg2Rad(90) = %v, want Pi/2 (%v)", got, math.Pi/2)
	}
}

func TestDeg2Rad_180(t *testing.T) {
	if got := Deg2Rad(180); !Equals(got, math.Pi, 1e-12) {
		t.Errorf("Deg2Rad(180) = %v, want Pi", got)
	}
}

func TestDeg2Rad_360(t *testing.T) {
	if got := Deg2Rad(360); !Equals(got, 2*math.Pi, 1e-12) {
		t.Errorf("Deg2Rad(360) = %v, want 2*Pi", got)
	}
}

func TestDeg2Rad_Negative(t *testing.T) {
	if got := Deg2Rad(-90); !Equals(got, -math.Pi/2, 1e-12) {
		t.Errorf("Deg2Rad(-90) = %v, want -Pi/2", got)
	}
}

func TestRad2Deg_Zero(t *testing.T) {
	if got := Rad2Deg(0); got != 0 {
		t.Errorf("Rad2Deg(0) = %v, want 0", got)
	}
}

func TestRad2Deg_PiOver2(t *testing.T) {
	if got := Rad2Deg(math.Pi / 2); !Equals(got, 90, 1e-12) {
		t.Errorf("Rad2Deg(Pi/2) = %v, want 90", got)
	}
}

func TestRad2Deg_Pi(t *testing.T) {
	if got := Rad2Deg(math.Pi); !Equals(got, 180, 1e-12) {
		t.Errorf("Rad2Deg(Pi) = %v, want 180", got)
	}
}

func TestDeg2Rad_Rad2Deg_RoundTrip(t *testing.T) {
	for _, deg := range []float64{0, 30, 45, 60, 90, 120, 180, 270, 360, -45} {
		got := Rad2Deg(Deg2Rad(deg))
		if !Equals(got, deg, 1e-12) {
			t.Errorf("Rad2Deg(Deg2Rad(%v)) = %v", deg, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Pow2LT
// ---------------------------------------------------------------------------

func TestPow2LT_PowersOfTwo(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{2, 1},
		{4, 2},
		{8, 4},
		{16, 8},
		{256, 128},
		{1024, 512},
	}
	for _, tc := range tests {
		if got := Pow2LT(tc.n); got != tc.want {
			t.Errorf("Pow2LT(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestPow2LT_NonPowersOfTwo(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{3, 2},
		{5, 4},
		{7, 4},
		{9, 8},
		{100, 64},
		{255, 128},
		{257, 256},
	}
	for _, tc := range tests {
		if got := Pow2LT(tc.n); got != tc.want {
			t.Errorf("Pow2LT(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestPow2LT_One(t *testing.T) {
	if got := Pow2LT(1); got != 0 {
		t.Errorf("Pow2LT(1) = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// Equals
// ---------------------------------------------------------------------------

func TestEquals_Identical(t *testing.T) {
	if !Equals(1.0, 1.0, 1e-9) {
		t.Error("Equals(1, 1, 1e-9) = false, want true")
	}
}

func TestEquals_WithinEpsilon(t *testing.T) {
	if !Equals(1.0, 1.0+5e-10, 1e-9) {
		t.Error("Equals(1.0, 1.0+5e-10, 1e-9) = false, want true")
	}
}

func TestEquals_OutsideEpsilon(t *testing.T) {
	if Equals(1.0, 1.0+2e-9, 1e-9) {
		t.Error("Equals(1.0, 1.0+2e-9, 1e-9) = true, want false")
	}
}

func TestEquals_NegativeDifference(t *testing.T) {
	if !Equals(1.0, 0.9999999999, 1e-9) {
		t.Error("Equals with negative difference should work (abs)")
	}
}

func TestEquals_Zero(t *testing.T) {
	if !Equals(0.0, 0.0, 1e-9) {
		t.Error("Equals(0, 0, 1e-9) = false, want true")
	}
}

func TestEquals_LargeEpsilon(t *testing.T) {
	if !Equals(1.0, 2.0, 1.5) {
		t.Error("Equals(1, 2, 1.5) = false, want true")
	}
}

func TestEquals_ExactlyAtEpsilon(t *testing.T) {
	// Equals uses strict < (not <=), so |1.0 - 2.0| = 1.0, not < 1.0
	if Equals(1.0, 2.0, 1.0) {
		t.Error("Equals(1, 2, 1.0) = true, want false (strict <)")
	}
}

// ---------------------------------------------------------------------------
// Pi constant
// ---------------------------------------------------------------------------

func TestPi(t *testing.T) {
	if Pi != math.Pi {
		t.Errorf("Pi = %v, want %v", Pi, math.Pi)
	}
}
