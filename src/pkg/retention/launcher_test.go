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

package retention

import (
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/q"
	_ "github.com/goharbor/harbor/src/pkg/retention/res/selectors/doublestar"
	hjob "github.com/goharbor/harbor/src/testing/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type fakeProjectManager struct {
	projects []*models.Project
}

func (f *fakeProjectManager) List(...*models.ProjectQueryParam) ([]*models.Project, error) {
	return f.projects, nil
}
func (f *fakeProjectManager) Get(idOrName interface{}) (*models.Project, error) {
	id, ok := idOrName.(int64)
	if ok {
		for _, pro := range f.projects {
			if pro.ProjectID == id {
				return pro, nil
			}
		}
		return nil, nil
	}
	name, ok := idOrName.(string)
	if ok {
		for _, pro := range f.projects {
			if pro.Name == name {
				return pro, nil
			}
		}
		return nil, nil
	}
	return nil, fmt.Errorf("invalid parameter: %v, should be ID(int64) or name(string)", idOrName)
}

type fakeRepositoryManager struct {
	imageRepositories []*models.RepoRecord
	chartRepositories []*chartserver.ChartInfo
}

func (f *fakeRepositoryManager) ListImageRepositories(projectID int64) ([]*models.RepoRecord, error) {
	return f.imageRepositories, nil
}
func (f *fakeRepositoryManager) ListChartRepositories(projectID int64) ([]*chartserver.ChartInfo, error) {
	return f.chartRepositories, nil
}

type fakeRetentionManager struct{}

func (f *fakeRetentionManager) CreatePolicy(p *policy.Metadata) (int64, error) {
	return 0, nil
}
func (f *fakeRetentionManager) UpdatePolicy(p *policy.Metadata) error {
	return nil
}
func (f *fakeRetentionManager) DeletePolicy(ID int64) error {
	return nil
}
func (f *fakeRetentionManager) GetPolicy(ID int64) (*policy.Metadata, error) {
	return nil, nil
}
func (f *fakeRetentionManager) CreateExecution(execution *Execution) (int64, error) {
	return 0, nil
}
func (f *fakeRetentionManager) UpdateExecution(execution *Execution) error {
	return nil
}
func (f *fakeRetentionManager) GetExecution(eid int64) (*Execution, error) {
	return nil, nil
}
func (f *fakeRetentionManager) ListTasks(query ...*q.TaskQuery) ([]*Task, error) {
	return nil, nil
}
func (f *fakeRetentionManager) CreateTask(task *Task) (int64, error) {
	return 0, nil
}
func (f *fakeRetentionManager) UpdateTask(task *Task, cols ...string) error {
	return nil
}
func (f *fakeRetentionManager) GetTaskLog(taskID int64) ([]byte, error) {
	return nil, nil
}
func (f *fakeRetentionManager) ListExecutions(query *q.Query) ([]*Execution, error) {
	return nil, nil
}
func (f *fakeRetentionManager) AppendHistory(history *History) error {
	return nil
}
func (f *fakeRetentionManager) ListHistories(executionID int64, query *q.Query) ([]*History, error) {
	return nil, nil
}

type launchTestSuite struct {
	suite.Suite
	projectMgr       project.Manager
	repositoryMgr    repository.Manager
	retentionMgr     Manager
	jobserviceClient job.Client
}

func (l *launchTestSuite) SetupTest() {
	pro := &models.Project{
		ProjectID: 1,
		Name:      "library",
	}
	l.projectMgr = &fakeProjectManager{
		projects: []*models.Project{
			pro,
		}}
	l.repositoryMgr = &fakeRepositoryManager{
		imageRepositories: []*models.RepoRecord{
			{
				Name: "library/image",
			},
		},
		chartRepositories: []*chartserver.ChartInfo{
			{
				Name: "chart",
			},
		},
	}
	l.retentionMgr = &fakeRetentionManager{}
	l.jobserviceClient = &hjob.MockJobClient{}
}

func (l *launchTestSuite) TestGetProjects() {
	projects, err := getProjects(l.projectMgr)
	require.Nil(l.T(), err)
	assert.Equal(l.T(), 1, len(projects))
	assert.Equal(l.T(), int64(1), projects[0].NamespaceID)
	assert.Equal(l.T(), "library", projects[0].Namespace)
}

func (l *launchTestSuite) TestGetRepositories() {
	repositories, err := getRepositories(l.projectMgr, l.repositoryMgr, 1)
	require.Nil(l.T(), err)
	assert.Equal(l.T(), 2, len(repositories))
	assert.Equal(l.T(), "library", repositories[0].Namespace)
	assert.Equal(l.T(), "image", repositories[0].Repository)
	assert.Equal(l.T(), "image", repositories[0].Kind)
	assert.Equal(l.T(), "library", repositories[1].Namespace)
	assert.Equal(l.T(), "chart", repositories[1].Repository)
	assert.Equal(l.T(), "chart", repositories[1].Kind)
}

func (l *launchTestSuite) TestLaunch() {
	launcher := &launcher{
		projectMgr:       l.projectMgr,
		repositoryMgr:    l.repositoryMgr,
		retentionMgr:     l.retentionMgr,
		jobserviceClient: l.jobserviceClient,
	}

	var ply *policy.Metadata
	// nil policy
	n, err := launcher.Launch(ply, 1)
	require.NotNil(l.T(), err)

	// nil rules
	ply = &policy.Metadata{}
	n, err = launcher.Launch(ply, 1)
	require.Nil(l.T(), err)
	assert.Equal(l.T(), int64(0), n)

	// nil scope
	ply = &policy.Metadata{
		Rules: []rule.Metadata{
			{},
		},
	}
	_, err = launcher.Launch(ply, 1)
	require.NotNil(l.T(), err)

	// system scope
	ply = &policy.Metadata{
		Scope: &policy.Scope{
			Level: "system",
		},
		Rules: []rule.Metadata{
			{
				ScopeSelectors: map[string][]*rule.Selector{
					"project": {
						{
							Kind:       "doublestar",
							Decoration: "nsMatches",
							Pattern:    "**",
						},
					},
					"repository": {
						{
							Kind:       "doublestar",
							Decoration: "repoMatches",
							Pattern:    "**",
						},
					},
				},
			},
		},
	}
	n, err = launcher.Launch(ply, 1)
	require.Nil(l.T(), err)
	assert.Equal(l.T(), int64(2), n)
}

func TestLaunchTestSuite(t *testing.T) {
	suite.Run(t, new(launchTestSuite))
}
