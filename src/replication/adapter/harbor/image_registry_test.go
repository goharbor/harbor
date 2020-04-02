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

func TestFetchArtifacts(t *testing.T) {
	server := test.NewServer([]*test.RequestHandlerMapping{
		{
			Method:  http.MethodGet,
			Pattern: "/api/v2.0/projects/library/repositories/hello-world/_self/artifacts",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				data := `[
  {
    "digest": "digest1",
    "tags": [
      {
        "name": "1.0"
      }
    ]
  },
  {
    "digest": "digest2",
    "tags": [
      {
        "name": "2.0"
      }
    ]
  }
]`
				w.Write([]byte(data))
			},
		},
		{
			Method:  http.MethodGet,
			Pattern: "/api/v2.0/projects/library/repositories",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				data := `[{
					"name": "library/hello-world"
				}]`
				w.Write([]byte(data))
			},
		},
		{
			Method:  http.MethodGet,
			Pattern: "/api/v2.0/projects",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				data := `[{
					"name": "library",
					"metadata": {"public":true}
				}]`
				w.Write([]byte(data))
			},
		},
	}...)
	defer server.Close()
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter, err := newAdapter(registry)
	require.Nil(t, err)
	// nil filter
	resources, err := adapter.FetchArtifacts(nil)
	require.Nil(t, err)
	assert.Equal(t, 1, len(resources))
	assert.Equal(t, model.ResourceTypeArtifact, resources[0].Type)
	assert.Equal(t, "library/hello-world", resources[0].Metadata.Repository.Name)
	assert.Equal(t, 2, len(resources[0].Metadata.Artifacts))
	assert.Equal(t, "1.0", resources[0].Metadata.Artifacts[0].Tags[0])
	assert.Equal(t, "2.0", resources[0].Metadata.Artifacts[1].Tags[0])
	// not nil filter
	filters := []*model.Filter{
		{
			Type:  model.FilterTypeName,
			Value: "library/*",
		},
		{
			Type:  model.FilterTypeTag,
			Value: "1.0",
		},
	}
	resources, err = adapter.FetchArtifacts(filters)
	require.Nil(t, err)
	assert.Equal(t, 1, len(resources))
	assert.Equal(t, model.ResourceTypeArtifact, resources[0].Type)
	assert.Equal(t, "library/hello-world", resources[0].Metadata.Repository.Name)
	assert.Equal(t, 1, len(resources[0].Metadata.Artifacts))
	assert.Equal(t, "1.0", resources[0].Metadata.Artifacts[0].Tags[0])
}
