package asm

import (
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/stretchr/testify/assert"
)

func TestSingleStringAsmStruct(t *testing.T) {
	var test testsupport.SingleStringAsmStruct
	tp := reflect.ValueOf(&test)
	node, err := reflectparser.New("eap", "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	asmr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = asmr.Get(&node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	res := node.Value().Interface().(testsupport.SingleStringAsmStruct)
	assert.Equal(t, "The name", res.Name)
}

func TestStructWithSubStruct(t *testing.T) {
	var test testsupport.StructWithSubStruct
	tp := reflect.ValueOf(&test)
	node, err := reflectparser.New("eap", "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	asmr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = asmr.Get(&node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 43, test.AsmSub.Apa2)
	assert.Equal(t, "test svc name", test.AsmSub.Nu2)

	res := node.Value().Interface().(testsupport.StructWithSubStruct)
	assert.Equal(t, 43, res.AsmSub.Apa2)
	assert.Equal(t, "test svc name", res.AsmSub.Nu2)
}
