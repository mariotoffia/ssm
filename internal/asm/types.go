package asm

import (
	"strings"

	"github.com/mariotoffia/ssm.git/parser"
)

// ParamTier specifies the parameter tier such as std, adv, or intelligent.
type ParamTier string

// ToAsmTag converts a StructTag into the AsmTag struct
// If fails, it return false.
func ToAsmTag(generictag *parser.StructNode) (*AsmTagStruct, bool) {
	if ntag, ok := generictag.Tag["asm"]; ok {
		return ntag.(*AsmTagStruct), true
	}
	return nil, false
}

// AsmTag is a generic interface
type AsmTag interface {
	parser.StructTag
	Prefix() string
	Name() string
	Tags() map[string]string
	FqName() string
	Secure() bool
	Description() string
	DefaultAccountKey() bool
	IsLocalKey() bool
	GetKeyName() string
	StringKey() string
	VersionStage() string
	VersionID() string
}

// AsmTagStruct is for AWS secets manager
type AsmTagStruct struct {
	// Extend StructTagImpl
	parser.StructTagImpl
}

// VersionID specifies the unique identifier of the version of the secret.
func (t *AsmTagStruct) VersionID() string { return t.StructTagImpl.Named["vid"] }

// VersionStage specifies the secret version that you want to retrieve by the staging label
// attached to the version. If no versionStage or versionID is specified AWSCURRENT
// as versionStage is used
func (t *AsmTagStruct) VersionStage() string { return t.StructTagImpl.Named["vs"] }

// StringKey is the name of the element in the JSON payload in value where secrets
// manager shall generate it's password into. This is done in creation time. The
// complete secret string is then encrypted. If this is nil, no generation is wanted.
// This is not used by the ssm go implementation. Instead this is for cloud formation
// that have the ability to generate a password upon deployment.
func (t *AsmTagStruct) StringKey() string { return t.StructTagImpl.Named["strkey"] }

// Description returns a  description describing the parameter (if any).
func (t *AsmTagStruct) Description() string { return t.StructTagImpl.Named["description"] }

// Prefix returns the prefix if any
func (t *AsmTagStruct) Prefix() string { return t.StructTagImpl.Named["prefix"] }

// Name returns the short name of the field
func (t *AsmTagStruct) Name() string { return t.StructTagImpl.Named["name"] }

// Tag is set of name values that is part of the Tags for the parameter
func (t *AsmTagStruct) Tag() map[string]string { return t.StructTagImpl.Tags }

// FqName is the full name including prefix
func (t *AsmTagStruct) FqName() string { return t.StructTagImpl.FullName }

// Secure returns true if this entry is backed by a encryption key
func (t *AsmTagStruct) Secure() bool { return t.StructTagImpl.Named["keyid"] != "" }

// DefaultAccountKey is for determine if the backing key is the account default KMS key
func (t *AsmTagStruct) DefaultAccountKey() bool { return t.StructTagImpl.Named["keyid"] == "default" }

// IsLocalKey returns true if the real arn to the key is registered in the encoder / decoder
func (t *AsmTagStruct) IsLocalKey() bool {
	return strings.HasPrefix(t.StructTagImpl.Named["keyid"], "local://")
}

// GetKeyName gets the keyname without namespaces for local, full for arn and default
// for such
func (t *AsmTagStruct) GetKeyName() string {
	if t.IsLocalKey() {
		return t.StructTagImpl.Named["keyid"][8:]
	}
	return t.StructTagImpl.Named["keyid"]
}
