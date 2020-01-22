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
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	"github.com/stretchr/testify/suite"
	"testing"
)

type controllerTestSuite struct {
	suite.Suite
	ctl     *controller
	proMgr  *project.FakeManager
	repoMgr *repository.FakeManager
}

func (c *controllerTestSuite) SetupTest() {
	c.proMgr = &project.FakeManager{}
	c.repoMgr = &repository.FakeManager{}
	c.ctl = &controller{
		proMgr:  c.proMgr,
		repoMgr: c.repoMgr,
	}
}

func (c *controllerTestSuite) TestEnsure() {
	// already exists
	c.repoMgr.On("List").Return(0, []*models.RepoRecord{
		{
			RepositoryID: 1,
			ProjectID:    1,
			Name:         "library/hello-world",
		},
	}, nil)
	created, id, err := c.ctl.Ensure(nil, "library/hello-world")
	c.Require().Nil(err)
	c.repoMgr.AssertExpectations(c.T())
	c.False(created)
	c.Equal(int64(1), id)

	// reset the mock
	c.SetupTest()

	// doesn't exist
	c.repoMgr.On("List").Return(0, []*models.RepoRecord{}, nil)
	c.proMgr.On("Get").Return(&models.Project{
		ProjectID: 1,
	}, nil)
	c.repoMgr.On("Create").Return(1, nil)
	created, id, err = c.ctl.Ensure(nil, "library/hello-world")
	c.Require().Nil(err)
	c.repoMgr.AssertExpectations(c.T())
	c.proMgr.AssertExpectations(c.T())
	c.True(created)
	c.Equal(int64(1), id)
}

func (c *controllerTestSuite) TestList() {
	c.repoMgr.On("List").Return(1, []*models.RepoRecord{
		{
			RepositoryID: 1,
		},
	}, nil)
	total, repositories, err := c.ctl.List(nil, nil)
	c.Require().Nil(err)
	c.repoMgr.AssertExpectations(c.T())
	c.Equal(int64(1), total)
	c.Require().Len(repositories, 1)
	c.Equal(int64(1), repositories[0].RepositoryID)
}

func (c *controllerTestSuite) TestGet() {
	c.repoMgr.On("Get").Return(&models.RepoRecord{
		RepositoryID: 1,
	}, nil)
	repository, err := c.ctl.Get(nil, 1)
	c.Require().Nil(err)
	c.repoMgr.AssertExpectations(c.T())
	c.Equal(int64(1), repository.RepositoryID)
}

func (c *controllerTestSuite) TestGetByName() {
	c.repoMgr.On("GetByName").Return(&models.RepoRecord{
		RepositoryID: 1,
	}, nil)
	repository, err := c.ctl.GetByName(nil, "library/hello-world")
	c.Require().Nil(err)
	c.repoMgr.AssertExpectations(c.T())
	c.Equal(int64(1), repository.RepositoryID)
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}
