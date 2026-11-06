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

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
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
	currentUser.Password = ""
	currentUser.Salt = ""
	log.Debugf("a session security context generated for request %s %s", req.Method, req.URL.Path)
	return local.NewSecurityContext(currentUser)
}
