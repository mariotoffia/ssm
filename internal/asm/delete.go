package asm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/mariotoffia/ssm/parser"
	"github.com/mariotoffia/ssm/support"
	"github.com/rs/zerolog/log"
)

// Delete will delete the paths found in nodes.
func (p *Serializer) Delete(
	node *parser.StructNode,
	filter *support.FieldFilters) (map[string]support.FullNameField, error) {

	m := map[string]*parser.StructNode{}
	svc := secretsmanager.NewFromConfig(p.config)

	parser.NodesToParameterMap(node, m, filter, []string{"asm"})

	im := map[string]support.FullNameField{}
	paths := parser.ExtractPaths(m)

	for _, path := range paths {

		err := internalDelete(
			svc,
			secretsmanager.DeleteSecretInput{SecretId: aws.String(path),
				ForceDeleteWithoutRecovery: true},
		)

		if err != nil {

			if val, ok := m[path]; ok {
				im[val.FqName] = support.FullNameField{
					RemoteName: path,
					LocalName:  val.FqName,
					Field:      val.Field,
					Value:      val.Value,
					Error:      err,
				}
			}

		}
	}

	return im, nil
}

// DeleteTree will delete all secrets that have a certain prefix.
// Since it is possible to specify many _prefixes_ this is able
// to delete several trees.
func (p *Serializer) DeleteTree(prefixes ...string) error {

	svc := secretsmanager.NewFromConfig(p.config)
	input := secretsmanager.ListSecretsInput{}

	for {

		resp, err := svc.ListSecrets(context.Background(), &input)

		if err != nil {

			log.Warn().Msgf("Failed to list asm-secrets %v", err)
			break

		}

		input.NextToken = resp.NextToken

		for _, s := range resp.SecretList {

			log.Debug().Msgf("Found asm-secret %s", *s.Name)

			if findPrefix(prefixes, *s.Name) {

				internalDelete(
					svc,
					secretsmanager.DeleteSecretInput{SecretId: aws.String(*s.Name),
						ForceDeleteWithoutRecovery: true},
				)

			}

		}

		if resp.NextToken == nil {

			log.Debug().Msg("No more asm-secrets to delete (note that you may to delete them some minutes after creation to be found!")
			break

		}

		input.NextToken = resp.NextToken
	}

	return nil
}

func findPrefix(array []string, val string) bool {

	for _, item := range array {
		if strings.HasPrefix(val, item) {
			return true
		}
	}

	return false
}

func internalDelete(svc *secretsmanager.Client, prms secretsmanager.DeleteSecretInput) error {

	fmt.Printf("deleting-asm %v", prms)

	if _, err := svc.DeleteSecret(context.Background(), &prms); err != nil {

		var resourceNotFound *smtypes.ResourceNotFoundException

		if errors.As(err, &resourceNotFound) {
			return nil
		}

		log.Warn().Msgf("Error when deleting %v", prms)
		return err
	}

	return nil
}
