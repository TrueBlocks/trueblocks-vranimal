// Package validator checks a parsed VRML97 scene graph for structural
// problems, invalid field values, and specification conformance issues.
//
// Create a Validator with New(), then call Validate(nodes) to obtain a
// slice of Finding values. Each Finding carries a Severity (Warning or
// Error), the offending node, and a human-readable message.
//
// Checks include: missing required children, out-of-range field values,
// empty geometry, dangling DEF/USE references, and invalid ROUTE
// endpoints.
//
// The vrml-validate CLI tool wraps this package for command-line use.
package validator
