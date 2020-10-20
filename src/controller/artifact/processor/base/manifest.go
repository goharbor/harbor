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
	"encoding/json"
	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

// NewManifestProcessor creates a new base manifest processor.
// All metadata read from config layer will be populated if specifying no "properties"
func NewManifestProcessor(properties ...string) *ManifestProcessor {
	return &ManifestProcessor{
		properties: properties,
		RegCli:     registry.Cli,
	}
}

// ManifestProcessor is a base processor to process artifact enveloped by OCI manifest or docker v2 manifest
type ManifestProcessor struct {
	properties []string
	RegCli     registry.Client
}

// AbstractMetadata abstracts metadata of artifact
func (m *ManifestProcessor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact, content []byte) error {
	// parse metadata from config layer
	metadata := map[string]interface{}{}
	if err := m.UnmarshalConfig(ctx, artifact.RepositoryName, content, &metadata); err != nil {
		return err
	}
	// if no properties specified, populate all metadata into the ExtraAttrs
	if len(m.properties) == 0 {
		artifact.ExtraAttrs = metadata
		return nil
	}

	if artifact.ExtraAttrs == nil {
		artifact.ExtraAttrs = map[string]interface{}{}
	}
	for _, property := range m.properties {
		artifact.ExtraAttrs[property] = metadata[property]
	}
	return nil
}

// AbstractAddition abstracts the addition of artifact
func (m *ManifestProcessor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*processor.Addition, error) {
	return nil, errors.New(nil).WithCode(errors.BadRequestCode).
		WithMessage("addition %s isn't supported", addition)
}

// GetArtifactType returns the artifact type
func (m *ManifestProcessor) GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string {
	return ""
}

// ListAdditionTypes returns the supported addition types
func (m *ManifestProcessor) ListAdditionTypes(ctx context.Context, artifact *artifact.Artifact) []string {
	return nil
}

// UnmarshalConfig unmarshal the config blob of the artifact into the specified object "v"
func (m *ManifestProcessor) UnmarshalConfig(ctx context.Context, repository string, manifest []byte, v interface{}) error {
	// unmarshal manifest
	mani := &v1.Manifest{}
	if err := json.Unmarshal(manifest, mani); err != nil {
		return err
	}
	// if the size of the config blob is 0(empty config blob), return directly
	if mani.Config.Size == 0 {
		return nil
	}
	// get config layer
	_, blob, err := m.RegCli.PullBlob(repository, mani.Config.Digest.String())
	if err != nil {
		return err
	}
	defer blob.Close()

	// unmarshal config layer
	return json.NewDecoder(blob).Decode(v)
}
