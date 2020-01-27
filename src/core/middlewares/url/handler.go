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
	"github.com/opencontainers/go-digest"
	"net/http"
	"regexp"
	"strings"
)

var (
	urlPatterns = []*regexp.Regexp{
		util.ManifestURLRe, util.TagListURLRe, util.BlobURLRe, util.BlobUploadURLRe,
	}
)

// urlHandler extracts the artifact info from the url of request to V2 handler and propagates it to context
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
	path := req.URL.Path
	log.Debugf("in url handler, path: %s", path)
	m, ok := parse(path)
	if !ok {
		uh.next.ServeHTTP(rw, req)
	}
	repo := m[util.RepositorySubexp]
	components := strings.SplitN(repo, "/", 2)
	if len(components) < 2 {
		http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("Bad repository name: %s", repo)), http.StatusBadRequest)
		return
	}
	art := util.ArtifactInfo{
		Repository:  repo,
		ProjectName: components[0],
	}
	if digest, ok := m[util.DigestSubexp]; ok {
		art.Digest = digest
	}
	if ref, ok := m[util.ReferenceSubexp]; ok {
		art.Reference = ref
	}

	if util.ManifestURLRe.MatchString(path) && req.Method == http.MethodGet { // Request for pulling manifest
		client, err := coreutils.NewRepositoryClientForUI(util.TokenUsername, art.Repository)
		if err != nil {
			log.Errorf("Error creating repository Client: %v", err)
			http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}
		digest, _, err := client.ManifestExist(art.Reference)
		if err != nil {
			log.Errorf("Failed to get digest for reference: %s, error: %v", art.Reference, err)
			http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}

		art.Digest = digest
		log.Debugf("artifact info of the request: %#v", art)
		ctx := context.WithValue(req.Context(), util.ArtifactInfoCtxKey, art)
		req = req.WithContext(ctx)
	}
	uh.next.ServeHTTP(rw, req)
}

func parse(urlPath string) (map[string]string, bool) {
	m := make(map[string]string)
	match := false
	for _, re := range urlPatterns {
		l := re.FindStringSubmatch(urlPath)
		if len(l) > 0 {
			match = true
			for i := 1; i < len(l); i++ {
				m[re.SubexpNames()[i]] = l[i]
			}
		}
	}
	if digest.DigestRegexp.MatchString(m[util.ReferenceSubexp]) {
		m[util.DigestSubexp] = m[util.ReferenceSubexp]
	}
	return m, match
}
