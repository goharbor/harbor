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

	"github.com/goharbor/harbor/src/controller/artifact/manifest"
	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// Abstractor abstracts the metadata of artifact
type Abstractor interface {
	// AbstractMetadata abstracts the metadata for the specific artifact type into the artifact model,
	AbstractMetadata(ctx context.Context, artifact *artifact.Artifact) error
}

// NewAbstractor creates a new abstractor
func NewAbstractor() Abstractor {
	return &abstractor{
		regCli: registry.Cli,
	}
}

type abstractor struct {
	regCli registry.Client
}

func (a *abstractor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact) error {
	// read m content
	m, _, err := a.regCli.PullManifest(artifact.RepositoryName, artifact.Digest)
	if err != nil {
		return err
	}
	manifestMediaType, content, err := m.Payload()
	if err != nil {
		return err
	}
	artifact.ManifestMediaType = manifestMediaType

	manifest, err := manifest.Get(manifestMediaType)
	if err != nil {
		return err
	}
	if err := manifest.AbstractManifestMetadata(ctx, artifact, content); err != nil {
		return err
	}

	return processor.Get(artifact.ResolveArtifactType()).AbstractMetadata(ctx, artifact, content)
}
