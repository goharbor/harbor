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

package security

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestIDToken(t *testing.T) {
	idToken := &idToken{}

	// not the OIDC mode
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	ctx := idToken.Generate(req)
	assert.Nil(t, ctx)

	// not the candidate request
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/chartrepo/", nil)
	require.Nil(t, err)
	req = req.WithContext(internal.WithAuthMode(req.Context(), common.DBAuth))
	ctx = idToken.Generate(req)
	assert.Nil(t, ctx)

	// contains no authorization header
	req, err = http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req = req.WithContext(internal.WithAuthMode(req.Context(), common.OIDCAuth))
	ctx = idToken.Generate(req)
	assert.Nil(t, ctx)
}
