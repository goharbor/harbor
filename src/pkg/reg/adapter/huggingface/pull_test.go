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
	"io"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/synth"
)

func readAll(t *testing.T, size int64, r io.ReadCloser, err error) []byte {
	t.Helper()
	require.NoError(t, err)
	defer r.Close()
	b, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, size, int64(len(b)))
	return b
}

func manifestOf(t *testing.T, a *adapter, repository, reference string) (v1.Manifest, digest.Digest) {
	t.Helper()
	mf, dgst, err := a.PullManifest(repository, reference)
	require.NoError(t, err)
	mediaType, payload, err := mf.Payload()
	require.NoError(t, err)
	assert.Equal(t, v1.MediaTypeImageManifest, mediaType)
	assert.Equal(t, digest.FromBytes(payload).String(), dgst)
	var m v1.Manifest
	require.NoError(t, json.Unmarshal(payload, &m))
	return m, digest.Digest(dgst)
}

func TestManifestReferences(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	byTag, dTag := manifestOf(t, a, repoTiny, "main")
	_, dBranch := manifestOf(t, a, repoTiny, "feature-x")
	_, dCommit := manifestOf(t, a, repoTiny, commitMain)
	_, dDigest := manifestOf(t, a, repoTiny, dTag.String())
	_, dV1 := manifestOf(t, a, "ORG/Tiny-Model", "v1")
	assert.Equal(t, dTag, dBranch)
	assert.Equal(t, dTag, dCommit)
	assert.Equal(t, dTag, dDigest)
	assert.NotEqual(t, dTag, dV1)

	assert.Equal(t, synth.ArtifactTypeModelManifest, byTag.ArtifactType)
	var paths []string
	for _, l := range byTag.Layers {
		paths = append(paths, l.Annotations[synth.AnnotationFilepath])
	}
	assert.Equal(t, []string{"README.md", "config.json", "model.safetensors", "sub/modeling.py"}, paths)

	exist, desc, err := a.ManifestExist(repoTiny, "main")
	require.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, dTag, desc.Digest)
	assert.Equal(t, v1.MediaTypeImageManifest, desc.MediaType)
	mf, _, _ := a.PullManifest(repoTiny, "main")
	_, payload, _ := mf.Payload()
	assert.Equal(t, int64(len(payload)), desc.Size)

	for _, ref := range []string{"nope", "cccccccccccccccccccccccccccccccccccccccd", digest.FromString("x").String()} {
		exist, desc, err = a.ManifestExist(repoTiny, ref)
		assert.NoError(t, err, ref)
		assert.False(t, exist, ref)
		assert.Nil(t, desc)
		_, _, err = a.PullManifest(repoTiny, ref)
		assert.True(t, errors.IsNotFoundErr(err), ref)
	}

	// A layer digest is not a manifest.
	_, _, err = a.PullManifest(repoTiny, byTag.Layers[0].Digest.String())
	assert.True(t, errors.IsNotFoundErr(err), err)
}

func TestPullBlob(t *testing.T) {
	model := tinyModel()
	h := hubtest.New(model)
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())
	m, _ := manifestOf(t, a, repoTiny, "main")

	files := model.Commits[commitMain].Files
	for _, l := range m.Layers {
		size, r, err := a.PullBlob(repoTiny, l.Digest.String())
		b := readAll(t, l.Size, r, err)
		assert.Equal(t, l.Size, size)
		assert.Equal(t, l.Digest, digest.FromBytes(b))
		assert.Equal(t, files[l.Annotations[synth.AnnotationFilepath]].Content, b)
	}
	size, r, err := a.PullBlob(repoTiny, m.Config.Digest.String())
	cfg := readAll(t, size, r, err)
	assert.Equal(t, m.Config.Digest, digest.FromBytes(cfg))
	assert.Equal(t, m.Config.Size, size)

	var weight v1.Descriptor
	for _, l := range m.Layers {
		if l.MediaType == synth.MediaTypeWeightRaw {
			weight = l
		}
	}
	size, r, err = a.PullBlobChunk(repoTiny, weight.Digest.String(), weight.Size, 100, 199)
	chunk := readAll(t, 100, r, err)
	assert.Equal(t, int64(100), size)
	assert.Equal(t, bigWeight[100:200], chunk)

	size, r, err = a.PullBlobChunk(repoTiny, m.Config.Digest.String(), m.Config.Size, 1, 3)
	assert.Equal(t, cfg[1:4], readAll(t, size, r, err))

	_, _, err = a.PullBlobChunk(repoTiny, weight.Digest.String(), weight.Size, weight.Size, weight.Size+10)
	assert.True(t, errors.IsErr(err, errors.BadRequestCode), err)
	_, _, err = a.PullBlobChunk(repoTiny, weight.Digest.String(), weight.Size, 5, 4)
	assert.True(t, errors.IsErr(err, errors.BadRequestCode), err)

	exist, err := a.BlobExist(repoTiny, weight.Digest.String())
	assert.NoError(t, err)
	assert.True(t, exist)
	exist, err = a.BlobExist(repoTiny, digest.FromString("nope").String())
	assert.NoError(t, err)
	assert.False(t, exist)

	_, _, err = a.PullBlob(repoTiny, digest.FromString("nope").String())
	assert.True(t, errors.IsNotFoundErr(err), err)
	_, _, err = a.PullBlob(repoTiny, "not-a-digest")
	assert.True(t, errors.IsErr(err, errors.BadRequestCode), err)
}

