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
)

// cannot verify the real scenario here
func TestSwitchQuota(t *testing.T) {
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    "/api/internal/switchquota",
			},
			code: http.StatusUnauthorized,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/internal/switchquota",
				credential: sysAdmin,
				bodyJSON: &QuotaSwitcher{
					Enabled: true,
				},
			},
			code: http.StatusOK,
		},
		// 403
		{
			request: &testingRequest{
				url:        "/api/internal/switchquota",
				method:     http.MethodPut,
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
	}
	runCodeCheckingCases(t, cases...)
}

// cannot verify the real scenario here
func TestSyncQuota(t *testing.T) {
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/internal/syncquota",
			},
			code: http.StatusUnauthorized,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/internal/syncquota",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
		// 403
		{
			request: &testingRequest{
				url:        "/api/internal/syncquota",
				method:     http.MethodPost,
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
	}
	runCodeCheckingCases(t, cases...)
}
