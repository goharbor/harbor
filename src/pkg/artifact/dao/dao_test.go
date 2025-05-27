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
	"time"

	beegoorm "github.com/beego/beego/v2/client/orm"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	tagdao "github.com/goharbor/harbor/src/pkg/tag/dao"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
)

type daoTestSuite struct {
	suite.Suite
	dao           DAO
	tagDAO        tagdao.DAO
	parentArtID   int64
	childArt01ID  int64
	childArt02ID  int64
	reference01ID int64
	reference02ID int64
	tagID         int64
	ctx           context.Context
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = New()
	d.tagDAO = tagdao.New()
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.NewContext(nil, beegoorm.NewOrm())
}

func (d *daoTestSuite) SetupTest() {
	now := time.Now()
	parentArt := &Artifact{
		Type:              "IMAGE",
		MediaType:         v1.MediaTypeImageConfig,
		ManifestMediaType: v1.MediaTypeImageIndex,
		ProjectID:         1,
		RepositoryID:      1,
		RepositoryName:    "library/hello-world",
		Digest:            "parent_digest",
		PushTime:          now,
		PullTime:          now,
		Annotations:       `{"anno1":"value1"}`,
	}
	id, err := d.dao.Create(d.ctx, parentArt)
	d.Require().Nil(err)
	d.parentArtID = id

	childArt01 := &Artifact{
		Type:              "IMAGE",
		MediaType:         v1.MediaTypeImageConfig,
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         1,
		RepositoryID:      1,
		RepositoryName:    "library/hello-world",
		Digest:            "child_digest_01",
		Size:              1024,
		PushTime:          now,
		PullTime:          now,
		ExtraAttrs:        `{"attr1":"value1"}`,
	}
	id, err = d.dao.Create(d.ctx, childArt01)
	d.Require().Nil(err)
	d.childArt01ID = id

	childArt02 := &Artifact{
		Type:              "IMAGE",
		MediaType:         v1.MediaTypeImageConfig,
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         1,
		RepositoryID:      1,
		RepositoryName:    "library/hello-world",
		Digest:            "child_digest_02",
		Size:              1024,
		PushTime:          now,
		PullTime:          now,
		ExtraAttrs:        `{"attr1":"value1"}`,
	}
	id, err = d.dao.Create(d.ctx, childArt02)
	d.Require().Nil(err)
	d.childArt02ID = id

	id, err = d.dao.CreateReference(d.ctx, &ArtifactReference{
		ParentID: d.parentArtID,
		ChildID:  d.childArt01ID,
	})
	d.Require().Nil(err)
	d.reference01ID = id

	id, err = d.dao.CreateReference(d.ctx, &ArtifactReference{
		ParentID: d.parentArtID,
		ChildID:  d.childArt02ID,
	})
	d.Require().Nil(err)
	d.reference02ID = id

	id, err = d.tagDAO.Create(d.ctx, &tag.Tag{
		RepositoryID: 1,
		ArtifactID:   d.childArt01ID,
		Name:         "latest",
		PushTime:     now,
		PullTime:     now,
	})
	d.Require().Nil(err)
	d.tagID = id
}

func (d *daoTestSuite) TearDownTest() {
	err := d.dao.DeleteReferences(d.ctx, d.parentArtID)
	d.Require().Nil(err)
	d.tagDAO.Delete(d.ctx, d.tagID)
	d.Require().Nil(err)
	err = d.dao.Delete(d.ctx, d.childArt01ID)
	d.Require().Nil(err)
	err = d.dao.Delete(d.ctx, d.childArt02ID)
	d.Require().Nil(err)
	err = d.dao.Delete(d.ctx, d.parentArtID)
	d.Require().Nil(err)
}

func (d *daoTestSuite) TestCount() {
	// query by repository ID: both tagged and untagged artifacts
	totalOfAll, err := d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
		},
	})
	d.Require().Nil(err)
	d.True(totalOfAll >= 2)

	// only query tagged artifacts
	totalOfTagged, err := d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
			"Tags":         "*",
		},
	})
	d.Require().Nil(err)
	d.Equal(totalOfAll-1, totalOfTagged)

	// only query untagged artifacts
	totalOfUnTagged, err := d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
			"Tags":         "nil",
		},
	})
	d.Require().Nil(err)
	d.Equal(totalOfAll-1, totalOfUnTagged)

	// specific tag value
	total, err := d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
			"Tags":         "latest",
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)

	// query by repository ID and digest
	total, err = d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
			"Digest":       "parent_digest",
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)

	// set pagination in query
	total, err = d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
		},
		PageNumber: 1,
		PageSize:   1,
	})
	d.Require().Nil(err)
	d.True(total > 1)
}

