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

	commonsecret "github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/common/security"
	securitysecret "github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
)

type secret struct{}

func (s *secret) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())
	sec := commonsecret.FromRequest(req)
	if len(sec) == 0 {
		return nil
	}
	log.Debugf("a secret security context generated for request %s %s", req.Method, req.URL.Path)
	return securitysecret.NewSecurityContext(sec, config.SecretStore)
}
