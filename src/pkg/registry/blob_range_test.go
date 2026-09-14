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

package registry

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
)

func TestPullBlobRange(t *testing.T) {
	for _, tt := range []struct {
		name         string
		status       int
		contentRange string
		body         string
	}{
		{"partial", http.StatusPartialContent, "bytes 2-5/10", "2345"},
		{"range ignored", http.StatusOK, "", "0123456789"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/v2/library/test/blobs/digest", r.URL.Path)
				assert.Equal(t, "bytes=2-5", r.Header.Get("Range"))
				assert.Equal(t, `"version"`, r.Header.Get("If-Range"))
				assert.Equal(t, "identity", r.Header.Get("Accept-Encoding"))
				w.Header().Set("Content-Range", tt.contentRange)
				w.Header().Set("Content-Length", strconv.Itoa(len(tt.body)))
				w.Header().Set("Docker-Content-Digest", "digest")
				w.Header().Set("ETag", `"version"`)
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			defer server.Close()
			client := NewClientWithAuthorizer(server.URL, nil, false, "")
			resp, err := client.PullBlobRange(context.Background(), "library/test", "digest", "bytes=2-5", `"version"`)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tt.status, resp.StatusCode)
			require.Equal(t, tt.contentRange, resp.Header.Get("Content-Range"))
			require.Equal(t, "digest", resp.Header.Get("Docker-Content-Digest"))
			require.Equal(t, `"version"`, resp.Header.Get("ETag"))
			require.EqualValues(t, len(tt.body), resp.ContentLength)
			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, tt.body, string(data))
		})
	}
}

func TestPullBlobRangeErrors(t *testing.T) {
	for _, tt := range []struct {
		status int
		code   string
	}{
		{http.StatusUnauthorized, errors.UnAuthorizedCode},
		{http.StatusForbidden, errors.ForbiddenCode},
		{http.StatusNotFound, errors.NotFoundCode},
		{http.StatusTooManyRequests, errors.RateLimitCode},
		{http.StatusRequestedRangeNotSatisfiable, errors.GeneralCode},
	} {
		t.Run(strconv.Itoa(tt.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()
			client := NewClientWithAuthorizer(server.URL, nil, false, "")
			resp, err := client.PullBlobRange(context.Background(), "test", "digest", "bytes=0-1", "")
			require.Nil(t, resp)
			require.True(t, errors.IsErr(err, tt.code), "%v", err)
		})
	}
}

func TestPullBlobRangeGetsChunkedResponseLengthFromHead(t *testing.T) {
	const blob = "0123456789"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v2/test/blobs/digest", r.URL.Path)
		assert.Equal(t, "identity", r.Header.Get("Accept-Encoding"))
		switch r.Method {
		case http.MethodGet:
			assert.Equal(t, "bytes=0-", r.Header.Get("Range"))
			w.WriteHeader(http.StatusOK)
			w.(http.Flusher).Flush()
			_, _ = io.WriteString(w, blob)
		case http.MethodHead:
			assert.Empty(t, r.Header.Get("Range"))
			w.Header().Set("Content-Length", strconv.Itoa(len(blob)))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()
	client := NewClientWithAuthorizer(server.URL, nil, false, "")
	resp, err := client.PullBlobRange(context.Background(), "test", "digest", "bytes=0-", "")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.EqualValues(t, len(blob), resp.ContentLength)
	require.Empty(t, resp.Header.Get("Content-Length"))
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, blob, string(data))
}

func TestPullBlobRangeRedirect(t *testing.T) {
	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "bytes=2-", r.Header.Get("Range"))
		assert.Equal(t, `"version"`, r.Header.Get("If-Range"))
		w.Header().Set("Content-Range", "bytes 2-3/4")
		w.WriteHeader(http.StatusPartialContent)
		_, _ = io.WriteString(w, "23")
	}))
	defer storage.Close()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, storage.URL, http.StatusTemporaryRedirect)
	}))
	defer upstream.Close()
	client := NewClientWithAuthorizer(upstream.URL, nil, false, "")
	resp, err := client.PullBlobRange(context.Background(), "test", "digest", "bytes=2-", `"version"`)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusPartialContent, resp.StatusCode)
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "23", string(data))
}

func TestPullBlobRangeCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := NewClientWithAuthorizer("http://127.0.0.1:1", nil, false, "")
	_, err := client.PullBlobRange(ctx, "test", "digest", "bytes=0-1", "")
	require.ErrorIs(t, err, context.Canceled)
}
