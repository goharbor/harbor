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

package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"runtime"
)

func currPath() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	return path.Dir(f)
}

// NewNotaryServer creates a notary server for testing.
func NewNotaryServer(endpoint string) *httptest.Server {
	mux := http.NewServeMux()
	validRoot := fmt.Sprintf("/v2/%s/library/busybox/_trust/tuf/", endpoint)
	invalidRoot := fmt.Sprintf("/v2/%s/library/busybox/fail/_trust/tuf/", endpoint)
	p := currPath()
	mux.Handle(validRoot, http.StripPrefix(validRoot, http.FileServer(http.Dir(path.Join(p, "valid")))))
	mux.Handle(invalidRoot, http.StripPrefix(invalidRoot, http.FileServer(http.Dir(path.Join(p, "invalid")))))
	return httptest.NewServer(mux)
}
