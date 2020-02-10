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
	"github.com/docker/distribution/manifest/schema1"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/api/artifact/descriptor"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

func init() {
	rslver := &manifestV1Resolver{}
	if err := resolver.Register(rslver, schema1.MediaTypeSignedManifest); err != nil {
		log.Errorf("failed to register resolver for media type %s: %v", schema1.MediaTypeSignedManifest, err)
		return
	}
	if err := descriptor.Register(rslver, schema1.MediaTypeSignedManifest); err != nil {
		log.Errorf("failed to register descriptor for media type %s: %v", schema1.MediaTypeSignedManifest, err)
		return
	}
}

// manifestV1Resolver resolve artifact with docker v1 manifest
type manifestV1Resolver struct {
}

func (m *manifestV1Resolver) ResolveMetadata(ctx context.Context, manifest []byte, artifact *artifact.Artifact) error {
	mani := &schema1.Manifest{}
	if err := json.Unmarshal([]byte(manifest), mani); err != nil {
		return err
	}
	if artifact.ExtraAttrs == nil {
		artifact.ExtraAttrs = map[string]interface{}{}
	}
	artifact.ExtraAttrs["architecture"] = mani.Architecture
	return nil
}

func (m *manifestV1Resolver) ResolveAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*resolver.Addition, error) {
	return nil, ierror.New(nil).WithCode(ierror.BadRequestCode).
		WithMessage("addition %s isn't supported for %s(manifest version 1)", addition, ArtifactTypeImage)
}

func (m *manifestV1Resolver) GetArtifactType() string {
	return ArtifactTypeImage
}

func (m *manifestV1Resolver) ListAdditionTypes() []string {
	return nil
}
