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

package rbac

import (
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/ui/projectmanager"
)

// SecurityContext implements security.Context interface based on database
type SecurityContext struct {
	user *models.User
	pm   projectmanager.ProjectManager
}

// NewSecurityContext ...
func NewSecurityContext(user *models.User, pm projectmanager.ProjectManager) *SecurityContext {
	return &SecurityContext{
		user: user,
		pm:   pm,
	}
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	return s.user != nil
}

// GetUsername returns the username of the authenticated user
// It returns null if the user has not been authenticated
func (s *SecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	return s.user.Username
}

// IsSysAdmin returns whether the authenticated user is system admin
// It returns false if the user has not been authenticated
func (s *SecurityContext) IsSysAdmin() bool {
	if !s.IsAuthenticated() {
		return false
	}
	return s.user.HasAdminRole == 1
}

// HasReadPerm returns whether the user has read permission to the project
func (s *SecurityContext) HasReadPerm(projectIDOrName interface{}) bool {
	// public project
	if s.pm.IsPublic(projectIDOrName) {
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

	roles := s.pm.GetRoles(s.GetUsername(), projectIDOrName)
	for _, role := range roles {
		switch role {
		case common.RoleProjectAdmin,
			common.RoleDeveloper,
			common.RoleGuest:
			return true
		}
	}

	return false
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

	roles := s.pm.GetRoles(s.GetUsername(), projectIDOrName)
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

	roles := s.pm.GetRoles(s.GetUsername(), projectIDOrName)
	for _, role := range roles {
		switch role {
		case common.RoleProjectAdmin:
			return true
		}
	}

	return false
}
