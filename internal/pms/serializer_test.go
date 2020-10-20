package pms

import (
	"flag"
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm/internal/testsupport"
	"github.com/mariotoffia/ssm/parser"
	"github.com/mariotoffia/ssm/support"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

// cSpell:disable
var stage string
var scope string
var provision bool

func init() {
	testing.Init() // Need to do this in order for flag.Parse() to work

	flag.StringVar(&scope, "scope", "", "Scope for test")
	flag.BoolVar(&provision, "provision", true,
		"Set this to false when no provision the ssm with default values shall take place",
	)

	flag.Parse()

	// TODO: just for manual testing
	//provision = false
	//scope = "rw"

	stage = testsupport.UnittestStage()
	log.Info().Msgf("Initializing PMS unittest with STAGE: %s", stage)

	if provision {
		err := testsupport.DefaultProvisionPms(stage)
		if err != nil {
			panic(err)
		}
	}
}

func TestMarshalSecureParamAccountKMSKey(t *testing.T) {
	if scope != "rw" {
		return
	}

	type Test struct {
		MySecret string `pms:"param, keyid=default"`
	}

	test := Test{MySecret: `{"user":"nisse@hult.com", "pass":"kalle", "apikey":"abc-122-abc"}`}
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsRepository, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsRepository.Upsert(node, support.NewFilters())
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var tr Test
	tp = reflect.ValueOf(&tr)

	node, err = parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsRepository.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t,
		`{"user":"nisse@hult.com", "pass":"kalle", "apikey":"abc-122-abc"}`,
		tr.MySecret,
	)
}

func TestMarshalAsJSONWithDefaultKMSAccountKey(t *testing.T) {
	if scope != "rw" {
		return
	}

	type Endpoint struct {
		User     string `json:"user"`
		Password string `json:"pass"`
		APIKey   string `json:"api-key"`
	}
	type Test struct {
		RemoteEndpoint Endpoint `pms:"remote-endpoint, keyid=default"`
	}

	test := Test{
		RemoteEndpoint: Endpoint{
			User:     "nisse@hult.com",
			Password: "kalle",
			APIKey:   "abc-122-abc",
		},
	}

	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	pmsRepository, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsRepository.Upsert(node, support.NewFilters())
	if len(result) > 0 {
		assert.Equal(t, nil, result)
	}

	var tr Test
	tp = reflect.ValueOf(&tr)

	node, err = parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsRepository.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "nisse@hult.com", tr.RemoteEndpoint.User)
	assert.Equal(t, "kalle", tr.RemoteEndpoint.Password)
	assert.Equal(t, "abc-122-abc", tr.RemoteEndpoint.APIKey)
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

	pmsRepository, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsRepository.Get(node, support.NewFilters())
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

	pmsRepository, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsRepository.Get(node, support.NewFilters())
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

	pmsRepository, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	_, err = pmsRepository.Get(node, support.NewFilters())
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

	pmsRepository, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsRepository.Upsert(node, support.NewFilters())
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

	_, err = pmsRepository.Get(node, support.NewFilters())
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

	pmsRepository, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsRepository.Upsert(node, support.NewFilters())
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

	_, err = pmsRepository.Get(node, support.NewFilters())
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

	pmsRepository, err := New("test-service")
	if err != nil {
		assert.Equal(t, nil, err)
	}

	result := pmsRepository.Upsert(node, support.NewFilters())
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

	_, err = pmsRepository.Get(node, support.NewFilters())
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, "My Connection Details", tr.Name)
	assert.Equal(t, "gördis", tr.Connection.User)
	assert.Equal(t, 1088, tr.Connection.Timeout)
	assert.Equal(t, "åaaäs2##!!äöå!#dfmklvmlkBBCH2¤", tr.Connection.Password)
}

// cSpell:enable