func TestNonLFSHashedOncePerBlob(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	c := newMemoryCache()

	a := newTestAdapter(t, hubRegistry(h, 1, ""), c)
	_, _ = manifestOf(t, a, repoTiny, commitMain)
	assert.Equal(t, 3, h.Calls("resolve"), "README.md, config.json and sub/modeling.py are downloaded to hash them")

	// Another registry shares the content-addressed blob cache: config.json and modeling.py have
	// the same git blob id at commitV1, only README.md differs.
	h.ResetCalls()
	b := newTestAdapter(t, hubRegistry(h, 2, ""), c)
	_, _ = manifestOf(t, b, repoTiny, commitV1)
	assert.Equal(t, 1, h.Calls("resolve"))
}

func TestListTagsAndReferrers(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	tags, err := a.ListTags(repoTiny)
	require.NoError(t, err)
	assert.Equal(t, []string{"feature-x", "main", "v1"}, tags)

	idx, header, err := a.ListReferrers(repoTiny, digest.FromString("x").String(), "")
	require.NoError(t, err)
	require.NotNil(t, idx)
	assert.NotNil(t, idx.Manifests)
	assert.Empty(t, idx.Manifests)
	assert.Equal(t, 2, idx.SchemaVersion)
	assert.NotNil(t, header)
}

func TestWarmPathHubCalls(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	c := newMemoryCache()

	// Cold: refs and snapshot, each via the Hub's 307 from the lowercase repository to the model ID.
	first := newTestAdapter(t, hubRegistry(h, 1, ""), c)
	_, _, err := first.ManifestExist(repoTiny, "main")
	require.NoError(t, err)
	assert.Equal(t, 2, h.Calls("refs"))
	assert.Equal(t, 2, h.Calls("revision"))

	// Once the model ID is known, a new commit costs no redirect.
	h.ResetCalls()
	_, _, err = first.ManifestExist(repoTiny, commitV1)
	require.NoError(t, err)
	assert.Equal(t, 1, h.Calls("revision"))

	// A proxied HEAD calls ManifestExist twice, each on a new adapter.
	h.ResetCalls()
	for range 2 {
		exist, _, err := newTestAdapter(t, hubRegistry(h, 1, ""), c).ManifestExist(repoTiny, "main")
		require.NoError(t, err)
		assert.True(t, exist)
	}
	assert.Equal(t, 0, h.Calls("refs"))
	assert.Equal(t, 0, h.Calls("revision"))
	assert.Equal(t, 0, h.Calls("resolve"))
}

func TestRegistryScopedCache(t *testing.T) {
	m := tinyModel()
	m.Gated = true
	h := hubtest.New(m)
	h.Token = "hf_token"
	defer h.Close()
	c := newMemoryCache()

	withToken := newTestAdapter(t, hubRegistry(h, 1, "hf_token"), c)
	manifest, dgst := manifestOf(t, withToken, repoTiny, "main")

	anonymous := newTestAdapter(t, hubRegistry(h, 2, ""), c)
	exist, _, err := anonymous.ManifestExist(repoTiny, "main")
	assert.False(t, exist)
	assert.True(t, errors.IsErr(err, errors.UnAuthorizedCode), err)
	_, _, err = anonymous.PullManifest(repoTiny, dgst.String())
	assert.True(t, errors.IsErr(err, errors.UnAuthorizedCode), err)
	_, _, err = anonymous.PullBlob(repoTiny, manifest.Layers[0].Digest.String())
	assert.True(t, errors.IsErr(err, errors.UnAuthorizedCode), err)
}

