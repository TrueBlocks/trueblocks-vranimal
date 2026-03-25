// Package base provides the foundation types for the VRML scene graph:
// BaseNode (reference-counted base), RuntimeClass (dynamic type info),
// and the Field system for VRML field definitions.
// Ported from vraniml/src/utils/basenode.h, runtimeclass.h, field.h, fieldval.h.
package base

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Field type constants -- token IDs from fieldval.h
// ---------------------------------------------------------------------------

const (
	SFBOOL     = 258
	SFINT32    = 259
	SFFLOAT    = 260
	SFTIME     = 261
	SFSTRING   = 263
	SFVEC2F    = 264
	SFVEC3F    = 265
	SFCOLOR    = 266
	SFROTATION = 267
	SFIMAGE    = 268
	SFNODE     = 289

	MFINT32    = 290
	MFFLOAT    = 291
	MFSTRING   = 292
	MFVEC2F    = 293
	MFVEC3F    = 294
	MFCOLOR    = 295
	MFROTATION = 296
	MFNODE     = 297
)

// EventType flags for VRML field event semantics.
type EventType int

const (
	EventOut     EventType = 0
	EventIn      EventType = 1
	ExposedField EventType = 2
	FieldOnly    EventType = 3
)

// ---------------------------------------------------------------------------
// FieldValue -- a dynamic VRML field value (replaces C++ _fieldValUnion)
// ---------------------------------------------------------------------------

// FieldValue holds a typed VRML value. Use type assertions to extract.
type FieldValue = any

// ---------------------------------------------------------------------------
// FieldDef -- one field definition on a runtime class
// ---------------------------------------------------------------------------

// FieldDef describes one field in a VRML node class.
type FieldDef struct {
	Name      string
	Type      int       // SFBOOL, SFINT32, etc.
	EventType EventType // eventIn, eventOut, exposedField, field
	ID        int32
}

// ---------------------------------------------------------------------------
// RuntimeClass -- metadata about a VRML node class
// ---------------------------------------------------------------------------

// RuntimeClass holds metadata about a VRML node type, supporting dynamic
// creation and runtime type checking.
type RuntimeClass struct {
	ClassName  string
	BaseClass  *RuntimeClass
	CreateFunc func() any
	Fields     []FieldDef
}

// IsDerivedFrom returns true if this class is a descendant of other.
func (rc *RuntimeClass) IsDerivedFrom(other *RuntimeClass) bool {
	for cur := rc; cur != nil; cur = cur.BaseClass {
		if cur == other {
			return true
		}
	}
	return false
}

// FindField returns the FieldDef with the given name, or nil.
func (rc *RuntimeClass) FindField(name string) *FieldDef {
	for cur := rc; cur != nil; cur = cur.BaseClass {
		for i := range cur.Fields {
			if cur.Fields[i].Name == name {
				return &cur.Fields[i]
			}
		}
	}
	return nil
}

