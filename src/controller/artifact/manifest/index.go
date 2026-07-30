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

package manifest

import (
	"context"
	"encoding/json"

	"github.com/docker/distribution/manifest/manifestlist"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

func init() {
	mediaTypes := []string{v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList}
	if err := Register(NewIndex(pkg.ArtifactMgr), mediaTypes...); err != nil {
		log.Errorf("failed to register the abstractor for manifest media types %v: %v", mediaTypes, err)
	}
}

// indexAbstractor abstracts artifacts enveloped by OCI index or docker manifest list.
type indexAbstractor struct {
	artMgr artifact.Manager
}

// NewIndex returns an abstractor for OCI indexes and docker manifest lists.
func NewIndex(artMgr artifact.Manager) Abstractor {
	return &indexAbstractor{artMgr: artMgr}
}

func (a *indexAbstractor) Abstract(ctx context.Context, art *artifact.Artifact, content []byte) error {
	// the identity of index is still in progress, we use the manifest mediaType
	// as the media type of artifact
	art.MediaType = art.ManifestMediaType

	index := &v1.Index{}
	if err := json.Unmarshal(content, index); err != nil {
		return err
	}

	/*
		https://github.com/opencontainers/distribution-spec/blob/v1.1.0/spec.md#listing-referrers
		For referrers list, If the artifactType is empty or missing in an index, the artifactType MUST be omitted.
	*/
	if index.ArtifactType != "" {
		art.ArtifactType = index.ArtifactType
	} else {
		art.ArtifactType = ""
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
