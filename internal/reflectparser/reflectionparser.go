package reflectparser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mariotoffia/ssm.git/internal/tagparser"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ReflectionParser is a parser that uses
// reflection to reflect structs and tags
type ReflectionParser struct {
	deployEnv string
	service   string
}

// New Creates a new reflection parser
func New(deployEnv string, service string) *ReflectionParser {
	return &ReflectionParser{deployEnv: deployEnv, service: service}
}

// Parse parses the value and creates a hiearchy to be used when marshal / unmarshal
func (p *ReflectionParser) Parse(prefix string, v reflect.Value) (SsmNode, error) {
	node := SsmNode{t: v.Type(), root: true, parent: nil}

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return node, errors.Errorf("Must pass struct by pointer and it must no be null - kind: %s", v.Kind().String())
	}

	// Dereference the pointer
	node.v = reflect.Indirect(v)
	nodes, err := p.parseStruct("", &node, prefix, node.v)

	if err != nil {
		return SsmNode{root: true, parent: nil}, err
	}

	node.childs = nodes
	return node, nil
}

func (p *ReflectionParser) parseStruct(nav string, parent *SsmNode, prefix string, v reflect.Value) ([]SsmNode, error) {
	t := v.Type()
	nodes := []SsmNode{}

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		switch fv.Kind() {
		case reflect.Struct:
			node, err := p.parseSubStruct(renderFqName(nav, ft), nodes, t, fv, ft, parent, prefix)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, *node)
		case reflect.Ptr:
			// Get the value it points to
			tv := reflect.Indirect(fv)
			if tv.IsValid() {
				node, err := p.parseSubStruct(renderFqName(nav, ft), nodes, t, tv, ft, parent, prefix)
				if err != nil {
					return nil, err
				}
				nodes = append(nodes, *node)
			}
		default:
			// Parse the tag on field
			tag, err := p.parseField(ft, prefix)
			if err != nil {
				return nil, errors.Errorf("The config %s could not parse field %s", t.Name(), ft.Name)
			}
			// Store tag for field
			if tag != nil {
				if e := log.Debug(); e.Enabled() {
					e.Str("svc", p.service).Msgf("struct: '%s' field: '%s' parsed: '%+v' full Name: '%s'", t.Name(), renderFqName(nav, ft), tag, tag.FullName())
				}

				nodes = append(nodes, SsmNode{fqname: renderFqName(nav, ft), t: t, f: ft, v: fv, tag: tag, root: false, parent: parent})
			}
		}
	}

	return nodes, nil
}

func renderFqName(nav string, ft reflect.StructField) string {
	if nav == "" {
		return ft.Name
	}
	return fmt.Sprintf("%s.%s", nav, ft.Name)
}

// The ft struct field is of a struct kind and hence we need to parse all it's
// fields and add those as children
func (p *ReflectionParser) parseSubStruct(nav string, nodes []SsmNode, t reflect.Type, fv reflect.Value,
	ft reflect.StructField, parent *SsmNode, prefix string) (*SsmNode, error) {

	node := SsmNode{fqname: nav, t: t, f: ft, v: fv, root: false, parent: parent}
	cn, err := p.parseStruct(nav, &node, prefix+"/"+strings.ToLower(ft.Name), fv)

	if err != nil {
		return nil, err
	}

	if len(cn) > 0 {
		node.childs = cn
	}
	return &node, nil
}

// Parses a single field in a structure and returns a ssmTag interface. If no tags
// is retrieved nil is returned. For example when a field is a sub struct hence this
// is a valid return. Errors may return.
func (p *ReflectionParser) parseField(f reflect.StructField, prefix string) (tagparser.SsmTag, error) {
	// Nothing to parse (this is not an error)
	if f.Tag == "" {
		return nil, nil
	}
	// Parameter store
	if value, ok := f.Tag.Lookup("pms"); ok {
		pmstag, err := tagparser.ParsePmsTagString(value, prefix, p.deployEnv, p.service)
		if err != nil {
			return nil, err
		}
		return pmstag, nil

	}
	// Secrets manager
	if value, ok := f.Tag.Lookup("asm"); ok {
		asmtag, err := tagparser.ParseAsmTagString(value, prefix, p.deployEnv, p.service)
		if err != nil {
			return nil, err
		}
		return asmtag, nil
	}
	// Nothing
	return nil, nil
}

// DumpNodes dumps info in the whole tree
func DumpNodes(nodes []SsmNode) {
	childNodes := []SsmNode{}

	for _, node := range nodes {
		if node.HasChildren() {
			childNodes = append(childNodes, node)
		}

		log.Debug().Msg(node.ToString(false))
	}

	for _, node := range childNodes {
		DumpNodes(node.Children())
	}
}
