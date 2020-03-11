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

	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/api/blob"
	"github.com/goharbor/harbor/src/api/chartmuseum"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/q"
	dr "github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/graph-gophers/dataloader"
)

func init() {
	dr.Register("project", newDriver())
}

type driver struct {
	cfg    *config.CfgManager
	loader *dataloader.Loader

	artifactCtl artifact.Controller
	blobCtl     blob.Controller
	chartCtl    chartmuseum.Controller
}

func (d *driver) Enabled(ctx context.Context, key string) (bool, error) {
	return d.cfg.Get(common.QuotaPerProjectEnable).GetBool(), nil
}

func (d *driver) HardLimits(ctx context.Context) types.ResourceList {
	return types.ResourceList{
		types.ResourceCount:   d.cfg.Get(common.CountPerProject).GetInt64(),
		types.ResourceStorage: d.cfg.Get(common.StoragePerProject).GetInt64(),
	}
}

func (d *driver) Load(ctx context.Context, key string) (dr.RefObject, error) {
	thunk := d.loader.Load(ctx, dataloader.StringKey(key))

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

func (d *driver) CalculateUsage(ctx context.Context, key string) (types.ResourceList, error) {
	projectID, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return nil, err
	}

	// HACK: base=* in KeyWords to filter all artifacts
	artifactsCount, err := d.artifactCtl.Count(ctx, q.New(q.KeyWords{"project_id": projectID, "base": "*"}))
	if err != nil {
		return nil, err
	}

	chartsCount, err := d.chartCtl.Count(ctx, projectID)
	if err != nil {
		return nil, err
	}

	size, err := d.blobCtl.CalculateTotalSizeByProject(ctx, projectID, true)
	if err != nil {
		return nil, err
	}

	return types.ResourceList{types.ResourceCount: artifactsCount + chartsCount, types.ResourceStorage: size}, nil
}

func newDriver() dr.Driver {
	cfg := config.NewDBCfgManager()

	loader := dataloader.NewBatchedLoader(getProjectsBatchFn)

	return &driver{
		cfg:         cfg,
		loader:      loader,
		artifactCtl: artifact.Ctl,
		blobCtl:     blob.Ctl,
		chartCtl:    chartmuseum.Ctl,
	}
}
