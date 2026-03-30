// Package primitives generates standard solid shapes as B-rep models using
// Euler operators from the euler package.
//
// Available primitives:
//   - MakeCube — axis-aligned cube centered at the origin.
//   - MakeSphere — UV-sphere with configurable latitude/longitude segments.
//   - MakeCylinder — capped cylinder with configurable radial segments.
//   - MakeTorus — torus with configurable major/minor radii and segments.
//   - MakePrism — triangular prism.
//   - MakeLamina — arbitrary planar polygon (lamina) from a vertex list.
//   - MakeCircle — regular polygon approximating a circle.
//
// Each function returns a *base.Solid ready for use in boolean operations,
// splitting, export, or direct rendering. All primitives are assigned a
// uniform face color passed as a vec.SFColor argument.
package primitives
