// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gorilla/mux"
)

// RequestHandlerMapping is a mapping between request and its handler
type RequestHandlerMapping struct {
	// Method is the method the request used
	Method string
	// Pattern is the pattern the request must match
	Pattern string
	// Handler is the handler which handles the request
	Handler func(http.ResponseWriter, *http.Request)
}

// ServeHTTP ...
func (rhm *RequestHandlerMapping) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(rhm.Method) != 0 && r.Method != strings.ToUpper(rhm.Method) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rhm.Handler(w, r)
}

// Response is a response used for unit test
type Response struct {
	// StatusCode is the status code of the response
	StatusCode int
	// Headers are the headers of the response
	Headers map[string]string
	// Boby is the body of the response
	Body []byte
}

// Handler returns a handler function which handle requst according to
// the response provided
func Handler(resp *Response) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if resp == nil {
			return
		}

		for k, v := range resp.Headers {
			w.Header().Add(http.CanonicalHeaderKey(k), v)
		}

		if resp.StatusCode == 0 {
			resp.StatusCode = http.StatusOK
		}
		w.WriteHeader(resp.StatusCode)

		if len(resp.Body) != 0 {
			io.Copy(w, bytes.NewReader(resp.Body))
		}
	}
}

// NewServer creates a HTTP server for unit test
func NewServer(mappings ...*RequestHandlerMapping) *httptest.Server {
	r := mux.NewRouter()

	for _, mapping := range mappings {
		r.PathPrefix(mapping.Pattern).Handler(mapping).Methods(mapping.Method)
	}

	return httptest.NewServer(r)
}
