package asm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/mariotoffia/ssm/internal/common"
	"github.com/mariotoffia/ssm/parser"
	"github.com/mariotoffia/ssm/support"
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

// New creates a repository using the default configuration.
func New(service string) (*Serializer, error) {

	awscfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return &Serializer{}, errors.Wrapf(err, "Failed to load AWS config")
	}

	return &Serializer{config: awscfg, service: service}, nil
}

// Get parameters from the secrets manager and populates the node graph with values.
// Any fields that was not able to be set is reported in the FullNameField string map.
// FullNameField do not include those fields filtered out in exclusion filter.
func (p *Serializer) Get(node *parser.StructNode,
	filter *support.FieldFilters) (map[string]support.FullNameField, error) {

	m := map[string]*parser.StructNode{}
	parser.NodesToParameterMap(node, m, filter, []string{"asm"})
	mprms := map[string]*secretsmanager.GetSecretValueOutput{}
	im := map[string]support.FullNameField{}
	extrprms := parser.ExtractPaths(m)

	log.Debug().Str("svc", p.service).
		Str("package", "asm").
		Str("method", "Get").
		Msgf("Fetching: %v", extrprms)

	for _, prm := range extrprms {
		if n, ok := m[prm]; ok {

			if nasm, ok := ToAsmTag(n); ok {
				result, err := p.getFromAws(prm, nasm)

				if err != nil {

					var resourceNotFound *types.ResourceNotFoundException

					if errors.As(err, &resourceNotFound) {

						im[n.FqName] = support.FullNameField{LocalName: n.FqName,
							RemoteName: prm, Field: node.Field, Value: node.Value}

					} else {

						return nil, errors.Wrapf(err, "Failed fetch asm config entry %s", prm)

					}

				} else {

					log.Debug().Str("svc", p.service).Str("method", "Get").Msgf("field %s", n.FqName)
					mprms[n.FqName] = result

				}
			} else {

				log.Warn().Str("svc", p.service).Msgf("tag is not asm tag! tag: %v", n)

			}
		}
	}

	populate(node, mprms)

	return im, nil
}

// Upsert creates or updates a secret.
func (p *Serializer) Upsert(node *parser.StructNode,
	filter *support.FieldFilters) map[string]support.FullNameField {

	m := map[string]*parser.StructNode{}
	parser.NodesToParameterMap(node, m, filter, []string{"asm"})

	// TODO: Implement me!
	im := map[string]support.FullNameField{}

	client := secretsmanager.NewFromConfig(p.config)
	params := genCreateSecretParams(m)

	for _, prm := range params {
		node := m[*prm.Name]

		_, err := p.createAwsSecret(client, prm)
		if err != nil {
			_, err := p.updateAwsSecret(client, prm)
			if err != nil {
				im[node.FqName] = support.FullNameField{LocalName: node.FqName,
					RemoteName: *prm.Name, Error: err, Field: node.Field, Value: node.Value}
			} else {

				if len(prm.Tags) > 0 {

					_, err = p.tagAwsSecret(client, prm)
					if err != nil {
						im[node.FqName] = support.FullNameField{LocalName: node.FqName,
							RemoteName: *prm.Name, Error: err, Field: node.Field, Value: node.Value}
					}
				}
			}
		}
	}

	return im
}

func populate(node *parser.StructNode, params map[string]*secretsmanager.GetSecretValueOutput) {
	node.EnsureInstance(false)

	if val, ok := params[node.FqName]; ok {
		if tag, ok := node.Tag["asm"]; ok {
			if tag.GetFullName() != "" {
				common.SetStructValueFromString(node, *val.Name, *val.SecretString)
			}
		}
	}

	if node.HasChildren() {
		for _, n := range node.Childs {
			populate(&n, params)
		}
		return
	}
}
