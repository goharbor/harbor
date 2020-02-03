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

package chart

import (
	"context"
	"encoding/json"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/blob"
	resolv "github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/repository"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// ArtifactTypeChart defines the artifact type for helm chart
	ArtifactTypeChart = "CHART"
	// TODO import it from helm chart repository
	mediaType = "application/vnd.cncf.helm.config.v1+json"
)

func init() {
	resolver := &resolver{
		repoMgr:     repository.Mgr,
		blobFetcher: blob.Fcher,
	}
	if err := resolv.Register(resolver, mediaType); err != nil {
		log.Errorf("failed to register resolver for artifact %s: %v", resolver.ArtifactType(), err)
		return
	}
}

type resolver struct {
	repoMgr     repository.Manager
	blobFetcher blob.Fetcher
}

func (r *resolver) ArtifactType() string {
	return ArtifactTypeChart
}

func (r *resolver) Resolve(ctx context.Context, manifest []byte, artifact *artifact.Artifact) error {
	repository, err := r.repoMgr.Get(ctx, artifact.RepositoryID)
	if err != nil {
		return err
	}
	m := &v1.Manifest{}
	if err := json.Unmarshal(manifest, m); err != nil {
		return err
	}
	digest := m.Config.Digest.String()
	layer, err := r.blobFetcher.FetchLayer(repository.Name, digest)
	if err != nil {
		return err
	}
	// TODO should we abstract all values?
	metadata := map[string]interface{}{}
	if err := json.Unmarshal(layer, &metadata); err != nil {
		return err
	}
	if artifact.ExtraAttrs == nil {
		artifact.ExtraAttrs = map[string]interface{}{}
	}
	for k, v := range metadata {
		artifact.ExtraAttrs[k] = v
	}

	return nil
}
