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

package processor

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

var (
	// Registry for registered artifact processors
	Registry = map[string]Processor{}
)

// Addition defines the specific addition of different artifacts: build history for image, values.yaml for chart, etc
type Addition struct {
	Content     []byte // the content of the addition
	ContentType string // the content type of the addition, returned as "Content-Type" header in API
}

// Processor processes specified artifact
type Processor interface {
	// GetArtifactType returns the type of one kind of artifact specified by media type
	GetArtifactType(ctx context.Context, artifact *artifact.Artifact) string
	// ListAdditionTypes returns the supported addition types of one kind of artifact specified by media type
	ListAdditionTypes(ctx context.Context, artifact *artifact.Artifact) []string
	// AbstractMetadata abstracts the metadata for the specific artifact type into the artifact model,
	// the metadata can be got from the manifest or other layers referenced by the manifest.
	AbstractMetadata(ctx context.Context, artifact *artifact.Artifact, manifest []byte) error
	// AbstractAddition abstracts the addition of the artifact.
	// The additions are different for different artifacts:
	// build history for image; values.yaml, readme and dependencies for chart, etc
	AbstractAddition(ctx context.Context, artifact *artifact.Artifact, additionType string) (addition *Addition, err error)
}

// Register artifact processor, one processor can process multiple media types for one kind of artifact
func Register(processor Processor, mediaTypes ...string) error {
	for _, mediaType := range mediaTypes {
		_, exist := Registry[mediaType]
		if exist {
			return fmt.Errorf("the processor to process media type %s already exists", mediaType)
		}
		Registry[mediaType] = processor
		log.Infof("the processor to process media type %s registered", mediaType)
	}
	return nil
}

// Get the artifact processor according to the media type
func Get(mediaType string) Processor {
	processor := Registry[mediaType]
	// no registered processor found, use the default one
	if processor == nil {
		log.Debugf("the processor for media type %s not found, use the default one", mediaType)
		processor = DefaultProcessor
	}
	return processor
}
