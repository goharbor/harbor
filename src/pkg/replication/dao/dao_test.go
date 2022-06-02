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
	"github.com/goharbor/harbor/src/pkg/replication/model"
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
	registry := &model.Policy{
		Name: "test-rule",
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
			"Name": "test-rule",
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)
}

func (d *daoTestSuite) TestList() {
	// nil query
	policies, err := d.dao.List(d.ctx, nil)
	d.Require().Nil(err)
	found := false
	for _, policy := range policies {
		if policy.ID == d.id {
			found = true
			break
		}
	}
	d.True(found)

	// query by name
	policies, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"Name": "test-rule",
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(1, len(policies))
	d.Equal(d.id, policies[0].ID)
}

func (d *daoTestSuite) TestGet() {
	// get the non-exist policy
	_, err := d.dao.Get(d.ctx, 10000)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))

	// get the exist policy
	policy, err := d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Require().NotNil(policy)
	d.Equal(d.id, policy.ID)
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass case is covered in Setup

	// conflict
	policy := &model.Policy{
		Name: "test-rule",
	}
	_, err := d.dao.Create(d.ctx, policy)
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
	err := d.dao.Update(d.ctx, &model.Policy{
		ID:          d.id,
		Description: "description",
	}, "Description")
	d.Require().Nil(err)

	policy, err := d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Require().NotNil(policy)
	d.Equal("description", policy.Description)

	// not exist
	err = d.dao.Update(d.ctx, &model.Policy{
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
