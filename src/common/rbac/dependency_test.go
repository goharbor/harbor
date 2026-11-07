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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/pkg/permission/types"
)

func TestMissingPermissionDependencies(t *testing.T) {
	cases := []struct {
		name    string
		granted []*types.Policy
		want    []string
	}{
		{"empty", nil, nil},
		{"retention read only", []*types.Policy{
			{Resource: ResourceTagRetention, Action: ActionRead},
			{Resource: ResourceTagRetention, Action: ActionList},
		}, nil},
		{"retention write with artifact delete", []*types.Policy{
			{Resource: ResourceTagRetention, Action: ActionCreate},
			{Resource: ResourceTagRetention, Action: ActionUpdate},
			{Resource: ResourceTagRetention, Action: ActionDelete},
			{Resource: ResourceTagRetention, Action: ActionOperate},
			{Resource: ResourceArtifact, Action: ActionDelete},
		}, nil},
		{"retention write without artifact delete", []*types.Policy{
			{Resource: ResourceTagRetention, Action: ActionCreate},
			{Resource: ResourceTagRetention, Action: ActionUpdate},
			{Resource: ResourceTagRetention, Action: ActionDelete},
			{Resource: ResourceArtifact, Action: ActionRead},
			nil,
		}, []string{
			"tag-retention:create requires artifact:delete",
			"tag-retention:update requires artifact:delete",
			"tag-retention:delete requires artifact:delete",
		}},
		{"retention run without artifact delete", []*types.Policy{
			{Resource: ResourceTagRetention, Action: ActionOperate},
		}, []string{"tag-retention:operate requires artifact:delete"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, MissingPermissionDependencies(tc.granted))
		})
	}
}

// Every permission catalog Harbor hands out must satisfy the dependencies it enforces on custom roles.
func TestPermissionCatalogsSatisfyDependencies(t *testing.T) {
	provider := &NolimitProvider{}
	for _, scope := range []scope{ScopeProject, ScopeRole, ScopeSystem} {
		assert.Empty(t, MissingPermissionDependencies(provider.GetPermissions(scope)), string(scope))
	}
}
