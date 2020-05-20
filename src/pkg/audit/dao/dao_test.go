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
	beegoorm "github.com/astaxie/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/audit/model"
	"github.com/stretchr/testify/suite"
	"testing"
)

type daoTestSuite struct {
	suite.Suite
	dao     DAO
	auditID int64
	ctx     context.Context
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = New()
	common_dao.PrepareTestForPostgresSQL()
	d.ctx = orm.NewContext(nil, beegoorm.NewOrm())
	artifactID, err := d.dao.Create(d.ctx, &model.AuditLog{
		Operation:    "Create",
		ResourceType: "artifact",
		Resource:     "library/test-audit",
		Username:     "admin",
	})
	d.Require().Nil(err)
	d.auditID = artifactID
}

func (d *daoTestSuite) TearDownSuite() {
	err := d.dao.Delete(d.ctx, d.auditID)
	d.Require().Nil(err)
}

func (d *daoTestSuite) TestCount() {
	total, err := d.dao.Count(d.ctx, nil)
	d.Require().Nil(err)
	d.True(total > 0)
	total, err = d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"Resource": "library/test-audit",
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)
}

func (d *daoTestSuite) TestList() {
	// nil query
	audits, err := d.dao.List(d.ctx, nil)
	d.Require().Nil(err)

	// query by repository ID and name
	audits, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"Resource": "library/test-audit",
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(1, len(audits))
	d.Equal("admin", audits[0].Username)
}

func (d *daoTestSuite) TestGet() {
	// get the non-exist tag
	_, err := d.dao.Get(d.ctx, 10000)
	d.Require().NotNil(err)
	d.True(errors.IsErr(err, errors.NotFoundCode))

	audit, err := d.dao.Get(d.ctx, d.auditID)
	d.Require().Nil(err)
	d.Require().NotNil(audit)
	d.Equal(d.auditID, audit.ID)
}

func (d *daoTestSuite) TestListPIDs() {
	// get the non-exist tag
	id1, err := d.dao.Create(d.ctx, &model.AuditLog{
		Operation:    "Create",
		ResourceType: "artifact",
		Resource:     "library/hello-world",
		Username:     "admin",
		ProjectID:    11,
	})
	d.Require().Nil(err)
	id2, err := d.dao.Create(d.ctx, &model.AuditLog{
		Operation:    "Create",
		ResourceType: "artifact",
		Resource:     "library/hello-world",
		Username:     "admin",
		ProjectID:    12,
	})
	d.Require().Nil(err)
	id3, err := d.dao.Create(d.ctx, &model.AuditLog{
		Operation:    "Delete",
		ResourceType: "artifact",
		Resource:     "library/hello-world",
		Username:     "admin",
		ProjectID:    13,
	})
	d.Require().Nil(err)

	// query by repository ID and name
	ol := &q.OrList{}
	for _, item := range []int64{11, 12, 13} {
		ol.Values = append(ol.Values, item)
	}
	audits, err := d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ProjectID": ol,
		},
	})
	d.Require().Nil(err)
	d.Require().Equal(3, len(audits))
	d.dao.Delete(d.ctx, id1)
	d.dao.Delete(d.ctx, id2)
	d.dao.Delete(d.ctx, id3)
}

func (d *daoTestSuite) TestCreate() {
	// conflict
	audit := &model.AuditLog{
		Operation:    "Create",
		ResourceType: "tag",
		Resource:     "library/hello-world",
		Username:     "admin",
	}
	_, err := d.dao.Create(d.ctx, audit)
	d.Require().Nil(err)
}

func (d *daoTestSuite) TestDelete() {
	err := d.dao.Delete(d.ctx, 10000)
	d.Require().NotNil(err)
	var e *errors.Error
	d.Require().True(errors.As(err, &e))
	d.Equal(errors.NotFoundCode, e.Code)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
