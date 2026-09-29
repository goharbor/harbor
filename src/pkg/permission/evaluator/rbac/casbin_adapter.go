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

type adapter struct {
	policies []*types.Policy
	username string
}

// LoadPolicy loads the visitor's own grants.
//
// Role grants are not here. They live in casbin_rule, they are held in memory
// by the process-wide store, and the evaluator asks that store directly rather
// than copying a role's policy into a fresh enforcer on every request.
//
// The rules go in as arrays rather than as "p, sub, obj, act" text, because
// casbin parses a policy line with a CSV reader and there is no reason to
// format a string only to have it taken apart again.
func (a *adapter) LoadPolicy(model model.Model) error {
	if a.username == "" {
		return nil
	}
	for _, policy := range a.policies {
		// casbin keeps the slice it is given, so each rule gets its own.
		rule := []string{"p", a.username, policy.Resource.String(), policy.Action.String(), policy.GetEffect()}
		if err := persist.LoadPolicyArray(rule, model); err != nil {
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
