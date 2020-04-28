package reflectparser

import (
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestSingleStringStruct(t *testing.T) {
	var test testsupport.SingleStringStruct
	tp := reflect.ValueOf(&test)
	node, err := NewReflectionParser("eap", "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	nodes := []SsmNode{}
	DumpNodes(append(nodes, node))
}

func TestStructWithSubStruct(t *testing.T) {
	var test testsupport.StructWithSubStruct
	tp := reflect.ValueOf(&test)
	node, err := NewReflectionParser("eap", "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	nodes := []SsmNode{}
	DumpNodes(append(nodes, node))
}
