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

package immutable

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match/rule"
	"net/http"
)

type immutableHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &immutableHandler{
		next: next,
	}
}

// ServeHTTP ...
func (rh immutableHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if match, _, _ := util.MatchPushManifest(req); !match {
		rh.next.ServeHTTP(rw, req)
		return
	}
	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		var err error
		info, err = util.ParseManifestInfoFromPath(req)
		if err != nil {
			log.Error(err)
			rh.next.ServeHTTP(rw, req)
			return
		}
	}

	_, repoName := common_util.ParseRepository(info.Repository)
	matched, err := rule.NewRuleMatcher(info.ProjectID).Match(art.Candidate{
		Repository:  repoName,
		Tag:         info.Tag,
		NamespaceID: info.ProjectID,
	})
	if err != nil {
		log.Error(err)
		rh.next.ServeHTTP(rw, req)
		return
	}
	if !matched {
		rh.next.ServeHTTP(rw, req)
		return
	}

	artifactQuery := &models.ArtifactQuery{
		PID:  info.ProjectID,
		Repo: info.Repository,
		Tag:  info.Tag,
	}
	afs, err := dao.ListArtifacts(artifactQuery)
	if err != nil {
		log.Error(err)
		rh.next.ServeHTTP(rw, req)
		return
	}
	if len(afs) == 0 {
		rh.next.ServeHTTP(rw, req)
		return
	}

	// rule matched and non-existent is a immutable tag
	http.Error(rw, util.MarshalError("DENIED",
		fmt.Sprintf("The tag:%s:%s is immutable, cannot be overwrite.", info.Repository, info.Tag)), http.StatusPreconditionFailed)
	return
}
