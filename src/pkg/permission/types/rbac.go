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

package types

// RBACRole the interface of rbac role
type RBACRole interface {
	// GetRoleName returns the role identity, if empty string role's policies will be ignore
	GetRoleName() string
	// GetPolicies returns the policies of the role
	GetPolicies() []*Policy
	// PolicyObject maps a resource into the object space this role's grants are
	// written against, and reports false when the resource is out of the role's
	// reach. A project role holds the same grants in every project it is held
	// in, so it maps /project/57/repository to /project/*/repository and
	// refuses a resource belonging to any other project.
	PolicyObject(Resource) (Resource, bool)
}

// RBACUser the interface of rbac user
type RBACUser interface {
	// GetUserName returns the user identity, if empty string user's all policies will be ignore
	GetUserName() string
	// GetPolicies returns special policies of the user
	GetPolicies() []*Policy
	// GetRoles returns roles the user owned
	GetRoles() []RBACRole
}
