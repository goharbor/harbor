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

package cnab

import (
	"context"
	"encoding/json"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/blob"
	resolv "github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/api/artifact/descriptor"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// const definitions
const (
	ArtifactTypeCNAB = "CNAB"
	mediaType        = "application/vnd.cnab.manifest.v1"
)

func init() {
	resolver := &resolver{
		argMgr:      artifact.Mgr,
		blobFetcher: blob.Fcher,
	}
	if err := resolv.Register(resolver, mediaType); err != nil {
		log.Errorf("failed to register resolver for media type %s: %v", mediaType, err)
		return
	}
	if err := descriptor.Register(resolver, mediaType); err != nil {
		log.Errorf("failed to register descriptor for media type %s: %v", mediaType, err)
		return
	}
}

type resolver struct {
	argMgr      artifact.Manager
	blobFetcher blob.Fetcher
}

func (r *resolver) ResolveMetadata(ctx context.Context, manifest []byte, art *artifact.Artifact) error {
	index := &v1.Index{}
	if err := json.Unmarshal(manifest, index); err != nil {
		return err
	}
	cfgManiDgt := ""
	// populate the referenced artifacts
	for _, mani := range index.Manifests {
		digest := mani.Digest.String()
		// make sure the child artifact exist
		ar, err := r.argMgr.GetByDigest(ctx, art.RepositoryName, digest)
		if err != nil {
			return err
		}
		art.References = append(art.References, &artifact.Reference{
			ChildID:     ar.ID,
			ChildDigest: digest,
			Platform:    mani.Platform,
			URLs:        mani.URLs,
			Annotations: mani.Annotations,
		})
		// try to get the digest of the manifest that the config layer is referenced by
		if mani.Annotations != nil &&
			mani.Annotations["io.cnab.manifest.type"] == "config" {
			cfgManiDgt = mani.Digest.String()
		}
	}
	if len(cfgManiDgt) == 0 {
		return nil
	}

	// resolve the config of CNAB
	// get the manifest that the config layer is referenced by
	_, cfgMani, err := r.blobFetcher.FetchManifest(art.RepositoryName, cfgManiDgt)
	if err != nil {
		return err
	}
	m := &v1.Manifest{}
	if err := json.Unmarshal(cfgMani, m); err != nil {
		return err
	}
	cfgDgt := m.Config.Digest.String()
	// get the config layer
	cfg, err := r.blobFetcher.FetchLayer(art.RepositoryName, cfgDgt)
	if err != nil {
		return err
	}
	metadata := map[string]interface{}{}
	if err := json.Unmarshal(cfg, &metadata); err != nil {
		return err
	}
	if art.ExtraAttrs == nil {
		art.ExtraAttrs = map[string]interface{}{}
	}
	for k, v := range metadata {
		art.ExtraAttrs[k] = v
	}
	return nil
}

func (r *resolver) ResolveAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*resolv.Addition, error) {
	return nil, ierror.New(nil).WithCode(ierror.BadRequestCode).
		WithMessage("addition %s isn't supported for %s", addition, ArtifactTypeCNAB)
}

func (r *resolver) GetArtifactType() string {
	return ArtifactTypeCNAB
}

func (r *resolver) ListAdditionTypes() []string {
	return nil
}
