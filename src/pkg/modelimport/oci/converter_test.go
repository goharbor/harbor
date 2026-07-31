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
	"io"
	"net/http"
	"testing"

	"github.com/docker/distribution"
	modelspec "github.com/modelpack/model-spec/specs-go/v1"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/modelimport/huggingface"
)

func TestPush(t *testing.T) {
	reg := &fakeRegistry{}
	converter := New(reg)
	snapshot := testSnapshot
	result, err := converter.Push(context.Background(), "library/model", "latest", snapshot, func(_ context.Context, filename string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader([]byte("content-" + filename))), nil
	})
	require.NoError(t, err)
	require.Equal(t, "sha256:manifest", result.Digest)
	require.Len(t, reg.blobs, 3)
	require.Len(t, reg.manifests, 1)

	var manifest ocispec.Manifest
	require.NoError(t, json.Unmarshal(reg.manifests[0], &manifest))
	require.Equal(t, modelspec.ArtifactTypeModelManifest, manifest.ArtifactType)
	require.Equal(t, "README.md", manifest.Layers[0].Annotations[modelspec.AnnotationFilepath])
	require.Equal(t, modelspec.MediaTypeModelDocRaw, manifest.Layers[0].MediaType)

	config := map[string]any{}
	require.NoError(t, json.Unmarshal(reg.blobs[manifest.Config.Digest.String()], &config))
	hf := config["huggingface"].(map[string]any)
	require.Equal(t, "org/model", hf["repo_id"])
	require.Equal(t, "abc123", hf["commit_sha"])
}

var testSnapshot = &huggingface.Snapshot{
	RepoID:    "org/model",
	Revision:  "main",
	CommitSHA: "abc123",
	SourceURL: "https://huggingface.co/org/model",
	CardData: map[string]any{
		"license": "apache-2.0",
	},
	Files: []huggingface.File{
		{Path: "README.md", Size: 10},
		{Path: "model.safetensors", Size: 20},
	},
}

type fakeRegistry struct {
	blobs     map[string][]byte
	manifests [][]byte
}

func (f *fakeRegistry) PushBlob(_ string, dgst string, _ int64, reader io.Reader) error {
	if f.blobs == nil {
		f.blobs = map[string][]byte{}
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	f.blobs[dgst] = data
	return nil
}

func (f *fakeRegistry) PushManifest(_, _, _ string, payload []byte) (string, error) {
	f.manifests = append(f.manifests, payload)
	return "sha256:manifest", nil
}

func (f *fakeRegistry) Ping() error                       { return nil }
func (f *fakeRegistry) Catalog() ([]string, error)        { return nil, nil }
func (f *fakeRegistry) ListTags(string) ([]string, error) { return nil, nil }
func (f *fakeRegistry) ManifestExist(string, string) (bool, *distribution.Descriptor, error) {
	return false, nil, nil
}
func (f *fakeRegistry) PullManifest(string, string, ...string) (distribution.Manifest, string, error) {
	return nil, "", nil
}
func (f *fakeRegistry) DeleteManifest(string, string) error    { return nil }
func (f *fakeRegistry) BlobExist(string, string) (bool, error) { return false, nil }
func (f *fakeRegistry) PullBlob(string, string) (int64, io.ReadCloser, error) {
	return 0, nil, nil
}
func (f *fakeRegistry) PullBlobChunk(string, string, int64, int64, int64) (int64, io.ReadCloser, error) {
	return 0, nil, nil
}
func (f *fakeRegistry) PushBlobChunk(string, string, int64, io.Reader, int64, int64, string) (string, int64, error) {
	return "", 0, nil
}
func (f *fakeRegistry) MountBlob(string, string, string) error          { return nil }
func (f *fakeRegistry) DeleteBlob(string, string) error                 { return nil }
func (f *fakeRegistry) Copy(string, string, string, string, bool) error { return nil }
func (f *fakeRegistry) Do(*http.Request) (*http.Response, error)        { return nil, nil }
func (f *fakeRegistry) ListReferrers(string, string, string) (*ocispec.Index, map[string][]string, error) {
	return nil, nil, nil
}
