// Package types defines the fundamental VRML97 scalar type aliases and
// small numeric utilities used across the library.
//
// Type aliases:
//   - SFBool  — bool
//   - SFInt32 — int64
//   - SFFloat — float64
//   - SFTime  — float64
//
// Utility functions provide clamping, range checks, linear interpolation,
// degree/radian conversion, and power-of-two helpers. All generic functions
// use Go 1.18+ type constraints to work on both integer and float operands.
package types
