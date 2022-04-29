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
	"github.com/stretchr/testify/suite"
)

type daoTestSuite struct {
	suite.Suite
	dao DAO
	ctx context.Context
	id  int64
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = NewDAO()
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.NewContext(nil, beegoorm.NewOrm())
}

func (d *daoTestSuite) SetupTest() {
	registry := &Registry{
		URL:      "http://harbor.local",
		Name:     "harbor",
		Type:     "harbor",
		Insecure: false,
		Status:   "health",
	}
	id, err := d.dao.Create(d.ctx, registry)
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
			"Name": "harbor",
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)
}

func (d *daoTestSuite) TestList() {
	// nil query
	registries, err := d.dao.List(d.ctx, nil)
	d.Require().Nil(err)
	found := false
	for _, registry := range registries {
		if registry.ID == d.id {
			found = true
			break
		}
	}
	d.True(found)

	// query by name
	registries, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"Name": "harbor",
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(1, len(registries))
	d.Equal(d.id, registries[0].ID)
}

func (d *daoTestSuite) TestGet() {
	// get the non-exist registry
	_, err := d.dao.Get(d.ctx, 10000)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))

	// get the exist registry
	registry, err := d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Require().NotNil(registry)
	d.Equal(d.id, registry.ID)
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass case is covered in Setup

	// conflict
	registry := &Registry{
		Name: "harbor",
	}
	_, err := d.dao.Create(d.ctx, registry)
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
	err := d.dao.Update(d.ctx, &Registry{
		ID:          d.id,
		Description: "description",
	}, "Description")
	d.Require().Nil(err)

	registry, err := d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Require().NotNil(registry)
	d.Equal("description", registry.Description)

	// not exist
	err = d.dao.Update(d.ctx, &Registry{
		ID: 10000,
	})
	d.Require().NotNil(err)
	var e *errors.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(errors.NotFoundCode, e.Code)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
