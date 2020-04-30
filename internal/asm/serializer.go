package asm

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/pkg/errors"
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
