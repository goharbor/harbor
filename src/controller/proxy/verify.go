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
	"io"

	"github.com/opencontainers/go-digest"

	"github.com/goharbor/harbor/src/lib/errors"
)

// verifyTailSize is the number of trailing bytes withheld until the digest is confirmed.
const verifyTailSize = 32 * 1024

// isEmptyBlobDigest reports whether raw addresses a zero-length blob. Parse rejects
// an unavailable algorithm, so FromBytes cannot panic here.
func isEmptyBlobDigest(raw string) bool {
	dgst, err := digest.Parse(raw)
	if err != nil {
		return false
	}

	return dgst == dgst.Algorithm().FromBytes(nil)
}

// verifyingReader hashes a blob as it streams, withholding the final verifyTailSize
// bytes until the digest is confirmed: the response commits its status and
// Content-Length before the body is written, so verifying only at EOF would leave a
// complete, cacheable 200 carrying the wrong content.
type verifyingReader struct {
	src      io.ReadCloser
	verifier digest.Verifier
	dgst     string
	buf      bytes.Buffer
	scratch  []byte
	eof      bool
	checked  bool
	err      error
}

// NewVerifyingReader wraps src so its bytes are verified against dgst as they are read.
func NewVerifyingReader(src io.ReadCloser, dgst string) (io.ReadCloser, error) {
	parsed, err := digest.Parse(dgst)
	if err != nil {
		return nil, errors.New(err).WithCode(errors.BadRequestCode).
			WithMessagef("invalid digest %s", dgst)
	}

	return &verifyingReader{
		src:      src,
		verifier: parsed.Verifier(),
		dgst:     dgst,
		// allocated once, a per-read allocation would put the whole blob through the GC
		scratch: make([]byte, 32*1024),
	}, nil
}

func (r *verifyingReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	// keep the tail buffered so it can still be withheld when EOF arrives
	for !r.eof && r.buf.Len() <= verifyTailSize {
		n, err := r.src.Read(r.scratch)
		if n > 0 {
			if _, werr := r.verifier.Write(r.scratch[:n]); werr != nil {
				r.err = werr
				return 0, r.err
			}
			r.buf.Write(r.scratch[:n])
		}
		if err == io.EOF {
			r.eof = true
			break
		}
		if err != nil {
			r.err = err
			return 0, r.err
		}
	}

	if r.eof {
		if !r.checked {
			if !r.verifier.Verified() {
				r.err = errors.New(nil).WithCode(errors.BadGatewayCode).
					WithMessagef("upstream content does not match digest %s", r.dgst)
				return 0, r.err
			}
			r.checked = true
		}
		// digest confirmed, release the withheld tail
		return r.buf.Read(p)
	}

	// the loop above guarantees buf holds more than verifyTailSize bytes
	releasable := r.buf.Len() - verifyTailSize
	if len(p) > releasable {
		p = p[:releasable]
	}

	return r.buf.Read(p)
}

func (r *verifyingReader) Close() error {
	return r.src.Close()
}
