package ssm

import (
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestSingleStringStruct(t *testing.T) {
	var test testsupport.SingleStringStruct
	tp := reflect.ValueOf(&test)
	node, err := newReflectionParser("eap", "test-service").parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	nodes := []ssmNode{}
	dumpNodes(append(nodes, node))
}

func TestStructWithSubStruct(t *testing.T) {
	var test testsupport.StructWithSubStruct
	tp := reflect.ValueOf(&test)
	node, err := newReflectionParser("eap", "test-service").parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	nodes := []ssmNode{}
	dumpNodes(append(nodes, node))
}
