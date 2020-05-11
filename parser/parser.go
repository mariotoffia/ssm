package parser

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Parser parses structs and produces a tree of nodes
// along with fields, values and possibly tags.
type Parser struct {
	// Contains a registration of what name of tag
	// and the parser instance to invoke. for exmaple
	// 'pms' and the parameter store tag parser instance.
	tagparsers map[string]TagParser
	// The service currently using this parser
	service string
	// The environment info
	environment string
	// If custom prefix is wanted instead of default or tag
	// based.
	prefix string
}

// New creates a new instrance of the Parser
func New(service string, environment string, prefix string) *Parser {
	return &Parser{
		tagparsers:  map[string]TagParser{},
		service:     service,
		environment: environment,
		prefix:      prefix,
	}
}

// RegisterTagParser registers a tag parser that parses
// a speicfied tag.
func (p *Parser) RegisterTagParser(tag string, parser TagParser) *Parser {
	p.tagparsers[tag] = parser
	return p
}

// Parse will parse the inparam value. It may either be a type
// such as var s MyStruct or a instance such as s := MyStruct{...}
// and then do reflect.ValueOf(&s) and send that to Parse.
func (p *Parser) Parse(v reflect.Value) (*StructNode, error) {
	node := &StructNode{Type: v.Type(), Owner: nil}

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return node, errors.Errorf("Must pass struct by pointer and it must no be null - kind: %s", v.Kind().String())
	}

	// Dereference the pointer
	node.Value = reflect.Indirect(v)

	nodes, err := p.parse("", node, node.Value)
	if err != nil {
		return nil, err
	}

	node.Childs = nodes
	return node, nil
}

// DumpNode dumps info in the whole tree
func DumpNode(node *StructNode) {
	dumpNodes(append([]StructNode{}, *node))
}

func dumpNodes(nodes []StructNode) {
	childNodes := []StructNode{}

	for _, node := range nodes {
		if len(node.Childs) > 0 {
			childNodes = append(childNodes, node)
		}

		log.Debug().Msg(node.ToString(false))
	}

	for _, node := range childNodes {
		dumpNodes(node.Childs)
	}
}
