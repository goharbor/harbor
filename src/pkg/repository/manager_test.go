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
	"context"
	"testing"

	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/testing/pkg/repository/dao"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type managerTestSuite struct {
	suite.Suite
	mgr *manager
	dao *dao.DAO
}

func (m *managerTestSuite) SetupTest() {
	m.dao = &dao.DAO{}
	m.mgr = &manager{
		dao: m.dao,
	}
}

func (m *managerTestSuite) TestCount() {
	m.dao.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	n, err := m.mgr.Count(context.Background(), nil)
	m.Nil(err)
	m.Equal(int64(1), n)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestList() {
	repository := &model.RepoRecord{
		RepositoryID: 1,
		ProjectID:    1,
		Name:         "library/hello-world",
	}
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.RepoRecord{repository}, nil)
	rpers, err := m.mgr.List(context.Background(), nil)
	m.Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGet() {
	repository := &model.RepoRecord{
		RepositoryID: 1,
		ProjectID:    1,
		Name:         "library/hello-world",
	}
	m.dao.On("Get", mock.Anything, mock.Anything).Return(repository, nil)
	repo, err := m.mgr.Get(context.Background(), 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Require().NotNil(repo)
	m.Equal(repository.RepositoryID, repo.RepositoryID)
}

func (m *managerTestSuite) TestGetByName() {
	repository := &model.RepoRecord{
		RepositoryID: 1,
		ProjectID:    1,
		Name:         "library/hello-world",
	}
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.RepoRecord{repository}, nil)
	repo, err := m.mgr.GetByName(context.Background(), "library/hello-world")
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Require().NotNil(repo)
	m.Equal(repository.RepositoryID, repo.RepositoryID)
}

func (m *managerTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	_, err := m.mgr.Create(context.Background(), &model.RepoRecord{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDelete() {
	m.dao.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.Delete(context.Background(), 1)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestUpdate() {
	m.dao.On("Update", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.Update(context.Background(), &model.RepoRecord{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestAddPullCount() {
	m.dao.On("AddPullCount", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.AddPullCount(context.Background(), 1, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestNonEmptyRepos() {
	repository := &model.RepoRecord{
		RepositoryID: 1,
		ProjectID:    1,
		Name:         "library/hello-world",
	}
	m.dao.On("NonEmptyRepos", mock.Anything).Return([]*model.RepoRecord{repository}, nil)
	repo, err := m.mgr.NonEmptyRepos(nil)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Equal(repository.RepositoryID, repo[0].RepositoryID)
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
