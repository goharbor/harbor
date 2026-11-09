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
	"fmt"

	"github.com/goharbor/harbor/src/pkg/permission/types"
)

type permissionDependency struct {
	Resource         types.Resource
	Action           types.Action
	RequiredResource types.Resource
	RequiredAction   types.Action
}

// permissionDependencies lists permissions that act as another permission in
// disguise. Writing or running a tag retention policy deletes artifacts, so a
// grant that holds it without artifact delete would bypass that right.
var permissionDependencies = []permissionDependency{
	{ResourceTagRetention, ActionCreate, ResourceArtifact, ActionDelete},
	{ResourceTagRetention, ActionUpdate, ResourceArtifact, ActionDelete},
	{ResourceTagRetention, ActionDelete, ResourceArtifact, ActionDelete},
	{ResourceTagRetention, ActionOperate, ResourceArtifact, ActionDelete},
}

// MissingPermissionDependencies returns one message per granted permission
// whose required companion permission is absent from the same grant.
func MissingPermissionDependencies(granted []*types.Policy) []string {
	has := func(resource types.Resource, action types.Action) bool {
		for _, p := range granted {
			if p != nil && p.Resource == resource && p.Action == action {
				return true
			}
		}
		return false
	}
	var missing []string
	for _, dep := range permissionDependencies {
		if has(dep.Resource, dep.Action) && !has(dep.RequiredResource, dep.RequiredAction) {
			missing = append(missing, fmt.Sprintf("%s:%s requires %s:%s",
				dep.Resource, dep.Action, dep.RequiredResource, dep.RequiredAction))
		}
	}
	return missing
}
