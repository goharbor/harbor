// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/controller/artifact/manifest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/controller/artifact/processor/wasm"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

func init() {
	v2 := &v2ManifestAbstractor{}
	if err := manifest.Register(v2,
		v1.MediaTypeImageManifest,
		v1.MediaTypeImageLayerGzip,
		schema2.MediaTypeManifest,
	); err != nil {
		log.Errorf("failed to register v2 manifest abstractor: %v", err)
	}
}

// v2ManifestAbstractor handles OCI image manifests and Docker V2 manifests
type v2ManifestAbstractor struct{}

func (a *v2ManifestAbstractor) AbstractManifestMetadata(_ context.Context, artifact *artifact.Artifact, content []byte) error {
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(content, manifest); err != nil {
		return err
	}
	// use the "manifest.config.mediatype" as the media type of the artifact
	artifact.MediaType = manifest.Config.MediaType
	if manifest.Annotations[wasm.AnnotationVariantKey] == wasm.AnnotationVariantValue || manifest.Annotations[wasm.AnnotationHandlerKey] == wasm.AnnotationHandlerValue {
		artifact.MediaType = wasm.MediaType
	}
	/*
		https://github.com/opencontainers/distribution-spec/blob/v1.1.0/spec.md#listing-referrers
		For referrers list, if the artifactType is empty or missing in the image manifest, the value of artifactType MUST be set to the config descriptor mediaType value
	*/
	if manifest.ArtifactType != "" {
		artifact.ArtifactType = manifest.ArtifactType
	} else {
		artifact.ArtifactType = manifest.Config.MediaType
	}

	// set size
	artifact.Size = int64(len(content)) + manifest.Config.Size
	for _, layer := range manifest.Layers {
		artifact.Size += layer.Size
	}
	// set annotations
	artifact.Annotations = manifest.Annotations
	return nil
}
