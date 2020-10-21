package parser

import (
	"fmt"
	"reflect"
)

// TagParser where external caller implements to parse a tag. It is also possible
// to use a default tag parser within this package.
// The tag parser has two reserved tags "name" and "prefix" that it must adhere to
// check the default implementation for the handling of those.
type TagParser interface {
	// ParseTagString is implemented by the user of the parser. The tagstring is the complete
	// tag string for the registered tag name. The parameter prefix is the navigation
	// when having sub struct separated with slashes. In this way the parser may
	// add it as prefix on some occasions. The env is the in param environment data and
	// that can be anything. The svc is the service currently using this parser.
	//
	// The ParseTagString returns either a StructTag (struct) or an error.
	ParseTagString(tagstring string, prefix string, env string, svc string) (StructTag, error)
}

// StructTag is a interface struct
// that represents a tag with parameters
type StructTag interface {
	// Named is the named parametres indexed by name
	// If not key = value, instead just a value
	// the key is "" and the value is the value.
	GetNamed() map[string]string
	// The other key = values where not special
	// meaning
	GetTags() map[string]string
	// FullName is a fully qualified name separated
	// with slashes. This is a combination of a
	// prefix and the Named["name"] of this StructTag.
	GetFullName() string
}

// StructTagImpl is a generic struct
// that represents a tag with parameters
type StructTagImpl struct {
	// Gets the named parametres indexed by name
	// If not key = value, instead just a value
	// the key is "" and the value is the value.
	Named map[string]string
	// The other key = values where not special
	// meaning
	Tags map[string]string
	// FullName is a fully qualified name separated
	// with slashes. This is a combination of a
	// prefix and the Named["name"] of this StructTag.
	FullName string
}

// GetNamed gets the named parametres indexed by name. If not key = value,
// instead just a value the key is "" and the value is the value.
func (t *StructTagImpl) GetNamed() map[string]string { return t.Named }

// GetTags contains all tags that do not have a special meaning.
func (t *StructTagImpl) GetTags() map[string]string { return t.Tags }

// GetFullName is the fully qualified name, i.e. prefix + name
func (t *StructTagImpl) GetFullName() string { return t.FullName }

// StructNode is a node representing a struct
// or a field in a struct.
type StructNode struct {
	// FqName is the fully qualified field name. If
	// within a substruct it will have a dotted path.
	// For example Address.Street. This means that
	// Street is this node and it is within a sub-struct
	// on a field called Address.
	FqName string
	// Type is the the struct type
	Type reflect.Type
	// Field is the field value (not set if root node)
	Field reflect.StructField
	// Tag is the parsed tags (if sub-struct attr will be nil)
	// each tag is mapped to the registered tag name e.g. pms, asm etc.
	Tag map[string]StructTag
	// Childs is any child nodes (if sub-struct)
	Childs []StructNode
	// Owner is the owning node (nil if root node) that either
	// owns the scalar field or a sub-struct
	Owner *StructNode
	// Value is used if sub-/root struct
	Value reflect.Value
}

// HasChildren returns true if this node has children
func (s *StructNode) HasChildren() bool { return len(s.Childs) > 0 }

// HasTag checks if in param tag name exists int the Tag map
func (s *StructNode) HasTag(tag string) bool {
	_, ok := s.Tag[tag]
	return ok
}

// EnsureInstance ensures that value part is set
// If the embedded reflect.Value is a pointer type
// and it is nil. It will create a new instance and
// assing the it to the pointer. All this is done
// through reflection.
func (s *StructNode) EnsureInstance(children bool) {
	// Only ptr and those who is nil can be created
	if s.Value.Kind() != reflect.Ptr || s.Value.IsValid() {
		return
	}

	s.Value.Set(reflect.New(s.Value.Elem().Type()))

	if !children {
		return
	}

	for _, child := range s.Childs {
		child.EnsureInstance(children)
	}
}

// ToString renders the node
func (s *StructNode) ToString(children bool) string {
	owner := ""
	str := ""
	if s.Owner != nil {
		owner = s.Field.Name
	}

	str += fmt.Sprintf("[fqn = %s]: ", s.FqName)
	if len(s.Tag) > 0 {
		for key, value := range s.Tag {
			str += fmt.Sprintf("fullname: %s ", value.GetFullName())
			for nk, nv := range value.GetNamed() {
				str += fmt.Sprintf("[%s, Named] %s = %s, ", key, nk, nv)
			}
			for nk, nv := range value.GetTags() {
				str += fmt.Sprintf("[%s, Tags] %s = %s, ", key, nk, nv)
			}
		}
	}

	ot := s.Type.Kind().String()
	if s.Type.Kind() == reflect.Ptr {
		ot = "*" + s.Type.Elem().Name()
	}
	str += fmt.Sprintf("[owning type: '%s' ('%s') field '%s' tag '%v' owning-property '%s']",
		ot, s.Type.Name(), s.FqName, s.Tag, owner)

	if children && s.Childs != nil && len(s.Childs) > 0 {
		for _, child := range s.Childs {
			str += fmt.Sprintf("[%s --> %s]", str, child.ToString(children))
		}
	}

	return str
}
