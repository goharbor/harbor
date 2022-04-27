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

package dao

import (
	"context"
	"testing"

	beegoorm "github.com/beego/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	artdao "github.com/goharbor/harbor/src/pkg/artifact/dao"
	"github.com/goharbor/harbor/src/pkg/label/model"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
)

type labelDaoTestSuite struct {
	suite.Suite
	dao    DAO
	artDAO artdao.DAO
	ctx    context.Context
	artID  int64
	id     int64
	refID  int64
}

func (l *labelDaoTestSuite) SetupSuite() {
	common_dao.PrepareTestForPostgresSQL()
	l.dao = &defaultDAO{}
	l.artDAO = artdao.New()
	l.ctx = orm.NewContext(nil, beegoorm.NewOrm())
}

func (l *labelDaoTestSuite) SetupTest() {
	id, err := l.dao.Create(l.ctx, &model.Label{
		Name:  "label_for_label_dao_test_suite",
		Scope: "g",
	})
	l.Require().Nil(err)
	l.id = id

	id, err = l.artDAO.Create(l.ctx, &artdao.Artifact{
		Type:              "IMAGE",
		MediaType:         v1.MediaTypeImageConfig,
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         1,
		RepositoryID:      1,
		Digest:            "sha256",
	})
	l.Require().Nil(err)
	l.artID = id

	id, err = l.dao.CreateReference(l.ctx, &model.Reference{
		LabelID:    l.id,
		ArtifactID: l.artID,
	})
	l.Require().Nil(err)
	l.refID = id
}

func (l *labelDaoTestSuite) TearDownTest() {
	err := l.dao.DeleteReference(l.ctx, l.refID)
	l.Require().Nil(err)

	err = l.dao.Delete(l.ctx, l.id)
	l.Require().Nil(err)

	err = l.artDAO.Delete(l.ctx, l.artID)
	l.Require().Nil(err)
}

func (l *labelDaoTestSuite) TestGet() {
	// not found
	_, err := l.dao.Get(l.ctx, 1000)
	l.Require().NotNil(err)
	l.True(errors.IsErr(err, errors.NotFoundCode))

	// success
	label, err := l.dao.Get(l.ctx, l.id)
	l.Require().Nil(err)
	l.Equal(l.id, label.ID)
}

func (l *labelDaoTestSuite) TestCreate() {
	// happy pass is covered by SetupTest

	// conflict
	_, err := l.dao.Create(l.ctx, &model.Label{
		Name:  "label_for_label_dao_test_suite",
		Scope: "g",
	})
	l.Require().NotNil(err)
	l.True(errors.IsErr(err, errors.ConflictCode))
}

func (l *labelDaoTestSuite) TestDelete() {
	// happy pass is covered by TearDownTest

	// not found
	err := l.dao.Delete(l.ctx, 1000)
	l.Require().NotNil(err)
	l.True(errors.IsErr(err, errors.NotFoundCode))
}

func (l *labelDaoTestSuite) TestListByResource() {
	labels, err := l.dao.ListByArtifact(l.ctx, l.artID)
	l.Require().Nil(err)
	l.Require().Len(labels, 1)
	l.Equal(l.id, labels[0].ID)
}

func (l *labelDaoTestSuite) TestCreateReference() {
	// happy pass is covered by SetupTest

	// conflict
	_, err := l.dao.CreateReference(l.ctx, &model.Reference{
		LabelID:    l.id,
		ArtifactID: l.artID,
	})
	l.Require().NotNil(err)
	l.True(errors.IsErr(err, errors.ConflictCode))

	// violating foreign key constraint: the label that the ref tries to refer doesn't exist
	_, err = l.dao.CreateReference(l.ctx, &model.Reference{
		LabelID:    1000,
		ArtifactID: l.artID,
	})
	l.Require().NotNil(err)
	l.True(errors.IsErr(err, errors.NotFoundCode))

	// violating foreign key constraint: the artifact that the ref tries to refer doesn't exist
	_, err = l.dao.CreateReference(l.ctx, &model.Reference{
		LabelID:    l.id,
		ArtifactID: 1000,
	})
	l.Require().NotNil(err)
	l.True(errors.IsErr(err, errors.NotFoundCode))
}

func (l *labelDaoTestSuite) DeleteReference() {
	// happy pass is covered by TearDownTest

	// not found
	err := l.dao.DeleteReference(l.ctx, 1000)
	l.Require().NotNil(err)
	l.True(errors.IsErr(err, errors.NotFoundCode))
}

func (l *labelDaoTestSuite) DeleteReferences() {
	n, err := l.dao.DeleteReferences(l.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"LabelID": 1000,
		},
	})
	l.Require().Nil(err)
	l.Equal(int64(0), n)
}

func TestLabelDaoTestSuite(t *testing.T) {
	suite.Run(t, &labelDaoTestSuite{})
}
