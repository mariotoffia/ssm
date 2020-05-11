package parser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

func (p *Parser) parse(nav string, owner *StructNode, v reflect.Value) ([]StructNode, error) {
	t := v.Type()
	nodes := []StructNode{}

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		node, err := p.handleKind(nav, owner, t, fv, ft)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, *node)
	}

	return nodes, nil
}

func (p *Parser) handleKind(nav string, owner *StructNode, t reflect.Type, fv reflect.Value,
	ft reflect.StructField) (*StructNode, error) {

	var node *StructNode
	var err error

	switch fv.Kind() {
	case reflect.Struct:
		node, err = p.parseStruct(renderFqName(nav, ft), owner, t, fv, ft)
	case reflect.Ptr:
		tv := reflect.Indirect(fv)
		if tv.IsValid() {
			node, err = p.handleKind(nav, owner, t, fv, ft)
		}
	default:
		node, err = p.parseField(nav, owner, t, fv, ft)
	}

	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, errors.Errorf("Failed to handle kind %s nav %s field %v",
			fv.Kind().String(), nav, ft)
	}

	return node, nil
}

func (p *Parser) parseField(nav string, owner *StructNode, t reflect.Type, fv reflect.Value,
	ft reflect.StructField) (*StructNode, error) {

	tag, err := p.parseTag(nav, ft)
	if err != nil {
		return nil, errors.Errorf("The config %s could not parse field %s", t.Name(), ft.Name)
	}

	return &StructNode{
		FqName: renderFqName(nav, ft),
		Field:  ft,
		Owner:  owner,
		Type:   t,
		Value:  fv,
		Tag:    tag,
	}, nil
}

func (p *Parser) parseStruct(nav string, owner *StructNode, t reflect.Type, fv reflect.Value,
	ft reflect.StructField) (*StructNode, error) {
	node := StructNode{FqName: nav, Field: ft, Owner: owner, Type: t, Value: fv}
	nodes, err := p.parse(nav, &node, fv)

	if err != nil {
		return nil, err
	}

	if len(nodes) > 0 {
		node.Childs = nodes
	}

	return &node, nil
}

func (p *Parser) parseTag(nav string, ft reflect.StructField) (map[string]StructTag, error) {
	tags := map[string]StructTag{}

	if ft.Tag == "" {
		return tags, nil
	}

	for name, tagparser := range p.tagparsers {
		if tagstring, ok := ft.Tag.Lookup(name); ok {

			tag, err := tagparser.ParseTagString(tagstring,
				p.renderPrefix(nav),
				p.environment,
				p.service)

			if err != nil {
				return nil, err
			}

			tags[name] = tag
		}
	}

	return tags, nil
}

func (p *Parser) renderPrefix(nav string) string {
	if len(p.prefix) > 0 {
		if len(nav) > 0 {
			return strings.ReplaceAll(fmt.Sprintf("%s.%s", p.prefix, nav), ".", "/")
		}

		return strings.ReplaceAll(p.prefix, ".", "/")
	}
	return strings.ReplaceAll(nav, ".", "/")

}
func renderFqName(nav string, ft reflect.StructField) string {
	if nav == "" {
		return ft.Name
	}
	return fmt.Sprintf("%s.%s", nav, ft.Name)
}
