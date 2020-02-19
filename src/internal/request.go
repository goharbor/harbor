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

package internal

import (
	"bytes"
	"io"
	"net/http"
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
		n.Seek(0, io.SeekStart)
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
		nc.Seek(0, io.SeekStart)
		return body
	}

	defer body.Close()

	var buf bytes.Buffer
	io.Copy(&buf, body)

	return nopCloser{bytes.NewReader(buf.Bytes())}
}

// NopCloseRequest ...
func NopCloseRequest(r *http.Request) *http.Request {
	if r != nil && r.Body != nil {
		r.Body = copyBody(r.Body)
	}

	return r
}
