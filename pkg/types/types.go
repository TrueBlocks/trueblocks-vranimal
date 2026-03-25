// Package types provides the basic VRML type aliases and utility functions,
// ported from vraniml/src/utils/types.h and utility.h.
package types

import "math"

// Basic VRML scalar types.
type (
	SFBool  = bool
	SFInt32 = int32
	SFFloat = float32
	SFTime  = float64
)

const Pi = math.Pi

// Clamp returns v clamped to the range [lo, hi].
func Clamp[T ~int32 | ~float32 | ~float64](v, lo, hi T) T {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// InRange returns true if v is in [lo, hi].
func InRange[T ~int32 | ~float32 | ~float64](v, lo, hi T) bool {
	return v >= lo && v <= hi
}

// Interpolate returns a + t*(b-a).
func Interpolate[T ~float32 | ~float64](a, b, t T) T {
	return a + t*(b-a)
}

// Deg2Rad converts degrees to radians.
func Deg2Rad(deg SFFloat) SFFloat {
	return deg * SFFloat(Pi) / 180.0
}

// Rad2Deg converts radians to degrees.
func Rad2Deg(rad SFFloat) SFFloat {
	return rad * 180.0 / SFFloat(Pi)
}

// Pow2LT returns the largest power-of-2 less than n.
func Pow2LT(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}
	return p >> 1
}

// Equals returns true if a and b are approximately equal within epsilon.
func Equals(a, b, epsilon SFFloat) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < epsilon
}
