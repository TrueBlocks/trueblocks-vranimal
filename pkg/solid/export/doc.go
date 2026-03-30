// Package export converts solid B-rep models to VRML97 text for visualization
// and interchange.
//
// Functions:
//   - VRML / VRMLFile — write a single solid as a complete .wrl file.
//   - MultiVRML / MultiVRMLFile — write multiple solids with translation
//     offsets, useful for side-by-side comparison of boolean operation inputs
//     and results.
//   - Shape — emit one solid as a VRML Shape (IndexedFaceSet with material)
//     for embedding inside a larger scene.
//   - Wireframe — emit a solid's edges as an IndexedLineSet.
//
// Output is valid VRML97 suitable for any compliant viewer or for re-parsing
// with the parser package.
package export
