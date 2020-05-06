package report

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/stretchr/testify/assert"
)

func TestReportSingleAsmStringNilStruct(t *testing.T) {
	var test testsupport.SingleStringAsmStruct
	tp := reflect.ValueOf(&test)
	node, err := reflectparser.New("prod", "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	reporter := NewWithTier(ssm.ParameterTierStandard)
	report, buff, err := reporter.RenderReport(&node, &support.FieldFilters{}, true)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 1, len(report.Parameters))
	assert.Contains(t, buff, "\"fqname\": \"/prod/simple/test\",")
	assert.Contains(t, buff, "\"details\": null")
	assert.Contains(t, buff, "\"type\": \"secrets-manager\"")
	assert.Contains(t, buff, "\"value\": \"\"")
}

func TestReportSingleAsmStringValueStruct(t *testing.T) {
	test := testsupport.SingleStringAsmStruct{Name: "kalle kula"}
	tp := reflect.ValueOf(&test)
	node, err := reflectparser.New("prod", "test-service").Parse("", tp)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	reporter := NewWithTier(ssm.ParameterTierStandard)
	report, buff, err := reporter.RenderReport(&node, &support.FieldFilters{}, true)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 1, len(report.Parameters))
	assert.Contains(t, buff, "\"fqname\": \"/prod/simple/test\",")
	assert.Contains(t, buff, "\"details\": null")
	assert.Contains(t, buff, "\"type\": \"secrets-manager\"")
	assert.Contains(t, buff, "\"value\": \"kalle kula\"")
}