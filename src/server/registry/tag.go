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
	"net/http"

	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/tag"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/server/registry/util"
	"github.com/goharbor/harbor/src/server/router"
)

func newTagHandler() http.Handler {
	return &tagHandler{
		repoCtl: repository.Ctl,
		tagCtl:  tag.Ctl,
	}
}

type tagHandler struct {
	repoCtl repository.Controller
	tagCtl  tag.Controller
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
	n, _, err := util.ParseNAndLastParameters(req)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	if n != nil && *n == 0 {
		util.SendListTagsResponse(w, req, nil)
		return
	}

	repositoryName := router.Param(req.Context(), ":splat")
	repository, err := t.repoCtl.GetByName(req.Context(), repositoryName)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	tags, err := t.tagCtl.List(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": repository.RepositoryID,
		}}, nil)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	tagNames := make([]string, len(tags))
	for i := range tags {
		tagNames[i] = tags[i].Name
	}

	util.SendListTagsResponse(w, req, tagNames)
}
