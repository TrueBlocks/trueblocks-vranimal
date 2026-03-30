// Package algorithms provides topological and geometric algorithms for
// solid B-rep models built on the base half-edge data structure.
//
// Core capabilities:
//   - Point-in-polygon containment (CheckForContainment, BoundaryContains).
//   - Coplanarity testing and classification (CoplanarTest).
//   - Topology verification (VerifyTopology, VerifyLoop) including Euler
//     formula checks and half-edge consistency audits.
//   - Translational sweep for extruding planar faces into solid volumes.
//   - Intersection record tracking for ray-cast and boolean operations.
//
// These algorithms are used internally by the boolop (boolean operations)
// and split packages, and are available for direct use in custom solid
// modeling workflows.
package algorithms
