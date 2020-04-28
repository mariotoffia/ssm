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
func (ssm *ssmNode) ToString(nochild bool) string {
	parent := ""
	if ssm.parent != nil {
		parent = ssm.f.Name
	}

	s := fmt.Sprintf("owning type: '%s' ('%s') field '%s' tag '%+v' parent-property '%s'",
		ssm.t.Kind().String(), ssm.t.Name(), ssm.f.Name, ssm.tag, parent)

	if !nochild && ssm.childs != nil && len(ssm.childs) > 0 {
		for _, chld := range ssm.childs {
			s = fmt.Sprintf("%s --> %s", s, chld.ToString(true))
		}
	}

	return s
}
