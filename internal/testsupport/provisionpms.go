package testsupport

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ProvisionPms will provision all parameters.
// If already existant, it will just be overwritten.
func provisionPms(prms []ssm.PutParameterInput) error {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to load AWS config")
	}

	client := ssm.New(awscfg)
	for _, p := range prms {
		req := client.PutParameterRequest(&p)
		resp, err := req.Send(context.Background())
		if err != nil {
			return err
		}

		log.Debug().Msgf("Wrote name: %s value: %s got version %d", *p.Name, *p.Value, *resp.Version)
	}
	return nil
}

// ListDeletePrms lists and deletes all parameters that begins with /unittest
func listDeletePrms() error {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return errors.Wrapf(err, "Failed to load AWS config")
	}

	log.Debug().Msgf("region %s", awscfg.Region)
	client := ssm.New(awscfg)

	inp := ssm.DescribeParametersInput{
		ParameterFilters: []ssm.ParameterStringFilter{{
			Key:    aws.String("Name"),
			Option: aws.String("BeginsWith"),
			Values: []string{"/unittest-"},
		}}}

	for {
		req := client.DescribeParametersRequest(&inp)
		res, err := req.Send(context.Background())
		if err != nil {
			log.Warn().Msgf("got error when listing params for deletion error: %v", err)
			return err
		}

		dprm := ssm.DeleteParametersInput{}
		for _, prm := range res.Parameters {
			log.Debug().Msgf("Deleting param name: %s version %d", *prm.Name, *prm.Version)
			dprm.Names = append(dprm.Names, *prm.Name)
		}

		dreq := client.DeleteParametersRequest(&dprm)
		_, err = dreq.Send(context.Background())
		if err != nil {
			log.Warn().Msgf("got error when deleting params error: %v", err)
			return err
		}

		if res.NextToken == nil {
			break
		}

		inp.NextToken = res.NextToken
	}

	return nil
}

// DefaultProvisionPms sets up a defualt test environment for PMS
func DefaultProvisionPms(stage string) error {
	listDeletePrms()

	return provisionPms([]ssm.PutParameterInput{
		{Name: aws.String(fmt.Sprintf("/%s/simple/test", stage)), Type: ssm.ParameterTypeString,
			Overwrite: aws.Bool(true), Value: aws.String("The name")},
		{Name: aws.String(fmt.Sprintf("/%s/test-service/sub/ext", stage)), Type: ssm.ParameterTypeString,
			Overwrite: aws.Bool(true), Value: aws.String("43")},
		{Name: aws.String(fmt.Sprintf("/%s/test-service/sub/myname", stage)), Type: ssm.ParameterTypeString,
			Overwrite: aws.Bool(true), Value: aws.String("test svc name")}})
}