// FindFieldByID returns the FieldDef with the given ID, or nil.
func (rc *RuntimeClass) FindFieldByID(id int32) *FieldDef {
	for cur := rc; cur != nil; cur = cur.BaseClass {
		for i := range cur.Fields {
			if cur.Fields[i].ID == id {
				return &cur.Fields[i]
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Type Registry -- global registry for dynamic class creation
// ---------------------------------------------------------------------------

var (
	registryMu sync.RWMutex
	registry   = map[string]*RuntimeClass{}
)

// RegisterClass registers a RuntimeClass by name.
func RegisterClass(rc *RuntimeClass) {
	registryMu.Lock()
	registry[rc.ClassName] = rc
	registryMu.Unlock()
}

// LookupClass finds a RuntimeClass by name.
func LookupClass(name string) *RuntimeClass {
	registryMu.RLock()
	rc := registry[name]
	registryMu.RUnlock()
	return rc
}

// CreateByName creates a new instance of the named class, or nil if unknown.
func CreateByName(name string) any {
	rc := LookupClass(name)
	if rc == nil || rc.CreateFunc == nil {
		return nil
	}
	return rc.CreateFunc()
}

// ---------------------------------------------------------------------------
// BaseNode -- reference-counted base object (replaces vrBaseNode)
// ---------------------------------------------------------------------------

// BaseNode is the universal base for all VRML objects with reference counting.
type BaseNode struct {
	refCount int32
	class    *RuntimeClass
}

// Reference increments the reference count.
func (n *BaseNode) Reference() { atomic.AddInt32(&n.refCount, 1) }

// Dereference decrements the reference count.
func (n *BaseNode) Dereference() { atomic.AddInt32(&n.refCount, -1) }

// IsReferenced returns true if the reference count is > 0.
func (n *BaseNode) IsReferenced() bool { return atomic.LoadInt32(&n.refCount) > 0 }

// RefCount returns the current reference count.
func (n *BaseNode) RefCount() int32 { return atomic.LoadInt32(&n.refCount) }

// SetRuntimeClass sets the runtime class of this node.
func (n *BaseNode) SetRuntimeClass(rc *RuntimeClass) { n.class = rc }

// GetRuntimeClass returns the runtime class of this node.
func (n *BaseNode) GetRuntimeClass() *RuntimeClass { return n.class }

// IsKindOf returns true if this node's class is a descendant of the given class.
func (n *BaseNode) IsKindOf(rc *RuntimeClass) bool {
	if n.class == nil {
		return false
	}
	return n.class.IsDerivedFrom(rc)
}

// ClassName returns the class name or "unknown".
func (n *BaseNode) ClassName() string {
	if n.class != nil {
		return n.class.ClassName
	}
	return "unknown"
}

// ---------------------------------------------------------------------------
// DumpContext -- output helper with indentation
// ---------------------------------------------------------------------------

// DumpContext is an output helper with indentation support.
type DumpContext struct {
	Indent int
	buf    []byte
}

// NewDumpContext creates a new empty dump context.
func NewDumpContext() *DumpContext { return &DumpContext{} }

// Write appends a line with current indentation.
func (dc *DumpContext) Write(s string) {
	for i := 0; i < dc.Indent; i++ {
		dc.buf = append(dc.buf, ' ', ' ')
	}
	dc.buf = append(dc.buf, s...)
	dc.buf = append(dc.buf, '\n')
}

// Writef appends a formatted line with current indentation.
func (dc *DumpContext) Writef(format string, args ...any) {
	dc.Write(fmt.Sprintf(format, args...))
}

// String returns the accumulated output.
func (dc *DumpContext) String() string { return string(dc.buf) }

// ---------------------------------------------------------------------------
// Field -- a named, typed VRML field on a concrete node instance
// ---------------------------------------------------------------------------

// Field represents a single VRML field value on a node instance.
type Field struct {
	Name      string
	Type      int // SFBOOL, SFINT32, etc.
	EventType EventType
	ID        int32
	Value     FieldValue
}

// NewField creates a field with the given properties.
func NewField(name string, typ int, eventType EventType, id int32, val FieldValue) Field {
	return Field{Name: name, Type: typ, EventType: eventType, ID: id, Value: val}
}

// ---------------------------------------------------------------------------
// Typed accessors for FieldValue
// ---------------------------------------------------------------------------

// AsBool extracts a bool from a FieldValue.
func AsBool(v FieldValue) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// AsInt32 extracts an int32 from a FieldValue.
func AsInt32(v FieldValue) int32 {
	if i, ok := v.(int32); ok {
		return i
	}
	return 0
}

// AsFloat32 extracts a float32 from a FieldValue.
func AsFloat32(v FieldValue) float32 {
	if f, ok := v.(float32); ok {
		return f
	}
	return 0
}

// AsFloat64 extracts a float64 from a FieldValue.
func AsFloat64(v FieldValue) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

// AsString extracts a string from a FieldValue.
func AsString(v FieldValue) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// AsVec2f extracts a SFVec2f from a FieldValue.
func AsVec2f(v FieldValue) vec.SFVec2f {
	if vv, ok := v.(vec.SFVec2f); ok {
		return vv
	}
	return vec.SFVec2f{}
}

// AsVec3f extracts a SFVec3f from a FieldValue.
func AsVec3f(v FieldValue) vec.SFVec3f {
	if vv, ok := v.(vec.SFVec3f); ok {
		return vv
	}
	return vec.SFVec3f{}
}

// AsColor extracts a SFColor from a FieldValue.
func AsColor(v FieldValue) vec.SFColor {
	if c, ok := v.(vec.SFColor); ok {
		return c
	}
	return vec.SFColor{}
}

// AsRotation extracts a SFRotation from a FieldValue.
func AsRotation(v FieldValue) vec.SFRotation {
	if r, ok := v.(vec.SFRotation); ok {
		return r
	}
	return vec.SFRotation{}
}

// AsImage extracts a SFImage from a FieldValue.
func AsImage(v FieldValue) vec.SFImage {
	if img, ok := v.(vec.SFImage); ok {
		return img
	}
	return vec.SFImage{}
}

// AsInt32Slice extracts an []int32 from a FieldValue.
func AsInt32Slice(v FieldValue) []int32 {
	if s, ok := v.([]int32); ok {
		return s
	}
	return nil
}

// AsFloat32Slice extracts a []float32 from a FieldValue.
func AsFloat32Slice(v FieldValue) []float32 {
	if s, ok := v.([]float32); ok {
		return s
	}
	return nil
}

// AsStringSlice extracts a []string from a FieldValue.
func AsStringSlice(v FieldValue) []string {
	if s, ok := v.([]string); ok {
		return s
	}
	return nil
}

// AsVec2fSlice extracts a []SFVec2f from a FieldValue.
func AsVec2fSlice(v FieldValue) []vec.SFVec2f {
	if s, ok := v.([]vec.SFVec2f); ok {
		return s
	}
	return nil
}

// AsVec3fSlice extracts a []SFVec3f from a FieldValue.
func AsVec3fSlice(v FieldValue) []vec.SFVec3f {
	if s, ok := v.([]vec.SFVec3f); ok {
		return s
	}
	return nil
}

// AsColorSlice extracts a []SFColor from a FieldValue.
func AsColorSlice(v FieldValue) []vec.SFColor {
	if s, ok := v.([]vec.SFColor); ok {
		return s
	}
	return nil
}

// AsRotationSlice extracts a []SFRotation from a FieldValue.
func AsRotationSlice(v FieldValue) []vec.SFRotation {
	if s, ok := v.([]vec.SFRotation); ok {
		return s
	}
	return nil
}
