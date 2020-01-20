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

package abstractor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/blob"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

// Abstractor abstracts the specific information for different types of artifacts
type Abstractor interface {
	// Abstract the specific information for the specific artifact type into the artifact model,
	// the information can be got from the manifest or other layers referenced by the manifest.
	Abstract(ctx context.Context, artifact *artifact.Artifact) error
}

// NewAbstractor returns an instance of the default abstractor
func NewAbstractor() Abstractor {
	return &abstractor{
		repoMgr:     repository.Mgr,
		blobFetcher: blob.Fcher,
	}
}

type abstractor struct {
	repoMgr     repository.Manager
	blobFetcher blob.Fetcher
}

// TODO try CNAB, how to forbid CNAB

// TODO add white list for supported artifact type
func (a *abstractor) Abstract(ctx context.Context, artifact *artifact.Artifact) error {
	repository, err := a.repoMgr.Get(ctx, artifact.RepositoryID)
	if err != nil {
		return err
	}
	// read manifest content
	manifestMediaType, content, err := a.blobFetcher.FetchManifest(repository.Name, artifact.Digest)
	if err != nil {
		return err
	}
	artifact.ManifestMediaType = manifestMediaType

	switch artifact.ManifestMediaType {
	// docker manifest v1
	case "", "application/json", schema1.MediaTypeSignedManifest:
		// unify the media type of v1 manifest to "schema1.MediaTypeSignedManifest"
		artifact.ManifestMediaType = schema1.MediaTypeSignedManifest
		// as no config layer in the docker v1 manifest, use the "schema1.MediaTypeSignedManifest"
		// as the media type of artifact
		artifact.MediaType = schema1.MediaTypeSignedManifest
		// there is no layer size in v1 manifest, doesn't set the artifact size
	// OCI manifest/docker manifest v2
	case v1.MediaTypeImageManifest, schema2.MediaTypeManifest:
		manifest := &v1.Manifest{}
		if err := json.Unmarshal(content, manifest); err != nil {
			return err
		}
		// use the "manifest.config.mediatype" as the media type of the artifact
		artifact.MediaType = manifest.Config.MediaType
		// set size
		artifact.Size = int64(len(content)) + manifest.Config.Size
		for _, layer := range manifest.Layers {
			artifact.Size += layer.Size
		}
		// set annotations
		artifact.Annotations = manifest.Annotations
	// OCI index/docker manifest list
	case v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList:
		// the identity of index is still in progress, only handle image index for now
		// and use the manifestMediaType as the media type of artifact
		// If we want to support CNAB, we should get the media type from annotation
		artifact.MediaType = artifact.ManifestMediaType

		index := &v1.Index{}
		if err := json.Unmarshal(content, index); err != nil {
			return err
		}
		// the size for image index is meaningless, doesn't set it for image index
		// but it is useful for CNAB or other artifacts, set it when needed

		// set annotations
		artifact.Annotations = index.Annotations
		// TODO handle references in resolvers
	default:
		return fmt.Errorf("unsupported manifest media type: %s", artifact.ManifestMediaType)
	}

	resolver, err := resolver.Get(artifact.MediaType)
	if err != nil {
		return err
	}
	artifact.Type = resolver.ArtifactType()
	return resolver.Resolve(ctx, content, artifact)
}
