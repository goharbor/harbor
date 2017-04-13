// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package auth

import (
	"net/http"
	"testing"

	"github.com/vmware/harbor/src/common/utils/test"
)

func TestAuthorizeOfStandardTokenAuthorizer(t *testing.T) {
	handler := test.Handler(&test.Response{
		Body: []byte(`
		{
			"token":"token",
			"expires_in":300,
			"issued_at":"2016-08-17T23:17:58+08:00"
		}
		`),
	})

	server := test.NewServer(&test.RequestHandlerMapping{
		Method:  "GET",
		Pattern: "/token",
		Handler: handler,
	})
	defer server.Close()

	authorizer := NewStandardTokenAuthorizer(nil, false, "", "repository", "library/ubuntu", "pull")
	req, err := http.NewRequest("GET", "http://registry", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	params := map[string]string{
		"realm": server.URL + "/token",
	}

	if err := authorizer.Authorize(req, params); err != nil {
		t.Fatalf("failed to authorize request: %v", err)
	}

	tk := req.Header.Get("Authorization")
	if tk != "Bearer token" {
		t.Errorf("unexpected token: %s != %s", tk, "Bearer token")
	}
}

func TestSchemeOfStandardTokenAuthorizer(t *testing.T) {
	authorizer := &standardTokenAuthorizer{}
	if authorizer.Scheme() != "bearer" {
		t.Errorf("unexpected scheme: %s != %s", authorizer.Scheme(), "bearer")
	}

}
