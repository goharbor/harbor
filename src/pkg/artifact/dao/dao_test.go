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
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	typee                   = "IMAGE"
	mediaType               = "application/vnd.oci.image.config.v1+json"
	manifestMediaType       = "application/vnd.oci.image.manifest.v1+json"
	projectID         int64 = 1
	repositoryID      int64 = 1
	digest                  = "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180"
)

type daoTestSuite struct {
	suite.Suite
	dao        DAO
	artifactID int64
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = New()
	common_dao.PrepareTestForPostgresSQL()
}

func (d *daoTestSuite) SetupTest() {
	artifact := &Artifact{
		Type:              typee,
		MediaType:         mediaType,
		ManifestMediaType: manifestMediaType,
		ProjectID:         projectID,
		RepositoryID:      repositoryID,
		Digest:            digest,
		Size:              1024,
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		ExtraAttrs:        `{"attr1":"value1"}`,
		Annotations:       `{"anno1":"value1"}`,
	}
	id, err := d.dao.Create(nil, artifact)
	d.Require().Nil(err)
	d.artifactID = id
}

func (d *daoTestSuite) TearDownTest() {
	err := d.dao.Delete(nil, d.artifactID)
	d.Require().Nil(err)
}

func (d *daoTestSuite) TestCount() {
	// nil query
	total, err := d.dao.Count(nil, nil)
	d.Require().Nil(err)
	d.True(total > 0)

	// query by repository ID and digest
	total, err = d.dao.Count(nil, &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"digest":        digest,
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)

	// query by repository ID and digest
	total, err = d.dao.Count(nil, &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"digest":        digest,
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)

	// populate more data
	id, err := d.dao.Create(nil, &Artifact{
		Type:              typee,
		MediaType:         mediaType,
		ManifestMediaType: manifestMediaType,
		ProjectID:         projectID,
		RepositoryID:      repositoryID,
		Digest:            "sha256:digest",
	})
	d.Require().Nil(err)
	defer func() {
		err = d.dao.Delete(nil, id)
		d.Require().Nil(err)
	}()
	// set pagination in query
	total, err = d.dao.Count(nil, &q.Query{
		PageNumber: 1,
		PageSize:   1,
	})
	d.Require().Nil(err)
	d.True(total > 1)
}

func (d *daoTestSuite) TestList() {
	// nil query
	artifacts, err := d.dao.List(nil, nil)
	d.Require().Nil(err)
	found := false
	for _, artifact := range artifacts {
		if artifact.ID == d.artifactID {
			found = true
			break
		}
	}
	d.True(found)

	// query by repository ID and digest
	artifacts, err = d.dao.List(nil, &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"digest":        digest,
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(1, len(artifacts))
	d.Equal(d.artifactID, artifacts[0].ID)
}

func (d *daoTestSuite) TestGet() {
	// get the non-exist artifact
	_, err := d.dao.Get(nil, 10000)
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.NotFoundCode))

	// get the exist artifact
	artifact, err := d.dao.Get(nil, d.artifactID)
	d.Require().Nil(err)
	d.Require().NotNil(artifact)
	d.Equal(d.artifactID, artifact.ID)
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass case is covered in Setup

	// conflict
	artifact := &Artifact{
		Type:              typee,
		MediaType:         mediaType,
		ManifestMediaType: manifestMediaType,
		ProjectID:         projectID,
		RepositoryID:      repositoryID,
		Digest:            digest,
		Size:              1024,
		PushTime:          time.Now(),
		PullTime:          time.Now(),
		ExtraAttrs:        `{"attr1":"value1"}`,
		Annotations:       `{"anno1":"value1"}`,
	}
	_, err := d.dao.Create(nil, artifact)
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.ConflictCode))
}

func (d *daoTestSuite) TestDelete() {
	// the happy pass case is covered in TearDown

	// not exist
	err := d.dao.Delete(nil, 100021)
	d.Require().NotNil(err)
	var e *ierror.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(ierror.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestUpdate() {
	// pass
	now := time.Now()
	err := d.dao.Update(nil, &Artifact{
		ID:       d.artifactID,
		PushTime: now,
	}, "PushTime")
	d.Require().Nil(err)

	artifact, err := d.dao.Get(nil, d.artifactID)
	d.Require().Nil(err)
	d.Require().NotNil(artifact)
	d.Equal(now.Unix(), artifact.PullTime.Unix())

	// not exist
	err = d.dao.Update(nil, &Artifact{
		ID: 10000,
	})
	d.Require().NotNil(err)
	var e *ierror.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(ierror.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestReference() {
	// create reference
	id, err := d.dao.CreateReference(nil, &ArtifactReference{
		ParentID: d.artifactID,
		ChildID:  10000,
	})
	d.Require().Nil(err)

	// conflict
	_, err = d.dao.CreateReference(nil, &ArtifactReference{
		ParentID: d.artifactID,
		ChildID:  10000,
	})
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.ConflictCode))

	// list reference
	references, err := d.dao.ListReferences(nil, &q.Query{
		Keywords: map[string]interface{}{
			"parent_id": d.artifactID,
		},
	})
	d.Require().Equal(1, len(references))
	d.Equal(id, references[0].ID)

	// delete reference
	err = d.dao.DeleteReferences(nil, d.artifactID)
	d.Require().Nil(err)

	// parent artifact not exist
	err = d.dao.DeleteReferences(nil, 10000)
	d.Require().NotNil(err)
	var e *ierror.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(ierror.NotFoundCode, e.Code)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
