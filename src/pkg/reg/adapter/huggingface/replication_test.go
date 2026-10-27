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

package huggingface

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"testing"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	trans "github.com/goharbor/harbor/src/controller/replication/transfer"
	_ "github.com/goharbor/harbor/src/controller/replication/transfer/image"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub/hubtest"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const memoryRegistryType = "hf-test-memory"

// memoryRegistry is an in-memory destination that verifies every pushed blob against its digest.
type memoryRegistry struct {
	mu            sync.Mutex
	blobs         map[string][]byte
	uploads       map[string]*bytes.Buffer
	manifests     map[string][]byte
	blobPushes    int
	chunkPushes   int
	manifestPushs int
}

var destination = &memoryRegistry{}

func (m *memoryRegistry) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blobs = map[string][]byte{}
	m.uploads = map[string]*bytes.Buffer{}
	m.manifests = map[string][]byte{}
	m.blobPushes, m.chunkPushes, m.manifestPushs = 0, 0, 0
}

type memoryFactory struct{}

func (memoryFactory) Create(*model.Registry) (adp.Adapter, error) { return destination, nil }
func (memoryFactory) AdapterPattern() *model.AdapterPattern       { return nil }

func init() {
	if err := adp.RegisterFactory(memoryRegistryType, memoryFactory{}); err != nil {
		panic(err)
	}
}

func (m *memoryRegistry) Info() (*model.RegistryInfo, error)     { return &model.RegistryInfo{}, nil }
func (m *memoryRegistry) PrepareForPush([]*model.Resource) error { return nil }
func (m *memoryRegistry) HealthCheck() (string, error)           { return model.Healthy, nil }
func (m *memoryRegistry) FetchArtifacts([]*model.Filter) ([]*model.Resource, error) {
	return nil, nil
}

func (m *memoryRegistry) ManifestExist(repository, reference string) (bool, *distribution.Descriptor, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	payload, ok := m.manifests[repository+":"+reference]
	if !ok {
		return false, nil, nil
	}
	return true, &distribution.Descriptor{MediaType: v1.MediaTypeImageManifest, Digest: digest.FromBytes(payload), Size: int64(len(payload))}, nil
}

func (m *memoryRegistry) PullManifest(string, string, ...string) (distribution.Manifest, string, error) {
	return nil, "", fmt.Errorf("not implemented")
}

func (m *memoryRegistry) PushManifest(repository, reference, _ string, payload []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var parsed v1.Manifest
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return "", err
	}
	for _, d := range append([]v1.Descriptor{parsed.Config}, parsed.Layers...) {
		if _, ok := m.blobs[d.Digest.String()]; !ok {
			return "", fmt.Errorf("manifest references missing blob %s", d.Digest)
		}
	}
	m.manifestPushs++
	m.manifests[repository+":"+reference] = payload
	m.manifests[repository+":"+digest.FromBytes(payload).String()] = payload
	return digest.FromBytes(payload).String(), nil
}

func (m *memoryRegistry) DeleteManifest(string, string) error { return nil }

func (m *memoryRegistry) BlobExist(_, dgst string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.blobs[dgst]
	return ok, nil
}

func (m *memoryRegistry) PullBlob(string, string) (int64, io.ReadCloser, error) {
	return 0, nil, fmt.Errorf("not implemented")
}

func (m *memoryRegistry) PullBlobChunk(string, string, int64, int64, int64) (int64, io.ReadCloser, error) {
	return 0, nil, fmt.Errorf("not implemented")
}

func (m *memoryRegistry) store(dgst string, size int64, content []byte) error {
	if int64(len(content)) != size {
		return fmt.Errorf("blob %s has %d bytes, expected %d", dgst, len(content), size)
	}
	if digest.FromBytes(content).String() != dgst {
		return fmt.Errorf("blob content does not match %s", dgst)
	}
	m.blobs[dgst] = content
	return nil
}

func (m *memoryRegistry) PushBlob(_, dgst string, size int64, blob io.Reader) error {
	content, err := io.ReadAll(blob)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blobPushes++
	return m.store(dgst, size, content)
}

