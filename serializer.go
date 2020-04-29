package ssm

import (
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/pms"
	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/support"
)

// Serializer handles un-/marshaling of SSM data
// back and forth go struct fields.
type Serializer struct {
	config  aws.Config
	region  string
	service string
	env     string
	tier    ssm.ParameterTier
}

// NewSsmSerializer creates a new serializer with default aws.Config
func NewSsmSerializer(env string, service string) *Serializer {
	return &Serializer{env: env, service: service, tier: ssm.ParameterTierStandard}
}

// SetTier allows for change the tier. By default Serializer uses
// the standard tier.
func (s *Serializer) SetTier(tier ssm.ParameterTier) *Serializer {
	s.tier = tier
	return s
}

// Unmarshal creates the inparam struct pointer (and sub structs as well).
// It will populate the fields that are denoted with pms and asm
// with data from the Systems Manager. It returns a map containg fields that
// where requested but not set
func (s *Serializer) Unmarshal(v interface{}) (map[string]support.FullNameField, error) {
	tp := reflect.ValueOf(v)
	node, err := reflectparser.New(s.env, s.service).Parse("", tp)
	if err != nil {
		return nil, err
	}

	pmsr, err := pms.New(s.service)
	if err != nil {
		return nil, err
	}

	pmsr.SetTier(s.tier)

	invalid, err := pmsr.Get(&node)
	return invalid, err
}
