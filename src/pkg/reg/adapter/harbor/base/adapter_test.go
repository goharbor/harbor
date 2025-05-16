// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package base

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestGetAPIVersion(t *testing.T) {
	adapter := &Adapter{
		Client: &Client{APIVersion: "1.0"},
	}
	assert.Equal(t, "1.0", adapter.GetAPIVersion())
}

func TestInfo(t *testing.T) {
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/systeminfo",
		Handler: func(w http.ResponseWriter, r *http.Request) {},
	})
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter, err := New(registry)
	require.Nil(t, err)
	info, err := adapter.Info()
	require.Nil(t, err)
	assert.Equal(t, model.RegistryTypeHarbor, info.Type)
	assert.Equal(t, 3, len(info.SupportedResourceFilters))
	assert.Equal(t, 2, len(info.SupportedTriggers))
	assert.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
	assert.Equal(t, model.RepositoryPathComponentTypeAtLeastTwo, info.SupportedRepositoryPathComponentType)
	server.Close()

	server = test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/systeminfo",
		Handler: func(w http.ResponseWriter, r *http.Request) {},
	})
	registry = &model.Registry{
		URL: server.URL,
	}
	adapter, err = New(registry)
	require.Nil(t, err)
	info, err = adapter.Info()
	require.Nil(t, err)
	assert.Equal(t, model.RegistryTypeHarbor, info.Type)
	assert.Equal(t, 3, len(info.SupportedResourceFilters))
	assert.Equal(t, 2, len(info.SupportedTriggers))
	assert.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
	assert.Equal(t, true, info.SupportedCopyByChunk)
	server.Close()
}

func TestPrepareForPush(t *testing.T) {
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodPost,
		Pattern: "/api/projects",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		},
	},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/api/projects",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[]`))
			},
		},
	)
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter, err := New(registry)
	require.Nil(t, err)
	// nil resource
	err = adapter.PrepareForPush([]*model.Resource{nil})
	require.NotNil(t, err)
	// nil metadata
	err = adapter.PrepareForPush([]*model.Resource{
		{},
	})
	require.NotNil(t, err)
	// nil repository
	err = adapter.PrepareForPush(
		[]*model.Resource{
			{
				Metadata: &model.ResourceMetadata{},
			},
		})
	require.NotNil(t, err)
	// nil repository name
	err = adapter.PrepareForPush(
		[]*model.Resource{
			{
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{},
				},
			},
		})
	require.NotNil(t, err)
	// project doesn't exist
	err = adapter.PrepareForPush(
		[]*model.Resource{
			{
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: "library/hello-world",
					},
				},
			},
		})
	require.Nil(t, err)

	server.Close()

	// project already exists
	server = test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/projects",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"name": "library"}]`))
		},
	})
	registry = &model.Registry{
		URL: server.URL,
	}
	adapter, err = New(registry)
	require.Nil(t, err)
	err = adapter.PrepareForPush(
		[]*model.Resource{
			{
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: "library/hello-world",
					},
				},
			},
		})
	require.Nil(t, err)

	// project already exists and the type is proxy cache
	server = test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/projects",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"name": "library", "registry_id": 1}]`))
		},
	})
	registry = &model.Registry{
		URL: server.URL,
	}
	adapter, err = New(registry)
	require.Nil(t, err)
	resources := []*model.Resource{
		{
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
			},
		},
	}
	err = adapter.PrepareForPush(resources)
	require.Nil(t, err)
	require.True(t, resources[0].Skip)
}

func TestParsePublic(t *testing.T) {
	cases := []struct {
		metadata map[string]any
		result   bool
	}{
		{nil, false},
		{map[string]any{}, false},
		{map[string]any{"public": true}, true},
		{map[string]any{"public": "not_bool"}, false},
		{map[string]any{"public": "true"}, true},
		{map[string]any{"public": struct{}{}}, false},
	}
	for _, c := range cases {
		assert.Equal(t, c.result, parsePublic(c.metadata))
	}
}

func TestMergeMetadata(t *testing.T) {
	cases := []struct {
		m1     map[string]any
		m2     map[string]any
		public bool
	}{
		{
			m1: map[string]any{
				"public": "true",
			},
			m2: map[string]any{
				"public": "true",
			},
			public: true,
		},
		{
			m1: map[string]any{
				"public": "false",
			},
			m2: map[string]any{
				"public": "true",
			},
			public: false,
		},
		{
			m1: map[string]any{
				"public": "false",
			},
			m2: map[string]any{
				"public": "false",
			},
			public: false,
		},
	}
	for _, c := range cases {
		m := mergeMetadata(c.m1, c.m2)
		assert.Equal(t, strconv.FormatBool(c.public), m["public"].(string))
	}
}

func TestAbstractPublicMetadata(t *testing.T) {
	// nil input metadata
	meta := abstractPublicMetadata(nil)
	assert.Nil(t, meta)

	// contains no public metadata
	metadata := map[string]any{
		"other": "test",
	}
	meta = abstractPublicMetadata(metadata)
	assert.Nil(t, meta)

	// contains public metadata
	metadata = map[string]any{
		"other":  "test",
		"public": "true",
	}
	meta = abstractPublicMetadata(metadata)
	require.NotNil(t, meta)
	require.Equal(t, 1, len(meta))
	require.Equal(t, "true", meta["public"].(string))
}

func TestListProjects(t *testing.T) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/api/projects",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[{"name": "p1"}, {"name": "p2"}]`))
			},
		},
	)

	defer server.Close()

	registry := &model.Registry{
		URL: server.URL,
	}
	adapter, err := New(registry)
	require.Nil(t, err)

	validPattern := "{p1,p2}/**"
	// has " " in the p2 project name
	invalidPattern := "{p1, p2}/**"
	filters := []*model.Filter{
		{
			Type:  "name",
			Value: validPattern,
		},
	}
	projects, err := adapter.ListProjects(filters)
	require.Nil(t, err)
	require.Len(t, projects, 2)
	require.Equal(t, "p1", projects[0].Name)
	require.Equal(t, "p2", projects[1].Name)

	// invalid pattern, should also work with trim white space in project name.
	filters[0].Value = invalidPattern
	_, err = adapter.ListProjects(filters)
	require.Nil(t, err)
	require.Len(t, projects, 2)
	require.Equal(t, "p1", projects[0].Name)
	require.Equal(t, "p2", projects[1].Name)
}
