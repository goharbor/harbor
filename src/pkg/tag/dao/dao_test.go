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
	"errors"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	repositoryID int64 = 1000
	artifactID   int64 = 1000
	name               = "latest"
)

type daoTestSuite struct {
	suite.Suite
	dao   DAO
	tagID int64
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = New()
	common_dao.PrepareTestForPostgresSQL()
}

func (d *daoTestSuite) SetupTest() {
	tag := &tag.Tag{
		RepositoryID: repositoryID,
		ArtifactID:   artifactID,
		Name:         name,
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}
	id, err := d.dao.Create(nil, tag)
	d.Require().Nil(err)
	d.tagID = id
}

func (d *daoTestSuite) TearDownTest() {
	err := d.dao.Delete(nil, d.tagID)
	d.Require().Nil(err)
}

func (d *daoTestSuite) TestCount() {
	// nil query
	total, err := d.dao.Count(nil, nil)
	d.Require().Nil(err)
	d.True(total > 0)
	// query by repository ID and name
	total, err = d.dao.Count(nil, &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"name":          name,
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)
}

func (d *daoTestSuite) TestList() {
	// nil query
	tags, err := d.dao.List(nil, nil)
	d.Require().Nil(err)
	found := false
	for _, tag := range tags {
		if tag.ID == d.tagID {
			found = true
			break
		}
	}
	d.True(found)

	// query by repository ID and name
	tags, err = d.dao.List(nil, &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"name":          name,
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(1, len(tags))
	d.Equal(d.tagID, tags[0].ID)
}

func (d *daoTestSuite) TestGet() {
	// get the non-exist tag
	_, err := d.dao.Get(nil, 10000)
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.NotFoundCode))

	// get the exist tag
	tag, err := d.dao.Get(nil, d.tagID)
	d.Require().Nil(err)
	d.Require().NotNil(tag)
	d.Equal(d.tagID, tag.ID)
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass case is covered in Setup

	// conflict
	tag := &tag.Tag{
		RepositoryID: repositoryID,
		ArtifactID:   artifactID,
		Name:         name,
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}
	_, err := d.dao.Create(nil, tag)
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.ConflictCode))
}

func (d *daoTestSuite) TestDelete() {
	// happy pass is covered in TearDown

	// not exist
	err := d.dao.Delete(nil, 10000)
	d.Require().NotNil(err)
	var e *ierror.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(ierror.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestUpdate() {
	// pass
	err := d.dao.Update(nil, &tag.Tag{
		ID:         d.tagID,
		ArtifactID: 2,
	}, "ArtifactID")
	d.Require().Nil(err)

	tg, err := d.dao.Get(nil, d.tagID)
	d.Require().Nil(err)
	d.Require().NotNil(tg)
	d.Equal(int64(2), tg.ArtifactID)

	// not exist
	err = d.dao.Update(nil, &tag.Tag{
		ID: 10000,
	})
	d.Require().NotNil(err)
	var e *ierror.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(ierror.NotFoundCode, e.Code)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
