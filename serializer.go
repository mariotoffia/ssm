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
	hasconfig bool
	config    aws.Config
	service   string
	env       string
	tier      ssm.ParameterTier
}

// NewSsmSerializer creates a new serializer with default aws.Config
func NewSsmSerializer(env string, service string) *Serializer {
	return &Serializer{env: env, service: service, tier: ssm.ParameterTierStandard}
}

// NewSsmSerializerFromConfig creates a new serializer using the inparam config instead
// of the default config.
func NewSsmSerializerFromConfig(env string, service string, config aws.Config) *Serializer {
	return &Serializer{env: env, service: service,
		tier: ssm.ParameterTierStandard, config: config,
		hasconfig: true}
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
// where requested but not set.
func (s *Serializer) Unmarshal(v interface{}) (map[string]support.FullNameField, error) {
	return s.UnmarshalFilterd(v, support.FieldFilters{})
}

// UnmarshalFilterd creates the inparam struct pointer (and sub structs as well).
// It will populate the fields that are denoted with pms and asm
// with data from the Systems Manager. It returns a map containg fields that
// where requested but not set. This version of Unmarshal accepts a set of inclusion
// & exclusion filters. The type is only initialized with the non excluded or explicit
// included field. By default  property is excluded. See @support.FieldFilters for more
// informatio about filtering.
func (s *Serializer) UnmarshalFilterd(v interface{}, filter support.FieldFilters) (map[string]support.FullNameField, error) {

	tp := reflect.ValueOf(v)
	node, err := reflectparser.New(s.env, s.service).Parse("", tp)

	if err != nil {
		return nil, err
	}

	pmsr, err := s.getAndConfigurePms()
	if err != nil {
		return nil, err
	}

	invalid, err := pmsr.Get(&node, filter)
	return invalid, err
}

func (s *Serializer) getAndConfigurePms() (*pms.Serializer, error) {
	if s.hasconfig {
		return pms.NewFromConfig(s.config, s.service).SetTier(s.tier), nil
	}

	pmsr, err := pms.New(s.service)
	if err != nil {
		return nil, err
	}

	return pmsr.SetTier(s.tier), nil
}
