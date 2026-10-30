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

package project

import (
	"path"
	"strings"

	"github.com/goharbor/harbor/src/pkg/permission/policy"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// projectRBACRole is one role a visitor holds in one project.
//
// It carries the role id project_member stores and the project the membership
// is in. What the role grants is not here: it is in the database, and the
// policy store holds it in memory for the life of the process.
type projectRBACRole struct {
	projectID int64
	roleID    int64
}

// GetRoleName returns how this role appears to the policy store.
//
// There is no switch on the role id here and no list of the ids Harbor happens
// to ship. A role id is looked up the same way whether it is 1 or 4001.
func (r *projectRBACRole) GetRoleName() string {
	return policy.Subject(r.roleID)
}

// PolicyObject maps a resource in this project into the object space the
// role's grants are written against.
//
// A role grants the same thing in every project it is held in, so the grant is
// stored against /project/:pid and the question is asked against /project/:pid
// too. A resource belonging to a different project is out of this membership's
// reach and is refused here rather than being left to the matcher.
func (r *projectRBACRole) PolicyObject(resource types.Resource) (types.Resource, bool) {
	relative, err := resource.RelativeTo(NewNamespace(r.projectID).Resource())
	if err != nil {
		return "", false
	}
	if relative == "." {
		return types.Resource(policy.Namespace), true
	}
	// Subresource joins, and joining cleans the path, so a ".." here would
	// climb back out of the project the membership is in and land on another
	// one. Only a path that is already canonical is mapped.
	if rel := relative.String(); rel != path.Clean(rel) || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", false
	}
	return types.Resource(policy.Namespace).Subresource(relative), true
}

// GetPolicies is not on the permission path any more. It stays for the
// interface and for the APIs that report what a role grants.
func (r *projectRBACRole) GetPolicies() []*types.Policy {
	return nil
}
