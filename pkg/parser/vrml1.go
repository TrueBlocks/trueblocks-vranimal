package parser

import (
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
)

// v1SeparatorTag marks Groups that originated from V1.0 Separator nodes.
// Separators create state scope; plain V1.0 Groups do not.
const v1SeparatorTag = "v1sep"

func isV1GroupType(typeName string) bool {
	switch typeName {
	case "Separator", "Group", "Switch":
		return true
	}
	return false
}

// parseV1GroupBody parses the body of a V1.0 Separator, Group, or Switch.
// In V1.0, children are listed directly (not wrapped in a "children [...]" field).
// Fields like whichChild may also appear interspersed with children.
func (p *Parser) parseV1GroupBody(n node.Node, _ string) {
	for p.lex.Peek() != TokCloseBrace && p.lex.Peek() != TokEOF {
		tok := p.lex.Peek()
		switch tok {
		case TokDEF:
			if child := p.parseDEF(); child != nil {
				addV1Child(n, child)
			}
		case TokUSE:
			if child := p.parseUSE(); child != nil {
				addV1Child(n, child)
			}
		case TokIdentifier:
			p.lex.Next()
			name := p.lex.StrVal()
			if p.lex.Peek() == TokOpenBrace {
				p.lex.Next() // consume '{'
				child := p.createNode(name)
				if child == nil {
					depth := 1
					for depth > 0 {
						t := p.lex.Next()
						switch t {
						case TokOpenBrace:
							depth++
						case TokCloseBrace:
							depth--
						case TokEOF:
							return
						}
					}
				} else {
					p.parseNodeBody(child, name)
					if p.lex.Next() != TokCloseBrace {
						p.errorf("expected '}' for V1.0 node %s", name)
					}
					addV1Child(n, child)
				}
			} else {
				p.parseV1Field(n, name)
			}
		default:
			p.lex.Next()
		}
	}
}

func addV1Child(parent node.Node, child node.Node) {
	switch p := parent.(type) {
	case *node.Group:
		p.AddChild(child)
	case *node.Switch:
		p.Choice = append(p.Choice, child)
	}
}

func (p *Parser) parseV1Field(n node.Node, fieldName string) {
	if sw, ok := n.(*node.Switch); ok && fieldName == "whichChild" {
		sw.WhichChoice = p.parseInt32()
		return
	}
	p.skipFieldValue()
}

// translateV1 converts a V1.0-style scene graph (siblings sharing state) into
// V2.0 Shape/Appearance-wrapped geometry.
func translateV1(nodes []node.Node) []node.Node {
	s := &v1State{}
	result, _ := s.walk(nodes)
	return result
}

type v1State struct {
	mat    *node.Material
	coord  *node.Coordinate
	normal *node.NormalNode
}

func (s *v1State) clone() *v1State {
	return &v1State{mat: s.mat, coord: s.coord, normal: s.normal}
}

func (s *v1State) walk(nodes []node.Node) ([]node.Node, *v1State) {
	var result []node.Node
	for _, n := range nodes {
		switch v := n.(type) {
		case *node.Material:
			s.mat = v
		case *node.Coordinate:
			s.coord = v
		case *node.NormalNode:
			s.normal = v
		case *node.IndexedFaceSet:
			shape := &node.Shape{Geometry: v}
			if s.mat != nil {
				shape.Appearance = &node.Appearance{Material: s.mat}
			}
			if s.coord != nil {
				v.Coord = s.coord
			}
			if s.normal != nil {
				v.Normal = s.normal
			}
			result = append(result, shape)
		case *node.Group:
			if isSeparator(v) {
				sub := s.clone()
				v.Children, _ = sub.walk(v.Children)
			} else {
				v.Children, s = s.walk(v.Children)
			}
			result = append(result, v)
		case *node.Switch:
			for i, ch := range v.Choice {
				if g, ok := ch.(*node.Group); ok {
					sub := s.clone()
					g.Children, _ = sub.walk(g.Children)
					v.Choice[i] = g
				}
			}
			result = append(result, v)
		default:
			result = append(result, n)
		}
	}
	return result, s
}

func isSeparator(g *node.Group) bool {
	for _, tag := range g.IsMaps {
		if tag == v1SeparatorTag {
			return true
		}
	}
	return false
}
