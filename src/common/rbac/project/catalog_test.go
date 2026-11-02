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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// TestScopeRoleCatalogMatchesProjectAdmin guards against drift between the
// hand-written ScopeRole catalog (common/rbac/const.go) and the projectAdmin
// built-in role: a custom role's permission ceiling is exactly what a project
// admin can hold, minus self:{read,update,delete} (project view/edit/delete,
// which are never selectable custom-role permissions — baseline visibility is
// granted by membership instead). If a new project resource is added to
// projectAdmin but forgotten in the ScopeRole catalog (or vice versa), it would
// silently become ungrantable/over-grantable for custom roles with no compile
// error; this test fails loudly instead.
func TestScopeRoleCatalogMatchesProjectAdmin(t *testing.T) {
	key := func(p *types.Policy) string {
		return fmt.Sprintf("%s:%s:%s", p.Resource, p.Action, p.Effect)
	}

	scopeRole := map[string]bool{}
	for _, p := range rbac.GetPermissionProvider().GetPermissions(rbac.ScopeRole) {
		scopeRole[key(p)] = true
	}

	projectAdmin := map[string]bool{}
	for _, p := range rolePoliciesMap["projectAdmin"] {
		if p.Resource == rbac.ResourceSelf {
			continue
		}
		projectAdmin[key(p)] = true
	}

	for k := range projectAdmin {
		if !scopeRole[k] {
			t.Errorf("permission %q is in projectAdmin (minus self) but missing from the ScopeRole catalog", k)
		}
	}
	for k := range scopeRole {
		if !projectAdmin[k] {
			t.Errorf("permission %q is in the ScopeRole catalog but not in projectAdmin (minus self)", k)
		}
	}
}
