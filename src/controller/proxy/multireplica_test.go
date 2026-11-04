//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/registry"
)

// replicatedRegistry emulates a registry running as several replicas behind a
// non-sticky service on shared storage.
//
// An upload opened by a POST is only known to the replica that served the POST,
// so a PUT routed to another replica fails with BLOB_UPLOAD_UNKNOWN even though
// nothing is wrong. Pushing a blob is exactly that pair of requests, which is why
// a single attempt is not enough to cache content reliably.
type replicatedRegistry struct {
	t *testing.T

	mu      sync.Mutex
	uploads map[int]map[string]bool // replica -> upload ids opened on it
	blobs   map[string][]byte       // digest -> stored content
	reqs    int                     // requests that reached the service
	ids     int

	// route reports which replica serves request number n (1-based).
	route func(n int) int
}

func newReplicatedRegistry(t *testing.T, route func(n int) int) *replicatedRegistry {
	return &replicatedRegistry{
		t:       t,
		uploads: map[int]map[string]bool{},
		blobs:   map[string][]byte{},
		route:   route,
	}
}

func (f *replicatedRegistry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.reqs++
	n, replica := f.reqs, f.route(f.reqs)
	f.mu.Unlock()

	switch {
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/blobs/uploads/"):
		f.mu.Lock()
		f.ids++
		id := fmt.Sprintf("upload-%d", f.ids)
		if f.uploads[replica] == nil {
			f.uploads[replica] = map[string]bool{}
		}
		f.uploads[replica][id] = true
		f.mu.Unlock()

		repo := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v2/"), "/blobs/uploads/")
		w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads/%s", repo, id))
		w.Header().Set("Docker-Upload-UUID", id)
		w.WriteHeader(http.StatusAccepted)
		f.t.Logf("request %d -> replica %d: POST opened upload %s", n, replica, id)

	case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/blobs/uploads/"):
		parts := strings.Split(r.URL.Path, "/")
		id := parts[len(parts)-1]

		f.mu.Lock()
		known := f.uploads[replica][id]
		f.mu.Unlock()
		if !known {
			f.t.Logf("request %d -> replica %d: PUT %s rejected, blob upload unknown", n, replica, id)
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errors":[{"code":"BLOB_UPLOAD_UNKNOWN","message":"blob upload unknown to registry"}]}`))
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		f.mu.Lock()
		f.blobs[r.URL.Query().Get("digest")] = body
		f.mu.Unlock()
		f.t.Logf("request %d -> replica %d: PUT %s stored %d bytes", n, replica, id, len(body))
		w.WriteHeader(http.StatusCreated)

	default:
		w.WriteHeader(http.StatusOK)
	}
}

func (f *replicatedRegistry) blob(dig digest.Digest) ([]byte, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.blobs[dig.String()]
	return b, ok
}

// countingRemote hands out a fresh reader over the blob on every call and counts
// how many times it was asked for one.
type countingRemote struct {
	fakeRemote
	opens int
}

func (s *countingRemote) BlobReader(repo, dig string) (int64, io.ReadCloser, error) {
	s.opens++
	return s.fakeRemote.BlobReader(repo, dig)
}

// splitFirstAttempt routes the first attempt's POST and PUT to different replicas
// and keeps everything afterwards on one replica, i.e. a later attempt gets the
// two requests to stick together.
//
// The registry client pings the endpoint before the push, so the first attempt's
// POST and PUT are requests 2 and 3.
func splitFirstAttempt(n int) int {
	if n == 3 {
		return 1
	}
	return 0
}

// TestPushBlobSingleAttemptFailsAcrossReplicas pins the behaviour that makes the
// retry necessary: on its own, one push attempt is lost when the service splits
// the upload across replicas.
func TestPushBlobSingleAttemptFailsAcrossReplicas(t *testing.T) {
	content := []byte("layer content that should end up cached locally")
	dig := digest.FromBytes(content)
	desc := distribution.Descriptor{Digest: dig, Size: int64(len(content))}

	reg := newReplicatedRegistry(t, splitFirstAttempt)
	srv := httptest.NewServer(reg)
	defer srv.Close()

	local := &localHelper{registry: registry.NewClient(srv.URL, "", "", true)}

	err := local.PushBlob("library/hello-world", desc, io.NopCloser(bytes.NewReader(content)))
	require.Error(t, err)
	require.Contains(t, err.Error(), "BLOB_UPLOAD_UNKNOWN")

	_, cached := reg.blob(dig)
	require.False(t, cached, "the blob must not have been stored")
}

// TestPutBlobToLocalRecoversAcrossReplicas is the regression test for #23926: the
// same split that loses a single attempt must not leave the artifact uncached.
func TestPutBlobToLocalRecoversAcrossReplicas(t *testing.T) {
	content := []byte("layer content that should end up cached locally")
	dig := digest.FromBytes(content)
	desc := distribution.Descriptor{Digest: dig, Size: int64(len(content))}
	repo := "library/hello-world"

	reg := newReplicatedRegistry(t, splitFirstAttempt)
	srv := httptest.NewServer(reg)
	defer srv.Close()

	local := &localHelper{registry: registry.NewClient(srv.URL, "", "", true)}
	remote := &countingRemote{fakeRemote: fakeRemote{content: content}}

	err := putBlobToLocal(context.Background(), local, repo, repo, desc, remote)
	require.NoError(t, err)

	cached, ok := reg.blob(dig)
	require.True(t, ok, "the blob must be stored after the retry")
	require.Equal(t, content, cached, "the retried push must store the whole blob")
	require.Equal(t, dig, digest.FromBytes(cached))

	// the second attempt read the blob from upstream again rather than replaying
	// the reader the failed attempt had already drained
	require.Equal(t, 2, remote.opens, "the upstream reader must be re-opened per attempt")
}
