// Package vec provides VRML vector, color, rotation, matrix, and image types,
// ported from vraniml/src/utils/geometry/ and containers/.
package vec

import (
	"fmt"
	"math"
)

// ---------------------------------------------------------------------------
// SFVec2f — 2D vector
// ---------------------------------------------------------------------------

// SFVec2f is a 2-component float64 vector.
type SFVec2f struct {
	X, Y float64
}

// NewVec2f creates a 2D vector.
func NewVec2f(x, y float64) SFVec2f { return SFVec2f{x, y} }

// Add returns the component-wise sum of two vectors.
func (v SFVec2f) Add(o SFVec2f) SFVec2f { return SFVec2f{v.X + o.X, v.Y + o.Y} }

// Sub returns the component-wise difference.
func (v SFVec2f) Sub(o SFVec2f) SFVec2f { return SFVec2f{v.X - o.X, v.Y - o.Y} }

// Scale returns the vector scaled by s.
func (v SFVec2f) Scale(s float64) SFVec2f { return SFVec2f{v.X * s, v.Y * s} }

// Dot returns the dot product.
func (v SFVec2f) Dot(o SFVec2f) float64 { return v.X*o.X + v.Y*o.Y }

// Length returns the Euclidean length.
func (v SFVec2f) Length() float64 { return float64(math.Sqrt(float64(v.Dot(v)))) }

// Normalize returns the unit-length vector.
func (v SFVec2f) Normalize() SFVec2f {
	l := v.Length()
	if l == 0 {
		return v
	}
	return v.Scale(1 / l)
}

// Eq returns true if both components are equal.
func (v SFVec2f) Eq(o SFVec2f) bool { return v.X == o.X && v.Y == o.Y }

// Negate returns the negated vector.
func (v SFVec2f) Negate() SFVec2f { return SFVec2f{-v.X, -v.Y} }

// Index returns the i-th component (0=X, 1=Y).
func (v SFVec2f) Index(i int) float64 {
	if i == 0 {
		return v.X
	}
	return v.Y
}

// String returns a VRML-style string representation.
func (v SFVec2f) String() string { return fmt.Sprintf("%g %g", v.X, v.Y) }

// ---------------------------------------------------------------------------
// SFVec3f — 3D vector
// ---------------------------------------------------------------------------

// SFVec3f is a 3-component float64 vector.
type SFVec3f struct {
	X, Y, Z float64
}

var (
	Vec3fZero = SFVec3f{}
	XAxis     = SFVec3f{1, 0, 0}
	YAxis     = SFVec3f{0, 1, 0}
	ZAxis     = SFVec3f{0, 0, 1}
)

// NewVec3f creates a 3D vector.
func NewVec3f(x, y, z float64) SFVec3f { return SFVec3f{x, y, z} }

// Add returns the component-wise sum.
func (v SFVec3f) Add(o SFVec3f) SFVec3f {
	return SFVec3f{v.X + o.X, v.Y + o.Y, v.Z + o.Z}
}

// Sub returns the component-wise difference.
func (v SFVec3f) Sub(o SFVec3f) SFVec3f {
	return SFVec3f{v.X - o.X, v.Y - o.Y, v.Z - o.Z}
}

// Scale returns the vector scaled by s.
func (v SFVec3f) Scale(s float64) SFVec3f {
	return SFVec3f{v.X * s, v.Y * s, v.Z * s}
}

// Dot returns the dot product.
func (v SFVec3f) Dot(o SFVec3f) float64 {
	return v.X*o.X + v.Y*o.Y + v.Z*o.Z
}

// Cross returns the cross product.
func (v SFVec3f) Cross(o SFVec3f) SFVec3f {
	return SFVec3f{
		v.Y*o.Z - v.Z*o.Y,
		v.Z*o.X - v.X*o.Z,
		v.X*o.Y - v.Y*o.X,
	}
}

// Length returns the Euclidean length.
func (v SFVec3f) Length() float64 {
	return float64(math.Sqrt(float64(v.Dot(v))))
}

// Normalize returns the unit-length vector.
func (v SFVec3f) Normalize() SFVec3f {
	l := v.Length()
	if l == 0 {
		return v
	}
	return v.Scale(1 / l)
}

// Negate returns the negated vector.
func (v SFVec3f) Negate() SFVec3f { return SFVec3f{-v.X, -v.Y, -v.Z} }

