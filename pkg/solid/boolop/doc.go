// Package boolop implements boolean set operations — Union, Intersection, and
// Difference — on closed solid B-rep models.
//
// The algorithm follows the classical generate-classify-connect-finish
// pipeline:
//
//  1. Generate: compute edge-face intersection points and insert new vertices
//     and edges into both solids along the intersection curve.
//  2. Classify: label each face region of both solids as inside, outside, or
//     on the boundary of the other solid, using vertex neighborhoods.
//  3. Connect: join null-edge pairs across the two solids to merge them into
//     a single topological shell.
//  4. Finish: select and assemble the faces that belong to the result based
//     on the requested operation, then separate the result solid.
//
// Robustness features include epsilon-based distance tolerances, automatic
// perturbation when degenerate configurations are detected, and a last-ditch
// fallback classifier.
//
// Entry points:
//   - Union(a, b) / Intersection(a, b) / Difference(a, b) — simple API.
//   - BoolOp(a, b, op) — generic with operation constant.
//   - BoolOpEx(a, b, op) — returns a BoolOpResult with diagnostics.
//   - ChainBoolOpEx(solids, op) — chain multiple operands left-to-right.
package boolop
