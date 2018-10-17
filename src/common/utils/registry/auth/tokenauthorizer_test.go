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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterReq(t *testing.T) {
	authorizer := tokenAuthorizer{}

	// v2
	req, err := http.NewRequest(http.MethodGet, "http://registry/v2/", nil)
	require.Nil(t, err)
	goon, err := authorizer.filterReq(req)
	assert.Nil(t, err)
	assert.True(t, goon)

	// catalog
	req, err = http.NewRequest(http.MethodGet, "http://registry/v2/_catalog?n=1000", nil)
	require.Nil(t, err)
	goon, err = authorizer.filterReq(req)
	assert.Nil(t, err)
	assert.True(t, goon)

	// contains two v2 in path
	req, err = http.NewRequest(http.MethodGet, "http://registry/v2/library/v2/tags/list", nil)
	require.Nil(t, err)
	goon, err = authorizer.filterReq(req)
	assert.Nil(t, err)
	assert.True(t, goon)

	// different scheme
	req, err = http.NewRequest(http.MethodGet, "https://registry/v2/library/golang/tags/list", nil)
	require.Nil(t, err)
	goon, err = authorizer.filterReq(req)
	assert.Nil(t, err)
	assert.False(t, goon)

	// different host
	req, err = http.NewRequest(http.MethodGet, "http://vmware.com/v2/library/golang/tags/list", nil)
	require.Nil(t, err)
	goon, err = authorizer.filterReq(req)
	assert.Nil(t, err)
	assert.False(t, goon)

	// different path
	req, err = http.NewRequest(http.MethodGet, "https://registry/s3/ssss", nil)
	require.Nil(t, err)
	goon, err = authorizer.filterReq(req)
	assert.Nil(t, err)
	assert.False(t, goon)
}

func TestParseScopes(t *testing.T) {
	// contains from in query string
	req, err := http.NewRequest(http.MethodGet, "http://registry/v2?from=library", nil)
	require.Nil(t, err)
	scopses, err := parseScopes(req)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(scopses))
	assert.EqualValues(t, &token.ResourceActions{
		Type: "repository",
		Name: "library",
		Actions: []string{
			"pull"},
	}, scopses[0])

	// v2
	req, err = http.NewRequest(http.MethodGet, "http://registry/v2", nil)
	require.Nil(t, err)
	scopses, err = parseScopes(req)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(scopses))

	// catalog
	req, err = http.NewRequest(http.MethodGet, "http://registry/v2/_catalog", nil)
	require.Nil(t, err)
	scopses, err = parseScopes(req)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(scopses))
	assert.EqualValues(t, &token.ResourceActions{
		Type: "registry",
		Name: "catalog",
		Actions: []string{
			"*"},
	}, scopses[0])

	// manifest
	req, err = http.NewRequest(http.MethodPut, "http://registry/v2/library/mysql/5.6/manifests/1", nil)
	require.Nil(t, err)
	scopses, err = parseScopes(req)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(scopses))
	assert.EqualValues(t, &token.ResourceActions{
		Type:    "repository",
		Name:    "library/mysql/5.6",
		Actions: []string{"pull", "push"},
	}, scopses[0])

	// invalid
	req, err = http.NewRequest(http.MethodPut, "http://registry/other", nil)
	require.Nil(t, err)
	scopses, err = parseScopes(req)
	assert.NotNil(t, err)
}

func TestGetAndUpdateCachedToken(t *testing.T) {
	authorizer := &tokenAuthorizer{
		cachedTokens: make(map[string]*models.Token),
	}

	// empty cache
	token := authorizer.getCachedToken("")
	assert.Nil(t, token)

	// put a valid token into cache
	token = &models.Token{
		Token:     "token",
		ExpiresIn: 60,
		IssuedAt:  time.Now().Format(time.RFC3339),
	}
	authorizer.updateCachedToken("", token)
	token2 := authorizer.getCachedToken("")
	assert.EqualValues(t, token, token2)

	// put a expired token into cache
	token = &models.Token{
		Token:     "token",
		ExpiresIn: 60,
		IssuedAt:  time.Now().Add(-time.Second * 120).Format("2006-01-02 15:04:05.999999999 -0700 MST"),
	}
	authorizer.updateCachedToken("", token)
	token2 = authorizer.getCachedToken("")
	assert.Nil(t, token2)
}

func TestModifyOfStandardTokenAuthorizer(t *testing.T) {
	token := &models.Token{
		Token:     "token",
		ExpiresIn: 3600,
		IssuedAt:  time.Now().String(),
	}
	data, err := json.Marshal(token)
	require.Nil(t, err)

	tokenHandler := test.Handler(&test.Response{
		Body: data,
	})

	tokenServer := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/service/token",
			Handler: tokenHandler,
		})
	defer tokenServer.Close()

	header := fmt.Sprintf("Bearer realm=\"%s/service/token\",service=\"registry\"",
		tokenServer.URL)
	pingHandler := test.Handler(&test.Response{
		StatusCode: http.StatusUnauthorized,
		Headers: map[string]string{
			"WWW-Authenticate": header,
		},
	})
	registryServer := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  "GET",
			Pattern: "/v2",
			Handler: pingHandler,
		})
	defer registryServer.Close()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/v2/", registryServer.URL), nil)
	require.Nil(t, err)

	authorizer := NewStandardTokenAuthorizer(http.DefaultClient, nil)

	err = authorizer.Modify(req)
	require.Nil(t, err)

	tk := req.Header.Get("Authorization")
	assert.Equal(t, strings.ToLower("Bearer "+token.Token), strings.ToLower(tk))
}

func TestUserAgentModifier(t *testing.T) {
	agent := "harbor-registry-client"
	modifier := &UserAgentModifier{
		UserAgent: agent,
	}
	req, err := http.NewRequest(http.MethodGet, "http://registry/v2/", nil)
	require.Nil(t, err)
	modifier.Modify(req)
	actual := req.Header.Get("User-Agent")
	if actual != agent {
		t.Errorf("expect request to have header User-Agent=%s, but got User-Agent=%s", agent, actual)
	}
}
