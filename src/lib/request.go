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

package lib

import (
	"bytes"
	"io"
	"net/http"

	"github.com/goharbor/harbor/src/lib/errors"
)

// nopCloser is just like ioutil's, but here to let us re-read the same
// buffer inside by moving position to the start every time we done with reading
type nopCloser struct {
	io.ReadSeeker
}

// Read just a wrapper around real Read which also moves position to the start if we get EOF
// to have it ready for next read-cycle
func (n nopCloser) Read(p []byte) (int, error) {
	num, err := n.ReadSeeker.Read(p)
	if err == io.EOF { // move to start to have it ready for next read cycle
		_, _ = n.Seek(0, io.SeekStart)
	}
	return num, err
}

// Close is a no-op Close
func (n nopCloser) Close() error {
	return nil
}

func copyBody(body io.ReadCloser) io.ReadCloser {
	// check if body was already read and converted into our nopCloser
	if nc, ok := body.(nopCloser); ok {
		_, _ = nc.Seek(0, io.SeekStart)
		return body
	}

	defer body.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, body)

	return nopCloser{bytes.NewReader(buf.Bytes())}
}

// NopCloseRequest makes r.Body re-readable so it can be consumed more than once.
func NopCloseRequest(r *http.Request) *http.Request {
	if r != nil && r.Body != nil {
		r.Body = copyBody(r.Body)
	}

	return r
}

// ReadRequestBody reads and returns r.Body while enforcing an optional size
// limit. When limit > 0 and the body is larger than limit, a
// RequestEntityTooLargeError is returned so callers surface a 413 instead of
// buffering or parsing a truncated body. On success r.Body is replaced with a
// re-readable buffer holding the same (<= limit) bytes for any downstream
// consumer.
func ReadRequestBody(r *http.Request, limit int64) ([]byte, error) {
	if r == nil || r.Body == nil {
		return nil, nil
	}

	// keep a reference to the original body so we can close it safely once we
	// have buffered its contents (or bailed out on an over-limit body)
	originalBody := r.Body
	defer originalBody.Close()

	var reader io.Reader = originalBody
	if limit > 0 {
		// read one extra byte so an over-limit body is detected rather than
		// silently truncated
		reader = io.LimitReader(originalBody, limit+1)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if limit > 0 && int64(len(data)) > limit {
		return nil, errors.RequestEntityTooLargeError(nil)
	}

	r.Body = nopCloser{bytes.NewReader(data)}
	return data, nil
}
