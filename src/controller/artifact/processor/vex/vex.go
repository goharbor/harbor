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

package vex

import (
	"context"
	"encoding/json"
	"io"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

const (
	// AdditionTypeVEX identifies the VEX document addition.
	AdditionTypeVEX = "VEX"
	// ArtifactTypeVEX is the UI type for VEX artifacts.
	ArtifactTypeVEX = "VEX"
	// ProcessorMediaTypeOpenVEX is the OCI artifact type for OpenVEX documents.
	ProcessorMediaTypeOpenVEX = "application/vnd.openvex+json"
)

func init() {
	pc := &Processor{ManifestProcessor: base.NewManifestProcessor()}
	if err := processor.Register(pc, ProcessorMediaTypeOpenVEX); err != nil {
		log.Errorf("failed to register processor for media type %s: %v", ProcessorMediaTypeOpenVEX, err)
	}
}

// Processor processes VEX artifacts (such as OpenVEX).
type Processor struct {
	*base.ManifestProcessor
}

// ListAdditionTypes returns the VEX document addition.
func (p *Processor) ListAdditionTypes(_ context.Context, _ *artifact.Artifact) []string {
	return []string{AdditionTypeVEX}
}

// AbstractAddition reads the VEX document from its manifest layer.
func (p *Processor) AbstractAddition(_ context.Context, art *artifact.Artifact, _ string) (*processor.Addition, error) {
	manifest, _, err := p.RegCli.PullManifest(art.RepositoryName, art.Digest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull manifest")
	}
	_, payload, err := manifest.Payload()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get payload")
	}
	vexManifest := &v1.Manifest{}
	if err := json.Unmarshal(payload, vexManifest); err != nil {
		return nil, err
	}
	if len(vexManifest.Layers) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).WithMessage("The VEX document is not found")
	}
	_, blob, err := p.RegCli.PullBlob(art.RepositoryName, vexManifest.Layers[0].Digest.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to pull the blob")
	}
	defer blob.Close()
	content, err := io.ReadAll(blob)
	if err != nil {
		return nil, err
	}
	return &processor.Addition{Content: content, ContentType: "application/json"}, nil
}

// GetArtifactType returns the type displayed for VEX artifacts.
func (p *Processor) GetArtifactType(_ context.Context, _ *artifact.Artifact) string {
	return ArtifactTypeVEX
}
