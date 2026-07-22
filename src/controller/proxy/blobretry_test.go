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
	"errors"
	"io"
	"testing"
	"time"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubReader serves data and then, once exhausted, returns err (or io.EOF if
// err is nil). It records whether it was closed.
type stubReader struct {
	data   []byte
	err    error
	closed bool
}

func (s *stubReader) Read(p []byte) (int, error) {
	if len(s.data) > 0 {
		n := copy(p, s.data)
		s.data = s.data[n:]
		return n, nil
	}
	if s.err != nil {
		return 0, s.err
	}
	return 0, io.EOF
}

func (s *stubReader) Close() error {
	s.closed = true
	return nil
}

// oneShotReader returns all of data together with err in a single Read call,
// unlike stubReader, which always separates the final data from a
// subsequent terminal error/EOF call - both are legal io.Reader behavior.
type oneShotReader struct {
	data []byte
	err  error
}

func (o *oneShotReader) Read(p []byte) (int, error) {
	n := copy(p, o.data)
	o.data = o.data[n:]
	return n, o.err
}

func (o *oneShotReader) Close() error { return nil }

func withShortBlobRetryBackoff(t *testing.T) {
	orig := blobFetchRetryBackoff
	blobFetchRetryBackoff = time.Millisecond
	t.Cleanup(func() { blobFetchRetryBackoff = orig })
}

func TestResumingBlobReader_ResumesAfterMidStreamError(t *testing.T) {
	withShortBlobRetryBackoff(t)

	first := &stubReader{data: []byte("hello "), err: io.ErrUnexpectedEOF}
	second := &stubReader{data: []byte("world")}

	var resumeOffsets []int64
	reader := newResumingBlobReader(context.Background(), first, 11, func(offset int64) (io.ReadCloser, error) {
		resumeOffsets = append(resumeOffsets, offset)
		return second, nil
	})

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(got))
	assert.Equal(t, []int64{6}, resumeOffsets)
	assert.True(t, first.closed, "the broken reader should be closed before resuming")
}

func TestResumingBlobReader_GivesUpAfterMaxRetries(t *testing.T) {
	withShortBlobRetryBackoff(t)

	first := &stubReader{err: io.ErrUnexpectedEOF}
	resumeAttempts := 0
	reader := newResumingBlobReader(context.Background(), first, 100, func(offset int64) (io.ReadCloser, error) {
		resumeAttempts++
		return nil, errors.New("upstream still unreachable")
	})

	_, err := io.ReadAll(reader)
	require.Error(t, err)
	assert.Equal(t, maxBlobFetchRetries, resumeAttempts)
}

func TestResumingBlobReader_DoesNotRetryOnContextCanceled(t *testing.T) {
	withShortBlobRetryBackoff(t)

	first := &stubReader{err: context.Canceled}
	resumeAttempts := 0
	reader := newResumingBlobReader(context.Background(), first, 100, func(offset int64) (io.ReadCloser, error) {
		resumeAttempts++
		return &stubReader{}, nil
	})

	_, err := io.ReadAll(reader)
	require.Error(t, err)
	assert.Equal(t, 0, resumeAttempts, "a canceled request should not be retried")
}

