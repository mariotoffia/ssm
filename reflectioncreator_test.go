package ssm

import (
	"reflect"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

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

func TestSingleStringStructCreate(t *testing.T) {
	var test SingleStringStruct
	tp := reflect.ValueOf(test)
	node, err := newReflectionParser("dev", "test-service").parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	v, err := create(&node)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	nodes := []ssmNode{}
	dumpNodes(append(nodes, node))
	s := v.Interface().(SingleStringStruct)
	s.Name = "pelle"
	log.Debug().Msgf("TestSingleStringStructCreate: s.Name = %s", s.Name)
}

func TestStructWithSubStructCreate(t *testing.T) {
	var test StructWithSubStruct
	tp := reflect.ValueOf(test)
	node, err := newReflectionParser("dev", "test-service").parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	v, err := create(&node)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	nodes := []ssmNode{}
	dumpNodes(append(nodes, node))
	s := v.Interface().(StructWithSubStruct)
	s.Name = "pelle"
	s.Sub.Apa = 17
	s.Sub.Nu = "test"
	log.Debug().Msgf("TestStructWithSubStructCreate name %s sub.apa %d sub.nu %s", s.Name, s.Sub.Apa, s.Sub.Nu)
}
