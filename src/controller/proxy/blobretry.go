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
	"time"

	"github.com/opencontainers/go-digest"

	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// maxBlobFetchRetries bounds how many times an interrupted upstream blob
	// fetch is resumed before the read error is returned to the caller.
	maxBlobFetchRetries = 3
)

// blobFetchRetryBackoff is the delay before the first resume attempt; it
// doubles on each subsequent attempt (1s, 2s, 4s by default). It's a var
// rather than a const so tests can shrink it.
var blobFetchRetryBackoff = time.Second

// resumeBlobFunc opens a reader for the remaining bytes of a blob, starting
// at offset. It's used to resume a fetch that broke mid-stream.
type resumeBlobFunc func(offset int64) (io.ReadCloser, error)

// resumingBlobReader wraps an upstream blob body and transparently resumes
// the fetch, via resume, if the connection breaks mid-stream - instead of
// failing the whole blob on a single dropped connection, the way a plain
// docker/OCI client retries an interrupted layer pull. Resuming is bounded by
// maxBlobFetchRetries and disabled when size is unknown (<= 0), since the
// read offset can't be validated against the blob length in that case.
//
// resume itself isn't cancelable once a request is in flight - the
// underlying RemoteInterface doesn't accept a context - but the backoff wait
// between attempts, and any attempt not yet started, respect ctx, so a
// canceled caller (e.g. a disconnected pulling client) stops this reader
// from waiting out further backoffs and opening further upstream requests.
type resumingBlobReader struct {
	ctx       context.Context
	reader    io.ReadCloser
	resume    resumeBlobFunc
	size      int64
	offset    int64
	retries   int
	completed bool
}

// newResumingBlobReader wraps reader so a mid-stream read failure is retried
// up to maxBlobFetchRetries times by resuming the fetch from the last
// successfully read byte offset (via an HTTP range request through resume).
// ctx bounds the retry loop itself (see resumingBlobReader); pass the
// pulling client's request context to give up promptly if it disconnects,
// or a detached context (e.g. context.Background()) for background work
// that should keep retrying regardless.
func newResumingBlobReader(ctx context.Context, reader io.ReadCloser, size int64, resume resumeBlobFunc) io.ReadCloser {
	return &resumingBlobReader{ctx: ctx, reader: reader, size: size, resume: resume}
}

func (r *resumingBlobReader) Read(p []byte) (int, error) {
	if r.completed {
		// A prior call already reached the declared size; self-terminate
		// instead of calling into a reader that may now be stale/exhausted.
		return 0, io.EOF
	}
	for {
		n, err := r.reader.Read(p)
		r.offset += int64(n)

		if r.size > 0 && r.offset >= r.size {
			// We now have the full declared size, regardless of what error
			// (if any) came with these final bytes - an io.Reader may
			// legally pair its last bytes with a non-nil error instead of
			// returning them cleanly and erroring only on a subsequent
			// call. Report success once; further calls are handled by the
			// completed check above rather than re-reading this reader.
			r.completed = true
			return n, nil
		}

		if err == nil {
			return n, nil
		}
		if err == io.EOF { // nolint:errorlint
			if r.size <= 0 {
				// Nothing to validate the EOF against; trust it.
				return n, err
			}
			// The stream ended cleanly but short of the expected size -
			// e.g. a network intermediary that terminates a chunked
			// response gracefully instead of resetting the connection.
			// Treat it the same as a dropped connection so it consumes
			// the retry budget instead of being accepted as complete.
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
		// n == 0: loop and read immediately from the freshly resumed reader.
	}
}

// reconnect closes the broken reader and resumes the fetch from the current
// offset, retrying the resume itself (bounded by the same retry budget) if
// establishing the new connection also fails. It gives up early, without
// waiting out the backoff or starting another attempt, if ctx is canceled.
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
			r.reader = next
			return nil
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

func (r *resumingBlobReader) Close() error {
	return r.reader.Close()
}

// isRetryableBlobReadErr reports whether err looks like a transient failure
// of the upstream connection - dropped, reset, or timed out mid-transfer -
// worth retrying, as opposed to the pulling client itself having given up.
func isRetryableBlobReadErr(err error) bool {
	return err != nil && !errors.Is(err, context.Canceled)
}

// verifyingReadCloser wraps a blob reader and verifies the bytes read so far
// match the expected digest as soon as the declared size has been read - not
// only when the wrapped reader happens to signal io.EOF on some later call.
// A consumer with its own framing (e.g. an HTTP request body capped by
// Content-Length) may never issue that extra call, since it can determine
// on its own that it has all the bytes it declared and stop reading. A
// mismatch is surfaced as a Read error, so a corrupted or truncated blob
// that still parses as a well-formed HTTP response is never mistaken for a
// complete, valid one by the caller (e.g. the local registry committing a
// blob push).
type verifyingReadCloser struct {
	io.ReadCloser
	verifier digest.Verifier
	dig      digest.Digest
	size     int64
	read     int64
	verified bool
}

// newVerifyingReadCloser wraps reader so that, once size bytes have been
// read (or the stream ends, if size is unknown), content not matching dig
// surfaces as a Read error.
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