// Eq returns true if all components are equal.
func (v SFVec3f) Eq(o SFVec3f) bool {
	return v.X == o.X && v.Y == o.Y && v.Z == o.Z
}

// Index returns the i-th component (0=X, 1=Y, 2=Z).
func (v SFVec3f) Index(i int) float64 {
	switch i {
	case 0:
		return v.X
	case 1:
		return v.Y
	default:
		return v.Z
	}
}

// String returns a VRML-style string representation.
func (v SFVec3f) String() string {
	return fmt.Sprintf("%g %g %g", v.X, v.Y, v.Z)
}

// Lerp linearly interpolates between two vectors.
func (v SFVec3f) Lerp(o SFVec3f, t float64) SFVec3f {
	return SFVec3f{
		v.X + t*(o.X-v.X),
		v.Y + t*(o.Y-v.Y),
		v.Z + t*(o.Z-v.Z),
	}
}

// ---------------------------------------------------------------------------
// SFVec4f — homogeneous 4D vector
// ---------------------------------------------------------------------------

// SFVec4f is a 4-component float64 vector.
type SFVec4f struct {
	X, Y, Z, W float64
}

// NewVec4f creates a 4D vector.
func NewVec4f(x, y, z, w float64) SFVec4f { return SFVec4f{x, y, z, w} }

// ---------------------------------------------------------------------------
// SFColor — RGBA color
// ---------------------------------------------------------------------------

// SFColor represents an RGBA color with float64 components in [0,1].
type SFColor struct {
	R, G, B, A float64
}

// NewColor creates an opaque color (alpha=1).
func NewColor(r, g, b float64) SFColor { return SFColor{r, g, b, 1} }

// NewColorA creates a color with explicit alpha.
func NewColorA(r, g, b, a float64) SFColor { return SFColor{r, g, b, a} }

var (
	Black   = SFColor{0, 0, 0, 1}
	White   = SFColor{1, 1, 1, 1}
	Red     = SFColor{1, 0, 0, 1}
	Green   = SFColor{0, 1, 0, 1}
	Blue    = SFColor{0, 0, 1, 1}
	Yellow  = SFColor{1, 1, 0, 1}
	Cyan    = SFColor{0, 1, 1, 1}
	Magenta = SFColor{1, 0, 1, 1}
	Grey    = SFColor{0.5, 0.5, 0.5, 1}
)

// Add returns the component-wise sum of two colors.
func (c SFColor) Add(o SFColor) SFColor {
	return SFColor{c.R + o.R, c.G + o.G, c.B + o.B, c.A + o.A}
}

// Sub returns the component-wise difference.
func (c SFColor) Sub(o SFColor) SFColor {
	return SFColor{c.R - o.R, c.G - o.G, c.B - o.B, c.A - o.A}
}

// Scale returns the color scaled by s.
func (c SFColor) Scale(s float64) SFColor {
	return SFColor{c.R * s, c.G * s, c.B * s, c.A * s}
}

// Eq returns true if all components match.
func (c SFColor) Eq(o SFColor) bool {
	return c.R == o.R && c.G == o.G && c.B == o.B && c.A == o.A
}

// NormalizeColor clamps each component to [0,1].
func (c SFColor) NormalizeColor() SFColor {
	cl := func(v float64) float64 {
		if v < 0 {
			return 0
		}
		if v > 1 {
			return 1
		}
		return v
	}
	return SFColor{cl(c.R), cl(c.G), cl(c.B), cl(c.A)}
}

// Vec3f returns the RGB components as a 3D vector.
func (c SFColor) Vec3f() SFVec3f { return SFVec3f{c.R, c.G, c.B} }

// String returns a VRML-style string representation.
func (c SFColor) String() string {
	return fmt.Sprintf("%g %g %g %g", c.R, c.G, c.B, c.A)
}

// ---------------------------------------------------------------------------
// SFRotation — axis-angle rotation
// ---------------------------------------------------------------------------

// SFRotation represents an axis-angle rotation. X,Y,Z define the axis, W is
// the angle in radians.
type SFRotation struct {
	X, Y, Z, W float64
}

// NewRotation creates a rotation from axis components and an angle.
func NewRotation(x, y, z, angle float64) SFRotation {
	return SFRotation{x, y, z, angle}
}

