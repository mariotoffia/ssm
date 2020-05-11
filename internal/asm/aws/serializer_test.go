package awsasm

import (
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/mariotoffia/ssm.git/parser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

// Since we need to remove all and if subsquent test is too fast run,
// the secrets manager will complain and faild stating that it is subject
// for deletion
var stage string

func init() {

	stage = testsupport.DefaultProvisionAsm()
	log.Info().Msgf("Initializing ASM unittest with STAGE: %s", stage)
}

func TestUnmarshalSingleStringAsmStruct(t *testing.T) {
	var test testsupport.SingleStringAsmStruct
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("asm", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	asmr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = asmr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	res := node.Value.Interface().(testsupport.SingleStringAsmStruct)
	assert.Equal(t, "The name", res.Name)
}

func TestUnmarshalStructWithSubStruct(t *testing.T) {
	var test testsupport.StructWithSubStruct
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("asm", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	asmr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = asmr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 43, test.AsmSub.Apa2)
	assert.Equal(t, "test svc name", test.AsmSub.Nu2)

	res := node.Value.Interface().(testsupport.StructWithSubStruct)
	assert.Equal(t, 43, res.AsmSub.Apa2)
	assert.Equal(t, "test svc name", res.AsmSub.Nu2)
}

func TestMarshalSingleStringAsmStruct(t *testing.T) {
	test := testsupport.SingleStringAsmStruct{Name: "testing write"}
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("asm", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	asmr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := asmr.Upsert(node, support.NewFilters())
	if len(result) > 0 {
		assert.Equal(t, nil, err)
	}

	var testr testsupport.SingleStringAsmStruct
	tpr := reflect.ValueOf(&testr)

	node, err = parser.New("test-service", stage, "").
		RegisterTagParser("asm", NewTagParser()).
		Parse(tpr)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = asmr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "testing write", testr.Name)
}

func TestMarshalStructWithSubStruct(t *testing.T) {
	test := testsupport.StructWithSubStruct{}
	test.AsmSub.Apa2 = 49
	test.AsmSub.Nu2 = "fluffy flow"

	tp := reflect.ValueOf(&test)
	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("asm", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	asmr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := asmr.Upsert(node, support.NewFilters())
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var testr testsupport.StructWithSubStruct
	tpr := reflect.ValueOf(&testr)

	node, err = parser.New("test-service", stage, "").
		RegisterTagParser("asm", NewTagParser()).
		Parse(tpr)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = asmr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 49, testr.AsmSub.Apa2)
	assert.Equal(t, "fluffy flow", testr.AsmSub.Nu2)
}
