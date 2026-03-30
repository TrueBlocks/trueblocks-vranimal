// Package euler implements the classical Euler operators for constructing and
// modifying boundary-representation (B-rep) solids.
//
// Euler operators maintain topological validity by construction: every
// operation preserves the Euler–Poincaré formula (V − E + F = 2 per shell,
// adjusted for holes). They are the only safe way to mutate the base.Solid
// data structure.
//
// Operators:
//   - Mvfs / Kvfs — make/kill a minimal solid (one vertex, one face, one
//     loop).
//   - Lmev / Lmev2 — split a loop edge to insert a new vertex.
//   - Lmef — split a loop to create a new face.
//   - Lkev / Lkef — inverse of Lmev / Lmef.
//   - Lkemr / Lmekr — kill/make edge to create/remove an inner ring (hole).
//   - Lmfkrh / Lkfmrh — promote/demote an inner loop to/from a separate
//     face.
//   - Lringmv — move a loop between faces.
//
// BuildFromIndexSet constructs a complete solid from indexed vertex/face
// data (the typical output of geometry generators). MarkCreases labels
// sharp edges for smooth/flat shading decisions.
package euler
