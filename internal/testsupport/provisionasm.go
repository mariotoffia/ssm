package testsupport

import (
	"context"
	"fmt"

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
func DefaultProvisionAsm() {
	ProvisionAsm(Secrets())
}

// Secrets generates all secrets managed by the test system
func Secrets() []secretsmanager.CreateSecretInput {
	return []secretsmanager.CreateSecretInput{
		{Name: aws.String("/eap/simple/test"),
			SecretString:       aws.String("The name"),
			ClientRequestToken: aws.String(uuid.New().String())},
		{Name: aws.String("/eap/test-service/asmsub/ext"),
			SecretString:       aws.String("43"),
			ClientRequestToken: aws.String(uuid.New().String())},
		{Name: aws.String("/eap/test-service/asmsub/myname"),
			SecretString:       aws.String("test svc name"),
			ClientRequestToken: aws.String(uuid.New().String())},
	}

	// if any changes, let the value settle
	// time.Sleep(1000 * time.Millisecond)
}

var delete = false

// ProvisionAsm provision secrets
func ProvisionAsm(prms []secretsmanager.CreateSecretInput) {

	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic(err)
	}

	svc := secretsmanager.New(awscfg)

	for _, p := range prms {
		log.Info().Msgf("Creating secret %s", *p.Name)
		req := svc.CreateSecretRequest(&p)
		if _, err := req.Send(context.Background()); err != nil {
			log.Debug().Msgf("Failed to create secret %s, checking if PUT works", *p.Name)

			if e := putSecret(svc, p); e != nil {
				log.Warn().Msgf("error when create secret %v", err)
			}
		}
	}
}

func putSecret(svc *secretsmanager.Client, i secretsmanager.CreateSecretInput) error {

	req := svc.PutSecretValueRequest(&secretsmanager.PutSecretValueInput{
		ClientRequestToken: i.ClientRequestToken,
		SecretId:           i.Name,
		SecretString:       i.SecretString,
	})

	if _, err := req.Send(context.Background()); err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Warn().Msgf("Failed to put secret %s - error: %v", *i.Name, aerr.Error())
		}

		return err
	}

	log.Info().Msgf("Secret %s successfully updated", *i.Name)

	return nil
}

// DeleteAsm deletes secrets
func DeleteAsm(prms []secretsmanager.CreateSecretInput) {
	for _, p := range prms {
		internalDelete(secretsmanager.DeleteSecretInput{SecretId: aws.String(*p.Name),
			ForceDeleteWithoutRecovery: aws.Bool(true)})
	}
}

func internalDelete(prms secretsmanager.DeleteSecretInput) error {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to load AWS config")
	}

	svc := secretsmanager.New(awscfg)

	fmt.Printf("deleting %v", prms)
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
