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

package label

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/label/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/label/dao"
	"github.com/stretchr/testify/suite"
	"testing"
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

func (m *managerTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	_, err := m.mgr.Create(context.Background(), &model.Label{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestCount() {
	m.dao.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	n, err := m.mgr.Count(context.Background(), nil)
	m.Nil(err)
	m.Equal(int64(1), n)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDelete() {
	m.dao.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.Delete(context.Background(), 1)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGet() {
	m.dao.On("Get", mock.Anything, mock.Anything).Return(&model.Label{
		ID:   1,
		Name: "label",
	}, nil)
	label, err := m.mgr.Get(context.Background(), 1)
	m.Nil(err)
	m.Equal("label", label.Name)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestUpdate() {
	m.dao.On("Update", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.Update(context.Background(), &model.Label{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestListByArtifact() {
	m.dao.On("ListByArtifact", mock.Anything, mock.Anything).Return([]*model.Label{
		{
			ID:   1,
			Name: "label",
		},
	}, nil)
	rpers, err := m.mgr.ListByArtifact(context.Background(), 1)
	m.Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestList() {
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.Label{
		{
			ID:   1,
			Name: "label",
		},
	}, nil)
	rpers, err := m.mgr.List(context.Background(), nil)
	m.Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestAddTo() {
	m.dao.On("CreateReference", mock.Anything, mock.Anything).Return(int64(1), nil)
	err := m.mgr.AddTo(context.Background(), 1, 1)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestRemoveFrom() {
	m.dao.On("DeleteReferences", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	err := m.mgr.RemoveFrom(context.Background(), 1, 1)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestRemoveAllFrom() {
	m.dao.On("DeleteReferences", mock.Anything, mock.Anything).Return(int64(1), nil)
	err := m.mgr.RemoveAllFrom(context.Background(), 1)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestRemoveFromAllArtifacts() {
	m.dao.On("DeleteReferences", mock.Anything, mock.Anything).Return(int64(1), nil)
	err := m.mgr.RemoveFromAllArtifacts(context.Background(), 1)
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
