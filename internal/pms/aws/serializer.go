package awspms

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/pkg/errors"
)

// Serializer handles the parameter store communication
type Serializer struct {
	// AWS Config to use when communicating
	config aws.Config
	// The name of the service using this library
	service string
	// Default tier if not specified.
	tier ssm.ParameterTier
}

// SeDefaulttTier allows for change the tier. By default the
// serializer uses the standard tier.
func (p *Serializer) SeDefaulttTier(tier ssm.ParameterTier) *Serializer {
	p.tier = tier
	return p
}

// NewFromConfig creates a repository using a existing configuration
func NewFromConfig(config aws.Config, service string) *Serializer {
	return &Serializer{config: config, service: service,
		tier: ssm.ParameterTierStandard}
}

// New creates a repository using the default AWS configuration
func New(service string) (*Serializer, error) {
	awscfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return &Serializer{}, errors.Wrapf(err, "Failed to load AWS config")
	}

	return &Serializer{config: awscfg, service: service,
		tier: ssm.ParameterTierStandard}, nil
}
