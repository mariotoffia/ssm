package pms

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/mariotoffia/ssm.git/parser"
)

// ParamTier specifies the parameter tier such as std, adv, or intelligent.
type ParamTier string

const (
	// Std is the standard tier that allows one accouint to store
	// 10,000 parameters for free
	Std ParamTier = "std"
	// Adv allows for storage of up 8kb of data and uses different encryption algorithms
	// This is not a free tier and cost will incur
	Adv = "adv"
	// Eval is trying to evaluates each request to determine if the parameter is standard
	// or advanced. If the request doesn't include any options that require an advanced parameter,
	// the parameter is created in the standard-parameter tier.
	Eval = "eval"
	// Default is the default when noting is set and it will fall back on the set default in the serializer
	Default = "default"
)

// ToPmsTag converts a StructTag into the PmsTag interface
// If fails, it return false.
func ToPmsTag(generictag *parser.StructNode) (*PmsTagStruct, bool) {
	if ntag, ok := generictag.Tag["pms"]; ok {
		return ntag.(*PmsTagStruct), true
	}
	return nil, false
}

// ParameterType gets the Parameter Store Parameter Type.
func ParameterType(node *parser.StructNode) ssm.ParameterType {
	if tag, ok := ToPmsTag(node); ok {
		if tag.Secure() {
			return ssm.ParameterTypeSecureString
		}
		if node.Value.Kind() == reflect.Slice {
			return ssm.ParameterTypeStringList
		}
	}
	return ssm.ParameterTypeString
}

// PmsTag is a generic interface
type PmsTag interface {
	parser.StructTag
	Prefix() string
	Overwrite() bool
	Name() string
	Tags() map[string]string
	FqName() string
	Secure() bool
	Description() string
	DefaultAccountKey() bool
	IsLocalKey() bool
	GetKeyName() string
	Tier() ParamTier
	SsmTier(defaultTier ssm.ParameterTier) ssm.ParameterTier
	Pattern() string
	SsmTags() []ssm.Tag
}

// PmsTagStruct is for AWS parameter store
type PmsTagStruct struct {
	// Extend StructTagImpl
	parser.StructTagImpl
}

// Tier specifies the parameter tier it may be of std, adv, and eval (intelligent tiering).
// If nothing is specified Default is returned and it will use the serializer default tier.
func (t *PmsTagStruct) Tier() ParamTier {
	str := t.StructTagImpl.Named["tier"]
	if str == "" {
		return Default
	}

	return ParamTier(str)
}

// SsmTags Converts the StructTag.Tags into ssm version of Tags
func (t *PmsTagStruct) SsmTags() []ssm.Tag {

	tags := []ssm.Tag{}

	for key, value := range t.StructTagImpl.Tags {
		tags = append(tags, ssm.Tag{Key: aws.String(key), Value: aws.String(value)})
	}

	return tags
}

// SsmTier returns the tier or a default specified in the in-param
func (t *PmsTagStruct) SsmTier(defaultTier ssm.ParameterTier) ssm.ParameterTier {

	switch t.Tier() {
	case Default:
		return defaultTier
	case Std:
		return ssm.ParameterTierStandard
	case Adv:
		return ssm.ParameterTierAdvanced
	case Eval:
		return ssm.ParameterTierIntelligentTiering
	}

	return defaultTier
}

// Overwrite returns true (default) if it will overwrite parameter upon write
func (t *PmsTagStruct) Overwrite() bool {
	overwrite := false
	if prm, ok := t.StructTagImpl.Named["overwrite"]; ok {
		overwrite, _ = strconv.ParseBool(prm)
	}
	return overwrite
}

// Pattern returns a optional regular expression to validate the parameter value.
func (t *PmsTagStruct) Pattern() string { return t.StructTagImpl.Named["pattern"] }

// Description returns a  description describing the parameter (if any).
func (t *PmsTagStruct) Description() string { return t.StructTagImpl.Named["description"] }

// Prefix returns the prefix if any
func (t *PmsTagStruct) Prefix() string { return t.StructTagImpl.Named["prefix"] }

// Name returns the short name of the field
func (t *PmsTagStruct) Name() string { return t.StructTagImpl.Named["name"] }

// Tag is set of name values that is part of the Tags for the parameter
func (t *PmsTagStruct) Tag() map[string]string { return t.StructTagImpl.Tags }

// FqName is the full name including prefix
func (t *PmsTagStruct) FqName() string { return t.StructTagImpl.FullName }

// Secure returns true if this entry is backed by a encryption key
func (t *PmsTagStruct) Secure() bool { return t.StructTagImpl.Named["keyid"] != "" }

// DefaultAccountKey is for determine if the backing key is the account default KMS key
func (t *PmsTagStruct) DefaultAccountKey() bool { return t.StructTagImpl.Named["keyid"] == "default" }

// IsLocalKey returns true if the real arn to the key is registered in the encoder / decoder
func (t *PmsTagStruct) IsLocalKey() bool {
	return strings.HasPrefix(t.StructTagImpl.Named["keyid"], "local://")
}

// GetKeyName gets the keyname without namespaces for local, full for arn and default
// for such
func (t *PmsTagStruct) GetKeyName() string {
	if t.IsLocalKey() {
		return t.StructTagImpl.Named["keyid"][8:]
	}
	return t.StructTagImpl.Named["keyid"]
}
