package parser

import (
	"reflect"

	"github.com/mariotoffia/ssm/support"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Parser parses struct(s) and produces a tree of nodes
// along with fields, values and possibly tags.
type Parser struct {
	// Contains a registration of what name of tag
	// and the parser instance to invoke. for example
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
//
// When a _prefix_ is passed, it acts as a default prefix if no
// prefix is specified in the tag. Prefix operates under two modes
// _Local_ and _Global_.
//
// .Local vs Global Mode
// [cols="1,1,4"]
// |===
// |Mode |Example |Description
//
// |Local
// |my-local-prefix/nested
// |This will render environment/service/my-local-prefix/nested/property. E.g. dev/tes-service/my-local-prefix/nested/password
//
// |Global
// |/my-global-prefix/nested
// |This will render environment/my-global-prefix/nested/property. E.g. dev/my-global-prefix/nested/password
//
// |===
//
// NOTE: When global prefix, the _service_ element is eliminated (in order to have singeltons).
func New(service string, environment string, prefix string) *Parser {
	return &Parser{
		tagparsers:  map[string]TagParser{},
		service:     service,
		environment: environment,
		prefix:      prefix,
	}
}

// RegisterTagParser registers a tag parser that parses
// a specified tag.
func (p *Parser) RegisterTagParser(tag string, parser TagParser) *Parser {
	p.tagparsers[tag] = parser
	return p
}

// Parse will parse the in param value. It may either be a type
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

// NodesToParameterMap grabs all tag FullNames on nodes that do have at least
// one tag in the StructNode.Tag property. The tags full name is the associated
// with the node itself. This is to gain a more accessable structure to search
// for nodes. Note if multiple tag FullName are present for same StructNode,
// multiple entries in the paths map will be created, one per tag.FullName.
func NodesToParameterMap(node *StructNode,
	paths map[string]*StructNode, filter *support.FieldFilters, tags []string) {

	if filter.IsIncluded(node.FqName) {
		for _, tagname := range tags {
			if tag, ok := node.Tag[tagname]; ok {
				if fullName := tag.GetFullName(); fullName != "" {
					paths[tag.GetFullName()] = node
				}
			}
		}
	}

	if node.HasChildren() {
		children := node.Childs
		for i := range node.Childs {
			NodesToParameterMap(&children[i], paths, filter, tags)
		}
	}
}

// ExtractPaths extracts all keys in the paths map and adds
// them to an array.
func ExtractPaths(paths map[string]*StructNode) []string {
	arr := make([]string, 0, len(paths))
	for key := range paths {
		arr = append(arr, key)
	}

	return arr
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