// Axis returns the rotation axis as a unit vector.
func (r SFRotation) Axis() SFVec3f {
	return SFVec3f{r.X, r.Y, r.Z}.Normalize()
}

// Angle returns the rotation angle in radians.
func (r SFRotation) Angle() float64 { return r.W }

// Eq returns true if both rotations are identical.
func (r SFRotation) Eq(o SFRotation) bool {
	return r.X == o.X && r.Y == o.Y && r.Z == o.Z && r.W == o.W
}

// String returns a VRML-style string representation.
func (r SFRotation) String() string {
	return fmt.Sprintf("%g %g %g %g", r.X, r.Y, r.Z, r.W)
}

// SlerpRotation interpolates between two axis-angle rotations using quaternion slerp.
func SlerpRotation(a, b SFRotation, t float64) SFRotation {
	// Convert axis-angle to quaternion
	q1 := axisAngleToQuat(a)
	q2 := axisAngleToQuat(b)

	// Slerp in quaternion space
	dot := q1[0]*q2[0] + q1[1]*q2[1] + q1[2]*q2[2] + q1[3]*q2[3]
	if dot < 0 {
		q2 = [4]float64{-q2[0], -q2[1], -q2[2], -q2[3]}
		dot = -dot
	}
	if dot > 0.9995 {
		// Linear interpolation for nearly identical quaternions
		for i := range q1 {
			q1[i] = q1[i] + float64(t)*(q2[i]-q1[i])
		}
	} else {
		theta := math.Acos(dot)
		sinTheta := math.Sin(theta)
		w1 := math.Sin(float64(1-t)*theta) / sinTheta
		w2 := math.Sin(float64(t)*theta) / sinTheta
		for i := range q1 {
			q1[i] = w1*q1[i] + w2*q2[i]
		}
	}

	// Normalize and convert back to axis-angle
	return quatToAxisAngle(q1)
}

func axisAngleToQuat(r SFRotation) [4]float64 {
	half := float64(r.W) / 2
	s := math.Sin(half)
	axis := SFVec3f{r.X, r.Y, r.Z}.Normalize()
	return [4]float64{
		float64(axis.X) * s,
		float64(axis.Y) * s,
		float64(axis.Z) * s,
		math.Cos(half),
	}
}

func quatToAxisAngle(q [4]float64) SFRotation {
	// Normalize quaternion
	n := math.Sqrt(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + q[3]*q[3])
	if n > 0 {
		q[0] /= n
		q[1] /= n
		q[2] /= n
		q[3] /= n
	}
	angle := 2 * math.Acos(math.Max(-1, math.Min(1, q[3])))
	s := math.Sqrt(1 - q[3]*q[3])
	if s < 1e-8 {
		return SFRotation{0, 1, 0, float64(angle)}
	}
	return SFRotation{
		X: float64(q[0] / s),
		Y: float64(q[1] / s),
		Z: float64(q[2] / s),
		W: float64(angle),
	}
}

// ---------------------------------------------------------------------------
// SFImage — pixel image data
// ---------------------------------------------------------------------------

// SFImage holds raw pixel data for VRML PixelTexture nodes.
type SFImage struct {
	Width, Height, NumComponents int64
	Pixels                       []uint8
}

// NewImage creates an empty image with allocated pixel buffer.
func NewImage(w, h, nc int64) SFImage {
	return SFImage{
		Width:         w,
		Height:        h,
		NumComponents: nc,
		Pixels:        make([]uint8, int(w)*int(h)*int(nc)),
	}
}

// ---------------------------------------------------------------------------
// Matrix — 4x4 transformation matrix (row-major)
// ---------------------------------------------------------------------------

// Matrix is a 4x4 float64 transformation matrix.
type Matrix [4][4]float64

// Identity returns the 4x4 identity matrix.
func Identity() Matrix {
	var m Matrix
	m[0][0] = 1
	m[1][1] = 1
	m[2][2] = 1
	m[3][3] = 1
	return m
}

// ScaleMatrix creates a non-uniform scaling matrix.
func ScaleMatrix(sx, sy, sz float64) Matrix {
	m := Identity()
	m[0][0] = sx
	m[1][1] = sy
	m[2][2] = sz
	return m
}

