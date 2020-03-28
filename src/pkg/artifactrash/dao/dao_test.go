package dao

import (
	"context"
	beegoorm "github.com/astaxie/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	errors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	artdao "github.com/goharbor/harbor/src/pkg/artifact/dao"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
	"testing"
)

type daoTestSuite struct {
	suite.Suite
	dao   DAO
	afDao artdao.DAO
	id    int64
	ctx   context.Context
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = New()
	d.afDao = artdao.New()
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.NewContext(nil, beegoorm.NewOrm())

	art1 := &artdao.Artifact{
		Type:              "image",
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         10,
		RepositoryID:      10,
		Digest:            "1234",
	}
	id, err := d.afDao.Create(d.ctx, art1)
	d.Require().Nil(err)
	err = d.afDao.Delete(d.ctx, id)
	d.Require().Nil(err)
	art2 := &artdao.Artifact{
		Type:              "image",
		ManifestMediaType: v1.MediaTypeImageManifest,
		ProjectID:         10,
		RepositoryID:      10,
		Digest:            "5678",
	}
	_, err = d.afDao.Create(d.ctx, art2)
	d.Require().Nil(err)

	aft := &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "test/hello-world",
		Digest:            "1234",
	}
	id, err = d.dao.Create(d.ctx, aft)
	d.Require().Nil(err)
	d.id = art2.ID
}

func (d *daoTestSuite) TearDownSuite() {
	d.afDao.Delete(d.ctx, d.id)
}

func (d *daoTestSuite) TestCreate() {
	// conflict
	aft := &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "test/hello-world",
		Digest:            "1234",
	}

	_, err := d.dao.Create(d.ctx, aft)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.ConflictCode))
}

func (d *daoTestSuite) TestDelete() {
	err := d.dao.Delete(d.ctx, 100021)
	d.Require().NotNil(err)
	var e *errors.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(errors.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestFilter() {
	afs, err := d.dao.Filter(d.ctx)
	d.Require().Nil(err)
	d.Require().Equal(afs[0].Digest, "1234")
}

func (d *daoTestSuite) TestFlush() {
	_, err := d.dao.Create(d.ctx, &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "hello-world",
		Digest:            "abcd",
	})
	d.Require().Nil(err)
	_, err = d.dao.Create(d.ctx, &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "hello-world2",
		Digest:            "efgh",
	})
	d.Require().Nil(err)
	_, err = d.dao.Create(d.ctx, &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "hello-world3",
		Digest:            "ijkl",
	})
	d.Require().Nil(err)

	err = d.dao.Flush(d.ctx)
	d.Require().Nil(err)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
