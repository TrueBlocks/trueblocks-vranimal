// Package converter bridges VRML97 scene graphs to the g3n 3D engine for
// OpenGL rendering. It maps VRML node types (Shape, Transform, IndexedFaceSet,
// lights, etc.) to their g3n equivalents and maintains a bidirectional NodeMap
// so that runtime updates to the VRML graph (e.g., via ROUTEs or sensors) can
// be reflected in the rendered scene.
//
// Typical usage: parse a .wrl file with the parser package, then call the
// converter from the viewer to produce a g3n scene graph suitable for display
// in a g3n application window.
package converter