func TestMissPathAfterBranchMoved(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()

	_, oldDigest := manifestOf(t, newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache()), repoTiny, "main")
	h.SetBranch("Org/Tiny-Model", "main", commitNew)
	h.SetBranch("Org/Tiny-Model", "feature/x", commitNew)

	// commitMain is no longer a ref head: its digest is gone after cache loss.
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())
	exist, _, err := a.ManifestExist(repoTiny, oldDigest.String())
	require.NoError(t, err)
	assert.False(t, exist)

	// Re-requesting by tag refills the index; the pinned commit still resolves.
	_, newDigest := manifestOf(t, a, repoTiny, "main")
	assert.NotEqual(t, oldDigest, newDigest)
	_, byCommit := manifestOf(t, a, repoTiny, commitMain)
	assert.Equal(t, oldDigest, byCommit)
	exist, _, err = a.ManifestExist(repoTiny, oldDigest.String())
	require.NoError(t, err)
	assert.True(t, exist)
}

func TestRateLimitPropagates(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	h.RateLimit("refs", 2)
	exist, _, err := a.ManifestExist(repoTiny, "main")
	assert.False(t, exist)
	assert.True(t, errors.IsRateLimitError(err), err)
}

func TestUnknownModel(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	// The anonymous Hub answers a bare 401 for a model that does not exist.
	exist, desc, err := a.ManifestExist("org/does-not-exist", "main")
	assert.NoError(t, err)
	assert.False(t, exist)
	assert.Nil(t, desc)
	_, _, err = a.PullBlob("org/does-not-exist", digest.FromString("x").String())
	assert.True(t, errors.IsNotFoundErr(err), err)
}

func TestDigestIndependentOfEndpoint(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	other := hubRegistry(h, 2, "")
	other.URL = strings.Replace(h.URL, "127.0.0.1", "localhost", 1) + "/"
	require.NotEqual(t, h.URL, other.URL)

	_, d1 := manifestOf(t, newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache()), repoTiny, "main")
	b := newTestAdapter(t, other, newMemoryCache())
	assert.Equal(t, strings.TrimSuffix(other.URL, "/"), b.hub.Endpoint())
	m, d2 := manifestOf(t, b, repoTiny, "main")
	assert.Equal(t, d1, d2)

	size, r, err := b.PullBlob(repoTiny, m.Config.Digest.String())
	assert.Contains(t, string(readAll(t, size, r, err)), `"sourceURL":"https://huggingface.co/Org/Tiny-Model"`)
}

func TestMissPathMainFirstAndNegativeCache(t *testing.T) {
	model := tinyModel()
	h := hubtest.New(model)
	defer h.Close()
	h.SetBranch(model.ID, "main", commitNew)

	_, _ = manifestOf(t, newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache()), repoTiny, commitNew)
	readme := digest.FromBytes(model.Commits[commitNew].Files["README.md"].Content)

	// Refs name commitMain and commitV1 as well, which sort before commitNew; main is tried first.
	h.ResetCalls()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())
	size, r, err := a.PullBlob(repoTiny, readme.String())
	readAll(t, size, r, err)
	assert.Equal(t, 2, h.Calls("revision"), "one snapshot, doubled by the case redirect")

	// An unknown digest costs one pass over the ref heads, then nothing while it is cached as missing.
	h.ResetCalls()
	unknown := digest.FromString("unknown")
	_, _, err = a.PullBlob(repoTiny, unknown.String())
	assert.True(t, errors.IsNotFoundErr(err), err)
	assert.Equal(t, 1, h.Calls("refs"))
	assert.Equal(t, 2, h.Calls("revision"), "commitMain and commitV1; commitNew is cached")
	h.ResetCalls()
	for range 3 {
		exist, err := a.BlobExist(repoTiny, unknown.String())
		assert.NoError(t, err)
		assert.False(t, exist)
	}
	assert.Zero(t, h.Calls("refs"))
	assert.Zero(t, h.Calls("revision"))
}

func TestHeadsMainFirst(t *testing.T) {
	assert.Equal(t, []string{commitNew, commitMain, commitV1}, headsMainFirst([]hub.Ref{
		{Name: "v1", Commit: commitV1}, {Name: "x", Commit: commitMain}, {Name: "main", Commit: commitNew}, {Name: "y", Commit: commitNew},
	}))
	assert.Equal(t, []string{commitMain, commitV1}, headsMainFirst([]hub.Ref{{Name: "v1", Commit: commitV1}, {Name: "x", Commit: commitMain}}))
	assert.Empty(t, headsMainFirst(nil))
}
