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

package registry

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func direct(req *http.Request) {
	req.Header.Add("test-key", "test-value")
}

func TestBasicAuthDirector(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "127.0.0.1", nil)
	os.Setenv("REGISTRY_CREDENTIAL_USERNAME", "testuser")
	os.Setenv("REGISTRY_CREDENTIAL_PASSWORD", "testpassword")
	defer func() {
		os.Unsetenv("REGISTRY_CREDENTIAL_USERNAME")
		os.Unsetenv("REGISTRY_CREDENTIAL_PASSWORD")
	}()

	d := basicAuthDirector(direct)
	d(req)
	assert.Equal(t, "test-value", req.Header.Get("test-key"))
	user, pass, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "testuser", user)
	assert.Equal(t, "testpassword", pass)
}
