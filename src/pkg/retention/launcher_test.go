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
	"context"
	"github.com/goharbor/harbor/src/lib/orm"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"testing"

	"github.com/goharbor/harbor/src/common/job"
	_ "github.com/goharbor/harbor/src/lib/selector/selectors/doublestar"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/q"
	hjob "github.com/goharbor/harbor/src/testing/job"
	"github.com/goharbor/harbor/src/testing/mock"
	projecttesting "github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type fakeRetentionManager struct{}

func (f *fakeRetentionManager) GetTotalOfRetentionExecs(policyID int64) (int64, error) {
	return 0, nil
}

func (f *fakeRetentionManager) GetTotalOfTasks(executionID int64) (int64, error) {
	return 0, nil
}

func (f *fakeRetentionManager) CreatePolicy(ctx context.Context, p *policy.Metadata) (int64, error) {
	return 0, nil
}
func (f *fakeRetentionManager) UpdatePolicy(ctx context.Context, p *policy.Metadata) error {
	return nil
}
func (f *fakeRetentionManager) DeletePolicy(ctx context.Context, ID int64) error {
	return nil
}
func (f *fakeRetentionManager) GetPolicy(ctx context.Context, ID int64) (*policy.Metadata, error) {
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
func (f *fakeRetentionManager) DeleteExecution(eid int64) error {
	return nil
}
func (f *fakeRetentionManager) ListTasks(query ...*q.TaskQuery) ([]*Task, error) {
	return []*Task{
		{
			ID:          1,
			ExecutionID: 1,
			JobID:       "1",
		},
	}, nil
}
func (f *fakeRetentionManager) GetTask(taskID int64) (*Task, error) {
	return nil, nil
}
func (f *fakeRetentionManager) CreateTask(task *Task) (int64, error) {
	return 0, nil
}
func (f *fakeRetentionManager) UpdateTask(task *Task, cols ...string) error {
	return nil
}
func (f *fakeRetentionManager) UpdateTaskStatus(int64, string, int64) error {
	return nil
}
func (f *fakeRetentionManager) GetTaskLog(taskID int64) ([]byte, error) {
	return nil, nil
}
func (f *fakeRetentionManager) ListExecutions(policyID int64, query *q.Query) ([]*Execution, error) {
	return nil, nil
}
func (f *fakeRetentionManager) AppendHistory(history *History) (int64, error) {
	return 0, nil
}
func (f *fakeRetentionManager) ListHistories(executionID int64, query *q.Query) ([]*History, error) {
	return nil, nil
}

type launchTestSuite struct {
	suite.Suite
	projectMgr       project.Manager
	execMgr          *tasktesting.ExecutionManager
	taskMgr          *tasktesting.Manager
	repositoryMgr    *repository.Manager
	retentionMgr     Manager
	jobserviceClient job.Client
}

func (l *launchTestSuite) SetupTest() {
	pro1 := &proModels.Project{
		ProjectID: 1,
		Name:      "library",
	}
	pro2 := &proModels.Project{
		ProjectID: 2,
		Name:      "test",
	}
	projectMgr := &projecttesting.Manager{}
	mock.OnAnything(projectMgr, "List").Return([]*proModels.Project{
		pro1, pro2,
	}, nil)
	l.projectMgr = projectMgr
	l.repositoryMgr = &repository.Manager{}
	l.retentionMgr = &fakeRetentionManager{}
	l.execMgr = &tasktesting.ExecutionManager{}
	l.taskMgr = &tasktesting.Manager{}
	l.jobserviceClient = &hjob.MockJobClient{
		JobUUID: []string{"1"},
	}
}

func (l *launchTestSuite) TestGetProjects() {
	ctx := orm.Context()
	projects, err := getProjects(ctx, l.projectMgr)
	require.Nil(l.T(), err)
	assert.Equal(l.T(), 2, len(projects))
	assert.Equal(l.T(), int64(1), projects[0].NamespaceID)
	assert.Equal(l.T(), "library", projects[0].Namespace)
}

func (l *launchTestSuite) TestGetRepositories() {
	l.repositoryMgr.On("List", mock.Anything, mock.Anything).Return([]*model.RepoRecord{
		{
			RepositoryID: 1,
			ProjectID:    1,
			Name:         "library/image",
		},
	}, nil)
	ctx := orm.Context()
	repositories, err := getRepositories(ctx, l.projectMgr, l.repositoryMgr, 1)
	require.Nil(l.T(), err)
	l.repositoryMgr.AssertExpectations(l.T())
	assert.Equal(l.T(), 1, len(repositories))
	assert.Equal(l.T(), "library", repositories[0].Namespace)
	assert.Equal(l.T(), "image", repositories[0].Repository)
	assert.Equal(l.T(), "image", repositories[0].Kind)
}

func (l *launchTestSuite) TestLaunch() {
	l.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	l.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	l.taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)

	launcher := &launcher{
		projectMgr:       l.projectMgr,
		repositoryMgr:    l.repositoryMgr,
		retentionMgr:     l.retentionMgr,
		execMgr:          l.execMgr,
		taskMgr:          l.taskMgr,
		jobserviceClient: l.jobserviceClient,
	}

	ctx := orm.Context()
	var ply *policy.Metadata
	// nil policy
	n, err := launcher.Launch(ctx, ply, 1, false)
	require.NotNil(l.T(), err)

	// nil rules
	ply = &policy.Metadata{}
	n, err = launcher.Launch(ctx, ply, 1, false)
	require.Nil(l.T(), err)
	assert.Equal(l.T(), int64(0), n)

	// nil scope
	ply = &policy.Metadata{
		Rules: []rule.Metadata{
			{},
		},
	}
	_, err = launcher.Launch(ctx, ply, 1, false)
	require.NotNil(l.T(), err)

	// system scope
	l.repositoryMgr.On("List", mock.Anything, mock.Anything).Return([]*model.RepoRecord{
		{
			RepositoryID: 1,
			ProjectID:    1,
			Name:         "library/image",
		},
	}, nil)
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
							Pattern:    "library",
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
			{
				Disabled: true,
				ScopeSelectors: map[string][]*rule.Selector{
					"project": {
						{
							Kind:       "doublestar",
							Decoration: "nsMatches",
							Pattern:    "library1",
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
	n, err = launcher.Launch(ctx, ply, 1, false)
	require.Nil(l.T(), err)
	l.repositoryMgr.AssertExpectations(l.T())
	assert.Equal(l.T(), int64(1), n)
}

func (l *launchTestSuite) TestStop() {
	t := l.T()
	l.execMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	launcher := &launcher{
		projectMgr:       l.projectMgr,
		repositoryMgr:    l.repositoryMgr,
		retentionMgr:     l.retentionMgr,
		execMgr:          l.execMgr,
		taskMgr:          l.taskMgr,
		jobserviceClient: l.jobserviceClient,
	}
	ctx := orm.Context()
	// invalid execution ID
	err := launcher.Stop(ctx, 0)
	require.NotNil(t, err)

	err = launcher.Stop(ctx, 1)
	require.Nil(t, err)
}

func TestLaunchTestSuite(t *testing.T) {
	suite.Run(t, new(launchTestSuite))
}
