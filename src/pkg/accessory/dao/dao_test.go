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

	beegoorm "github.com/beego/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	errors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	artdao "github.com/goharbor/harbor/src/pkg/artifact/dao"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type daoTestSuite struct {
	htesting.Suite
	dao           DAO
	artDAO        artdao.DAO
	artifactID    int64
	subArtifactID int64
	accID         int64
	ctx           context.Context
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = New()
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.NewContext(nil, beegoorm.NewOrm())
	d.ClearTables = []string{"artifact", "artifact_accessory"}

	d.artDAO = artdao.New()
	artifactID, err := d.artDAO.Create(d.ctx, &artdao.Artifact{
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1000,
		Digest:            d.DigestString(),
	})
	d.Require().Nil(err)
	d.subArtifactID = artifactID

	d.artDAO = artdao.New()
	artifactID, err = d.artDAO.Create(d.ctx, &artdao.Artifact{
		Type:              "Signature",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1000,
		Digest:            d.DigestString(),
	})
	d.Require().Nil(err)
	d.artifactID = artifactID

	accID, err := d.dao.Create(d.ctx, &Accessory{
		ArtifactID:        d.artifactID,
		SubjectArtifactID: d.subArtifactID,
		Digest:            d.DigestString(),
		Size:              1234,
		Type:              "cosign.signature",
	})
	d.Require().Nil(err)
	d.accID = accID
}

func (d *daoTestSuite) TearDownSuite() {
	err := d.dao.Delete(d.ctx, d.accID)
	d.Require().Nil(err)
	err = d.artDAO.Delete(d.ctx, d.artifactID)
	d.Require().Nil(err)
	err = d.artDAO.Delete(d.ctx, d.subArtifactID)
	d.Require().Nil(err)
}

func (d *daoTestSuite) SetupTest() {
}

func (d *daoTestSuite) TearDownTest() {
}

func (d *daoTestSuite) TestCount() {
	// nil query
	total, err := d.dao.Count(d.ctx, nil)
	d.Require().Nil(err)
	d.True(total > 0)
	total, err = d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"SubjectArtifactID": d.subArtifactID,
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)
}

func (d *daoTestSuite) TestList() {
	// nil query
	accs, err := d.dao.List(d.ctx, nil)
	d.Require().Nil(err)
	found := false
	for _, acc := range accs {
		if acc.Type == "cosign.signature" {
			found = true
			break
		}
	}
	d.True(found)

	accs, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"SubjectArtifactID": d.subArtifactID,
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(1, len(accs))
	d.Equal(d.accID, accs[0].ID)
}

func (d *daoTestSuite) TestGet() {
	_, err := d.dao.Get(d.ctx, 10000)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))

	acc, err := d.dao.Get(d.ctx, d.accID)
	d.Require().Nil(err)
	d.Require().NotNil(acc)
	d.Equal(d.accID, acc.ID)
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass case is covered in Setup

	// conflict
	acc := &Accessory{
		ArtifactID:        d.artifactID,
		SubjectArtifactID: d.subArtifactID,
		Digest:            d.DigestString(),
		Size:              1234,
		Type:              "cosign.signature",
	}
	_, err := d.dao.Create(d.ctx, acc)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ConflictCode))

	// violating foreign key constraint: the artifact that the tag tries to attach doesn't exist
	acc = &Accessory{
		ArtifactID:        999,
		SubjectArtifactID: 998,
		Digest:            d.DigestString(),
		Size:              1234,
		Type:              "cosign.signature",
	}
	_, err = d.dao.Create(d.ctx, acc)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ViolateForeignKeyConstraintCode))
}

func (d *daoTestSuite) TestDelete() {
	// happy pass is covered in TearDown

	// not exist
	err := d.dao.Delete(d.ctx, 10000)
	d.Require().NotNil(err)
	var e *errors.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(errors.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestDeleteOfArtifact() {
	subArtID, err := d.artDAO.Create(d.ctx, &artdao.Artifact{
		Type:              "IMAGE",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1000,
		Digest:            d.DigestString(),
	})
	d.Require().Nil(err)
	defer d.artDAO.Delete(d.ctx, subArtID)

	artID1, err := d.artDAO.Create(d.ctx, &artdao.Artifact{
		Type:              "Signature",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1000,
		Digest:            d.DigestString(),
	})
	d.Require().Nil(err)
	defer d.artDAO.Delete(d.ctx, artID1)

	artID2, err := d.artDAO.Create(d.ctx, &artdao.Artifact{
		Type:              "Signature",
		MediaType:         "application/vnd.oci.image.config.v1+json",
		ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
		ProjectID:         1,
		RepositoryID:      1000,
		Digest:            d.DigestString(),
	})
	d.Require().Nil(err)
	defer d.artDAO.Delete(d.ctx, artID2)

	acc1 := &Accessory{
		ArtifactID:        artID1,
		SubjectArtifactID: subArtID,
		Digest:            d.DigestString(),
		Size:              1234,
		Type:              "cosign.signature",
	}
	_, err = d.dao.Create(d.ctx, acc1)
	d.Require().Nil(err)

	acc2 := &Accessory{
		ArtifactID:        artID2,
		SubjectArtifactID: subArtID,
		Digest:            d.DigestString(),
		Size:              1234,
		Type:              "cosign.signature",
	}
	_, err = d.dao.Create(d.ctx, acc2)
	d.Require().Nil(err)

	accs, err := d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"SubjectArtifactID": subArtID,
		},
	})
	for _, acc := range accs {
		fmt.Println(acc.ID)
	}
	d.Require().Nil(err)
	d.Require().Len(accs, 2)

	_, err = d.dao.DeleteAccessories(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"SubjectArtifactID": subArtID,
		},
	})
	d.Require().Nil(err)

	accs, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"SubjectArtifactID": subArtID,
		},
	})
	d.Require().Nil(err)
	d.Require().Len(accs, 0)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
