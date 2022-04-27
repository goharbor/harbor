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
	"fmt"
	"testing"
	"time"

	beegoorm "github.com/beego/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	af_dao "github.com/goharbor/harbor/src/pkg/artifact/dao"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	tag_dao "github.com/goharbor/harbor/src/pkg/tag/dao"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
)

var (
	repository = fmt.Sprintf("library/%d", time.Now().Unix())
)

type daoTestSuite struct {
	suite.Suite
	dao    DAO
	tagDao tag_dao.DAO
	afDao  af_dao.DAO
	id     int64
	ctx    context.Context
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = New()
	d.tagDao = tag_dao.New()
	d.afDao = af_dao.New()
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.NewContext(nil, beegoorm.NewOrm())
}

func (d *daoTestSuite) SetupTest() {
	repository := &model.RepoRecord{
		Name:        repository,
		ProjectID:   1,
		Description: "",
	}
	id, err := d.dao.Create(d.ctx, repository)
	d.Require().Nil(err)
	d.id = id
}

func (d *daoTestSuite) TearDownTest() {
	err := d.dao.Delete(d.ctx, d.id)
	d.Require().Nil(err)
}

func (d *daoTestSuite) TestCount() {
	// nil query
	total, err := d.dao.Count(d.ctx, nil)
	d.Require().Nil(err)
	d.True(total > 0)

	// query by name
	total, err = d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"name": repository,
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)
}

func (d *daoTestSuite) TestList() {
	// nil query
	repositories, err := d.dao.List(d.ctx, nil)
	d.Require().Nil(err)
	found := false
	for _, repository := range repositories {
		if repository.RepositoryID == d.id {
			found = true
			break
		}
	}
	d.True(found)

	// query by name
	repositories, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"name": repository,
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(1, len(repositories))
	d.Equal(d.id, repositories[0].RepositoryID)
}

func (d *daoTestSuite) TestGet() {
	// get the non-exist repository
	_, err := d.dao.Get(d.ctx, 10000)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))

	// get the exist repository
	repository, err := d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Require().NotNil(repository)
	d.Equal(d.id, repository.RepositoryID)
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass case is covered in Setup

	// conflict
	repository := &model.RepoRecord{
		Name:      repository,
		ProjectID: 1,
	}
	_, err := d.dao.Create(d.ctx, repository)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ConflictCode))
}

func (d *daoTestSuite) TestDelete() {
	// the happy pass case is covered in TearDown

	// not exist
	err := d.dao.Delete(d.ctx, 100021)
	d.Require().NotNil(err)
	var e *errors.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(errors.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestUpdate() {
	// pass
	err := d.dao.Update(d.ctx, &model.RepoRecord{
		RepositoryID: d.id,
		PullCount:    1,
	}, "PullCount")
	d.Require().Nil(err)

	repository, err := d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Require().NotNil(repository)
	d.Equal(int64(1), repository.PullCount)

	// not exist
	err = d.dao.Update(d.ctx, &model.RepoRecord{
		RepositoryID: 10000,
	})
	d.Require().NotNil(err)
	var e *errors.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(errors.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestAddPullCount() {
	repository := &model.RepoRecord{
		Name:        "test/pullcount",
		ProjectID:   10,
		Description: "test pull count",
		PullCount:   1,
	}
	id, err := d.dao.Create(d.ctx, repository)
	d.Require().Nil(err)

	err = d.dao.AddPullCount(d.ctx, id, 1)
	d.Require().Nil(err)

	repository, err = d.dao.Get(d.ctx, id)
	d.Require().Nil(err)
	d.Require().NotNil(repository)
	d.Equal(int64(2), repository.PullCount)

	d.dao.Delete(d.ctx, id)
}

func (d *daoTestSuite) TestNonEmptyRepos() {
	repository := &model.RepoRecord{
		Name:        "TestNonEmptyRepos",
		ProjectID:   10,
		Description: "test pull count",
		PullCount:   1,
	}
	id, err := d.dao.Create(d.ctx, repository)
	d.Require().Nil(err)

	art := &af_dao.Artifact{
		Type:              "IMAGE",
		MediaType:         v1.MediaTypeImageConfig,
		ManifestMediaType: v1.MediaTypeImageIndex,
		ProjectID:         1,
		RepositoryID:      1,
		RepositoryName:    "library/hello-world",
		Digest:            "parent_digest",
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		Annotations:       `{"anno1":"value1"}`,
	}
	afID, err := d.afDao.Create(d.ctx, art)
	d.Require().Nil(err)

	// Same repository with two tags, the result should only contain one repository record.
	tag1 := &tag.Tag{
		RepositoryID: id,
		ArtifactID:   afID,
		Name:         "tag1",
		PushTime:     time.Now(),
		PullTime:     time.Now(),
	}
	_, err = d.tagDao.Create(d.ctx, tag1)
	d.Require().Nil(err)
	tag2 := &tag.Tag{
		RepositoryID: id,
		ArtifactID:   afID,
		Name:         "tag2",
		PushTime:     time.Now(),
		PullTime:     time.Now(),
	}
	_, err = d.tagDao.Create(d.ctx, tag2)
	d.Require().Nil(err)

	repos, err := d.dao.NonEmptyRepos(d.ctx)
	d.Require().Nil(err)

	var success bool
	var count int
	for _, repo := range repos {
		if repo.Name == "TestNonEmptyRepos" {
			success = true
			count++
		}
	}
	if !success {
		d.Fail("TestNonEmptyRepos failure: no NonEmpty repository found")
	}

	if count != 1 {
		d.Fail("TestNonEmptyRepos failure: duplicate repository record")
	}

}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
