package asm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/google/uuid"
	"github.com/mariotoffia/ssm/internal/common"
	"github.com/mariotoffia/ssm/parser"
	"github.com/rs/zerolog/log"
)

func genCreateSecretParams(nodes map[string]*parser.StructNode) []secretsmanager.CreateSecretInput {

	prms := []secretsmanager.CreateSecretInput{}
	for _, node := range nodes {
		prms = append(prms, genCreateSecretParam(node))
	}

	return prms
}

func genCreateSecretParam(node *parser.StructNode) secretsmanager.CreateSecretInput {
	var keyid *string = nil
	var tags []types.Tag = nil

	if tag, ok := ToAsmTag(node); ok {
		if !tag.DefaultAccountKey() {
			keyid = aws.String(tag.GetKeyName())
		}

		t := tag.Tag()
		if len(t) > 0 {
			tags = []types.Tag{}
			for key := range t {
				tags = append(tags, types.Tag{Key: aws.String(key), Value: aws.String(t[key])})
			}
		}

		return secretsmanager.CreateSecretInput{
			ClientRequestToken: aws.String(uuid.New().String()),
			Name:               aws.String(tag.GetFullName()),
			Description:        aws.String(tag.Description()),
			KmsKeyId:           keyid,
			SecretString:       aws.String(common.GetStringValueFromField(node)),
			Tags:               tags,
		}
	}

	panic(node)
}

// Invoke get towards aws secrets manager
func (p *Serializer) getFromAws(prm string,
	nasm *AsmTagStruct) (*secretsmanager.GetSecretValueOutput, error) {

	var params *secretsmanager.GetSecretValueInput
	if nasm.VersionID() == "" && nasm.VersionStage() == "" {
		params = &secretsmanager.GetSecretValueInput{SecretId: aws.String(prm)}
	} else if nasm.VersionStage() != "" {
		params = &secretsmanager.GetSecretValueInput{SecretId: aws.String(prm), VersionStage: aws.String(nasm.VersionStage())}
	} else {
		params = &secretsmanager.GetSecretValueInput{SecretId: aws.String(prm), VersionId: aws.String(nasm.VersionID())}
	}

	client := secretsmanager.NewFromConfig(p.config)

	resp, err := client.GetSecretValue(context.Background(), params)

	if err != nil {
		log.Debug().Msgf("error for '%s': %v err %v", prm, resp, err)
		return nil, err
	}
	return resp, nil
}

func (p *Serializer) createAwsSecret(client *secretsmanager.Client,
	secret secretsmanager.CreateSecretInput) (*secretsmanager.CreateSecretOutput, error) {

	resp, err := client.CreateSecret(context.Background(), &secret)

	if err != nil {
		log.Debug().Msgf("create error for '%s': %v err %v", *secret.Name, resp, err)
		return nil, err
	}

	log.Debug().Str("svc", p.service).Str("method", "createAwsSecret").
		Msgf("created secret %s value ***", *secret.Name)

	return resp, nil

}

func (p *Serializer) updateAwsSecret(client *secretsmanager.Client,
	secret secretsmanager.CreateSecretInput) (*secretsmanager.UpdateSecretOutput, error) {

	resp, err := client.UpdateSecret(context.Background(), &secretsmanager.UpdateSecretInput{
		ClientRequestToken: secret.ClientRequestToken,
		Description:        secret.Description,
		KmsKeyId:           secret.KmsKeyId,
		SecretId:           secret.Name,
		SecretString:       secret.SecretString,
	})

	if err != nil {
		log.Debug().Msgf("update error for '%s': %v err %v", *secret.Name, resp, err)
		return nil, err
	}

	log.Debug().Str("svc", p.service).Str("method", "updateAwsSecret").
		Msgf("updated secret %s value ***", *secret.Name)

	return resp, nil

}

func (p *Serializer) tagAwsSecret(client *secretsmanager.Client,
	secret secretsmanager.CreateSecretInput) (*secretsmanager.TagResourceOutput, error) {

	resp, err := client.TagResource(context.Background(), &secretsmanager.TagResourceInput{
		SecretId: secret.Name,
		Tags:     secret.Tags,
	})

	if err != nil {
		log.Debug().Msgf("update tgs error for '%s': %v err %v", *secret.Name, resp, err)
		return nil, err
	}

	log.Debug().Str("svc", p.service).Str("method", "tagAwsSecret").
		Msgf("tagged secret %s tags %v", *secret.Name, secret.Tags)

	return resp, nil

}
