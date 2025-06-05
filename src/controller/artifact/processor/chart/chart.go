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

package chart

import (
	"context"
	"encoding/json"
	"io"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	ps "github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/chart"
)

// const definitions
const (
	// ArtifactTypeChart defines the artifact type for helm chart
	ArtifactTypeChart        = "CHART"
	AdditionTypeValues       = "VALUES.YAML"
	AdditionTypeReadme       = "README.MD"
	AdditionTypeDependencies = "DEPENDENCIES"

	// as helm put the media type definition under "internal" package, we cannot
	// import it, defines it by ourselves
	mediaType = "application/vnd.cncf.helm.config.v1+json"
)

func init() {
	pc := &processor{
		chartOperator: chart.Optr,
	}
	pc.ManifestProcessor = base.NewManifestProcessor()
	if err := ps.Register(pc, mediaType); err != nil {
		log.Errorf("failed to register processor for media type %s: %v", mediaType, err)
		return
	}
}

type processor struct {
	*base.ManifestProcessor
	chartOperator chart.Operator
}

func (p *processor) AbstractAddition(_ context.Context, artifact *artifact.Artifact, addition string) (*ps.Addition, error) {
	if addition != AdditionTypeValues && addition != AdditionTypeReadme && addition != AdditionTypeDependencies {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("addition %s isn't supported for %s", addition, ArtifactTypeChart)
	}

	m, _, err := p.RegCli.PullManifest(artifact.RepositoryName, artifact.Digest)
	if err != nil {
		return nil, err
	}
	_, payload, err := m.Payload()
	if err != nil {
		return nil, err
	}
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(payload, manifest); err != nil {
		return nil, err
	}

	for _, layer := range manifest.Layers {
		// chart do have two layers, one is config, we should resolve the other one.
		layerDgst := layer.Digest.String()
		if layerDgst != manifest.Config.Digest.String() {
			_, blob, err := p.RegCli.PullBlob(artifact.RepositoryName, layerDgst)
			if err != nil {
				return nil, err
			}
			defer blob.Close()
			content, err := io.ReadAll(blob)
			if err != nil {
				return nil, err
			}
			chartDetails, err := p.chartOperator.GetDetails(content)
			if err != nil {
				return nil, err
			}

			var additionContent []byte
			var additionContentType string

			switch addition {
			case AdditionTypeValues:
				additionContent = []byte(chartDetails.Files[AdditionTypeValues])
				additionContentType = "text/plain; charset=utf-8"
			case AdditionTypeReadme:
				additionContent = []byte(chartDetails.Files[AdditionTypeReadme])
				additionContentType = "text/markdown; charset=utf-8"
			case AdditionTypeDependencies:
				additionContent, err = json.Marshal(chartDetails.Dependencies)
				if err != nil {
					return nil, err
				}
				additionContentType = "application/json; charset=utf-8"
			}

			return &ps.Addition{
				Content:     additionContent,
				ContentType: additionContentType,
			}, nil
		}
	}
	return nil, nil
}

func (p *processor) GetArtifactType(_ context.Context, _ *artifact.Artifact) string {
	return ArtifactTypeChart
}

func (p *processor) ListAdditionTypes(_ context.Context, _ *artifact.Artifact) []string {
	return []string{AdditionTypeValues, AdditionTypeReadme, AdditionTypeDependencies}
}
