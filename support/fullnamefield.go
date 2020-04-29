package support

import "reflect"

// FullNameField contains the full name both locally
// using field navigation and the remove name e.g. if
// pms it could be /prod/test-service/myproperty. It also
// includes field metadata including the reflect.Value.
type FullNameField struct {
	LocalName  string
	RemoteName string
	Field      reflect.StructField
	Value      reflect.Value
}
