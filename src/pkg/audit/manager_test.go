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

package audit

import (
	"context"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/audit/model"
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
func (f *fakeDao) List(ctx context.Context, query *q.Query) ([]*model.AuditLog, error) {
	args := f.Called()
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}
func (f *fakeDao) Get(ctx context.Context, id int64) (*model.AuditLog, error) {
	args := f.Called()
	return args.Get(0).(*model.AuditLog), args.Error(1)
}
func (f *fakeDao) Create(ctx context.Context, repository *model.AuditLog) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) Delete(ctx context.Context, id int64) error {
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
	audit := &model.AuditLog{
		ProjectID:    1,
		Resource:     "library/hello-world",
		ResourceType: "artifact",
	}
	m.dao.On("List", mock.Anything).Return([]*model.AuditLog{audit}, nil)
	auditLogs, err := m.mgr.List(nil, nil)
	m.Require().Nil(err)
	m.Equal(1, len(auditLogs))
	m.Equal(audit.Resource, auditLogs[0].Resource)
}

func (m *managerTestSuite) TestGet() {
	audit := &model.AuditLog{
		ProjectID:    1,
		Resource:     "library/hello-world",
		ResourceType: "artifact",
	}
	m.dao.On("Get", mock.Anything).Return(audit, nil)
	au, err := m.mgr.Get(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Require().NotNil(au)
	m.Equal(audit.Resource, au.Resource)
}

func (m *managerTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything).Return(1, nil)
	id, err := m.mgr.Create(nil, &model.AuditLog{
		ProjectID:    1,
		Resource:     "library/hello-world",
		ResourceType: "artifact",
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

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
