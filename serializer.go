package ssm

import (
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/mariotoffia/ssm/internal/asm"
	"github.com/mariotoffia/ssm/internal/pms"
	"github.com/mariotoffia/ssm/parser"
	"github.com/mariotoffia/ssm/report"
	"github.com/mariotoffia/ssm/support"
)

// Usage determines how the tags on the struct are evaluated
type Usage string

const (
	// UsePms will enable the Systems Manager Parameter Store tags
	UsePms Usage = "pms"
	// UseAsm will enable the AWS Secrets Manager tags
	UseAsm = "asm"
)

// NoFilter specified that no Filtering shall be done
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
	tier      types.ParameterTier
	usage     []Usage
	parser    map[string]parser.TagParser
	prefix    string
}

// NewSsmSerializer creates a new serializer with default aws.Config
func NewSsmSerializer(env string, service string) *Serializer {
	return &Serializer{
		env:     env,
		service: service,
		tier:    types.ParameterTierStandard,
		parser:  map[string]parser.TagParser{},
	}
}

// NewSsmSerializerFromConfig creates a new serializer using the in param config instead
// of the default config.
func NewSsmSerializerFromConfig(env string, service string, config aws.Config) *Serializer {
	return &Serializer{
		env:       env,
		service:   service,
		tier:      types.ParameterTierStandard,
		config:    config,
		hasconfig: true,
		parser:    map[string]parser.TagParser{},
	}
}

// UseTagParser registers a custom tag parser to participate in the reflective
// parsing when Marshal or Unmarshal operations.
func (s *Serializer) UseTagParser(tag string, parser parser.TagParser) *Serializer {
	s.parser[tag] = parser
	return s
}

// UsePrefix acts as a default prefix if no prefix is specified in the tag.
//
// Prefix operates under two modes: _Local_ and _Global_.
//
// .Local vs Global Mode
// [cols="1,1,4"]
// |===
// |Mode |Example |Description
//
// |Local
// |my-local-prefix/nested
// |This will render environment/service/my-local-prefix/nested/property. E.g. dev/tes-service/my-local-prefix/nested/password
//
// |Global
// |/my-global-prefix/nested
// |This will render environment/my-global-prefix/nested/property. E.g. dev/my-global-prefix/nested/password
//
// |===
//
// NOTE: When global prefix, the _service_ element is eliminated (in order to have singeltons).
func (s *Serializer) UsePrefix(prefix string) *Serializer {
	s.prefix = prefix
	return s
}

// SetTier allows for change the tier. By default Serializer uses
// the standard tier.
func (s *Serializer) SetTier(tier types.ParameterTier) *Serializer {
	s.tier = tier
	return s
}

// Delete creates the in param struct pointer (and sub struct as well).
// It will search the fields that are denoted with pms and asm
// with data from the Systems Manager. It tries to delete all keys. It returns
// a map contains fields that where failed to be deleted.
func (s *Serializer) Delete(v interface{}) (map[string]support.FullNameField, error) {
	inv, _, err := s.delete(v, nil, nil)
	return inv, err
}

// DeleteWithOpts creates the in param struct pointer (and sub struct as well).
// It will search the fields that are denoted with pms and asm
// with data from the Systems Manager. It tries to delete all keys. It returns
// a map contains fields that where failed to be deleted.
//
// This version of Delete accepts a set of inclusion & exclusion filters. The type
// is only parsed with the non excluded or explicit included field. By default
// property is excluded. See @support.FieldFilters for more information about filtering.
// It also accepts a set of usage directives that the calling code may turn off or on
// certain tags (for example do only delete PMS data). By default the serializer will
// use all supported tags.
func (s *Serializer) DeleteWithOpts(v interface{},
	filter *support.FieldFilters, usage []Usage) (map[string]support.FullNameField, error) {
	inv, _, err := s.delete(v, filter, usage)
	return inv, err
}

// AdvDeleteWithOpts is the same function as the DeleteWithOpts but is
// also returns the tree of parsed node to be post-processed.
func (s *Serializer) AdvDeleteWithOpts(v interface{},
	filter *support.FieldFilters,
	usage []Usage) (map[string]support.FullNameField, *parser.StructNode, error) {
	return s.delete(v, filter, usage)
}

