package parser

import (
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestSingleStringStruct(t *testing.T) {
	var test testsupport.SingleStringPmsStruct
	tp := reflect.ValueOf(&test)
	node, err := New("test-service", "dev", "").
		RegisterTagParser("pms", NewTagParser([]string{})).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	DumpNode(node)
}
