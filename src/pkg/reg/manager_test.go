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

package reg

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/reg/dao"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/testing/mock"
	testingdao "github.com/goharbor/harbor/src/testing/pkg/reg/dao"
	"github.com/stretchr/testify/suite"
)

type managerTestSuite struct {
	suite.Suite
	mgr *manager
	dao *testingdao.DAO
}

func (m *managerTestSuite) SetupTest() {
	m.dao = &testingdao.DAO{}
	m.mgr = &manager{
		dao: m.dao,
	}
}

func (m *managerTestSuite) TestCount() {
	mock.OnAnything(m.dao, "Count").Return(int64(1), nil)
	n, err := m.mgr.Count(nil, nil)
	m.Require().Nil(err)
	m.Equal(int64(1), n)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestList() {
	mock.OnAnything(m.dao, "List").Return([]*dao.Registry{
		{
			ID: 1,
		},
	}, nil)
	registries, err := m.mgr.List(nil, nil)
	m.Require().Nil(err)
	m.Require().Equal(1, len(registries))
	m.Equal(int64(1), registries[0].ID)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGet() {
	mock.OnAnything(m.dao, "Get").Return(&dao.Registry{
		ID: 1,
	}, nil)
	registry, err := m.mgr.Get(nil, 1)
	m.Require().Nil(err)
	m.Equal(int64(1), registry.ID)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestCreate() {
	mock.OnAnything(m.dao, "Create").Return(int64(1), nil)
	_, err := m.mgr.Create(nil, &model.Registry{})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDelete() {
	mock.OnAnything(m.dao, "Delete").Return(nil)
	err := m.mgr.Delete(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestUpdate() {
	mock.OnAnything(m.dao, "Update").Return(nil)
	err := m.mgr.Update(nil, &model.Registry{})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
