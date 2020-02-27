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

package descriptor

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"regexp"
	"strings"
)

// ArtifactTypeUnknown defines the type for the unknown artifacts
const ArtifactTypeUnknown = "UNKNOWN"

var (
	registry           = map[string]Descriptor{}
	artifactTypeRegExp = regexp.MustCompile(`^application/vnd\.[^.]*\.(.*)\.config\.[^.]*\+json$`)
)

// Descriptor describes the static information for one kind of media type
type Descriptor interface {
	// GetArtifactType returns the type of one kind of artifact specified by media type
	GetArtifactType() string
	// ListAdditionTypes returns the supported addition types of one kind of artifact specified by media type
	ListAdditionTypes() []string
}

// Register descriptor, one descriptor can handle multiple media types for one kind of artifact
func Register(descriptor Descriptor, mediaTypes ...string) error {
	for _, mediaType := range mediaTypes {
		_, exist := registry[mediaType]
		if exist {
			return fmt.Errorf("descriptor to handle media type %s already exists", mediaType)
		}
		registry[mediaType] = descriptor
		log.Infof("descriptor to handle media type %s registered", mediaType)
	}
	return nil
}

// Get the descriptor according to the media type
func Get(mediaType string) Descriptor {
	return registry[mediaType]
}

// GetArtifactType gets the artifact type according to the media type
func GetArtifactType(mediaType string) string {
	descriptor := Get(mediaType)
	if descriptor != nil {
		return descriptor.GetArtifactType()
	}
	// if got no descriptor, try to parse the artifact type based on the media type
	return parseArtifactType(mediaType)
}

// ListAdditionTypes lists the supported addition types according to the media type
func ListAdditionTypes(mediaType string) []string {
	descriptor := Get(mediaType)
	if descriptor != nil {
		return descriptor.ListAdditionTypes()
	}
	return nil
}

func parseArtifactType(mediaType string) string {
	strs := artifactTypeRegExp.FindStringSubmatch(mediaType)
	if len(strs) == 2 {
		return strings.ToUpper(strs[1])
	}
	// can not get the artifact type from the media type, return unknown
	return ArtifactTypeUnknown
}
