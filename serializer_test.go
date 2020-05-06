package ssm

import (
	"fmt"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

var stage string

func init() {
	stage = testsupport.DefaultProvisionAsm()
	log.Info().Msgf("Initializing main serializer unittest with STAGE: %s", stage)

	err := testsupport.DefaultProvisionPms(stage)
	if err != nil {
		panic(err)
	}
}

func TestWihSingleStringStructPms(t *testing.T) {
	var test testsupport.SingleStringPmsStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.Unmarshal(&test)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
}

func TestWihSingleStringStructAsm(t *testing.T) {
	var test testsupport.SingleStringAsmStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.Unmarshal(&test)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
}

func TestWihSingleNestedStructPms(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyPms)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 43, test.Sub.Apa)
	assert.Equal(t, "test svc name", test.Sub.Nu)
}

func TestWihSingleNestedStructAsm(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyAsm)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 43, test.AsmSub.Apa2)
	assert.Equal(t, "test svc name", test.AsmSub.Nu2)
}

func TestWihSingleNestedStructPmsAndAsm(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.UnmarshalWithOpts(&test, NoFilter, AllTags)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 43, test.Sub.Apa)
	assert.Equal(t, "test svc name", test.Sub.Nu)
	assert.Equal(t, 43, test.AsmSub.Apa2)
	assert.Equal(t, "test svc name", test.AsmSub.Nu2)
}

func TestWhenOnlyAsmEnabledPmsWillNotBePopulated(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyAsm)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "", test.Name)
	assert.Equal(t, 0, test.Sub.Apa)
	assert.Equal(t, "", test.Sub.Nu)
	assert.Equal(t, 43, test.AsmSub.Apa2)
	assert.Equal(t, "test svc name", test.AsmSub.Nu2)
}

func TestWhenOnlyPmsEnabledAsmWillNotBePopulated(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyAsm)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "", test.Name)
	assert.Equal(t, 0, test.Sub.Apa)
	assert.Equal(t, "", test.Sub.Nu)
	assert.Equal(t, 43, test.AsmSub.Apa2)
	assert.Equal(t, "test svc name", test.AsmSub.Nu2)
}

func TestWihSingleNestedStructFilteredPms(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.UnmarshalWithOpts(&test,
		support.NewFilters().
			Exclude("Sub.Apa"), OnlyPms)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 0, test.Sub.Apa) // Since not included
	assert.Equal(t, "test svc name", test.Sub.Nu)
}

func TestWihSingleNestedStructFilteredAsm(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.UnmarshalWithOpts(&test,
		support.NewFilters().
			Exclude("AsmSub.Apa2"), OnlyAsm)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 0, test.AsmSub.Apa2) // Since not included
	assert.Equal(t, "test svc name", test.AsmSub.Nu2)
}

func TestNonBackedVariableInStructReturnsAsMissingFullNameFieldPms(t *testing.T) {
	var test testsupport.StructPmsWithNonExistantVariable

	s := NewSsmSerializer(stage, "test-service")
	invalid, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyPms)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
	assert.Equal(t, 43, test.Sub.Apa)
	assert.Equal(t, "test svc name", test.Sub.Nu)
	assert.Equal(t, "", test.Sub.Missing)
	assert.Equal(t, 1, len(invalid))
	assert.Equal(t, "Sub.Missing", invalid["Sub.Missing"].LocalName)
	assert.Equal(t, fmt.Sprintf("/%s/test-service/sub/gonemissing", stage), invalid["Sub.Missing"].RemoteName)
}

func TestNonBackedVariableInStructReturnsAsMissingFullNameFieldAsm(t *testing.T) {
	var test testsupport.StructPmsWithNonExistantVariable

	s := NewSsmSerializer(stage, "test-service")
	invalid, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyAsm)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 43, test.AsmSub.Apa2)
	assert.Equal(t, "test svc name", test.AsmSub.Nu2)
	assert.Equal(t, "", test.AsmSub.Missing2)
	assert.Equal(t, 1, len(invalid))
	assert.Equal(t, "AsmSub.Missing2", invalid["AsmSub.Missing2"].LocalName)
	assert.Equal(t, fmt.Sprintf("/%s/test-service/asmsub/gonemissing", stage), invalid["AsmSub.Missing2"].RemoteName)
}
