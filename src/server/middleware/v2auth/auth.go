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

package v2auth

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/filter"
	"github.com/goharbor/harbor/src/core/promgr"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/server/middleware"
	reg_err "github.com/goharbor/harbor/src/server/registry/error"
	"net/http"
)

type reqChecker struct {
	pm promgr.ProjectManager
}

func (rc *reqChecker) check(req *http.Request) error {
	securityCtx, err := filter.GetSecurityContext(req)
	if err != nil {
		return err
	}

	if a, ok := middleware.ArtifactInfoFromContext(req.Context()); ok {
		action := getAction(req)
		if action == "" {
			return nil
		}
		log.Debugf("action: %s, repository: %s", action, a.Repository)
		pid, err := rc.projectID(a.ProjectName)
		if err != nil {
			return err
		}
		resource := rbac.NewProjectNamespace(pid).Resource(rbac.ResourceRepository)
		if !securityCtx.Can(action, resource) {
			return fmt.Errorf("unauthorized to access repository: %s, action: %s", a.Repository, action)
		}
		if req.Method == http.MethodPost && a.BlobMountProjectName != "" { // check permission for the source of blob mount
			p, err := rc.pm.Get(a.BlobMountProjectName)
			if err != nil {
				return err
			}
			resource := rbac.NewProjectNamespace(p.ProjectID).Resource(rbac.ResourceRepository)
			if !securityCtx.Can(rbac.ActionPull, resource) {
				return fmt.Errorf("unauthorized to access repository from which to mount blob: %s, action: %s", a.BlobMountRepository, rbac.ActionPull)
			}
		}
	} else if len(middleware.V2CatalogURLRe.FindStringSubmatch(req.URL.Path)) == 1 && !securityCtx.IsSysAdmin() {
		return fmt.Errorf("unauthorized to list catalog")
	} else if req.URL.Path == "/v2/" && !securityCtx.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}
	return nil
}

func (rc *reqChecker) projectID(name string) (int64, error) {
	p, err := rc.pm.Get(name)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, fmt.Errorf("project not found, name: %s", name)
	}
	return p.ProjectID, nil
}

func getAction(req *http.Request) rbac.Action {
	pushActions := map[string]struct{}{
		http.MethodPost:   {},
		http.MethodDelete: {},
		http.MethodPatch:  {},
		http.MethodPut:    {},
	}
	pullActions := map[string]struct{}{
		http.MethodGet:  {},
		http.MethodHead: {},
	}
	if _, ok := pushActions[req.Method]; ok {
		return rbac.ActionPush
	}
	if _, ok := pullActions[req.Method]; ok {
		return rbac.ActionPull
	}
	return ""

}

var checker = reqChecker{
	pm: config.GlobalProjectMgr,
}

// Middleware checks the permission of the request to access the artifact
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if err := checker.check(req); err != nil {
				reg_err.Handle(rw, req, ierror.UnauthorizedError(err))
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}