func (d *daoTestSuite) TestList() {
	// query by repository ID: both tagged and untagged artifacts
	artifacts, err := d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
		},
	})
	d.Require().Nil(err)

	parentArtFound := false
	childArt01Found := false
	childArt02Found := false
	for _, artifact := range artifacts {
		if artifact.ID == d.parentArtID {
			parentArtFound = true
			continue
		}
		if artifact.ID == d.childArt01ID {
			childArt01Found = true
			continue
		}
		if artifact.ID == d.childArt02ID {
			childArt02Found = true
			continue
		}
	}
	d.True(parentArtFound)
	d.True(childArt01Found)
	d.False(childArt02Found)

	// only query tagged artifacts
	artifacts, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
			"Tags":         "*",
		},
	})
	d.Require().Nil(err)
	parentArtFound = false
	childArt01Found = false
	childArt02Found = false
	for _, artifact := range artifacts {
		if artifact.ID == d.parentArtID {
			parentArtFound = true
			continue
		}
		if artifact.ID == d.childArt01ID {
			childArt01Found = true
			continue
		}
		if artifact.ID == d.childArt02ID {
			childArt02Found = true
			continue
		}
	}
	d.False(parentArtFound)
	d.True(childArt01Found)
	d.False(childArt02Found)

	// only query untagged artifacts
	artifacts, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
			"Tags":         "nil",
		},
	})
	d.Require().Nil(err)
	parentArtFound = false
	childArt01Found = false
	childArt02Found = false
	for _, artifact := range artifacts {
		if artifact.ID == d.parentArtID {
			parentArtFound = true
			continue
		}
		if artifact.ID == d.childArt01ID {
			childArt01Found = true
			continue
		}
		if artifact.ID == d.childArt02ID {
			childArt02Found = true
			continue
		}
	}
	d.True(parentArtFound)
	d.False(childArt01Found)
	d.False(childArt02Found)

	// query by repository ID and digest
	artifacts, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
			"Digest":       "parent_digest",
		},
	})
	d.Require().Nil(err)
	d.Require().Len(artifacts, 1)
	d.Equal(d.parentArtID, artifacts[0].ID)

	// set pagination in query
	artifacts, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]any{
			"RepositoryID": 1,
		},
		PageNumber: 1,
		PageSize:   1,
	})
	d.Require().Nil(err)
	d.Require().Len(artifacts, 1)
}

func (d *daoTestSuite) TestGet() {
	// get the non-exist artifact
	_, err := d.dao.Get(d.ctx, 10000)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))

	// get the exist artifact
	artifact, err := d.dao.Get(d.ctx, d.parentArtID)
	d.Require().Nil(err)
	d.Require().NotNil(artifact)
	d.Equal(d.parentArtID, artifact.ID)
}

func (d *daoTestSuite) TestGetByDigest() {
	// get the non-exist artifact
	_, err := d.dao.GetByDigest(d.ctx, "library/hello-world", "non_existing_digest")
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))

	// get the exist artifact
	artifact, err := d.dao.GetByDigest(d.ctx, "library/hello-world", "child_digest_02")
	d.Require().Nil(err)
	d.Require().NotNil(artifact)
	d.Equal(d.childArt02ID, artifact.ID)
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass case is covered in Setup

	// conflict
	artifact := &Artifact{
		Type:              "IMAGE",
		MediaType:         v1.MediaTypeImageConfig,
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         1,
		RepositoryID:      1,
		Digest:            "child_digest_01",
		Size:              1024,
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		ExtraAttrs:        `{"attr1":"value1"}`,
		Annotations:       `{"anno1":"value1"}`,
	}
	_, err := d.dao.Create(d.ctx, artifact)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ConflictCode))
}

func (d *daoTestSuite) TestDelete() {
	// the happy pass case is covered in TearDown

	// not exist
	err := d.dao.Delete(d.ctx, 100021)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))

	// foreign key constraint
	err = d.dao.Delete(d.ctx, d.childArt01ID)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ViolateForeignKeyConstraintCode))
}

func (d *daoTestSuite) TestUpdate() {
	// pass
	now := time.Now()
	err := d.dao.Update(d.ctx, &Artifact{
		ID:       d.parentArtID,
		PullTime: now,
	}, "PullTime")
	d.Require().Nil(err)

	artifact, err := d.dao.Get(d.ctx, d.parentArtID)
	d.Require().Nil(err)
	d.Require().NotNil(artifact)
	d.Equal(now.Unix(), artifact.PullTime.Unix())

	// not exist
	err = d.dao.Update(d.ctx, &Artifact{
		ID: 10000,
	})
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))
}

