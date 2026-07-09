package filter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestRepositoryNameFilters(t *testing.T) {
	repositories := []*model.Repository{
		{Name: "library/nginx"},
		{Name: "library/redis"},
		{Name: "jar/app"},
		{Name: "demo/app"},
	}

	filters := []*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "library/**",
		},
	}

	repoFilters, err := BuildRepositoryFilters(filters)
	require.NoError(t, err)

	repos, err := repoFilters.Filter(repositories)
	require.NoError(t, err)
	require.Len(t, repos, 2)
	require.Equal(t, "library/nginx", repos[0].Name)
	require.Equal(t, "library/redis", repos[1].Name)

	filters = []*model.Filter{
		{
			Type:       model.FilterTypeName,
			Value:      "jar/**",
			Decoration: model.Excludes,
		},
	}

	repoFilters, err = BuildRepositoryFilters(filters)
	require.NoError(t, err)

	repos, err = repoFilters.Filter(repositories)
	require.NoError(t, err)
	require.Len(t, repos, 3)
	require.Equal(t, "library/nginx", repos[0].Name)
	require.Equal(t, "library/redis", repos[1].Name)
	require.Equal(t, "demo/app", repos[2].Name)
}

func TestRepositoryNameFilterValidate(t *testing.T) {
	filter := &model.Filter{
		Type:       model.FilterTypeName,
		Value:      "jar/**",
		Decoration: model.Excludes,
	}
	require.NoError(t, filter.Validate())

	filter = &model.Filter{
		Type:  model.FilterTypeResource,
		Value: model.ResourceTypeImage,
		Decoration: model.Excludes,
	}
	require.Error(t, filter.Validate())
}
