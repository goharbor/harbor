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
	"path"
	"strings"
)

const (
	// EffectAllow allow effect
	EffectAllow = Effect("allow")
	// EffectDeny deny effect
	EffectDeny = Effect("deny")
)

// Resource the type of resource
type Resource string

// RelativeTo returns relative resource to other resource
func (res Resource) RelativeTo(other Resource) (Resource, error) {
	prefix := other.String()
	str := res.String()

	if !strings.HasPrefix(str, prefix) {
		return Resource(""), errors.New("value error")
	}

	relative := strings.TrimPrefix(str, prefix)
	if strings.HasPrefix(relative, "/") {
		relative = relative[1:]
	}

	if relative == "" {
		relative = "."
	}

	return Resource(relative), nil
}

func (res Resource) String() string {
	return string(res)
}

// Subresource returns subresource
func (res Resource) Subresource(resources ...Resource) Resource {
	elements := []string{res.String()}

	for _, resource := range resources {
		elements = append(elements, resource.String())
	}

	return Resource(path.Join(elements...))
}

// GetNamespace returns namespace from resource
func (res Resource) GetNamespace() (Namespace, error) {
	for _, parser := range namespaceParsers {
		namespace, err := parser(res)
		if err == nil {
			return namespace, nil
		}
	}

	return nil, fmt.Errorf("no namespace found for %s", res)
}

// Action the type of action
type Action string

func (act Action) String() string {
	return string(act)
}

// Effect the type of effect
type Effect string

func (eff Effect) String() string {
	return string(eff)
}

// Policy the type of policy
type Policy struct {
	Resource
	Action
	Effect
}

// GetEffect returns effect of resource, default is allow
func (p *Policy) GetEffect() string {
	eft := p.Effect
	if eft == "" {
		eft = EffectAllow
	}

	return eft.String()
}

func (p *Policy) String() string {
	return p.Resource.String() + ":" + p.Action.String() + ":" + p.GetEffect()
}

// Role the interface of rbac role
type Role interface {
	// GetRoleName returns the role identity, if empty string role's policies will be ignore
	GetRoleName() string
	GetPolicies() []*Policy
}

// User the interface of rbac user
type User interface {
	// GetUserName returns the user identity, if empty string user's all policies will be ignore
	GetUserName() string
	GetPolicies() []*Policy
	GetRoles() []Role
}

// BaseUser the type implement User interface whose policies are empty
type BaseUser struct{}

// GetRoles returns roles of the user
func (u *BaseUser) GetRoles() []Role {
	return nil
}

// GetUserName returns user identity
func (u *BaseUser) GetUserName() string {
	return ""
}

// GetPolicies returns policies of the user
func (u *BaseUser) GetPolicies() []*Policy {
	return nil
}

// HasPermission returns whether the user has action permission on resource
func HasPermission(user User, resource Resource, action Action) bool {
	return enforcerForUser(user).Enforce(user.GetUserName(), resource.String(), action.String())
}
