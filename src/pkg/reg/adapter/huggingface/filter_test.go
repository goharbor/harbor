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

	trans "github.com/goharbor/harbor/src/controller/replication/transfer"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/synth"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func manifestDigest(t *testing.T, a *adapter, ref string) digest.Digest {
	t.Helper()
	_, d, err := a.PullManifest(repoTiny, ref)
	require.NoError(t, err)
	return digest.Digest(d)
}

func layerPaths(t *testing.T, a *adapter, ref string) []string {
	t.Helper()
	m, _, err := a.PullManifest(repoTiny, ref)
	require.NoError(t, err)
	_, payload, err := m.Payload()
	require.NoError(t, err)
	var parsed v1.Manifest
	require.NoError(t, json.Unmarshal(payload, &parsed))
	var paths []string
	for _, l := range parsed.Layers {
		paths = append(paths, l.Annotations[synth.AnnotationFilepath])
	}
	return paths
}

func TestFileFilterDeterminism(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()

	adapterWith := func(f *model.FileFilter) *adapter {
		a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())
		require.NoError(t, a.ConfigureSource(&model.SourceOptions{FileFilter: f}))
		return a
	}

	full := manifestDigest(t, adapterWith(nil), "main")
	jsonOnly := &model.FileFilter{Include: []string{"*.json"}}
	filtered := manifestDigest(t, adapterWith(jsonOnly), "main")

	assert.NotEqual(t, full, filtered, "a filter that drops files changes the digest")
	assert.Equal(t, filtered, manifestDigest(t, adapterWith(jsonOnly), "main"), "same commit and filter, same digest")
	assert.Equal(t, filtered, manifestDigest(t, adapterWith(&model.FileFilter{Include: []string{"*.json", "*.json"}}), "main"),
		"duplicate patterns select the same files")
	assert.Equal(t, []string{"config.json"}, layerPaths(t, adapterWith(jsonOnly), "main"))

	noWeights := &model.FileFilter{Exclude: []string{"*.safetensors"}}
	assert.Equal(t, []string{"README.md", "config.json", "sub/modeling.py"}, layerPaths(t, adapterWith(noWeights), "main"))
	assert.NotEqual(t, filtered, manifestDigest(t, adapterWith(noWeights), "main"))

	// The digest records the selected files, not the filter: selecting everything is the full artifact.
	assert.Equal(t, full, manifestDigest(t, adapterWith(&model.FileFilter{Include: []string{"**"}}), "main"))

	// A pattern with a slash matches the full path.
	assert.Equal(t, []string{"sub/modeling.py"}, layerPaths(t, adapterWith(&model.FileFilter{Include: []string{"sub/**"}}), "main"))

	_, _, err := adapterWith(&model.FileFilter{Include: []string{"*.gguf"}}).PullManifest(repoTiny, "main")
	require.Error(t, err)
	assert.True(t, errors.IsErr(err, errors.BadRequestCode))
}

func TestFileFilterSkipsHashingUnselectedFiles(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())
	require.NoError(t, a.ConfigureSource(&model.SourceOptions{FileFilter: &model.FileFilter{Include: []string{"config.json"}}}))
	manifestDigest(t, a, "main")
	assert.Equal(t, 1, h.Calls("resolve"), "only the selected non-LFS file is downloaded")
}

// A filtered manifest and config are served by an adapter without the filter, e.g. after the
// digest index is the only state left.
func TestFilteredDigestsLocatableWithoutFilter(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	c := newMemoryCache()
	filtered := newTestAdapter(t, hubRegistry(h, 1, ""), c)
	require.NoError(t, filtered.ConfigureSource(&model.SourceOptions{FileFilter: &model.FileFilter{Include: []string{"*.json"}}}))
	m, d, err := filtered.PullManifest(repoTiny, "main")
	require.NoError(t, err)

	plain := newTestAdapter(t, hubRegistry(h, 1, ""), c)
	_, again, err := plain.PullManifest(repoTiny, d)
	require.NoError(t, err)
	assert.Equal(t, d, again)
	for _, ref := range m.References() {
		size, body, err := plain.PullBlob(repoTiny, ref.Digest.String())
		require.NoError(t, err)
		body.Close()
		assert.Equal(t, ref.Size, size)
	}
}

func TestConfigureSourceRejectsMalformedFilter(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())
	require.Error(t, a.ConfigureSource(&model.SourceOptions{FileFilter: &model.FileFilter{Include: []string{"[a"}}}))
}

// TestReplicationTransferWithFileFilter runs the real transfer with source options carried on
// the source resource, as the copy flow serializes them into the job parameters.
func TestReplicationTransferWithFileFilter(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	destination.reset()
	clearProcessCache(t)

	factory, err := adp.GetFactory(model.RegistryTypeHuggingFace)
	require.NoError(t, err)
	src, err := factory.Create(hubRegistry(h, 8, ""))
	require.NoError(t, err)
	resources, err := src.(adp.ArtifactRegistry).FetchArtifacts([]*model.Filter{
		{Type: model.FilterTypeName, Value: "org/tiny-model"},
		{Type: model.FilterTypeTag, Value: "main"},
		{Type: model.FilterTypeFile, Value: "*.json,*.md"},
	})
	require.NoError(t, err)
	require.Len(t, resources, 1)

	fileFilter, err := model.FileFilterOf([]*model.Filter{{Type: model.FilterTypeFile, Value: "*.json,*.md"}})
	require.NoError(t, err)
	srcRes := resources[0]
	srcRes.SourceOptions = &model.SourceOptions{FileFilter: fileFilter}
	// The job parameters are JSON.
	data, err := json.Marshal(srcRes)
	require.NoError(t, err)
	srcRes = &model.Resource{}
	require.NoError(t, json.Unmarshal(data, srcRes))

	dstRes := &model.Resource{
		Type:     srcRes.Type,
		Registry: &model.Registry{Type: memoryRegistryType},
		Metadata: &model.ResourceMetadata{
			Repository: &model.Repository{Name: "library/tiny-model"},
			Artifacts:  srcRes.Metadata.Artifacts,
		},
	}
	run := func() digest.Digest {
		t.Helper()
		tf, err := trans.GetFactory(srcRes.Type)
		require.NoError(t, err)
		tr, err := tf(log.DefaultLogger(), func() bool { return false })
		require.NoError(t, err)
		require.NoError(t, tr.Transfer(srcRes, dstRes, &trans.Options{}))
		exist, desc, err := destination.ManifestExist("library/tiny-model", "main")
		require.NoError(t, err)
		require.True(t, exist)
		return desc.Digest
	}

	first := run()
	var parsed v1.Manifest
	require.NoError(t, json.Unmarshal(destination.manifests["library/tiny-model:main"], &parsed))
	var paths []string
	for _, l := range parsed.Layers {
		paths = append(paths, l.Annotations[synth.AnnotationFilepath])
	}
	assert.Equal(t, []string{"README.md", "config.json"}, paths)
	assert.NotContains(t, destination.blobs, digest.FromBytes(bigWeight).String(), "the excluded weight is never copied")

	clearProcessCache(t)
	assert.Equal(t, first, run(), "a rerun after cache loss gives the same digest")
	assert.Equal(t, 1, destination.manifestPushs, "the rerun skips the existing artifact")
}
