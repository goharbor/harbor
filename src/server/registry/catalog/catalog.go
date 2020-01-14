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

package catalog

import (
	"encoding/json"
	"fmt"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/repository"
	reg_error "github.com/goharbor/harbor/src/server/registry/error"
	"github.com/goharbor/harbor/src/server/registry/util"
	"net/http"
	"sort"
	"strconv"
)

// NewHandler returns the handler to handler catalog request
func NewHandler(repoMgr repository.Manager) http.Handler {
	return &handler{
		repoMgr: repoMgr,
	}
}

type handler struct {
	repoMgr repository.Manager
}

// ServeHTTP ...
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		h.get(w, req)
	}
}

func (h *handler) get(w http.ResponseWriter, req *http.Request) {
	var maxEntries int
	var err error

	reqQ := req.URL.Query()
	lastEntry := reqQ.Get("last")
	withN := reqQ.Get("n") != ""
	if withN {
		maxEntries, err = strconv.Atoi(reqQ.Get("n"))
		if err != nil || maxEntries < 0 {
			err := ierror.New(err).WithCode(ierror.BadRequestCode).WithMessage("the N must be a positive int type")
			reg_error.Handle(w, req, err)
			return
		}
	}

	var repoNames []string
	// get all repositories
	// ToDo filter out the untagged repos
	total, repoRecords, err := h.repoMgr.List(req.Context(), nil)
	if err != nil {
		reg_error.Handle(w, req, err)
		return
	}
	if total <= 0 {
		h.sendResponse(w, req, repoNames)
		return
	}
	for _, r := range repoRecords {
		repoNames = append(repoNames, r.Name)
	}
	sort.Strings(repoNames)
	if !withN {
		h.sendResponse(w, req, repoNames)
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
			reg_error.Handle(w, req, err)
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
			reg_error.Handle(w, req, err)
			return
		}
		w.Header().Set("Link", urlStr)
	}

	h.sendResponse(w, req, resRepos)
	return
}

// sendResponse ...
func (h *handler) sendResponse(w http.ResponseWriter, req *http.Request, repositoryNames []string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	if err := enc.Encode(catalogAPIResponse{
		Repositories: repositoryNames,
	}); err != nil {
		reg_error.Handle(w, req, err)
		return
	}
}

type catalogAPIResponse struct {
	Repositories []string `json:"repositories"`
}
