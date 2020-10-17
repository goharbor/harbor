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

package base

import (
	"context"
	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// NewIndexProcessor creates a new base index processor.
func NewIndexProcessor() *IndexProcessor {
	return &IndexProcessor{
		RegCli: registry.Cli,
	}
}

// IndexProcessor is a base processor to process artifact enveloped by OCI index or docker manifest list
// Currently, it is just a null implementation
type IndexProcessor struct {
	RegCli registry.Client
}

// AbstractMetadata abstracts metadata of artifact
func (m *IndexProcessor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact, content []byte) error {
	return nil
}

// AbstractAddition abstracts the addition of artifact
func (m *IndexProcessor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*processor.Addition, error) {
	return nil, errors.New(nil).WithCode(errors.BadRequestCode).
		WithMessage("addition %s isn't supported", addition)
}

// GetArtifactType returns the artifact type
func (m *IndexProcessor) GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string {
	return ""
}

// ListAdditionTypes returns the supported addition types
func (m *IndexProcessor) ListAdditionTypes(ctx context.Context, artifact *artifact.Artifact) []string {
	return nil
}
