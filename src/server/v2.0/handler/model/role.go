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

package model

import (
	"github.com/goharbor/harbor/src/controller/role"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Robot ...
type Role struct {
	*role.Role
}

// ToSwagger ...
func (r *Role) ToSwagger() *models.Role {
	perms := []*models.RolePermission{}
	for _, p := range r.Permissions {
		temp := &models.RolePermission{}
		if err := lib.JSONCopy(temp, p); err != nil {
			log.Warningf("failed to do JSONCopy on RolePermission, error: %v", err)
		}
		log.Debug("*** toSwagger -- appending permission")
		perms = append(perms, temp)
	}

	return &models.Role{
		ID:          r.ID,
		Name:        r.Name,
		RoleMask:    r.RoleMask,
		RoleCode:    r.RoleCode,
		Permissions: perms,
	}
}

// NewRole ...
func NewRole(r *role.Role) *Role {
	return &Role{
		Role: r,
	}
}
