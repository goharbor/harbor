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

package proxycachesecret

import (
	"context"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// const definition
const (
	// contains "#" to avoid the conflict with normal user
	ProxyCacheService = "harbor#proxy-cache-service"
)

// SecurityContext is the security context for proxy cache secret
type SecurityContext struct {
	repository string
	ctl        project.Controller
}

// NewSecurityContext returns an instance of the proxy cache secret security context
func NewSecurityContext(repository string) *SecurityContext {
	return &SecurityContext{
		repository: repository,
		ctl:        project.Ctl,
	}
}

// Name returns the name of the security context
func (s *SecurityContext) Name() string {
	return "proxy_cache_secret"
}

// IsAuthenticated always returns true
func (s *SecurityContext) IsAuthenticated() bool {
	return true
}

// GetUsername returns the name of proxy cache service
func (s *SecurityContext) GetUsername() string {
	return ProxyCacheService
}

// IsSysAdmin always returns false
func (s *SecurityContext) IsSysAdmin() bool {
	return false
}

// IsSolutionUser always returns false
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// Can returns true only when requesting pull/push operation against the specific project
func (s *SecurityContext) Can(ctx context.Context, action types.Action, resource types.Resource) bool {
	if !(action == rbac.ActionPull || action == rbac.ActionPush) {
		log.Debugf("unauthorized for action %s", action)
		return false
	}
	namespace, ok := rbac_project.NamespaceParse(resource)
	if !ok {
		log.Debugf("got no namespace from the resource %s", resource)
		return false
	}

	p, err := s.ctl.Get(ctx, namespace.Identity().(int64))
	if err != nil {
		log.Errorf("failed to get project %v: %v", namespace.Identity(), err)
		return false
	}

	pro, _ := utils.ParseRepository(s.repository)
	if p.Name != pro {
		log.Debugf("unauthorized for project %s", p.Name)
		return false
	}
	return true
}
