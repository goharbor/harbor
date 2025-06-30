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

// NewLicense creates a new license parser.
func NewLicense(cli registry.Client) Parser {
	return &license{
		base: newBase(cli),
	}
}

// license is the parser for License file.
type license struct {
	*base
}

// Parse parses the License file.
func (l *license) Parse(ctx context.Context, artifact *artifact.Artifact, manifest *ocispec.Manifest) (string, []byte, error) {
	if manifest == nil {
		return "", nil, errors.New("manifest cannot be nil")
	}

	// lookup the license file layer
	var layer *ocispec.Descriptor
	for _, desc := range manifest.Layers {
		if slices.Contains([]string{
			modelspec.MediaTypeModelDoc,
			modelspec.MediaTypeModelDocRaw,
		}, desc.MediaType) {
			if desc.Annotations != nil {
				filepath := desc.Annotations[modelspec.AnnotationFilepath]
				if filepath == "LICENSE" || filepath == "LICENSE.txt" {
					layer = &desc
					break
				}
			}
		}
	}

	if layer == nil {
		return "", nil, errors.NotFoundError(fmt.Errorf("license layer not found"))
	}

	return l.base.Parse(ctx, artifact, layer)
}
