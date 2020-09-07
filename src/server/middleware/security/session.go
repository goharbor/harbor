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

	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
)

// UserIDSessionKey is the key for storing user id in session
const UserIDSessionKey = "user_id"

type session struct{}

func (s *session) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	store, err := beego.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	if err != nil {
		log.Errorf("failed to get the session store for request: %v", err)
		return nil
	}
	userInterface := store.Get(UserIDSessionKey)
	if userInterface == nil {
		return nil
	}
	uid, ok := userInterface.(int)
	if !ok {
		log.Warning("can not convert the data in session to user id")
		return nil
	}
	user, err := dao.GetUser(models.User{UserID: uid})
	if err != nil {
		log.Errorf("failed to get user info, error: %v, skip generating security context via session", err)
		return nil
	}
	if user == nil {
		log.Errorf("the user id from session: %d, not exist in DB", uid)
		return nil
	}
	log.Debugf("a session security context generated for request %s %s", req.Method, req.URL.Path)
	return local.NewSecurityContext(user, config.GlobalProjectMgr)
}
