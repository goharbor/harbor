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

package oci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	modelspec "github.com/modelpack/model-spec/specs-go/v1"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/pkg/modelimport/huggingface"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// FileOpener opens a source file stream.
type FileOpener func(ctx context.Context, filename string) (io.ReadCloser, error)

// Metadata is source metadata written into the model-spec config extension.
type Metadata struct {
	RepoID    string
	Revision  string
	CommitSHA string
	SourceURL string
	CardData  map[string]any
}

// Result contains the pushed OCI model artifact information.
type Result struct {
	Digest string
}

// Converter converts a Hugging Face snapshot into an OCI model-spec artifact.
type Converter struct {
	regCli registry.Client
}

// New returns a converter.
func New(regCli registry.Client) *Converter {
	return &Converter{regCli: regCli}
}

// Push converts and pushes the source files into the target repository.
func (c *Converter) Push(ctx context.Context, repository, tag string, snapshot *huggingface.Snapshot, opener FileOpener) (*Result, error) {
	if c.regCli == nil {
		c.regCli = registry.Cli
	}
	var layers []ocispec.Descriptor
	var diffIDs []digest.Digest
	for _, file := range snapshot.Files {
		desc, diffID, err := c.pushFile(ctx, repository, file, opener)
		if err != nil {
			return nil, err
		}
		layers = append(layers, *desc)
		diffIDs = append(diffIDs, diffID)
	}

	configBytes, err := json.Marshal(modelConfig(snapshot, diffIDs))
	if err != nil {
		return nil, err
	}
	configDigest := digest.FromBytes(configBytes)
	if err := c.regCli.PushBlob(repository, configDigest.String(), int64(len(configBytes)), bytes.NewReader(configBytes)); err != nil {
		return nil, fmt.Errorf("push model config: %w", err)
	}

	manifest := ocispec.Manifest{
		Versioned:    specs.Versioned{SchemaVersion: 2},
		MediaType:    ocispec.MediaTypeImageManifest,
		ArtifactType: modelspec.ArtifactTypeModelManifest,
		Config: ocispec.Descriptor{
			MediaType: modelspec.MediaTypeModelConfig,
			Digest:    configDigest,
			Size:      int64(len(configBytes)),
		},
		Layers: layers,
		Annotations: map[string]string{
			"org.opencontainers.image.source":   snapshot.SourceURL,
			"org.opencontainers.image.revision": snapshot.CommitSHA,
		},
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return nil, err
	}
	dgst, err := c.regCli.PushManifest(repository, tag, ocispec.MediaTypeImageManifest, payload)
	if err != nil {
		return nil, fmt.Errorf("push model manifest: %w", err)
	}
	return &Result{Digest: dgst}, nil
}

func (c *Converter) pushFile(ctx context.Context, repository string, file huggingface.File, opener FileOpener) (*ocispec.Descriptor, digest.Digest, error) {
	const maxAttempts = 4
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		desc, dgst, err := c.pushFileOnce(ctx, repository, file, opener)
		if err == nil {
			return desc, dgst, nil
		}
		lastErr = err
		if attempt == maxAttempts || ctx.Err() != nil || !isTransientDownloadErr(err) {
			break
		}
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		case <-time.After(time.Duration(attempt) * time.Second):
		}
	}
	return nil, "", lastErr
}

func (c *Converter) pushFileOnce(ctx context.Context, repository string, file huggingface.File, opener FileOpener) (*ocispec.Descriptor, digest.Digest, error) {
	reader, err := opener(ctx, file.Path)
	if err != nil {
		return nil, "", err
	}
	defer reader.Close()

	tmp, err := os.CreateTemp("", "harbor-model-import-*")
	if err != nil {
		return nil, "", err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	digester := digest.Canonical.Digester()
	written, err := io.Copy(io.MultiWriter(tmp, digester.Hash()), reader)
	if err != nil {
		return nil, "", err
	}
	if _, err = tmp.Seek(0, io.SeekStart); err != nil {
		return nil, "", err
	}
	dgst := digester.Digest()
	if err = c.regCli.PushBlob(repository, dgst.String(), written, tmp); err != nil {
		return nil, "", fmt.Errorf("push blob %s: %w", file.Path, err)
	}
	mediaType := mediaTypeFor(file.Path)
	return &ocispec.Descriptor{
		MediaType: mediaType,
		Digest:    dgst,
		Size:      written,
		Annotations: map[string]string{
			modelspec.AnnotationFilepath: file.Path,
		},
	}, dgst, nil
}

func isTransientDownloadErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	for _, needle := range []string{
		"unexpected eof",
		"connection reset",
		"broken pipe",
		"i/o timeout",
		"tls handshake timeout",
		"http2: server sent goaway",
		"too many requests",
		"failed: 429",
		"failed: 500",
		"failed: 502",
		"failed: 503",
		"failed: 504",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func modelConfig(snapshot *huggingface.Snapshot, diffIDs []digest.Digest) map[string]any {
	descriptor := modelspec.ModelDescriptor{
		Name:        repoName(snapshot.RepoID),
		Title:       stringFromCard(snapshot.CardData, "title"),
		Description: stringFromCard(snapshot.CardData, "description"),
		SourceURL:   snapshot.SourceURL,
		Revision:    snapshot.CommitSHA,
		Vendor:      "Hugging Face",
		Licenses:    stringSliceFromCard(snapshot.CardData, "license"),
	}
	if descriptor.Title == "" {
		descriptor.Title = descriptor.Name
	}
	config := modelspec.ModelConfig{
		Architecture: stringFromCard(snapshot.CardData, "architecture"),
		Format:       stringFromCard(snapshot.CardData, "library_name"),
	}
	return map[string]any{
		"descriptor": descriptor,
		"config":     config,
		"modelfs": modelspec.ModelFS{
			Type:    "layers",
			DiffIDs: diffIDs,
		},
		"huggingface": map[string]any{
			"repo_id":    snapshot.RepoID,
			"revision":   snapshot.Revision,
			"commit_sha": snapshot.CommitSHA,
			"source_url": snapshot.SourceURL,
			"card_data":  snapshot.CardData,
		},
	}
}

func mediaTypeFor(name string) string {
	lower := strings.ToLower(name)
	switch {
	case lower == "readme.md" || strings.HasPrefix(lower, "readme.") ||
		lower == "license" || strings.HasPrefix(lower, "license."):
		return modelspec.MediaTypeModelDocRaw
	case strings.HasSuffix(lower, ".py") || strings.HasSuffix(lower, ".sh"):
		return modelspec.MediaTypeModelCodeRaw
	case strings.HasSuffix(lower, ".json") || strings.HasSuffix(lower, ".txt") ||
		strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") ||
		strings.Contains(lower, "tokenizer") || strings.Contains(lower, "config"):
		return modelspec.MediaTypeModelWeightConfigRaw
	default:
		return modelspec.MediaTypeModelWeightRaw
	}
}

func repoName(repoID string) string {
	base := filepath.Base(repoID)
	if base == "." || base == "/" {
		return repoID
	}
	return base
}

func stringFromCard(card map[string]any, key string) string {
	if card == nil {
		return ""
	}
	v, ok := card[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func stringSliceFromCard(card map[string]any, key string) []string {
	if card == nil {
		return nil
	}
	v, ok := card[key]
	if !ok || v == nil {
		return nil
	}
	switch value := v.(type) {
	case string:
		if value == "" {
			return nil
		}
		return []string{value}
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if item != nil {
				out = append(out, fmt.Sprintf("%v", item))
			}
		}
		return out
	default:
		return []string{fmt.Sprintf("%v", value)}
	}
}
