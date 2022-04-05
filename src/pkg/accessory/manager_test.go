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

package accessory

import (
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/accessory/dao"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/testing/mock"
	testingdao "github.com/goharbor/harbor/src/testing/pkg/accessory/dao"
	"github.com/stretchr/testify/suite"
	"testing"
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

func (m *managerTestSuite) TestEnsure() {
	mock.OnAnything(m.dao, "List").Return([]*dao.Accessory{}, nil)
	mock.OnAnything(m.dao, "Create").Return(int64(1), nil)
	err := m.mgr.Ensure(nil, int64(1), int64(1), int64(1), "sha256:1234", model.TypeCosignSignature)
	m.Require().Nil(err)
}

func (m *managerTestSuite) TestList() {
	acc := &dao.Accessory{
		ID:   1,
		Type: model.TypeCosignSignature,
	}
	mock.OnAnything(m.dao, "List").Return([]*dao.Accessory{
		acc,
	}, nil)
	accs, err := m.mgr.List(nil, nil)
	m.Require().Nil(err)
	m.Require().Equal(1, len(accs))
	m.Equal(int64(1), accs[0].GetData().ID)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGet() {
	acc := &dao.Accessory{
		ID:   1,
		Type: model.TypeCosignSignature,
	}
	mock.OnAnything(m.dao, "Get").Return(acc, nil)
	accessory, err := m.mgr.Get(nil, 1)
	m.Require().Nil(err)
	m.Equal(int64(1), accessory.GetData().ID)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestCreate() {
	mock.OnAnything(m.dao, "Create").Return(int64(1), nil)
	_, err := m.mgr.Create(nil, model.AccessoryData{
		ArtifactID: 1,
		Size:       1,
		Type:       model.TypeCosignSignature,
	})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDelete() {
	mock.OnAnything(m.dao, "Delete").Return(nil)
	err := m.mgr.Delete(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestCount() {
	mock.OnAnything(m.dao, "Count").Return(int64(1), nil)
	n, err := m.mgr.Count(nil, nil)
	m.Require().Nil(err)
	m.Equal(int64(1), n)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDeleteOfArtifact() {
	mock.OnAnything(m.dao, "DeleteAccessories").Return(int64(1), nil)
	err := m.mgr.DeleteAccessories(nil, q.New(q.KeyWords{"ArtifactID": 1}))
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGetIcon() {
	var icon string
	icon = m.mgr.GetIcon("")
	m.Require().Empty(icon, "empty icon")
	icon = m.mgr.GetIcon("signature.cosign")
	m.Require().Equal("sha256:20401d5b3a0f6dbc607c8d732eb08471af4ae6b19811a4efce8c6a724aed2882", icon)
	icon = m.mgr.GetIcon("unknown")
	m.Require().Empty(icon, "empty icon")
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
