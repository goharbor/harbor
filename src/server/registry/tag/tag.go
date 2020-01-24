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

package tag

import (
	"encoding/json"
	"fmt"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/tag"
	reg_error "github.com/goharbor/harbor/src/server/registry/error"
	"github.com/goharbor/harbor/src/server/registry/util"
	"github.com/gorilla/mux"
	"net/http"
	"sort"
	"strconv"
)

// NewHandler returns the handler to handle listing tag request
func NewHandler(repoMgr repository.Manager, tagMgr tag.Manager) http.Handler {
	return &handler{
		repoMgr: repoMgr,
		tagMgr:  tagMgr,
	}
}

type handler struct {
	repoMgr repository.Manager
	tagMgr  tag.Manager

	repositoryName string
}

// ServeHTTP ...
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		h.get(w, req)
	}
}

// get return the list of tags

// Content-Type: application/json
// Link: <<url>?n=<n from the request>&last=<last tag value from previous response>>; rel="next"
//
// {
//    "name": "<name>",
//    "tags": [
//      "<tag>",
//      ...
//    ]
// }
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

	var repoID int64
	var tagNames []string
	vars := mux.Vars(req)
	h.repositoryName = vars["name"]

	repoID, err = h.getRepoID(req)
	if err != nil {
		reg_error.Handle(w, req, err)
		return
	}

	// get tags ...
	total, tags, err := h.tagMgr.List(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": repoID,
		},
	})
	if err != nil {
		reg_error.Handle(w, req, err)
		return
	}
	if total == 0 {
		h.sendResponse(w, req, tagNames)
		return
	}

	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	sort.Strings(tagNames)
	if !withN {
		h.sendResponse(w, req, tagNames)
		return
	}

	// handle the pagination
	resTags := tagNames
	// with "last", get items form lastEntryIndex+1 to lastEntryIndex+maxEntries
	// without "last", get items from 0 to maxEntries'
	if lastEntry != "" {
		lastEntryIndex := util.IndexString(tagNames, lastEntry)
		if lastEntryIndex == -1 {
			err := ierror.New(nil).WithCode(ierror.BadRequestCode).WithMessage(fmt.Sprintf("the last: %s should be a valid tag name.", lastEntry))
			reg_error.Handle(w, req, err)
			return
		}
		resTags = tagNames[lastEntryIndex+1 : lastEntryIndex+maxEntries]
	} else {
		resTags = tagNames[0:maxEntries]
	}

	// compare the last item to define whether return the link header.
	// if equals, means that there is no more items in DB. Do not need to give the link header.
	if tagNames[len(tagNames)-1] != resTags[len(resTags)-1] {
		urlStr, err := util.SetLinkHeader(req.URL.String(), maxEntries, resTags[len(resTags)-1])
		if err != nil {
			reg_error.Handle(w, req, err)
			return
		}
		w.Header().Set("Link", urlStr)
	}
	h.sendResponse(w, req, resTags)
	return
}

// getRepoID ...
func (h *handler) getRepoID(req *http.Request) (int64, error) {
	total, repoRecord, err := h.repoMgr.List(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"name": h.repositoryName,
		},
	})
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		err := ierror.New(nil).WithCode(ierror.NotFoundCode).WithMessage("repositoryNotFound")
		return 0, err
	}
	return repoRecord[0].RepositoryID, nil
}

// sendResponse ...
func (h *handler) sendResponse(w http.ResponseWriter, req *http.Request, tagNames []string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	if err := enc.Encode(tagsAPIResponse{
		Name: h.repositoryName,
		Tags: tagNames,
	}); err != nil {
		reg_error.Handle(w, req, err)
		return
	}
}

type tagsAPIResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
