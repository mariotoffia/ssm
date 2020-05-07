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

// SsmTag is used to encapsulate both PmsTag and AsmTag
type SsmTag interface {
	SsmType() StoreType
	Prefix() string
	Name() string
	Tags() map[string]string
	FullName() string
	Secure() bool
	Description() string
	DefaultAccountKey() bool
	IsLocalKey() bool
	GetKeyName() string
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
	// for the value
	tags map[string]string
	// An optional regular expression to validate the parameter value. E.g. for integer ^\d+$
	pattern string
	// An optional description describing the parameter.
	description string
	// If set to true (default) it will overwrite in a create operation - hence upsert
	overwrite bool
	// If not set it will use "std" (free) by default. Otherwise choose from adv(anced) or
	// eval (Intelligent-Tiering)
	tier ParamTier
}

// Tier specifies the parameter tier it may be of std, adv, and eval (intelligent tiering).
// If nothing is specified Default is returned and it will use the serializer default tier.
func (t *PmsTag) Tier() ParamTier {
	if t.tier == "" {
		return Default
	}

	return t.tier
}

// Overwrite returns true (default) if it will overwrite parameter upon write
func (t *PmsTag) Overwrite() bool { return t.overwrite }

// Pattern returns a optional regular expression to validate the parameter value.
func (t *PmsTag) Pattern() string { return t.pattern }

// Description returns a  description describing the parameter (if any).
func (t *PmsTag) Description() string { return t.description }

// SsmType returns Pms
func (t *PmsTag) SsmType() StoreType { return Pms }

// Prefix returns the prefix if any
func (t *PmsTag) Prefix() string { return t.prefix }

// Name returns the short name of the field
func (t *PmsTag) Name() string { return t.name }

// Tags is set of name values that is part of the Tags for the parameter
func (t *PmsTag) Tags() map[string]string { return t.tags }

// FullName is the full name including prefix
func (t *PmsTag) FullName() string { return fmt.Sprintf("%s/%s", t.prefix, t.name) }

// Secure returns true if this entry is backed by a encryption key
func (t *PmsTag) Secure() bool { return t.keyID != "" }

// DefaultAccountKey is for determine if the backing key is the account default KMS key
func (t *PmsTag) DefaultAccountKey() bool { return t.keyID == "default" }

// IsLocalKey returns true if the real arn to the key is registered in the encoder / decoder
func (t *PmsTag) IsLocalKey() bool { return strings.HasPrefix(t.keyID, "local://") }

// GetKeyName gets the keyname without namespaces for local, full for arn and default
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
	// This value is typically a UUID-type (https://wikipedia.org/wiki/Universally_unique_identifier)
	// value with 32 hexadecimal digits. Specifies the unique identifier of the version of the secret.
	versionID string
	// Specifies the secret version that you want to retrieve by the staging label
	// attached to the version. If no versionStage or versionID is specified AWSCURRENT
	// as versionStage is used
	versionStage string
	// All key values that do not have a special meaning will end up as tags
	// for the value
	tags map[string]string
	// The description for this secret (if any)
	description string
	// If the value of this secret is a JSON payload. This identifies where in the JSON payload
	// the secrets manager shall genereate the password into. For example {"user":"nisse"}
	// and this property is password. Then the secrets manager will generate password into the "password"
	// property. The resulting secret string is then {"user":"nisse", "password":"12djkscnji "}.If this is
	// not set,  nothing is generated instead the data is only ecrypted / decrypted.
	stringkey string
}

// StringKey is the name of the element in the JSON payload in value where secrets
// manager shall generate it's password into. This is done in creation time. The
// complete secret string is then encrypted. If this is nil, no generation is wanted.
// This is not used by the ssm go implementation. Instead this is for cloud formation
// that have the ability to generate a password upon deployment.
func (t *AsmTag) StringKey() string { return t.stringkey }

// Description returns the description for this secret (if any)
func (t *AsmTag) Description() string { return t.description }

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

// VersionID specifies the unique identifier of the version of the secret.
func (t *AsmTag) VersionID() string { return t.versionID }

// VersionStage specifies the secret version that you want to retrieve by the staging label
// attached to the version. If no versionStage or versionID is specified AWSCURRENT
// as versionStage is used
func (t *AsmTag) VersionStage() string { return t.versionStage }

// DefaultAccountKey is for determine if the backing key is the account default KMS key
func (t *AsmTag) DefaultAccountKey() bool { return t.keyID == "default" }

// IsLocalKey returns true if the real arn to the key is registered in the encoder / decoder
func (t *AsmTag) IsLocalKey() bool { return strings.HasPrefix(t.keyID, "local://") }

// GetKeyName gets the keyname without namespaces for local, full for arn and default
// for such
func (t *AsmTag) GetKeyName() string {
	if t.IsLocalKey() {
		return t.keyID[8:]
	}
	return t.keyID
}
