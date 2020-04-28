package ssm

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type reflectionParser struct {
	deployEnv string
	service   string
}

func newReflectionParser(deployEnv string, service string) *reflectionParser {
	return &reflectionParser{deployEnv: deployEnv, service: service}
}

func (p *reflectionParser) parse(prefix string, v reflect.Value) (ssmNode, error) {
	node := ssmNode{t: v.Type(), v: v, root: true, parent: nil}
	nodes, err := p.parseStruct(&node, prefix, v)

	if err != nil {
		return ssmNode{root: true, parent: nil}, err
	}

	node.childs = nodes
	return node, nil
}

func (p *reflectionParser) parseStruct(parent *ssmNode, prefix string, v reflect.Value) ([]ssmNode, error) {
	t := v.Type()
	nodes := []ssmNode{}
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		switch fv.Kind() {
		case reflect.Struct:
			err := p.parseSubStruct(nodes, t, fv, ft, parent, prefix)
			if err != nil {
				return nil, err
			}
			continue
		case reflect.Ptr:
			// Get the value it points to
			tv := fv.Elem()
			if tv.IsValid() {
				err := p.parseSubStruct(nodes, t, tv, ft, parent, prefix)
				if err != nil {
					return nil, err
				}
				continue
			}
		}

		// Parse the tag on field
		tag, err := p.parseField(ft, prefix)
		if err != nil {
			return nil, errors.Errorf("The config %s could not parse field %s", t.Name(), ft.Name)
		}
		// Store tag for field
		if tag != nil {
			if e := log.Debug(); e.Enabled() {
				e.Str("svc", p.service).Msgf("struct: '%s' field: '%s' parsed: '%+v' full Name: '%s'", t.Name(), ft.Name, tag, tag.FullName())
			}

			nodes = append(nodes, ssmNode{t: t, f: ft, v: fv, tag: tag, root: false, parent: parent})
		}
	}

	return nodes, nil
}

// The ft struct field is of a struct kind and hence we need to parse all it's
// fields and add those as children
func (p *reflectionParser) parseSubStruct(nodes []ssmNode, t reflect.Type, fv reflect.Value,
	ft reflect.StructField, parent *ssmNode, prefix string) error {

	node := ssmNode{t: t, f: ft, v: fv, root: false, parent: parent}
	cn, err := p.parseStruct(&node, prefix+"/"+strings.ToLower(ft.Name), fv)

	if err != nil {
		return err
	}

	if len(cn) > 0 {
		node.childs = cn
		nodes = append(nodes, node)
	}
	return nil
}

// Parses a single field in a structure and returns a ssmTag interface. If no tags
// is retrieved nil is returned. For example when a field is a sub struct hence this
// is a valid return. Errors may return.
func (p *reflectionParser) parseField(f reflect.StructField, prefix string) (ssmTag, error) {
	// Nothing to parse (this is not an error)
	if f.Tag == "" {
		return nil, nil
	}
	// Parameter store
	if value, ok := f.Tag.Lookup("pms"); ok {
		pmstag, err := parsePmsTagString(value, prefix, p.deployEnv, p.service)
		if err != nil {
			return nil, err
		}
		return pmstag, nil

	}
	// Secrets manager
	if value, ok := f.Tag.Lookup("asm"); ok {
		asmtag, err := parseAsmTagString(value, prefix, p.deployEnv, p.service)
		if err != nil {
			return nil, err
		}
		return asmtag, nil
	}
	// Nothing
	return nil, nil
}

func dumpNodes(nodes []ssmNode) {
	childNodes := []ssmNode{}

	for _, node := range nodes {
		if node.HasChildren() {
			childNodes = append(childNodes, node)
		}

		log.Debug().Msg(node.ToString(true))
	}

	for _, node := range childNodes {
		dumpNodes(node.childs)
	}
}
