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
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/authproxy"
)

type authProxy struct{}

func (a *authProxy) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	if lib.GetAuthMode(req.Context()) != common.HTTPAuth {
		return nil
	}
	// only support docker login
	if !strings.HasPrefix(req.URL.Path, "/v2") {
		return nil
	}
	proxyUserName, proxyPwd, ok := req.BasicAuth()
	if !ok {
		return nil
	}
	rawUserName, match := a.matchAuthProxyUserName(proxyUserName)
	if !match {
		log.Errorf("user name %s doesn't meet the auth proxy name pattern", proxyUserName)
		return nil
	}
	httpAuthProxyConf, err := config.HTTPAuthProxySetting()
	if err != nil {
		log.Errorf("failed to get auth proxy settings: %v", err)
		return nil
	}
	tokenReviewStatus, err := authproxy.TokenReview(proxyPwd, httpAuthProxyConf)
	if err != nil {
		log.Errorf("failed to review token: %v", err)
		return nil
	}
	if rawUserName != tokenReviewStatus.User.Username {
		log.Errorf("user name doesn't match with token: %s", rawUserName)
		return nil
	}
	user, err := dao.GetUser(models.User{
		Username: rawUserName,
	})
	if err != nil {
		log.Errorf("failed to get user %s: %v", rawUserName, err)
		return nil
	}
	if user == nil {
		// onboard user if it's not yet onboarded.
		uid, err := auth.SearchAndOnBoardUser(rawUserName)
		if err != nil {
			log.Errorf("failed to search and onboard user %s: %v", rawUserName, err)
			return nil
		}
		user, err = dao.GetUser(models.User{
			UserID: uid,
		})
		if err != nil {
			log.Errorf("failed to get user, name: %s, ID: %d: %v", rawUserName, uid, err)
			return nil
		}
	}
	u2, err := authproxy.UserFromReviewStatus(tokenReviewStatus)
	if err != nil {
		log.Errorf("failed to get user information from token review status: %v", err)
		return nil
	}
	user.GroupIDs = u2.GroupIDs
	log.Debugf("an auth proxy security context generated for request %s %s", req.Method, req.URL.Path)
	return local.NewSecurityContext(user, config.GlobalProjectMgr)
}

func (a *authProxy) matchAuthProxyUserName(name string) (string, bool) {
	if !strings.HasPrefix(name, common.AuthProxyUserNamePrefix) {
		return "", false
	}
	return strings.Replace(name, common.AuthProxyUserNamePrefix, "", -1), true
}
