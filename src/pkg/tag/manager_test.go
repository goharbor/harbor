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

package tag

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type fakeDao struct {
	mock.Mock
}

func (f *fakeDao) Count(ctx context.Context, query *q.Query) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) List(ctx context.Context, query *q.Query) ([]*tag.Tag, error) {
	args := f.Called()
	return args.Get(0).([]*tag.Tag), args.Error(1)
}
func (f *fakeDao) Get(ctx context.Context, id int64) (*tag.Tag, error) {
	args := f.Called()
	return args.Get(0).(*tag.Tag), args.Error(1)
}
func (f *fakeDao) Create(ctx context.Context, tag *tag.Tag) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) GetOrCreate(ctx context.Context, tag *tag.Tag) (bool, int64, error) {
	args := f.Called()
	return args.Bool(0), int64(args.Int(1)), args.Error(2)
}
func (f *fakeDao) Update(ctx context.Context, tag *tag.Tag, props ...string) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) DeleteOfArtifact(ctx context.Context, artifactID int64) error {
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
	tg := &tag.Tag{
		ID:           1,
		RepositoryID: 1,
		ArtifactID:   1,
		Name:         "latest",
		PushTime:     time.Now(),
		PullTime:     time.Now(),
	}
	m.dao.On("List", mock.Anything).Return([]*tag.Tag{tg}, nil)
	tags, err := m.mgr.List(nil, nil)
	m.Require().Nil(err)
	m.Equal(1, len(tags))
	m.Equal(tg.ID, tags[0].ID)
}

func (m *managerTestSuite) TestGet() {
	m.dao.On("Get", mock.Anything).Return(&tag.Tag{}, nil)
	_, err := m.mgr.Get(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything).Return(1, nil)
	_, err := m.mgr.Create(nil, nil)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestGetOrCreate() {
	m.dao.On("GetOrCreate", mock.Anything).Return(false, 1, nil)
	created, id, err := m.mgr.GetOrCreate(nil, nil)
	m.Require().Nil(err)
	m.False(created)
	m.Equal(int64(1), id)
}

func (m *managerTestSuite) TestUpdate() {
	m.dao.On("Update", mock.Anything).Return(nil)
	err := m.mgr.Update(nil, nil)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDelete() {
	m.dao.On("Delete", mock.Anything).Return(nil)
	err := m.mgr.Delete(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestDeleteOfArtifact() {
	m.dao.On("DeleteOfArtifact", mock.Anything).Return(nil)
	err := m.mgr.DeleteOfArtifact(nil, 1)
	m.Require().Nil(err)
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
