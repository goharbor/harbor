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

package event

import (
	"context"

	"testing"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/require"
)

type graphReader map[string]*artifactNode

func (g graphReader) Read(_ context.Context, _, digest string) (*artifactNode, error) {
	if n := g[digest]; n != nil {
		return n, nil
	}
	return nil, errors.NotFoundError(nil)
}
func seed(digest string, tags ...string) *model.Resource {
	return &model.Resource{Type: model.ResourceTypeArtifact, Metadata: &model.ResourceMetadata{
		Repository: &model.Repository{Name: "test/harbor-core"}, Artifacts: []*model.Artifact{{Digest: digest, Tags: tags}},
	}}
}
func tagFilter(pattern string) []*model.Filter {
	return []*model.Filter{{Type: model.FilterTypeTag, Value: pattern}}
}
func digests(r *model.Resource) []string {
	var ds []string
	if r != nil {
		for _, a := range r.Metadata.Artifacts {
			ds = append(ds, a.Digest)
		}
	}
	return ds
}

func TestDiscoverIncludesChildAndNestedAccessories(t *testing.T) {
	g := graphReader{
		"root":   {artifact: model.Artifact{Digest: "root", Tags: []string{"other"}}, children: []string{"child"}, accessories: []string{"sig"}},
		"child":  {artifact: model.Artifact{Digest: "child"}, accessories: []string{"sbom"}},
		"sig":    {artifact: model.Artifact{Digest: "sig", Tags: []string{"sig-tag"}}, accessories: []string{"nested"}},
		"nested": {artifact: model.Artifact{Digest: "nested"}},
		"sbom":   {artifact: model.Artifact{Digest: "sbom"}},
	}
	original := seed("root", "release")
	accessories, err := findAccessories(context.Background(), original, g.Read)
	r := &model.Resource{Metadata: &model.ResourceMetadata{Artifacts: accessories}}
	require.NoError(t, err)
	require.Equal(t, []string{"sbom", "sig", "nested"}, digests(r))
	require.Equal(t, []string{"release"}, original.Metadata.Artifacts[0].Tags)
	require.Equal(t, []string{"sig-tag"}, r.Metadata.Artifacts[1].Tags)
	require.Len(t, original.Metadata.Artifacts, 1)
}
func TestDiscoverReadError(t *testing.T) {
	g := graphReader{"root": {accessories: []string{"missing"}}}
	_, err := findAccessories(context.Background(), seed("root"), g.Read)
	require.Error(t, err)
	require.True(t, errors.IsNotFoundErr(err))
	require.ErrorContains(t, err, "test/harbor-core@missing")
}

// The index is an accessory of the manifest it also references as a child.
// These relationships form a cycle only when both edge types are traversed.
func TestDiscoverIndexSubjectBackEdge(t *testing.T) {
	g := graphReader{
		"index":   {artifact: model.Artifact{Digest: "index", Tags: []string{"index-tag"}}, children: []string{"subject"}, accessories: []string{"nested"}},
		"subject": {artifact: model.Artifact{Digest: "subject"}, accessories: []string{"index"}},
		"nested":  {artifact: model.Artifact{Digest: "nested"}},
		"root":    {artifact: model.Artifact{Digest: "root"}, children: []string{"index"}},
	}
	for _, tc := range []struct {
		root string
		want []string
	}{
		{"index", []string{"nested"}},
		{"subject", []string{"index", "nested"}},
		{"root", []string{"index", "nested"}},
	} {
		t.Run(tc.root, func(t *testing.T) {
			reads := map[string]int{}
			read := func(ctx context.Context, repo, digest string) (*artifactNode, error) {
				reads[digest]++
				require.Equal(t, 1, reads[digest], "each node must be read once")
				return g.Read(ctx, repo, digest)
			}
			original := seed(tc.root, "release")
			accessories, err := findAccessories(context.Background(), original, read)
			require.NoError(t, err)
			require.Equal(t, tc.want, digests(&model.Resource{Metadata: &model.ResourceMetadata{Artifacts: accessories}}))
			require.Len(t, original.Metadata.Artifacts, 1)
			require.Equal(t, []string{"release"}, original.Metadata.Artifacts[0].Tags)
		})
	}
}

func TestDiscoverAccessoryPreviouslyVisitedAsChild(t *testing.T) {
	g := graphReader{
		"root": {children: []string{"acc"}, accessories: []string{"acc"}},
		"acc":  {artifact: model.Artifact{Digest: "acc", Tags: []string{"signature-tag"}}},
	}
	accessories, err := findAccessories(context.Background(), seed("root"), g.Read)
	require.NoError(t, err)
	require.Len(t, accessories, 1)
	require.Equal(t, "acc", accessories[0].Digest)
	require.Equal(t, []string{"signature-tag"}, accessories[0].Tags)
}

func TestDiscoverSharedAccessory(t *testing.T) {
	g := graphReader{
		"root":  {children: []string{"child"}, accessories: []string{"sig"}},
		"child": {accessories: []string{"sig"}},
		"sig":   {artifact: model.Artifact{Digest: "sig"}},
	}
	accessories, err := findAccessories(context.Background(), seed("root"), g.Read)
	r := &model.Resource{Metadata: &model.ResourceMetadata{Artifacts: accessories}}
	require.NoError(t, err)
	require.Equal(t, []string{"sig"}, digests(r))
}
