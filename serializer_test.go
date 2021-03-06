package ssm

import (
	"flag"
	"fmt"
	"testing"

	"github.com/mariotoffia/ssm/internal/testsupport"
	"github.com/mariotoffia/ssm/support"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

var stage string
var scope string

func init() {
	testing.Init() // Need to do this in order for flag.Parse() to work
	flag.StringVar(&scope, "scope", "", "Scope for test")
	flag.Parse()

	if scope != "clean" {

		stage = testsupport.DefaultProvisionAsm()
		log.Info().Msgf("Initializing main serializer unittest with STAGE: %s", stage)

		err := testsupport.DefaultProvisionPms(stage)
		if err != nil {
			panic(err)
		}
	}
}

func TestCleanAll(t *testing.T) {
	if scope != "clean" {
		t.Skip("Only run when explicit run with parameter clean")
	}

	testsupport.DeleteAllUnittestSecrets()
	testsupport.ListDeletePrms()
}

func TestReportNestedStructValues(t *testing.T) {
	set := testsupport.StructWithSubStruct{Name: "Thy name"}
	set.Sub.Apa = 44
	set.Sub.Nu = "hibby bibby"
	set.AsmSub.Apa2 = 444
	set.AsmSub.Nu2 = "ingen fantasi"

	s := NewSsmSerializer(stage, "test-service")
	objs, json, err := s.ReportWithOpts(&set, NoFilter, true)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Contains(t, json, "hibby bibby")
	assert.Equal(t, 5, len(objs.Parameters))
}

func TestUnmarshalWihSingleStringStructPms(t *testing.T) {
	var test testsupport.SingleStringPmsStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.Unmarshal(&test)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
}

func TestUnmarshalWihSingleStringStructAsm(t *testing.T) {
	var test testsupport.SingleStringAsmStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.Unmarshal(&test)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "The name", test.Name)
}

func TestUnmarshalWihSingleNestedStructPms(t *testing.T) {
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

func TestUnmarshalWihSingleNestedStructAsm(t *testing.T) {
	var test testsupport.StructWithSubStruct

	s := NewSsmSerializer(stage, "test-service")
	_, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyAsm)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 43, test.AsmSub.Apa2)
	assert.Equal(t, "test svc name", test.AsmSub.Nu2)
}

func TestUnmarshalWihSingleNestedStructPmsAndAsm(t *testing.T) {
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

func TestUnmarshalWhenOnlyAsmEnabledPmsWillNotBePopulated(t *testing.T) {
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

func TestUnmarshalWhenOnlyPmsEnabledAsmWillNotBePopulated(t *testing.T) {
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

func TestUnmarshalWihSingleNestedStructFilteredPms(t *testing.T) {
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

func TestUnmarshalWihSingleNestedStructFilteredAsm(t *testing.T) {
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

func TestUnmarshalNonBackedVariableInStructReturnsAsMissingFullNameFieldPms(t *testing.T) {
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

func TestUnmarshalNonBackedVariableInStructReturnsAsMissingFullNameFieldAsm(t *testing.T) {
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

func TestMarshalWihSingleStringStructPms(t *testing.T) {
	if scope != "rw" {
		return
	}

	test := testsupport.SingleStringPmsStruct{Name: "stored from ssm"}

	s := NewSsmSerializer(stage, "test-service")
	errors := s.Marshal(&test)
	if len(errors) > 0 {
		assert.Equal(t, nil, errors)
	}

	var testr testsupport.SingleStringPmsStruct
	_, err := s.Unmarshal(&testr)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "stored from ssm", testr.Name)
}

func TestMarshalWihSingleStringStructAsm(t *testing.T) {
	if scope != "rw" {
		return
	}

	set := testsupport.SingleStringAsmStruct{Name: "hobby bobby"}

	s := NewSsmSerializer(stage, "test-service")
	result := s.Marshal(&set)
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var test testsupport.SingleStringAsmStruct
	_, err := s.Unmarshal(&test)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "hobby bobby", test.Name)
}

func TestMarshalWihSingleNestedStructPms(t *testing.T) {
	if scope != "rw" {
		return
	}

	set := testsupport.StructWithSubStruct{Name: "nisse hult"}
	set.Sub.Apa = 88
	set.Sub.Nu = "bubben här"

	s := NewSsmSerializer(stage, "test-service")
	result := s.MarshalWithOpts(&set, NoFilter, OnlyPms)
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var test testsupport.StructWithSubStruct
	_, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyPms)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "nisse hult", test.Name)
	assert.Equal(t, 88, test.Sub.Apa)
	assert.Equal(t, "bubben här", test.Sub.Nu)
}

func TestMarshalWihSingleNestedStructAsm(t *testing.T) {
	if scope != "rw" {
		return
	}

	set := testsupport.StructWithSubStruct{}
	set.AsmSub.Apa2 = 188
	set.AsmSub.Nu2 = "bubben här igen"

	s := NewSsmSerializer(stage, "test-service")

	result := s.MarshalWithOpts(&set, NoFilter, OnlyAsm)
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var test testsupport.StructWithSubStruct
	_, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyAsm)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 188, test.AsmSub.Apa2)
	assert.Equal(t, "bubben här igen", test.AsmSub.Nu2)
}

func TestMarshalWihSingleNestedStructPmsAndAsm(t *testing.T) {
	if scope != "rw" {
		return
	}

	set := testsupport.StructWithSubStruct{Name: "Thy name"}
	set.Sub.Apa = 44
	set.Sub.Nu = "hibby bibby"
	set.AsmSub.Apa2 = 444
	set.AsmSub.Nu2 = "ingen fantasi"

	s := NewSsmSerializer(stage, "test-service")
	result := s.MarshalWithOpts(&set, NoFilter, AllTags)
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var test testsupport.StructWithSubStruct
	_, err := s.UnmarshalWithOpts(&test, NoFilter, AllTags)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "Thy name", test.Name)
	assert.Equal(t, 44, test.Sub.Apa)
	assert.Equal(t, "hibby bibby", test.Sub.Nu)
	assert.Equal(t, 444, test.AsmSub.Apa2)
	assert.Equal(t, "ingen fantasi", test.AsmSub.Nu2)
}

