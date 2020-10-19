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

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/proxycachesecret"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	ps "github.com/goharbor/harbor/src/pkg/proxy/secret"
)

type proxyCacheSecret struct{}

func (p *proxyCacheSecret) Generate(req *http.Request) security.Context {
	log := log.G(req.Context())

	artifact := lib.GetArtifactInfo(req.Context())
	if artifact == (lib.ArtifactInfo{}) {
		return nil
	}
	secret := ps.GetSecret(req)
	if !ps.GetManager().Verify(secret, artifact.Repository) {
		return nil
	}
	log.Debugf("a proxy cache secret security context generated for request %s %s", req.Method, req.URL.Path)
	return proxycachesecret.NewSecurityContext(artifact.Repository)
}
