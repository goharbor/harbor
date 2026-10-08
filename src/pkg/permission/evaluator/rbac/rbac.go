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
	"context"
	"sync"

	"github.com/casbin/casbin/v3"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/policy"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

var _ evaluator.Evaluator = &Evaluator{}

// Evaluator is the permission evaluator for an rbac user.
//
// A question is answered in two steps, in this order:
//
//  1. What the visitor's roles grant. That is casbin's policy, held in memory
//     by the process-wide store and reloaded when it changes, so this step
//     makes no database call and allocates nothing on a cache hit.
//
//  2. What the visitor has been granted directly, which for a robot account is
//     its own permission list and for a public project is what any visitor
//     gets. These are not roles -- they belong to one principal and are built
//     for one request -- so they stay in a per-request enforcer, exactly as
//     they were before.
//
// Most visitors have no direct grants at all, so step 2 builds nothing.
type Evaluator struct {
	rbacUser types.RBACUser
	store    *policy.Store

	once     sync.Once
	enforcer *casbin.Enforcer

	directOnce sync.Once
	direct     []*types.Policy
}

// HasPermission returns true when the rbac user has action permission for the resource
func (e *Evaluator) HasPermission(_ context.Context, resource types.Resource, action types.Action) bool {
	if e.rbacUser.GetUserName() == "" {
		return false
	}
	if e.rolesGrant(resource, action) {
		return true
	}
	return e.directGrants(resource, action)
}

func (e *Evaluator) rolesGrant(resource types.Resource, action types.Action) bool {
	for _, r := range e.rbacUser.GetRoles() {
		name := r.GetRoleName()
		if name == "" {
			continue
		}
		object, ok := r.PolicyObject(resource)
		if !ok {
			continue
		}
		granted, err := e.store.Enforce(name, object.String(), action.String())
		if err != nil {
			log.Errorf("failed to evaluate role %q: %v", name, err)
			continue
		}
		if granted {
			return true
		}
	}
	return false
}

func (e *Evaluator) directGrants(resource types.Resource, action types.Action) bool {
	// GetPolicies builds its result, so it is asked once per evaluator rather
	// than once per question.
	e.directOnce.Do(func() { e.direct = e.rbacUser.GetPolicies() })
	if len(e.direct) == 0 {
		return false
	}
	e.once.Do(func() {
		enforcer, err := makeEnforcer(e.direct, e.rbacUser.GetUserName())
		if err != nil {
			log.Errorf("failed to build the permission enforcer: %v", err)
			return
		}
		e.enforcer = enforcer
	})
	if e.enforcer == nil {
		return false
	}

	granted, err := e.enforcer.Enforce(e.rbacUser.GetUserName(), resource.String(), action.String())
	if err != nil {
		log.Errorf("failed to evaluate the permission: %v", err)
		return false
	}
	return granted
}

// New returns evaluator.Evaluator for the RBACUser
func New(rbacUser types.RBACUser) *Evaluator {
	return &Evaluator{rbacUser: rbacUser, store: policy.Default()}
}
