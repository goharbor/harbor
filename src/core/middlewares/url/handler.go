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

package url

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	"net/http"
	"strings"
)

type urlHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &urlHandler{
		next: next,
	}
}

// ServeHTTP ...
func (uh urlHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Debugf("in url handler, path: %s", req.URL.Path)
	flag, repository, reference := util.MatchPullManifest(req)
	if flag {
		components := strings.SplitN(repository, "/", 2)
		if len(components) < 2 {
			http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("Bad repository name: %s", repository)), http.StatusBadRequest)
			return
		}

		client, err := coreutils.NewRepositoryClientForUI(util.TokenUsername, repository)
		if err != nil {
			log.Errorf("Error creating repository Client: %v", err)
			http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}
		digest, _, err := client.ManifestExist(reference)
		if err != nil {
			log.Errorf("Failed to get digest for reference: %s, error: %v", reference, err)
			http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}

		img := util.ImageInfo{
			Repository:  repository,
			Reference:   reference,
			ProjectName: components[0],
			Digest:      digest,
		}

		log.Debugf("image info of the request: %#v", img)
		ctx := context.WithValue(req.Context(), util.ImageInfoCtxKey, img)
		req = req.WithContext(ctx)
	}
	uh.next.ServeHTTP(rw, req)
}
