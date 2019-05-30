package chain

import (
	"testing"

	"github.com/goharbor/harbor/src/common/retention"
	"github.com/goharbor/harbor/src/retention/filter"
	"github.com/stretchr/testify/require"
)

func TestBuild_Valid(t *testing.T) {
	meta := &retention.Policy{
		Filters: []*retention.FilterMetadata{
			{
				Type:    filter.TypeDeleteOlderThan,
				Options: map[string]interface{}{filter.MetaDataKeyN: 1},
			},
			{
				Type:    filter.TypeDeleteRegex,
				Options: map[string]interface{}{filter.MetaDataKeyMatch: ".*"},
			},
			{
				Type: filter.TypeKeepEverything,
			},
			{
				Type:    filter.TypeKeepMostRecentN,
				Options: map[string]interface{}{filter.MetaDataKeyN: 1},
			},
			{
				Type:    filter.TypeKeepRegex,
				Options: map[string]interface{}{filter.MetaDataKeyMatch: ".*"},
			},
			{
				Type: filter.TypeDeleteEverything,
			},
		},
	}

	result, err := Build(meta)

	require.NoError(t, err)
	require.Len(t, result, 6)
}

func TestBuild_UnknownType(t *testing.T) {
	meta := &retention.Policy{
		Filters: []*retention.FilterMetadata{
			{
				Type: "some_unknown_filter",
			},
		},
	}

	result, err := Build(meta)

	require.EqualError(t, err, "unknown filter type: some_unknown_filter")
	require.Empty(t, result)
}

func TestBuild_BuilderErr(t *testing.T) {
	meta := &retention.Policy{
		Filters: []*retention.FilterMetadata{
			{
				Type:    filter.TypeDeleteOlderThan,
				Options: map[string]interface{}{filter.MetaDataKeyN: -1},
			},
		},
	}

	result, err := Build(meta)

	require.EqualError(t, err, "filter: metadata: n: cannot be negative")
	require.Empty(t, result)
}
