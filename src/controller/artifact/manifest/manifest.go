// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manifest

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

var (
	ManifestAbstractorRegistry = map[string]ManifestAbstractor{}
)

// ManifestAbstractor abstracts the metadata of artifact
type ManifestAbstractor interface {
	// AbstractManifest abstracts the metadata for the specific artifact type into the artifact model,
	AbstractManifestMetadata(ctx context.Context, artifact *artifact.Artifact, content []byte) error
}

// Register registers the manifest abstractor for specific media types
func Register(abstractor ManifestAbstractor, mediaTypes ...string) error {
	for _, mediaType := range mediaTypes {
		if _, exist := ManifestAbstractorRegistry[mediaType]; exist {
			err := errors.New(fmt.Sprintf("the manifest abstractor to process media type %s already exists", mediaType))
			return err
		}
		ManifestAbstractorRegistry[mediaType] = abstractor
		log.Infof("the manifest abstractor to process media type %s registered", mediaType)
	}
	return nil
}

func Get(mediaType string) (ManifestAbstractor, error) {
	abstractor, exist := ManifestAbstractorRegistry[mediaType]
	if !exist {
		return nil, fmt.Errorf("no manifest abstractor found to process media type %s", mediaType)
	}
	return abstractor, nil
}
