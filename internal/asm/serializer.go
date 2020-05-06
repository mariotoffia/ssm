package asm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/uuid"
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
							im[n.FqName()] = support.FullNameField{LocalName: n.FqName(),
								RemoteName: prm, Field: node.Field(), Value: node.Value()}
						default:
							return nil, errors.Wrapf(err, "Failed fetch asm config entry %s", prm)
						}
					}
				} else {
					log.Debug().Str("svc", p.service).Str("method", "Get").Msgf("field %s = got %s", n.FqName(), *result.SecretString)
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

// Upsert creates or updates a secret.
func (p *Serializer) Upsert(node *reflectparser.SsmNode,
	filter *support.FieldFilters) map[string]support.FullNameField {

	m := map[string]*reflectparser.SsmNode{}
	common.NodesToParameterMap(node, m, filter, tagparser.Asm)

	// TODO: Implement me!
	im := map[string]support.FullNameField{}

	client := secretsmanager.New(p.config)
	params := genCreateSecretParams(m)

	for _, prm := range params {
		node := m[*prm.Name]

		_, err := p.createAwsSecret(client, prm)
		if err != nil {
			_, err := p.updateAwsSecret(client, prm)
			if err != nil {
				im[node.FqName()] = support.FullNameField{LocalName: node.FqName(),
					RemoteName: *prm.Name, Error: err, Field: node.Field(), Value: node.Value()}
			} else {

				if len(prm.Tags) > 0 {

					_, err = p.tagAwsSecret(client, prm)
					if err != nil {
						im[node.FqName()] = support.FullNameField{LocalName: node.FqName(),
							RemoteName: *prm.Name, Error: err, Field: node.Field(), Value: node.Value()}
					}
				}
			}
		}
	}

	return im
}

func genCreateSecretParams(nodes map[string]*reflectparser.SsmNode) []secretsmanager.CreateSecretInput {

	prms := []secretsmanager.CreateSecretInput{}
	for _, node := range nodes {
		prms = append(prms, genCreateSecretParam(node))
	}

	return prms
}

func genCreateSecretParam(node *reflectparser.SsmNode) secretsmanager.CreateSecretInput {
	var keyid *string = nil
	var tags []secretsmanager.Tag = nil

	if !node.Tag().DefaultAccountKey() {
		keyid = aws.String(node.Tag().GetKeyName())
	}

	t := node.Tag().Tags()
	if len(t) > 0 {
		tags = []secretsmanager.Tag{}
		for key := range t {
			tags = append(tags, secretsmanager.Tag{Key: aws.String(key), Value: aws.String(t[key])})
		}
	}

	return secretsmanager.CreateSecretInput{
		ClientRequestToken: aws.String(uuid.New().String()),
		Name:               aws.String(node.Tag().FullName()),
		Description:        aws.String(node.Tag().Description()),
		KmsKeyId:           keyid,
		SecretString:       aws.String(common.GetStringValueFromField(node)),
		Tags:               tags,
	}
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

	client := secretsmanager.New(p.config)
	req := client.GetSecretValueRequest(params)
	resp, err := req.Send(context.Background())

	if err != nil {
		log.Debug().Msgf("error for '%s': %v err %v", prm, resp, err)
		return nil, err
	}
	return resp.GetSecretValueOutput, nil
}

func (p *Serializer) createAwsSecret(client *secretsmanager.Client,
	secret secretsmanager.CreateSecretInput) (*secretsmanager.CreateSecretOutput, error) {

	req := client.CreateSecretRequest(&secret)
	resp, err := req.Send(context.Background())

	if err != nil {
		log.Debug().Msgf("create error for '%s': %v err %v", *secret.Name, resp, err)
		return nil, err
	}

	log.Debug().Str("svc", p.service).Str("method", "createAwsSecret").
		Msgf("created secret %s value %s", *secret.Name, *secret.SecretString)

	return resp.CreateSecretOutput, nil

}

func (p *Serializer) updateAwsSecret(client *secretsmanager.Client,
	secret secretsmanager.CreateSecretInput) (*secretsmanager.UpdateSecretOutput, error) {

	req := client.UpdateSecretRequest(&secretsmanager.UpdateSecretInput{
		ClientRequestToken: secret.ClientRequestToken,
		Description:        secret.Description,
		KmsKeyId:           secret.KmsKeyId,
		SecretId:           secret.Name,
		SecretString:       secret.SecretString,
	})
	resp, err := req.Send(context.Background())
	if err != nil {
		log.Debug().Msgf("update error for '%s': %v err %v", *secret.Name, resp, err)
		return nil, err
	}

	log.Debug().Str("svc", p.service).Str("method", "updateAwsSecret").
		Msgf("updated secret %s value %s", *secret.Name, *secret.SecretString)

	return resp.UpdateSecretOutput, nil

}

func (p *Serializer) tagAwsSecret(client *secretsmanager.Client,
	secret secretsmanager.CreateSecretInput) (*secretsmanager.TagResourceOutput, error) {

	req := client.TagResourceRequest(&secretsmanager.TagResourceInput{
		SecretId: secret.Name,
		Tags:     secret.Tags,
	})
	resp, err := req.Send(context.Background())
	if err != nil {
		log.Debug().Msgf("update tgs error for '%s': %v err %v", *secret.Name, resp, err)
		return nil, err
	}

	log.Debug().Str("svc", p.service).Str("method", "tagAwsSecret").
		Msgf("tagged secret %s tags %v", *secret.Name, secret.Tags)

	return resp.TagResourceOutput, nil

}
