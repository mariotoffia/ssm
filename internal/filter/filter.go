package filter

// Action states which action to use when
// filtering (include / exclude)
type Action int

const (
	// Include states that it shall be included
	Include Action = iota
	// Exclude states that it shall be excluded
	Exclude
)

// FieldFilter is a filter that expresses inclusion or exclusion per field basis
// When a field is e.g. a sub-struct the exclusion or inclusion will be for the
// whole subgraph. The field navigation is based on the field name and sub-structs
// are navigate by a dot e.g. Address.Street would imply that Street is part of a sub
// struct on parent struct called Address.
type FieldFilter struct {
	// Name of the field. Use dot (.) to
	// navigate sub-struct(s).
	Name string
	// Action states which action include / exclude it should do
	Action Action
}