func TestMarshalWihSingleNestedStructFilteredPms(t *testing.T) {
	if scope != "rw" {
		return
	}

	test := testsupport.StructWithSubStruct{Name: "hej o hå"}
	test.Sub.Apa = 999
	test.Sub.Nu = "johoo"

	s := NewSsmSerializer(stage, "test-service")
	result := s.MarshalWithOpts(&test,
		support.NewFilters().
			Exclude("Sub.Apa"), OnlyPms)

	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var testr testsupport.StructWithSubStruct
	_, err := s.UnmarshalWithOpts(&testr, NoFilter, OnlyPms)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "hej o hå", testr.Name)
	assert.NotEqual(t, 999, testr.Sub.Apa) // Since not included
	assert.Equal(t, "johoo", testr.Sub.Nu)
}

func TestMarshalWihSingleNestedStructFilteredAsm(t *testing.T) {
	if scope != "rw" {
		return
	}

	set := testsupport.StructWithSubStruct{}
	set.AsmSub.Apa2 = 999
	set.AsmSub.Nu2 = "japp"

	s := NewSsmSerializer(stage, "test-service")
	result := s.MarshalWithOpts(&set,
		support.NewFilters().
			Exclude("AsmSub.Apa2"), OnlyAsm)

	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var test testsupport.StructWithSubStruct
	_, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyAsm)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.NotEqual(t, 999, test.AsmSub.Apa2) // Since not included
	assert.Equal(t, "japp", test.AsmSub.Nu2)
}

func TestDeletePmsTaggedStruct(t *testing.T) {
	if scope != "rw" {
		return
	}

	type Test struct {
		Name string `pms:"test, prefix=simple"`
		Sub  struct {
			Apa int    `pms:"ext"`
			Nu  string `pms:"myname"`
		}
		AsmSub struct {
			Apa2 int    `asm:"ext"`
			Nu2  string `asm:"myname"`
		}
	}

	set := Test{Name: "nisse manpower"}
	set.Sub.Apa = 88
	set.Sub.Nu = "bubben här"
	set.AsmSub.Apa2 = 99
	set.AsmSub.Nu2 = "doris simmar"

	s := NewSsmSerializer(stage, "test-service")
	result := s.MarshalWithOpts(&set, NoFilter, OnlyPms)
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var test Test
	_, err := s.UnmarshalWithOpts(&test, NoFilter, OnlyPms)
	assert.Equal(t, nil, err)

	assert.Equal(t, "nisse manpower", test.Name)
	assert.Equal(t, 88, test.Sub.Apa)
	assert.Equal(t, "bubben här", test.Sub.Nu)

	var test2 Test
	fields, _ := s.DeleteWithOpts(&test2, NoFilter, OnlyPms)

	assert.Equal(t, 0, len(fields), "all fields deleted")

	var test3 Test
	fields, _ = s.UnmarshalWithOpts(&test3, NoFilter, OnlyPms)
	assert.Equal(t, 3, len(fields), "we should get three errors")
}

func TestGetWithIncorrectPrefix(t *testing.T) {
	if scope != "rw" {
		return
	}

	type Test struct {
		Name string `pms:"test"`
		Sub  struct {
			Apa int    `pms:"ext"`
			Nu  string `pms:"myname"`
		}
	}

	set := Test{Name: "nisse manpower"}
	set.Sub.Apa = 88
	set.Sub.Nu = "bubben här"

	s := NewSsmSerializer(stage, "test-service").
		UsePrefix("/global/endeavor")

	result := s.Marshal(&set)
	if len(result) > 0 {

		assert.Equal(t, 0, len(result),
			"should not return any fields, since this indicates error %v", result,
		)

		return
	}

	defer func() {

		var test2 Test

		fields, _ := s.UsePrefix("/global/endeavor").
			Delete(&test2)

		assert.Equal(t, 0, len(fields), "You need to manually delete %v", fields)

	}()

	var test Test

	fields2, err := s.UsePrefix("/global/endeavor2").
		Unmarshal(&test)

	assert.Equal(t, nil, err)

	assert.Equal(t, 3, len(fields2),
		"Since returning no fields indicates that is could read and that's wrong",
	)
}
