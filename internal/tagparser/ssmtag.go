package tagparser

import (
	"fmt"
	"strings"
)

// StoreType is the type of storage backed in AWS
type StoreType int

const (
	// Pms is Parameter Store
	Pms StoreType = iota
	// Asm is AWS Secrets Manager
	Asm
)

// SsmTag is used to encapsulate both PmsTag and AsmTag
type SsmTag interface {
	SsmType() StoreType
	Prefix() string
	Name() string
	Tags() map[string]string
	FullName() string
	Secure() bool
}

// PmsTag is for AWS parameter store
type PmsTag struct {
	// Overrides the prefix of the path to the name. Default is /service name/name of parameter.
	// It does not, however, override the environment prefix such as /dev/... /prod/... etc.
	prefix string
	// name of the entry - this is never a fully qualified.
	name string
	// If encrypted, which key id (registered with the encoder) should
	// be provided to parameter store when created. This will resolve to
	// a arn to be used when manageing encrypted entries. If ARN is specified
	// it will use that explicit ARN instead. If id is default it uses
	// the default account KMS key to encrypt the value. All local registered
	// keys (with the encoder / decoder) begins with local://
	keyID string
	// All key values that do not have a special meaning will end up as tags
	tags map[string]string
}

// SsmType returns Pms
func (t *PmsTag) SsmType() StoreType { return Pms }

// Prefix returns the prefix if any
func (t *PmsTag) Prefix() string { return t.prefix }

// Name returns the short name of the field
func (t *PmsTag) Name() string { return t.name }

// Tags is set when a standard field with a PmsTag
func (t *PmsTag) Tags() map[string]string { return t.tags }

// FullName is the full name including prefix
func (t *PmsTag) FullName() string { return fmt.Sprintf("%s/%s", t.prefix, t.name) }

// Secure returns true if this entry is backed by a encryption key
func (t *PmsTag) Secure() bool { return t.keyID != "" }

// DefaultAccountKey is for determine if the backing key is the account default KMS key
func (t *PmsTag) DefaultAccountKey() bool { return t.keyID == "default" }

// IsLocalKey returns true if the real arn to the key is registered in the encoder / decoder
func (t *PmsTag) IsLocalKey() bool { return strings.HasPrefix(t.keyID, "local://") }

// GetKeyName gets the keyname without namespaces for local, full for arn and default-account
// for such
func (t *PmsTag) GetKeyName() string {
	if t.IsLocalKey() {
		return t.keyID[8:]
	}
	return t.keyID
}

// AsmTag encapsulates a AWS Secrets Manager backed storage
type AsmTag struct {
	// Overrides the prefix of the path to the name. Default is /service name/name of parameter.
	// It does not, however, override the environment prefix such as /dev/... /prod/... etc.
	prefix string
	// The name of the secret
	name string
	// (Optional) Specifies the ARN, Key ID, or alias of the AWS KMS customer master
	// key (CMK) to be used to encrypt the SecretString or SecretBinary values in
	// the versions stored in this secret	keyID string
	// All key values that do not have a special meaning will end up as tags
	// If you don't specify this value, then Secrets Manager defaults to using the
	// AWS account's default CMK (the one named aws/secretsmanager).
	// If id is default it uses the default account KMS key to encrypt the value.
	// All local registered keys (with the encoder / decoder) begins with local://
	// and the name of the key to be resolved to a ARN, Key ID, or alias.
	keyID string
	tags  map[string]string
}

// SsmType returns Asm
func (t *AsmTag) SsmType() StoreType { return Asm }

// Prefix returns the prefix if any
func (t *AsmTag) Prefix() string { return t.prefix }

// Name returns the short name of the field
func (t *AsmTag) Name() string { return t.name }

// Tags is set when a standard field with a AsmTag
func (t *AsmTag) Tags() map[string]string { return t.tags }

// FullName is the full name including prefix
func (t *AsmTag) FullName() string { return fmt.Sprintf("%s/%s", t.prefix, t.name) }

// Secure returns true if this entry is backed by a encryption key
func (t *AsmTag) Secure() bool { return true }

// DefaultAccountKey is for determine if the backing key is the account default KMS key
func (t *AsmTag) DefaultAccountKey() bool { return t.keyID == "default" }

// IsLocalKey returns true if the real arn to the key is registered in the encoder / decoder
func (t *AsmTag) IsLocalKey() bool { return strings.HasPrefix(t.keyID, "local://") }

// GetKeyName gets the keyname without namespaces for local, full for arn and default-account
// for such
func (t *AsmTag) GetKeyName() string {
	if t.IsLocalKey() {
		return t.keyID[8:]
	}
	return t.keyID
}
