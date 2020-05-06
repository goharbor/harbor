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

package rbac

import (
	"errors"
	"fmt"

	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

var (
	errNotImplemented = errors.New("not implemented")
)

func policyLinesOfRole(rbacRole types.RBACRole) []string {
	lines := []string{}

	roleName := rbacRole.GetRoleName()
	// returns empty policy lines if role name is empty
	if roleName == "" {
		return lines
	}

	for _, policy := range rbacRole.GetPolicies() {
		line := fmt.Sprintf("p, %s, %s, %s, %s", roleName, policy.Resource, policy.Action, policy.GetEffect())
		lines = append(lines, line)
	}

	return lines
}

func policyLinesOfRBACUser(rbacUser types.RBACUser) []string {
	lines := []string{}

	username := rbacUser.GetUserName()
	for _, policy := range rbacUser.GetPolicies() {
		line := fmt.Sprintf("p, %s, %s, %s, %s", username, policy.Resource, policy.Action, policy.GetEffect())
		lines = append(lines, line)
	}

	return lines
}

type adapter struct {
	rbacUser types.RBACUser
}

func (a *adapter) getPolicyLines() []string {
	lines := []string{}

	username := a.rbacUser.GetUserName()
	// returns empty policy lines if username is empty
	if username == "" {
		return lines
	}

	lines = append(lines, policyLinesOfRBACUser(a.rbacUser)...)

	for _, role := range a.rbacUser.GetRoles() {
		lines = append(lines, policyLinesOfRole(role)...)
		lines = append(lines, fmt.Sprintf("g, %s, %s", username, role.GetRoleName()))
	}

	return lines
}

func (a *adapter) LoadPolicy(model model.Model) error {
	for _, line := range a.getPolicyLines() {
		persist.LoadPolicyLine(line, model)
	}

	return nil
}

func (a *adapter) SavePolicy(model model.Model) error {
	return errNotImplemented
}

func (a *adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return errNotImplemented
}

func (a *adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return errNotImplemented
}

func (a *adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errNotImplemented
}
