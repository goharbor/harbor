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
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	var reg, err = adp.NewDefaultImageRegistry(registry)
	assert.NotNil(t, reg)
	assert.Nil(t, err)
	var adapter = Native{
		DefaultImageRegistry: reg,
		registry:             registry,
	}
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
