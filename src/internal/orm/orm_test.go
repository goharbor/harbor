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

package orm

import (
	"context"
	"errors"
	"testing"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/suite"
)

func addProject(ctx context.Context, project models.Project) (int64, error) {
	o, ok := FromContext(ctx)
	if !ok {
		return 0, errors.New("orm not found in context")
	}

	return o.Insert(&project)
}

func readProject(ctx context.Context, id int64) (*models.Project, error) {
	o, ok := FromContext(ctx)
	if !ok {
		return nil, errors.New("orm not found in context")
	}

	project := &models.Project{
		ProjectID: id,
	}

	if err := o.Read(project, "project_id"); err != nil {
		return nil, err
	}

	return project, nil
}

func deleteProject(ctx context.Context, id int64) error {
	o, ok := FromContext(ctx)
	if !ok {
		return errors.New("orm not found in context")
	}

	project := &models.Project{
		ProjectID: id,
	}

	_, err := o.Delete(project, "project_id")
	return err
}

func existProject(ctx context.Context, id int64) bool {
	o, ok := FromContext(ctx)
	if !ok {
		return false
	}

	project := &models.Project{
		ProjectID: id,
	}

	if err := o.Read(project, "project_id"); err != nil {
		return false
	}

	return true
}

// Suite ...
type OrmSuite struct {
	suite.Suite
}

// SetupSuite ...
func (suite *OrmSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()
}

func (suite *OrmSuite) TestContext() {
	ctx := context.TODO()

	o, ok := FromContext(ctx)
	suite.False(ok)
	suite.Nil(o)

	o, ok = FromContext(NewContext(ctx, orm.NewOrm()))
	suite.True(ok)
	suite.NotNil(o)
}

func (suite *OrmSuite) TestWithTransaction() {
	ctx := NewContext(context.TODO(), orm.NewOrm())

	var id int64
	t1 := WithTransaction(func(ctx context.Context) (err error) {
		id, err = addProject(ctx, models.Project{Name: "t1", OwnerID: 1})
		return err
	})

	suite.Nil(t1(ctx))
	suite.True(existProject(ctx, id))
	suite.Nil(deleteProject(ctx, id))
}

func (suite *OrmSuite) TestSequentialTransactions() {
	ctx := NewContext(context.TODO(), orm.NewOrm())

	var id1, id2 int64
	t1 := func(ctx context.Context, retErr error) error {
		return WithTransaction(func(ctx context.Context) (err error) {
			id1, err = addProject(ctx, models.Project{Name: "t1", OwnerID: 1})
			if err != nil {
				return err
			}

			// Ensure t1 created success
			suite.True(existProject(ctx, id1))

			return retErr
		})(ctx)
	}
	t2 := func(ctx context.Context, retErr error) error {
		return WithTransaction(func(ctx context.Context) (err error) {
			id2, _ = addProject(ctx, models.Project{Name: "t2", OwnerID: 1})
			if err != nil {
				return err
			}

			// Ensure t2 created success
			suite.True(existProject(ctx, id2))

			return retErr
		})(ctx)
	}

	if suite.Nil(t1(ctx, nil)) {
		suite.True(existProject(ctx, id1))
	}

	if suite.Nil(t2(ctx, nil)) {
		suite.True(existProject(ctx, id2))
	}

	// delete project t1 and t2 in db
	suite.Nil(deleteProject(ctx, id1))
	suite.Nil(deleteProject(ctx, id2))

	if suite.Error(t1(ctx, errors.New("oops"))) {
		suite.False(existProject(ctx, id1))
	}

	if suite.Nil(t2(ctx, nil)) {
		suite.True(existProject(ctx, id2))
		suite.Nil(deleteProject(ctx, id2))
	}
}

