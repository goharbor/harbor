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

package wasm

import (
	"context"
	"encoding/json"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// const definitions
const (
	// ArtifactTypeWASM is the artifact type for image
	ArtifactTypeWASM         = "WASM"
	AdditionTypeBuildHistory = "BUILD_HISTORY"

	// AnnotationVariantKey and AnnotationVariantValue is available key-value pair to identify an annotation fashion wasm artifact
	AnnotationVariantKey   = "module.wasm.image/variant"
	AnnotationVariantValue = "compat"

	// AnnotationHandlerKey and AnnotationHandlerValue is another available key-value pair to identify an annotation fashion wasm artifact
	AnnotationHandlerKey   = "run.oci.handler"
	AnnotationHandlerValue = "wasm"

	MediaType = "application/vnd.wasm.config.v1+json"
)

func init() {
	pc := &Processor{}
	pc.ManifestProcessor = base.NewManifestProcessor()
	mediaTypes := []string{
		MediaType,
	}
	if err := processor.Register(pc, mediaTypes...); err != nil {
		log.Errorf("failed to register processor for media type %v: %v", mediaTypes, err)
		return
	}
}

// Processor processes image with OCI manifest and docker v2 manifest
type Processor struct {
	*base.ManifestProcessor
}

func (m *Processor) AbstractMetadata(ctx context.Context, art *artifact.Artifact, manifestBody []byte) error {
	art.ExtraAttrs = map[string]interface{}{}
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(manifestBody, manifest); err != nil {
		return err
	}

	if art.ExtraAttrs == nil {
		art.ExtraAttrs = map[string]interface{}{}
	}
	if manifest.Annotations[AnnotationVariantKey] == AnnotationVariantValue || manifest.Annotations[AnnotationHandlerKey] == AnnotationHandlerValue {
		// for annotation way
		config := &v1.Image{}
		if err := m.UnmarshalConfig(ctx, art.RepositoryName, manifestBody, config); err != nil {
			return err
		}
		art.ExtraAttrs["manifest.config.mediaType"] = manifest.Config.MediaType
		art.ExtraAttrs["created"] = config.Created
		art.ExtraAttrs["architecture"] = config.Architecture
		art.ExtraAttrs["os"] = config.OS
		art.ExtraAttrs["config"] = config.Config
		// if the author is null, try to get it from labels:
		// https://docs.docker.com/engine/reference/builder/#maintainer-deprecated
		author := config.Author
		if len(author) == 0 && len(config.Config.Labels) > 0 {
			author = config.Config.Labels["maintainer"]
		}
		art.ExtraAttrs["author"] = author
	} else {
		// for wasm-to-oci way
		art.ExtraAttrs["manifest.config.mediaType"] = MediaType
		if len(manifest.Layers) > 0 {
			art.ExtraAttrs["manifest.layers.mediaType"] = manifest.Layers[0].MediaType
			art.ExtraAttrs["org.opencontainers.image.title"] = manifest.Layers[0].Annotations["org.opencontainers.image.title"]
		}
	}
	return nil
}

func (m *Processor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*processor.Addition, error) {
	if addition != AdditionTypeBuildHistory {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("addition %s isn't supported for %s(manifest version 2)", addition, ArtifactTypeWASM)
	}

	mani, _, err := m.RegCli.PullManifest(artifact.RepositoryName, artifact.Digest)
	if err != nil {
		return nil, err
	}
	_, content, err := mani.Payload()
	if err != nil {
		return nil, err
	}
	config := &v1.Image{}
	if err = m.ManifestProcessor.UnmarshalConfig(ctx, artifact.RepositoryName, content, config); err != nil {
		return nil, err
	}
	content, err = json.Marshal(config.History)
	if err != nil {
		return nil, err
	}
	return &processor.Addition{
		Content:     content,
		ContentType: "application/json; charset=utf-8",
	}, nil
}

func (m *Processor) GetArtifactType(_ context.Context, _ *artifact.Artifact) string {
	return ArtifactTypeWASM
}

func (m *Processor) ListAdditionTypes(_ context.Context, _ *artifact.Artifact) []string {
	return []string{AdditionTypeBuildHistory}
}
