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

package artifact

import (
	"context"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/artifact/dao"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type fakeDao struct {
	mock.Mock
}

func (f *fakeDao) Count(ctx context.Context, query *q.Query) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) List(ctx context.Context, query *q.Query) ([]*dao.Artifact, error) {
	args := f.Called()
	return args.Get(0).([]*dao.Artifact), args.Error(1)
}
func (f *fakeDao) Get(ctx context.Context, id int64) (*dao.Artifact, error) {
	args := f.Called()
	return args.Get(0).(*dao.Artifact), args.Error(1)
}
func (f *fakeDao) GetByDigest(ctx context.Context, repository, digest string) (*dao.Artifact, error) {
	args := f.Called()
	return args.Get(0).(*dao.Artifact), args.Error(1)
}
func (f *fakeDao) Create(ctx context.Context, artifact *dao.Artifact) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) Delete(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) Update(ctx context.Context, artifact *dao.Artifact, props ...string) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) UpdatePullTime(ctx context.Context, id int64, pullTime time.Time) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) CreateReference(ctx context.Context, reference *dao.ArtifactReference) (int64, error) {
	args := f.Called()
	return int64(args.Int(0)), args.Error(1)
}
func (f *fakeDao) ListReferences(ctx context.Context, query *q.Query) ([]*dao.ArtifactReference, error) {
	args := f.Called()
	return args.Get(0).([]*dao.ArtifactReference), args.Error(1)
}
func (f *fakeDao) DeleteReference(ctx context.Context, id int64) error {
	args := f.Called()
	return args.Error(0)
}
func (f *fakeDao) DeleteReferences(ctx context.Context, parentID int64) error {
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

func (m *managerTestSuite) TestAssemble() {
	art := &dao.Artifact{
		ID:                1,
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: v1.MediaTypeImageIndex,
		ProjectID:         1,
		RepositoryID:      1,
		Digest:            "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
		Size:              1024,
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		ExtraAttrs:        `{"attr1":"value1"}`,
		Annotations:       `{"anno1":"value1"}`,
	}
	m.dao.On("ListReferences").Return([]*dao.ArtifactReference{
		{
			ID:       1,
			ParentID: 1,
			ChildID:  2,
		},
		{
			ID:       2,
			ParentID: 1,
			ChildID:  3,
		},
	}, nil)
	artifact, err := m.mgr.assemble(nil, art)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Require().NotNil(artifact)
	m.Equal(art.ID, artifact.ID)
	m.Equal(2, len(artifact.References))
}

func (m *managerTestSuite) TestList() {
	art := &dao.Artifact{
		ID:                1,
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1,
		Digest:            "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
		Size:              1024,
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		ExtraAttrs:        `{"attr1":"value1"}`,
		Annotations:       `{"anno1":"value1"}`,
	}
	m.dao.On("List", mock.Anything).Return([]*dao.Artifact{art}, nil)
	m.dao.On("ListReferences").Return([]*dao.ArtifactReference{}, nil)
	artifacts, err := m.mgr.List(nil, nil)
	m.Require().Nil(err)
	m.Equal(1, len(artifacts))
	m.Equal(art.ID, artifacts[0].ID)
}

func (m *managerTestSuite) TestGet() {
	art := &dao.Artifact{
		ID:                1,
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: v1.MediaTypeImageIndex,
		ProjectID:         1,
		RepositoryID:      1,
		Digest:            "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
		Size:              1024,
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		ExtraAttrs:        `{"attr1":"value1"}`,
		Annotations:       `{"anno1":"value1"}`,
	}
	m.dao.On("Get", mock.Anything).Return(art, nil)
	m.dao.On("ListReferences").Return([]*dao.ArtifactReference{}, nil)
	artifact, err := m.mgr.Get(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Require().NotNil(artifact)
	m.Equal(art.ID, artifact.ID)
}

func (m *managerTestSuite) TestGetByDigest() {
	art := &dao.Artifact{
		ID:                1,
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1,
		Digest:            "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
		Size:              1024,
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		ExtraAttrs:        `{"attr1":"value1"}`,
		Annotations:       `{"anno1":"value1"}`,
	}
	m.dao.On("GetByDigest", mock.Anything).Return(art, nil)
	m.dao.On("ListReferences").Return([]*dao.ArtifactReference{}, nil)
	artifact, err := m.mgr.GetByDigest(nil, "library/hello-world", "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180")
	m.Require().Nil(err)
	m.Require().NotNil(artifact)
	m.Equal(art.ID, artifact.ID)
}

func (m *managerTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything).Return(1, nil)
	m.dao.On("CreateReference").Return(1, nil)
	id, err := m.mgr.Create(nil, &Artifact{
		References: []*Reference{
			{
				ChildID: 2,
			},
		},
	})
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
	m.Equal(int64(1), id)
}

func (m *managerTestSuite) TestDelete() {
	m.dao.On("Delete", mock.Anything).Return(nil)
	m.dao.On("DeleteReferences").Return(nil)
	err := m.mgr.Delete(nil, 1)
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestUpdate() {
	m.dao.On("Update", mock.Anything).Return(nil)
	err := m.mgr.Update(nil, &Artifact{
		ID:       1,
		PullTime: time.Now(),
	}, "PullTime")
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestUpdatePullTime() {
	m.dao.On("UpdatePullTime", mock.Anything).Return(nil)
	err := m.mgr.UpdatePullTime(nil, 1, time.Now())
	m.Require().Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestListReferences() {
	m.dao.On("ListReferences").Return([]*dao.ArtifactReference{
		{
			ID:       1,
			ParentID: 1,
			ChildID:  2,
		},
	}, nil)
	references, err := m.mgr.ListReferences(nil, nil)
	m.Require().Nil(err)
	m.Require().Len(references, 1)
	m.Equal(int64(1), references[0].ID)
}

func (m *managerTestSuite) TestDeleteReference() {
	m.dao.On("DeleteReference").Return(nil)
	err := m.mgr.DeleteReference(nil, 1)
	m.Require().Nil(err)
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
