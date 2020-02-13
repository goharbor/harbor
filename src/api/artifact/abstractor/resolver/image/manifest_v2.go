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
	"github.com/goharbor/harbor/src/api/artifact/abstractor/blob"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/api/artifact/descriptor"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

// const definitions
const (
	// ArtifactTypeImage is the artifact type for image
	ArtifactTypeImage           = "IMAGE"
	AdditionTypeBuildHistory    = "BUILD_HISTORY"
	AdditionTypeVulnerabilities = "VULNERABILITIES"
)

func init() {
	rslver := &manifestV2Resolver{
		repoMgr:     repository.Mgr,
		blobFetcher: blob.Fcher,
	}
	mediaTypes := []string{
		v1.MediaTypeImageConfig,
		schema2.MediaTypeImageConfig,
	}
	if err := resolver.Register(rslver, mediaTypes...); err != nil {
		log.Errorf("failed to register resolver for media type %v: %v", mediaTypes, err)
		return
	}
	if err := descriptor.Register(rslver, mediaTypes...); err != nil {
		log.Errorf("failed to register descriptor for media type %v: %v", mediaTypes, err)
		return
	}
}

// manifestV2Resolver resolve artifact with OCI manifest and docker v2 manifest
type manifestV2Resolver struct {
	repoMgr     repository.Manager
	blobFetcher blob.Fetcher
}

func (m *manifestV2Resolver) ResolveMetadata(ctx context.Context, content []byte, artifact *artifact.Artifact) error {
	repository, err := m.repoMgr.Get(ctx, artifact.RepositoryID)
	if err != nil {
		return err
	}
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(content, manifest); err != nil {
		return err
	}
	digest := manifest.Config.Digest.String()
	layer, err := m.blobFetcher.FetchLayer(repository.Name, digest)
	if err != nil {
		return err
	}
	image := &v1.Image{}
	if err := json.Unmarshal(layer, image); err != nil {
		return err
	}
	if artifact.ExtraAttrs == nil {
		artifact.ExtraAttrs = map[string]interface{}{}
	}
	artifact.ExtraAttrs["created"] = image.Created
	artifact.ExtraAttrs["author"] = image.Author
	artifact.ExtraAttrs["architecture"] = image.Architecture
	artifact.ExtraAttrs["os"] = image.OS
	return nil
}

func (m *manifestV2Resolver) ResolveAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*resolver.Addition, error) {
	if addition != AdditionTypeBuildHistory {
		return nil, ierror.New(nil).WithCode(ierror.BadRequestCode).
			WithMessage("addition %s isn't supported for %s(manifest version 2)", addition, ArtifactTypeImage)
	}
	repository, err := m.repoMgr.Get(ctx, artifact.RepositoryID)
	if err != nil {
		return nil, err
	}
	_, content, err := m.blobFetcher.FetchManifest(repository.Name, artifact.Digest)
	if err != nil {
		return nil, err
	}
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(content, manifest); err != nil {
		return nil, err
	}
	content, err = m.blobFetcher.FetchLayer(repository.Name, manifest.Config.Digest.String())
	if err != nil {
		return nil, err
	}
	image := &v1.Image{}
	if err := json.Unmarshal(content, image); err != nil {
		return nil, err
	}
	content, err = json.Marshal(image.History)
	if err != nil {
		return nil, err
	}
	return &resolver.Addition{
		Content:     content,
		ContentType: "application/json; charset=utf-8",
	}, nil
}

func (m *manifestV2Resolver) GetArtifactType() string {
	return ArtifactTypeImage
}

func (m *manifestV2Resolver) ListAdditionTypes() []string {
	return []string{AdditionTypeBuildHistory, AdditionTypeVulnerabilities}
}
