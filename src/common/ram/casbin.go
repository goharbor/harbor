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

package ram

import (
	"errors"
	"fmt"

	"github.com/casbin/casbin"
	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
)

var (
	errNotImplemented = errors.New("Not implemented")
)

// Syntax for models see https://casbin.org/docs/en/syntax-for-models
const modelText = `
# Request definition
[request_definition]
r = sub, obj, act

# Policy definition
[policy_definition]
p = sub, obj, act, eft

# Role definition
[role_definition]
g = _, _

# Policy effect
[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

# Matchers
[matchers]
m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && (r.act == p.act || p.act == '*')
`

type userAdapter struct {
	User
}

func (a *userAdapter) getRolePolicyLines(role Role) []string {
	lines := []string{}

	roleName := role.GetRoleName()
	// returns empty policy lines if role name is empty
	if roleName == "" {
		return lines
	}

	for _, policy := range role.GetPolicies() {
		line := fmt.Sprintf("p, %s, %s, %s, %s", roleName, policy.Resource, policy.Action, policy.GetEffect())
		lines = append(lines, line)
	}

	return lines
}

func (a *userAdapter) getUserPolicyLines() []string {
	lines := []string{}

	username := a.GetUserName()
	// returns empty policy lines if username is empty
	if username == "" {
		return lines
	}

	for _, policy := range a.GetPolicies() {
		line := fmt.Sprintf("p, %s, %s, %s, %s", username, policy.Resource, policy.Action, policy.GetEffect())
		lines = append(lines, line)
	}

	return lines
}

func (a *userAdapter) getUserAllPolicyLines() []string {
	lines := []string{}

	username := a.GetUserName()
	// returns empty policy lines if username is empty
	if username == "" {
		return lines
	}

	lines = append(lines, a.getUserPolicyLines()...)

	for _, role := range a.GetRoles() {
		lines = append(lines, a.getRolePolicyLines(role)...)
		lines = append(lines, fmt.Sprintf("g, %s, %s", username, role.GetRoleName()))
	}

	return lines
}

func (a *userAdapter) LoadPolicy(model model.Model) error {
	for _, line := range a.getUserAllPolicyLines() {
		persist.LoadPolicyLine(line, model)
	}

	return nil
}

func (a *userAdapter) SavePolicy(model model.Model) error {
	return errNotImplemented
}

func (a *userAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	return errNotImplemented
}

func (a *userAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return errNotImplemented
}

func (a *userAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return errNotImplemented
}

func enforcerForUser(user User) *casbin.Enforcer {
	m := model.Model{}
	m.LoadModelFromText(modelText)
	return casbin.NewEnforcer(m, &userAdapter{User: user})
}
