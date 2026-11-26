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

package registry

import (
	"encoding/json"
	"net/http"
	"sort"

	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/server/registry/util"
)

func newRepositoryHandler() http.Handler {
	return &repositoryHandler{
		repoMgr: pkg.RepositoryMgr,
	}
}

type repositoryHandler struct {
	repoMgr repository.Manager
}

func (r *repositoryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	n, lastEntry, err := util.ParseNAndLastParameters(req)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	repoNames := make([]string, 0)
	// get all the non repositories
	repoRecords, err := r.repoMgr.NonEmptyRepos(req.Context())
	if err != nil {
		lib_http.SendError(w, err)
		return
	}
	if len(repoRecords) <= 0 {
		r.sendResponse(w, req, repoNames)
		return
	}
	for _, repo := range repoRecords {
		repoNames = append(repoNames, repo.Name)
	}
	sort.Strings(repoNames)

	resRepos, nextLast := util.PickItems(repoNames, n, lastEntry)
	if nextLast != "" && n != nil {
		urlStr, err := util.SetLinkHeader(req.URL.String(), *n, nextLast)
		if err != nil {
			lib_http.SendError(w, err)
			return
		}
		w.Header().Set("Link", urlStr)
	}

	r.sendResponse(w, req, resRepos)
}

// sendResponse ...
func (r *repositoryHandler) sendResponse(w http.ResponseWriter, _ *http.Request, repositoryNames []string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	if err := enc.Encode(catalogAPIResponse{
		Repositories: repositoryNames,
	}); err != nil {
		lib_http.SendError(w, err)
		return
	}
}

type catalogAPIResponse struct {
	Repositories []string `json:"repositories"`
}
