package report

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/common"
	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/tagparser"
	"github.com/mariotoffia/ssm.git/support"
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
}

// PmsParameterDetails describes the details for a parameter store parameter
type PmsParameterDetails struct {
	// Pattern is a regexp to validate against (optional)
	Pattern string `json:"pattern"`
	// Tier specifies the tier for the parameter
	Tier ssm.ParameterTier `json:"tier"`
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
func (r *Reporter) RenderReport(node *reflectparser.SsmNode,
	filter *support.FieldFilters, value bool) (*Report, string, error) {

	params := []Parameter{}
	params = r.renderReport(node, filter, params, value)

	report := &Report{Parameters: params}
	buff, err := json.MarshalIndent(report, "", "  ")

	return report, string(buff), err
}

func (r *Reporter) renderReport(node *reflectparser.SsmNode,
	filter *support.FieldFilters, params []Parameter, value bool) []Parameter {

	if node.HasChildren() {
		children := node.Children()
		for i := range node.Children() {
			params = r.renderReport(&children[i], filter, params, value)
		}
	} else {
		if filter.IsIncluded(node.FqName()) {

			var keyid string
			if node.Tag().IsLocalKey() {
				// TODO: need to resolve it to an ARN
			} else if !node.Tag().DefaultAccountKey() {
				// Key is ARN
				keyid = node.Tag().GetKeyName()
			}

			prm := Parameter{
				Name:        node.Tag().FullName(),
				Description: node.Tag().Description(),
				Tags:        node.Tag().Tags(),
				KeyID:       keyid,
			}

			if node.Tag().SsmType() == tagparser.Pms {
				if tag, ok := node.Tag().(*tagparser.PmsTag); ok {

					prm.Details = PmsParameterDetails{
						Pattern: tag.Pattern(),
						Tier:    r.getTierFromTag(tag),
					}

					prm.Type = ParameterStore
				}
			} else {
				prm.Type = SecretsManager
			}

			if value {
				prm.Value = common.GetStringValueFromField(node)
			}

			params = append(params, prm)
		}
	}

	return params
}

func (r *Reporter) getTierFromTag(pmstag *tagparser.PmsTag) ssm.ParameterTier {

	switch pmstag.Tier() {
	case tagparser.Default:
		return r.tier
	case tagparser.Std:
		return ssm.ParameterTierStandard
	case tagparser.Adv:
		return ssm.ParameterTierAdvanced
	case tagparser.Eval:
		return ssm.ParameterTierIntelligentTiering
	}

	return r.tier
}