// Unmarshal creates the in param struct pointer (and sub struct as well).
// It will populate the fields that are denoted with pms and asm
// with data from the Systems Manager. It returns a map contains fields that
// where requested but not set.
func (s *Serializer) Unmarshal(v interface{}) (map[string]support.FullNameField, error) {
	inv, _, err := s.unmarshal(v, nil, nil)
	return inv, err
}

// UnmarshalWithOpts creates the in param struct pointer (and sub struct as well).
// It will populate the fields that are denoted with pms and asm
// with data from the Systems Manager. It returns a map contains fields that
// where requested but not set. This version of Unmarshal accepts a set of inclusion
// & exclusion filters. The type is only initialized with the non excluded or explicit
// included field. By default  property is excluded. See @support.FieldFilters for more
// information about filtering. It also accepts a set of usage directives that the calling
// code may turn off or on certain tags (for example do only unmarshal PMS data). By
// default the serializer will use all supported tags.
func (s *Serializer) UnmarshalWithOpts(v interface{},
	filter *support.FieldFilters, usage []Usage) (map[string]support.FullNameField, error) {
	inv, _, err := s.unmarshal(v, filter, usage)
	return inv, err
}

// AdvUnmarshalWithOpts is the same function as the UnmarshalWithOpts but is
// also returns the tree of parsed node to be post-processed.
func (s *Serializer) AdvUnmarshalWithOpts(v interface{},
	filter *support.FieldFilters,
	usage []Usage) (map[string]support.FullNameField, *parser.StructNode, error) {
	return s.unmarshal(v, filter, usage)
}

// Marshal serializes the struct and sub-struct onto parameter store and AWS secrets
// manager. The values are not checked, it will bluntly **upsert** the data onto the
// remote storage. It returns a map contains fields that where tried to be set but for
// some reason fails. The error property is always filled in using Marshal (as opposed to
// Unmarshal where it is never filled in). If any non field related error occurs an empty
// support.FullNameField is returned with only the Error field populated.
func (s *Serializer) Marshal(v interface{}) map[string]support.FullNameField {
	inv, _ := s.marshal(v, nil, nil)
	return inv
}

// MarshalWithOpts serializes the struct and sub-struct onto parameter store and AWS secrets
// manager. The values are not checked, it will bluntly **upsert** the data onto the
// remote storage. It returns a map contains fields that where tried to be set but for
// some reason fails. The error property is always filled in using Marshal (as opposed to
// Unmarshal where it is never filled in). If any non field related error occurs an empty
// support.FullNameField is returned with only the Error field populated.
//
// In this version of Marshal it is possible to specify exactly which parameters to be
// marshalled and which are not (see UnmarshalWithOpts for filter discussion). It is also
// possible to explicitly enable / disable PMS or ASM and hence gain a little optimization.
func (s *Serializer) MarshalWithOpts(v interface{},
	filter *support.FieldFilters, usage []Usage) map[string]support.FullNameField {
	inv, _ := s.marshal(v, filter, usage)
	return inv
}

// AdvMarshalWithOpts is exactly as MarshalWithOpts but it also returns
// the parsed node tree. It is possible to register your own tag parsers using
// UseTagParser(name, tagparser) method.
func (s *Serializer) AdvMarshalWithOpts(v interface{},
	filter *support.FieldFilters,
	usage []Usage) (map[string]support.FullNameField, *parser.StructNode) {
	return s.marshal(v, filter, usage)
}

// ReportWithOpts generates a struct based and JSON based report of the in param type
// or actual struct value to have default values generated.
// The JSON report is on the following example format:
// "parameters": [
//	  {
//		  "type": "secrets-manager",
//		  "fqname": "/prod/simple/test",
//		  "keyid": "",
//		  "description": "A test secret",
//		  "tags": {"test": "true"},
//		  "details": null,
//		  "value": "777"
//	  }
//  ]
//}
func (s *Serializer) ReportWithOpts(v interface{},
	filter *support.FieldFilters,
	values bool) (*report.Report, string, error) {

	tp := reflect.ValueOf(v)
	node, err := parser.New("test-service", "prod", s.prefix).
		RegisterTagParser("asm", asm.NewTagParser()).
		RegisterTagParser("pms", pms.NewTagParser()).
		Parse(tp)

	if err != nil {
		return nil, "", err
	}

	reporter := report.NewWithTier(s.tier)
	return reporter.RenderReport(node, filter, true)
}
