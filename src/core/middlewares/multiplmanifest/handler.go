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

package multiplmanifest

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
	"strings"
)

type multipleManifestHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &multipleManifestHandler{
		next: next,
	}
}

// ServeHTTP The handler is responsible for blocking request to upload manifest list by docker client, which is not supported so far by Harbor.
func (mh multipleManifestHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, _, _ := util.MatchPushManifest(req)
	if match {
		contentType := req.Header.Get("Content-type")
		// application/vnd.docker.distribution.manifest.list.v2+json
		if strings.Contains(contentType, "manifest.list.v2") {
			log.Debugf("Content-type: %s is not supported, failing the response.", contentType)
			http.Error(rw, util.MarshalError("UNSUPPORTED_MEDIA_TYPE", "Manifest.list is not supported."), http.StatusUnsupportedMediaType)
			return
		}
	}
	mh.next.ServeHTTP(rw, req)
}
