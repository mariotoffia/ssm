package pms

import (
	"flag"
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/mariotoffia/ssm.git/parser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

var stage string
var scope string

func init() {
	testing.Init() // Need to do this in order for flag.Parse() to work
	flag.StringVar(&scope, "scope", "", "Scope for test")
	flag.Parse()

	stage = testsupport.UnittestStage()
	log.Info().Msgf("Initializing PMS unittest with STAGE: %s", stage)

	err := testsupport.DefaultProvisionPms(stage)
	if err != nil {
		panic(err)
	}
}

func TestUnmarshalWihSingleStringStruct(t *testing.T) {
	var test testsupport.SingleStringPmsStruct
	tp := reflect.ValueOf(&test)
	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	res := node.Value.Interface().(testsupport.SingleStringPmsStruct)
	assert.Equal(t, "The name", res.Name)
}

func TestUnmarshalWihSingleNestedStruct(t *testing.T) {
	var test testsupport.StructWithSubStruct
	tp := reflect.ValueOf(&test)
	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 43, test.Sub.Apa)
	assert.Equal(t, "test svc name", test.Sub.Nu)

	res := node.Value.Interface().(testsupport.StructWithSubStruct)
	assert.Equal(t, "The name", res.Name)
}

func TestUnmarshalNestedJsonStructValue(t *testing.T) {
	var test testsupport.MyDbServiceConfig
	tp := reflect.ValueOf(&test)
	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, "gurka", test.Connection.User)
	assert.Equal(t, 17, test.Connection.Timeout)
	assert.Equal(t, "", test.Connection.Password)
}

func TestMarshalWihSingleStringStruct(t *testing.T) {
	if scope != "rw" {
		return
	}

	test := testsupport.SingleStringPmsStruct{Name: "my-custom name"}
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsr.Upsert(node, support.NewFilters())
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var tr testsupport.SingleStringPmsStruct
	tp = reflect.ValueOf(&tr)

	node, err = parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "my-custom name", tr.Name)
}

func TestMarshalWihSingleNestedStruct(t *testing.T) {
	if scope != "rw" {
		return
	}

	test := testsupport.StructWithSubStruct{Name: "kalle kula"}
	test.Sub.Apa = 18
	test.Sub.Nu = "krafso blafso"

	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsr.Upsert(node, support.NewFilters())
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	// Read it back
	var tr testsupport.StructWithSubStruct
	tp = reflect.ValueOf(&tr)

	node, err = parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "kalle kula", tr.Name)
	assert.Equal(t, 18, tr.Sub.Apa)
	assert.Equal(t, "krafso blafso", tr.Sub.Nu)
}

func TestMarshalSubStructAsJSON(t *testing.T) {
	if scope != "rw" {
		return
	}

	test := testsupport.MyDbServiceConfig{}
	test.Name = "My Connection Details"
	test.Connection.User = "gördis"
	test.Connection.Timeout = 1088
	test.Connection.Password = "åaaäs2##!!äöå!#dfmklvmlkBBCH2¤"
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsr.Upsert(node, support.NewFilters())
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	// Read it back
	var tr testsupport.MyDbServiceConfig
	tp = reflect.ValueOf(&tr)

	node, err = parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "My Connection Details", tr.Name)
	assert.Equal(t, "gördis", tr.Connection.User)
	assert.Equal(t, 1088, tr.Connection.Timeout)
	assert.Equal(t, "åaaäs2##!!äöå!#dfmklvmlkBBCH2¤", tr.Connection.Password)
}
