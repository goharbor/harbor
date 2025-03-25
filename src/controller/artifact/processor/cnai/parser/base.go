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

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/registry"
)

var (
	// errFileTooLarge is returned when the file is too large to be processed.
	errFileTooLarge = errors.New("The file is too large to be processed")
)

const (
	// contentTypeTextPlain is the content type of text/plain.
	contentTypeTextPlain = "text/plain; charset=utf-8"
	// contentTypeTextMarkdown is the content type of text/markdown.
	contentTypeMarkdown = "text/markdown; charset=utf-8"
	// contentTypeJSON is the content type of application/json.
	contentTypeJSON = "application/json; charset=utf-8"

	// defaultFileSizeLimit is the default file size limit.
	defaultFileSizeLimit = 1024 * 1024 * 4 // 4MB
)

// newBase creates a new base parser.
func newBase(cli registry.Client) *base {
	return &base{
		regCli: cli,
	}
}

// base provides a default implementation for other parsers to build upon.
type base struct {
	regCli registry.Client
}

// Parse is the common implementation for parsing layer.
func (b *base) Parse(_ context.Context, artifact *artifact.Artifact, layer *ocispec.Descriptor) (string, []byte, error) {
	if artifact == nil || layer == nil {
		return "", nil, fmt.Errorf("artifact or manifest cannot be nil")
	}

	if layer.Size > defaultFileSizeLimit {
		return "", nil, errors.RequestEntityTooLargeError(errFileTooLarge)
	}

	_, stream, err := b.regCli.PullBlob(artifact.RepositoryName, layer.Digest.String())
	if err != nil {
		return "", nil, fmt.Errorf("failed to pull blob from registry: %w", err)
	}

	defer stream.Close()
	content, err := untar(stream)
	if err != nil {
		return "", nil, fmt.Errorf("failed to untar the content: %w", err)
	}

	return contentTypeTextPlain, content, nil
}
