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

package bearer

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestStringOfScope(t *testing.T) {
	scope := &scope{
		Type:    scopeTypeRepository,
		Name:    "library/hello-world",
		Actions: []string{scopeActionPull, scopeActionPush},
	}
	assert.Equal(t, "repository:library/hello-world:pull,push", scope.String())
}

func TestParseScopes(t *testing.T) {
	// base
	req, _ := http.NewRequest(http.MethodGet, "/v2/", nil)
	scopes := parseScopes(req)
	require.Nil(t, scopes)

	// catalog
	req, _ = http.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	scopes = parseScopes(req)
	require.Len(t, scopes, 1)
	assert.Equal(t, scopeTypeRegistry, scopes[0].Type)
	assert.Equal(t, "catalog", scopes[0].Name)
	require.Len(t, scopes[0].Actions, 1)
	assert.Equal(t, scopeActionAll, scopes[0].Actions[0])

	// list tags
	req, _ = http.NewRequest(http.MethodGet, "/v2/library/hello-world/tags/list", nil)
	scopes = parseScopes(req)
	require.Len(t, scopes, 1)
	assert.Equal(t, scopeTypeRepository, scopes[0].Type)
	assert.Equal(t, "library/hello-world", scopes[0].Name)
	require.Len(t, scopes[0].Actions, 1)
	assert.Equal(t, scopeActionPull, scopes[0].Actions[0])

	// get manifest by tag
	req, _ = http.NewRequest(http.MethodGet, "/v2/library/hello-world/manifests/latest", nil)
	scopes = parseScopes(req)
	require.Len(t, scopes, 1)
	assert.Equal(t, scopeTypeRepository, scopes[0].Type)
	assert.Equal(t, "library/hello-world", scopes[0].Name)
	require.Len(t, scopes[0].Actions, 1)
	assert.Equal(t, scopeActionPull, scopes[0].Actions[0])

	// get manifest by digest
	req, _ = http.NewRequest(http.MethodGet, "/v2/library/hello-world/manifests/sha256:eec76eedea59f7bf39a2713bfd995c82cfaa97724ee5b7f5aba253e07423d0ae", nil)
	scopes = parseScopes(req)
	require.Len(t, scopes, 1)
	assert.Equal(t, scopeTypeRepository, scopes[0].Type)
	assert.Equal(t, "library/hello-world", scopes[0].Name)
	require.Len(t, scopes[0].Actions, 1)
	assert.Equal(t, scopeActionPull, scopes[0].Actions[0])

	// push manifest
	req, _ = http.NewRequest(http.MethodPut, "/v2/library/hello-world/manifests/sha256:eec76eedea59f7bf39a2713bfd995c82cfaa97724ee5b7f5aba253e07423d0ae", nil)
	scopes = parseScopes(req)
	require.Len(t, scopes, 1)
	assert.Equal(t, scopeTypeRepository, scopes[0].Type)
	assert.Equal(t, "library/hello-world", scopes[0].Name)
	require.Len(t, scopes[0].Actions, 2)
	assert.Equal(t, scopeActionPull, scopes[0].Actions[0])
	assert.Equal(t, scopeActionPush, scopes[0].Actions[1])

	// delete manifest
	req, _ = http.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/sha256:eec76eedea59f7bf39a2713bfd995c82cfaa97724ee5b7f5aba253e07423d0ae", nil)
	scopes = parseScopes(req)
	require.Len(t, scopes, 1)
	assert.Equal(t, scopeTypeRepository, scopes[0].Type)
	assert.Equal(t, "library/hello-world", scopes[0].Name)
	require.Len(t, scopes[0].Actions, 1)
	assert.Equal(t, scopeActionAll, scopes[0].Actions[0])

	// mount blob
	req, _ = http.NewRequest(http.MethodPost, "/v2/library/hello-world/blobs/uploads/?mount=sha256:eec76eedea59f7bf39a2713bfd995c82cfaa97724ee5b7f5aba253e07423d0ae&from=library/alpine", nil)
	scopes = parseScopes(req)
	require.Len(t, scopes, 2)
	assert.Equal(t, scopeTypeRepository, scopes[0].Type)
	assert.Equal(t, "library/alpine", scopes[0].Name)
	require.Len(t, scopes[0].Actions, 1)
	assert.Equal(t, scopeActionPull, scopes[1].Actions[0])
	assert.Equal(t, scopeTypeRepository, scopes[1].Type)
	assert.Equal(t, "library/hello-world", scopes[1].Name)
	require.Len(t, scopes[1].Actions, 2)
	assert.Equal(t, scopeActionPull, scopes[1].Actions[0])
	assert.Equal(t, scopeActionPush, scopes[1].Actions[1])

	// no match
	req, _ = http.NewRequest(http.MethodPost, "/api/others", nil)
	scopes = parseScopes(req)
	require.Nil(t, scopes)
}
