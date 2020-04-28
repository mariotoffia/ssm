package ssm

import (
	"fmt"
	"reflect"
)

type ssmNode struct {
	// If this node is a root node or not.
	root bool
	// The struct type
	t reflect.Type
	// The field value (not set if root node)
	f reflect.StructField
	// The parsed tag (if sub-struct attr will be nil)
	tag ssmTag
	// Any child nodes (if sub-struct)
	childs []ssmNode
	// The parent node (nil if root node)
	parent *ssmNode
	// Used if sub-/root struct
	v reflect.Value
}

func (ssm *ssmNode) IsRoot() bool               { return ssm.root }
func (ssm *ssmNode) HasChildren() bool          { return nil != ssm.childs }
func (ssm *ssmNode) IsReflectValueNilPtr() bool { return !ssm.v.IsValid() }
func (ssm *ssmNode) ToString(children bool) string {
	parent := ""
	if ssm.parent != nil {
		parent = ssm.f.Name
	}

	s := fmt.Sprintf("owning type: '%s' ('%s') field '%s' tag '%+v' parent-property '%s'",
		ssm.t.Kind().String(), ssm.t.Name(), ssm.f.Name, ssm.tag, parent)

	if children && ssm.childs != nil && len(ssm.childs) > 0 {
		for _, chld := range ssm.childs {
			s = fmt.Sprintf("%s --> %s", s, chld.ToString(children))
		}
	}

	return s
}

// If the embedded reflect.Value is a pointer type
// and it is nil. It will create a new instance and
// assing the it to the pointer. All this is done
// through reflection.
func (ssm *ssmNode) EnsureInstance(children bool) {
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
