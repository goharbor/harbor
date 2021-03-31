package githubcr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_native_Info(t *testing.T) {
	var registry = &model.Registry{URL: "abc"}
	adapter := newAdapter(registry)
	assert.NotNil(t, adapter)

	info, err := adapter.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, model.RegistryTypeGithubCR, info.Type)
	assert.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, 2, len(info.SupportedResourceFilters))
	assert.Equal(t, 2, len(info.SupportedTriggers))
	assert.Equal(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func Test_getAdapterPattern(t *testing.T) {
	var pattern = getAdapterPattern()
	assert.NotNil(t, pattern)
	assert.Equal(t, model.EndpointPatternTypeFix, pattern.EndpointPattern.EndpointType)
}

func mockGHCR() (mock *httptest.Server) {
	return test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/_catalog",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				r.Response.StatusCode = http.StatusNotFound
				w.Write([]byte(``))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/kofj/a1/tags/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"kofj/a1","tags":["tag11"]}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/kofj/b2/tags/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"kofj/b2","tags":["tag11","tag2","tag13"]}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/kofj/c3/l3/tags/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"kofj/c3/l3","tags":["tag4"]}`))
			},
		},
	)
}

func Test_native_FetchArtifacts(t *testing.T) {
	var mock = mockGHCR()
	defer mock.Close()
	fmt.Println("mockGHCR URL: ", mock.URL)

	var registry = &model.Registry{
		Type:     model.RegistryTypeDockerRegistry,
		URL:      mock.URL,
		Insecure: true,
	}
	adapter := newAdapter(registry)
	assert.NotNil(t, adapter)

	tests := []struct {
		name    string
		filters []*model.Filter
		want    []*model.Resource
		wantErr bool
	}{
		{
			name: "repository not exist",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "b1",
				},
			},
			wantErr: false,
		},
		{
			name: "tag not exist",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeTag,
					Value: "this_tag_not_exist_in_the_mock_server",
				},
			},
			wantErr: false,
		},
		{
			name:    "no filters",
			filters: []*model.Filter{},
			want:    []*model.Resource{},
			wantErr: true,
		},

		{
			name: "only special repository",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "kofj/a1",
				},
			},
			want: []*model.Resource{
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "kofj/a1"},
						Artifacts: []*model.Artifact{
							{
								Tags: []string{"tag11"},
							},
						},
					},
				},
			},
			wantErr: false,
		},

		{
			name: "only special tag",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeTag,
					Value: "tag11",
				},
			},
			want:    []*model.Resource{},
			wantErr: true,
		},

		{
			name: "special repository and special tag",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "kofj/b2",
				},
				{
					Type:  model.FilterTypeTag,
					Value: "tag2",
				},
			},
			want: []*model.Resource{
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "kofj/b2"},
						Artifacts: []*model.Artifact{
							{
								Tags: []string{"tag2"},
							},
						},
					},
				},
			},

			wantErr: false,
		},
		{
			name: "only wildcard repository",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "kofj/b*",
				},
			},
			want:    []*model.Resource{},
			wantErr: true,
		},
		{
			name: "only wildcard tag",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeTag,
					Value: "tag1*",
				},
			},
			want:    []*model.Resource{},
			wantErr: true,
		},
		{
			name: "wildcard repository and wildcard tag",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "kofj/b*",
				},
				{
					Type:  model.FilterTypeTag,
					Value: "tag1*",
				},
			},
			want:    []*model.Resource{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources, err = adapter.FetchArtifacts(tt.filters)
			if tt.wantErr {
				require.Len(t, resources, 0)
				require.NotNil(t, err)
			} else {
				if err != nil {
					t.Logf("Name=%s, error: %v", t.Name(), err)
				}
				require.Equal(t, len(tt.want), len(resources))
				for i, resource := range resources {
					require.NotNil(t, resource.Metadata)
					assert.Equal(t, tt.want[i].Metadata.Repository, resource.Metadata.Repository)
					assert.ElementsMatch(t, tt.want[i].Metadata.Artifacts, resource.Metadata.Artifacts)
				}
			}
		})
	}
}

// ***************************************
