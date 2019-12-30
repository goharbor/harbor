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

package resolver

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

var (
	registry = map[string]Resolver{}
)

// Resolver resolves the detail information for a specific kind of artifact
type Resolver interface {
	// ArtifactType returns the type of artifact that the resolver handles
	ArtifactType() string
	// Resolve receives the manifest content, resolves the detail information
	// from the manifest or the layers referenced by the manifest, and populates
	// the detail information into the artifact
	Resolve(ctx context.Context, manifest []byte, artifact *artifact.Artifact) error
}

// Register resolver, one resolver can handle multiple media types for one kind of artifact
func Register(resolver Resolver, mediaTypes ...string) error {
	for _, mediaType := range mediaTypes {
		_, exist := registry[mediaType]
		if exist {
			return fmt.Errorf("resolver to handle media type %s already exists", mediaType)
		}
		registry[mediaType] = resolver
		log.Infof("resolver to handle media type %s registered", mediaType)
	}
	return nil
}

// Get the resolver according to the media type
func Get(mediaType string) (Resolver, error) {
	resolver, exist := registry[mediaType]
	if !exist {
		return nil, fmt.Errorf("resolver resolves %s not found", mediaType)
	}
	return resolver, nil
}