func (suite *OrmSuite) TestNestedTransaction() {
	ctx := NewContext(context.TODO(), orm.NewOrm())

	var id1, id2 int64
	nt1 := WithTransaction(func(ctx context.Context) (err error) {
		id1, err = addProject(ctx, models.Project{Name: "nt1", OwnerID: 1})
		return err
	})
	nt2 := WithTransaction(func(ctx context.Context) (err error) {
		id2, err = addProject(ctx, models.Project{Name: "nt2", OwnerID: 1})
		return err
	})

	nt := func(ctx context.Context, retErr error) error {
		return WithTransaction(func(ctx context.Context) error {
			if err := nt1(ctx); err != nil {
				return err
			}

			if err := nt2(ctx); err != nil {
				return err
			}

			// Ensure nt1 and nt2 created success
			suite.True(existProject(ctx, id1))
			suite.True(existProject(ctx, id2))

			return retErr
		})(ctx)
	}

	if suite.Nil(nt(ctx, nil)) {
		suite.True(existProject(ctx, id1))
		suite.True(existProject(ctx, id2))

		// delete project nt1 and nt2 in db
		suite.Nil(deleteProject(ctx, id1))
		suite.Nil(deleteProject(ctx, id2))
		suite.False(existProject(ctx, id1))
		suite.False(existProject(ctx, id2))
	}

	if suite.Error(nt(ctx, errors.New("oops"))) {
		suite.False(existProject(ctx, id1))
		suite.False(existProject(ctx, id2))
	}

	// test nt1 failed but we skip it and nt2 success
	suite.Nil(nt1(ctx))
	suite.True(existProject(ctx, id1))

	// delete nt1 here because id1 will overwrite in the following transaction
	defer func(id int64) {
		suite.Nil(deleteProject(ctx, id))
	}(id1)

	t := WithTransaction(func(ctx context.Context) error {
		suite.Error(nt1(ctx))

		if err := nt2(ctx); err != nil {
			return err
		}

		// Ensure t2 created success
		suite.True(existProject(ctx, id2))

		return nil
	})

	if suite.Nil(t(ctx)) {
		suite.True(existProject(ctx, id2))

		// delete project t2 in db
		suite.Nil(deleteProject(ctx, id2))
	}
}

func (suite *OrmSuite) TestNestedSavepoint() {
	ctx := NewContext(context.TODO(), orm.NewOrm())

	var id1, id2 int64
	ns1 := WithTransaction(func(ctx context.Context) (err error) {
		id1, err = addProject(ctx, models.Project{Name: "ns1", OwnerID: 1})
		return err
	})
	ns2 := WithTransaction(func(ctx context.Context) (err error) {
		id2, err = addProject(ctx, models.Project{Name: "ns2", OwnerID: 1})
		return err
	})

	ns := func(ctx context.Context, retErr error) error {
		return WithTransaction(func(ctx context.Context) error {
			if err := ns1(ctx); err != nil {
				return err
			}

			if err := ns2(ctx); err != nil {
				return err
			}

			// Ensure nt1 and nt2 created success
			suite.True(existProject(ctx, id1))
			suite.True(existProject(ctx, id2))

			return retErr
		})(ctx)
	}

	t := func(ctx context.Context, tErr, pErr error) error {
		return WithTransaction(func(c context.Context) error {
			ns(c, pErr)
			return tErr
		})(ctx)
	}

	// transaction commit and s1s2 commit
	suite.Nil(t(ctx, nil, nil))
	// Ensure nt1 and nt2 created success
	suite.True(existProject(ctx, id1))
	suite.True(existProject(ctx, id2))
	// delete project nt1 and nt2 in db
	suite.Nil(deleteProject(ctx, id1))
	suite.Nil(deleteProject(ctx, id2))
	suite.False(existProject(ctx, id1))
	suite.False(existProject(ctx, id2))

	// transaction commit and s1s2 rollback
	suite.Nil(t(ctx, nil, errors.New("oops")))
	// Ensure nt1 and nt2 created failed
	suite.False(existProject(ctx, id1))
	suite.False(existProject(ctx, id2))

	// transaction rollback and s1s2 commit
	suite.Error(t(ctx, errors.New("oops"), nil))
	// Ensure nt1 and nt2 created failed
	suite.False(existProject(ctx, id1))
	suite.False(existProject(ctx, id2))
}

func TestRunOrmSuite(t *testing.T) {
	suite.Run(t, new(OrmSuite))
}
