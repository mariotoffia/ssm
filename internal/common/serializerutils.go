package common

import (
	"reflect"
	"strconv"

	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/tagparser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

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
				log.Info().Msgf("ADDING %s = %v", node.Tag().FullName(), node)
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
		setStructIntValue(node, name, value)
	}

	return nil
}

func setStructIntValue(node *reflectparser.SsmNode, name string, value string) error {
	ival, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return errors.Wrapf(err, "Config value %s = %s is not a valid integer", name, value)
	}
	node.Value().SetInt(ival)
	return nil
}
