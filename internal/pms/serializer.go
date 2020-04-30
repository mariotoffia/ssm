package pms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/common"
	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/tagparser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Serializer handles the parameter store communication
type Serializer struct {
	config  aws.Config
	service string
	tier    ssm.ParameterTier
}

// SetTier allows for change the tier. By default PmsRepo uses
// the standard tier.
func (p *Serializer) SetTier(tier ssm.ParameterTier) *Serializer {
	p.tier = tier
	return p
}

// NewFromConfig creates a repository using a existing configuration
func NewFromConfig(config aws.Config, service string) *Serializer {
	return &Serializer{config: config, service: service,
		tier: ssm.ParameterTierStandard}
}

// NewWithRegion creates a repository using the default configuration
// with optional region override.
func NewWithRegion(region string, service string) (*Serializer, error) {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return &Serializer{}, errors.Wrapf(err, "Failed to load AWS config")
	}

	if len(region) > 0 {
		awscfg.Region = region
	}

	return &Serializer{config: awscfg, service: service,
		tier: ssm.ParameterTierStandard}, nil
}

// New creates a repository using the default configuration.
func New(service string) (*Serializer, error) {
	return NewWithRegion("", service)
}

// Get parameters from the parameterstore and populates the node graph with values.
// Any fields that was not able to be set is reported in the FullNameField string map.
// FullNameField do not include those fields filtered out in exclusion filter.
func (p *Serializer) Get(node *reflectparser.SsmNode,
	filter *support.FieldFilters) (map[string]support.FullNameField, error) {

	m := map[string]*reflectparser.SsmNode{}
	issecure := common.NodesToParameterMap(node, m, filter, tagparser.Pms)
	paths := common.ExtractParameters(m)

	params := &ssm.GetParametersInput{
		Names:          paths,
		WithDecryption: aws.Bool(issecure),
	}

	log.Debug().Str("svc", p.service).Msgf("Fetching: %v", params)

	prms, invalid, err := p.getFromAws(params)
	if err != nil {
		return nil, err
	}

	im := p.handleInvalidRequestParameters(invalid, m)
	err = p.populate(node, prms)

	return im, err
}

func (p *Serializer) handleInvalidRequestParameters(invalid []string,
	m map[string]*reflectparser.SsmNode) map[string]support.FullNameField {

	im := map[string]support.FullNameField{}

	if len(invalid) > 0 {
		for _, name := range invalid {
			if val, ok := m[name]; ok {
				im[val.FqName()] = support.FullNameField{RemoteName: val.Tag().FullName(),
					LocalName: val.FqName(), Field: val.Field(), Value: val.Value()}
			} else {
				log.Warn().Str("service", p.service).Msgf("Could not find %s in node map", name)
			}
		}
	}

	if len(im) > 0 {
		for key, val := range im {
			log.Debug().Msgf("not fetched: %s [%s]", key, val.RemoteName)
		}
	}
	return im
}
func (p *Serializer) populate(node *reflectparser.SsmNode, params map[string]ssm.Parameter) error {
	node.EnsureInstance(false)

	if node.HasChildren() {
		for _, n := range node.Children() {
			p.populate(&n, params)
		}
		return nil
	}

	if node.Tag().SsmType() != tagparser.Pms {
		log.Debug().Msgf("Node %s is not of pms type", node.Tag().FullName())
		return nil
	}

	if val, ok := params[node.Tag().FullName()]; ok {
		common.SetStructValueFromString(node, *val.Name, *val.Value)
	}

	return nil
}

// Invoke get towards aws parameter store
func (p *Serializer) getFromAws(params *ssm.GetParametersInput) (map[string]ssm.Parameter, []string, error) {
	client := ssm.New(p.config)

	var resp *ssm.GetParametersResponse
	var err error
	success := false
	for i := 0; i < 3 && !success; i++ {
		req := client.GetParametersRequest(params)

		resp, err = req.Send(context.TODO())
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Failed fetch pms config entries %v", params)
		}

		if len(resp.Parameters) == 0 && len(resp.InvalidParameters) > 0 {
			time.Sleep(400 * time.Millisecond)
		} else {
			success = true
		}
	}

	m := map[string]ssm.Parameter{}
	for _, p := range resp.Parameters {
		key := *p.Name
		m[key] = p
	}

	return m, resp.InvalidParameters, nil
}
