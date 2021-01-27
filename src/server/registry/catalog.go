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
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/server/registry/util"
	"net/http"
	"sort"
	"strconv"
	"sync"
)

func newRepositoryHandler() http.Handler {
	return &repositoryHandler{
		repoCtl: repository.Ctl,
		tagCtl:  tag.Ctl,
		wp:      lib.NewWorkerPool(20),
	}
}

type repositoryHandler struct {
	repoCtl repository.Controller
	tagCtl  tag.Controller
	wp      *lib.WorkerPool
}

func (r *repositoryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var maxEntries int
	var err error
	var wg sync.WaitGroup

	reqQ := req.URL.Query()
	lastEntry := reqQ.Get("last")
	withN := reqQ.Get("n") != ""
	if withN {
		maxEntries, err = strconv.Atoi(reqQ.Get("n"))
		if err != nil || maxEntries < 0 {
			err := errors.New(err).WithCode(errors.BadRequestCode).WithMessage("the N must be a positive int type")
			lib_http.SendError(w, err)
			return
		}
	}

	repoNames := make([]string, 0)
	// get all repositories
	repoRecords, err := r.repoCtl.List(req.Context(), nil)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}
	if len(repoRecords) <= 0 {
		r.sendResponse(w, req, repoNames)
		return
	}

	repoNamesChan := make(chan string)
	done := make(chan bool, 1)
	errChan := make(chan error, 1)
	wg.Add(len(repoRecords))

	go func() {
		defer func() {
			done <- true
		}()

		for {
			select {
			case str, ok := <-repoNamesChan:
				if !ok {
					return
				}
				repoNames = append(repoNames, str)

			case e := <-errChan:
				if err == nil {
					err = errors.Wrap(e, "catalog API error")
				} else {
					err = errors.Wrap(e, err.Error())
				}
			}
		}
	}()

	for _, repo := range repoRecords {
		r.wp.GetWorker()
		go func(repo *models.RepoRecord) {
			defer func() {
				r.wp.ReleaseWorker()
				wg.Done()
			}()

			total, err := r.tagCtl.Count(orm.Clone(req.Context()), &q.Query{
				Keywords: map[string]interface{}{
					"repository_id": repo.RepositoryID,
				},
			})
			if err != nil {
				errChan <- err
				return
			}
			if total != 0 {
				repoNamesChan <- repo.Name
			}
		}(repo)
	}

	wg.Wait()
	close(repoNamesChan)
	<-done

	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	sort.Strings(repoNames)
	if !withN {
		r.sendResponse(w, req, repoNames)
		return
	}

	// handle the pagination
	resRepos := repoNames
	repoNamesLen := len(repoNames)
	// with "last", get items form lastEntryIndex+1 to lastEntryIndex+maxEntries
	// without "last", get items from 0 to maxEntries'
	if lastEntry != "" {
		lastEntryIndex := util.IndexString(repoNames, lastEntry)
		if lastEntryIndex == -1 {
			err := errors.New(nil).WithCode(errors.BadRequestCode).WithMessage(fmt.Sprintf("the last: %s should be a valid repository name.", lastEntry))
			lib_http.SendError(w, err)
			return
		}
		if lastEntryIndex+1+maxEntries > repoNamesLen {
			resRepos = repoNames[lastEntryIndex+1 : repoNamesLen]
		} else {
			resRepos = repoNames[lastEntryIndex+1 : lastEntryIndex+1+maxEntries]
		}
	} else {
		if maxEntries > repoNamesLen {
			maxEntries = repoNamesLen
		}
		resRepos = repoNames[0:maxEntries]
	}

	if len(resRepos) == 0 {
		r.sendResponse(w, req, resRepos)
		return
	}

	// compare the last item to define whether return the link header.
	// if equals, means that there is no more items in DB. Do not need to give the link header.
	if repoNames[len(repoNames)-1] != resRepos[len(resRepos)-1] {
		urlStr, err := util.SetLinkHeader(req.URL.String(), maxEntries, resRepos[len(resRepos)-1])
		if err != nil {
			lib_http.SendError(w, err)
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
		lib_http.SendError(w, err)
		return
	}
}

type catalogAPIResponse struct {
	Repositories []string `json:"repositories"`
}
