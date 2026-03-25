package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
)

// FieldKind distinguishes the four VRML97 interface declaration types.
type FieldKind int

const (
	KindField FieldKind = iota
	KindExposedField
	KindEventIn
	KindEventOut
)

// ProtoFieldDecl represents one interface field of a PROTO.
type ProtoFieldDecl struct {
	Kind     FieldKind
	TypeName string // e.g. "SFVec3f", "MFNode"
	Name     string
	Default  string // raw VRML text of default value (empty for eventIn/eventOut)
}

// ProtoDefinition stores a parsed PROTO declaration.
type ProtoDefinition struct {
	Name   string
	Fields []ProtoFieldDecl
	Body   string   // raw VRML text between the body braces
	URLs   []string // EXTERNPROTO URL list (empty for regular PROTOs)
}

// ---------------------------------------------------------------------------
// Lexer: CaptureBlock reads raw text between balanced delimiters.
// The opening delimiter must already have been consumed by the caller.
// ---------------------------------------------------------------------------

func (l *Lexer) CaptureBlock(open, close byte) string {
	var buf strings.Builder
	depth := 1
	for depth > 0 {
		ch, err := l.readByte()
		if err != nil {
			break
		}
		if ch == '\n' {
			l.line++
		}
		// Handle quoted strings verbatim
		if ch == '"' {
			buf.WriteByte(ch)
			for {
				sc, err := l.readByte()
				if err != nil {
					return buf.String()
				}
				buf.WriteByte(sc)
				if sc == '\n' {
					l.line++
				}
				if sc == '\\' {
					ec, err := l.readByte()
					if err != nil {
						return buf.String()
					}
					buf.WriteByte(ec)
					continue
				}
				if sc == '"' {
					break
				}
			}
			continue
		}
		// Handle comments verbatim
		if ch == '#' {
			buf.WriteByte(ch)
			for {
				cc, err := l.readByte()
				if err != nil {
					return buf.String()
				}
				buf.WriteByte(cc)
				if cc == '\n' {
					l.line++
					break
				}
			}
			continue
		}
		if ch == open {
			depth++
		}
		if ch == close {
			depth--
			if depth == 0 {
				break
			}
		}
		buf.WriteByte(ch)
	}
	return buf.String()
}

// ---------------------------------------------------------------------------
// parsePROTOFull: parse a PROTO declaration and store it.
// ---------------------------------------------------------------------------

func (p *Parser) parsePROTOFull() {
	p.lex.Next() // consume PROTO

	if p.lex.Peek() != TokIdentifier {
		p.errorf("expected name after PROTO")
		return
	}
	p.lex.Next()
	name := p.lex.StrVal()

	// Parse interface declarations [...]
	if p.lex.Next() != TokOpenBracket {
		p.errorf("expected '[' in PROTO %s", name)
		return
	}
	fields := p.parseProtoInterface()

	// Capture body {...} as raw text
	if p.lex.Next() != TokOpenBrace {
		p.errorf("expected '{' in PROTO %s", name)
		return
	}
	body := p.lex.CaptureBlock('{', '}')

	def := &ProtoDefinition{
		Name:   name,
		Fields: fields,
		Body:   body,
	}
	p.protoTable[name] = def
}

// parseProtoInterface parses the interface declarations between [ and ].
func (p *Parser) parseProtoInterface() []ProtoFieldDecl {
	var fields []ProtoFieldDecl
	for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
		tok := p.lex.Peek()
		var kind FieldKind
		switch tok {
		case TokField:
			kind = KindField
		case TokExposedField:
			kind = KindExposedField
		case TokEventIn:
			kind = KindEventIn
		case TokEventOut:
			kind = KindEventOut
		default:
			p.lex.Next() // skip unexpected token
			continue
		}
		p.lex.Next() // consume field/exposedField/eventIn/eventOut

		if p.lex.Next() != TokIdentifier {
			p.errorf("expected type name in interface declaration")
			continue
		}
		typeName := p.lex.StrVal()

		if p.lex.Next() != TokIdentifier {
			p.errorf("expected field name in interface declaration")
			continue
		}
		fieldName := p.lex.StrVal()

		var defaultText string
		if kind == KindField || kind == KindExposedField {
			defaultText = p.captureValueText(typeName)
		}

		fields = append(fields, ProtoFieldDecl{
			Kind:     kind,
			TypeName: typeName,
			Name:     fieldName,
			Default:  defaultText,
		})
	}
	if p.lex.Peek() == TokCloseBracket {
		p.lex.Next()
	}
	return fields
}

