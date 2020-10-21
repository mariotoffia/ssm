package parser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFieldWithPrefixNoGlobal makes sure that inline prefix
// with no slash at the beginning gets inserted after service
// and before field name.
func TestFieldWithPrefixNoGlobal(t *testing.T) {

	type Test struct {
		Name string `pms:"test, prefix=simple,tag1=nanna banna panna"`
	}

	var test Test
	tp := reflect.ValueOf(&test)
	node, err := New("test-service", "dev", "").
		RegisterTagParser("pms", NewTagParser([]string{})).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t,
		"/dev/test-service/simple",
		node.Childs[0].Tag["pms"].GetNamed()["prefix"],
	)

	// DumpNode(node)
}

// TestFieldWithPrefixGlobal makes sure that inline prefix
// with a slash at the beginning gets inserted after environment
// and removes the service name. It is appended directly after
// the environment and thus is a _global_ parameter.
func TestFieldWithPrefixGlobal(t *testing.T) {

	type Test struct {
		Name string `pms:"test, prefix=/global/simple,tag1=nanna banna panna"`
	}

	var test Test
	tp := reflect.ValueOf(&test)
	node, err := New("test-service", "dev", "").
		RegisterTagParser("pms", NewTagParser([]string{})).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t,
		"/dev/global/simple",
		node.Childs[0].Tag["pms"].GetNamed()["prefix"],
	)

	// DumpNode(node)
}

func TestFieldNoGlobalPrefixInParser(t *testing.T) {

	type Test struct {
		HasLocalPrefix string `pms:"hasprefix, prefix=simple"`
		HasNoPrefix    string `pms:"hasnoprefix"`
	}

	var test Test
	tp := reflect.ValueOf(&test)
	node, err := New("test-service", "dev", "parser-prefix").
		RegisterTagParser("pms", NewTagParser([]string{})).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	prefix1 := node.Childs[0].Tag["pms"].GetNamed()["prefix"]
	prefix2 := node.Childs[1].Tag["pms"].GetNamed()["prefix"]

	assert.Equal(t, "/dev/test-service/simple", prefix1)
	assert.Equal(t, "/dev/test-service/parser-prefix", prefix2)
	//DumpNode(node)
}

func TestFieldGlobalPrefixInParser(t *testing.T) {

	type Test struct {
		HasLocalPrefix string `pms:"hasprefix, prefix=simple"`
		HasNoPrefix    string `pms:"hasnoprefix"`
	}

	var test Test
	tp := reflect.ValueOf(&test)
	node, err := New("test-service", "dev", "/global/parser-prefix").
		RegisterTagParser("pms", NewTagParser([]string{})).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	prefix1 := node.Childs[0].Tag["pms"].GetNamed()["prefix"]
	prefix2 := node.Childs[1].Tag["pms"].GetNamed()["prefix"]

	assert.Equal(t, "/dev/test-service/simple", prefix1)
	assert.Equal(t, "/dev/global/parser-prefix", prefix2)

	fmt.Println(prefix1)
	fmt.Println(prefix2)
	//DumpNode(node)
}

func TestNestedPropertyJsonExpanded(t *testing.T) {

	type Test struct {
		Connection struct {
			User     string `json:"user"`
			Password string `json:"password,omitempty"`
			Timeout  int    `json:"timeout"`
		} `pms:"bubbibobbo"`
	}

	var test Test
	tp := reflect.ValueOf(&test)
	node, err := New("test-service", "dev", "").
		RegisterTagParser("pms", NewTagParser([]string{})).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	DumpNode(node)
}
