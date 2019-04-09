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

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

func TestInfo(t *testing.T) {
	// chart museum enabled
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/systeminfo",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			data := `{"with_chartmuseum":true}`
			w.Write([]byte(data))
		},
	})
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter := newAdapter(registry)
	info, err := adapter.Info()
	require.Nil(t, err)
	assert.Equal(t, model.RegistryTypeHarbor, info.Type)
	assert.Equal(t, 3, len(info.SupportedResourceFilters))
	assert.Equal(t, 3, len(info.SupportedTriggers))
	assert.Equal(t, 2, len(info.SupportedResourceTypes))
	assert.Equal(t, model.ResourceTypeRepository, info.SupportedResourceTypes[0])
	assert.Equal(t, model.ResourceTypeChart, info.SupportedResourceTypes[1])
	server.Close()

	// chart museum disabled
	server = test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/systeminfo",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			data := `{"with_chartmuseum":false}`
			w.Write([]byte(data))
		},
	})
	registry = &model.Registry{
		URL: server.URL,
	}
	adapter = newAdapter(registry)
	info, err = adapter.Info()
	require.Nil(t, err)
	assert.Equal(t, model.RegistryTypeHarbor, info.Type)
	assert.Equal(t, 3, len(info.SupportedResourceFilters))
	assert.Equal(t, 3, len(info.SupportedTriggers))
	assert.Equal(t, 1, len(info.SupportedResourceTypes))
	assert.Equal(t, model.ResourceTypeRepository, info.SupportedResourceTypes[0])
	server.Close()
}

func TestListNamespaces(t *testing.T) {
	// TODO
}

func TestCreateNamespace(t *testing.T) {
	// project doesn't exist
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodPost,
		Pattern: "/api/projects",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		},
	})
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter := newAdapter(registry)
	err := adapter.CreateNamespace(&model.Namespace{
		Name: "library",
	})
	require.Nil(t, err)
	server.Close()

	// project already exists
	server = test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodPost,
		Pattern: "/api/projects",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
		},
	})
	registry = &model.Registry{
		URL: server.URL,
	}
	adapter = newAdapter(registry)
	err = adapter.CreateNamespace(&model.Namespace{
		Name: "library",
	})
	require.Nil(t, err)
	server.Close()
}

func TestGetNamespace(t *testing.T) {
	// project exists
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/projects",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			data := `[{
				"name": "library",
				"metadata": {"public":true}

			}]`
			w.Write([]byte(data))
		},
	})
	defer server.Close()
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter := newAdapter(registry)
	namespace, err := adapter.GetNamespace("library")
	require.Nil(t, err)
	assert.Equal(t, "library", namespace.Name)
	assert.True(t, namespace.Metadata["public"].(bool))

	// project doesn't exists
	namespace, err = adapter.GetNamespace("test")
	require.NotNil(t, err)
}
