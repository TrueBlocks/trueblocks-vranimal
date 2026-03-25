package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/node"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Token types (from tokens.h)
// ---------------------------------------------------------------------------

// Token represents a lexer token type.
type Token int

const (
	TokEOF Token = iota
	TokOpenBrace
	TokCloseBrace
	TokOpenBracket
	TokCloseBracket
	TokPeriod
	TokComma

	// Keywords
	TokDEF
	TokEXTERNPROTO
	TokFALSE
	TokIS
	TokNULL
	TokPROTO
	TokROUTE
	TokTO
	TokTRUE
	TokUSE
	TokEventIn
	TokEventOut
	TokExposedField
	TokField

	// VRML header
	TokHeader

	// Value tokens
	TokInt
	TokFloat
	TokString
	TokIdentifier
)

// ---------------------------------------------------------------------------
// Lexer
// ---------------------------------------------------------------------------

// Lexer tokenizes a VRML input stream.
type Lexer struct {
	reader   *bufio.Reader
	line     int
	tokType  Token
	intVal   int64
	floatVal float64
	strVal   string
	peeked   bool
}

// NewLexer creates a lexer from an io.Reader.
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		reader: bufio.NewReader(r),
		line:   1,
	}
}

// Line returns the current line number.
func (l *Lexer) Line() int { return l.line }

// Next advances to and returns the next token type.
func (l *Lexer) Next() Token {
	if l.peeked {
		l.peeked = false
		return l.tokType
	}
	l.tokType = l.scan()
	return l.tokType
}

// Peek returns the next token without consuming it.
func (l *Lexer) Peek() Token {
	if !l.peeked {
		l.tokType = l.scan()
		l.peeked = true
	}
	return l.tokType
}

// IntVal returns the last scanned integer value.
func (l *Lexer) IntVal() int64 { return l.intVal }

// FloatVal returns the last scanned float value.
func (l *Lexer) FloatVal() float64 { return l.floatVal }

// StrVal returns the last scanned string or identifier.
func (l *Lexer) StrVal() string { return l.strVal }

func (l *Lexer) scan() Token {
	l.skipWhitespaceAndComments()

	ch, err := l.readByte()
	if err != nil {
		return TokEOF
	}

	switch ch {
	case '{':
		return TokOpenBrace
	case '}':
		return TokCloseBrace
	case '[':
		return TokOpenBracket
	case ']':
		return TokCloseBracket
	case '.':
		// Peek: if the next byte is a digit, this is a number like .5
		next, nerr := l.readByte()
		if nerr != nil {
			return TokPeriod
		}
		if next >= '0' && next <= '9' {
			l.unreadByte()
			return l.scanNumber(ch)
		}
		l.unreadByte()
		return TokPeriod
	case ',':
		return TokComma
	case '"':
		return l.scanString()
	}

	if ch == '-' || ch == '+' || ch == '.' || (ch >= '0' && ch <= '9') {
		return l.scanNumber(ch)
	}

	if isIdentStart(ch) {
		return l.scanIdentifier(ch)
	}

	return TokEOF
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		ch, err := l.readByte()
		if err != nil {
			return
		}
		if ch == '\n' {
			l.line++
			continue
		}
		if ch == '#' {
			for {
				c, err := l.readByte()
				if err != nil || c == '\n' {
					l.line++
					break
				}
			}
			continue
		}
		if ch <= ' ' {
			continue
		}
		l.unreadByte()
		return
	}
}

func (l *Lexer) scanString() Token {
	var sb strings.Builder
	for {
		ch, err := l.readByte()
		if err != nil || ch == '"' {
			break
		}
		if ch == '\\' {
			esc, err := l.readByte()
			if err != nil {
				break
			}
			switch esc {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			default:
				sb.WriteByte(esc)
			}
			continue
		}
		if ch == '\n' {
			l.line++
		}
		sb.WriteByte(ch)
	}
	l.strVal = sb.String()
	return TokString
}

func (l *Lexer) scanNumber(first byte) Token {
	var sb strings.Builder
	sb.WriteByte(first)
	isFloat := first == '.'
	for {
		ch, err := l.readByte()
		if err != nil {
			break
		}
		if ch == '.' || ch == 'e' || ch == 'E' || ch == '+' || ch == '-' {
			if ch == '.' || ch == 'e' || ch == 'E' {
				isFloat = true
			}
			sb.WriteByte(ch)
			continue
		}
		if ch >= '0' && ch <= '9' {
			sb.WriteByte(ch)
			continue
		}
		if ch == 'x' || ch == 'X' || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
			sb.WriteByte(ch)
			continue
		}
		l.unreadByte()
		break
	}
	s := sb.String()
	if isFloat {
		l.floatVal, _ = strconv.ParseFloat(s, 64)
		return TokFloat
	}
	l.intVal, _ = strconv.ParseInt(s, 0, 64)
	l.floatVal = float64(l.intVal)
	return TokInt
}

