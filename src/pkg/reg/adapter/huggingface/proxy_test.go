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
	"encoding/json"
	"testing"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
)

// TestProxyPath follows the proxy cache call sequence (HEAD tag, GET manifest by digest, GET
// blobs) with a new adapter and an empty cache for every request, so each by-digest lookup
// takes the miss path, as after a core restart or on another core replica.
func TestProxyPath(t *testing.T) {
	model := tinyModel()
	h := hubtest.New(model)
	defer h.Close()
	fresh := func() *adapter { return newTestAdapter(t, hubRegistry(h, 3, ""), newMemoryCache()) }

	// HEAD by tag: ManifestExist twice (UseLocalManifest, then proxyManifestHead).
	exist, head, err := fresh().ManifestExist(repoTiny, "main")
	require.NoError(t, err)
	require.True(t, exist)
	exist, head2, err := fresh().ManifestExist(repoTiny, "main")
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, head, head2)
	assert.Equal(t, v1.MediaTypeImageManifest, head.MediaType)
	assert.Positive(t, head.Size)

	// GET by digest: ManifestExist then PullManifest on one instance.
	a := fresh()
	exist, byDigest, err := a.ManifestExist(repoTiny, head.Digest.String())
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, head.Digest, byDigest.Digest)
	mf, dgst, err := a.PullManifest(repoTiny, head.Digest.String())
	require.NoError(t, err)
	assert.Equal(t, head.Digest.String(), dgst)
	_, payload, err := mf.Payload()
	require.NoError(t, err)
	assert.Equal(t, head.Size, int64(len(payload)))
	var m v1.Manifest
	require.NoError(t, json.Unmarshal(payload, &m))

	// GET blob: each on a new adapter; the proxy calls PullBlob twice per cold blob.
	files := model.Commits[commitMain].Files
	for _, d := range append([]v1.Descriptor{m.Config}, m.Layers...) {
		for range 2 {
			size, r, err := fresh().PullBlob(repoTiny, d.Digest.String())
			content := readAll(t, d.Size, r, err)
			assert.Equal(t, d.Size, size)
			assert.Equal(t, d.Digest, digest.FromBytes(content))
			if path, ok := d.Annotations["org.cncf.model.filepath"]; ok {
				assert.Equal(t, files[path].Content, content)
			}
		}
	}

	// The same sequence from the factory, which uses the shared process cache.
	h.ResetCalls()
	clearProcessCache(t)
	viaFactory := func() *adapter {
		created, err := (&factory{}).Create(hubRegistry(h, 3, ""))
		require.NoError(t, err)
		return created.(*adapter)
	}
	_, warm, err := viaFactory().ManifestExist(repoTiny, "main")
	require.NoError(t, err)
	_, _, err = viaFactory().PullManifest(repoTiny, warm.Digest.String())
	require.NoError(t, err)
	_, r, err := viaFactory().PullBlob(repoTiny, m.Config.Digest.String())
	readAll(t, m.Config.Size, r, err)
	assert.Equal(t, head.Digest, warm.Digest)
	assert.Equal(t, 2, h.Calls("refs"), "one refs call, doubled by the case redirect")
	assert.Equal(t, 2, h.Calls("revision"), "one snapshot call, doubled by the case redirect")
}
