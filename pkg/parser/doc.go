// Package parser implements a full VRML97 lexer and parser that reads .wrl
// text and produces a typed scene graph of node.Node values.
//
// The lexer (NewLexer) tokenizes the input into VRML97 tokens: braces,
// brackets, keywords (DEF, USE, ROUTE, PROTO, EXTERNPROTO, NULL, TRUE,
// FALSE), identifiers, and literal values (integers, floats, strings).
//
// The parser (NewParser) consumes tokens to build group hierarchies,
// geometry, appearance, sensors, interpolators, lights, and all other
// VRML97 node types. It resolves DEF/USE references, expands PROTO and
// EXTERNPROTO definitions, and collects ROUTE declarations.
//
// Usage:
//
//	f, _ := os.Open("scene.wrl")
//	p := parser.NewParser(f)
//	p.SetBaseDir(filepath.Dir("scene.wrl"))
//	nodes := p.Parse()
//	if errs := p.Errors(); len(errs) > 0 {
//	    // handle parse errors
//	}
//	routes := p.GetRoutes()
package parser
