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

package synth

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/docker/distribution"
	_ "github.com/docker/distribution/manifest/ocischema"
	modelspec "github.com/modelpack/model-spec/specs-go/v1"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
)

const (
	qwenCommit = "c1899de289a04d12100db370d81485cdf75e47ca"
	// Independently reproduced with a Python json.dumps of the same structure. Changing it requires bumping Version.
	qwenManifestDigest = "sha256:abe67638e7a34128814f19351ede02006096d2c168adedd8e04ff747953abb56"
)

func loadQwen(t *testing.T) (*hub.Snapshot, map[string]digest.Digest) {
	t.Helper()
	data, err := os.ReadFile("../testdata/qwen3-0.6b-revision.json")
	require.NoError(t, err)
	s, err := hub.DecodeSnapshot(data)
	require.NoError(t, err)

	raw, err := os.ReadFile("../testdata/qwen3-0.6b-digests.json")
	require.NoError(t, err)
	digests := map[string]digest.Digest{}
	require.NoError(t, json.Unmarshal(raw, &digests))
	return s, digests
}

func TestBuildGoldenDigest(t *testing.T) {
	s, digests := loadQwen(t)
	art, err := Build(s, digests)
	require.NoError(t, err)
	assert.Equal(t, qwenManifestDigest, art.Digest.String())
	assert.Equal(t, digest.FromBytes(art.Manifest), art.Digest)
	assert.Equal(t, digest.FromBytes(art.Config), art.ConfigDigest)

	again, err := Build(s, digests)
	require.NoError(t, err)
	assert.Equal(t, art.Manifest, again.Manifest)
}

func TestBuildQwenContent(t *testing.T) {
	s, digests := loadQwen(t)
	art, err := Build(s, digests)
	require.NoError(t, err)

	var m v1.Manifest
	require.NoError(t, json.Unmarshal(art.Manifest, &m))
	assert.Equal(t, 2, m.SchemaVersion)
	assert.Equal(t, v1.MediaTypeImageManifest, m.MediaType)
	assert.Equal(t, ArtifactTypeModelManifest, m.ArtifactType)
	assert.Equal(t, map[string]string{AnnotationSynthesisVersion: Version}, m.Annotations)
	assert.Equal(t, MediaTypeModelConfig, m.Config.MediaType)
	assert.Equal(t, art.ConfigDigest, m.Config.Digest)
	assert.Equal(t, int64(len(art.Config)), m.Config.Size)

	type layer struct{ path, mediaType string }
	var got []layer
	for _, l := range m.Layers {
		got = append(got, layer{l.Annotations[AnnotationFilepath], l.MediaType})
		assert.Len(t, l.Annotations, 1)
	}
	assert.Equal(t, []layer{
		{"LICENSE", MediaTypeDocRaw},
		{"README.md", MediaTypeDocRaw},
		{"config.json", MediaTypeWeightConfigRaw},
		{"generation_config.json", MediaTypeWeightConfigRaw},
		{"merges.txt", MediaTypeWeightConfigRaw},
		{"model.safetensors", MediaTypeWeightRaw},
		{"tokenizer.json", MediaTypeWeightConfigRaw},
		{"tokenizer_config.json", MediaTypeWeightConfigRaw},
		{"vocab.json", MediaTypeWeightConfigRaw},
	}, got)
	assert.Equal(t, "sha256:f47f71177f32bcd101b7573ec9171e6a57f4f4d31148d38e382306f42996874b", m.Layers[5].Digest.String())
	assert.Equal(t, int64(1503300328), m.Layers[5].Size)
	assert.Equal(t, digests["config.json"], m.Layers[2].Digest)

	var cfg map[string]any
	require.NoError(t, json.Unmarshal(art.Config, &cfg))
	assert.Equal(t, map[string]any{
		"createdAt": "2025-07-26T03:46:27Z",
		"name":      "Qwen/Qwen3-0.6B",
		"sourceURL": "https://huggingface.co/Qwen/Qwen3-0.6B",
		"revision":  qwenCommit,
		"licenses":  []any{"apache-2.0"},
	}, cfg["descriptor"])
	assert.Equal(t, map[string]any{"format": "safetensors"}, cfg["config"])
	fs := cfg["modelfs"].(map[string]any)
	assert.Equal(t, "layers", fs["type"])
	require.Len(t, fs["diffIds"], len(m.Layers))
	for i, l := range m.Layers {
		assert.Equal(t, l.Digest.String(), fs["diffIds"].([]any)[i])
	}

	// distribution must keep the exact bytes, including artifactType.
	mf, desc, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, art.Manifest)
	require.NoError(t, err)
	assert.Equal(t, art.Digest, desc.Digest)
	_, payload, err := mf.Payload()
	require.NoError(t, err)
	assert.Equal(t, art.Manifest, payload)
	assert.Len(t, mf.References(), len(m.Layers)+1)
}

