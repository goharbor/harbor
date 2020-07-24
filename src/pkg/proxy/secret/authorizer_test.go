//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package secret

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestAuthorizer(t *testing.T) {
	authorizer := &authorizer{}

	// not manifest/blob requests
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1/v2/_catalog", nil)
	err := authorizer.Modify(req)
	require.Nil(t, err)
	assert.Empty(t, GetSecret(req))

	// pass, manifest URL
	req, _ = http.NewRequest(http.MethodGet, "http://127.0.0.1/v2/library/hello-world/manifests/latest", nil)
	err = authorizer.Modify(req)
	require.Nil(t, err)
	assert.NotEmpty(t, GetSecret(req))

	// pass, blob URL
	req, _ = http.NewRequest(http.MethodGet, "http://127.0.0.1/v2/library/hello-world/blobs/sha256:e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9", nil)
	err = authorizer.Modify(req)
	require.Nil(t, err)
	assert.NotEmpty(t, GetSecret(req))

	// pass, blob upload URL
	req, _ = http.NewRequest(http.MethodGet, "http://127.0.0.1/v2/library/hello-world/blobs/uploads/uuid", nil)
	err = authorizer.Modify(req)
	require.Nil(t, err)
	assert.NotEmpty(t, GetSecret(req))
}
