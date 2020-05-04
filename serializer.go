package ssm

import (
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/internal/asm"
	"github.com/mariotoffia/ssm.git/internal/pms"
	"github.com/mariotoffia/ssm.git/internal/reflectparser"
	"github.com/mariotoffia/ssm.git/support"
)

// Usage determines how the tags on the structs are evaluated
type Usage string

const (
	// UsePms will enable the Systems Manager Parameter Store tags
	UsePms Usage = "pms"
	// UseAsm will enable the AWS Secrets Manager tags
	UseAsm = "asm"
)

// NoFilter specifieds that no Filtering shall be done
var NoFilter *support.FieldFilters = &support.FieldFilters{}

// AllTags will use all tags in the struct that this un-/marshaller supports
var AllTags []Usage = []Usage{UsePms, UseAsm}

// OnlyPms will enable only Parameter Store tags
var OnlyPms []Usage = []Usage{UsePms}

// OnlyAsm will enable only Secrets Manager tags
var OnlyAsm []Usage = []Usage{UseAsm}

// Serializer handles un-/marshaling of SSM data
// back and forth go struct fields. Default is
// all tags used when un-/marshal
type Serializer struct {
	hasconfig bool
	config    aws.Config
	service   string
	env       string
	tier      ssm.ParameterTier
	usage     []Usage
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
	return s.unmarshal(v, nil, nil)
}

// UnmarshalWithOpts creates the inparam struct pointer (and sub structs as well).
// It will populate the fields that are denoted with pms and asm
// with data from the Systems Manager. It returns a map containg fields that
// where requested but not set. This version of Unmarshal accepts a set of inclusion
// & exclusion filters. The type is only initialized with the non excluded or explicit
// included field. By default  property is excluded. See @support.FieldFilters for more
// informatio about filtering. It also accepts a set of usage directives that the calling
// code may turn off or on certain tags (for example do only unmarshal PMS data). By
// default the serializer will use all supported tags.
func (s *Serializer) UnmarshalWithOpts(v interface{},
	filter *support.FieldFilters, usage []Usage) (map[string]support.FullNameField, error) {
	return s.unmarshal(v, filter, usage)
}

func (s *Serializer) unmarshal(v interface{},
	filter *support.FieldFilters, usage []Usage) (map[string]support.FullNameField, error) {

	if len(usage) == 0 {
		if len(s.usage) > 0 {
			usage = s.usage
		} else {
			usage = []Usage{UsePms, UseAsm}
		}
	}

	if nil == filter {
		filter = support.NewFilters()
	}

	tp := reflect.ValueOf(v)
	node, err := reflectparser.New(s.env, s.service).Parse("", tp)

	if err != nil {
		return nil, err
	}

	var invalid map[string]support.FullNameField

	if _, found := find(usage, UsePms); found {
		pmsr, err := s.getAndConfigurePms()
		if err != nil {
			return nil, err
		}

		invalid, err = pmsr.Get(&node, filter)
	}

	if _, found := find(usage, UseAsm); found {
		asmr, err := s.getAndConfigureAsm()
		if err != nil {
			return nil, err
		}

		invalid2, err := asmr.Get(&node, filter)
		if invalid == nil && len(invalid2) > 0 {
			invalid = map[string]support.FullNameField{}
		}

		// Merge field errors from ASM with PMS errors
		for key, value := range invalid2 {
			invalid[key] = value
		}
	}

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

func (s *Serializer) getAndConfigureAsm() (*asm.Serializer, error) {
	if s.hasconfig {
		return asm.NewFromConfig(s.config, s.service), nil
	}

	asmr, err := asm.New(s.service)
	if err != nil {
		return nil, err
	}

	return asmr, nil
}

func find(slice []Usage, val Usage) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