func (d *daoTestSuite) TestUpdatePullTime() {
	artifact, err := d.dao.Get(d.ctx, d.parentArtID)
	d.Require().Nil(err)
	d.Require().NotNil(artifact)
	oldPullTime := artifact.PullTime
	// case for pull_time before db pull_time.
	before := oldPullTime.AddDate(0, 0, -1)
	err = d.dao.UpdatePullTime(d.ctx, d.parentArtID, before)
	d.Require().Nil(err)

	artifact, err = d.dao.Get(d.ctx, d.parentArtID)
	d.Require().Nil(err)
	d.Require().NotNil(artifact)
	d.Equal(oldPullTime.Unix(), artifact.PullTime.Unix())

	// case for pull_time after db pull_time.
	after := oldPullTime.AddDate(0, 0, 1)
	err = d.dao.UpdatePullTime(d.ctx, d.parentArtID, after)
	d.Require().Nil(err)

	artifact, err = d.dao.Get(d.ctx, d.parentArtID)
	d.Require().Nil(err)
	d.Require().NotNil(artifact)
	d.Equal(after.Unix(), artifact.PullTime.Unix())
}

func (d *daoTestSuite) TestCreateReference() {
	// happy pass is covered in SetupTest

	// conflict
	_, err := d.dao.CreateReference(d.ctx, &ArtifactReference{
		ParentID: d.parentArtID,
		ChildID:  d.childArt01ID,
	})
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ConflictCode))

	// foreign key constraint
	_, err = d.dao.CreateReference(d.ctx, &ArtifactReference{
		ParentID: d.parentArtID,
		ChildID:  1000,
	})
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ViolateForeignKeyConstraintCode))
}

func (d *daoTestSuite) TestListReferences() {
	references, err := d.dao.ListReferences(d.ctx, &q.Query{
		Keywords: map[string]any{
			"ParentID": d.parentArtID,
			"ChildID":  d.childArt01ID,
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(1, len(references))
	d.Equal(d.reference01ID, references[0].ID)
}

func (d *daoTestSuite) TestDeleteReference() {
	// not exist
	err := d.dao.DeleteReference(d.ctx, 10000)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))
}

func (d *daoTestSuite) TestDeleteReferences() {
	// happy pass is covered in TearDownTest

	// parent artifact not exist
	err := d.dao.DeleteReferences(d.ctx, 10000)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))
}

func (d *daoTestSuite) TestListWithLatest() {
	now := time.Now()
	art := &Artifact{
		Type:              "IMAGE",
		MediaType:         v1.MediaTypeImageConfig,
		ManifestMediaType: v1.MediaTypeImageIndex,
		ProjectID:         1234,
		RepositoryID:      1234,
		RepositoryName:    "library2/hello-world1",
		Digest:            "digest",
		PushTime:          now,
		PullTime:          now,
		Annotations:       `{"anno1":"value1"}`,
	}
	id, err := d.dao.Create(d.ctx, art)
	d.Require().Nil(err)

	time.Sleep(1 * time.Second)
	now = time.Now()

	art2 := &Artifact{
		Type:              "IMAGE",
		MediaType:         v1.MediaTypeImageConfig,
		ManifestMediaType: v1.MediaTypeImageIndex,
		ProjectID:         1234,
		RepositoryID:      1235,
		RepositoryName:    "library2/hello-world2",
		Digest:            "digest",
		PushTime:          now,
		PullTime:          now,
		Annotations:       `{"anno1":"value1"}`,
	}
	id1, err := d.dao.Create(d.ctx, art2)
	d.Require().Nil(err)

	time.Sleep(1 * time.Second)
	now = time.Now()

	art3 := &Artifact{
		Type:              "IMAGE",
		MediaType:         v1.MediaTypeImageConfig,
		ManifestMediaType: v1.MediaTypeImageIndex,
		ProjectID:         1234,
		RepositoryID:      1235,
		RepositoryName:    "library2/hello-world2",
		Digest:            "digest2",
		PushTime:          now,
		PullTime:          now,
		Annotations:       `{"anno1":"value1"}`,
	}
	id2, err := d.dao.Create(d.ctx, art3)
	d.Require().Nil(err)

	latest, err := d.dao.ListWithLatest(d.ctx, &q.Query{
		Keywords: map[string]any{
			"ProjectID":  1234,
			"media_type": v1.MediaTypeImageConfig,
		},
	})

	d.Require().Nil(err)
	d.Require().Equal(2, len(latest))
	d.Equal("library2/hello-world1", latest[0].RepositoryName)

	defer d.dao.Delete(d.ctx, id)
	defer d.dao.Delete(d.ctx, id1)
	defer d.dao.Delete(d.ctx, id2)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
