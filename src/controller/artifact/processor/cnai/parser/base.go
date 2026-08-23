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

package parser // nolint:revive

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

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

	// formatTar is the format of tar file.
	formatTar = ".tar"
	// formatRaw is the format of raw file.
	formatRaw = ".raw"
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

	// Reject early based on the manifest-declared size. This is only a cheap
	// pre-check: layer.Size is attacker-controlled and reflects the packed blob
	// size, which can be far smaller than the number of bytes materialized when
	// the content is decompressed/expanded (e.g. GNU tar sparse files).
	if layer.Size > defaultFileSizeLimit {
		return "", nil, errors.RequestEntityTooLargeError(errFileTooLarge)
	}

	_, stream, err := b.regCli.PullBlob(artifact.RepositoryName, layer.Digest.String())
	if err != nil {
		return "", nil, fmt.Errorf("failed to pull blob from registry: %w", err)
	}

	defer stream.Close()

	// Enforce the size limit against the actual bytes materialized, not just the
	// declared blob size, to prevent decompression/sparse-file bombs from
	// exhausting memory.
	content, err := decodeContent(layer.MediaType, stream, defaultFileSizeLimit)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decode content: %w", err)
	}

	return contentTypeTextPlain, content, nil
}

// decodeContent decodes the content read from reader according to mediaType,
// enforcing that no more than limit bytes are materialized in memory.
func decodeContent(mediaType string, reader io.Reader, limit int64) ([]byte, error) {
	format := filepath.Ext(mediaType)
	switch format {
	case formatTar:
		return untar(reader, limit)
	case formatRaw:
		return readAllLimited(reader, limit)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// readAllLimited reads from reader until EOF, but fails once more than limit
// bytes have been read.
func readAllLimited(reader io.Reader, limit int64) ([]byte, error) {
	if limit < 0 {
		return nil, fmt.Errorf("invalid limit: %d", limit)
	}

	// Read up to limit+1 bytes so we can distinguish "exactly at the limit"
	// from "over the limit".
	content, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}

	if int64(len(content)) > limit {
		return nil, errors.RequestEntityTooLargeError(errFileTooLarge)
	}

	return content, nil
}