// Harbor's cnai processor and parsers key on model-spec's constants; a typo here would skip them
// while the golden digest stays green.
func TestMediaTypesMatchModelSpec(t *testing.T) {
	assert.Equal(t, modelspec.ArtifactTypeModelManifest, ArtifactTypeModelManifest)
	assert.Equal(t, modelspec.MediaTypeModelConfig, MediaTypeModelConfig)
	assert.Equal(t, modelspec.MediaTypeModelWeightRaw, MediaTypeWeightRaw)
	assert.Equal(t, modelspec.MediaTypeModelWeightConfigRaw, MediaTypeWeightConfigRaw)
	assert.Equal(t, modelspec.MediaTypeModelDocRaw, MediaTypeDocRaw)
	assert.Equal(t, modelspec.MediaTypeModelCodeRaw, MediaTypeCodeRaw)
	assert.Equal(t, modelspec.AnnotationFilepath, AnnotationFilepath)
}

func TestBuildErrors(t *testing.T) {
	commit := "0123456789abcdef0123456789abcdef01234567"
	cases := []struct {
		name     string
		snapshot *hub.Snapshot
		digests  map[string]digest.Digest
	}{
		{"nil snapshot", nil, nil},
		{"invalid commit", &hub.Snapshot{ModelID: "a/b", Commit: "main", Files: []hub.File{{Path: "x.safetensors", Size: 1, SHA256: sha(1)}}}, nil},
		{"no files", &hub.Snapshot{ModelID: "a/b", Commit: commit}, nil},
		{"only gitattributes", &hub.Snapshot{ModelID: "a/b", Commit: commit, Files: []hub.File{{Path: ".gitattributes", Size: 1}}}, nil},
		{"missing non-LFS digest", &hub.Snapshot{ModelID: "a/b", Commit: commit, Files: []hub.File{{Path: "config.json", Size: 1}}}, nil},
		{"non-sha256 digest", &hub.Snapshot{ModelID: "a/b", Commit: commit, Files: []hub.File{{Path: "config.json", Size: 1}}},
			map[string]digest.Digest{"config.json": digest.SHA512.FromString("x")}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Build(c.snapshot, c.digests)
			assert.Error(t, err)
		})
	}
}

func TestBuildDeterminism(t *testing.T) {
	commit := "0123456789abcdef0123456789abcdef01234567"
	files := []hub.File{
		{Path: "b/model.safetensors", Size: 10, SHA256: sha(1)},
		{Path: "a.json", Size: 2},
		{Path: "B.md", Size: 3},
	}
	digests := map[string]digest.Digest{"a.json": digest.FromString("a"), "B.md": digest.FromString("b")}
	reversed := []hub.File{files[2], files[1], files[0]}
	s1 := &hub.Snapshot{ModelID: "a/b", Commit: commit, Files: files, LastModified: time.Date(2020, 1, 2, 3, 4, 5, 0, time.FixedZone("x", 3600))}
	s2 := &hub.Snapshot{ModelID: "a/b", Commit: commit, Files: reversed, LastModified: s1.LastModified.UTC()}

	a1, err := Build(s1, digests)
	require.NoError(t, err)
	a2, err := Build(s2, digests)
	require.NoError(t, err)
	assert.Equal(t, a1.Manifest, a2.Manifest)
	assert.Equal(t, []string{"B.md", "a.json", "b/model.safetensors"}, []string{a1.Layers[0].Path, a1.Layers[1].Path, a1.Layers[2].Path})
	assert.Contains(t, string(a1.Config), `"createdAt":"2020-01-02T02:04:05Z"`)

	s3 := *s1
	s3.Commit = "1123456789abcdef0123456789abcdef01234567"
	a3, err := Build(&s3, digests)
	require.NoError(t, err)
	assert.NotEqual(t, a1.Digest, a3.Digest)
}

func TestFormat(t *testing.T) {
	cases := []struct {
		files []string
		want  string
	}{
		{[]string{"model-00001-of-00002.safetensors", "model-00002-of-00002.safetensors", "config.json"}, "safetensors"},
		{[]string{"q4.gguf", "q8.gguf"}, "gguf"},
		{[]string{"model.onnx", "model.onnx_data"}, "onnx"},
		{[]string{"model.safetensors", "pytorch_model.bin"}, ""},
		{[]string{"config.json"}, ""},
	}
	for _, c := range cases {
		var layers []Layer
		for _, f := range c.files {
			layers = append(layers, Layer{Path: f, MediaType: classify(f, 1).mediaType()})
		}
		assert.Equal(t, c.want, format(layers), c.files)
	}
}

func sha(n byte) string {
	return digest.FromBytes([]byte{n}).Encoded()
}
