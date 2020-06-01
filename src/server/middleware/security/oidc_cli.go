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
	"fmt"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils/oidc"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/handler/base"
)

type oidcCli struct{}

func (o *oidcCli) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	path := req.URL.Path
	if !isTarget(path) {
		return nil
	}
	if lib.GetAuthMode(req.Context()) != common.OIDCAuth {
		return nil
	}
	username, secret, ok := req.BasicAuth()
	if !ok {
		return nil
	}
	user, err := oidc.VerifySecret(req.Context(), username, secret)
	if err != nil {
		log.Errorf("failed to verify secret: %v", err)
		return nil
	}
	log.Debugf("an OIDC CLI security context generated for request %s %s", req.Method, req.URL.Path)
	return local.NewSecurityContext(user, config.GlobalProjectMgr)
}

// only handles request by docker CLI or helm CLI
func isTarget(path string) bool {
	if path == "/service/token" {
		return true
	}
	if path == "/service/token" ||
		strings.HasPrefix(path, "/v2") ||
		strings.HasPrefix(path, "/chartrepo/") ||
		strings.HasPrefix(path, "/api/chartrepo/") {
		return true
	}
	for _, version := range base.AvailableAPIVersions {
		if strings.HasPrefix(path, fmt.Sprintf("/api/%s/chartrepo/", version)) {
			return true
		}
	}
	return false
}
