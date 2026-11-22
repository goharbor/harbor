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

package proxy

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"testing/synctest"
	"time"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib"
	testregistry "github.com/goharbor/harbor/src/testing/pkg/registry"
)

func TestLocalPushWaitsForInflightResult(t *testing.T) {
	for _, operation := range []string{"blob", "manifest"} {
		for _, fails := range []bool{false, true} {
			name := operation + "/success"
			var pushErr error
			if fails {
				name = operation + "/failure"
				pushErr = errors.New("registry unavailable")
			}
			t.Run(name, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					client := &testregistry.Client{}
					repo := "library/concurrent"
					desc := distribution.Descriptor{Digest: digest.FromString("layer"), Size: 5}
					manifest := &mockManifest{}
					manifest.On("Payload").Return("application/json", []byte("{}"), nil)
					var call *mock.Call
					push := func(local *localHelper) error {
						reader := io.NopCloser(bytes.NewBufferString("layer"))
						defer reader.Close()
						return local.PushBlob(repo, desc, reader)
					}
					if operation == "blob" {
						call = client.On("PushBlob", repo, desc.Digest.String(), desc.Size, mock.Anything).Return(pushErr)
					} else {
						call = client.On("PushManifest", repo, desc.Digest.String(), "application/json", []byte("{}")).Return("", pushErr)
						push = func(local *localHelper) error {
							return local.PushManifest(repo, desc.Digest.String(), manifest)
						}
					}
					release := make(chan struct{})
					call.Run(func(mock.Arguments) { <-release }).Twice()
					results := make(chan error, 2)
					go func() { results <- push(&localHelper{registry: client}) }()
					synctest.Wait()
					// A separate helper must join the same push, not report success
					// while the registry call is still blocked.
					go func() { results <- push(&localHelper{registry: client}) }()
					synctest.Wait()
					require.Empty(t, results, "both callers must wait for the registry")
					close(release)
					for range 2 {
						require.ErrorIs(t, <-results, pushErr)
					}
					// Completed calls must not be retained, including failures:
					// a subsequent attempt needs to contact the registry again.
					require.ErrorIs(t, push(&localHelper{registry: client}), pushErr)
					client.AssertExpectations(t)
				})
			})
		}
	}
}

func TestManifestCacheWaitsForInflightBlob(t *testing.T) {
	for _, fails := range []bool{false, true} {
		name := "success"
		var pushErr error
		if fails {
			name = "failure"
			pushErr = errors.New("blob upload failed")
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				client := &testregistry.Client{}
				local := &localHelper{registry: client}
				repo := "library/cache-concurrent"
				content := []byte("layer")
				desc := distribution.Descriptor{Digest: digest.FromBytes(content), Size: int64(len(content))}
				manifest := &mockManifest{}
				manifest.On("References").Return([]distribution.Descriptor{desc})
				manifest.On("Payload").Return("application/json", []byte("{}"), nil)
				client.On("BlobExist", repo, desc.Digest.String()).Return(false, nil)
				release := make(chan struct{})
				client.On("PushBlob", repo, desc.Digest.String(), desc.Size, mock.Anything).
					Run(func(mock.Arguments) { <-release }).Return(pushErr).Once()
				if fails {
					client.On("PushBlob", repo, desc.Digest.String(), desc.Size, mock.Anything).
						Return(pushErr).Times(localPushAttempts - 1)
				} else {
					client.On("PushManifest", repo, "latest", "application/json", []byte("{}")).Return("", nil).Once()
				}
				owner := make(chan error, 1)
				go func() {
					reader := io.NopCloser(bytes.NewReader(content))
					defer reader.Close()
					owner <- local.PushBlob(repo, desc, reader)
				}()
				synctest.Wait()
				cached := make(chan struct{})
				go func() {
					cache := &ManifestCache{local: local}
					cache.CacheContent(context.Background(), repo, manifest,
						lib.ArtifactInfo{Repository: repo, Tag: "latest"}, &fakeRemote{content: content}, "")
					close(cached)
				}()
				// Advance virtual time past the dependency polling period so
				// CacheContent reaches the blob push while its owner is blocked.
				time.Sleep((maxManifestWait*sleepIntervalSec + 1) * time.Second)
				synctest.Wait()
				select {
				case <-cached:
					t.Fatal("caching finished before the in-flight blob upload")
				default:
				}
				client.AssertNotCalled(t, "PushManifest", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				close(release)
				require.ErrorIs(t, <-owner, pushErr)
				<-cached
				client.AssertExpectations(t)
			})
		})
	}
}
