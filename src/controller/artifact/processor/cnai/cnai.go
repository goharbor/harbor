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

package cnai

import (
	"context"
	"encoding/json"

	modelspec "github.com/CloudNativeAI/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	ps "github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/controller/artifact/processor/cnai/parser"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// const definitions
const (
	// ArtifactTypeCNAI defines the artifact type for CNAI model.
	ArtifactTypeCNAI = "CNAI"

	// AdditionTypeReadme defines the addition type readme for API.
	AdditionTypeReadme = "README.MD"
	// AdditionTypeLicense defines the addition type license for API.
	AdditionTypeLicense = "LICENSE"
	// AdditionTypeFiles defines the addition type files for API.
	AdditionTypeFiles = "FILES"
)

func init() {
	pc := &processor{
		ManifestProcessor: base.NewManifestProcessor(),
	}

	if err := ps.Register(pc, modelspec.ArtifactTypeModelManifest); err != nil {
		log.Errorf("failed to register processor for artifact type %s: %v", modelspec.ArtifactTypeModelManifest, err)
		return
	}
}

type processor struct {
	*base.ManifestProcessor
}

func (p *processor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*ps.Addition, error) {
	var additionParser parser.Parser
	switch addition {
	case AdditionTypeReadme:
		additionParser = parser.NewReadme(p.RegCli)
	case AdditionTypeLicense:
		additionParser = parser.NewLicense(p.RegCli)
	case AdditionTypeFiles:
		additionParser = parser.NewFiles(p.RegCli)
	default:
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("addition %s isn't supported for %s", addition, ArtifactTypeCNAI)
	}

	mf, _, err := p.RegCli.PullManifest(artifact.RepositoryName, artifact.Digest)
	if err != nil {
		return nil, err
	}

	_, payload, err := mf.Payload()
	if err != nil {
		return nil, err
	}

	manifest := &ocispec.Manifest{}
	if err := json.Unmarshal(payload, manifest); err != nil {
		return nil, err
	}

	contentType, content, err := additionParser.Parse(ctx, artifact, manifest)
	if err != nil {
		return nil, err
	}

	return &ps.Addition{
		ContentType: contentType,
		Content:     content,
	}, nil
}

func (p *processor) GetArtifactType(_ context.Context, _ *artifact.Artifact) string {
	return ArtifactTypeCNAI
}

func (p *processor) ListAdditionTypes(_ context.Context, _ *artifact.Artifact) []string {
	return []string{AdditionTypeReadme, AdditionTypeLicense, AdditionTypeFiles}
}
