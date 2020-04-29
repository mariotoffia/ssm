package support

import "reflect"

// FullNameField contains the full name both locally
// using field navigation and the remove name e.g. if
// pms it could be /prod/test-service/myproperty. It also
// includes field metadata including the reflect.Value.
type FullNameField struct {
	// Local name in dotted navigation format
	LocalName string
	// Remove name as required by AWS
	// (for PMS this is not a ARN)
	RemoteName string
	// The field within the struct that is reffered
	Field reflect.StructField
	// The value accessor to the field. Note if this is
	// a pointer; it may not have a value do check IsValid
	// before accessing.
	Value reflect.Value
}
