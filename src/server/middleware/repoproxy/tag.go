//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import (
	"context"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/controller/proxy"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib/errors"
	libhttp "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/registry/util"
)

// TagsListMiddleware handle list tags request
func TagsListMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		ctx := r.Context()

		art, p, _, err := preCheck(ctx)
		if err != nil {
			libhttp.SendError(w, err)
			return
		}

		if !canProxy(ctx, p) {
			next.ServeHTTP(w, r)
			return
		}

		n, _, err := util.ParseNAndLastParameters(r)
		if err != nil {
			libhttp.SendError(w, err)
			return
		}

		if n != nil && *n == 0 {
			util.SendListTagsResponse(w, r, nil)
			return
		}

		logger := log.G(ctx).WithField("project", art.ProjectName).WithField("repository", art.Repository)

		tags, err := getLocalTags(ctx, art.Repository)
		if err != nil {
			libhttp.SendError(w, err)
			return
		}

		defer func() {
			util.SendListTagsResponse(w, r, tags)
		}()

		remote, err := proxy.NewRemoteHelper(ctx, p.RegistryID)
		if err != nil {
			logger.Warningf("failed to get remote interface, error: %v, fallback to local tags", err)
			return
		}

		remoteRepository := strings.TrimPrefix(art.Repository, art.ProjectName+"/")
		remoteTags, err := remote.ListTags(remoteRepository)
		if err != nil {
			logger.Warningf("failed to get remote tags, error: %v, fallback to local tags", err)
			return
		}

		tags = append(tags, remoteTags...)
	})
}

var getLocalTags = func(ctx context.Context, repo string) ([]string, error) {
	r, err := repository.Ctl.GetByName(ctx, repo)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, err
	}

	items, err := tag.Ctl.List(ctx, q.New(q.KeyWords{"RepositoryID": r.RepositoryID}), nil)
	if err != nil {
		return nil, err
	}

	tags := make([]string, len(items))
	for i := range items {
		tags[i] = items[i].Name
	}

	return tags, nil
}