// captureValueText reads a field value as raw text given the VRML type name.
func (p *Parser) captureValueText(typeName string) string {
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		inner := p.lex.CaptureBlock('[', ']')
		return "[ " + inner + " ]"
	}
	if p.lex.Peek() == TokNULL {
		p.lex.Next()
		return "NULL"
	}

	count := sfTokenCount(typeName)
	if count > 0 {
		var parts []string
		for i := 0; i < count; i++ {
			parts = append(parts, p.captureOneToken())
		}
		return strings.Join(parts, " ")
	}

	// SFNode or MFNode without brackets: capture a full node value
	if typeName == "SFNode" || typeName == "MFNode" {
		return p.captureNodeValue()
	}

	// Other MF types without brackets: single scalar value
	return p.captureOneToken()
}

// captureNodeValue captures SFNode/MFNode without brackets: DEF/USE/NodeType { ... }.
func (p *Parser) captureNodeValue() string {
	if p.lex.Peek() == TokNULL {
		p.lex.Next()
		return "NULL"
	}
	if p.lex.Peek() == TokDEF {
		p.lex.Next() // consume DEF
		if p.lex.Next() != TokIdentifier {
			return "DEF"
		}
		name := p.lex.StrVal()
		rest := p.captureNodeValue()
		return "DEF " + name + " " + rest
	}
	if p.lex.Peek() == TokUSE {
		p.lex.Next() // consume USE
		if p.lex.Next() != TokIdentifier {
			return "USE"
		}
		return "USE " + p.lex.StrVal()
	}
	if p.lex.Peek() == TokIdentifier {
		p.lex.Next()
		typeName := p.lex.StrVal()
		if p.lex.Peek() == TokOpenBrace {
			p.lex.Next()
			inner := p.lex.CaptureBlock('{', '}')
			return typeName + " { " + inner + " }"
		}
		return typeName
	}
	return p.captureOneToken()
}

// captureOneToken reads one token and returns it as raw text.
func (p *Parser) captureOneToken() string {
	tok := p.lex.Next()
	switch tok {
	case TokFloat:
		return strconv.FormatFloat(p.lex.FloatVal(), 'f', -1, 64)
	case TokInt:
		return strconv.FormatInt(p.lex.IntVal(), 10)
	case TokString:
		return fmt.Sprintf("%q", p.lex.StrVal())
	case TokTRUE:
		return "TRUE"
	case TokFALSE:
		return "FALSE"
	case TokIdentifier:
		return p.lex.StrVal()
	case TokNULL:
		return "NULL"
	case TokOpenBrace:
		inner := p.lex.CaptureBlock('{', '}')
		return "{ " + inner + " }"
	default:
		return ""
	}
}

// sfTokenCount returns the number of tokens for an SF type.
func sfTokenCount(typeName string) int {
	switch typeName {
	case "SFBool", "SFInt32", "SFFloat", "SFString", "SFTime":
		return 1
	case "SFVec2f":
		return 2
	case "SFVec3f", "SFColor":
		return 3
	case "SFRotation":
		return 4
	default:
		return 0
	}
}

// ---------------------------------------------------------------------------
// PROTO instantiation
// ---------------------------------------------------------------------------

// instantiateProto creates nodes from a PROTO instance.
// Called after the opening '{' of the instance has been consumed.
func (p *Parser) instantiateProto(def *ProtoDefinition) []node.Node {
	// Resolve EXTERNPROTO if needed
	if def.Body == "" && len(def.URLs) > 0 {
		if !p.resolveExternProto(def) {
			p.errorf("could not resolve EXTERNPROTO %s", def.Name)
			for p.lex.Peek() != TokCloseBrace && p.lex.Peek() != TokEOF {
				p.lex.Next()
			}
			if p.lex.Peek() == TokCloseBrace {
				p.lex.Next()
			}
			return nil
		}
	}

	// Build default values map
	values := make(map[string]string)
	for _, f := range def.Fields {
		if f.Default != "" {
			values[f.Name] = f.Default
		}
	}

	// Parse instance field overrides
	for p.lex.Peek() != TokCloseBrace && p.lex.Peek() != TokEOF {
		if p.lex.Peek() != TokIdentifier {
			p.lex.Next()
			continue
		}
		p.lex.Next()
		fieldName := p.lex.StrVal()

		fd := findFieldDecl(def, fieldName)
		if fd == nil {
			p.errorf("unknown field %q in PROTO %s instance", fieldName, def.Name)
			p.skipFieldValue()
			continue
		}
		values[fieldName] = p.captureValueText(fd.TypeName)
	}
	if p.lex.Peek() == TokCloseBrace {
		p.lex.Next()
	}

	// Substitute IS references in body text
	body := substituteIS(def.Body, values)

	// Parse substituted body with a sub-parser
	sub := NewParser(strings.NewReader(body))
	sub.defTable = p.defTable
	sub.protoTable = p.protoTable
	sub.baseDir = p.baseDir
	sub.urlFetcher = p.urlFetcher
	nodes := sub.Parse()

	// Propagate errors and DEF names
	for _, e := range sub.errors {
		p.errors = append(p.errors, e)
	}
	for k, v := range sub.defTable {
		p.defTable[k] = v
	}
	return nodes
}

