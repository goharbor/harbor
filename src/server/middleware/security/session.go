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

package security

import (
	"net/http"
	"net/http/httptest"

	"github.com/beego/beego/v2/server/web"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	pkguser "github.com/goharbor/harbor/src/pkg/user"
)

type session struct{}

func (s *session) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	store, err := web.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	if err != nil {
		log.Errorf("failed to get the session store for request: %v", err)
		return nil
	}
	userInterface := store.Get(req.Context(), "user")
	if userInterface == nil {
		return nil
	}
	user, ok := userInterface.(models.User)
	if !ok {
		log.Warning("can not convert the user in session to user model")
		return nil
	}
	if user.UserID <= 0 {
		// An old proxy session may not contain an ID. Looking up by name would
		// allow a deleted user's session to attach to a new account of that name.
		log.Warning("session user has no valid ID; reauthentication is required")
		return nil
	}
	if lib.GetAuthMode(req.Context()) == common.HTTPAuth && user.SessionAuthMode != common.HTTPAuth && user.SessionAuthMode != common.DBAuth {
		// Legacy sessions do not identify their authenticator. In particular, a
		// proxy identity can have the same name and ID as the local superuser.
		log.Warning("session authentication source is unknown; reauthentication is required")
		return nil
	}
	currentUser, err := pkguser.Mgr.Get(req.Context(), user.UserID)
	if err != nil {
		log.Warningf("failed to load user %d for session: %v", user.UserID, err)
		return nil
	}
	if currentUser == nil || currentUser.Deleted {
		return nil
	}
	// These fields are supplied by the authentication provider and are not stored in the user table.
	currentUser.AdminRoleInAuth = user.AdminRoleInAuth
	currentUser.GroupIDs = user.GroupIDs
	if user.SessionAuthMode == common.HTTPAuth {
		// Proxy users derive administrator authority only from the provider.
		currentUser.SysAdminFlag = false
	}
	currentUser.Password = ""
	currentUser.Salt = ""
	log.Debugf("a session security context generated for request %s %s", req.Method, req.URL.Path)
	return local.NewSecurityContext(currentUser)
}
