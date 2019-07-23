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

package repository

import (
	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/project"
)

// Manager is used for repository management
// currently, the interface only defines the methods needed for tag retention
// will expand it when doing refactor
type Manager interface {
	// List image repositories under the project specified by the ID
	ListImageRepositories(projectID int64) ([]*models.RepoRecord, error)
	// List chart repositories under the project specified by the ID
	ListChartRepositories(projectID int64) ([]*chartserver.ChartInfo, error)
	// IsChartServerEnabled returns whether the chart server is enabled
	IsChartServerEnabled() bool
}

// New returns a default implementation of Manager
func New(projectMgr project.Manager, chartCtl *chartserver.Controller) Manager {
	return &manager{
		projectMgr: projectMgr,
		chartCtl:   chartCtl,
	}
}

type manager struct {
	projectMgr project.Manager
	chartCtl   *chartserver.Controller
}

// List image repositories under the project specified by the ID
func (m *manager) ListImageRepositories(projectID int64) ([]*models.RepoRecord, error) {
	return dao.GetRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{projectID},
	})
}

// List chart repositories under the project specified by the ID
func (m *manager) ListChartRepositories(projectID int64) ([]*chartserver.ChartInfo, error) {
	project, err := m.projectMgr.Get(projectID)
	if err != nil {
		return nil, err
	}
	return m.chartCtl.ListCharts(project.Name)
}

func (m *manager) IsChartServerEnabled() bool {
	return m.chartCtl != nil
}
