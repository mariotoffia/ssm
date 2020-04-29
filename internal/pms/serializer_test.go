package pms

import (
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestWihSingleStringStruct(t *testing.T) {
	var test testsupport.SingleStringStruct
	tp := reflect.ValueOf(&test)
	node, err := reflectparser.New("eap", "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(&node)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	res := node.Value().Interface().(testsupport.SingleStringStruct)
	assert.Equal(t, "The name", res.Name)
}

func TestWihSingleNestedStruct(t *testing.T) {
	var test testsupport.StructWithSubStruct
	tp := reflect.ValueOf(&test)
	node, err := reflectparser.New("eap", "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	nnn := []reflectparser.SsmNode{}
	reflectparser.DumpNodes(append(nnn, node))

	pmsr, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsr.Get(&node)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 43, test.Sub.Apa)
	assert.Equal(t, "test svc name", test.Sub.Nu)

	res := node.Value().Interface().(testsupport.StructWithSubStruct)
	assert.Equal(t, "The name", res.Name)
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
