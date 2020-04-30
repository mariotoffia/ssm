package pms

import (
	"context"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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
		return &Serializer{}, errors.Errorf("Failed to load AWS config %v", err)
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
func (p *Serializer) Get(node reflectparser.SsmNode,
	filter *support.FieldFilters) (map[string]support.FullNameField, error) {

	m := map[string]reflectparser.SsmNode{}
	issecure := p.nodesToParameterMap(node, m, filter)
	// TODO: need to split up into 10 parameters per get
	paths := p.extractParameters(m)

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
	m map[string]reflectparser.SsmNode) map[string]support.FullNameField {

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
func (p *Serializer) populate(node reflectparser.SsmNode, params map[string]ssm.Parameter) error {
	node.EnsureInstance(false)

	if node.HasChildren() {
		for _, n := range node.Children() {
			p.populate(n, params)
		}
		return nil
	}

	if node.Tag().SsmType() != tagparser.Pms {
		log.Debug().Msgf("Node %s is not of pms type", node.Tag().FullName())
		return nil
	}

	if val, ok := params[node.Tag().FullName()]; ok {
		setStructValue(node, val)
	}

	return nil
}

func setStructValue(node reflectparser.SsmNode, val ssm.Parameter) error {

	log.Debug().Msgf("setting: %s (%s) val: %s", node.Tag().FullName(), *val.Name, *val.Value)

	switch node.Value().Kind() {

	case reflect.String:
		node.Value().SetString(*val.Value)

	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
		setStructIntValue(node, val)
	}

	return nil
}

func setStructIntValue(node reflectparser.SsmNode, val ssm.Parameter) error {
	ival, err := strconv.ParseInt(*val.Value, 10, 64)
	if err != nil {
		return errors.Errorf("Config value %s = %s is not a valid integer", *val.Name, *val.Value)
	}
	node.Value().SetInt(ival)
	return nil
}

var cnt int = 0

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
			return nil, nil, errors.Errorf("Failed fetch pms config entries %+v", params)
		}

		if len(resp.Parameters) == 0 && len(resp.InvalidParameters) > 0 {
			time.Sleep(400 * time.Millisecond)
		} else {
			success = true
		}
	}

	log.Debug().Msg("done getfrom aws!")

	m := map[string]ssm.Parameter{}
	for _, p := range resp.Parameters {
		key := *p.Name
		m[key] = p
	}

	return m, resp.InvalidParameters, nil
}

// Flattern the parameters in order to provide queries against
// the parameter store.
func (p *Serializer) extractParameters(paths map[string]reflectparser.SsmNode) []string {
	arr := make([]string, 0, len(paths))
	for key := range paths {
		arr = append(arr, key)
	}

	return arr
}

// Grabs all FullNames on nodes that do have tag set
// in order to get data fom parameter store. Note that
// it chcks for the tag SsmType = pms. The full name is
// the associated with the node itself. This is to gain
// a more accessable structure to seach for nodes.
func (p *Serializer) nodesToParameterMap(node reflectparser.SsmNode,
	paths map[string]reflectparser.SsmNode, filter *support.FieldFilters) bool {
	issecure := false
	if node.HasChildren() {
		for _, n := range node.Children() {
			if p.nodesToParameterMap(n, paths, filter) {
				issecure = true
			}
		}
	} else {
		if node.Tag().SsmType() == tagparser.Pms {
			if filter.IsIncluded(node.FqName()) {
				paths[node.Tag().FullName()] = node
				if node.Tag().Secure() {
					issecure = true
				}
			}
		}
	}

	return issecure
}
