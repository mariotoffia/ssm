package ssm

import (
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/mariotoffia/ssm.git/internal/pms"
	"github.com/mariotoffia/ssm.git/internal/reflectparser"
)

// Serializer handles un-/marshaling of SSM data
// back and forth go struct fields.
type Serializer struct {
	config  aws.Config
	service string
	env     string
}

// Unmarshal creates the inparam struct pointer (and sub structs as well).
// It will populate the fields that are denoted with pms and asm
// with data from the Systems Manager.
func (s *Serializer) Unmarshal(v interface{}) error {
	tp := reflect.ValueOf(v)
	node, err := reflectparser.New(s.env, s.service).Parse("", tp)
	if err != nil {
		return err
	}

	pmsr, err := pms.New(s.service)
	if err != nil {
		return err
	}

	_, err = pmsr.Get(&node)
	return err
}
