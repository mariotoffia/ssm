package testsupport

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// DefaultProvisionAsm provisions a test default environment
// for AWS Secrets Manager. Since the time to create / delete
// and when the value is available or may be re-created is
// lagging (e.g. the notorious the value is schedule for deletion)
// this function will check if exist, if so update the value, if
// missing it will create the secret. It will not delete the
// secrets by default.
func DefaultProvisionAsm() string {
	stage := UnittestStage()
	DeleteAllUnittestSecrets()
	ProvisionAsm(Secrets(stage))

	return stage
}

// Secrets generates all secrets managed by the test system
func Secrets(stage string) []secretsmanager.CreateSecretInput {
	return []secretsmanager.CreateSecretInput{
		{Name: aws.String(fmt.Sprintf("/%s/simple/test", stage)),
			SecretString:       aws.String("The name"),
			ClientRequestToken: aws.String(uuid.New().String())},
		{Name: aws.String(fmt.Sprintf("/%s/test-service/asmsub/ext", stage)),
			SecretString:       aws.String("43"),
			ClientRequestToken: aws.String(uuid.New().String())},
		{Name: aws.String(fmt.Sprintf("/%s/test-service/asmsub/myname", stage)),
			SecretString:       aws.String("test svc name"),
			ClientRequestToken: aws.String(uuid.New().String())},
		{Name: aws.String(fmt.Sprintf("/%s/test-service/bubbibobbo", stage)),
			SecretString:       aws.String(`{"user":"gurkaburka","timeout":998}`),
			ClientRequestToken: aws.String(uuid.New().String())},
	}
}

// ProvisionAsm provision secrets
func ProvisionAsm(prms []secretsmanager.CreateSecretInput) {

	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic(err)
	}

	svc := secretsmanager.New(awscfg)

	for _, p := range prms {
		log.Info().Msgf("Creating asm-secret %s", *p.Name)
		req := svc.CreateSecretRequest(&p)
		if _, err := req.Send(context.Background()); err != nil {
			log.Debug().Msgf("Failed to create asm-secret %s", *p.Name)
			panic(err)
		}
	}
}

// DeleteAllUnittestSecrets tries to delete all unit test secrets
func DeleteAllUnittestSecrets() error {
	inp := secretsmanager.ListSecretsInput{}

	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to load AWS config")
	}

	svc := secretsmanager.New(awscfg)
	for {
		req := svc.ListSecretsRequest(&inp)
		resp, err := req.Send(context.Background())
		if err != nil {
			log.Warn().Msgf("Failed to list asm-secrets %v", err)
			break
		}

		inp.NextToken = resp.NextToken
		for _, s := range resp.SecretList {
			log.Debug().Msgf("Found asm-secret %s", *s.Name)
			if strings.HasPrefix(*s.Name, "/unittest-") {
				internalDelete(secretsmanager.DeleteSecretInput{SecretId: aws.String(*s.Name),
					ForceDeleteWithoutRecovery: aws.Bool(true)})
			}
		}

		if resp.NextToken == nil {
			log.Debug().Msg("No more asm-secrets to delete (note that you may to delete them some minutes after creation to be found!")
			break
		}
	}

	return nil
}

func internalDelete(prms secretsmanager.DeleteSecretInput) error {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to load AWS config")
	}

	svc := secretsmanager.New(awscfg)

	fmt.Printf("deleting-asm %v", prms)
	req := svc.DeleteSecretRequest(&prms)
	if _, err := req.Send(context.Background()); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				break
			default:
				log.Warn().Msgf("Error when deleting %v", prms)
				return err
			}
		}
	}
	return nil
}
