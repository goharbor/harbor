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
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/oidc"
)

type idToken struct{}

func (i *idToken) Generate(req *http.Request) security.Context {
	ctx := req.Context()
	log := log.G(ctx)
	if lib.GetAuthMode(ctx) != common.OIDCAuth {
		return nil
	}
	if !strings.HasPrefix(req.URL.Path, "/api") {
		return nil
	}
	token := bearerToken(req)
	if len(token) == 0 {
		return nil
	}
	claims, err := oidc.VerifyToken(ctx, token)
	if err != nil {
		log.Warningf("failed to verify token: %v", err)
		return nil
	}
	u, err := user.Ctl.GetBySubIss(ctx, claims.Subject, claims.Issuer)
	if err != nil {
		log.Warningf("failed to get user based on token claims: %v", err)
		return nil
	}
	setting, err := config.OIDCSetting(ctx)
	if err != nil {
		log.Errorf("failed to get OIDC settings: %v", err)
		return nil
	}
	info, err := oidc.UserInfoFromIDToken(ctx, &oidc.Token{RawIDToken: token}, *setting)
	if err != nil {
		log.Errorf("Failed to get user info from ID token: %v", err)
		return nil
	}
	oidc.InjectGroupsToUser(info, u)
	log.Debugf("an ID token security context generated for request %s %s", req.Method, req.URL.Path)
	return local.NewSecurityContext(u)
}