func TestResumingBlobReader_StopsRetryingWhenContextCanceled(t *testing.T) {
	withShortBlobRetryBackoff(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled before the reader ever attempts to resume

	first := &stubReader{err: io.ErrUnexpectedEOF}
	resumeAttempts := 0
	reader := newResumingBlobReader(ctx, first, 100, func(offset int64) (io.ReadCloser, error) {
		resumeAttempts++
		return &stubReader{}, nil
	})

	_, err := io.ReadAll(reader)
	require.Error(t, err)
	assert.Equal(t, 0, resumeAttempts, "a canceled context should stop retrying before opening another upstream request")
}

func TestResumingBlobReader_DoesNotRetryWithUnknownSize(t *testing.T) {
	withShortBlobRetryBackoff(t)

	first := &stubReader{err: io.ErrUnexpectedEOF}
	resumeAttempts := 0
	reader := newResumingBlobReader(context.Background(), first, 0, func(offset int64) (io.ReadCloser, error) {
		resumeAttempts++
		return &stubReader{}, nil
	})

	_, err := io.ReadAll(reader)
	require.Error(t, err)
	assert.Equal(t, 0, resumeAttempts, "resuming an unknown-length blob can't be validated, so it shouldn't be attempted")
}

func TestResumingBlobReader_RecoversFromMultipleInterruptions(t *testing.T) {
	withShortBlobRetryBackoff(t)

	first := &stubReader{data: []byte("aaa"), err: io.ErrUnexpectedEOF}
	second := &stubReader{data: []byte("bbb"), err: io.ErrUnexpectedEOF}
	third := &stubReader{data: []byte("ccc")}
	var resumeOffsets []int64
	readers := []*stubReader{second, third}
	reader := newResumingBlobReader(context.Background(), first, 9, func(offset int64) (io.ReadCloser, error) {
		resumeOffsets = append(resumeOffsets, offset)
		next := readers[0]
		readers = readers[1:]
		return next, nil
	})

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "aaabbbccc", string(got))
	assert.Equal(t, []int64{3, 6}, resumeOffsets)
}

func TestResumingBlobReader_RetriesOnPrematureCleanEOF(t *testing.T) {
	withShortBlobRetryBackoff(t)

	// first ends with a clean nil error (io.EOF on the next call), not a
	// transport error - e.g. a network intermediary that terminates a
	// chunked response gracefully instead of resetting the connection.
	first := &stubReader{data: []byte("hello ")}
	second := &stubReader{data: []byte("world")}

	var resumeOffsets []int64
	reader := newResumingBlobReader(context.Background(), first, 11, func(offset int64) (io.ReadCloser, error) {
		resumeOffsets = append(resumeOffsets, offset)
		return second, nil
	})

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "hello world", string(got))
	assert.Equal(t, []int64{6}, resumeOffsets, "a clean EOF short of the declared size must still trigger a resume")
}

func TestResumingBlobReader_TreatsErrorAccompanyingFinalBytesAsSuccess(t *testing.T) {
	// An io.Reader may legally return its final bytes together with a
	// non-nil, non-EOF error in the same call, instead of returning them
	// cleanly and erroring only on a subsequent call. If those bytes bring
	// the reader to the full declared size, that's a complete blob and must
	// not be treated as a failed/retryable read.
	content := []byte("complete blob")
	reader := newResumingBlobReader(context.Background(), &oneShotReader{data: content, err: io.ErrUnexpectedEOF}, int64(len(content)),
		func(offset int64) (io.ReadCloser, error) {
			t.Fatal("resume should not be called when the blob was already fully read")
			return nil, nil
		})

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestVerifyingReadCloser_PassesThroughValidContent(t *testing.T) {
	content := []byte("valid blob content")
	dig := digest.FromBytes(content)

	reader := newVerifyingReadCloser(io.NopCloser(&stubReader{data: content}), dig, int64(len(content)))
	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestVerifyingReadCloser_RejectsMismatchedContent(t *testing.T) {
	content := []byte("corrupted blob content")
	wrongDigest := digest.FromBytes([]byte("something else entirely"))

	reader := newVerifyingReadCloser(io.NopCloser(&stubReader{data: content}), wrongDigest, int64(len(content)))
	_, err := io.ReadAll(reader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not match expected digest")
}

func TestVerifyingReadCloser_VerifiesAsSoonAsSizeReachedWithoutFurtherEOFCall(t *testing.T) {
	// bytes.Reader returns the final chunk with a nil error and only signals
	// io.EOF on a subsequent call - the same pattern net/http's
	// io.LimitReader-wrapped request body uses. A consumer capped by
	// Content-Length (e.g. the real blob upload) may never make that
	// subsequent call, so verification must not depend on it.
	content := []byte("AAAAAAAAAA")
	wrongDigest := digest.FromBytes([]byte("BBBBBBBBBB"))

	reader := newVerifyingReadCloser(io.NopCloser(bytes.NewReader(content)), wrongDigest, int64(len(content)))
	buf := make([]byte, len(content))
	n, err := reader.Read(buf)
	require.Error(t, err, "the read call that reaches the declared size must itself surface the mismatch")
	assert.Equal(t, len(content), n)
}
