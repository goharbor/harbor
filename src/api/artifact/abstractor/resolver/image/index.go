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

package image

import (
	"context"
	"encoding/json"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/api/artifact/descriptor"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func init() {
	rslver := &indexResolver{
		artMgr: artifact.Mgr,
	}
	mediaTypes := []string{
		v1.MediaTypeImageIndex,
		manifestlist.MediaTypeManifestList,
	}
	if err := resolver.Register(rslver, mediaTypes...); err != nil {
		log.Errorf("failed to register resolver for media type %v: %v", mediaTypes, err)
		return
	}
	if err := descriptor.Register(rslver, mediaTypes...); err != nil {
		log.Errorf("failed to register descriptor for media type %v: %v", mediaTypes, err)
		return
	}
}

// indexResolver resolves artifact with OCI index and docker manifest list
type indexResolver struct {
	artMgr artifact.Manager
}

func (i *indexResolver) ResolveMetadata(ctx context.Context, manifest []byte, art *artifact.Artifact) error {
	index := &v1.Index{}
	if err := json.Unmarshal(manifest, index); err != nil {
		return err
	}
	// populate the referenced artifacts
	for _, mani := range index.Manifests {
		digest := mani.Digest.String()
		// make sure the child artifact exist
		ar, err := i.artMgr.GetByDigest(ctx, art.RepositoryName, digest)
		if err != nil {
			return err
		}
		art.References = append(art.References, &artifact.Reference{
			ChildID:  ar.ID,
			Platform: mani.Platform,
		})
	}
	return nil
}

func (i *indexResolver) ResolveAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*resolver.Addition, error) {
	return nil, ierror.New(nil).WithCode(ierror.BadRequestCode).
		WithMessage("addition %s isn't supported for %s(index)", addition, ArtifactTypeImage)
}

func (i *indexResolver) GetArtifactType() string {
	return ArtifactTypeImage
}

func (i *indexResolver) ListAdditionTypes() []string {
	return nil
}
