package asm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/mariotoffia/ssm.git/internal/common"
	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/internal/tagparser"
	"github.com/mariotoffia/ssm.git/support"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Serializer handles the secrets manager communication
type Serializer struct {
	config  aws.Config
	service string
}

// NewFromConfig creates a repository using a existing configuration
func NewFromConfig(config aws.Config, service string) *Serializer {
	return &Serializer{config: config, service: service}
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

	return &Serializer{config: awscfg, service: service}, nil
}

// New creates a repository using the default configuration.
func New(service string) (*Serializer, error) {
	return NewWithRegion("", service)
}

// Get parameters from the secrets manager and populates the node graph with values.
// Any fields that was not able to be set is reported in the FullNameField string map.
// FullNameField do not include those fields filtered out in exclusion filter.
func (p *Serializer) Get(node *reflectparser.SsmNode,
	filter *support.FieldFilters) (map[string]support.FullNameField, error) {

	m := map[string]*reflectparser.SsmNode{}
	common.NodesToParameterMap(node, m, filter, tagparser.Asm)

	for key, val := range m {
		log.Info().Msgf("m[%s] = '%v'", key, *val)
	}

	mprms := map[string]*secretsmanager.GetSecretValueOutput{}
	im := map[string]support.FullNameField{}

	for _, prm := range common.ExtractParameters(m) {
		if n, ok := m[prm]; ok {

			if nasm, ok := n.Tag().(*tagparser.AsmTag); ok {
				result, err := p.getFromAws(prm, nasm)

				if err != nil {
					if aerr, ok := err.(awserr.Error); ok {
						switch aerr.Code() {
						case secretsmanager.ErrCodeResourceNotFoundException:
							im[nasm.FullName()] = support.FullNameField{LocalName: nasm.FullName(),
								RemoteName: prm, Field: node.Field(), Value: node.Value()}
						default:
							return nil, errors.Wrapf(err, "Failed fetch asm config entry %s", prm)
						}
					}
				} else {
					log.Debug().Str("svc", p.service).Msgf("mprms[%s] = %v", n.FqName(), result)
					mprms[n.FqName()] = result
				}
			} else {
				log.Warn().Str("svc", p.service).Msgf("tag is not asm tag! tag: %v", n)
			}
		}
	}

	populate(node, mprms)

	return im, nil
}

func populate(node *reflectparser.SsmNode, params map[string]*secretsmanager.GetSecretValueOutput) {
	node.EnsureInstance(false)

	if node.HasChildren() {
		for _, n := range node.Children() {
			populate(&n, params)
		}
		return
	}

	if node.Tag().SsmType() != tagparser.Asm {
		log.Debug().Msgf("Node %s is not of asm type", node.FqName())
		return
	}

	if val, ok := params[node.FqName()]; ok {
		common.SetStructValueFromString(node, *val.Name, *val.SecretString)
	} else {
		log.Warn().Msgf("could not find parameter %s in params", node.FqName())
	}

	return
}

// Invoke get towards aws secrets manager
func (p *Serializer) getFromAws(prm string,
	nasm *tagparser.AsmTag) (*secretsmanager.GetSecretValueOutput, error) {

	var params *secretsmanager.GetSecretValueInput
	if nasm.VersionID() == "" && nasm.VersionStage() == "" {
		params = &secretsmanager.GetSecretValueInput{SecretId: aws.String(prm)}
	} else if nasm.VersionStage() != "" {
		params = &secretsmanager.GetSecretValueInput{SecretId: aws.String(prm), VersionStage: aws.String(nasm.VersionStage())}
	} else {
		params = &secretsmanager.GetSecretValueInput{SecretId: aws.String(prm), VersionId: aws.String(nasm.VersionID())}
	}

	log.Debug().Str("svc", p.service).Msgf("Fetching %v", *params)

	client := secretsmanager.New(p.config)
	req := client.GetSecretValueRequest(params)
	resp, err := req.Send(context.TODO())

	if err != nil {
		log.Debug().Msgf("error for '%s': %v err %v", prm, resp, err)
		return nil, err
	}
	return resp.GetSecretValueOutput, nil
}
