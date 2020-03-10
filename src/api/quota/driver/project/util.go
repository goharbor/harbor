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
	"context"
	"fmt"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/graph-gophers/dataloader"
)

func getProjectsBatchFn(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
	handleError := func(err error) []*dataloader.Result {
		var results []*dataloader.Result
		var result dataloader.Result
		result.Error = err
		results = append(results, &result)
		return results
	}

	var projectIDs []int64
	for _, key := range keys {
		id, err := strconv.ParseInt(key.String(), 10, 64)
		if err != nil {
			return handleError(err)
		}
		projectIDs = append(projectIDs, id)
	}

	projects, err := project.Mgr.List(&models.ProjectQueryParam{ProjectIDs: projectIDs})
	if err != nil {
		return handleError(err)
	}

	var ownerIDs []int
	var projectsMap = make(map[int64]*models.Project, len(projectIDs))
	for _, project := range projects {
		ownerIDs = append(ownerIDs, project.OwnerID)
		projectsMap[project.ProjectID] = project
	}

	// TODO: change to use user manager
	owners, err := dao.ListUsers(&models.UserQuery{UserIDs: ownerIDs})
	if err != nil {
		return handleError(err)
	}

	var ownersMap = make(map[int]*models.User, len(owners))
	for i, owner := range owners {
		ownersMap[owner.UserID] = &owners[i]
	}

	var results []*dataloader.Result
	for _, projectID := range projectIDs {
		project, ok := projectsMap[projectID]
		if !ok {
			return handleError(fmt.Errorf("project not found, "+"project_id: %d", projectID))
		}

		owner, ok := ownersMap[project.OwnerID]
		if ok {
			project.OwnerName = owner.Username
		}

		result := dataloader.Result{
			Data:  project,
			Error: nil,
		}
		results = append(results, &result)
	}

	return results
}
