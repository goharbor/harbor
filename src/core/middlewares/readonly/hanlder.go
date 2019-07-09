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

package readonly

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
)

type readonlyHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &readonlyHandler{
		next: next,
	}
}

// ServeHTTP ...
func (rh readonlyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if config.ReadOnly() {
		if req.Method == http.MethodDelete || req.Method == http.MethodPost || req.Method == http.MethodPatch || req.Method == http.MethodPut {
			log.Warningf("The request is prohibited in readonly mode, url is: %s", req.URL.Path)
			http.Error(rw, util.MarshalError("DENIED", "The system is in read only mode. Any modification is prohibited."), http.StatusForbidden)
			return
		}
	}
	rh.next.ServeHTTP(rw, req)
}
