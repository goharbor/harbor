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
	"errors"
	"net/http"
)

// ResponseBuffer is a wrapper for the http.ResponseWriter to buffer the response data
type ResponseBuffer struct {
	w           http.ResponseWriter
	code        int
	header      http.Header
	buffer      bytes.Buffer
	wroteHeader bool
	flushed     bool
}

// NewResponseBuffer creates a ResponseBuffer object
func NewResponseBuffer(w http.ResponseWriter) *ResponseBuffer {
	return &ResponseBuffer{
		w:      w,
		header: http.Header{},
		buffer: bytes.Buffer{},
	}
}

// WriteHeader writes the status code into the buffer without writing to the underlying response writer
func (r *ResponseBuffer) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.code = statusCode
}

// Write writes the data into the buffer without writing to the underlying response writer
func (r *ResponseBuffer) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.buffer.Write(data)
}

// Header returns the header of the buffer
func (r *ResponseBuffer) Header() http.Header {
	return r.header
}

// Flush the status code, header and data into the underlying response writer
func (r *ResponseBuffer) Flush() (int, error) {
	r.flushed = true

	header := r.w.Header()
	for k, vs := range r.header {
		for _, v := range vs {
			header.Add(k, v)
		}
	}
	if r.code > 0 {
		r.w.WriteHeader(r.code)
	}
	return r.w.Write(r.buffer.Bytes())
}

// Success checks whether the status code is >= 200 & <= 399
func (r *ResponseBuffer) Success() bool {
	return r.code >= http.StatusOK && r.code < http.StatusBadRequest
}

// Reset reset the response buffer
func (r *ResponseBuffer) Reset() error {
	if r.flushed {
		return errors.New("response flushed")
	}

	r.code = 0
	r.wroteHeader = false
	r.header = http.Header{}
	r.buffer = bytes.Buffer{}

	return nil
}