func (l *Lexer) scanIdentifier(first byte) Token {
	var sb strings.Builder
	sb.WriteByte(first)
	for {
		ch, err := l.readByte()
		if err != nil {
			break
		}
		if isIdentPart(ch) {
			sb.WriteByte(ch)
		} else {
			l.unreadByte()
			break
		}
	}
	l.strVal = sb.String()
	return l.classifyKeyword(l.strVal)
}

func (l *Lexer) classifyKeyword(s string) Token {
	switch s {
	case "DEF":
		return TokDEF
	case "EXTERNPROTO":
		return TokEXTERNPROTO
	case "FALSE":
		return TokFALSE
	case "IS":
		return TokIS
	case "NULL":
		return TokNULL
	case "PROTO":
		return TokPROTO
	case "ROUTE":
		return TokROUTE
	case "TO":
		return TokTO
	case "TRUE":
		return TokTRUE
	case "USE":
		return TokUSE
	case "eventIn":
		return TokEventIn
	case "eventOut":
		return TokEventOut
	case "exposedField":
		return TokExposedField
	case "field":
		return TokField
	}
	return TokIdentifier
}

func (l *Lexer) readByte() (byte, error) {
	return l.reader.ReadByte()
}

func (l *Lexer) unreadByte() {
	_ = l.reader.UnreadByte()
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9') || ch == '-' || ch == '+'
}

// ---------------------------------------------------------------------------
// Parser
// ---------------------------------------------------------------------------

// Parser reads VRML tokens and constructs a scene graph.
type Parser struct {
	lex        *Lexer
	defTable   map[string]node.Node
	protoTable map[string]*ProtoDefinition
	errors     []string
	baseDir    string
	urlFetcher func(url string) (io.ReadCloser, error)
	routes     []*node.Route
	// Temp storage for interpolator keyValue parsing
	lastMFVec3f    []vec.SFVec3f
	lastMFRotation []vec.SFRotation
	lastMFColor    []vec.SFColor
	lastMFFloat    []float32
}

// NewParser creates a parser from an io.Reader.
func NewParser(r io.Reader) *Parser {
	return &Parser{
		lex:        NewLexer(r),
		defTable:   make(map[string]node.Node),
		protoTable: make(map[string]*ProtoDefinition),
	}
}

// SetBaseDir sets the base directory for resolving relative EXTERNPROTO URLs.
func (p *Parser) SetBaseDir(dir string) { p.baseDir = dir }

// SetURLFetcher sets a custom URL fetcher (for testing).
func (p *Parser) SetURLFetcher(f func(string) (io.ReadCloser, error)) { p.urlFetcher = f }

// Errors returns any parse errors encountered.
func (p *Parser) Errors() []string { return p.errors }

func (p *Parser) errorf(format string, args ...any) {
	msg := fmt.Sprintf("line %d: %s", p.lex.Line(), fmt.Sprintf(format, args...))
	p.errors = append(p.errors, msg)
}

// Parse reads the entire VRML file and returns the top-level children.
func (p *Parser) Parse() []node.Node {
	var children []node.Node
	for p.lex.Peek() != TokEOF {
		n := p.parseStatement()
		if n != nil {
			children = append(children, n)
		}
	}
	return children
}

func (p *Parser) parseStatement() node.Node {
	tok := p.lex.Peek()
	switch tok {
	case TokDEF:
		return p.parseDEF()
	case TokUSE:
		return p.parseUSE()
	case TokROUTE:
		p.parseROUTE()
		return nil
	case TokPROTO:
		p.parsePROTO()
		return nil
	case TokEXTERNPROTO:
		p.parseEXTERNPROTO()
		return nil
	case TokIdentifier:
		return p.parseNode()
	default:
		p.lex.Next()
		return nil
	}
}

func (p *Parser) parseDEF() node.Node {
	p.lex.Next()
	if p.lex.Next() != TokIdentifier {
		p.errorf("expected name after DEF")
		return nil
	}
	name := p.lex.StrVal()
	n := p.parseNode()
	if n != nil {
		n.SetName(name)
		p.defTable[name] = n
	}
	return n
}

func (p *Parser) parseUSE() node.Node {
	p.lex.Next()
	if p.lex.Next() != TokIdentifier {
		p.errorf("expected name after USE")
		return nil
	}
	name := p.lex.StrVal()
	n, ok := p.defTable[name]
	if !ok {
		p.errorf("USE of undefined name: %s", name)
		return nil
	}
	return n
}

func (p *Parser) parseROUTE() {
	p.lex.Next() // consume ROUTE
	// ROUTE srcNode.srcField TO dstNode.dstField
	srcName, srcField := p.parseDottedPair()
	if p.lex.Peek() == TokTO {
		p.lex.Next() // consume TO
	}
	dstName, dstField := p.parseDottedPair()

	srcNode, srcOK := p.defTable[srcName]
	dstNode, dstOK := p.defTable[dstName]
	if srcOK && dstOK {
		r := node.NewRoute(srcNode, srcField, dstNode, dstField)
		p.routes = append(p.routes, r)
	}
}

