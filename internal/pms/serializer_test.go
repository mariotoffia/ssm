package pms

import (
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

var stage string

func init() {
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
	node, err := reflectparser.New(stage, "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(&node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	res := node.Value().Interface().(testsupport.SingleStringPmsStruct)
	assert.Equal(t, "The name", res.Name)
}

func TestUnmarshalWihSingleNestedStruct(t *testing.T) {
	var test testsupport.StructWithSubStruct
	tp := reflect.ValueOf(&test)
	node, err := reflectparser.New(stage, "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(&node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 43, test.Sub.Apa)
	assert.Equal(t, "test svc name", test.Sub.Nu)

	res := node.Value().Interface().(testsupport.StructWithSubStruct)
	assert.Equal(t, "The name", res.Name)
}

func TestMarshalWihSingleStringStruct(t *testing.T) {
	test := testsupport.SingleStringPmsStruct{Name: "my-custom name"}
	tp := reflect.ValueOf(&test)
	node, err := reflectparser.New(stage, "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsr.Upsert(&node, support.NewFilters())
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var tr testsupport.SingleStringPmsStruct
	tp = reflect.ValueOf(&tr)
	node, err = reflectparser.New(stage, "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(&node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "my-custom name", tr.Name)
}

func TestMarshalWihSingleNestedStruct(t *testing.T) {

	test := testsupport.StructWithSubStruct{Name: "kalle kula"}
	test.Sub.Apa = 18
	test.Sub.Nu = "krafso blafso"

	tp := reflect.ValueOf(&test)
	node, err := reflectparser.New(stage, "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsr.Upsert(&node, support.NewFilters())
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	// Read it back
	var tr testsupport.StructWithSubStruct
	tp = reflect.ValueOf(&tr)
	node, err = reflectparser.New(stage, "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(&node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "kalle kula", tr.Name)
	assert.Equal(t, 18, tr.Sub.Apa)
	assert.Equal(t, "krafso blafso", tr.Sub.Nu)
}

//
// LEK O LÄR...
//

type A struct {
	Greeting string
	Message  string
	Pi       float64
}

type B struct {
	Struct    A
	Ptr       *A
	Answer    int
	Map       map[string]string
	StructMap map[string]interface{}
	Slice     []string
}

func TestLek(t *testing.T) {
	// https://golang.org/src/encoding/json/encode.go
	// https://golang.org/src/encoding/json/decode.go
	// https://gist.github.com/hvoecking/10772475#file-translate-go-L191
	// https://medium.com/capital-one-tech/learning-to-use-go-reflection-part-2-c91657395066
	var s B
	dumpStruct(reflect.ValueOf(s))
}

func dumpStruct(v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		switch fv.Kind() {
		case reflect.Ptr:
			// Get the value it points to
			tv := fv.Elem()
			if !tv.IsValid() {
				log.Debug().Msgf("name: %s field %s, type: %s kind: %s, value = nil",
					t.Name(), ft.Name, fv.Type().Name(), fv.Kind().String())
			} else {
				log.Debug().Msgf("name: %s field %s, type: %s kind: %s ptr-to-type: %s ptr-to-kind: %s",
					t.Name(), ft.Name, fv.Type().Name(), fv.Kind().String(), tv.Type().Name(), tv.Kind().String())
			}
			continue
		case reflect.Struct:
			dumpStruct(fv)
		}

		log.Debug().Msgf("name: %s field %s, type: %s kind: %s",
			t.Name(), ft.Name, fv.Type().Name(), fv.Kind().String())
	}
}
