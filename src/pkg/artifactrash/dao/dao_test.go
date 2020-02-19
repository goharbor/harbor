package dao

import (
	"context"
	"errors"
	beegoorm "github.com/astaxie/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/internal/orm"
	artdao "github.com/goharbor/harbor/src/pkg/artifact/dao"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
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
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.NewContext(nil, beegoorm.NewOrm())
}

func (d *daoTestSuite) SetupTest() {
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
	d.id = id
}

func (d *daoTestSuite) TearDownTest() {
	err := d.dao.Delete(d.ctx, d.id)
	d.Require().Nil(err)
}

func (d *daoTestSuite) TestCreate() {
	// conflict
	aft := &model.ArtifactTrash{
		ManifestMediaType: v1.MediaTypeImageManifest,
		RepositoryName:    "test/hello-world",
	}

	_, err := d.dao.Create(d.ctx, aft)
	d.Require().NotNil(err)
	d.True(ierror.IsErr(err, ierror.ConflictCode))
}

func (d *daoTestSuite) TestDelete() {
	err := d.dao.Delete(d.ctx, 100021)
	d.Require().NotNil(err)
	var e *ierror.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(ierror.NotFoundCode, e.Code)
}

func (d *daoTestSuite) TestFilter() {
	afs, err := d.dao.Filter(d.ctx)
	d.Require().NotNil(err)
	d.Require().Equal(afs[0].Digest, "1234")
}
