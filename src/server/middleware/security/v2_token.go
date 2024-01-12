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

	registry_token "github.com/docker/distribution/registry/auth/token"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/v2token"
	svc_token "github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/token"
	v2 "github.com/goharbor/harbor/src/pkg/token/claims/v2"
)

type v2TokenClaims struct {
	v2.Claims
	Access []*registry_token.ResourceActions `json:"access"`
}

func (vtc *v2TokenClaims) Valid() error {
	if err := vtc.Claims.Valid(); err != nil {
		return err
	}
	if !vtc.VerifyAudience(svc_token.Registry, true) {
		return fmt.Errorf("invalid token audience: %s", vtc.Audience)
	}
	return nil
}

type v2Token struct{}

func (vt *v2Token) Generate(req *http.Request) security.Context {
	logger := log.G(req.Context())
	if !strings.HasPrefix(req.URL.Path, "/v2") {
		return nil
	}
	tokenStr := bearerToken(req)
	if len(tokenStr) == 0 {
		return nil
	}

	defaultOpt := token.DefaultTokenOptions()
	if defaultOpt == nil {
		logger.Warningf("failed to get default options")
		return nil
	}
	cl := &v2TokenClaims{}
	t, err := token.Parse(defaultOpt, tokenStr, cl)
	if err != nil {
		logger.Warningf("failed to decode bearer token: %v", err)
		return nil
	}
	if err := t.Claims.Valid(); err != nil {
		logger.Warningf("failed to decode bearer token: %v", err)
		return nil
	}
	claims, ok := t.Claims.(*v2TokenClaims)
	if !ok {
		logger.Warningf("invalid token claims.")
		return nil
	}
	return v2token.New(req.Context(), claims.Subject, claims.Access)
}
