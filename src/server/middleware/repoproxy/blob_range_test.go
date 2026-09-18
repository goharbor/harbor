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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsSingleRange(t *testing.T) {
	for _, value := range []string{
		"bytes=0-0",
		"bytes=1024-2047",
		"bytes=1-",
		"bytes=0-9223372036854775807",
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
	rec := httptest.NewRecorder()
	resp := &http.Response{
		StatusCode: http.StatusPartialContent,
		Header: http.Header{
			"Content-Range": {"bytes 2-5/10"},
		},
		ContentLength: -1,
		Body:          io.NopCloser(bytes.NewBufferString("2345")),
	}
	require.NoError(t, serveBlobRange(rec, resp))
	require.Equal(t, "4", rec.Header().Get("Content-Length"))
	require.Equal(t, "2345", rec.Body.String())
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
