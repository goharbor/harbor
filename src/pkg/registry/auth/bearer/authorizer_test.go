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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

func TestModify(t *testing.T) {
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiIsImtpZCI6IlBZWU86VEVXVTpWN0pIOjI2SlY6QVFUWjpMSkMzOlNYVko6WEdIQTozNEYyOjJMQVE6WlJNSzpaN1E2In0.eyJpc3MiOiJhdXRoLmRvY2tlci5jb20iLCJzdWIiOiJqbGhhd24iLCJhdWQiOiJyZWdpc3RyeS5kb2NrZXIuY29tIiwiZXhwIjoxNDE1Mzg3MzE1LCJuYmYiOjE0MTUzODcwMTUsImlhdCI6MTQxNTM4NzAxNSwianRpIjoidFlKQ08xYzZjbnl5N2tBbjBjN3JLUGdiVjFIMWJGd3MiLCJhY2Nlc3MiOlt7InR5cGUiOiJyZXBvc2l0b3J5IiwibmFtZSI6InNhbWFsYmEvbXktYXBwIiwiYWN0aW9ucyI6WyJwdXNoIl19XX0.QhflHPfbd6eVF4lM9bwYpFZIV0PfikbyXuLx959ykRTBpe3CYnzs6YBK8FToVb5R47920PVLrh8zuLzdCr9t3w"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || u != "username" || p != "password" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Write([]byte(fmt.Sprintf(`{"token": "%s", "expires_in": 3600,"issued_at": "2009-11-10T23:00:00Z"}`, token)))
	}))
	defer server.Close()

	// invalid credential
	a := basic.NewAuthorizer("username", "invalid_password")
	authorizer := NewAuthorizer(server.URL, "service", a, commonhttp.NewTransport())
	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	err := authorizer.Modify(req)
	require.NotNil(t, err)
	assert.True(t, errors.IsErr(err, errors.UnAuthorizedCode))

	// valid credential
	a = basic.NewAuthorizer("username", "password")
	authorizer = NewAuthorizer(server.URL, "service", a, commonhttp.NewTransport())
	req, _ = http.NewRequest(http.MethodGet, server.URL, nil)
	err = authorizer.Modify(req)
	require.Nil(t, err)
	assert.Equal(t, fmt.Sprintf("Bearer %s", token), req.Header.Get("Authorization"))
}
