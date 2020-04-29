package support

import "strings"

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

// FieldFilters contains all filters to be applied
// on an operation
type FieldFilters struct {
	// Zero or more filters
	Filters []FieldFilter
}

// NewFilters creates a empty filter set. Use Methods Exclude and Include
func NewFilters() *FieldFilters {
	filters := []FieldFilter{}
	return &FieldFilters{Filters: filters}
}

// NewIncludeExcludeFilters creates a new filter set that may be used to apply to filtering
func NewIncludeExcludeFilters(includes []string, excludes []string) *FieldFilters {
	filters := []FieldFilter{}

	if nil == includes && nil == excludes {
		return &FieldFilters{Filters: filters}
	}

	for _, include := range includes {
		filters = append(filters, FieldFilter{Name: include, Action: Include})
	}

	for _, exclude := range excludes {
		filters = append(filters, FieldFilter{Name: exclude, Action: Exclude})
	}

	return &FieldFilters{Filters: filters}
}

// Exclude adds an exclusion filter to the set
func (f *FieldFilters) Exclude(path string) *FieldFilters {
	if len(path) > 0 {
		f.Filters = append(f.Filters, FieldFilter{Name: path, Action: Exclude})
	}
	return f
}

// Include adds an inclusion filter to the set
func (f *FieldFilters) Include(path string) *FieldFilters {
	if len(path) > 0 {
		f.Filters = append(f.Filters, FieldFilter{Name: path, Action: Include})
	}
	return f
}

// Excludes gives all Exclude filters in the set
func (f *FieldFilters) Excludes() []FieldFilter {
	filters := []FieldFilter{}
	for _, f := range f.Filters {
		if f.Action == Exclude {
			filters = append(filters, f)
		}
	}

	return filters
}

// Includes gives all Include filters in the set
func (f *FieldFilters) Includes() []FieldFilter {
	filters := []FieldFilter{}
	for _, f := range f.Filters {
		if f.Action == Include {
			filters = append(filters, f)
		}
	}

	return filters
}

// IsIncluded checks if the fqpath to a field is not excluded from the filter.
// Note that an exclusion may be done on a higer position in the tree but may
// be overridden by either a explicit include or a inclusion node below the
// exclusion but higer up than the field specified by fqpath
func (f *FieldFilters) IsIncluded(fqpath string) bool {
	included := true
	currlen := 0
	for _, x := range f.Filters {
		if strings.HasPrefix(fqpath, x.Name) {
			locallen := len(x.Name)
			if locallen > currlen {
				if x.Action == Include {
					included = true
				} else {
					included = false
				}
			}
			currlen = locallen
		}
	}

	return included
}
