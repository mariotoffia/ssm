package testsupport

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ProvisionPms will provision all parameters.
// If already existant, it will just be overwritten.
func ProvisionPms(prms []ssm.PutParameterInput) error {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to load AWS config")
	}

	delete := ssm.DeleteParametersInput{}
	for _, p := range prms {
		if nil != p.Overwrite && *p.Overwrite {
			delete.Names = append(delete.Names, *p.Name)
		}
	}

	if len(delete.Names) > 0 {
		if err := DeletePms(delete); err != nil {
			return err
		}
	}

	client := ssm.New(awscfg)
	for _, p := range prms {
		req := client.PutParameterRequest(&p)
		resp, err := req.Send(context.TODO())
		if err != nil {
			return err
		}

		log.Debug().Msgf("Wrote name: %s value: %s got version %d", *p.Name, *p.Value, *resp.Version)
	}
	return nil
}

// DeletePms removes a set of Parameter Store Parameters
func DeletePms(prms ssm.DeleteParametersInput) error {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to load AWS config")
	}

	client := ssm.New(awscfg)
	req := client.DeleteParametersRequest(&prms)
	resp, err := req.Send(context.TODO())
	if err != nil {
		return err
	}

	log.Debug().Msgf("deleted: %s invalid: %s", resp.DeletedParameters, resp.InvalidParameters)
	return nil
}

// DefaultProvisionPms sets up a defualt test environment for PMS
func DefaultProvisionPms() error {
	return ProvisionPms([]ssm.PutParameterInput{
		{Name: aws.String("/eap/simple/test"), Type: ssm.ParameterTypeString,
			Overwrite: aws.Bool(true), Value: aws.String("The name")},
		{Name: aws.String("/eap/test-service/sub/ext"), Type: ssm.ParameterTypeString,
			Overwrite: aws.Bool(true), Value: aws.String("43")},
		{Name: aws.String("/eap/test-service/sub/myname"), Type: ssm.ParameterTypeString,
			Overwrite: aws.Bool(true), Value: aws.String("test svc name")}})
}
