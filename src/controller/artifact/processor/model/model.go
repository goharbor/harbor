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

package model

import (
	"context"

	ps "github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// const definitions
const (
	// ArtifactTypeChart defines the artifact type for helm chart
	ArtifactTypeChart = "MODEL"

	// as helm put the media type definition under "internal" package, we cannot
	// import it, defines it by ourselves
	mediaType = "application/vnd.caicloud.model.config.v1alpha1+json"
)

func init() {
	pc := &processor{}
	pc.ManifestProcessor = base.NewManifestProcessor()
	if err := ps.Register(pc, mediaType); err != nil {
		log.Errorf("failed to register processor for media type %s: %v", mediaType, err)
		return
	}
}

type processor struct {
	*base.ManifestProcessor
}

// AbstractAddition abstracts the addition of artifact
func (p *processor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*ps.Addition, error) {
	return nil, nil
}

func (p *processor) GetArtifactType() string {
	return ArtifactTypeChart
}

func (p *processor) ListAdditionTypes() []string {
	return []string{}
}
