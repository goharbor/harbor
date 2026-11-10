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
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"

	libErrors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

// maxBlobFetchRetries bounds how often an interrupted upstream blob fetch is
// resumed before the read error is returned.
const maxBlobFetchRetries = 3

// blobFetchRetryBackoff is the wait before the first resume; it doubles on
// each further attempt. It's a var so tests can shrink it.
var blobFetchRetryBackoff = time.Second

// resumeBlobFunc opens a reader for the rest of a blob, starting at offset.
type resumeBlobFunc func(offset int64) (io.ReadCloser, error)

// resumingBlobReader wraps an upstream blob body and, when a read fails
// mid-stream, resumes the fetch from the current offset instead of failing
// the blob. It needs the blob size to tell a complete stream from a truncated
// one, so nothing is resumed when size is unknown (<= 0).
//
// Canceling ctx or calling Close ends the retry loop, though neither can
// abort a resume request already in flight: RemoteInterface takes no context.
// Close may be called while a Read is in flight, as net/http does to a
// request body when the request fails early.
type resumingBlobReader struct {
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.Mutex // guards reader: Close may race a resume
	reader    io.ReadCloser
	resume    resumeBlobFunc
	size      int64
	offset    int64
	retries   int
	completed bool
}

// newResumingBlobReader wraps reader, the body of a blob of the given size.
// Pass the pulling client's request context to stop retrying once it
// disconnects, or context.Background() for work that outlives the request.
func newResumingBlobReader(ctx context.Context, reader io.ReadCloser, size int64, resume resumeBlobFunc) io.ReadCloser {
	ctx, cancel := context.WithCancel(ctx)
	return &resumingBlobReader{ctx: ctx, cancel: cancel, reader: reader, size: size, resume: resume}
}

func (r *resumingBlobReader) Read(p []byte) (int, error) {
	if r.completed {
		return 0, io.EOF
	}
	for {
		n, err := r.reader.Read(p)
		r.offset += int64(n)

		if r.size > 0 && r.offset >= r.size {
			// The blob is complete, whatever error came with its final
			// bytes. Later calls return EOF without reading again.
			r.completed = true
			return n, nil
		}

		if err == nil {
			return n, nil
		}
		if err == io.EOF { // nolint:errorlint
			if r.size <= 0 {
				return n, err
			}
			// A clean EOF short of size is a truncated stream, e.g. an
			// intermediary ending a chunked response gracefully.
			err = io.ErrUnexpectedEOF
		}
		if !r.canRetry(err) {
			return n, err
		}
		if rErr := r.reconnect(err); rErr != nil {
			return n, rErr
		}
		if n > 0 {
			return n, nil
		}
	}
}

// reconnect closes the broken reader and resumes the fetch at the current
// offset. It backs off before every attempt, and a failed attempt is retried
// within the same budget.
func (r *resumingBlobReader) reconnect(cause error) error {
	_ = r.reader.Close()
	for {
		if ctxErr := r.ctx.Err(); ctxErr != nil {
			return fmt.Errorf("blob fetch canceled while resuming after %v: %w", cause, ctxErr)
		}
		r.retries++
		backoff := blobFetchRetryBackoff * time.Duration(uint64(1)<<uint(r.retries-1))
		log.Warningf("proxy cache: upstream blob read interrupted at byte %d/%d (%v), resuming attempt %d/%d in %s",
			r.offset, r.size, cause, r.retries, maxBlobFetchRetries, backoff)
		timer := time.NewTimer(backoff)
		select {
		case <-r.ctx.Done():
			timer.Stop()
			return fmt.Errorf("blob fetch canceled while waiting to resume after %v: %w", cause, r.ctx.Err())
		case <-timer.C:
		}
		next, err := r.resume(r.offset)
		if err == nil {
			r.mu.Lock()
			defer r.mu.Unlock()
			if ctxErr := r.ctx.Err(); ctxErr != nil {
				// Closed or canceled while the request was in flight; nobody
				// is left to close next.
				_ = next.Close()
				return fmt.Errorf("blob fetch canceled while resuming after %v: %w", cause, ctxErr)
			}
			r.reader = next
			return nil
		}
		if libErrors.IsRateLimitError(err) {
			// A few seconds of backoff won't clear a rate limit, and more
			// requests only add load to an upstream that's throttling.
			return fmt.Errorf("upstream rate-limited the resume attempt after %v: %w", cause, err)
		}
		if r.retries >= maxBlobFetchRetries {
			return fmt.Errorf("failed to resume interrupted blob fetch after %v: %w", cause, err)
		}
		cause = err
	}
}

func (r *resumingBlobReader) canRetry(err error) bool {
	return r.size > 0 && r.retries < maxBlobFetchRetries && r.offset < r.size && isRetryableBlobReadErr(err)
}

// Close cancels any pending or future resume before closing the current
// reader, so the Read it interrupts isn't mistaken for an upstream failure.
func (r *resumingBlobReader) Close() error {
	r.cancel()
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.reader.Close()
}

// isRetryableBlobReadErr reports whether a read error is worth a resume:
// anything but the request itself having been canceled.
func isRetryableBlobReadErr(err error) bool {
	return err != nil && !errors.Is(err, context.Canceled)
}

// openCacheBlobReader opens the upstream blob desc for copying into the local
// cache: interrupted reads are resumed, and the content is verified against
// desc.Digest so a corrupted blob fails the push instead of being committed.
// The copy is detached from any pulling client, hence its own context.
func openCacheBlobReader(r RemoteInterface, remoteRepo string, desc distribution.Descriptor) (io.ReadCloser, error) {
	// Digest.Verifier panics on a digest it can't hash, and the callers run in
	// goroutines nothing recovers from, so reject those up front.
	if err := desc.Digest.Validate(); err != nil {
		return nil, fmt.Errorf("unable to verify blob %q: %w", desc.Digest, err)
	}
	_, bReader, err := r.BlobReader(remoteRepo, string(desc.Digest))
	if err != nil {
		return nil, err
	}
	bReader = newResumingBlobReader(context.Background(), bReader, desc.Size, func(offset int64) (io.ReadCloser, error) {
		return r.BlobReaderAt(remoteRepo, string(desc.Digest), desc.Size, offset)
	})
	return newVerifyingReadCloser(bReader, desc.Digest, desc.Size), nil
}

// verifyingReadCloser fails the Read that completes a blob - size bytes
// read, or EOF when size is unknown - if the content doesn't match dig, so a
// consumer never takes a corrupted blob for a successful read.
type verifyingReadCloser struct {
	io.ReadCloser
	verifier digest.Verifier
	dig      digest.Digest
	size     int64
	read     int64
	verified bool
}

// newVerifyingReadCloser wraps reader. dig must be valid: Digest.Verifier
// panics otherwise.
func newVerifyingReadCloser(reader io.ReadCloser, dig digest.Digest, size int64) io.ReadCloser {
	return &verifyingReadCloser{ReadCloser: reader, verifier: dig.Verifier(), dig: dig, size: size}
}

func (v *verifyingReadCloser) Read(p []byte) (int, error) {
	n, err := v.ReadCloser.Read(p)
	if n > 0 {
		_, _ = v.verifier.Write(p[:n])
		v.read += int64(n)
	}
	if !v.verified && ((v.size > 0 && v.read >= v.size) || err == io.EOF) { // nolint:errorlint
		v.verified = true
		if !v.verifier.Verified() {
			return n, fmt.Errorf("blob content does not match expected digest %s after fetching from upstream", v.dig)
		}
	}
	return n, err
}
