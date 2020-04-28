package ssm

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type pmsRepo struct {
	config  aws.Config
	service string
}

func newPmsFromConfig(config aws.Config, service string) *pmsRepo {
	return &pmsRepo{config: config, service: service}
}

func newPmsWithRegion(region string, service string) (*pmsRepo, error) {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return &pmsRepo{}, errors.Errorf("Failed to load AWS config %v", err)
	}

	if len(region) > 0 {
		awscfg.Region = region
	}

	return &pmsRepo{config: awscfg, service: service}, nil
}

func newPms(service string) (*pmsRepo, error) {
	return newPmsWithRegion("", service)
}

func (p *pmsRepo) get(node *ssmNode) error {
	m := map[string]*ssmNode{}
	issecure := p.nodesToParameterMap(node, m)
	paths := p.extractParameters(m)

	params := &ssm.GetParametersInput{
		Names:          paths,
		WithDecryption: aws.Bool(issecure),
	}

	log.Debug().Str("svc", p.service).Str("region", p.config.Region).Msgf("Fetching: %v", params)

	prms, invalid, err := p.getFromAws(params)

	if len(invalid) > 0 {
		log.Debug().Str("service", p.service).Msgf("Invalid Parameter(s): %v", invalid)
		for _, p := range invalid {
			// TODO: this should be handeled
			fmt.Printf("%s\n", p)
		}
	}

	p.populate(node, prms)
	return err
}

func (p *pmsRepo) populate(node *ssmNode, params map[string]ssm.Parameter) {
	node.EnsureInstance(false)

	if node.HasChildren() {
		for _, n := range node.childs {
			p.populate(&n, params)
		}
		return
	}

	if val, ok := params[node.tag.FullName()]; ok {
		log.Debug().Msgf("name: %s (%s) val: %s", node.tag.FullName(), *val.Name, *val.Value)
		switch node.v.Kind() {
		case reflect.String:
			node.v.SetString(*val.Value)
		}
	} else {
		log.Debug().Msgf("no value for property name: %s", node.tag.FullName())
	}
}

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

func (p *pmsRepo) extractParameters(paths map[string]*ssmNode) []string {
	arr := make([]string, 0, len(paths))
	for key := range paths {
		arr = append(arr, key)
	}

	return arr
}

func (p *pmsRepo) nodesToParameterMap(node *ssmNode, paths map[string]*ssmNode) bool {
	issecure := false
	if node.HasChildren() {
		for _, n := range node.childs {
			if p.nodesToParameterMap(&n, paths) {
				issecure = true
			}
		}
	} else {
		paths[node.tag.FullName()] = node
		if node.tag.Secure() {
			issecure = true
		}
	}

	return issecure
}