func findFieldDecl(def *ProtoDefinition, name string) *ProtoFieldDecl {
	for i := range def.Fields {
		if def.Fields[i].Name == name {
			return &def.Fields[i]
		}
	}
	return nil
}

// substituteIS replaces "IS <fieldName>" patterns in body text with values.
func substituteIS(body string, values map[string]string) string {
	var result strings.Builder
	i := 0
	n := len(body)
	for i < n {
		ch := body[i]

		// Skip quoted strings
		if ch == '"' {
			result.WriteByte(ch)
			i++
			for i < n {
				sc := body[i]
				result.WriteByte(sc)
				i++
				if sc == '\\' && i < n {
					result.WriteByte(body[i])
					i++
					continue
				}
				if sc == '"' {
					break
				}
			}
			continue
		}

		// Skip comments
		if ch == '#' {
			for i < n && body[i] != '\n' {
				result.WriteByte(body[i])
				i++
			}
			continue
		}

		// Check for 'IS' keyword as standalone token
		if ch == 'I' && i+1 < n && body[i+1] == 'S' {
			before := i == 0 || isSpace(body[i-1])
			after := i+2 >= n || isSpace(body[i+2])
			if before && after {
				j := i + 2
				for j < n && isSpace(body[j]) {
					j++
				}
				k := j
				for k < n && isProtoIdentChar(body[k]) {
					k++
				}
				if k > j {
					fieldName := body[j:k]
					if val, ok := values[fieldName]; ok {
						result.WriteString(val)
						i = k
						continue
					}
				}
			}
		}

		result.WriteByte(ch)
		i++
	}
	return result.String()
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isProtoIdentChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') || ch == '_' || ch == '-'
}

// ---------------------------------------------------------------------------
// parseEXTERNPROTOFull: stores a stub definition for EXTERNPROTO.
// ---------------------------------------------------------------------------

func (p *Parser) parseEXTERNPROTOFull() {
	p.lex.Next() // consume EXTERNPROTO

	if p.lex.Peek() != TokIdentifier {
		p.errorf("expected name after EXTERNPROTO")
		return
	}
	p.lex.Next()
	name := p.lex.StrVal()

	// Parse interface declarations [...]
	if p.lex.Next() != TokOpenBracket {
		p.errorf("expected '[' in EXTERNPROTO %s", name)
		return
	}
	fields := p.parseExternProtoInterface()

	// Parse URL list
	urls := p.parseURLList()

	def := &ProtoDefinition{
		Name:   name,
		Fields: fields,
		Body:   "",
		URLs:   urls,
	}
	p.protoTable[name] = def
}

// parseExternProtoInterface parses EXTERNPROTO interface (no default values).
func (p *Parser) parseExternProtoInterface() []ProtoFieldDecl {
	var fields []ProtoFieldDecl
	for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
		tok := p.lex.Peek()
		var kind FieldKind
		switch tok {
		case TokField:
			kind = KindField
		case TokExposedField:
			kind = KindExposedField
		case TokEventIn:
			kind = KindEventIn
		case TokEventOut:
			kind = KindEventOut
		default:
			p.lex.Next()
			continue
		}
		p.lex.Next()

		if p.lex.Next() != TokIdentifier {
			p.errorf("expected type name in EXTERNPROTO interface")
			continue
		}
		typeName := p.lex.StrVal()

		if p.lex.Next() != TokIdentifier {
			p.errorf("expected field name in EXTERNPROTO interface")
			continue
		}
		fieldName := p.lex.StrVal()

		fields = append(fields, ProtoFieldDecl{
			Kind:     kind,
			TypeName: typeName,
			Name:     fieldName,
		})
	}
	if p.lex.Peek() == TokCloseBracket {
		p.lex.Next()
	}
	return fields
}

// ProtoTable returns the proto definitions (for testing).
func (p *Parser) ProtoTable() map[string]*ProtoDefinition {
	return p.protoTable
}