// TranslationMatrix creates a translation matrix.
func TranslationMatrix(tx, ty, tz float64) Matrix {
	m := Identity()
	m[3][0] = tx
	m[3][1] = ty
	m[3][2] = tz
	return m
}

// RotationMatrix creates a rotation matrix from an axis-angle rotation.
func RotationMatrix(rot SFRotation) Matrix {
	axis := rot.Axis()
	angle := float64(rot.W)
	c := float64(math.Cos(angle))
	s := float64(math.Sin(angle))
	t := 1 - c
	x, y, z := axis.X, axis.Y, axis.Z

	m := Identity()
	m[0][0] = t*x*x + c
	m[0][1] = t*x*y + s*z
	m[0][2] = t*x*z - s*y
	m[1][0] = t*x*y - s*z
	m[1][1] = t*y*y + c
	m[1][2] = t*y*z + s*x
	m[2][0] = t*x*z + s*y
	m[2][1] = t*y*z - s*x
	m[2][2] = t*z*z + c
	return m
}

// Mul multiplies two 4x4 matrices.
func (a Matrix) Mul(b Matrix) Matrix {
	var r Matrix
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				r[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return r
}

// TransformPoint multiplies a point (w=1) by the matrix.
func (m Matrix) TransformPoint(v SFVec3f) SFVec3f {
	return SFVec3f{
		X: m[0][0]*v.X + m[1][0]*v.Y + m[2][0]*v.Z + m[3][0],
		Y: m[0][1]*v.X + m[1][1]*v.Y + m[2][1]*v.Z + m[3][1],
		Z: m[0][2]*v.X + m[1][2]*v.Y + m[2][2]*v.Z + m[3][2],
	}
}

// TransformDirection multiplies a direction vector (w=0) by the matrix.
func (m Matrix) TransformDirection(v SFVec3f) SFVec3f {
	return SFVec3f{
		X: m[0][0]*v.X + m[1][0]*v.Y + m[2][0]*v.Z,
		Y: m[0][1]*v.X + m[1][1]*v.Y + m[2][1]*v.Z,
		Z: m[0][2]*v.X + m[1][2]*v.Y + m[2][2]*v.Z,
	}
}

// Transpose returns the transposed matrix.
func (m Matrix) Transpose() Matrix {
	var r Matrix
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			r[i][j] = m[j][i]
		}
	}
	return r
}

// Invert returns the inverse of the matrix using Gauss-Jordan elimination.
func (m Matrix) Invert() Matrix {
	var inv Matrix
	var aug [4][8]float64
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			aug[i][j] = float64(m[i][j])
		}
		aug[i][i+4] = 1
	}
	for col := 0; col < 4; col++ {
		pivot := col
		for row := col + 1; row < 4; row++ {
			if math.Abs(aug[row][col]) > math.Abs(aug[pivot][col]) {
				pivot = row
			}
		}
		aug[col], aug[pivot] = aug[pivot], aug[col]
		d := aug[col][col]
		if d == 0 {
			return Identity()
		}
		for j := 0; j < 8; j++ {
			aug[col][j] /= d
		}
		for row := 0; row < 4; row++ {
			if row == col {
				continue
			}
			f := aug[row][col]
			for j := 0; j < 8; j++ {
				aug[row][j] -= f * aug[col][j]
			}
		}
	}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			inv[i][j] = float64(aug[i][j+4])
		}
	}
	return inv
}

// Determinant calculates the 4x4 matrix determinant.
func (m Matrix) Determinant() float64 {
	return m[0][0]*m.cofactor(0, 0) -
		m[0][1]*m.cofactor(0, 1) +
		m[0][2]*m.cofactor(0, 2) -
		m[0][3]*m.cofactor(0, 3)
}

func (m Matrix) cofactor(row, col int) float64 {
	var sub [3][3]float64
	si := 0
	for i := 0; i < 4; i++ {
		if i == row {
			continue
		}
		sj := 0
		for j := 0; j < 4; j++ {
			if j == col {
				continue
			}
			sub[si][sj] = m[i][j]
			sj++
		}
		si++
	}
	return sub[0][0]*(sub[1][1]*sub[2][2]-sub[1][2]*sub[2][1]) -
		sub[0][1]*(sub[1][0]*sub[2][2]-sub[1][2]*sub[2][0]) +
		sub[0][2]*(sub[1][0]*sub[2][1]-sub[1][1]*sub[2][0])
}