// parseDottedPair consumes "nodeName.fieldName" and returns both parts.
func (p *Parser) parseDottedPair() (string, string) {
	nodeName := ""
	fieldName := ""
	if p.lex.Peek() == TokIdentifier {
		p.lex.Next()
		nodeName = p.lex.StrVal()
	}
	if p.lex.Peek() == TokPeriod {
		p.lex.Next() // consume '.'
		if p.lex.Peek() == TokIdentifier {
			p.lex.Next()
			fieldName = p.lex.StrVal()
		}
	}
	return nodeName, fieldName
}

// GetRoutes returns all ROUTE statements parsed from the file.
func (p *Parser) GetRoutes() []*node.Route {
	return p.routes
}

func (p *Parser) parsePROTO() {
	p.parsePROTOFull()
}

func (p *Parser) parseEXTERNPROTO() {
	p.parseEXTERNPROTOFull()
}

func (p *Parser) parseNode() node.Node {
	if p.lex.Next() != TokIdentifier {
		return nil
	}
	typeName := p.lex.StrVal()
	if p.lex.Next() != TokOpenBrace {
		p.errorf("expected '{' after node type %s", typeName)
		return nil
	}

	n := p.createNode(typeName)
	if n == nil {
		// Check if this is a PROTO instance
		if def, ok := p.protoTable[typeName]; ok {
			nodes := p.instantiateProto(def)
			if len(nodes) == 1 {
				return nodes[0]
			}
			if len(nodes) > 1 {
				g := &node.Group{}
				g.BboxSize = vec.SFVec3f{X: -1, Y: -1, Z: -1}
				g.Children = nodes
				return g
			}
			return nil
		}
		// Unknown node — skip balanced braces
		depth := 1
		for depth > 0 {
			tok := p.lex.Next()
			switch tok {
			case TokOpenBrace:
				depth++
			case TokCloseBrace:
				depth--
			case TokEOF:
				return nil
			}
		}
		return nil
	}

	p.parseNodeBody(n, typeName)

	// Resolve Inline nodes: fetch external file and populate children
	if inl, ok := n.(*node.Inline); ok && len(inl.URL) > 0 {
		p.resolveInline(inl)
	}

	if p.lex.Next() != TokCloseBrace {
		p.errorf("expected '}' for node type %s", typeName)
	}
	return n
}

func (p *Parser) createNode(typeName string) node.Node {
	switch typeName {
	case "Appearance":
		return &node.Appearance{}
	case "Material":
		return node.NewMaterial()
	case "ImageTexture":
		return node.NewImageTexture()
	case "MovieTexture":
		return node.NewMovieTexture()
	case "PixelTexture":
		return node.NewPixelTexture()
	case "TextureTransform":
		return node.NewTextureTransform()
	case "FontStyle":
		return node.NewFontStyle()
	case "Background":
		return &node.Background{}
	case "Fog":
		return node.NewFog()
	case "NavigationInfo":
		return node.NewNavigationInfo()
	case "Viewpoint":
		return node.NewViewpoint()
	case "Shape":
		return &node.Shape{}
	case "DirectionalLight":
		return node.NewDirectionalLight()
	case "PointLight":
		return node.NewPointLight()
	case "SpotLight":
		return node.NewSpotLight()
	case "WorldInfo":
		return &node.WorldInfo{}
	case "Script":
		return &node.Script{}
	case "Sound":
		return node.NewSound()
	case "AudioClip":
		return node.NewAudioClip()
	case "Transform":
		return node.NewTransform()
	case "Group":
		g := &node.Group{}
		g.BboxSize = vec.SFVec3f{X: -1, Y: -1, Z: -1}
		return g
	case "Anchor":
		return &node.Anchor{}
	case "Billboard":
		return node.NewBillboard()
	case "Collision":
		return node.NewCollision()
	case "Inline":
		return &node.Inline{}
	case "LOD":
		return node.NewLOD()
	case "Switch":
		return node.NewSwitch()
	case "Box":
		return node.NewBox()
	case "Sphere":
		return node.NewSphere()
	case "Cone":
		return node.NewCone()
	case "Cylinder":
		return node.NewCylinder()
	case "Extrusion":
		return node.NewExtrusion()
	case "Text":
		return &node.Text{}
	case "IndexedFaceSet":
		return node.NewIndexedFaceSet()
	case "IndexedLineSet":
		return node.NewIndexedLineSet()
	case "PointSet":
		return node.NewPointSet()
	case "ElevationGrid":
		return node.NewElevationGrid()
	case "Color":
		return &node.ColorNode{}
	case "Coordinate":
		return &node.Coordinate{}
	case "Normal":
		return &node.NormalNode{}
	case "TextureCoordinate":
		return &node.TextureCoordinate{}
	case "ColorInterpolator":
		return &node.ColorInterpolator{}
	case "PositionInterpolator":
		return &node.PositionInterpolator{}
	case "OrientationInterpolator":
		return &node.OrientationInterpolator{}
	case "ScalarInterpolator":
		return &node.ScalarInterpolator{}
	case "CoordinateInterpolator":
		return &node.CoordinateInterpolator{}
	case "NormalInterpolator":
		return &node.NormalInterpolator{}
	case "TouchSensor":
		return node.NewTouchSensor()
	case "TimeSensor":
		return node.NewTimeSensor()
	case "ProximitySensor":
		return node.NewProximitySensor()
	case "CylinderSensor":
		return node.NewCylinderSensor()
	case "PlaneSensor":
		return node.NewPlaneSensor()
	case "SphereSensor":
		return node.NewSphereSensor()
	case "VisibilitySensor":
		return node.NewVisibilitySensor()
	}
	return nil
}

