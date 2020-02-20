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
	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/api/repository"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/q"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/registry/util"
	"github.com/goharbor/harbor/src/server/router"
	"net/http"
	"sort"
	"strconv"
)

func newTagHandler() http.Handler {
	return &tagHandler{
		repoCtl: repository.Ctl,
		artCtl:  artifact.Ctl,
	}
}

type tagHandler struct {
	repoCtl        repository.Controller
	artCtl         artifact.Controller
	repositoryName string
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
func (t *tagHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

	var tagNames []string

	t.repositoryName = router.Param(req.Context(), ":splat")
	repository, err := t.repoCtl.GetByName(req.Context(), t.repositoryName)
	if err != nil {
		serror.SendError(w, err)
		return
	}

	// get tags ...
	tags, err := t.artCtl.ListTags(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": repository.RepositoryID,
		}}, nil)
	if err != nil {
		serror.SendError(w, err)
		return
	}
	if len(tags) == 0 {
		t.sendResponse(w, req, tagNames)
		return
	}

	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	sort.Strings(tagNames)
	if !withN {
		t.sendResponse(w, req, tagNames)
		return
	}

	// handle the pagination
	resTags := tagNames
	tagNamesLen := len(tagNames)
	// with "last", get items form lastEntryIndex+1 to lastEntryIndex+maxEntries
	// without "last", get items from 0 to maxEntries'
	if lastEntry != "" {
		lastEntryIndex := util.IndexString(tagNames, lastEntry)
		if lastEntryIndex == -1 {
			err := ierror.New(nil).WithCode(ierror.BadRequestCode).WithMessage(fmt.Sprintf("the last: %s should be a valid tag name.", lastEntry))
			serror.SendError(w, err)
			return
		}
		if lastEntryIndex+1+maxEntries > tagNamesLen {
			resTags = tagNames[lastEntryIndex+1 : tagNamesLen]
		} else {
			resTags = tagNames[lastEntryIndex+1 : lastEntryIndex+1+maxEntries]
		}
	} else {
		if maxEntries > tagNamesLen {
			maxEntries = tagNamesLen
		}
		resTags = tagNames[0:maxEntries]
	}

	if len(resTags) == 0 {
		t.sendResponse(w, req, resTags)
		return
	}

	// compare the last item to define whether return the link header.
	// if equals, means that there is no more items in DB. Do not need to give the link header.
	if tagNames[len(tagNames)-1] != resTags[len(resTags)-1] {
		urlStr, err := util.SetLinkHeader(req.URL.String(), maxEntries, resTags[len(resTags)-1])
		if err != nil {
			serror.SendError(w, err)
			return
		}
		w.Header().Set("Link", urlStr)
	}
	t.sendResponse(w, req, resTags)
	return
}

// sendResponse ...
func (t *tagHandler) sendResponse(w http.ResponseWriter, req *http.Request, tagNames []string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	if err := enc.Encode(tagsAPIResponse{
		Name: t.repositoryName,
		Tags: tagNames,
	}); err != nil {
		serror.SendError(w, err)
		return
	}
}

type tagsAPIResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
