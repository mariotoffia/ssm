package ssm

import "fmt"

type ssmType int

const (
	// Parameter Store
	pms ssmType = iota
	// Secrets Manager
	asm
)

type ssmTag interface {
	SsmType() ssmType
	Prefix() string
	Name() string
	Tags() map[string]string
	FullName() string
	Secure() bool
}

// pmsTag is for AWS parameter store
type pmsTag struct {
	// Overrides the prefix of the path to the name. Default is /service name/name of parameter.
	// It does not, however, override the environment prefix such as /dev/... /prod/... etc.
	prefix string
	// name of the entry - this is never a fully qualified.
	name string
	// If encrypted, which key id (registered with the encoder) should
	// be provided to parameter store when created. This will resolve to
	// a arn to be used when manageing encrypted entries. If ARN is specified
	// it will use that explicit ARN instead.
	keyID string
	// All key values that do not have a special meaning will end up as tags
	tags map[string]string
}

func (t pmsTag) SsmType() ssmType        { return pms }
func (t pmsTag) Prefix() string          { return t.prefix }
func (t pmsTag) Name() string            { return t.name }
func (t pmsTag) Tags() map[string]string { return t.tags }
func (t pmsTag) FullName() string        { return fmt.Sprintf("%s/%s", t.prefix, t.name) }
func (t pmsTag) Secure() bool            { return t.keyID != "" }

type asmTag struct {
	// Overrides the prefix of the path to the name. Default is /service name/name of parameter.
	// It does not, however, override the environment prefix such as /dev/... /prod/... etc.
	prefix string
	// The name of the secret
	name string
	// All key values that do not have a special meaning will end up as tags
	tags map[string]string
}

func (t asmTag) SsmType() ssmType        { return asm }
func (t asmTag) Prefix() string          { return t.prefix }
func (t asmTag) Name() string            { return t.name }
func (t asmTag) Tags() map[string]string { return t.tags }
func (t asmTag) FullName() string        { return fmt.Sprintf("%s/%s", t.prefix, t.name) }
func (t asmTag) Secure() bool            { return true }
