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

package repoproxy

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/controller/project"
	registryCtl "github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/lib"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/reg"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	registrytesting "github.com/goharbor/harbor/src/testing/controller/registry"
	testreg "github.com/goharbor/harbor/src/testing/pkg/reg"
)

func TestBlobGetMiddlewareRangeCase(t *testing.T) {
	// The proxy controller uses the shared insecure transport for local cache lookups.
	local := httptest.NewServer(http.NotFoundHandler())
	defer local.Close()
	transport := commonhttp.GetHTTPTransport(commonhttp.WithInsecure(true)).(*http.Transport)
	originalDial, originalProxy := transport.DialContext, transport.Proxy
	transport.CloseIdleConnections()
	transport.Proxy = nil
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, local.Listener.Addr().String())
	}
	defer func() {
		transport.CloseIdleConnections()
		transport.DialContext, transport.Proxy = originalDial, originalProxy
	}()

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			return
		}
		if r.Header.Get("Range") == "" {
			// Background cache fills are covered by the controller tests.
			w.WriteHeader(http.StatusNotFound)
			return
		}
		assert.Equal(t, "/v2/test/blobs/digest", r.URL.Path)
		http.ServeContent(w, r, "blob", time.Time{}, bytes.NewReader([]byte("0123456789")))
	}))
	defer upstream.Close()

	projectController := &projecttesting.Controller{}
	registryController := &registrytesting.Controller{}
	registryManager := &testreg.Manager{}
	originalProject, originalRegistry, originalManager := project.Ctl, registryCtl.Ctl, reg.Mgr
	project.Ctl, registryCtl.Ctl, reg.Mgr = projectController, registryController, registryManager
	defer func() {
		project.Ctl, registryCtl.Ctl, reg.Mgr = originalProject, originalRegistry, originalManager
		projectController.AssertExpectations(t)
		registryController.AssertExpectations(t)
		registryManager.AssertExpectations(t)
	}()
	art := lib.ArtifactInfo{ProjectName: "proxy", Repository: "proxy/test", Digest: "digest"}
	p := &proModels.Project{Name: art.ProjectName, RegistryID: 1}
	registry := &model.Registry{ID: 1, URL: upstream.URL, Type: model.RegistryTypeDockerRegistry, Status: model.Healthy}
	projectController.On("GetByName", mock.Anything, art.ProjectName, mock.Anything).Return(p, nil)
	registryController.On("Get", mock.Anything, int64(1)).Return(registry, nil)
	registryManager.On("Get", mock.Anything, int64(1)).Return(registry, nil)
	handler := BlobGetMiddleware()(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("unexpected local blob fallback")
	}))
	for _, byteRange := range []string{"bytes=0-1", "Bytes=0-1", "BYTES=0-1", "bYtEs=0-1"} {
		t.Run(byteRange, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v2/proxy/test/blobs/digest", nil)
			req = req.WithContext(lib.WithArtifactInfo(req.Context(), art))
			req.Header.Set("Range", byteRange)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			require.Equal(t, http.StatusPartialContent, rec.Code)
			require.Equal(t, "bytes 0-1/10", rec.Header().Get("Content-Range"))
			require.Equal(t, "2", rec.Header().Get("Content-Length"))
			require.Equal(t, "01", rec.Body.String())
		})
	}
}

func TestIsSingleRange(t *testing.T) {
	for _, value := range []string{
		"bytes=0-0",
		"bytes=1024-2047",
		"bytes=1-",
		"bytes=0-9223372036854775807",
		"Bytes=0-1",
		"BYTES=1-",
		"bYtEs=2-5",
	} {
		require.True(t, isSingleRange(value), value)
	}
	for _, value := range []string{
		"",
		"bytes=-10",
		"bytes=",
		"bytes=2-1",
		"bytes=0-1,4-5",
		"bytes=+1-2",
		"bytes=1-a",
		"bytes=1-2-3",
		"items=1-2",
		"bytes=9223372036854775808-",
		"bytes=0-9223372036854775808",
		"bytes= 1-2",
		"Bytes=-10",
		"BYTES=0-1,4-5",
		"Bytes=2-1",
		"bytes =0-1",
	} {
		require.False(t, isSingleRange(value), value)
	}
}

func TestServeBlobRange(t *testing.T) {
	for _, tt := range []struct {
		status       int
		contentRange string
		body         string
	}{
		{http.StatusPartialContent, "bytes 2-5/10", "2345"},
		{http.StatusOK, "", "0123456789"},
	} {
		t.Run(strconv.Itoa(tt.status), func(t *testing.T) {
			rec := httptest.NewRecorder()
			resp := &http.Response{
				StatusCode: tt.status,
				Header: http.Header{
					"Content-Length":        {strconv.Itoa(len(tt.body))},
					"Content-Range":         {tt.contentRange},
					"Content-Type":          {"application/octet-stream"},
					"Docker-Content-Digest": {"digest"},
					"Etag":                  {`"version"`},
					"Accept-Ranges":         {"bytes"},
					"Set-Cookie":            {"upstream=private"},
				},
				ContentLength: int64(len(tt.body)),
				Body:          io.NopCloser(bytes.NewBufferString(tt.body)),
			}
			require.NoError(t, serveBlobRange(rec, resp))
			require.Equal(t, tt.status, rec.Code)
			require.Equal(t, tt.body, rec.Body.String())
			for _, name := range []string{
				"Content-Length",
				"Content-Range",
				"Content-Type",
				"Docker-Content-Digest",
				"ETag",
				"Accept-Ranges",
			} {
				require.Equal(t, resp.Header.Get(name), rec.Header().Get(name), name)
			}
			require.Empty(t, rec.Header().Get("Set-Cookie"))
		})
	}
}

