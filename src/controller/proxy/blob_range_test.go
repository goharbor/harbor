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
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/docker/distribution"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/reg"
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	testreg "github.com/goharbor/harbor/src/testing/pkg/reg"
)

type cachedRangeBlob struct {
	repository string
	desc       distribution.Descriptor
	data       string
	err        error
}

type rangeLocalHelper struct {
	localInterface
	cached chan cachedRangeBlob
}

func (l *rangeLocalHelper) PushBlob(repo string, desc distribution.Descriptor, reader io.ReadCloser) error {
	data, err := io.ReadAll(reader)
	l.cached <- cachedRangeBlob{
		repository: repo,
		desc:       desc,
		data:       string(data),
		err:        err,
	}
	return err
}

func setupRangeRegistry(t *testing.T, upstream string) *proModels.Project {
	t.Helper()
	original := reg.Mgr
	manager := &testreg.Manager{}
	reg.Mgr = manager
	t.Cleanup(func() {
		reg.Mgr = original
		manager.AssertExpectations(t)
	})
	manager.On("Get", mock.Anything, int64(1)).Return(&model.Registry{
		ID:     1,
		URL:    upstream,
		Type:   model.RegistryTypeDockerRegistry,
		Status: model.Healthy,
	}, nil).Once()
	return &proModels.Project{
		RegistryID: 1,
		Metadata:   map[string]string{proModels.ProMetaProxySpeed: "1024"},
	}
}

// TestProxyBlobRangeCachesFullBlob exercises the public controller, native adapter and registry client,
// including the separate full download used for caching.
func TestProxyBlobRangeCachesFullBlob(t *testing.T) {
	for _, status := range []int{http.StatusPartialContent, http.StatusOK} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			const fullBlob = "0123456789"
			frontBlob := fullBlob
			if status == http.StatusPartialContent {
				frontBlob = "2345"
			}
			var requests atomic.Int32
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v2/" {
					return
				}
				assert.Equal(t, "/v2/library/test/blobs/digest", r.URL.Path)
				requests.Add(1)
				if r.Header.Get("Range") == "" {
					assert.Empty(t, r.Header.Get("If-Range"))
					w.Header().Set("Content-Length", strconv.Itoa(len(fullBlob)))
					_, _ = io.WriteString(w, fullBlob)
					return
				}
				assert.Equal(t, "bytes=2-5", r.Header.Get("Range"))
				assert.Equal(t, `"version"`, r.Header.Get("If-Range"))
				w.Header().Set("Content-Length", strconv.Itoa(len(frontBlob)))
				if status == http.StatusPartialContent {
					w.Header().Set("Content-Range", "bytes 2-5/10")
				}
				w.WriteHeader(status)
				_, _ = io.WriteString(w, frontBlob)
			}))
			defer upstream.Close()
			project := setupRangeRegistry(t, upstream.URL)
			local := &rangeLocalHelper{cached: make(chan cachedRangeBlob, 1)}
			ctl := &controller{local: local}
			art := lib.ArtifactInfo{ProjectName: "proxy", Repository: "proxy/library/test", Digest: "digest"}
			resp, err := ctl.ProxyBlobRange(context.Background(), project, art, "bytes=2-5", `"version"`)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, status, resp.StatusCode)
			require.EqualValues(t, len(frontBlob), resp.ContentLength)
			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, frontBlob, string(data))
			select {
			case cached := <-local.cached:
				require.NoError(t, cached.err)
				require.Equal(t, art.Repository, cached.repository)
				require.Equal(t, art.Digest, string(cached.desc.Digest))
				require.EqualValues(t, len(fullBlob), cached.desc.Size)
				require.Equal(t, fullBlob, cached.data)
			case <-time.After(5 * time.Second):
				t.Fatal("complete blob was not cached")
			}
			require.EqualValues(t, 2, requests.Load())
		})
	}
}

func TestProxyBlobRangeNotFound(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			return
		}
		assert.Equal(t, "bytes=0-1", r.Header.Get("Range"))
		w.WriteHeader(http.StatusNotFound)
	}))
	defer upstream.Close()
	project := setupRangeRegistry(t, upstream.URL)
	ctl := &controller{}
	art := lib.ArtifactInfo{ProjectName: "proxy", Repository: "proxy/test", Digest: "digest"}
	resp, err := ctl.ProxyBlobRange(context.Background(), project, art, "bytes=0-1", "")
	require.True(t, errors.IsNotFoundErr(err), "%v", err)
	require.Nil(t, resp)
}

// TestProxyBlobRangeBackgroundCache verifies that a slow complete download does not block
// the ranged response or inherit its cancellation.
func TestProxyBlobRangeBackgroundCache(t *testing.T) {
	fullRequested := make(chan struct{})
	releaseFull := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			return
		}
		if r.Header.Get("Range") != "" {
			w.Header().Set("Content-Range", "bytes 2-5/10")
			w.Header().Set("Content-Length", "4")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = io.WriteString(w, "2345")
			return
		}
		close(fullRequested)
		<-releaseFull
		w.Header().Set("Content-Length", "10")
		_, _ = io.WriteString(w, "0123456789")
	}))
	defer upstream.Close()
	defer func() {
		select {
		case <-releaseFull:
		default:
			close(releaseFull)
		}
	}()
	project := setupRangeRegistry(t, upstream.URL)
	local := &rangeLocalHelper{cached: make(chan cachedRangeBlob, 1)}
	ctl := &controller{local: local}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	art := lib.ArtifactInfo{ProjectName: "proxy", Repository: "proxy/library/test", Digest: "digest"}
	resp, err := ctl.ProxyBlobRange(ctx, project, art, "bytes=2-5", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "2345", string(data))
	select {
	case <-fullRequested:
	case <-time.After(5 * time.Second):
		t.Fatal("background full download was not started")
	}
	cancel()
	close(releaseFull)
	select {
	case cached := <-local.cached:
		require.NoError(t, cached.err)
		require.EqualValues(t, 10, cached.desc.Size)
		require.Equal(t, "0123456789", cached.data)
	case <-time.After(5 * time.Second):
		t.Fatal("background cache fill did not finish after the range request was canceled")
	}
}
