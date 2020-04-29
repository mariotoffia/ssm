package ssm

import (
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/stretchr/testify/assert"
)

func TestWihSingleStringStruct(t *testing.T) {
	var test testsupport.SingleStringStruct

	s := NewSsmSerializer("eap", "test-service")
	err := s.Unmarshal(&test)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
}