func TestServeBlobRangeLengthFromContentRange(t *testing.T) {
	for _, unit := range []string{"bytes", "Bytes", "BYTES", "bYtEs"} {
		t.Run(unit, func(t *testing.T) {
			rec := httptest.NewRecorder()
			resp := &http.Response{
				StatusCode: http.StatusPartialContent,
				Header: http.Header{
					"Content-Range": {unit + " 2-5/10"},
				},
				ContentLength: -1,
				Body:          io.NopCloser(bytes.NewBufferString("2345")),
			}
			require.NoError(t, serveBlobRange(rec, resp))
			require.Equal(t, "4", rec.Header().Get("Content-Length"))
			require.Equal(t, "2345", rec.Body.String())
		})
	}
}

func TestServeBlobRangeUnsatisfiable(t *testing.T) {
	for _, tt := range []struct {
		name          string
		body          string
		contentLength int64
		streamError   bool
	}{
		{name: "known length", body: "out of range", contentLength: 12},
		{name: "unknown length", body: "out of range", contentLength: -1},
		{name: "empty"},
		{name: "truncated", body: "out of", contentLength: 12, streamError: true},
		{name: "unknown length read error", body: "out of", contentLength: -1, streamError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			var reader io.Reader = bytes.NewBufferString(tt.body)
			if tt.streamError {
				reader = io.MultiReader(reader, &errReader{err: io.ErrUnexpectedEOF})
			}
			resp := &http.Response{
				StatusCode: http.StatusRequestedRangeNotSatisfiable,
				Header: http.Header{
					"Content-Range": {"bytes */10"},
					"Content-Type":  {"text/plain"},
				},
				ContentLength: tt.contentLength,
				Body:          io.NopCloser(reader),
			}
			err := serveBlobRange(rec, resp)
			if tt.streamError {
				require.ErrorIs(t, err, errBlobRangeStream)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, http.StatusRequestedRangeNotSatisfiable, rec.Code)
			require.Equal(t, "bytes */10", rec.Header().Get("Content-Range"))
			require.Equal(t, "text/plain", rec.Header().Get("Content-Type"))
			if tt.contentLength >= 0 {
				require.Equal(t, strconv.FormatInt(tt.contentLength, 10), rec.Header().Get("Content-Length"))
			} else {
				require.Empty(t, rec.Header().Get("Content-Length"))
			}
			require.Equal(t, tt.body, rec.Body.String())
		})
	}
}

func TestServeBlobRangeRejectsUnknownLength(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		header http.Header
	}{
		{name: "OK", status: http.StatusOK, header: make(http.Header)},
		{name: "missing content range", status: http.StatusPartialContent, header: make(http.Header)},
		{
			name:   "invalid content range",
			status: http.StatusPartialContent,
			header: http.Header{"Content-Range": {"bytes 5-2/10"}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			resp := &http.Response{
				StatusCode:    tt.status,
				Header:        tt.header,
				ContentLength: -1,
				Body:          io.NopCloser(bytes.NewBufferString("partial")),
			}
			require.Error(t, serveBlobRange(rec, resp))
			require.Empty(t, rec.Body.String())
		})
	}
}

func TestServeBlobRangeTruncated(t *testing.T) {
	const size = 8192
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := &http.Response{
			StatusCode: http.StatusPartialContent,
			Header: http.Header{
				"Content-Range": {"bytes 0-8191/16384"},
			},
			ContentLength: -1,
			Body: io.NopCloser(io.MultiReader(
				bytes.NewReader(bytes.Repeat([]byte("a"), 4096)),
				&errReader{err: io.ErrUnexpectedEOF},
			)),
		}
		if err := serveBlobRange(w, resp); !errors.Is(err, errBlobRangeStream) {
			t.Error(err)
		}
	}))
	defer server.Close()
	resp, err := http.Get(server.URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	require.Error(t, err)
}

func TestServeBlobRangeStreamError(t *testing.T) {
	for _, partial := range []string{"", "partial"} {
		t.Run(partial, func(t *testing.T) {
			rec := httptest.NewRecorder()
			resp := &http.Response{
				StatusCode:    http.StatusPartialContent,
				Header:        http.Header{"Content-Range": {"bytes 0-9/20"}, "Content-Length": {"10"}},
				ContentLength: 10,
				Body: io.NopCloser(io.MultiReader(
					bytes.NewBufferString(partial),
					&errReader{err: io.ErrUnexpectedEOF},
				)),
			}
			err := serveBlobRange(rec, resp)
			require.ErrorIs(t, err, errBlobRangeStream)
			require.Equal(t, http.StatusPartialContent, rec.Code)
			require.Equal(t, partial, rec.Body.String())
		})
	}
}
