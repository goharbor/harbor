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

package system

import (
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

type rbacUser struct {
	username string
	policies []*types.Policy
}

// GetUserName returns username of the visitor
func (sru *rbacUser) GetUserName() string {
	return sru.username
}

// GetPolicies returns policies of the visitor
func (sru *rbacUser) GetPolicies() []*types.Policy {
	return sru.policies
}

// GetRoles returns roles of the visitor
func (sru *rbacUser) GetRoles() []types.RBACRole {
	return nil
}
