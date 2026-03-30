// Package geo provides geometric primitives used throughout the solid-modeling
// and rendering pipelines: rays, planes, axis-aligned bounding boxes, and 2-D
// rectangles.
//
// Key capabilities:
//   - Ray: origin + direction, with Evaluate, reflection, and matrix transform.
//   - Plane: normal + signed distance, constructed from three points or
//     directly. Supports ray and plane–plane intersections.
//   - BoundingBox: AABB with union, overlap test, and matrix transform.
//   - Rect2D: simple 2-D rectangle for viewport and picking math.
//
// All types use float64 precision and operate in the same coordinate system
// as the vec package.
package geo
