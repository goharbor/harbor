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
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type fakeDao struct {
	mock.Mock
}

func (f *fakeDao) Count(ctx context.Context, query *q.Query) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) List(ctx context.Context, query *q.Query) ([]*models.RepoRecord, error) {
	args := f.Called()
	return args.Get(0).([]*models.RepoRecord), args.Error(1)
}
func (f *fakeDao) Get(ctx context.Context, id int64) (*models.RepoRecord, error) {
	args := f.Called()
	return args.Get(0).(*models.RepoRecord), args.Error(1)
}
func (f *fakeDao) Create(ctx context.Context, repository *models.RepoRecord) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) Update(ctx context.Context, repository *models.RepoRecord, props ...string) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) AddPullCount(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}

type managerTestSuite struct {
	suite.Suite
	mgr *manager
	dao *fakeDao
}

func (m *managerTestSuite) SetupTest() {
	m.dao = &fakeDao{}
	m.mgr = &manager{
		dao: m.dao,
	}
}

func (m *managerTestSuite) TestCount() {
	m.dao.On("Count", mock.Anything).Return(1, nil)
	total, err := m.mgr.Count(nil, nil)
	m.Require().Nil(err)
	m.Equal(int64(1), total)
}

func (m *managerTestSuite) TestList() {
	repository := &models.RepoRecord{
		RepositoryID: 1,
		ProjectID:    1,
		Name:         "library/hello-world",
	}
	m.dao.On("List", mock.Anything).Return([]*models.RepoRecord{repository}, nil)
	repositories, err := m.mgr.List(nil, nil)
	m.Require().Nil(err)
	m.Equal(1, len(repositories))
	m.Equal(repository.RepositoryID, repositories[0].RepositoryID)
}

func (m *managerTestSuite) TestGet() {
	repository := &models.RepoRecord{
		RepositoryID: 1,
		ProjectID:    1,
		Name:         "library/hello-world",
	}
	m.dao.On("Get", mock.Anything).Return(repository, nil)
	repo, err := m.mgr.Get(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Require().NotNil(repo)
	m.Equal(repository.RepositoryID, repo.RepositoryID)
}

func (m *managerTestSuite) TestGetByName() {
	repository := &models.RepoRecord{
		RepositoryID: 1,
		ProjectID:    1,
		Name:         "library/hello-world",
	}
	m.dao.On("List", mock.Anything).Return([]*models.RepoRecord{repository}, nil)
	repo, err := m.mgr.GetByName(nil, "library/hello-world")
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Require().NotNil(repo)
	m.Equal(repository.RepositoryID, repo.RepositoryID)
}

func (m *managerTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything).Return(1, nil)
	id, err := m.mgr.Create(nil, &models.RepoRecord{
		ProjectID: 1,
		Name:      "library/hello-world",
	})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Equal(int64(1), id)
}

func (m *managerTestSuite) TestDelete() {
	m.dao.On("Delete", mock.Anything).Return(nil)
	err := m.mgr.Delete(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestUpdate() {
	m.dao.On("Update", mock.Anything).Return(nil)
	err := m.mgr.Update(nil, &models.RepoRecord{
		RepositoryID: 1,
	})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestAddPullCount() {
	m.dao.On("AddPullCount", mock.Anything).Return(nil)
	err := m.mgr.AddPullCount(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
