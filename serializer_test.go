package ssm

import (
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/mariotoffia/ssm.git/support"
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

func TestWihSingleNestedStructFiltered(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer("eap", "test-service")
	_, err := s.UnmarshalFilterd(&test,
		support.NewFilters().
			Exclude("Sub.Apa"))

	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 0, test.Sub.Apa) // Since not included
	assert.Equal(t, "test svc name", test.Sub.Nu)
}

func TestNonBackedVariableInStructReturnsAsMissingFullNameField(t *testing.T) {
	var test testsupport.StructWithNonExistantVariable

	s := NewSsmSerializer("eap", "test-service")
	invalid, err := s.Unmarshal(&test)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 43, test.Sub.Apa)
	assert.Equal(t, "test svc name", test.Sub.Nu)
	assert.Equal(t, "", test.Sub.Missing)
	assert.Equal(t, 1, len(invalid))
	assert.Equal(t, "Sub.Missing", invalid["Sub.Missing"].LocalName)
	assert.Equal(t, "/eap/test-service/sub/gonemissing", invalid["Sub.Missing"].RemoteName)
}
