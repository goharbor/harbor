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

package policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/modelsync/policy/model"
)

type fakeDAO struct {
	mock.Mock
}

func (f *fakeDAO) Count(ctx context.Context, query *q.Query) (int64, error) {
	args := f.Called(ctx, query)
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDAO) List(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	args := f.Called(ctx, query)
	return args.Get(0).([]*model.Policy), args.Error(1)
}
func (f *fakeDAO) Get(ctx context.Context, id int64) (*model.Policy, error) {
	args := f.Called(ctx, id)
	return args.Get(0).(*model.Policy), args.Error(1)
}
func (f *fakeDAO) Create(ctx context.Context, policy *model.Policy) (int64, error) {
	args := f.Called(ctx, policy)
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDAO) Update(ctx context.Context, policy *model.Policy, props ...string) error {
	args := f.Called(ctx, policy, props)
	return args.Error(0)
}
func (f *fakeDAO) Delete(ctx context.Context, id int64) error {
	args := f.Called(ctx, id)
	return args.Error(0)
}

type managerTestSuite struct {
	suite.Suite
	mgr *manager
	dao *fakeDAO
}

func (m *managerTestSuite) SetupTest() {
	m.dao = &fakeDAO{}
	m.mgr = &manager{dao: m.dao}
}

func (m *managerTestSuite) TestAll() {
	ctx := context.Background()
	m.dao.On("Count", mock.Anything, mock.Anything).Return(1, nil)
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.Policy{{ID: 1}}, nil)
	m.dao.On("Get", mock.Anything, int64(1)).Return(&model.Policy{ID: 1}, nil)
	m.dao.On("Create", mock.Anything, mock.Anything).Return(1, nil)
	m.dao.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	m.dao.On("Delete", mock.Anything, int64(1)).Return(nil)

	n, err := m.mgr.Count(ctx, nil)
	m.NoError(err)
	m.Equal(int64(1), n)
	policies, err := m.mgr.List(ctx, nil)
	m.NoError(err)
	m.Len(policies, 1)
	p, err := m.mgr.Get(ctx, 1)
	m.NoError(err)
	m.Equal(int64(1), p.ID)
	id, err := m.mgr.Create(ctx, &model.Policy{})
	m.NoError(err)
	m.Equal(int64(1), id)
	m.NoError(m.mgr.Update(ctx, &model.Policy{ID: 1}, "Enabled"))
	m.NoError(m.mgr.Delete(ctx, 1))
	m.dao.AssertExpectations(m.T())
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
