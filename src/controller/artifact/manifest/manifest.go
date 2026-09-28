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

package manifest

import (
	"context"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// Registry holds the registered manifest abstractors, keyed by manifest media type.
var Registry = map[string]Abstractor{}

// Abstractor abstracts the metadata carried by one manifest format into the
// artifact model.
type Abstractor interface {
	Abstract(ctx context.Context, art *artifact.Artifact, content []byte) error
}

// Register an abstractor for one or more manifest media types. Registration
// failures are fatal at startup: an unregistered media type would fail every
// push and pull of that manifest format at runtime, far from the actual cause.
// The batch is validated before anything is written, since a half-applied batch
// would make dispatch depend on the order of the media types.
func Register(abstractor Abstractor, mediaTypes ...string) error {
	seen := make(map[string]struct{}, len(mediaTypes))
	for _, mediaType := range mediaTypes {
		if _, exist := Registry[mediaType]; exist {
			return errors.Errorf("the abstractor for manifest media type %s already exists", mediaType)
		}
		if _, duplicate := seen[mediaType]; duplicate {
			return errors.Errorf("the manifest media type %s is listed twice", mediaType)
		}
		seen[mediaType] = struct{}{}
	}

	for _, mediaType := range mediaTypes {
		Registry[mediaType] = abstractor
		log.Debugf("the abstractor for manifest media type %s registered", mediaType)
	}
	return nil
}

// Get the abstractor for the manifest media type. Unlike processor.Get there is
// no default fallback: guessing at an unknown format would corrupt the artifact model.
func Get(mediaType string) (Abstractor, error) {
	abstractor, exist := Registry[mediaType]
	if !exist {
		return nil, errors.Errorf("unsupported manifest media type: %s", mediaType).WithCode(errors.UNSUPPORTED)
	}
	return abstractor, nil
}
