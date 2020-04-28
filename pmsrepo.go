package ssm

import (
	"context"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/tagparser"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type fullNameField struct {
	FullName string
	Field    reflect.StructField
	Value    reflect.Value
}

type pmsRepo struct {
	config  aws.Config
	service string
	tier    ssm.ParameterTier
}

// Allows for change the tier. By default pmsRepo uses
// the standard tier.
func (p *pmsRepo) setTier(tier ssm.ParameterTier) *pmsRepo {
	p.tier = tier
	return p
}

func newPmsFromConfig(config aws.Config, service string) *pmsRepo {
	return &pmsRepo{config: config, service: service,
		tier: ssm.ParameterTierStandard}
}

func newPmsWithRegion(region string, service string) (*pmsRepo, error) {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return &pmsRepo{}, errors.Errorf("Failed to load AWS config %v", err)
	}

	if len(region) > 0 {
		awscfg.Region = region
	}

	return &pmsRepo{config: awscfg, service: service,
		tier: ssm.ParameterTierStandard}, nil
}

func newPms(service string) (*pmsRepo, error) {
	return newPmsWithRegion("", service)
}

func (p *pmsRepo) get(node *ssmNode) (map[string]fullNameField, error) {
	m := map[string]*ssmNode{}
	issecure := p.nodesToParameterMap(node, m)
	paths := p.extractParameters(m)

	params := &ssm.GetParametersInput{
		Names:          paths,
		WithDecryption: aws.Bool(issecure),
	}

	log.Debug().Str("svc", p.service).Str("region", p.config.Region).Msgf("Fetching: %v", params)

	prms, invalid, err := p.getFromAws(params)
	if err != nil {
		return nil, err
	}

	im := map[string]fullNameField{}
	if len(invalid) > 0 {
		for _, name := range invalid {
			if val, ok := m[name]; ok {
				im[name] = fullNameField{FullName: val.tag.FullName(), Field: val.f, Value: val.v}
			} else {
				log.Warn().Str("service", p.service).Msgf("Could not find %s in node map", name)
			}
		}
	}

	err = p.populate(node, prms)
	return im, err
}

func (p *pmsRepo) populate(node *ssmNode, params map[string]ssm.Parameter) error {
	node.EnsureInstance(false)

	if node.HasChildren() {
		for _, n := range node.childs {
			p.populate(&n, params)
		}
		return nil
	}

	if node.tag.SsmType() != tagparser.Pms {
		log.Debug().Msgf("Node %s is not of pms type", node.tag.FullName())
		return nil
	}

	if val, ok := params[node.tag.FullName()]; ok {
		log.Debug().Msgf("setting: %s (%s) val: %s", node.tag.FullName(), *val.Name, *val.Value)
		switch node.v.Kind() {
		case reflect.String:
			node.v.SetString(*val.Value)
		case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8:
			ival, err := strconv.ParseInt(*val.Value, 10, 64)
			if err != nil {
				return errors.Errorf("Config value %s = %s is not a valid integer", *val.Name, *val.Value)
			}
			node.v.SetInt(ival)
		}
	} else {
		log.Debug().Msgf("no value for property name: %s", node.tag.FullName())
	}

	return nil
}

// Invoke get towards aws parameter store
func (p *pmsRepo) getFromAws(params *ssm.GetParametersInput) (map[string]ssm.Parameter, []string, error) {
	client := ssm.New(p.config)
	req := client.GetParametersRequest(params)

	resp, err := req.Send(context.TODO())

	if err != nil {
		return nil, nil, errors.Errorf("Failed fetch pms config entries %+v", params)
	}

	m := map[string]ssm.Parameter{}
	for _, p := range resp.Parameters {
		key := *p.Name
		m[key] = p
	}

	return m, resp.InvalidParameters, nil
}

// Flattern the parameters in order to provide queries against
// the parameter store.
func (p *pmsRepo) extractParameters(paths map[string]*ssmNode) []string {
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
func (p *pmsRepo) nodesToParameterMap(node *ssmNode, paths map[string]*ssmNode) bool {
	issecure := false
	if node.HasChildren() {
		for _, n := range node.childs {
			if p.nodesToParameterMap(&n, paths) {
				issecure = true
			}
		}
	} else {
		if node.tag.SsmType() == tagparser.Pms {
			paths[node.tag.FullName()] = node
			if node.tag.Secure() {
				issecure = true
			}
		}
	}

	return issecure
}