func (m *memoryRegistry) PushBlobChunk(_, dgst string, size int64, chunk io.Reader, start, end int64, location string) (string, int64, error) {
	content, err := io.ReadAll(chunk)
	if err != nil {
		return location, start - 1, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chunkPushes++
	if location == "" {
		location = "upload-" + strconv.Itoa(len(m.uploads))
		m.uploads[location] = &bytes.Buffer{}
	}
	buf := m.uploads[location]
	if int64(buf.Len()) != start || int64(len(content)) != end-start+1 {
		return location, start - 1, fmt.Errorf("chunk %d-%d does not continue upload at %d", start, end, buf.Len())
	}
	buf.Write(content)
	if end == size-1 {
		return location, end, m.store(dgst, size, buf.Bytes())
	}
	return location, end, nil
}

func (m *memoryRegistry) MountBlob(string, string, string) error { return nil }
func (m *memoryRegistry) CanBeMount(string) (bool, string, error) {
	return false, "", nil
}
func (m *memoryRegistry) DeleteTag(string, string) error { return nil }
func (m *memoryRegistry) ListTags(string) ([]string, error) {
	return nil, nil
}
func (m *memoryRegistry) ListReferrers(string, string, string) (*v1.Index, map[string][]string, error) {
	return nil, nil, nil
}

func clearProcessCache(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	it, err := processCache().Scan(ctx, "huggingface:")
	require.NoError(t, err)
	for it.Next(ctx) {
		require.NoError(t, processCache().Delete(ctx, it.Val()))
	}
}

// TestReplicationTransfer runs the real transfer with the adapter created by its factory, as
// jobservice does, against an in-memory destination.
func TestReplicationTransfer(t *testing.T) {
	for _, byChunk := range []bool{false, true} {
		t.Run(fmt.Sprintf("copy by chunk %v", byChunk), func(t *testing.T) {
			h := hubtest.New(tinyModel())
			defer h.Close()
			destination.reset()
			clearProcessCache(t)

			srcReg := hubRegistry(h, 7, "")
			factory, err := adp.GetFactory(model.RegistryTypeHuggingFace)
			require.NoError(t, err)
			src, err := factory.Create(srcReg)
			require.NoError(t, err)
			resources, err := src.(adp.ArtifactRegistry).FetchArtifacts([]*model.Filter{
				{Type: model.FilterTypeName, Value: "org/tiny-model"},
				{Type: model.FilterTypeTag, Value: "{main,v1}"},
			})
			require.NoError(t, err)
			require.Len(t, resources, 1)
			srcRes := resources[0]

			dstRes := &model.Resource{
				Type:     srcRes.Type,
				Registry: &model.Registry{Type: memoryRegistryType},
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{Name: "library/tiny-model"},
					Artifacts:  srcRes.Metadata.Artifacts,
				},
			}
			run := func() {
				t.Helper()
				tf, err := trans.GetFactory(srcRes.Type)
				require.NoError(t, err)
				tr, err := tf(log.DefaultLogger(), func() bool { return false })
				require.NoError(t, err)
				require.NoError(t, tr.Transfer(srcRes, dstRes, &trans.Options{CopyByChunk: byChunk}))
			}
			digestOf := func(tag string) digest.Digest {
				exist, desc, err := destination.ManifestExist("library/tiny-model", tag)
				require.NoError(t, err)
				require.True(t, exist, tag)
				return desc.Digest
			}

			run()
			mainDigest, v1Digest := digestOf("main"), digestOf("v1")
			assert.Equal(t, 2, destination.manifestPushs)
			// main and v1 share config.json, modeling.py and the weight; both READMEs and configs differ.
			if byChunk {
				assert.Positive(t, destination.chunkPushes)
			} else {
				assert.Zero(t, destination.chunkPushes)
			}
			pushes := destination.blobPushes + destination.chunkPushes

			h.ResetCalls()
			run()
			assert.Equal(t, mainDigest, digestOf("main"))
			assert.Equal(t, v1Digest, digestOf("v1"))
			assert.Equal(t, 2, destination.manifestPushs, "run two skips both artifacts")
			assert.Equal(t, pushes, destination.blobPushes+destination.chunkPushes)
			assert.Zero(t, h.Calls("resolve"), "run two downloads nothing")

			// A jobservice restart loses the process cache; the digests stay the same.
			clearProcessCache(t)
			run()
			assert.Equal(t, mainDigest, digestOf("main"))
			assert.Equal(t, v1Digest, digestOf("v1"))
			assert.Equal(t, 2, destination.manifestPushs)
		})
	}
}
