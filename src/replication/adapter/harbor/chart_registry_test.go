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
	"bytes"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchCharts(t *testing.T) {
	server := test.NewServer([]*test.RequestHandlerMapping{
		{
			Method:  http.MethodGet,
			Pattern: "/api/chartrepo/library/charts/harbor",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				data := `[{
				"name": "harbor",
				"version":"1.0"
			},{
				"name": "harbor",
				"version":"2.0"
			}]`
				w.Write([]byte(data))
			},
		},
		{
			Method:  http.MethodGet,
			Pattern: "/api/chartrepo/library/charts",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				data := `[{
				"name": "harbor"
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
	resources, err := adapter.FetchCharts([]string{"library"}, nil)
	require.Nil(t, err)
	assert.Equal(t, 2, len(resources))
	assert.Equal(t, model.ResourceTypeChart, resources[0].Type)
	assert.Equal(t, "harbor", resources[0].Metadata.Repository.Name)
	assert.Equal(t, "library", resources[0].Metadata.Namespace.Name)
	assert.Equal(t, 1, len(resources[0].Metadata.Vtags))
	assert.Equal(t, "1.0", resources[0].Metadata.Vtags[0])
}

func TestChartExist(t *testing.T) {
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodGet,
		Pattern: "/api/chartrepo/library/charts/harbor/1.0",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			data := `{
				"metadata": {
					"urls":["http://127.0.0.1/charts"]
				}
			}`
			w.Write([]byte(data))
		},
	})
	defer server.Close()
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter, err := newAdapter(registry)
	require.Nil(t, err)
	exist, err := adapter.ChartExist("library/harbor", "1.0")
	require.Nil(t, err)
	require.True(t, exist)
}

func TestDownloadChart(t *testing.T) {
	server := test.NewServer([]*test.RequestHandlerMapping{
		{
			Method:  http.MethodGet,
			Pattern: "/api/chartrepo/library/charts/harbor/1.0",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				data := `{
				"metadata": {
					"urls":["charts/harbor-1.0.tgz"]
				}
			}`
				w.Write([]byte(data))
			},
		},
		{
			Method:  http.MethodGet,
			Pattern: "/api/chartrepo/library/charts/harbor-1.0.tgz",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
	}...)
	defer server.Close()
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter, err := newAdapter(registry)
	require.Nil(t, err)
	_, err = adapter.DownloadChart("library/harbor", "1.0")
	require.Nil(t, err)
}

func TestUploadChart(t *testing.T) {
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodPost,
		Pattern: "/api/chartrepo/library/charts",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	})
	defer server.Close()
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter, err := newAdapter(registry)
	require.Nil(t, err)
	err = adapter.UploadChart("library/harbor", "1.0", bytes.NewBuffer(nil))
	require.Nil(t, err)
}

func TestDeleteChart(t *testing.T) {
	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  http.MethodDelete,
		Pattern: "/api/chartrepo/library/charts/harbor/1.0",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	})
	defer server.Close()
	registry := &model.Registry{
		URL: server.URL,
	}
	adapter, err := newAdapter(registry)
	require.Nil(t, err)
	err = adapter.DeleteChart("library/harbor", "1.0")
	require.Nil(t, err)
}
