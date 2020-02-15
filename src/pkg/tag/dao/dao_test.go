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
	"errors"
	beegoorm "github.com/astaxie/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/internal/orm"
	artdao "github.com/goharbor/harbor/src/pkg/artifact/dao"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type daoTestSuite struct {
	suite.Suite
	dao        DAO
	artDAO     artdao.DAO
	tagID      int64
	artifactID int64
	ctx        context.Context
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = New()
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.NewContext(nil, beegoorm.NewOrm())
	d.artDAO = artdao.New()
	artifactID, err := d.artDAO.Create(d.ctx, &artdao.Artifact{
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1000,
		Digest:            "sha256:digest",
	})
	d.Require().Nil(err)
	d.artifactID = artifactID
}

func (d *daoTestSuite) TearDownSuite() {
	err := d.artDAO.Delete(d.ctx, d.artifactID)
	d.Require().Nil(err)
}

func (d *daoTestSuite) SetupTest() {
	tag := &tag.Tag{
		RepositoryID: 1000,
		ArtifactID:   d.artifactID,
		Name:         "latest",
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}
	id, err := d.dao.Create(d.ctx, tag)
	d.Require().Nil(err)
	d.tagID = id
}

func (d *daoTestSuite) TearDownTest() {
	err := d.dao.Delete(d.ctx, d.tagID)
	d.Require().Nil(err)
}

func (d *daoTestSuite) TestCount() {
	// nil query
	total, err := d.dao.Count(d.ctx, nil)
	d.Require().Nil(err)
	d.True(total > 0)
	// query by repository ID and name
	total, err = d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": 1000,
			"name":          "latest",
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)
}

func (d *daoTestSuite) TestList() {
	// nil query
	tags, err := d.dao.List(d.ctx, nil)
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
	tags, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": 1000,
			"name":          "latest",
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(1, len(tags))
	d.Equal(d.tagID, tags[0].ID)
}

func (d *daoTestSuite) TestGet() {
	// get the non-exist tag
	_, err := d.dao.Get(d.ctx, 10000)
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.NotFoundCode))

	// get the exist tag
	tag, err := d.dao.Get(d.ctx, d.tagID)
	d.Require().Nil(err)
	d.Require().NotNil(tag)
	d.Equal(d.tagID, tag.ID)
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass case is covered in Setup

	// conflict
	tg := &tag.Tag{
		RepositoryID: 1000,
		ArtifactID:   d.artifactID,
		Name:         "latest",
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}
	_, err := d.dao.Create(d.ctx, tg)
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.ConflictCode))

	// violating foreign key constraint: the artifact that the tag tries to attach doesn't exist
	tg = &tag.Tag{
		RepositoryID: 1000,
		ArtifactID:   1000,
		Name:         "latest2",
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}
	_, err = d.dao.Create(d.ctx, tg)
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.ViolateForeignKeyConstraintCode))
}

func (d *daoTestSuite) TestDelete() {
	// happy pass is covered in TearDown

	// not exist
	err := d.dao.Delete(d.ctx, 10000)
	d.Require().NotNil(err)
	var e *ierror.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(ierror.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestUpdate() {
	// pass
	artifactID, err := d.artDAO.Create(d.ctx, &artdao.Artifact{
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1000,
		Digest:            "sha256:digest2",
	})
	d.Require().Nil(err)
	defer func() {
		err := d.artDAO.Delete(d.ctx, artifactID)
		d.Require().Nil(err)
	}()

	err = d.dao.Update(d.ctx, &tag.Tag{
		ID:         d.tagID,
		ArtifactID: artifactID,
	}, "ArtifactID")
	d.Require().Nil(err)

	tg, err := d.dao.Get(d.ctx, d.tagID)
	d.Require().Nil(err)
	d.Require().NotNil(tg)
	d.Equal(artifactID, tg.ArtifactID)

	err = d.dao.Update(d.ctx, &tag.Tag{
		ID:         d.tagID,
		ArtifactID: d.artifactID,
	}, "ArtifactID")
	d.Require().Nil(err)

	// violating foreign key constraint: the artifact that the tag tries to attach doesn't exist
	err = d.dao.Update(d.ctx, &tag.Tag{
		ID:         d.tagID,
		ArtifactID: 2,
	}, "ArtifactID")
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.ViolateForeignKeyConstraintCode))

	// not exist
	err = d.dao.Update(d.ctx, &tag.Tag{
		ID: 10000,
	})
	d.Require().NotNil(err)
	var e *ierror.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(ierror.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestDeleteOfArtifact() {
	artifactID, err := d.artDAO.Create(d.ctx, &artdao.Artifact{
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1000,
		Digest:            "sha256:digest02",
	})
	d.Require().Nil(err)
	defer d.artDAO.Delete(d.ctx, artifactID)

	tag1 := &tag.Tag{
		RepositoryID: 1000,
		ArtifactID:   artifactID,
		Name:         "tag1",
	}
	_, err = d.dao.Create(d.ctx, tag1)
	d.Require().Nil(err)
	tag2 := &tag.Tag{
		RepositoryID: 1000,
		ArtifactID:   artifactID,
		Name:         "tag2",
	}
	_, err = d.dao.Create(d.ctx, tag2)
	d.Require().Nil(err)

	tags, err := d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ArtifactID": artifactID,
		},
	})
	d.Require().Nil(err)
	d.Require().Len(tags, 2)

	err = d.dao.DeleteOfArtifact(d.ctx, artifactID)
	d.Require().Nil(err)

	tags, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ArtifactID": artifactID,
		},
	})
	d.Require().Nil(err)
	d.Require().Len(tags, 0)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
