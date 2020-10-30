package common

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/mariotoffia/ssm/parser"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func setStructIntValue(rv reflect.Value, name string, value string) error {
	ival, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return errors.Wrapf(err, "Config value %s = %s is not a valid integer", name, value)
	}
	rv.SetInt(ival)
	return nil
}

// SetStructValueFromString sets a field in a struct to the specified value.
func SetStructValueFromString(node *parser.StructNode, name string, value string) error {

	log.Debug().Msgf("setting: %s (%s)", node.FqName, name)

	switch node.Value.Kind() {
	case reflect.Struct:
		setSubStructViaJSONString(node, value)
	case reflect.String:
		node.Value.SetString(value)

	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
		setStructIntValue(node.Value, name, value)
	}

	return nil
}

func setSubStructViaJSONString(node *parser.StructNode, value string) error {
	if err := json.Unmarshal([]byte(value), node.Value.Addr().Interface()); err != nil {
		return err
	}
	return nil
}

// getJSONViaSubStruct marshals a substruct into JSON payload
func getJSONViaSubStruct(node *parser.StructNode) (string, error) {
	data, err := json.Marshal(node.Value.Addr().Interface())
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetStringValueFromField retrieves the value from the field and
// converts it to a string
func GetStringValueFromField(node *parser.StructNode) string {

	switch node.Value.Kind() {
	case reflect.Struct:
		data, _ := getJSONViaSubStruct(node)
		return data
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
