// Package traverser provides the visitor-pattern infrastructure for walking
// VRML97 scene graphs, plus concrete traverser implementations.
//
// Base supplies a matrix stack (PushMatrix, PopMatrix, TopMatrix) that tracks
// the cumulative Transform as the traversal descends through grouping nodes.
// Concrete traversers embed Base and override per-node methods.
//
// WriteTraverser serializes a scene graph back to formatted VRML97 text,
// handling DEF/USE references, indentation, and all field types. It is used
// by the vrml-fmt CLI tool.
//
// The action and picking traversers extend this framework with event
// processing and ray-cast hit testing, respectively.
package traverser
