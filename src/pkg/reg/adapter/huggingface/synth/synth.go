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

// Package synth turns a Hub snapshot into a byte-deterministic ModelPack model-spec
// OCI artifact. It performs no I/O.
package synth

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
)

// Version identifies the synthesis rules. Any change to the bytes produced for the same
// snapshot must bump it.
const Version = "1"

// Media types and annotations of the ModelPack model-spec. They are declared here rather than
// imported from model-spec because every value is part of the manifest digest.
const (
	ArtifactTypeModelManifest = "application/vnd.cncf.model.manifest.v1+json"
	MediaTypeModelConfig      = "application/vnd.cncf.model.config.v1+json"
	MediaTypeWeightRaw        = "application/vnd.cncf.model.weight.v1.raw"
	MediaTypeWeightConfigRaw  = "application/vnd.cncf.model.weight.config.v1.raw"
	MediaTypeDocRaw           = "application/vnd.cncf.model.doc.v1.raw"
	MediaTypeCodeRaw          = "application/vnd.cncf.model.code.v1.raw"
	AnnotationFilepath        = "org.cncf.model.filepath"

	AnnotationSynthesisVersion = "org.goharbor.huggingface.synthesis.version"
)

// Artifact is a synthesized model artifact.
type Artifact struct {
	Manifest     []byte
	Digest       digest.Digest
	Config       []byte
	ConfigDigest digest.Digest
	Layers       []Layer
}

// Layer is one raw file of the artifact.
type Layer struct {
	Path      string
	Digest    digest.Digest
	Size      int64
	MediaType string
}

// Included reports whether a repository path becomes a layer.
func Included(filePath string) bool {
	return filePath != ".gitattributes"
}

// Build synthesizes the artifact of a snapshot. fileDigests must hold the content digest of every
// included non-LFS file; LFS files use the sha256 published by the Hub.
func Build(s *hub.Snapshot, fileDigests map[string]digest.Digest) (*Artifact, error) {
	if s == nil {
		return nil, fmt.Errorf("nil snapshot")
	}
	if !hub.IsCommit(s.Commit) {
		return nil, fmt.Errorf("invalid commit %q", s.Commit)
	}
	var layers []Layer
	for _, f := range s.Files {
		if !Included(f.Path) {
			continue
		}
		var d digest.Digest
		if f.IsLFS() {
			d = digest.NewDigestFromEncoded(digest.SHA256, f.SHA256)
		} else {
			d = fileDigests[f.Path]
		}
		if err := d.Validate(); err != nil || d.Algorithm() != digest.SHA256 {
			return nil, fmt.Errorf("no valid sha256 digest for %s of model %s at %s", f.Path, s.ModelID, s.Commit)
		}
		layers = append(layers, Layer{
			Path:      f.Path,
			Digest:    d,
			Size:      f.Size,
			MediaType: classify(f.Path, f.Size).mediaType(),
		})
	}
	if len(layers) == 0 {
		return nil, fmt.Errorf("model %s at %s has no files", s.ModelID, s.Commit)
	}
	sort.Slice(layers, func(i, j int) bool { return layers[i].Path < layers[j].Path })
	for i := 1; i < len(layers); i++ {
		if layers[i].Path == layers[i-1].Path {
			return nil, fmt.Errorf("model %s at %s lists %s twice", s.ModelID, s.Commit, layers[i].Path)
		}
	}

	cfg := modelConfig{
		Descriptor: modelDescriptor{
			Name: s.ModelID,
			// Pinned to the public Hub rather than the configured endpoint, so the digest is the
			// same on every Harbor and behind any mirror.
			SourceURL: hub.DefaultEndpoint + "/" + s.ModelID,
			Revision:  s.Commit,
			Licenses:  s.Licenses,
		},
		ModelFS: modelFS{Type: "layers"},
		Config:  modelConfigSection{Format: format(layers)},
	}
	if !s.LastModified.IsZero() {
		cfg.Descriptor.CreatedAt = s.LastModified.UTC().Format(time.RFC3339)
	}
	for _, l := range layers {
		cfg.ModelFS.DiffIDs = append(cfg.ModelFS.DiffIDs, l.Digest)
	}
	configBytes, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	configDigest := digest.FromBytes(configBytes)

	m := manifest{
		SchemaVersion: 2,
		MediaType:     v1.MediaTypeImageManifest,
		ArtifactType:  ArtifactTypeModelManifest,
		Config: descriptor{
			MediaType: MediaTypeModelConfig,
			Digest:    configDigest,
			Size:      int64(len(configBytes)),
		},
		Layers:      make([]descriptor, 0, len(layers)),
		Annotations: map[string]string{AnnotationSynthesisVersion: Version},
	}
	for _, l := range layers {
		m.Layers = append(m.Layers, descriptor{
			MediaType:   l.MediaType,
			Digest:      l.Digest,
			Size:        l.Size,
			Annotations: map[string]string{AnnotationFilepath: l.Path},
		})
	}
	manifestBytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &Artifact{
		Manifest:     manifestBytes,
		Digest:       digest.FromBytes(manifestBytes),
		Config:       configBytes,
		ConfigDigest: configDigest,
		Layers:       layers,
	}, nil
}

// format names the weight format when all weight layers share one.
func format(layers []Layer) string {
	result := ""
	for _, l := range layers {
		if l.MediaType != MediaTypeWeightRaw {
			continue
		}
		f := weightFormat(l.Path)
		if f == "" || (result != "" && f != result) {
			return ""
		}
		result = f
	}
	return result
}

func weightFormat(filePath string) string {
	base := strings.ToLower(path.Base(filePath))
	switch {
	case strings.HasSuffix(base, ".safetensors"):
		return "safetensors"
	case strings.HasSuffix(base, ".gguf"):
		return "gguf"
	case strings.HasSuffix(base, ".onnx"), strings.Contains(base, ".onnx_data"):
		return "onnx"
	default:
		return ""
	}
}

type manifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType"`
	ArtifactType  string            `json:"artifactType"`
	Config        descriptor        `json:"config"`
	Layers        []descriptor      `json:"layers"`
	Annotations   map[string]string `json:"annotations"`
}

type descriptor struct {
	MediaType   string            `json:"mediaType"`
	Digest      digest.Digest     `json:"digest"`
	Size        int64             `json:"size"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// The config structs mirror model-spec specs-go/v1 field names and order but are local, so a
// model-spec dependency bump cannot add a field and change every digest.
type modelConfig struct {
	Descriptor modelDescriptor    `json:"descriptor"`
	ModelFS    modelFS            `json:"modelfs"`
	Config     modelConfigSection `json:"config"`
}

type modelDescriptor struct {
	CreatedAt string   `json:"createdAt,omitempty"`
	Name      string   `json:"name,omitempty"`
	SourceURL string   `json:"sourceURL,omitempty"`
	Revision  string   `json:"revision,omitempty"`
	Licenses  []string `json:"licenses,omitempty"`
}

type modelFS struct {
	Type    string          `json:"type"`
	DiffIDs []digest.Digest `json:"diffIds"`
}

type modelConfigSection struct {
	Format string `json:"format,omitempty"`
}
