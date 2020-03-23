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

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/core/config"
)

type basicAuth struct{}

func (b *basicAuth) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	username, password, ok := req.BasicAuth()
	if !ok {
		return nil
	}
	user, err := auth.Login(models.AuthModel{
		Principal: username,
		Password:  password,
	})
	if err != nil {
		log.Errorf("failed to authenticate %s: %v", username, err)
		return nil
	}
	if user == nil {
		log.Debug("basic auth user is nil")
		return nil
	}
	log.Debugf("a basic auth security context generated for request %s %s", req.Method, req.URL.Path)
	return local.NewSecurityContext(user, config.GlobalProjectMgr)
}
