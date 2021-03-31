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

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuthorizer(t *testing.T) {
	type suite struct {
		key   string
		value string
		in    string
	}

	var (
		s          suite
		authorizer modifier.Modifier
		request    *http.Request
		err        error
	)

	// set in header
	s = suite{key: "Authorization", value: "Basic abc", in: "header"}
	authorizer = NewAPIKeyAuthorizer(s.key, s.value, s.in)
	request, err = http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.Nil(t, err)
	err = authorizer.Modify(request)
	assert.Nil(t, err)
	assert.Equal(t, s.value, request.Header.Get(s.key))

	// set in query
	s = suite{key: "private_token", value: "abc", in: "query"}
	authorizer = NewAPIKeyAuthorizer(s.key, s.value, s.in)
	request, err = http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.Nil(t, err)
	err = authorizer.Modify(request)
	assert.Nil(t, err)
	assert.Equal(t, s.value, request.URL.Query().Get(s.key))

	// set in invalid location
	s = suite{key: "", value: "", in: "invalid"}
	authorizer = NewAPIKeyAuthorizer(s.key, s.value, s.in)
	request, err = http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.Nil(t, err)
	err = authorizer.Modify(request)
	assert.NotNil(t, err)
}
