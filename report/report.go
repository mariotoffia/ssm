package report

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/asm"
	"github.com/mariotoffia/ssm.git/internal/common"
	"github.com/mariotoffia/ssm.git/internal/pms"
	"github.com/mariotoffia/ssm.git/parser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/rs/zerolog/log"
)

// ParameterType specifies if a parameter is a secrets
// maanger or parameter store parameter
type ParameterType string

const (
	// SecretsManager specifies that a parameter is a secrets manager parameter
	SecretsManager ParameterType = "secrets-manager"
	// ParameterStore specifies that a parameter is a parameter store parameter
	ParameterStore ParameterType = "parameter-store"
)

// Report contains the complete report
type Report struct {
	Parameters []Parameter `json:"parameters"`
}

// Parameter specifies a parameter configuration
type Parameter struct {
	// Type specifies the parameter type
	Type ParameterType `json:"type"`
	// Name specifies the fully qualified name of the parameter
	Name string `json:"fqname"`
	// KeyID, if set it is a ARN to a key ID to be used when encrypting / decrypting th key
	KeyID string `json:"keyid"`
	// Description describes the key
	Description string `json:"description"`
	// Tags is a key value that are applied to a parameter or secret
	Tags map[string]string `json:"tags"`
	// Details is an optional property to specify details that is exclusive for the ParameterType
	// the ParameterStore do have PsmParameterDetails as mandatory.
	Details interface{} `json:"details"`
	// Value the value of this parameter
	Value string `json:"value"`
	// The type of the value. For example PMS has String|StringList|SecureString.
	ValueType string `json:"valuetype"`
}

// PmsParameterDetails describes the details for a parameter store parameter
type PmsParameterDetails struct {
	// Pattern is a regexp to validate against (optional)
	Pattern string `json:"pattern"`
	// Tier specifies the tier for the parameter
	Tier ssm.ParameterTier `json:"tier"`
}

// AsmParameterDetails specifies Secrets Manager secret specifics
type AsmParameterDetails struct {
	// StringKey is the name of the json key where the Secrets Manager shall genereate it's secret into.
	// This is for template driven secrets where a JSON payload is set into the SecretString
	StringKey string `json:"strkey"`
}

// Reporter is the type to produce report of the configuration
type Reporter struct {
	tier ssm.ParameterTier
}

// New creates a new Reporter with ssm.ParameterTierStandard
func New() *Reporter {
	return NewWithTier(ssm.ParameterTierStandard)
}

// NewWithTier creates a new Reporter with specified tier
func NewWithTier(tier ssm.ParameterTier) *Reporter {
	return &Reporter{tier: tier}
}

// RenderReport renders a JSON report based on the node tree and filters
func (r *Reporter) RenderReport(node *parser.StructNode,
	filter *support.FieldFilters, value bool) (*Report, string, error) {

	params := []Parameter{}
	params = r.renderReport(node, filter, params, value)

	report := &Report{Parameters: params}
	buff, err := json.MarshalIndent(report, "", "  ")

	return report, string(buff), err
}

func (r *Reporter) renderReport(node *parser.StructNode,
	filter *support.FieldFilters, params []Parameter, value bool) []Parameter {

	var prm *Parameter
	if filter.IsIncluded(node.FqName) {
		if pmstag, ok := pms.ToPmsTag(node); ok {
			prm = r.handlePmsTag(pmstag)
		} else if asmtag, ok := asm.ToAsmTag(node); ok {
			prm = r.handleAsmTag(asmtag)
		} else {
			log.Debug().Msgf("node %s has not pms or asm tag", node.FqName)
		}

		if prm != nil {
			if value {
				prm.Value = common.GetStringValueFromField(node)
			}

			params = append(params, *prm)
		}
	}

	if node.HasChildren() {
		if prm != nil {
			prm.Value = common.GetStringValueFromField(node)
		} else {
			children := node.Childs
			for i := range node.Childs {
				params = r.renderReport(&children[i], filter, params, value)
			}
		}
	}

	return params
}

func (r *Reporter) handleAsmTag(asmtag *asm.AsmTagStruct) *Parameter {
	prm := &Parameter{
		Name:        asmtag.GetFullName(),
		Description: asmtag.Description(),
		Tags:        asmtag.GetTags(),
		Type:        ParameterStore,
	}
	prm.Type = SecretsManager
	prm.ValueType = "SecureString"
	prm.Details = AsmParameterDetails{
		StringKey: asmtag.StringKey(),
	}

	if asmtag.IsLocalKey() {
		// TODO: need to resolve it to an ARN
	} else if !asmtag.DefaultAccountKey() {
		// Key is ARN
		prm.KeyID = asmtag.GetKeyName()
	}
	return prm
}

func (r *Reporter) handlePmsTag(pmstag *pms.PmsTagStruct) *Parameter {
	prm := &Parameter{
		Name:        pmstag.GetFullName(),
		Description: pmstag.Description(),
		Tags:        pmstag.GetTags(),
		Type:        ParameterStore,
	}

	prm.Details = PmsParameterDetails{
		Pattern: pmstag.Pattern(),
		Tier:    pmstag.SsmTier(r.tier),
	}

	if pmstag.IsLocalKey() {
		// TODO: need to resolve it to an ARN
	} else if !pmstag.DefaultAccountKey() {
		// Key is ARN
		prm.KeyID = pmstag.GetKeyName()
	}

	if pmstag.Secure() {
		prm.ValueType = "SecureString"
	} else {
		prm.ValueType = "String"
	}

	return prm
}
