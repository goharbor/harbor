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
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/controller/proxy"
	"github.com/goharbor/harbor/src/lib/errors"
	libhttp "github.com/goharbor/harbor/src/lib/http"
)

// These tests cover the verifying reader and serveBlob together: the reader withholds
// a tail on mismatch, and serveBlob must turn that into a response clients reject.

func blobBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)

	return b
}

// serveVerified mirrors handleBlob: wrap the upstream reader, then serve it.
func serveVerified(t *testing.T, body []byte, requested string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reader, err := proxy.NewVerifyingReader(io.NopCloser(bytes.NewReader(body)), requested)
		if err != nil {
			libhttp.SendError(w, err)
			return
		}
		defer reader.Close()
		if err := serveBlob(w, reader, int64(len(body)), requested); err != nil {
			libhttp.SendError(w, err)
		}
	}))
}

func TestServeBlobVerifiedContentCompletes(t *testing.T) {
	body := blobBytes(t, 5*128*1024)
	dgst := digest.FromBytes(body).String()

	server := serveVerified(t, body, dgst)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, dgst, resp.Header.Get("Docker-Content-Digest"))

	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, body, got)
}

// The 200 and Content-Length are already committed, so the protection is that the
// body arrives short and no error payload is appended to the blob stream.
func TestServeBlobLargeMismatchTruncatesResponse(t *testing.T) {
	body := blobBytes(t, 5*128*1024)
	wrong := digest.FromString("a different blob").String()

	server := serveVerified(t, body, wrong)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	got, readErr := io.ReadAll(resp.Body)
	require.Error(t, readErr, "client must not see a complete body")
	assert.ErrorIs(t, readErr, io.ErrUnexpectedEOF)
	assert.Less(t, len(got), len(body), "response must be short of Content-Length")
	assert.NotContains(t, string(got), "errors", "no JSON error payload may be appended to the blob stream")
}

// A body fitting inside the withheld tail never reaches the client, so serveBlob can
// still drop the blob headers and the caller emits a clean error.
func TestServeBlobSmallMismatchReturnsError(t *testing.T) {
	body := []byte("<!doctype html><html><body>not a blob</body></html>")
	wrong := digest.FromString("a real blob").String()

	server := serveVerified(t, body, wrong)
	defer server.Close()

	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("Docker-Content-Digest"),
		"the digest claim must not survive on an error response")

	got, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.NotContains(t, string(got), "not a blob", "upstream content must not be served")
}

// An unparsable digest is a bad request, not a bad gateway.
func TestVerifyingReaderRejectsInvalidDigestAsBadRequest(t *testing.T) {
	_, err := proxy.NewVerifyingReader(io.NopCloser(bytes.NewReader([]byte("x"))), "not-a-digest")
	require.Error(t, err)
	assert.True(t, errors.IsErr(err, errors.BadRequestCode))
}
