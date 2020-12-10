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

package scheduler

import (
	"context"
	"testing"

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
	d.dao = &dao{}
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.Context()
}

func (d *daoTestSuite) SetupTest() {
	schedule := &schedule{
		VendorType:        "Vendor",
		VendorID:          1,
		CRON:              "0 * * * * *",
		CallbackFuncName:  "callback_func_01",
		CallbackFuncParam: "callback_func_params",
		ExtraAttrs:        `{"key":"value"}`,
	}
	id, err := d.dao.Create(d.ctx, schedule)
	d.Require().Nil(err)
	d.id = id
}

func (d *daoTestSuite) TearDownTest() {
	d.Require().Nil(d.dao.Delete(d.ctx, d.id))
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass is covered in SetupTest

	// conflict
	schedule := &schedule{
		VendorType:        "Vendor",
		VendorID:          1,
		CRON:              "0 * * * * *",
		CallbackFuncName:  "callback_func_01",
		CallbackFuncParam: "callback_func_params",
		ExtraAttrs:        `{"key":"value"}`,
	}
	_, err := d.dao.Create(d.ctx, schedule)
	d.Require().NotNil(err)
	d.True(errors.IsConflictErr(err))
}

func (d *daoTestSuite) TestList() {
	schedules, err := d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"CallbackFuncName": "callback_func_01",
		},
	})
	d.Require().Nil(err)
	d.Require().Len(schedules, 1)
	d.Equal(d.id, schedules[0].ID)
}

func (d *daoTestSuite) TestGet() {
	// not found
	schedule, err := d.dao.Get(d.ctx, 10000)
	d.True(errors.IsNotFoundErr(err))

	// pass
	schedule, err = d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Equal(d.id, schedule.ID)
	d.Equal("{\"key\":\"value\"}", schedule.ExtraAttrs)
}

func (d *daoTestSuite) TestDelete() {
	// the happy pass is covered in TearDownTest

	// not found
	err := d.dao.Delete(d.ctx, 10000)
	d.True(errors.IsNotFoundErr(err))
}

func (d *daoTestSuite) TestUpdate() {
	// not found
	err := d.dao.Update(d.ctx, &schedule{
		ID: 10000,
	}, "CRON")
	d.True(errors.IsNotFoundErr(err))

	// pass
	err = d.dao.Update(d.ctx, &schedule{
		ID:   d.id,
		CRON: "* */2 * * * *",
	}, "CRON")
	d.Require().Nil(err)

	schedule, err := d.dao.Get(d.ctx, d.id)
	d.Require().Nil(err)
	d.Equal("* */2 * * * *", schedule.CRON)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
