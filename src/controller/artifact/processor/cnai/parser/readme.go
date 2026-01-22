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

package parser

import (
	"context"
	"fmt"
	"slices"

	modelspec "github.com/CloudNativeAI/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// NewReadme creates a new readme parser.
func NewReadme(cli registry.Client) Parser {
	return &readme{
		base: newBase(cli),
	}
}

// readme is the parser for README.md file.
type readme struct {
	*base
}

// Parse parses the README.md file.
func (r *readme) Parse(ctx context.Context, artifact *artifact.Artifact, manifest *ocispec.Manifest) (string, []byte, error) {
	if manifest == nil {
		return "", nil, errors.New("manifest cannot be nil")
	}

	// lookup the readme file layer.
	var layer *ocispec.Descriptor
	for _, desc := range manifest.Layers {
		if slices.Contains([]string{
			modelspec.MediaTypeModelDoc,
			modelspec.MediaTypeModelDocRaw,
		}, desc.MediaType) {
			if desc.Annotations != nil {
				filepath := desc.Annotations[modelspec.AnnotationFilepath]
				if filepath == "README" || filepath == "README.md" {
					layer = &desc
					break
				}
			}
		}
	}

	if layer == nil {
		return "", nil, errors.NotFoundError(fmt.Errorf("readme layer not found"))
	}

	_, content, err := r.base.Parse(ctx, artifact, layer)
	if err != nil {
		return "", nil, err
	}

	return contentTypeMarkdown, content, nil
}
