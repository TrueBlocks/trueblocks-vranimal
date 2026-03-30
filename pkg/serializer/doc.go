// Package serializer provides binary serialization and deserialization of
// VRML97 scene graphs using Go's encoding/gob format.
//
// Encode writes a parsed scene graph ([]node.Node) to an io.Writer as a
// compact binary stream. Decode reads it back. This is useful as a cache:
// parsing large .wrl files is slower than deserializing the gob representation,
// so tools can parse once and serialize to a .vrb sidecar file for fast
// subsequent loads.
//
// The format preserves the full node hierarchy, field values, DEF names, and
// ROUTE declarations.
package serializer
