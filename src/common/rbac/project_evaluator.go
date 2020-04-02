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
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/namespace"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/rbac"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// NewProjectRBACEvaluator returns permission evaluator for project
func NewProjectRBACEvaluator(ctx security.Context, pm promgr.ProjectManager) evaluator.Evaluator {
	return namespace.New(ProjectNamespaceKind, func(ns types.Namespace) evaluator.Evaluator {
		project, err := pm.Get(ns.Identity())
		if err != nil || project == nil {
			if err != nil {
				log.Warningf("Failed to get info of project %d for permission evaluator, error: %v", ns.Identity(), err)
			}
			return nil
		}

		if ctx.IsAuthenticated() {
			roles := ctx.GetProjectRoles(project.ProjectID)
			return rbac.New(NewProjectRBACUser(project, ctx.GetUsername(), roles...))
		} else if project.IsPublic() {
			// anonymous access and the project is public
			return rbac.New(NewProjectRBACUser(project, "anonymous"))
		} else {
			return nil
		}
	})
}

// NewProjectRobotEvaluator returns robot permission evaluator for project
func NewProjectRobotEvaluator(ctx security.Context, pm promgr.ProjectManager,
	robotFactory func(types.Namespace) types.RBACUser) evaluator.Evaluator {

	return namespace.New(ProjectNamespaceKind, func(ns types.Namespace) evaluator.Evaluator {
		project, err := pm.Get(ns.Identity())
		if err != nil || project == nil {
			if err != nil {
				log.Warningf("Failed to get info of project %d for permission evaluator, error: %v", ns.Identity(), err)
			}
			return nil
		}

		if ctx.IsAuthenticated() {
			evaluators := evaluator.Evaluators{
				rbac.New(robotFactory(ns)), // robot account access
			}

			if project.IsPublic() {
				// authenticated access and the project is public
				evaluators = evaluators.Add(rbac.New(NewProjectRBACUser(project, ctx.GetUsername())))
			}

			return evaluators
		} else if project.IsPublic() {
			// anonymous access and the project is public
			return rbac.New(NewProjectRBACUser(project, "anonymous"))
		} else {
			return nil
		}
	})
}
