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

package artifact

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/api/artifact/processor"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// Abstractor abstracts the metadata of artifact
type Abstractor interface {
	// AbstractMetadata abstracts the metadata for the specific artifact type into the artifact model,
	AbstractMetadata(ctx context.Context, artifact *artifact.Artifact) error
}

// NewAbstractor creates a new abstractor
func NewAbstractor() Abstractor {
	return &abstractor{
		artMgr: artifact.Mgr,
		regCli: registry.Cli,
	}
}

type abstractor struct {
	artMgr artifact.Manager
	regCli registry.Client
}

func (a *abstractor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact) error {
	// read manifest content
	manifest, _, err := a.regCli.PullManifest(artifact.RepositoryName, artifact.Digest)
	if err != nil {
		return err
	}
	manifestMediaType, content, err := manifest.Payload()
	if err != nil {
		return err
	}
	artifact.ManifestMediaType = manifestMediaType

	switch artifact.ManifestMediaType {
	case "", "application/json", schema1.MediaTypeSignedManifest:
		a.abstractManifestV1Metadata(artifact)
	case v1.MediaTypeImageManifest, schema2.MediaTypeManifest:
		if err = a.abstractManifestV2Metadata(content, artifact); err != nil {
			return err
		}
	case v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList:
		if err = a.abstractIndexMetadata(ctx, content, artifact); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported manifest media type: %s", artifact.ManifestMediaType)
	}
	return processor.Get(artifact.MediaType).AbstractMetadata(ctx, content, artifact)
}

// the artifact is enveloped by docker manifest v1
func (a *abstractor) abstractManifestV1Metadata(artifact *artifact.Artifact) {
	// unify the media type of v1 manifest to "schema1.MediaTypeSignedManifest"
	artifact.ManifestMediaType = schema1.MediaTypeSignedManifest
	// as no config layer in the docker v1 manifest, use the "schema1.MediaTypeSignedManifest"
	// as the media type of artifact
	artifact.MediaType = schema1.MediaTypeSignedManifest
	// there is no layer size in v1 manifest, doesn't set the artifact size
}

// the artifact is enveloped by OCI manifest or docker manifest v2
func (a *abstractor) abstractManifestV2Metadata(content []byte, artifact *artifact.Artifact) error {
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
	return nil
}

// the artifact is enveloped by OCI index or docker manifest list
func (a *abstractor) abstractIndexMetadata(ctx context.Context, content []byte, art *artifact.Artifact) error {
	// the identity of index is still in progress, we use the manifest mediaType
	// as the media type of artifact
	art.MediaType = art.ManifestMediaType

	index := &v1.Index{}
	if err := json.Unmarshal(content, index); err != nil {
		return err
	}

	// set annotations
	art.Annotations = index.Annotations

	art.Size += int64(len(content))
	// populate the referenced artifacts
	for _, mani := range index.Manifests {
		digest := mani.Digest.String()
		// make sure the child artifact exist
		ar, err := a.artMgr.GetByDigest(ctx, art.RepositoryName, digest)
		if err != nil {
			return err
		}
		art.Size += ar.Size
		art.References = append(art.References, &artifact.Reference{
			ChildID:     ar.ID,
			ChildDigest: digest,
			Platform:    mani.Platform,
			URLs:        mani.URLs,
			Annotations: mani.Annotations,
		})
	}

	// Currently, CNAB put its media type inside the annotations
	// try to parse the artifact media type from the annotations
	if art.Annotations != nil {
		mediaType := art.Annotations["org.opencontainers.artifactType"]
		if len(mediaType) > 0 {
			art.MediaType = mediaType
		}
	}

	return nil
}
