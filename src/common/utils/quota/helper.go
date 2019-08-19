// Copyright 2018 Project Harbor Authors
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

package quota

import (
	"fmt"
	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	common_quota "github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/pkg/errors"
	"strconv"
)

// AlignQuota align data bases on db record.
func AlignQuota(chartController *chartserver.Controller) error {
	projects, err := dao.GetProjects(nil)
	if err != nil {
		return err
	}
	for _, project := range projects {
		var size, count int64
		aligners := getAligner(project, chartController)
		for _, aligner := range aligners {
			sizeTmp, err := aligner.Size()
			if err != nil {
				logger.Error(err)
				continue
			}
			size += sizeTmp
			countTmp, err := aligner.Count()
			if err != nil {
				logger.Error(err)
				continue
			}
			count += countTmp
		}
		if err := ensure(project.ProjectID, size, count); err != nil {
			logger.Error(err)
			continue
		}
	}
	return nil
}

func ensure(pid, size, count int64) error {
	quotaMgr, err := common_quota.NewManager("project", strconv.FormatInt(pid, 10))
	if err != nil {
		logger.Errorf("Error occurred when to new quota manager %v, just skip it.", err)
		return err
	}
	used := common_quota.ResourceList{
		common_quota.ResourceStorage: size,
		common_quota.ResourceCount:   count,
	}
	if err := quotaMgr.EnsureQuota(used); err != nil {
		logger.Errorf("cannot ensure quota for the project: %d, err: %v, just skip it.", pid, err)
		return err
	}
	return nil
}

// Aligner ...
type Aligner interface {
	// Count return the count of backend service bases on DB
	Count() (int64, error)

	// Size return the size of backend service bases on DB
	Size() (int64, error)
}

func getAligner(project *models.Project, chartController *chartserver.Controller) []Aligner {
	var aligners []Aligner
	reg := NewRegistryAligner(project)
	aligners = append(aligners, reg)
	if config.WithChartMuseum() {
		chart := NewChartAligner(project, chartController)
		aligners = append(aligners, chart)
	}
	return aligners
}

type registry struct {
	project *models.Project
}

// NewRegistryAligner ...
func NewRegistryAligner(project *models.Project) Aligner {
	return &registry{
		project: project,
	}
}

// Count ...
func (r registry) Count() (int64, error) {
	afQuery := &models.ArtifactQuery{
		PID: r.project.ProjectID,
	}
	afs, err := dao.ListArtifacts(afQuery)
	if err != nil {
		logger.Warningf("error happen on counting number of project:%d , error:%v, just skip it.", r.project.ProjectID, err)
		return int64(0), err
	}
	return int64(len(afs)), nil
}

// Size ...
func (r registry) Size() (int64, error) {
	size, err := dao.CountSizeOfProject(r.project.ProjectID)
	if err != nil {
		logger.Warningf("error happen on counting size of project:%d , error:%v, just skip it.", r.project.ProjectID, err)
		return int64(0), err
	}
	return size, nil
}

type chart struct {
	project         *models.Project
	chartController *chartserver.Controller
}

// NewChartAligner ...
func NewChartAligner(project *models.Project, chartController *chartserver.Controller) Aligner {
	return &chart{
		project:         project,
		chartController: chartController,
	}
}

// Count ...
func (c chart) Count() (int64, error) {
	count, err := c.chartController.GetCountOfCharts([]string{c.project.Name})
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("get chart count of project %d failed", c.project.ProjectID))
		logger.Error(err)
		return int64(0), err
	}
	return int64(count), nil
}

// Size ...
func (c chart) Size() (int64, error) {
	return int64(0), nil
}
