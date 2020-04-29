package ssm

import (
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestWihSingleStringStruct(t *testing.T) {
	var test testsupport.SingleStringStruct

	s := NewSsmSerializer("eap", "test-service")
	_, err := s.Unmarshal(&test)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
}

func TestWihSingleNestedStruct(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer("eap", "test-service")
	_, err := s.Unmarshal(&test)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 43, test.Sub.Apa)
	assert.Equal(t, "test svc name", test.Sub.Nu)
}
