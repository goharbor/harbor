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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	trans "github.com/goharbor/harbor/src/controller/replication/transfer"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const (
	revisionTagMain = "sha-aaaaaaaaaaaa"
	revisionTagV1   = "sha-bbbbbbbbbbbb"
)

func TestFetchArtifactsWithRevisionTag(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	fetch := func(tagFilter string) []*model.Artifact {
		t.Helper()
		res, err := a.FetchArtifacts([]*model.Filter{
			{Type: model.FilterTypeName, Value: "org/tiny-model"},
			{Type: model.FilterTypeTag, Value: tagFilter},
		})
		require.NoError(t, err)
		require.Len(t, res, 1)
		return res[0].Metadata.Artifacts
	}

	without := fetch("**")
	require.Len(t, without, 2)
	assert.Equal(t, []string{"feature-x", "main"}, without[0].Tags)

	require.NoError(t, a.ConfigureSource(&model.SourceOptions{RevisionTag: true}))
	with := fetch("**")
	require.Len(t, with, 2)
	assert.Equal(t, []string{"feature-x", "main", revisionTagMain}, with[0].Tags)
	assert.Equal(t, []string{"v1", revisionTagV1}, with[1].Tags)

	only := fetch("main")
	require.Len(t, only, 1)
	assert.Equal(t, []string{"main", revisionTagMain}, only[0].Tags, "the revision tag follows the tags the filter keeps")
}

func TestPullByRevisionTag(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	a := newTestAdapter(t, hubRegistry(h, 1, ""), newMemoryCache())

	assert.Equal(t, manifestDigest(t, a, "main"), manifestDigest(t, a, revisionTagMain))
	assert.Equal(t, manifestDigest(t, a, "v1"), manifestDigest(t, a, revisionTagV1))

	_, _, err := a.PullManifest(repoTiny, "sha-cccccccccccc")
	assert.True(t, errors.IsNotFoundErr(err), "a commit no ref points at has no revision tag lookup")

	r, err := a.Resolve(repoTiny, "main", nil)
	require.NoError(t, err)
	assert.Equal(t, revisionTagMain, r.RevisionTag)
}

func TestReplicationTransferWithRevisionTag(t *testing.T) {
	h := hubtest.New(tinyModel())
	defer h.Close()
	destination.reset()
	clearProcessCache(t)

	a := newTestAdapter(t, hubRegistry(h, 9, ""), newMemoryCache())
	opts := &model.SourceOptions{RevisionTag: true}
	require.NoError(t, a.ConfigureSource(opts))
	res, err := a.FetchArtifacts([]*model.Filter{
		{Type: model.FilterTypeName, Value: "org/tiny-model"},
		{Type: model.FilterTypeTag, Value: "main"},
	})
	require.NoError(t, err)
	srcRes := res[0]
	srcRes.Registry = hubRegistry(h, 9, "")
	srcRes.SourceOptions = opts
	dstRes := &model.Resource{
		Type:     srcRes.Type,
		Registry: &model.Registry{Type: memoryRegistryType},
		Metadata: &model.ResourceMetadata{
			Repository: &model.Repository{Name: "library/tiny-model"},
			Artifacts:  srcRes.Metadata.Artifacts,
		},
	}
	tf, err := trans.GetFactory(srcRes.Type)
	require.NoError(t, err)
	tr, err := tf(log.DefaultLogger(), func() bool { return false })
	require.NoError(t, err)
	require.NoError(t, tr.Transfer(srcRes, dstRes, &trans.Options{}))

	exist, mainDesc, err := destination.ManifestExist("library/tiny-model", "main")
	require.NoError(t, err)
	require.True(t, exist)
	exist, revDesc, err := destination.ManifestExist("library/tiny-model", revisionTagMain)
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, mainDesc.Digest, revDesc.Digest)
}
