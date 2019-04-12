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

package harbor

import (
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchImages(t *testing.T) {
	server := test.NewServer([]*test.RequestHandlerMapping{
		{
			Method:  http.MethodGet,
			Pattern: "/api/projects",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				data := `[{
					"name": "library",
					"metadata": {"public":true}
	
				}]`
				w.Write([]byte(data))
			},
		},
		{
			Method:  http.MethodGet,
			Pattern: "/api/repositories/library/hello-world/tags",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				data := `[{
				"name": "1.0"
			},{
				"name": "2.0"
			}]`
				w.Write([]byte(data))
			},
		},
		{
			Method:  http.MethodGet,
			Pattern: "/api/repositories",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				data := `[{
				"name": "library/hello-world"
			}]`
				w.Write([]byte(data))
			},
		},
	}...)
	defer server.Close()
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter := newAdapter(registry)
	resources, err := adapter.FetchImages([]string{"library"}, nil)
	require.Nil(t, err)
	assert.Equal(t, 1, len(resources))
	assert.Equal(t, model.ResourceTypeRepository, resources[0].Type)
	assert.Equal(t, "hello-world", resources[0].Metadata.Repository.Name)
	assert.Equal(t, "library", resources[0].Metadata.Namespace.Name)
	assert.Equal(t, 2, len(resources[0].Metadata.Vtags))
	assert.Equal(t, "1.0", resources[0].Metadata.Vtags[0])
	assert.Equal(t, "2.0", resources[0].Metadata.Vtags[1])
}

func TestDeleteManifest(t *testing.T) {
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodDelete,
		Pattern: "/api/repositories/library/hello-world/tags/1.0",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}})
	defer server.Close()
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter := newAdapter(registry)
	err := adapter.DeleteManifest("library/hello-world", "1.0")
	require.Nil(t, err)
}
