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

package huggingface

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestFactory(t *testing.T) {
	f, err := adp.GetFactory(model.RegistryTypeHuggingFace)
	require.NoError(t, err)
	a, err := f.Create(&model.Registry{URL: DefaultEndpoint})
	require.NoError(t, err)
	assert.NotNil(t, a)

	pattern := f.AdapterPattern()
	assert.Equal(t, model.EndpointPatternTypeList, pattern.EndpointPattern.EndpointType)
	assert.Equal(t, DefaultEndpoint, pattern.EndpointPattern.Endpoints[0].Value)
	assert.Equal(t, model.AccessKeyTypeFix, pattern.CredentialPattern.AccessKeyType)
}

func TestInfo(t *testing.T) {
	info, err := New(&model.Registry{}).Info()
	require.NoError(t, err)
	assert.Equal(t, model.RegistryTypeHuggingFace, info.Type)
	// must not be a replication candidate
	assert.Empty(t, info.SupportedResourceTypes)
}

func TestEndpointAndToken(t *testing.T) {
	assert.Equal(t, DefaultEndpoint, Endpoint(&model.Registry{}))
	assert.Equal(t, "https://mirror.example.com", Endpoint(&model.Registry{URL: "https://mirror.example.com/"}))
	assert.Equal(t, "", Token(&model.Registry{}))
	assert.Equal(t, "hf_x", Token(&model.Registry{Credential: &model.Credential{AccessSecret: " hf_x "}}))
}

func TestHealthCheck(t *testing.T) {
	var gotPath, gotAuth string
	status := http.StatusOK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(status)
	}))
	defer server.Close()

	// anonymous
	a := New(&model.Registry{URL: server.URL})
	s, err := a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Healthy, s)
	assert.Equal(t, "/api/models", gotPath)
	assert.Equal(t, "", gotAuth)

	// with token
	a = New(&model.Registry{URL: server.URL, Credential: &model.Credential{AccessSecret: "hf_token"}})
	s, err = a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Healthy, s)
	assert.Equal(t, "/api/whoami-v2", gotPath)
	assert.Equal(t, "Bearer hf_token", gotAuth)

	// invalid token
	status = http.StatusUnauthorized
	s, err = a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Unhealthy, s)

	// unreachable
	server.Close()
	s, err = a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Unhealthy, s)
}

func TestPrepareForPush(t *testing.T) {
	assert.Error(t, New(&model.Registry{}).PrepareForPush(nil))
}
