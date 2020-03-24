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
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/lib"
	ierror "github.com/goharbor/harbor/src/lib/error"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/middleware"
)

type reqChecker struct {
	pm promgr.ProjectManager
}

func (rc *reqChecker) check(req *http.Request) error {
	if rc.hasRegistryCred(req) {
		// TODO: May consider implement a local authorizer for registry, more details see #10602
		return nil
	}
	securityCtx, ok := security.FromContext(req.Context())
	if !ok {
		return fmt.Errorf("the security context got from request is nil")
	}
	none := lib.ArtifactInfo{}
	if a := lib.GetArtifactInfo(req.Context()); a != none {
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
			pid, err := rc.projectID(a.BlobMountProjectName)
			if err != nil {
				return err
			}
			resource := rbac.NewProjectNamespace(pid).Resource(rbac.ResourceRepository)
			if !securityCtx.Can(rbac.ActionPull, resource) {
				return fmt.Errorf("unauthorized to access repository from which to mount blob: %s, action: %s", a.BlobMountRepository, rbac.ActionPull)
			}
		}
	} else if len(middleware.V2CatalogURLRe.FindStringSubmatch(req.URL.Path)) == 1 && !securityCtx.IsSysAdmin() {
		return fmt.Errorf("unauthorized to list catalog")
	} else if req.URL.Path == "/v2/" && !securityCtx.IsAuthenticated() {
		return errors.New("unauthorized")
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

func (rc *reqChecker) hasRegistryCred(req *http.Request) bool {
	u, p, ok := req.BasicAuth()
	regUser, regPass := config.RegistryCredential()
	return ok && u == regUser && p == regPass
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

var (
	once    sync.Once
	checker reqChecker
)

// Middleware checks the permission of the request to access the artifact
func Middleware() func(http.Handler) http.Handler {
	once.Do(func() {
		if checker.pm == nil { // for UT, where pm has been set to a mock value
			checker = reqChecker{
				pm: config.GlobalProjectMgr,
			}
		}
	})
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if err := checker.check(req); err != nil {
				// the header is needed for "docker manifest" commands: https://github.com/docker/cli/issues/989
				rw.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
				rw.Header().Set("Www-Authenticate", `Basic realm="harbor"`)
				serror.SendError(rw, ierror.UnauthorizedError(err).WithMessage(err.Error()))
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}
