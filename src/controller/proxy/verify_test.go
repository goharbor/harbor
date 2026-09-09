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
	"crypto/rand"
	"io"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
)

func randomBlob(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := rand.Read(b)
	require.NoError(t, err)

	return b
}

func TestVerifyingReaderMatchingDigest(t *testing.T) {
	// sizes either side of the withheld tail exercise both the buffered and streaming paths
	for _, size := range []int{0, 1, verifyTailSize - 1, verifyTailSize, verifyTailSize + 1, 5 * verifyTailSize} {
		blob := randomBlob(t, size)
		dgst := digest.FromBytes(blob).String()

		r, err := NewVerifyingReader(io.NopCloser(bytes.NewReader(blob)), dgst)
		require.NoError(t, err)

		got, err := io.ReadAll(r)
		require.NoError(t, err, "size %d should verify", size)
		assert.Equal(t, blob, got, "size %d should stream through unchanged", size)
		require.NoError(t, r.Close())
	}
}

// Content not matching the requested digest must not be delivered in full.
func TestVerifyingReaderMismatchIsNotDeliveredInFull(t *testing.T) {
	// large enough that most of it streams before the mismatch is detected
	blob := randomBlob(t, 5*verifyTailSize)
	wrongDigest := digest.FromString("something else entirely").String()

	r, err := NewVerifyingReader(io.NopCloser(bytes.NewReader(blob)), wrongDigest)
	require.NoError(t, err)

	got, err := io.ReadAll(r)
	require.Error(t, err, "a digest mismatch must surface as a read error")
	assert.True(t, errors.IsErr(err, errors.BadGatewayCode))

	assert.Less(t, len(got), len(blob))
	assert.LessOrEqual(t, len(blob)-len(got), verifyTailSize+1,
		"only the withheld tail should be missing")
}

// A body fitting entirely in the withheld tail must not reach the client at all.
func TestVerifyingReaderSmallMismatchYieldsNothing(t *testing.T) {
	page := []byte("<!doctype html><html><body>not a blob</body></html>")
	wrongDigest := digest.FromString("expected blob").String()

	r, err := NewVerifyingReader(io.NopCloser(bytes.NewReader(page)), wrongDigest)
	require.NoError(t, err)

	got, err := io.ReadAll(r)
	require.Error(t, err)
	assert.True(t, errors.IsErr(err, errors.BadGatewayCode))
	assert.Empty(t, got, "nothing should be written before the mismatch is caught")
}

func TestVerifyingReaderInvalidDigest(t *testing.T) {
	_, err := NewVerifyingReader(io.NopCloser(bytes.NewReader(nil)), "not-a-digest")
	require.Error(t, err)
	assert.True(t, errors.IsErr(err, errors.BadRequestCode))
}

// io.CopyN(w, reader, 0) never calls Read, so only the empty blob's own digest may
// be served at zero length, in whichever algorithm addressed it.
func TestIsEmptyBlobDigest(t *testing.T) {
	for _, algo := range []digest.Algorithm{digest.SHA256, digest.SHA384, digest.SHA512} {
		if !algo.Available() {
			continue
		}
		empty := algo.FromBytes(nil).String()
		assert.True(t, isEmptyBlobDigest(empty), "empty blob in %s must be recognised", algo)
		assert.False(t, isEmptyBlobDigest(algo.FromBytes([]byte("a real blob")).String()),
			"non-empty blob in %s must not be treated as empty", algo)

		// The zero-length read that the caller would perform verifies cleanly.
		r, err := NewVerifyingReader(io.NopCloser(bytes.NewReader(nil)), empty)
		require.NoError(t, err)
		got, err := io.ReadAll(r)
		require.NoError(t, err)
		assert.Empty(t, got)
	}

	assert.False(t, isEmptyBlobDigest("not-a-digest"))
	assert.False(t, isEmptyBlobDigest(""))
}
