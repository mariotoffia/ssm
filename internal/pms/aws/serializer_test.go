package awspms

import (
	"reflect"
	"testing"

	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/mariotoffia/ssm.git/parser"
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
	node, err := parser.New("test-service", stage, "").
		RegisterTagParser("pms", NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	parser.DumpNode(node)

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
