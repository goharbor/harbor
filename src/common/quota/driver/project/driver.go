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

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	dr "github.com/goharbor/harbor/src/common/quota/driver"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/graph-gophers/dataloader"
)

func init() {
	dr.Register("project", newDriver())
}

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

	projects, err := dao.GetProjects(&models.ProjectQueryParam{})
	if err != nil {
		return handleError(err)
	}

	var projectsMap = make(map[int64]*models.Project, len(projectIDs))
	for _, project := range projects {
		projectsMap[project.ProjectID] = project
	}

	var results []*dataloader.Result
	for _, projectID := range projectIDs {
		project, ok := projectsMap[projectID]
		if !ok {
			return handleError(fmt.Errorf("project not found, "+"project_id: %d", projectID))
		}

		result := dataloader.Result{
			Data:  project,
			Error: nil,
		}
		results = append(results, &result)
	}

	return results
}

type driver struct {
	cfg    *config.CfgManager
	loader *dataloader.Loader
}

func (d *driver) HardLimits() types.ResourceList {
	return types.ResourceList{
		types.ResourceCount:   d.cfg.Get(common.CountPerProject).GetInt64(),
		types.ResourceStorage: d.cfg.Get(common.StoragePerProject).GetInt64(),
	}
}

func (d *driver) Load(key string) (dr.RefObject, error) {
	thunk := d.loader.Load(context.TODO(), dataloader.StringKey(key))

	result, err := thunk()
	if err != nil {
		return nil, err
	}

	project, ok := result.(*models.Project)
	if !ok {
		return nil, fmt.Errorf("bad result for project: %s", key)
	}

	return dr.RefObject{
		"id":         project.ProjectID,
		"name":       project.Name,
		"owner_name": project.OwnerName,
	}, nil
}

func (d *driver) Validate(hardLimits types.ResourceList) error {
	resources := map[types.ResourceName]bool{
		types.ResourceCount:   true,
		types.ResourceStorage: true,
	}

	for resource, value := range hardLimits {
		if !resources[resource] {
			return fmt.Errorf("resource %s not support", resource)
		}

		if value <= 0 && value != types.UNLIMITED {
			return fmt.Errorf("invalid value for resource %s", resource)
		}
	}

	for resource := range resources {
		if _, found := hardLimits[resource]; !found {
			return fmt.Errorf("resource %s not found", resource)
		}
	}

	return nil
}

func newDriver() dr.Driver {
	cfg := config.NewDBCfgManager()

	loader := dataloader.NewBatchedLoader(getProjectsBatchFn)

	return &driver{cfg: cfg, loader: loader}
}
