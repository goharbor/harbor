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

package auth

import (
	"net/http"
	"testing"
)

func TestAddAuthorizationOfBasicAuthCredential(t *testing.T) {
	cred := NewBasicAuthCredential("usr", "pwd")
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	cred.Modify(req)

	usr, pwd, ok := req.BasicAuth()
	if !ok {
		t.Fatal("basic auth not found")
	}

	if usr != "usr" {
		t.Errorf("unexpected username: %s != usr", usr)
	}

	if pwd != "pwd" {
		t.Errorf("unexpected password: %s != pwd", pwd)
	}
}
