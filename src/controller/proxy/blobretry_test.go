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
	"strings"
	"testing"
	"time"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libErrors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/adapter"
	testproxy "github.com/goharbor/harbor/src/testing/controller/proxy"
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

// oneShotReader returns data together with err in a single Read call, which
// an io.Reader may do, where stubReader returns them in separate calls.
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

func TestResumingBlobReader_GivesUpImmediatelyOnRateLimitedResume(t *testing.T) {
	withShortBlobRetryBackoff(t)

	first := &stubReader{err: io.ErrUnexpectedEOF}
	resumeAttempts := 0
	rateLimitErr := libErrors.New("too many requests").WithCode(libErrors.RateLimitCode)
	reader := newResumingBlobReader(context.Background(), first, 100, func(offset int64) (io.ReadCloser, error) {
		resumeAttempts++
		return nil, rateLimitErr
	})

	_, err := io.ReadAll(reader)
	require.Error(t, err)
	assert.Equal(t, 1, resumeAttempts, "a rate-limited resume shouldn't be retried - a few seconds of backoff won't help and it just adds more load")
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

	// first ends in a clean io.EOF rather than a transport error.
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
	// The error arrives with the bytes that complete the blob, so there's
	// nothing left to retry.
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

func TestResumingBlobReader_CloseDuringReadStopsRetrying(t *testing.T) {
	withShortBlobRetryBackoff(t)

	// A pipe's Read blocks until the pipe is closed and then fails, like an
	// upstream body whose consumer gives up mid-read: net/http closes a
	// request body from another goroutine when the request fails early.
	upstream, _ := io.Pipe()
	resumeAttempts := 0
	reader := newResumingBlobReader(context.Background(), upstream, 100, func(int64) (io.ReadCloser, error) {
		resumeAttempts++
		return &stubReader{}, nil
	})

	readErr := make(chan error, 1)
	go func() {
		_, err := reader.Read(make([]byte, 8))
		readErr <- err
	}()
	require.NoError(t, reader.Close())

	require.Error(t, <-readErr)
	assert.Equal(t, 0, resumeAttempts, "a read interrupted by Close isn't an upstream failure and must not be resumed")
}

func TestResumingBlobReader_ClosesReaderResumedAfterClose(t *testing.T) {
	withShortBlobRetryBackoff(t)

	// Close lands while the resume request is still in flight, so the reader
	// that request eventually returns has no one left to close it.
	late := &stubReader{data: []byte("late")}
	resuming, release := make(chan struct{}), make(chan struct{})
	reader := newResumingBlobReader(context.Background(), &stubReader{err: io.ErrUnexpectedEOF}, 100, func(int64) (io.ReadCloser, error) {
		close(resuming)
		<-release
		return late, nil
	})

	readErr := make(chan error, 1)
	go func() {
		_, err := reader.Read(make([]byte, 8))
		readErr <- err
	}()
	<-resuming
	require.NoError(t, reader.Close())
	close(release)

	require.Error(t, <-readErr)
	assert.True(t, late.closed, "a reader resumed after Close must be closed, not installed")
}

func TestOpenCacheBlobReader_RejectsUnverifiableDigest(t *testing.T) {
	// The mock has no expectations: these must be rejected before any upstream
	// request, and without digest.Verifier panicking on them.
	remote := &testproxy.RemoteInterface{}
	for _, dig := range []digest.Digest{"md5:d41d8cd98f00b204e9800998ecf8427e", "sha256:tooshort", "no-separator"} {
		_, err := openCacheBlobReader(remote, "library/hello-world", distribution.Descriptor{Digest: dig, Size: 1})
		assert.Error(t, err, dig)
	}
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
	// bytes.Reader returns its last bytes with a nil error and io.EOF only on
	// the next call, which a consumer that knows the size may never make.
	content := []byte("AAAAAAAAAA")
	wrongDigest := digest.FromBytes([]byte("BBBBBBBBBB"))

	reader := newVerifyingReadCloser(io.NopCloser(bytes.NewReader(content)), wrongDigest, int64(len(content)))
	buf := make([]byte, len(content))
	n, err := reader.Read(buf)
	require.Error(t, err, "the read call that reaches the declared size must itself surface the mismatch")
	assert.Equal(t, len(content), n)
}

// chunkRegistry is an ArtifactRegistry that can only serve ranges of blob.
type chunkRegistry struct {
	adapter.ArtifactRegistry
	blob string
}

func (c *chunkRegistry) PullBlobChunk(_, _ string, _, start, end int64) (int64, io.ReadCloser, error) {
	if end >= int64(len(c.blob)) {
		return 0, nil, errors.New("range not satisfiable")
	}
	return end - start + 1, io.NopCloser(strings.NewReader(c.blob[start : end+1])), nil
}

func TestRemoteHelper_BlobReaderAt(t *testing.T) {
	for _, speed := range []int32{0, 1024} { // unthrottled and throttled
		r := &remoteHelper{registry: &chunkRegistry{blob: "hello world"}, opts: NewOptions(WithSpeed(speed))}

		reader, err := r.BlobReaderAt("library/hello-world", "digest", 11, 6)
		require.NoError(t, err)
		got, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, "world", string(got), "everything from the offset through the last byte")

		_, err = r.BlobReaderAt("library/hello-world", "digest", 12, 6)
		assert.Error(t, err)
	}
}
