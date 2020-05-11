package common

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/tagparser"
	"github.com/mariotoffia/ssm.git/parser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// RenderPrefix renders a prefix based on the inparam strings
func RenderPrefix(prefix string, env string, svc string) string {
	if strings.HasPrefix(env, "/") {
		env = env[1:]
	}
	if strings.HasSuffix(env, "/") {
		env = env[:1]
	}
	if strings.HasPrefix(svc, "/") {
		svc = svc[1:]
	}
	if strings.HasSuffix(svc, "/") {
		svc = svc[:1]
	}
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if strings.HasSuffix(prefix, "/") {
		prefix = prefix[:1]
	}

	if prefix == "" {
		return fmt.Sprintf("/%s/%s", env, svc)
	}
	if svc == "" {
		return fmt.Sprintf("/%s%s", env, prefix)
	}
	return fmt.Sprintf("/%s/%s%s", env, svc, prefix)
}

// NodesToParameterMap grabs all FullNames on nodes that do have tag set
// in order to get data fom parameter store. Note that it chcks for the
// tag SsmType = st. The full name is the associated with the node itself.
// This is to gain a more accessable structure to seach for nodes.
func NodesToParameterMap(node *reflectparser.SsmNode,
	paths map[string]*reflectparser.SsmNode, filter *support.FieldFilters, st tagparser.StoreType) bool {
	issecure := false
	if node.HasChildren() {
		children := node.Children()
		for i := range node.Children() {
			if NodesToParameterMap(&children[i], paths, filter, st) {
				issecure = true
			}
		}
	} else {
		if node.Tag().SsmType() == st {
			if filter.IsIncluded(node.FqName()) {
				paths[node.Tag().FullName()] = node
				if node.Tag().Secure() {
					issecure = true
				}
			}
		}
	}

	return issecure
}

// ExtractParameters flattern the parameters into a single array
func ExtractParameters(paths map[string]*reflectparser.SsmNode) []string {
	arr := make([]string, 0, len(paths))
	for key := range paths {
		arr = append(arr, key)
	}

	return arr
}

// SetStructValueFromString sets a field in a struct to the specified value.
func SetStructValueFromString(node *reflectparser.SsmNode, name string, value string) error {

	log.Debug().Msgf("setting: %s (%s) val: %s", node.Tag().FullName(), name, value)

	switch node.Value().Kind() {

	case reflect.String:
		node.Value().SetString(value)

	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
		setStructIntValue(node.Value(), name, value)
	}

	return nil
}

func setStructIntValue(rv reflect.Value, name string, value string) error {
	ival, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return errors.Wrapf(err, "Config value %s = %s is not a valid integer", name, value)
	}
	rv.SetInt(ival)
	return nil
}

// SetStructValueFromString2 sets a field in a struct to the specified value.
func SetStructValueFromString2(node *parser.StructNode, name string, value string) error {

	log.Debug().Msgf("setting: %s (%s) val: %s", node.FqName, name, value)

	switch node.Value.Kind() {

	case reflect.String:
		node.Value.SetString(value)

	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
		setStructIntValue(node.Value, name, value)
	}

	return nil
}

// GetStringValueFromField retrieves the value from the field and
// converts it to a string
func GetStringValueFromField(node *reflectparser.SsmNode) string {

	switch node.Value().Kind() {
	case reflect.String:
		return node.Value().String()
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
		return strconv.FormatInt(node.Value().Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(node.Value().Uint(), 10)
	case reflect.Bool:
		return strconv.FormatBool(node.Value().Bool())
	case reflect.Float32:
		return strconv.FormatFloat(node.Value().Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(node.Value().Float(), 'f', -1, 64)
	}

	return ""
}

// GetStringValueFromField2 retrieves the value from the field and
// converts it to a string
func GetStringValueFromField2(node *parser.StructNode) string {

	switch node.Value.Kind() {
	case reflect.String:
		return node.Value.String()
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
		return strconv.FormatInt(node.Value.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(node.Value.Uint(), 10)
	case reflect.Bool:
		return strconv.FormatBool(node.Value.Bool())
	case reflect.Float32:
		return strconv.FormatFloat(node.Value.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(node.Value.Float(), 'f', -1, 64)
	}

	return ""
}
