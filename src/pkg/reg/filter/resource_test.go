package filter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/model"
)


func TestArtifactTagRegexFilters(t *testing.T) {
	var artifacts = []*model.Artifact{
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "aaaaa",
			Tags: []string{
				"test1",
				"test2",
				"harbor1",
				"1.8.0",
			},
		},
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "bbbbb",
			Tags: []string{
				"test3",
				"harbor2",
			},
		},
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "ccccc",
			Tags: []string{
				"harbor3",
				"1.9.01",
			},
		},
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "ddddd",
		},
	}

	var filters = []*model.Filter{

		//The following pattern is from doublestar testing
		//It works but number of filtered artifacts is mismatched
		//So the test validating below will fail
		//TODO: Remove following comment block
		// {
		// 	Type:  model.FilterTypeTagRegex,
		// 	Value: "test.*",
		// },

		{
			Type:  model.FilterTypeTagRegex,
			Value: `^((\d\d?).(\d\d?).(\d\d?(-stable)?))$`,
		},

	}

	artFilters, err := BuildArtifactFilters(filters)
	require.Nil(t, err)

	arts, err := artFilters.Filter(artifacts)
	require.Nil(t, err)
	require.Equal(t, 2, len(arts))
	require.EqualValues(t, "aaaaa", arts[0].Digest)
	fmt.Println(arts[0].Tags)
	require.EqualValues(t, []string{"1.8.0"}, arts[0].Tags)
	require.EqualValues(t, "ccccc", arts[1].Digest)
	require.EqualValues(t, []string{"1.9.01"}, arts[1].Tags)

	filters = []*model.Filter{
		{
			Type:       model.FilterTypeTagRegex,
			Value:      "test*",
			Decoration: model.Excludes,
		},
	}

	artFilters, err = BuildArtifactFilters(filters)
	require.Nil(t, err)

	arts, err = artFilters.Filter(artifacts)
	require.Nil(t, err)
	require.Equal(t, 4, len(arts))
	require.EqualValues(t, "aaaaa", arts[0].Digest)
	require.EqualValues(t, []string{"harbor1", "1.8.0"}, arts[0].Tags)
	require.EqualValues(t, "bbbbb", arts[1].Digest)
	require.EqualValues(t, []string{"harbor2"}, arts[1].Tags)
	require.EqualValues(t, "ccccc", arts[2].Digest)
	require.EqualValues(t, []string{"harbor3", "1.9.01"}, arts[2].Tags)
	require.EqualValues(t, "ddddd", arts[3].Digest)
	require.Nil(t, arts[3].Tags)
}

func TestArtifactTagFilters(t *testing.T) {
	var artifacts = []*model.Artifact{
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "aaaaa",
			Tags: []string{
				"test1",
				"test2",
				"harbor1",
			},
		},
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "bbbbb",
			Tags: []string{
				"test3",
				"harbor2",
			},
		},
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "ccccc",
			Tags: []string{
				"harbor3",
			},
		},
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "ddddd",
		},
	}

	var filters = []*model.Filter{
		{
			Type:  model.FilterTypeTag,
			Value: "test*",
		},
	}

	artFilters, err := BuildArtifactFilters(filters)
	require.Nil(t, err)

	arts, err := artFilters.Filter(artifacts)
	require.Nil(t, err)
	require.Equal(t, 2, len(arts))
	require.EqualValues(t, "aaaaa", arts[0].Digest)
	require.EqualValues(t, []string{"test1", "test2"}, arts[0].Tags)
	require.EqualValues(t, "bbbbb", arts[1].Digest)
	require.EqualValues(t, []string{"test3"}, arts[1].Tags)

	filters = []*model.Filter{
		{
			Type:       model.FilterTypeTag,
			Value:      "test*",
			Decoration: model.Excludes,
		},
	}

	artFilters, err = BuildArtifactFilters(filters)
	require.Nil(t, err)

	arts, err = artFilters.Filter(artifacts)
	require.Nil(t, err)
	require.Equal(t, 4, len(arts))
	require.EqualValues(t, "aaaaa", arts[0].Digest)
	require.EqualValues(t, []string{"harbor1"}, arts[0].Tags)
	require.EqualValues(t, "bbbbb", arts[1].Digest)
	require.EqualValues(t, []string{"harbor2"}, arts[1].Tags)
	require.EqualValues(t, "ccccc", arts[2].Digest)
	require.EqualValues(t, []string{"harbor3"}, arts[2].Tags)
	require.EqualValues(t, "ddddd", arts[3].Digest)
	require.Nil(t, arts[3].Tags)
}

func TestArtifactLabelFilters(t *testing.T) {
	var artifacts = []*model.Artifact{
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "aaaaa",
			Tags: []string{
				"test1",
				"test2",
				"harbor1",
			},
			Labels: []string{
				"label1",
			},
		},
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "bbbbb",
			Tags: []string{
				"test3",
				"harbor2",
			},
			Labels: []string{
				"label1",
				"label2",
			},
		},
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "ccccc",
			Tags: []string{
				"harbor3",
			},
			Labels: []string{
				"label3",
			},
		},
		{
			Type:   model.ResourceTypeArtifact,
			Digest: "ddddd",
		},
	}

	var filters = []*model.Filter{
		{
			Type:  model.FilterTypeLabel,
			Value: []string{"label1"},
		},
	}

	artFilters, err := BuildArtifactFilters(filters)
	require.Nil(t, err)

	arts, err := artFilters.Filter(artifacts)
	require.Nil(t, err)
	require.Equal(t, 2, len(arts))
	require.EqualValues(t, "aaaaa", arts[0].Digest)
	require.EqualValues(t, []string{"label1"}, arts[0].Labels)
	require.EqualValues(t, "bbbbb", arts[1].Digest)
	require.EqualValues(t, []string{"label1", "label2"}, arts[1].Labels)

	filters = []*model.Filter{
		{
			Type:       model.FilterTypeLabel,
			Value:      []string{"label1"},
			Decoration: model.Excludes,
		},
	}

	artFilters, err = BuildArtifactFilters(filters)
	require.Nil(t, err)

	arts, err = artFilters.Filter(artifacts)
	require.Nil(t, err)
	require.Equal(t, 2, len(arts))
	require.EqualValues(t, "ccccc", arts[0].Digest)
	require.EqualValues(t, []string{"label3"}, arts[0].Labels)
	require.EqualValues(t, "ddddd", arts[1].Digest)
	require.Nil(t, arts[1].Labels)
}
