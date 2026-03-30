// Package split implements plane-split operations on solid B-rep models.
//
// Split takes a solid and a splitting plane and produces two new solids: one
// containing the geometry above the plane and one below. The operation uses
// vertex classification (above/on/below the plane), edge-plane intersection
// to insert new vertices along the cut, and face reconstruction to close the
// two halves.
//
// The algorithm handles degeneracies where vertices or edges lie exactly on
// the plane, using epsilon-based tolerances consistent with the rest of the
// solid-modeling library.
//
// This package is used by the boolean operations pipeline (boolop) and is
// available for direct use in modeling workflows that need planar sectioning.
package split
