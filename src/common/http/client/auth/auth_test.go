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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizeOfCookieAuthorizer(t *testing.T) {
	name, value := "name", "value"
	authorizer := NewCookieAuthorizer(name, value)

	// nil request
	require.NotNil(t, authorizer.Authorize(nil))

	// valid request
	req, err := http.NewRequest("", "", nil)
	require.Nil(t, err)
	require.Nil(t, authorizer.Authorize(req))
	require.Equal(t, 1, len(req.Cookies()))
	v, err := req.Cookie(name)
	require.Nil(t, err)
	assert.Equal(t, value, v.Value)
}

func TestAuthorizeOfSecretAuthorizer(t *testing.T) {
	secret := "secret"
	authorizer := NewSecretAuthorizer(secret)

	// nil request
	require.NotNil(t, authorizer.Authorize(nil))

	// valid request
	req, err := http.NewRequest("", "", nil)
	require.Nil(t, err)
	require.Nil(t, authorizer.Authorize(req))
	require.Equal(t, 1, len(req.Cookies()))
	v, err := req.Cookie(secretCookieName)
	require.Nil(t, err)
	assert.Equal(t, secret, v.Value)
}
