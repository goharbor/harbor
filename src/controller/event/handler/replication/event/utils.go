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
	"fmt"
	"slices"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/tag"
)

// artifactNode for discover the accessories and child artifacts
type artifactNode struct {
	artifact    model.Artifact
	children    []string
	accessories []string
}

// findAccessories traverses index children and nested accessories. Only
// accessories are returned; the transfer layer already cover copies of index children.
// Event roots are excluded so their original tags remain unchanged.
func findAccessories(ctx context.Context, resource *model.Resource, get func(context.Context, string, string) (*artifactNode, error)) ([]*model.Artifact, error) {
	repo := resource.Metadata.Repository.Name
	var result []*model.Artifact
	nodes := map[string]*artifactNode{}
	included := map[string]bool{}
	for _, root := range resource.Metadata.Artifacts {
		included[root.Digest] = true
	}
	var walk func(string, bool) error
	walk = func(digest string, include bool) error {
		node, seen := nodes[digest]
		if !seen {
			var err error
			node, err = get(ctx, repo, digest)
			if err != nil {
				return fmt.Errorf("get replication artifact %s@%s: %w", repo, digest, err)
			}
			nodes[digest] = node
		}
		if include && !included[digest] {
			art := node.artifact
			art.Tags = slices.Clone(art.Tags)
			result = append(result, &art)
			included[digest] = true
		}
		// skip duplicate traverse
		if seen {
			return nil
		}
		for _, child := range node.children {
			if err := walk(child, false); err != nil {
				return err
			}
		}
		for _, acc := range node.accessories {
			if err := walk(acc, true); err != nil {
				return err
			}
		}
		return nil
	}
	for _, root := range resource.Metadata.Artifacts {
		if err := walk(root.Digest, false); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// getArtifactNode loads metadata and relationships within the event repository.
func getArtifactNode(ctx context.Context, repository, digest string) (*artifactNode, error) {
	art, err := pkg.ArtifactMgr.GetByDigest(ctx, repository, digest)
	if err != nil {
		return nil, err
	}
	n := &artifactNode{artifact: model.Artifact{Type: art.Type, Digest: art.Digest}}
	tags, err := tag.Mgr.List(ctx, q.New(q.KeyWords{"ArtifactID": art.ID}))
	if err != nil {
		return nil, err
	}
	for _, t := range tags {
		n.artifact.Tags = append(n.artifact.Tags, t.Name)
	}
	for _, ref := range art.References {
		n.children = append(n.children, ref.ChildDigest)
	}
	accessories, err := accessory.Mgr.List(ctx, q.New(q.KeyWords{"SubjectArtifactRepo": repository, "SubjectArtifactDigest": digest}))
	if err != nil {
		return nil, err
	}
	for _, a := range accessories {
		n.accessories = append(n.accessories, a.GetData().Digest)
	}
	return n, nil
}
