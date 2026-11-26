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

	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"

	"github.com/goharbor/harbor/src/pkg/permission/types"
)

var (
	errNotImplemented = errors.New("not implemented")
)

func policyLinesOfRole(rbacRole types.RBACRole) [][]string {
	lines := [][]string{}

	roleName := rbacRole.GetRoleName()
	// returns empty policy lines if role name is empty
	if roleName == "" {
		return lines
	}

	for _, policy := range rbacRole.GetPolicies() {
		line := []string{"p", roleName, policy.Resource.String(), policy.Action.String(), policy.GetEffect()}
		lines = append(lines, line)
	}

	return lines
}

func policyLinesOfRBACUser(rbacUser types.RBACUser) [][]string {
	lines := [][]string{}

	username := rbacUser.GetUserName()
	for _, policy := range rbacUser.GetPolicies() {
		line := []string{"p", username, policy.Resource.String(), policy.Action.String(), policy.GetEffect()}
		lines = append(lines, line)
	}

	return lines
}

type adapter struct {
	rbacUser types.RBACUser
}

func (a *adapter) getPolicyLines() [][]string {
	lines := [][]string{}

	username := a.rbacUser.GetUserName()
	// returns empty policy lines if username is empty
	if username == "" {
		return lines
	}

	lines = append(lines, policyLinesOfRBACUser(a.rbacUser)...)

	for _, role := range a.rbacUser.GetRoles() {
		lines = append(lines, policyLinesOfRole(role)...)
		lines = append(lines, []string{"g", username, role.GetRoleName()})
	}

	return lines
}

func (a *adapter) LoadPolicy(model model.Model) error {
	// arrays, not text lines: casbin v3 parses lines as CSV, which breaks on names with quotes or commas
	for _, line := range a.getPolicyLines() {
		if err := persist.LoadPolicyArray(line, model); err != nil {
			return err
		}
	}

	return nil
}

func (a *adapter) SavePolicy(_ model.Model) error {
	return errNotImplemented
}

func (a *adapter) AddPolicy(_ string, _ string, _ []string) error {
	return errNotImplemented
}

func (a *adapter) RemovePolicy(_ string, _ string, _ []string) error {
	return errNotImplemented
}

func (a *adapter) RemoveFilteredPolicy(_ string, _ string, _ int, _ ...string) error {
	return errNotImplemented
}
