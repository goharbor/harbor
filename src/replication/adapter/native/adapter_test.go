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

package native

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newAdapter(t *testing.T) {
	tests := []struct {
		name     string
		registry *model.Registry
		wantErr  bool
	}{
		{name: "Nil Registry URL", registry: &model.Registry{}, wantErr: true},
		{name: "Right", registry: &model.Registry{URL: "abc"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAdapter(tt.registry)
			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Nil(t, got)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func Test_native_Info(t *testing.T) {
	var registry = &model.Registry{URL: "abc"}
	adapter, err := NewAdapter(registry)
	require.Nil(t, err)
	assert.NotNil(t, adapter)

	info, err := adapter.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, model.RegistryTypeDockerRegistry, info.Type)
	assert.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, 2, len(info.SupportedResourceFilters))
	assert.Equal(t, 2, len(info.SupportedTriggers))
	assert.Equal(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func Test_native_PrepareForPush(t *testing.T) {
	var registry = &model.Registry{URL: "abc"}
	adapter, err := NewAdapter(registry)
	require.Nil(t, err)
	assert.NotNil(t, adapter)

	err = adapter.PrepareForPush(nil)
	assert.Nil(t, err)
}

func mockNativeRegistry() (mock *httptest.Server) {
	return test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/_catalog",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"repositories":["test/a1","test/b2","test/c3/3level"]}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/test/a1/tags/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"test/a1","tags":["tag11"]}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/test/b2/tags/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"test/b2","tags":["tag11","tag2","tag13"]}`))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/test/c3/3level/tags/list",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"test/c3/3level","tags":["tag4"]}`))
			},
		},
	)
}
func Test_native_FetchImages(t *testing.T) {
	var mock = mockNativeRegistry()
	defer mock.Close()
	fmt.Println("mockNativeRegistry URL: ", mock.URL)

	var registry = &model.Registry{
		Type:     model.RegistryTypeDockerRegistry,
		URL:      mock.URL,
		Insecure: true,
	}
	adapter, err := NewAdapter(registry)
	assert.Nil(t, err)
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
			want: []*model.Resource{
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/a1"},
						Vtags:      []string{"tag11"},
					},
				},
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/b2"},
						Vtags:      []string{"tag11", "tag2", "tag13"},
					},
				},
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/c3/3level"},
						Vtags:      []string{"tag4"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "only special repository",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "test/a1",
				},
			},
			want: []*model.Resource{
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/a1"},
						Vtags:      []string{"tag11"},
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
			want: []*model.Resource{
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/a1"},
						Vtags:      []string{"tag11"},
					},
				},
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/b2"},
						Vtags:      []string{"tag11"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "special repository and special tag",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "test/b2",
				},
				{
					Type:  model.FilterTypeTag,
					Value: "tag2",
				},
			},
			want: []*model.Resource{
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/b2"},
						Vtags:      []string{"tag2"},
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
					Value: "test/b*",
				},
			},
			want: []*model.Resource{
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/b2"},
						Vtags:      []string{"tag11", "tag2", "tag13"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "only wildcard tag",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeTag,
					Value: "tag1*",
				},
			},
			want: []*model.Resource{
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/a1"},
						Vtags:      []string{"tag11"},
					},
				},
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/b2"},
						Vtags:      []string{"tag11", "tag13"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "wildcard repository and wildcard tag",
			filters: []*model.Filter{
				{
					Type:  model.FilterTypeName,
					Value: "test/b*",
				},
				{
					Type:  model.FilterTypeTag,
					Value: "tag1*",
				},
			},
			want: []*model.Resource{
				{
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{Name: "test/b2"},
						Vtags:      []string{"tag11", "tag13"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resources, err = adapter.FetchImages(tt.filters)
			if tt.wantErr {
				require.Len(t, resources, 0)
				require.NotNil(t, err)
			} else {
				require.Equal(t, len(tt.want), len(resources))
				for i, resource := range resources {
					require.NotNil(t, resource.Metadata)
					assert.Equal(t, tt.want[i].Metadata.Repository, resource.Metadata.Repository)
					assert.Equal(t, tt.want[i].Metadata.Vtags, resource.Metadata.Vtags)
				}
			}
		})
	}
}

func TestIsDigest(t *testing.T) {
	cases := []struct {
		str      string
		isDigest bool
	}{
		{
			str:      "",
			isDigest: false,
		},
		{
			str:      "latest",
			isDigest: false,
		},
		{
			str:      "sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			isDigest: true,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.isDigest, isDigest(c.str))
	}
}
