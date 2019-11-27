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

package util

import (
	"net/http"
)

// CustomResponseWriter write the response code into the status
type CustomResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

// NewCustomResponseWriter ...
func NewCustomResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{ResponseWriter: w}
}

// Status ...
func (w *CustomResponseWriter) Status() int {
	return w.status
}

// Header ...
func (w CustomResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Write ...
func (w *CustomResponseWriter) Write(p []byte) (n int, err error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

// WriteHeader ...
func (w *CustomResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
}
