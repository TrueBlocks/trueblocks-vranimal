// Package vec implements the VRML97 vector, color, rotation, and matrix
// types with full arithmetic operations. All values use float64 precision.
//
// Types:
//   - SFVec2f, SFVec3f, SFVec4f — 2-D, 3-D, and 4-D vectors with Add, Sub,
//     Scale, Dot, Cross, Length, Normalize, Lerp, and Clamp.
//   - SFColor — RGBA color with named constants (Black, White, Red, …) and
//     normalization.
//   - SFRotation — axis-angle rotation with SLERP interpolation and
//     composition.
//   - SFImage — pixel data for PixelTexture nodes.
//   - Matrix — 4×4 transformation matrix with translation, rotation, scale
//     constructors, multiplication, inversion, and point/vector transform.
//
// This package has no external dependencies and is the lowest-level building
// block used by every other package in the library.
package vec
