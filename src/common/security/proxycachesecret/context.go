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

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project"
)

// const definition
const (
	// contains "#" to avoid the conflict with normal user
	ProxyCacheService = "harbor#proxy-cache-service"
)

// SecurityContext is the security context for proxy cache secret
type SecurityContext struct {
	repository string
	getProject func(interface{}) (*models.Project, error)
}

// NewSecurityContext returns an instance of the proxy cache secret security context
func NewSecurityContext(ctx context.Context, repository string) *SecurityContext {
	return &SecurityContext{
		repository: repository,
		getProject: func(i interface{}) (*models.Project, error) {
			return project.Mgr.Get(ctx, i)
		},
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
func (s *SecurityContext) Can(action types.Action, resource types.Resource) bool {
	if !(action == rbac.ActionPull || action == rbac.ActionPush) {
		log.Debugf("unauthorized for action %s", action)
		return false
	}
	namespace, ok := rbac.ProjectNamespaceParse(resource)
	if !ok {
		log.Debugf("got no namespace from the resource %s", resource)
		return false
	}
	project, err := s.getProject(namespace.Identity())
	if err != nil {
		log.Errorf("failed to get project %v: %v", namespace.Identity(), err)
		return false
	}
	if project == nil {
		log.Debugf("project not found %v", namespace.Identity())
		return false
	}
	pro, _ := utils.ParseRepository(s.repository)
	if project.Name != pro {
		log.Debugf("unauthorized for project %s", project.Name)
		return false
	}
	return true
}
