package report

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/asm"
	"github.com/mariotoffia/ssm.git/internal/pms"
	"github.com/mariotoffia/ssm.git/internal/testsupport"
	"github.com/mariotoffia/ssm.git/parser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/stretchr/testify/assert"
)

func TestReportSingleAsmStringNilStruct(t *testing.T) {
	var test testsupport.SingleStringAsmStruct
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", "prod", "").
		RegisterTagParser("asm", asm.NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	parser.DumpNode(node)

	reporter := NewWithTier(ssm.ParameterTierStandard)
	report, buff, err := reporter.RenderReport(node, &support.FieldFilters{}, true)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 1, len(report.Parameters))
	assert.Contains(t, buff, "\"fqname\": \"/prod/simple/test\",")
	assert.Contains(t, buff, `"strkey": ""`)
	assert.Contains(t, buff, "\"type\": \"secrets-manager\"")
	assert.Contains(t, buff, "\"value\": \"\"")
}

func TestReportSingleAsmStringValueStruct(t *testing.T) {
	test := testsupport.SingleStringAsmStruct{Name: "kalle kula"}
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", "prod", "").
		RegisterTagParser("asm", asm.NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	reporter := NewWithTier(ssm.ParameterTierStandard)
	report, buff, err := reporter.RenderReport(node, &support.FieldFilters{}, true)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 1, len(report.Parameters))
	assert.Contains(t, buff, "\"fqname\": \"/prod/simple/test\",")
	assert.Contains(t, buff, `"strkey": ""`)
	assert.Contains(t, buff, "\"type\": \"secrets-manager\"")
	assert.Contains(t, buff, "\"value\": \"kalle kula\"")
}

func TestReportSingleAsmSecretSubJsonStruct(t *testing.T) {
	test := testsupport.MyDbServiceConfigAsm{}
	test.Connection.User = "gördis"
	test.Connection.Timeout = 1088
	test.Connection.Password = "åaaäs2##!!äöå!#dfmklvmlkBBCH2¤"
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", "prod", "").
		RegisterTagParser("asm", asm.NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	reporter := NewWithTier(ssm.ParameterTierStandard)
	report, buff, err := reporter.RenderReport(node, &support.FieldFilters{}, true)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 1, len(report.Parameters))
	assert.Contains(t, buff, `"value": "{\"user\":\"gördis\",\"password\":\"åaaäs2##!!äöå!#dfmklvmlkBBCH2¤\",\"timeout\":1088}"`)
	assert.Contains(t, buff, `"strkey": "password"`)
	assert.Contains(t, buff, `"fqname": "/prod/test-service/bubbibobbo"`)
	fmt.Println(buff)
}

func TestReportSingleAsmSecretSubJsonNilStruct(t *testing.T) {
	var test testsupport.MyDbServiceConfigAsm
	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", "prod", "").
		RegisterTagParser("asm", asm.NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	reporter := NewWithTier(ssm.ParameterTierStandard)
	report, buff, err := reporter.RenderReport(node, &support.FieldFilters{}, true)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 1, len(report.Parameters))
	assert.Contains(t, buff, `"value": "{\"user\":\"\",\"timeout\":0}"`)
	assert.Contains(t, buff, `"strkey": "password"`)
	assert.Contains(t, buff, `"fqname": "/prod/test-service/bubbibobbo"`)
	fmt.Println(buff)
}

func TestReportPostgresSQLTemplateStruct(t *testing.T) {
	test := testsupport.MyContextPostgresSQL{}
	test.DbCtx.DbName = "mydb"
	test.DbCtx.Engine = support.PostgresDBEngine
	test.DbCtx.Host = "pgsql-17.toffia.se"
	test.DbCtx.Username = "gördis"
	test.Settings.BatchSize = 77
	test.Settings.Signer = "mto"

	tp := reflect.ValueOf(&test)

	node, err := parser.New("test-service", "prod", "").
		RegisterTagParser("asm", asm.NewTagParser()).
		RegisterTagParser("pms", pms.NewTagParser()).
		Parse(tp)

	if err != nil {
		assert.Equal(t, nil, err)
	}

	reporter := NewWithTier(ssm.ParameterTierStandard)
	report, buff, err := reporter.RenderReport(node, &support.FieldFilters{}, true)
	if err != nil {
		assert.Equal(t, nil, err)
	}

	assert.Equal(t, 2, len(report.Parameters))
	fmt.Println(buff)
}
