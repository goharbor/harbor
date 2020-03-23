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

package image

import (
	"context"
	"encoding/json"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	ierror "github.com/goharbor/harbor/src/lib/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

// const definitions
const (
	// ArtifactTypeImage is the artifact type for image
	ArtifactTypeImage        = "IMAGE"
	AdditionTypeBuildHistory = "BUILD_HISTORY"
)

func init() {
	pc := &manifestV2Processor{}
	pc.ManifestProcessor = base.NewManifestProcessor("created", "author", "architecture", "os")
	mediaTypes := []string{
		v1.MediaTypeImageConfig,
		schema2.MediaTypeImageConfig,
	}
	if err := processor.Register(pc, mediaTypes...); err != nil {
		log.Errorf("failed to register processor for media type %v: %v", mediaTypes, err)
		return
	}
}

// manifestV2Processor processes image with OCI manifest and docker v2 manifest
type manifestV2Processor struct {
	*base.ManifestProcessor
}

func (m *manifestV2Processor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*processor.Addition, error) {
	if addition != AdditionTypeBuildHistory {
		return nil, ierror.New(nil).WithCode(ierror.BadRequestCode).
			WithMessage("addition %s isn't supported for %s(manifest version 2)", addition, ArtifactTypeImage)
	}
	mani, _, err := m.RegCli.PullManifest(artifact.RepositoryName, artifact.Digest)
	if err != nil {
		return nil, err
	}
	_, content, err := mani.Payload()
	if err != nil {
		return nil, err
	}
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(content, manifest); err != nil {
		return nil, err
	}
	_, blob, err := m.RegCli.PullBlob(artifact.RepositoryName, manifest.Config.Digest.String())
	if err != nil {
		return nil, err
	}
	image := &v1.Image{}
	if err := json.NewDecoder(blob).Decode(image); err != nil {
		return nil, err
	}
	content, err = json.Marshal(image.History)
	if err != nil {
		return nil, err
	}
	return &processor.Addition{
		Content:     content,
		ContentType: "application/json; charset=utf-8",
	}, nil
}

func (m *manifestV2Processor) GetArtifactType() string {
	return ArtifactTypeImage
}

func (m *manifestV2Processor) ListAdditionTypes() []string {
	return []string{AdditionTypeBuildHistory}
}
