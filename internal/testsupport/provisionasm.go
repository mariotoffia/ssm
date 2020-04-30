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
)

// DefaultProvisionAsm provisions a test default
func DefaultProvisionAsm() {
	ProvisionAsm([]secretsmanager.CreateSecretInput{
		{Name: aws.String("/eap/simple/test"),
			SecretString:       aws.String("The name"),
			ClientRequestToken: aws.String(uuid.New().String())},
		{Name: aws.String("/eap/test-service/asmsub/ext"),
			SecretString:       aws.String("43"),
			ClientRequestToken: aws.String(uuid.New().String())},
		{Name: aws.String("/eap/test-service/asmsub/myname"),
			SecretString:       aws.String("test svc name"),
			ClientRequestToken: aws.String(uuid.New().String())},
	})
}

var onlydelete = false
var disable = false

// ProvisionAsm provision secrets
func ProvisionAsm(prms []secretsmanager.CreateSecretInput) {
	if disable {
		return
	}

	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic(err)
	}

	for _, p := range prms {
		DeleteAsm(secretsmanager.DeleteSecretInput{SecretId: aws.String(*p.Name),
			ForceDeleteWithoutRecovery: aws.Bool(true)})
	}

	if onlydelete {
		return
	}
	svc := secretsmanager.New(awscfg)

	for _, p := range prms {
		req := svc.CreateSecretRequest(&p)
		if _, err := req.Send(context.Background()); err != nil {
			fmt.Printf("error when create secret %v", err)
		}
	}
}

// DeleteAsm deletes secrets
func DeleteAsm(prms secretsmanager.DeleteSecretInput) error {
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
				fmt.Printf("Error when deleting %v", prms)
				return err
			}
		}
	}
	return nil
}
