package support

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilIncludeExcludeGivesZeroFilters(t *testing.T) {
	filters := NewIncludeExcludeFilters(nil, nil)
	assert.Equal(t, 0, len(filters.Filters))
}

func TestNilIncludeOneExcludeGivesOneExclude(t *testing.T) {
	filters := NewIncludeExcludeFilters(nil, []string{"Name"})
	assert.Equal(t, 1, len(filters.Filters))
	assert.Equal(t, "Name", filters.Filters[0].Name)
	assert.Equal(t, Exclude, filters.Filters[0].Action)
}

func TestNilExcludeOneIncludeGivesOneIncludeFilter(t *testing.T) {
	filters := NewIncludeExcludeFilters([]string{"Name"}, nil)
	assert.Equal(t, 1, len(filters.Filters))
	assert.Equal(t, "Name", filters.Filters[0].Name)
	assert.Equal(t, Include, filters.Filters[0].Action)
}

func TestOneExcludeOneIncludeGivesOneIncludeFilterAndOneExcludeFilter(t *testing.T) {
	filters := NewIncludeExcludeFilters([]string{"Name"}, []string{"Sub"})
	includes := filters.Includes()
	excludes := filters.Excludes()

	assert.Equal(t, 1, len(includes))
	assert.Equal(t, "Name", includes[0].Name)
	assert.Equal(t, Include, includes[0].Action)
	assert.Equal(t, 1, len(excludes))
	assert.Equal(t, "Sub", excludes[0].Name)
	assert.Equal(t, Exclude, excludes[0].Action)
}

func TestDefaultIsInclude(t *testing.T) {
	filters := NewFilters()

	assert.Equal(t, 0, len(filters.Filters))
	assert.Equal(t, true, filters.IsIncluded("Sub.Apa"))
}

func TestExplicitExclusionShallExclude(t *testing.T) {
	filters := NewFilters().Exclude("Sub.Apa")

	assert.Equal(t, 1, len(filters.Filters))
	assert.Equal(t, false, filters.IsIncluded("Sub.Apa"))
}

func TestExplicitInclusionShallExclude(t *testing.T) {
	filters := NewFilters().Include("Sub.Apa")

	assert.Equal(t, 1, len(filters.Filters))
	assert.Equal(t, true, filters.IsIncluded("Sub.Apa"))
}

func TestExcludeIsOverridedByIncludeDownTheChain(t *testing.T) {
	filters := NewFilters().Exclude("Sub").Include("Sub.Bobban")

	assert.Equal(t, 2, len(filters.Filters))
	assert.Equal(t, true, filters.IsIncluded("Sub.Bobban.Apa"))
}

func TestIncludeIsOverridedByExcludeDownTheChain(t *testing.T) {
	filters := NewFilters().Exclude("Sub.Bobban").Include("Sub")

	assert.Equal(t, 2, len(filters.Filters))
	assert.Equal(t, false, filters.IsIncluded("Sub.Bobban.Apa"))
}

func TestExcludeIsOverridedByExplicitInclude(t *testing.T) {
	filters := NewFilters().Exclude("Sub").Include("Sub.Bobban.Apa")

	assert.Equal(t, 2, len(filters.Filters))
	assert.Equal(t, true, filters.IsIncluded("Sub.Bobban.Apa"))
}

func TestIncludeIsOverridedByExplicitExclude(t *testing.T) {
	filters := NewFilters().Include("Sub").Exclude("Sub.Bobban.Apa")

	assert.Equal(t, 2, len(filters.Filters))
	assert.Equal(t, false, filters.IsIncluded("Sub.Bobban.Apa"))
}

func TestExcludeIsNotAffectingOtherBranch(t *testing.T) {
	filters := NewFilters().Include("Sub").Exclude("Sub.Bobban")

	assert.Equal(t, 2, len(filters.Filters))
	assert.Equal(t, true, filters.IsIncluded("Sub.Bubbu.Nisse.Apa"))
}

func TestIncludeIsOverridedByVaribleUpstreamExcludeButBelowInclude(t *testing.T) {
	filters := NewFilters().Include("Sub").Exclude("Sub.Bobban")

	assert.Equal(t, 2, len(filters.Filters))
	assert.Equal(t, false, filters.IsIncluded("Sub.Bobban.Bubben.Bibbi.Apa"))
}
