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
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestFactory(t *testing.T) {
	f, err := adp.GetFactory(model.RegistryTypeHuggingFace)
	require.NoError(t, err)
	a, err := f.Create(&model.Registry{})
	require.NoError(t, err)
	_, ok := a.(adp.ArtifactRegistry)
	assert.True(t, ok)
	assert.Equal(t, hub.DefaultEndpoint, a.(*adapter).hub.Endpoint())

	p := f.AdapterPattern()
	assert.Equal(t, model.EndpointPatternTypeList, p.EndpointPattern.EndpointType)
	assert.Equal(t, hub.DefaultEndpoint, p.EndpointPattern.Endpoints[0].Value)
	assert.Nil(t, p.CredentialPattern)

	_, err = f.Create(&model.Registry{URL: "not a url"})
	assert.Error(t, err)
}

func TestInfo(t *testing.T) {
	a := newTestAdapter(t, &model.Registry{URL: "https://huggingface.co/"}, newMemoryCache())
	assert.Equal(t, "https://huggingface.co", a.hub.Endpoint())
	info, err := a.Info()
	require.NoError(t, err)
	assert.Equal(t, model.RegistryTypeHuggingFace, info.Type)
	assert.Equal(t, []string{model.ResourceTypeArtifact}, info.SupportedResourceTypes)
	require.Len(t, info.SupportedResourceFilters, 2)
	assert.Equal(t, model.FilterTypeName, info.SupportedResourceFilters[0].Type)
	assert.Equal(t, model.FilterTypeTag, info.SupportedResourceFilters[1].Type)
	assert.Equal(t, []string{model.TriggerTypeManual, model.TriggerTypeScheduled}, info.SupportedTriggers)
	assert.True(t, info.SupportedCopyByChunk)
}

func TestHealthCheck(t *testing.T) {
	h := hubtest.New(tinyModel())
	h.Token = "hf_good"
	defer h.Close()

	cases := []struct {
		name  string
		token string
		want  string
	}{
		{"anonymous pings the model api", "", model.Healthy},
		{"valid token", "hf_good", model.Healthy},
		{"invalid token", "hf_bad", model.Unhealthy},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := newTestAdapter(t, hubRegistry(h, 1, c.token), newMemoryCache())
			status, err := a.HealthCheck()
			require.NoError(t, err)
			assert.Equal(t, c.want, status)
		})
	}

	h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())
	status, err := a.HealthCheck()
	require.NoError(t, err)
	assert.Equal(t, model.Unhealthy, status)
}

func TestSourceOnly(t *testing.T) {
	a := newTestAdapter(t, &model.Registry{}, newMemoryCache())
	isMethodNotAllowed := func(err error) {
		t.Helper()
		assert.True(t, errors.IsErr(err, errors.MethodNotAllowedCode), err)
	}
	isMethodNotAllowed(a.PrepareForPush(nil))
	_, err := a.PushManifest("a/b", "t", "m", nil)
	isMethodNotAllowed(err)
	isMethodNotAllowed(a.DeleteManifest("a/b", "t"))
	isMethodNotAllowed(a.PushBlob("a/b", "d", 0, bytes.NewReader(nil)))
	_, _, err = a.PushBlobChunk("a/b", "d", 0, bytes.NewReader(nil), 0, 0, "")
	isMethodNotAllowed(err)
	isMethodNotAllowed(a.MountBlob("a/b", "d", "c/d"))
	isMethodNotAllowed(a.DeleteTag("a/b", "t"))
	mount, _, err := a.CanBeMount("d")
	assert.NoError(t, err)
	assert.False(t, mount)
}
