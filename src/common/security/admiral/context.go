// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package admiral

import (
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/security/admiral/authcontext"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/promgr"
)

// SecurityContext implements security.Context interface based on
// auth context and project manager
type SecurityContext struct {
	ctx *authcontext.AuthContext
	pm  promgr.ProMgr
}

// NewSecurityContext ...
func NewSecurityContext(ctx *authcontext.AuthContext, pm promgr.ProMgr) *SecurityContext {
	return &SecurityContext{
		ctx: ctx,
		pm:  pm,
	}
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	if s.ctx == nil {
		return false
	}
	return len(s.ctx.PrincipalID) > 0
}

// GetUsername returns the username of the authenticated user
// It returns null if the user has not been authenticated
func (s *SecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	return s.ctx.PrincipalID
}

// IsSysAdmin returns whether the authenticated user is system admin
// It returns false if the user has not been authenticated
func (s *SecurityContext) IsSysAdmin() bool {
	if !s.IsAuthenticated() {
		return false
	}

	return s.ctx.IsSysAdmin()
}

// IsSolutionUser ...
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// HasReadPerm returns whether the user has read permission to the project
func (s *SecurityContext) HasReadPerm(projectIDOrName interface{}) bool {
	public, err := s.pm.IsPublic(projectIDOrName)
	if err != nil {
		log.Errorf("failed to check the public of project %v: %v",
			projectIDOrName, err)
		return false
	}
	if public {
		return true
	}

	// private project
	if !s.IsAuthenticated() {
		return false
	}

	// system admin
	if s.IsSysAdmin() {
		return true
	}

	roles := s.GetProjectRoles(projectIDOrName)

	return len(roles) > 0
}

// HasWritePerm returns whether the user has write permission to the project
func (s *SecurityContext) HasWritePerm(projectIDOrName interface{}) bool {
	if !s.IsAuthenticated() {
		return false
	}

	// system admin
	if s.IsSysAdmin() {
		return true
	}

	roles := s.GetProjectRoles(projectIDOrName)
	for _, role := range roles {
		switch role {
		case common.RoleProjectAdmin,
			common.RoleDeveloper:
			return true
		}
	}

	return false
}

// HasAllPerm returns whether the user has all permissions to the project
func (s *SecurityContext) HasAllPerm(projectIDOrName interface{}) bool {
	if !s.IsAuthenticated() {
		return false
	}

	// system admin
	if s.IsSysAdmin() {
		return true
	}

	roles := s.GetProjectRoles(projectIDOrName)
	for _, role := range roles {
		switch role {
		case common.RoleProjectAdmin:
			return true
		}
	}

	return false
}

// GetMyProjects ...
func (s *SecurityContext) GetMyProjects() ([]*models.Project, error) {
	return s.ctx.GetMyProjects(), nil
}

// GetProjectRoles ...
func (s *SecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	if !s.IsAuthenticated() || projectIDOrName == nil {
		return []int{}
	}

	return s.ctx.GetProjectRoles(projectIDOrName)
}
