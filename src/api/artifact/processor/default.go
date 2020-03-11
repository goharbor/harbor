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
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"regexp"
	"strings"
)

// ArtifactTypeUnknown defines the type for the unknown artifacts
const ArtifactTypeUnknown = "UNKNOWN"

var (
	artifactTypeRegExp = regexp.MustCompile(`^application/vnd\.[^.]*\.(.*)\.config\.[^.]*\+json$`)
)

// the default processor to process artifact
// currently, it only tries to parse the artifact type from media type
type defaultProcessor struct {
	mediaType string
}

func (d *defaultProcessor) GetArtifactType() string {
	// try to parse the type from the media type
	strs := artifactTypeRegExp.FindStringSubmatch(d.mediaType)
	if len(strs) == 2 {
		return strings.ToUpper(strs[1])
	}
	// can not get the artifact type from the media type, return unknown
	return ArtifactTypeUnknown
}
func (d *defaultProcessor) ListAdditionTypes() []string {
	return nil
}
func (d *defaultProcessor) AbstractMetadata(ctx context.Context, manifest []byte, artifact *artifact.Artifact) error {
	// do nothing currently
	// we can extend this function to abstract the metadata in the future if needed
	return nil
}
func (d *defaultProcessor) AbstractAddition(ctx context.Context, artifact *artifact.Artifact, addition string) (*Addition, error) {
	// return error directly
	return nil, ierror.New(nil).WithCode(ierror.BadRequestCode).
		WithMessage("the processor for artifact %s not found, cannot get the addition", artifact.Type)
}
