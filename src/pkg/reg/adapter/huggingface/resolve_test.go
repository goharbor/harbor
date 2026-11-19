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
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestResolve(t *testing.T) {
	m := tinyModel()
	m.PipelineTag = "text-generation"
	h := hubtest.New(m)
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	r, err := a.Resolve("Org/Tiny-Model", "", nil)
	require.NoError(t, err)
	assert.Equal(t, repoTiny, r.Repository)
	assert.Equal(t, "Org/Tiny-Model", r.SourceID)
	assert.Equal(t, "main", r.Reference)
	assert.Equal(t, commitMain, r.Revision)
	assert.Equal(t, "https://huggingface.co/Org/Tiny-Model", r.SourceURL)
	assert.Equal(t, []string{"feature-x", "main"}, r.Tags)
	assert.Equal(t, []string{"mit"}, r.Licenses)
	assert.Equal(t, "text-generation", r.Task)
	require.NotNil(t, r.LastModified)
	assert.Equal(t, 2024, r.LastModified.Year())

	files := m.Commits[commitMain].Files
	want := []*model.ResolvedFile{
		{Path: "README.md", Size: int64(len(files["README.md"].Content)), Class: model.FileClassDoc, SHA256: sha256Hex(files["README.md"].Content), Matched: true},
		{Path: "config.json", Size: int64(len(files["config.json"].Content)), Class: model.FileClassConfig, SHA256: sha256Hex(files["config.json"].Content), Matched: true},
		{Path: "model.safetensors", Size: int64(len(bigWeight)), Class: model.FileClassWeight, SHA256: sha256Hex(bigWeight), Matched: true},
		{Path: "sub/modeling.py", Size: int64(len(files["sub/modeling.py"].Content)), Class: model.FileClassCode, SHA256: sha256Hex(files["sub/modeling.py"].Content), Matched: true},
	}
	assert.Equal(t, want, r.Files, ".gitattributes is never listed")
	assert.Equal(t, 4, r.TotalFiles)
	assert.Equal(t, 4, r.MatchedFiles)
	assert.Equal(t, r.TotalSize, r.MatchedSize)

	_, pulled, err := a.PullManifest(repoTiny, "main")
	require.NoError(t, err)
	assert.Equal(t, pulled, r.Digest, "the preview digest is the digest of a copy")
}

func TestResolveWithFileFilter(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())
	jsonOnly := &model.FileFilter{Include: []string{"*.json"}}

	r, err := a.Resolve(repoTiny, "main", jsonOnly)
	require.NoError(t, err)
	assert.Equal(t, 4, r.TotalFiles)
	assert.Equal(t, 1, r.MatchedFiles)
	assert.Equal(t, int64(len(`{"model_type":"tiny"}`)), r.MatchedSize)
	for _, f := range r.Files {
		assert.Equal(t, f.Path == "config.json", f.Matched, f.Path)
		if f.Path == "README.md" {
			assert.Empty(t, f.SHA256, "an unselected non-LFS file is not downloaded")
		}
	}

	copier := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())
	require.NoError(t, copier.ConfigureSource(&model.SourceOptions{FileFilter: jsonOnly}))
	_, pulled, err := copier.PullManifest(repoTiny, "main")
	require.NoError(t, err)
	assert.Equal(t, pulled, r.Digest)

	none, err := a.Resolve(repoTiny, "main", &model.FileFilter{Include: []string{"*.gguf"}})
	require.NoError(t, err)
	assert.Zero(t, none.MatchedFiles)
	assert.Empty(t, none.Digest)
}

func TestResolveReferences(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	r, err := a.Resolve(repoTiny, "v1", nil)
	require.NoError(t, err)
	assert.Equal(t, commitV1, r.Revision)
	assert.Equal(t, []string{"v1"}, r.Tags)

	r, err = a.Resolve(repoTiny, commitNew, nil)
	require.NoError(t, err)
	assert.Equal(t, commitNew, r.Revision)
	assert.Equal(t, []string{}, r.Tags, "a commit no ref points at has no tags")

	_, err = a.Resolve(repoTiny, "nope", nil)
	assert.True(t, errors.IsNotFoundErr(err))

	_, err = a.Resolve("no-slash", "main", nil)
	assert.True(t, errors.IsErr(err, errors.BadRequestCode))

	_, err = a.Resolve(repoTiny, "main", &model.FileFilter{Include: []string{"[a"}})
	assert.True(t, errors.IsErr(err, errors.BadRequestCode))
}
