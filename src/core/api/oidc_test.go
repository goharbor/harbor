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

package api

import (
	"github.com/goharbor/harbor/src/common/utils/oidc"
	"net/http"
	"testing"
)

func TestOIDCAPI_Ping(t *testing.T) {
	url := "/api/system/oidc/ping"
	cases := []*codeCheckingCase{
		{ // 401
			request: &testingRequest{
				method:   http.MethodPost,
				bodyJSON: oidc.Conn{},
				url:      url,
			},
			code: http.StatusUnauthorized,
		},
		{ // 403
			request: &testingRequest{
				method:     http.MethodPost,
				bodyJSON:   oidc.Conn{},
				url:        url,
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		{ // 400
			request: &testingRequest{
				method: http.MethodPost,
				bodyJSON: oidc.Conn{
					URL:        "https://www.baidu.com",
					VerifyCert: true,
				},
				url:        url,
				credential: sysAdmin,
			},
			code: http.StatusBadRequest,
		},
		{ // 200
			request: &testingRequest{
				method: http.MethodPost,
				bodyJSON: oidc.Conn{
					URL:        "https://accounts.google.com",
					VerifyCert: true,
				},
				url:        url,
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}
