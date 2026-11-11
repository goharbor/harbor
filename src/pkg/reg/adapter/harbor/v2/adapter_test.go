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

package v2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/cosign"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/harbor/base"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func TestListArtifactsNestedAccessoriesInheritSubjectType(t *testing.T) {
	const artifactsPath = "/api/v2.0/projects/library/repositories/image/artifacts"
	// Accessory artifact responses report UNKNOWN even though their subject is an IMAGE.
	responses := map[string]string{
		artifactsPath: `[{
			"type": "IMAGE", "digest": "sha256:image", "tags": [{"name": "1.0"}],
			"accessories": [{"type": "signature.cosign", "digest": "sha256:acc1"}]
		}]`,
		artifactsPath + "/sha256:acc1": `{
			"type": "UNKNOWN", "digest": "sha256:acc1",
			"accessories": [{"type": "signature.cosign", "digest": "sha256:subacc1"}]
		}`,
		artifactsPath + "/sha256:subacc1": `{
			"type": "UNKNOWN", "digest": "sha256:subacc1", "accessories": []
		}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "true", r.URL.Query().Get("with_accessory"))
		response, ok := responses[r.URL.Path]
		if !ok {
			t.Errorf("unexpected request: %s", r.URL)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(response))
		assert.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	a := New(&base.Adapter{
		Client: &base.Client{
			URL:        server.URL,
			APIVersion: "v2.0",
			C:          commonhttp.NewClient(server.Client()),
		},
	}).(*adapter)
	expected := []*model.Artifact{
		{Type: "IMAGE", Digest: "sha256:image", Tags: []string{"1.0"}},
		{Type: "IMAGE", Digest: "sha256:acc1", IsAcc: true, ParentTags: []string{"1.0"}},
		{Type: "IMAGE", Digest: "sha256:subacc1", IsAcc: true, ParentTags: []string{"1.0"}},
	}
	for _, tc := range []struct {
		name    string
		filters []*model.Filter
	}{
		{name: "collection inherits subject type"},
		{
			name: "image filter keeps both accessory levels",
			filters: []*model.Filter{
				{Type: model.FilterTypeResource, Value: model.ResourceTypeImage},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			artifacts, err := a.listArtifacts("library/image", tc.filters)
			require.NoError(t, err)
			require.Equal(t, expected, artifacts)
		})
	}
}
