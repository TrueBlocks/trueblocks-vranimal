// Package writer serializes a VRML97 scene graph back to valid .wrl text
// with proper formatting and indentation.
//
// Create a Writer with New(w), then call WriteScene(nodes) to emit the full
// scene. The writer tracks DEF names so that shared nodes are emitted once
// with DEF and referenced thereafter with USE, producing minimal output.
//
// All VRML97 field types (SFBool, SFFloat, MFVec3f, etc.) are formatted
// according to the specification. The output is suitable for feeding back
// into the parser or any standards-compliant VRML97 viewer.
package writer
