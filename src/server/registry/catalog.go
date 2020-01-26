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
	"fmt"
	"github.com/goharbor/harbor/src/api/repository"
	ierror "github.com/goharbor/harbor/src/internal/error"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/registry/util"
	"net/http"
	"sort"
	"strconv"
)

func newRepositoryHandler() http.Handler {
	return &repositoryHandler{
		repoCtl: repository.Ctl,
	}
}

type repositoryHandler struct {
	repoCtl repository.Controller
}

func (r *repositoryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var maxEntries int
	var err error

	reqQ := req.URL.Query()
	lastEntry := reqQ.Get("last")
	withN := reqQ.Get("n") != ""
	if withN {
		maxEntries, err = strconv.Atoi(reqQ.Get("n"))
		if err != nil || maxEntries < 0 {
			err := ierror.New(err).WithCode(ierror.BadRequestCode).WithMessage("the N must be a positive int type")
			serror.SendError(w, err)
			return
		}
	}

	var repoNames []string
	// get all repositories
	// ToDo filter out the untagged repos
	total, repoRecords, err := r.repoCtl.List(req.Context(), nil)
	if err != nil {
		serror.SendError(w, err)
		return
	}
	if total <= 0 {
		r.sendResponse(w, req, repoNames)
		return
	}
	for _, r := range repoRecords {
		repoNames = append(repoNames, r.Name)
	}
	sort.Strings(repoNames)
	if !withN {
		r.sendResponse(w, req, repoNames)
		return
	}

	// handle the pagination
	resRepos := repoNames
	// with "last", get items form lastEntryIndex+1 to lastEntryIndex+maxEntries
	// without "last", get items from 0 to maxEntries'
	if lastEntry != "" {
		lastEntryIndex := util.IndexString(repoNames, lastEntry)
		if lastEntryIndex == -1 {
			err := ierror.New(nil).WithCode(ierror.BadRequestCode).WithMessage(fmt.Sprintf("the last: %s should be a valid repository name.", lastEntry))
			serror.SendError(w, err)
			return
		}
		resRepos = repoNames[lastEntryIndex+1 : lastEntryIndex+maxEntries]
	} else {
		resRepos = repoNames[0:maxEntries]
	}

	// compare the last item to define whether return the link header.
	// if equals, means that there is no more items in DB. Do not need to give the link header.
	if repoNames[len(repoNames)-1] != resRepos[len(resRepos)-1] {
		urlStr, err := util.SetLinkHeader(req.URL.String(), maxEntries, resRepos[len(resRepos)-1])
		if err != nil {
			serror.SendError(w, err)
			return
		}
		w.Header().Set("Link", urlStr)
	}

	r.sendResponse(w, req, resRepos)
	return
}

// sendResponse ...
func (r *repositoryHandler) sendResponse(w http.ResponseWriter, req *http.Request, repositoryNames []string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	if err := enc.Encode(catalogAPIResponse{
		Repositories: repositoryNames,
	}); err != nil {
		serror.SendError(w, err)
		return
	}
}

type catalogAPIResponse struct {
	Repositories []string `json:"repositories"`
}
