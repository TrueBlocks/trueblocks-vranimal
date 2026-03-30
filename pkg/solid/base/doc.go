// Package base defines the half-edge boundary representation (B-rep) data
// structure that underpins all solid modeling in the library.
//
// The five core entities are:
//   - Solid: top-level container owning linked lists of faces, edges, and
//     vertices.  Provides iteration (ForEachFace, ForEachEdge, ForEachVertex),
//     statistics (NFaces, NEdges, NVerts), extents, and geometry transforms.
//   - Face: an oriented polygonal surface bounded by one outer Loop and zero
//     or more inner Loops (holes). Carries a Plane equation and a color.
//   - Edge: a shared edge connecting two HalfEdges in adjacent faces.
//   - HalfEdge: a directed edge within a single face Loop, linking to its
//     Vertex, Next/Prev half-edges, mate (opposite direction), and parent
//     Loop.
//   - Vertex: a 3-D point (Loc) with a back-pointer to one incident
//     HalfEdge.
//   - Loop: a circular ring of HalfEdges forming an outer boundary or an
//     inner hole of a Face.
//
// This package is intentionally low-level. Higher-level construction is done
// via the euler package; algorithms, boolop, split, and export operate on
// these structures.
package base
