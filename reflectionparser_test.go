package ssm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SingleStringStruct struct {
	Name string `pms:"test, prefix=simple,tag1=nanna banna panna"`
}

type StructWithSubStruct struct {
	Name string `pms:"name"`
	Sub  struct {
		Apa int    `pms:"apa"`
		Nu  string `pms:"nisse"`
	}
}

func TestSingleStringStruct(t *testing.T) {
	var test SingleStringStruct
	tp := reflect.ValueOf(test)
	node, err := newReflectionParser("dev", "test-service").parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	nodes := []ssmNode{}
	dumpNodes(append(nodes, node))
}

func TestStructWithSubStruct(t *testing.T) {
	var test StructWithSubStruct
	tp := reflect.ValueOf(test)
	node, err := newReflectionParser("dev", "test-service").parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	nodes := []ssmNode{}
	dumpNodes(append(nodes, node))
}
