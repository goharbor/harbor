// Copyright 2018 Project Harbor Authors
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

package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

func TestReplicationAdapterAPIList(t *testing.T) {
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/replication/adapters",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/adapters",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/adapters",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func fakedFactory(*model.Registry) (adapter.Adapter, error) {
	return nil, nil
}
func TestReplicationAdapterAPIGet(t *testing.T) {
	err := adapter.RegisterFactory(
		&adapter.Info{
			Type:                   "harbor",
			SupportedResourceTypes: []model.ResourceType{"image"},
		}, fakedFactory)
	require.Nil(t, err)

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/replication/adapters/harbor",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/adapters/harbor",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/adapters/gcs",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/adapters/harbor",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}