func (p *Parser) parseNodeBody(n node.Node, typeName string) {
	for p.lex.Peek() != TokCloseBrace && p.lex.Peek() != TokEOF {
		tok := p.lex.Peek()
		if tok == TokIdentifier {
			p.lex.Next()
			fieldName := p.lex.StrVal()
			p.parseFieldValue(n, typeName, fieldName)
		} else {
			p.lex.Next()
		}
	}
}

func (p *Parser) parseFieldValue(n node.Node, typeName, fieldName string) {
	_ = typeName
	switch v := n.(type) {
	case *node.Transform:
		p.parseTransformField(v, fieldName)
	case *node.Group:
		p.parseGroupField(&v.GroupingNode, fieldName)
	case *node.Shape:
		p.parseShapeField(v, fieldName)
	case *node.Appearance:
		p.parseAppearanceField(v, fieldName)
	case *node.Material:
		p.parseMaterialField(v, fieldName)
	case *node.Box:
		p.parseBoxField(v, fieldName)
	case *node.Sphere:
		p.parseSphereField(v, fieldName)
	case *node.Cone:
		p.parseConeField(v, fieldName)
	case *node.Cylinder:
		p.parseCylinderField(v, fieldName)
	case *node.IndexedFaceSet:
		p.parseDataSetField(&v.DataSet, fieldName)
	case *node.IndexedLineSet:
		p.parseDataSetField(&v.DataSet, fieldName)
	case *node.PointSet:
		p.parseDataSetField(&v.DataSet, fieldName)
	case *node.ElevationGrid:
		p.parseElevationGridField(v, fieldName)
	case *node.Coordinate:
		p.parseCoordinateField(v, fieldName)
	case *node.NormalNode:
		p.parseNormalField(v, fieldName)
	case *node.ColorNode:
		p.parseColorNodeField(v, fieldName)
	case *node.TextureCoordinate:
		p.parseTexCoordField(v, fieldName)
	case *node.ImageTexture:
		p.parseImageTextureField(v, fieldName)
	case *node.Viewpoint:
		p.parseViewpointField(v, fieldName)
	case *node.DirectionalLight:
		p.parseDirLightField(v, fieldName)
	case *node.PointLight:
		p.parsePointLightField(v, fieldName)
	case *node.SpotLight:
		p.parseSpotLightField(v, fieldName)
	case *node.Inline:
		p.parseInlineField(v, fieldName)
	case *node.NavigationInfo:
		p.parseNavigationInfoField(v, fieldName)
	case *node.Fog:
		p.parseFogField(v, fieldName)
	case *node.Switch:
		p.parseSwitchField(v, fieldName)
	case *node.LOD:
		p.parseLODField(v, fieldName)
	case *node.Anchor:
		p.parseAnchorField(v, fieldName)
	case *node.Billboard:
		p.parseBillboardField(v, fieldName)
	case *node.Collision:
		p.parseCollisionField(v, fieldName)
	case *node.TimeSensor:
		p.parseTimeSensorField(v, fieldName)
	case *node.PositionInterpolator:
		p.parseInterpolatorField(&v.Interpolator, fieldName)
		if fieldName == "keyValue" {
			v.KeyValue = p.lastMFVec3f
		}
	case *node.OrientationInterpolator:
		p.parseInterpolatorField(&v.Interpolator, fieldName)
		if fieldName == "keyValue" {
			v.KeyValue = p.lastMFRotation
		}
	case *node.ColorInterpolator:
		p.parseInterpolatorField(&v.Interpolator, fieldName)
		if fieldName == "keyValue" {
			v.KeyValue = p.lastMFColor
		}
	case *node.ScalarInterpolator:
		p.parseInterpolatorField(&v.Interpolator, fieldName)
		if fieldName == "keyValue" {
			v.KeyValue = p.lastMFFloat
		}
	case *node.CoordinateInterpolator:
		p.parseInterpolatorField(&v.Interpolator, fieldName)
		if fieldName == "keyValue" {
			v.KeyValue = p.lastMFVec3f
		}
	case *node.NormalInterpolator:
		p.parseInterpolatorField(&v.Interpolator, fieldName)
		if fieldName == "keyValue" {
			v.KeyValue = p.lastMFVec3f
		}
	case *node.TouchSensor:
		p.parseTouchSensorField(v, fieldName)
	case *node.ProximitySensor:
		p.parseProximitySensorField(v, fieldName)
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseTransformField(t *node.Transform, field string) {
	switch field {
	case "children":
		p.parseChildren(&t.GroupingNode)
	case "translation":
		t.Translation = p.parseVec3f()
	case "rotation":
		t.Rotation = p.parseRotation()
	case "scale":
		t.Scale = p.parseVec3f()
	case "scaleOrientation":
		t.ScaleOrientation = p.parseRotation()
	case "center":
		t.Center = p.parseVec3f()
	case "bboxCenter":
		t.BboxCenter = p.parseVec3f()
	case "bboxSize":
		t.BboxSize = p.parseVec3f()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseGroupField(g *node.GroupingNode, field string) {
	switch field {
	case "children":
		p.parseChildren(g)
	case "bboxCenter":
		g.BboxCenter = p.parseVec3f()
	case "bboxSize":
		g.BboxSize = p.parseVec3f()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseShapeField(s *node.Shape, field string) {
	switch field {
	case "appearance":
		n := p.parseStatement()
		if a, ok := n.(*node.Appearance); ok {
			s.Appearance = a
		}
	case "geometry":
		n := p.parseStatement()
		if g, ok := n.(node.GeometryNode); ok {
			s.Geometry = g
		}
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseAppearanceField(a *node.Appearance, field string) {
	switch field {
	case "material":
		n := p.parseStatement()
		if m, ok := n.(*node.Material); ok {
			a.Material = m
		}
	case "texture":
		n := p.parseStatement()
		a.Texture = n
	case "textureTransform":
		n := p.parseStatement()
		if tt, ok := n.(*node.TextureTransform); ok {
			a.TextureTransform = tt
		}
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseMaterialField(m *node.Material, field string) {
	switch field {
	case "ambientIntensity":
		m.AmbientIntensity = p.parseFloat()
	case "diffuseColor":
		c := p.parseVec3f()
		m.DiffuseColor = vec.NewColor(c.X, c.Y, c.Z)
	case "emissiveColor":
		c := p.parseVec3f()
		m.EmissiveColor = vec.NewColor(c.X, c.Y, c.Z)
	case "shininess":
		m.Shininess = p.parseFloat()
	case "specularColor":
		c := p.parseVec3f()
		m.SpecularColor = vec.NewColor(c.X, c.Y, c.Z)
	case "transparency":
		m.Transparency = p.parseFloat()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseBoxField(b *node.Box, field string) {
	if field == "size" {
		b.Size = p.parseVec3f()
	} else {
		p.skipFieldValue()
	}
}

func (p *Parser) parseSphereField(s *node.Sphere, field string) {
	if field == "radius" {
		s.Radius = p.parseFloat()
	} else {
		p.skipFieldValue()
	}
}

func (p *Parser) parseConeField(c *node.Cone, field string) {
	switch field {
	case "bottomRadius":
		c.BottomRadius = p.parseFloat()
	case "height":
		c.Height = p.parseFloat()
	case "side":
		c.Side = p.parseBool()
	case "bottom":
		c.Bottom = p.parseBool()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseCylinderField(c *node.Cylinder, field string) {
	switch field {
	case "bottom":
		c.Bottom = p.parseBool()
	case "height":
		c.Height = p.parseFloat()
	case "radius":
		c.Radius = p.parseFloat()
	case "side":
		c.Side = p.parseBool()
	case "top":
		c.Top = p.parseBool()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseDataSetField(ds *node.DataSet, field string) {
	switch field {
	case "color":
		n := p.parseStatement()
		if c, ok := n.(*node.ColorNode); ok {
			ds.Color = c
		}
	case "coord":
		n := p.parseStatement()
		if c, ok := n.(*node.Coordinate); ok {
			ds.Coord = c
		}
	case "normal":
		n := p.parseStatement()
		if c, ok := n.(*node.NormalNode); ok {
			ds.Normal = c
		}
	case "texCoord":
		n := p.parseStatement()
		if c, ok := n.(*node.TextureCoordinate); ok {
			ds.TexCoord = c
		}
	case "colorIndex":
		ds.ColorIndex = p.parseMFInt32()
	case "coordIndex":
		ds.CoordIndex = p.parseMFInt32()
	case "normalIndex":
		ds.NormalIndex = p.parseMFInt32()
	case "texCoordIndex":
		ds.TexCoordIndex = p.parseMFInt32()
	case "colorPerVertex":
		ds.ColorPerVertex = p.parseBool()
	case "normalPerVertex":
		ds.NormalPerVertex = p.parseBool()
	case "ccw":
		ds.Ccw = p.parseBool()
	case "convex":
		ds.Convex = p.parseBool()
	case "creaseAngle":
		ds.CreaseAngle = p.parseFloat()
	case "solid":
		ds.IsSolid = p.parseBool()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseElevationGridField(eg *node.ElevationGrid, field string) {
	switch field {
	case "height":
		eg.Heights = p.parseMFFloat()
	case "xDimension":
		eg.XDimension = p.parseInt32()
	case "xSpacing":
		eg.XSpacing = p.parseFloat()
	case "zDimension":
		eg.ZDimension = p.parseInt32()
	case "zSpacing":
		eg.ZSpacing = p.parseFloat()
	default:
		p.parseDataSetField(&eg.DataSet, field)
	}
}

func (p *Parser) parseCoordinateField(c *node.Coordinate, field string) {
	if field == "point" {
		c.Point = p.parseMFVec3f()
	} else {
		p.skipFieldValue()
	}
}

func (p *Parser) parseNormalField(n *node.NormalNode, field string) {
	if field == "vector" {
		n.Vector = p.parseMFVec3f()
	} else {
		p.skipFieldValue()
	}
}

func (p *Parser) parseColorNodeField(c *node.ColorNode, field string) {
	if field == "color" {
		c.Color = p.parseMFColor()
	} else {
		p.skipFieldValue()
	}
}

func (p *Parser) parseTexCoordField(tc *node.TextureCoordinate, field string) {
	if field == "point" {
		tc.Point = p.parseMFVec2f()
	} else {
		p.skipFieldValue()
	}
}

func (p *Parser) parseImageTextureField(it *node.ImageTexture, field string) {
	switch field {
	case "url":
		it.URL = p.parseMFString()
	case "repeatS":
		it.RepeatS = p.parseBool()
	case "repeatT":
		it.RepeatT = p.parseBool()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseViewpointField(vp *node.Viewpoint, field string) {
	switch field {
	case "description":
		vp.Description = p.parseString()
	case "fieldOfView":
		vp.FieldOfView = p.parseFloat()
	case "jump":
		vp.Jump = p.parseBool()
	case "orientation":
		vp.Orientation = p.parseRotation()
	case "position":
		vp.Position = p.parseVec3f()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseNavigationInfoField(ni *node.NavigationInfo, field string) {
	switch field {
	case "avatarSize":
		ni.AvatarSize = p.parseMFFloat()
	case "headlight":
		ni.Headlight = p.parseBool()
	case "speed":
		ni.Speed = p.parseFloat()
	case "type":
		ni.Type = p.parseMFString()
	case "visibilityLimit":
		ni.VisibilityLimit = p.parseFloat()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseFogField(f *node.Fog, field string) {
	switch field {
	case "color":
		c := p.parseVec3f()
		f.Color = vec.NewColor(c.X, c.Y, c.Z)
	case "fogType":
		f.FogType = p.parseString()
	case "visibilityRange":
		f.VisibilityRange = p.parseFloat()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseSwitchField(sw *node.Switch, field string) {
	switch field {
	case "whichChoice":
		sw.WhichChoice = p.parseInt32()
	case "choice":
		sw.Choice = p.parseMFNode()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseLODField(lod *node.LOD, field string) {
	switch field {
	case "center":
		lod.Center = p.parseVec3f()
	case "range":
		lod.Range = p.parseMFFloat()
	case "level":
		lod.Level = p.parseMFNode()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseAnchorField(a *node.Anchor, field string) {
	switch field {
	case "children":
		p.parseChildren(&a.GroupingNode)
	case "url":
		a.URL = p.parseMFString()
		a.OrigURL = append([]string(nil), a.URL...)
	case "description":
		a.Description = p.parseString()
	case "parameter":
		a.Parameter = p.parseMFString()
	case "bboxCenter":
		a.BboxCenter = p.parseVec3f()
	case "bboxSize":
		a.BboxSize = p.parseVec3f()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseBillboardField(bb *node.Billboard, field string) {
	switch field {
	case "children":
		p.parseChildren(&bb.GroupingNode)
	case "axisOfRotation":
		bb.AxisOfRotation = p.parseVec3f()
	case "bboxCenter":
		bb.BboxCenter = p.parseVec3f()
	case "bboxSize":
		bb.BboxSize = p.parseVec3f()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseCollisionField(col *node.Collision, field string) {
	switch field {
	case "children":
		p.parseChildren(&col.GroupingNode)
	case "collide":
		col.Collide = p.parseBool()
	case "proxy":
		col.Proxy = p.parseStatement()
	case "bboxCenter":
		col.BboxCenter = p.parseVec3f()
	case "bboxSize":
		col.BboxSize = p.parseVec3f()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseMFNode() []node.Node {
	var nodes []node.Node
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokComma {
				p.lex.Next()
				continue
			}
			child := p.parseStatement()
			if child != nil {
				nodes = append(nodes, child)
			}
		}
		p.lex.Next()
	} else {
		child := p.parseStatement()
		if child != nil {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

func (p *Parser) parseDirLightField(dl *node.DirectionalLight, field string) {
	switch field {
	case "direction":
		dl.Direction = p.parseVec3f()
	case "on":
		dl.On = p.parseBool()
	case "color":
		c := p.parseVec3f()
		dl.Color = vec.NewColor(c.X, c.Y, c.Z)
	case "intensity":
		dl.Intensity = p.parseFloat()
	case "ambientIntensity":
		dl.AmbientIntensity = p.parseFloat()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parsePointLightField(pl *node.PointLight, field string) {
	switch field {
	case "location":
		pl.Location = p.parseVec3f()
	case "radius":
		pl.Radius = p.parseFloat()
	case "on":
		pl.On = p.parseBool()
	case "color":
		c := p.parseVec3f()
		pl.Color = vec.NewColor(c.X, c.Y, c.Z)
	case "intensity":
		pl.Intensity = p.parseFloat()
	case "ambientIntensity":
		pl.AmbientIntensity = p.parseFloat()
	case "attenuation":
		pl.Attenuation = p.parseVec3f()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseSpotLightField(sl *node.SpotLight, field string) {
	switch field {
	case "location":
		sl.Location = p.parseVec3f()
	case "direction":
		sl.Direction = p.parseVec3f()
	case "radius":
		sl.Radius = p.parseFloat()
	case "beamWidth":
		sl.BeamWidth = p.parseFloat()
	case "cutOffAngle":
		sl.CutOffAngle = p.parseFloat()
	case "on":
		sl.On = p.parseBool()
	case "color":
		c := p.parseVec3f()
		sl.Color = vec.NewColor(c.X, c.Y, c.Z)
	case "intensity":
		sl.Intensity = p.parseFloat()
	case "ambientIntensity":
		sl.AmbientIntensity = p.parseFloat()
	case "attenuation":
		sl.Attenuation = p.parseVec3f()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseInlineField(inl *node.Inline, field string) {
	switch field {
	case "url":
		inl.URL = p.parseMFString()
	case "bboxCenter":
		inl.BboxCenter = p.parseVec3f()
	case "bboxSize":
		inl.BboxSize = p.parseVec3f()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) resolveInline(inl *node.Inline) {
	for _, rawURL := range inl.URL {
		r, err := p.fetchURL(rawURL)
		if err != nil {
			continue
		}
		sub := NewParser(r)
		sub.baseDir = p.baseDir
		sub.urlFetcher = p.urlFetcher
		sub.defTable = p.defTable
		sub.protoTable = p.protoTable
		nodes := sub.Parse()
		r.Close()
		for _, e := range sub.errors {
			p.errors = append(p.errors, e)
		}
		for k, v := range sub.defTable {
			p.defTable[k] = v
		}
		for k, v := range sub.protoTable {
			if _, exists := p.protoTable[k]; !exists {
				p.protoTable[k] = v
			}
		}
		inl.Children = nodes
		return
	}
}

func (p *Parser) parseChildren(g *node.GroupingNode) {
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokComma {
				p.lex.Next()
				continue
			}
			child := p.parseStatement()
			if child != nil {
				g.AddChild(child)
			}
		}
		p.lex.Next()
	} else {
		child := p.parseStatement()
		if child != nil {
			g.AddChild(child)
		}
	}
}

// ---------------------------------------------------------------------------
// Animation node field parsers
// ---------------------------------------------------------------------------

func (p *Parser) parseTimeSensorField(ts *node.TimeSensor, field string) {
	switch field {
	case "cycleInterval":
		ts.CycleInterval = float64(p.parseFloat())
	case "loop":
		ts.Loop = p.parseBool()
	case "startTime":
		ts.StartTime = float64(p.parseFloat())
	case "stopTime":
		ts.StopTime = float64(p.parseFloat())
	case "enabled":
		ts.Enabled = p.parseBool()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseInterpolatorField(interp *node.Interpolator, field string) {
	switch field {
	case "key":
		interp.Key = p.parseMFFloat()
	case "keyValue":
		// keyValue type depends on interpolator type; dispatch stores result
		// in lastMF* fields which the caller reads.
		p.parseInterpolatorKeyValue()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseInterpolatorKeyValue() {
	// We don't know the type here, so we try parsing as the most general form.
	// The caller (in parseFieldValue switch) reads the appropriate lastMF* field.
	// We peek at the structure: [ x y z, ... ] vs [ x y z w, ... ] etc.
	// Strategy: parse all floats, then the caller picks the right lastMF* field
	// based on its type. We store in ALL lastMF* fields that might apply.
	p.lastMFVec3f = nil
	p.lastMFRotation = nil
	p.lastMFColor = nil
	p.lastMFFloat = nil

	var floats []float32
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokComma {
				p.lex.Next()
				continue
			}
			floats = append(floats, p.parseFloat())
		}
		p.lex.Next()
	} else {
		floats = append(floats, p.parseFloat())
	}

	// Store as float array
	p.lastMFFloat = floats

	// Store as Vec3f array (groups of 3)
	if len(floats) >= 3 {
		for i := 0; i+2 < len(floats); i += 3 {
			p.lastMFVec3f = append(p.lastMFVec3f, vec.SFVec3f{X: floats[i], Y: floats[i+1], Z: floats[i+2]})
		}
	}

	// Store as Color array (groups of 3)
	if len(floats) >= 3 {
		for i := 0; i+2 < len(floats); i += 3 {
			p.lastMFColor = append(p.lastMFColor, vec.SFColor{R: floats[i], G: floats[i+1], B: floats[i+2]})
		}
	}

	// Store as Rotation array (groups of 4)
	if len(floats) >= 4 {
		for i := 0; i+3 < len(floats); i += 4 {
			p.lastMFRotation = append(p.lastMFRotation, vec.SFRotation{X: floats[i], Y: floats[i+1], Z: floats[i+2], W: floats[i+3]})
		}
	}
}

func (p *Parser) parseTouchSensorField(ts *node.TouchSensor, field string) {
	switch field {
	case "enabled":
		ts.Enabled = p.parseBool()
	default:
		p.skipFieldValue()
	}
}

func (p *Parser) parseProximitySensorField(ps *node.ProximitySensor, field string) {
	switch field {
	case "center":
		ps.Center = p.parseVec3f()
	case "size":
		ps.Size = p.parseVec3f()
	case "enabled":
		ps.Enabled = p.parseBool()
	default:
		p.skipFieldValue()
	}
}

// ---------------------------------------------------------------------------
// Primitive value parsers
// ---------------------------------------------------------------------------

func (p *Parser) parseFloat() float32 {
	tok := p.lex.Next()
	if tok == TokFloat || tok == TokInt {
		return float32(p.lex.FloatVal())
	}
	p.errorf("expected float, got %v", tok)
	return 0
}

func (p *Parser) parseInt32() int32 {
	tok := p.lex.Next()
	if tok == TokInt || tok == TokFloat {
		return int32(p.lex.IntVal())
	}
	p.errorf("expected int, got %v", tok)
	return 0
}

func (p *Parser) parseBool() bool {
	tok := p.lex.Next()
	if tok == TokTRUE {
		return true
	}
	if tok == TokFALSE {
		return false
	}
	if tok == TokIdentifier {
		return strings.EqualFold(p.lex.StrVal(), "true")
	}
	return false
}

func (p *Parser) parseString() string {
	if p.lex.Next() == TokString {
		return p.lex.StrVal()
	}
	return ""
}

func (p *Parser) parseVec2f() vec.SFVec2f {
	x := p.parseFloat()
	y := p.parseFloat()
	return vec.SFVec2f{X: x, Y: y}
}

func (p *Parser) parseVec3f() vec.SFVec3f {
	x := p.parseFloat()
	y := p.parseFloat()
	z := p.parseFloat()
	return vec.SFVec3f{X: x, Y: y, Z: z}
}

func (p *Parser) parseRotation() vec.SFRotation {
	x := p.parseFloat()
	y := p.parseFloat()
	z := p.parseFloat()
	w := p.parseFloat()
	return vec.SFRotation{X: x, Y: y, Z: z, W: w}
}

func (p *Parser) parseMFInt32() []int32 {
	var vals []int32
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokComma {
				p.lex.Next()
				continue
			}
			vals = append(vals, p.parseInt32())
		}
		p.lex.Next()
	} else {
		vals = append(vals, p.parseInt32())
	}
	return vals
}

func (p *Parser) parseMFFloat() []float32 {
	var vals []float32
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokComma {
				p.lex.Next()
				continue
			}
			vals = append(vals, p.parseFloat())
		}
		p.lex.Next()
	} else {
		vals = append(vals, p.parseFloat())
	}
	return vals
}

func (p *Parser) parseMFString() []string {
	var vals []string
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokComma {
				p.lex.Next()
				continue
			}
			vals = append(vals, p.parseString())
		}
		p.lex.Next()
	} else {
		vals = append(vals, p.parseString())
	}
	return vals
}

func (p *Parser) parseMFVec2f() []vec.SFVec2f {
	var vals []vec.SFVec2f
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokComma {
				p.lex.Next()
				continue
			}
			vals = append(vals, p.parseVec2f())
		}
		p.lex.Next()
	} else {
		vals = append(vals, p.parseVec2f())
	}
	return vals
}

func (p *Parser) parseMFVec3f() []vec.SFVec3f {
	var vals []vec.SFVec3f
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokComma {
				p.lex.Next()
				continue
			}
			vals = append(vals, p.parseVec3f())
		}
		p.lex.Next()
	} else {
		vals = append(vals, p.parseVec3f())
	}
	return vals
}

func (p *Parser) parseMFColor() []vec.SFColor {
	var vals []vec.SFColor
	if p.lex.Peek() == TokOpenBracket {
		p.lex.Next()
		for p.lex.Peek() != TokCloseBracket && p.lex.Peek() != TokEOF {
			if p.lex.Peek() == TokComma {
				p.lex.Next()
				continue
			}
			v := p.parseVec3f()
			vals = append(vals, vec.NewColor(v.X, v.Y, v.Z))
		}
		p.lex.Next()
	} else {
		v := p.parseVec3f()
		vals = append(vals, vec.NewColor(v.X, v.Y, v.Z))
	}
	return vals
}

func (p *Parser) skipFieldValue() {
	tok := p.lex.Peek()
	if tok == TokOpenBracket {
		p.lex.Next()
		depth := 1
		for depth > 0 {
			t := p.lex.Next()
			switch t {
			case TokOpenBracket:
				depth++
			case TokCloseBracket:
				depth--
			case TokEOF:
				return
			}
		}
		return
	}
	if tok == TokOpenBrace {
		p.skipBlock()
		return
	}
	p.lex.Next()
}

func (p *Parser) skipBlock() {
	if p.lex.Peek() == TokOpenBrace || p.lex.Peek() == TokOpenBracket {
		open := p.lex.Next()
		var closeT Token
		if open == TokOpenBrace {
			closeT = TokCloseBrace
		} else {
			closeT = TokCloseBracket
		}
		depth := 1
		for depth > 0 {
			t := p.lex.Next()
			switch t {
			case open:
				depth++
			case closeT:
				depth--
			case TokEOF:
				return
			}
		}
	}
}
