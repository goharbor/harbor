// Copyright 2017 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package casbin

import "github.com/casbin/casbin/util"

// GetRolesForUser gets the roles that a user has.
func (e *Enforcer) GetRolesForUser(name string) ([]string, error) {
	res, err := e.model["g"]["g"].RM.GetRoles(name)
	return res, err
}

// GetUsersForRole gets the users that has a role.
func (e *Enforcer) GetUsersForRole(name string) ([]string, error) {
	res, err := e.model["g"]["g"].RM.GetUsers(name)
	return res, err
}

// HasRoleForUser determines whether a user has a role.
func (e *Enforcer) HasRoleForUser(name string, role string) (bool, error) {
	roles, err := e.GetRolesForUser(name)
	if err != nil {
		return false, err
	}
	hasRole := false
	for _, r := range roles {
		if r == role {
			hasRole = true
			break
		}
	}

	return hasRole, nil
}

// AddRoleForUser adds a role for a user.
// Returns false if the user already has the role (aka not affected).
func (e *Enforcer) AddRoleForUser(user string, role string) bool {
	return e.AddGroupingPolicy(user, role)
}

// DeleteRoleForUser deletes a role for a user.
// Returns false if the user does not have the role (aka not affected).
func (e *Enforcer) DeleteRoleForUser(user string, role string) bool {
	return e.RemoveGroupingPolicy(user, role)
}

// DeleteRolesForUser deletes all roles for a user.
// Returns false if the user does not have any roles (aka not affected).
func (e *Enforcer) DeleteRolesForUser(user string) bool {
	return e.RemoveFilteredGroupingPolicy(0, user)
}

// DeleteUser deletes a user.
// Returns false if the user does not exist (aka not affected).
func (e *Enforcer) DeleteUser(user string) bool {
	return e.RemoveFilteredGroupingPolicy(0, user)
}

// DeleteRole deletes a role.
func (e *Enforcer) DeleteRole(role string) {
	e.RemoveFilteredGroupingPolicy(1, role)
	e.RemoveFilteredPolicy(0, role)
}

// DeletePermission deletes a permission.
// Returns false if the permission does not exist (aka not affected).
func (e *Enforcer) DeletePermission(permission ...string) bool {
	return e.RemoveFilteredPolicy(1, permission...)
}

// AddPermissionForUser adds a permission for a user or role.
// Returns false if the user or role already has the permission (aka not affected).
func (e *Enforcer) AddPermissionForUser(user string, permission ...string) bool {
	return e.AddPolicy(util.JoinSlice(user, permission...))
}

// DeletePermissionForUser deletes a permission for a user or role.
// Returns false if the user or role does not have the permission (aka not affected).
func (e *Enforcer) DeletePermissionForUser(user string, permission ...string) bool {
	return e.RemovePolicy(util.JoinSlice(user, permission...))
}

// DeletePermissionsForUser deletes permissions for a user or role.
// Returns false if the user or role does not have any permissions (aka not affected).
func (e *Enforcer) DeletePermissionsForUser(user string) bool {
	return e.RemoveFilteredPolicy(0, user)
}

// GetPermissionsForUser gets permissions for a user or role.
func (e *Enforcer) GetPermissionsForUser(user string) [][]string {
	return e.GetFilteredPolicy(0, user)
}

// HasPermissionForUser determines whether a user has a permission.
func (e *Enforcer) HasPermissionForUser(user string, permission ...string) bool {
	return e.HasPolicy(util.JoinSlice(user, permission...))
}

// GetImplicitRolesForUser gets implicit roles that a user has.
// Compared to GetRolesForUser(), this function retrieves indirect roles besides direct roles.
// For example:
// g, alice, role:admin
// g, role:admin, role:user
//
// GetRolesForUser("alice") can only get: ["role:admin"].
// But GetImplicitRolesForUser("alice") will get: ["role:admin", "role:user"].
func (e *Enforcer) GetImplicitRolesForUser(name string, domain ...string) []string {
	res := []string{}
	roleSet := make(map[string]bool)
	roleSet[name] = true

	q := make([]string, 0)
	q = append(q, name)

	for len(q) > 0 {
		name := q[0]
		q = q[1:]

		roles, err := e.rm.GetRoles(name, domain...)
		if err != nil {
			panic(err)
		}
		for _, r := range roles {
			if _, ok := roleSet[r]; !ok {
				res = append(res, r)
				q = append(q, r)
				roleSet[r] = true
			}
		}
	}

	return res
}

// GetImplicitPermissionsForUser gets implicit permissions for a user or role.
// Compared to GetPermissionsForUser(), this function retrieves permissions for inherited roles.
// For example:
// p, admin, data1, read
// p, alice, data2, read
// g, alice, admin
//
// GetPermissionsForUser("alice") can only get: [["alice", "data2", "read"]].
// But GetImplicitPermissionsForUser("alice") will get: [["admin", "data1", "read"], ["alice", "data2", "read"]].
func (e *Enforcer) GetImplicitPermissionsForUser(user string, domain ...string) [][]string {
	roles := e.GetImplicitRolesForUser(user, domain...)
	roles = append([]string{user}, roles...)

	withDomain := false
	if len(domain) == 1 {
		withDomain = true
	} else if len(domain) > 1 {
		panic("error: domain should be 1 parameter")
	}

	res := [][]string{}
	permissions := [][]string{}
	for _, role := range roles {
		if withDomain {
			permissions = e.GetPermissionsForUserInDomain(role, domain[0])
		} else {
			permissions = e.GetPermissionsForUser(role)
		}
		res = append(res, permissions...)
	}

	return res
}

// GetImplicitUsersForPermission gets implicit users for a permission.
// For example:
// p, admin, data1, read
// p, bob, data1, read
// g, alice, admin
//
// GetImplicitUsersForPermission("data1", "read") will get: ["alice", "bob"].
// Note: only users will be returned, roles (2nd arg in "g") will be excluded.
func (e *Enforcer) GetImplicitUsersForPermission(permission ...string) []string {
	subjects := e.GetAllSubjects()
	roles := e.GetAllRoles()

	users := util.SetSubtract(subjects, roles)

	res := []string{}
	for _, user := range users {
		req := util.JoinSliceAny(user, permission...)
		if e.Enforce(req...) {
			res = append(res, user)
		}
	}

	return res
}
