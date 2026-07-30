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

	"github.com/docker/distribution/manifest/schema2"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/controller/artifact/processor/wasm"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

func init() {
	mediaTypes := []string{v1.MediaTypeImageManifest, schema2.MediaTypeManifest}
	if err := Register(NewV2(), mediaTypes...); err != nil {
		log.Errorf("failed to register the abstractor for manifest media types %v: %v", mediaTypes, err)
	}
}

// v2Abstractor abstracts artifacts enveloped by OCI manifest or docker manifest v2.
type v2Abstractor struct{}

func NewV2() Abstractor {
	return &v2Abstractor{}
}

func (a *v2Abstractor) Abstract(_ context.Context, art *artifact.Artifact, content []byte) error {
	mf := &v1.Manifest{}
	if err := json.Unmarshal(content, mf); err != nil {
		return err
	}
	// use the "manifest.config.mediatype" as the media type of the artifact
	art.MediaType = mf.Config.MediaType
	if mf.Annotations[wasm.AnnotationVariantKey] == wasm.AnnotationVariantValue || mf.Annotations[wasm.AnnotationHandlerKey] == wasm.AnnotationHandlerValue {
		art.MediaType = wasm.MediaType
	}
	/*
		https://github.com/opencontainers/distribution-spec/blob/v1.1.0/spec.md#listing-referrers
		For referrers list, if the artifactType is empty or missing in the image manifest, the value of artifactType MUST be set to the config descriptor mediaType value
	*/
	if mf.ArtifactType != "" {
		art.ArtifactType = mf.ArtifactType
	} else {
		art.ArtifactType = mf.Config.MediaType
	}

	// set size
	art.Size = int64(len(content)) + mf.Config.Size
	for _, layer := range mf.Layers {
		art.Size += layer.Size
	}
	// set annotations
	art.Annotations = mf.Annotations
	return nil
}
