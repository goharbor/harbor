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
	"github.com/stretchr/testify/suite"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/auditext/model"
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
	artifactID, err := d.dao.Create(d.ctx, &model.AuditLogExt{
		Operation:            "Create",
		ResourceType:         "user",
		Resource:             "user01",
		Username:             "admin",
		OperationDescription: "Create user",
		OperationResult:      true,
		OpTime:               time.Now().AddDate(0, 0, -8),
	})
	d.Require().Nil(err)
	d.auditID = artifactID
}

func (d *daoTestSuite) TearDownSuite() {
	ormer, err := orm.FromContext(d.ctx)
	d.Require().Nil(err)
	_, err = ormer.Raw("delete from audit_log_ext").Exec()
	d.Require().Nil(err)

}

func (d *daoTestSuite) TestList() {
	// nil query
	audits, err := d.dao.List(d.ctx, nil)
	d.Require().Nil(err)

	// query by repository ID and name
	audits, err = d.dao.List(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"Resource": "user01",
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

func (d *daoTestSuite) TestCount() {
	total, err := d.dao.Count(d.ctx, nil)
	d.Require().Nil(err)
	d.True(total > 0)
	total, err = d.dao.Count(d.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"Resource": "user01",
		},
	})
	d.Require().Nil(err)
	d.Equal(int64(1), total)
}

func (d *daoTestSuite) TestListPIDs() {
	// get the non-exist tag
	id1, err := d.dao.Create(d.ctx, &model.AuditLogExt{
		Operation:    "create",
		ResourceType: "artifact",
		Resource:     "library/hello-world",
		Username:     "admin",
		ProjectID:    11,
	})
	d.Require().Nil(err)
	id2, err := d.dao.Create(d.ctx, &model.AuditLogExt{
		Operation:    "create",
		ResourceType: "artifact",
		Resource:     "library/hello-world",
		Username:     "admin",
		ProjectID:    12,
	})
	d.Require().Nil(err)
	id3, err := d.dao.Create(d.ctx, &model.AuditLogExt{
		Operation:    "delete",
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
	audit := &model.AuditLogExt{
		Operation:       "create",
		ResourceType:    "user",
		Resource:        "user02",
		OperationResult: true,
		Username:        "admin",
	}
	_, err := d.dao.Create(d.ctx, audit)
	d.Require().Nil(err)
}

func (d *daoTestSuite) TestPurge() {
	// try to purge the audit log ext with the time range of 30 days, false
	result, err := d.dao.Purge(d.ctx, 24*30, []string{"create_user"}, true)
	d.Require().Nil(err)
	d.Require().Equal(int64(0), result)
	// try to purge the audit log ext with the time range of 7 days, true
	result1, err := d.dao.Purge(d.ctx, 24*7, []string{"create_user"}, true)
	d.Require().Nil(err)
	d.Require().Equal(int64(1), result1)

}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}

func TestPermitEventTypes(t *testing.T) {
	// test permit event types
	eventTypes := permitEventTypes([]string{"create_user", "delete_user", "delete_anything"})
	if len(eventTypes) != 2 {
		t.Errorf("permitEventTypes failed")
	}
	eventTypes2 := permitEventTypes([]string{})
	if len(eventTypes2) != 0 {
		t.Errorf("permitEventTypes failed")
	}

}
