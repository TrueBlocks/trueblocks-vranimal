// Package browser implements the VRML97 browser runtime: scene graph hosting,
// event routing, time-based animation, and binding-stack management for
// bindable nodes (Viewpoint, Background, Fog, NavigationInfo).
//
// A Browser holds the parsed scene graph and an ordered list of Traversers
// that walk the graph each frame. Call NewBrowser to create an instance, add
// one or more traversers with AddTraverser, and then drive the frame loop
// externally (e.g., from the viewer's render callback).
//
// Binding stacks follow the VRML97 specification: the first bindable node
// encountered during parsing is bound automatically, and subsequent
// BindViewpoint / BindNavigationInfo calls push or pop the stack.
//
// This package is the top-level entry point for headless and headed VRML
// applications. For rendering, pair it with the converter package to bridge
// the scene graph to g3n.
package browser
