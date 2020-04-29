package reflectparser

import (
	"fmt"
	"reflect"

	"github.com/mariotoffia/ssm.git/internal/tagparser"
)

// SsmNode encapsulates a struct or a struct field
type SsmNode struct {
	// If this node is a root node or not.
	root bool
	// This is the fully qualified field name. If
	// within a substruct it will have a dotted path.
	// For example Address.Street. This means that
	// Street is this node and it is within a sub-struct
	// on a field called Address.
	fqname string
	// The struct type
	t reflect.Type
	// The field value (not set if root node)
	f reflect.StructField
	// The parsed tag (if sub-struct attr will be nil)
	tag tagparser.SsmTag
	// Any child nodes (if sub-struct)
	childs []SsmNode
	// The parent node (nil if root node)
	parent *SsmNode
	// Used if sub-/root struct
	v reflect.Value
}

// IsRoot determines if this node is a root node or a child
func (ssm *SsmNode) IsRoot() bool { return ssm.root }

// FqName returns the fully qualified name. Sub-struct fields
// have dotted navigation and the last component is the name
// of the field
func (ssm *SsmNode) FqName() string { return ssm.fqname }

// HasChildren checks if the node has child nodes (e.g. a struct field)
func (ssm *SsmNode) HasChildren() bool { return nil != ssm.childs }

// Children return the childs of this node if any. Use HasChildren before
// accessing this property
func (ssm *SsmNode) Children() []SsmNode { return ssm.childs }

// IsReflectValueNilPtr returns true if the value part is Nil ptr or a valid pointer
func (ssm *SsmNode) IsReflectValueNilPtr() bool { return !ssm.v.IsValid() }

// Tag obtains the underlying tag (if any)
func (ssm *SsmNode) Tag() tagparser.SsmTag { return ssm.tag }

// Field returns the underlying field (only root node will not have this set)
func (ssm *SsmNode) Field() reflect.StructField { return ssm.f }

// Value returns the underlying value (if any). Check if this value is valid
// sinc ptrs may not been initialized
func (ssm *SsmNode) Value() reflect.Value { return ssm.v }

// ToString renders the node
func (ssm *SsmNode) ToString(children bool) string {
	parent := ""
	if ssm.parent != nil {
		parent = ssm.f.Name
	}

	s := fmt.Sprintf("owning type: '%s' ('%s') field '%s' tag '%+v' parent-property '%s'",
		ssm.t.Kind().String(), ssm.t.Name(), ssm.fqname, ssm.tag, parent)

	if children && ssm.childs != nil && len(ssm.childs) > 0 {
		for _, chld := range ssm.childs {
			s = fmt.Sprintf("%s --> %s", s, chld.ToString(children))
		}
	}

	return s
}

// EnsureInstance ensures that value part is set
// If the embedded reflect.Value is a pointer type
// and it is nil. It will create a new instance and
// assing the it to the pointer. All this is done
// through reflection.
func (ssm *SsmNode) EnsureInstance(children bool) {
	// Only ptr and those who is nil can be created
	if ssm.v.Kind() != reflect.Ptr || ssm.v.IsValid() {
		return
	}

	ssm.v.Set(reflect.New(ssm.v.Elem().Type()))

	if !children {
		return
	}

	for _, child := range ssm.childs {
		child.EnsureInstance(children)
	}
}
