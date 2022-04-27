//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package test

import (
	"context"
	"errors"
	"sync"
	"testing"

	. "github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"

	"github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/stretchr/testify/suite"
)

type Foo struct {
	ID   int64  `orm:"pk;auto;column(id)"`
	Name string `orm:"column(name)"`
}

func (foo *Foo) TableName() string {
	return "foo"
}

func (foo *Foo) GetID() int64 {
	return foo.ID
}

func addFoo(ctx context.Context, foo Foo) (int64, error) {
	o, err := FromContext(ctx)
	if err != nil {
		return 0, err
	}

	return o.Insert(&foo)
}

func readFoo(ctx context.Context, id int64) (*Foo, error) {
	o, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}

	foo := &Foo{
		ID: id,
	}

	if err := o.Read(foo, "id"); err != nil {
		return nil, err
	}

	return foo, nil
}

func deleteFoo(ctx context.Context, id int64) error {
	o, err := FromContext(ctx)
	if err != nil {
		return err
	}
	foo := &Foo{
		ID: id,
	}

	_, err = o.Delete(foo, "id")
	return err
}

func existFoo(ctx context.Context, id int64) bool {
	o, err := FromContext(ctx)
	if err != nil {
		return false
	}

	foo := &Foo{
		ID: id,
	}

	if err := o.Read(foo, "id"); err != nil {
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
	RegisterModel(&Foo{})
	dao.PrepareTestForPostgresSQL()

	o, err := FromContext(Context())
	if err != nil {
		suite.Fail("got error %v", err)
	}

	sql := `
	CREATE TABLE IF NOT EXISTS foo (
		id SERIAL PRIMARY KEY NOT NULL,
		name VARCHAR (30),
		UNIQUE (name)
	)
	`

	_, err = o.Raw(sql).Exec()
	if err != nil {
		suite.Fail("got error %v", err)
	}
}

func (suite *OrmSuite) TearDownSuite() {
	o, err := FromContext(Context())
	if err != nil {
		suite.Fail("got error %v", err)
	}

	sql := `DROP TABLE foo`

	_, err = o.Raw(sql).Exec()
	if err != nil {
		suite.Fail("got error %v", err)
	}
}

func (suite *OrmSuite) TestContext() {
	ctx := context.TODO()

	o, err := FromContext(ctx)
	suite.NotNil(err)

	o, err = FromContext(NewContext(ctx, orm.NewOrm()))
	suite.Nil(err)
	suite.NotNil(o)
}

func (suite *OrmSuite) TestWithTransaction() {
	ctx := NewContext(context.TODO(), orm.NewOrm())

	var id int64
	t1 := WithTransaction(func(ctx context.Context) (err error) {
		id, err = addFoo(ctx, Foo{Name: "t1"})
		return err
	})

	suite.Nil(t1(ctx))
	suite.True(existFoo(ctx, id))
	suite.Nil(deleteFoo(ctx, id))
}

func (suite *OrmSuite) TestSequentialTransactions() {
	ctx := NewContext(context.TODO(), orm.NewOrm())

	var id1, id2 int64
	t1 := func(ctx context.Context, retErr error) error {
		return WithTransaction(func(ctx context.Context) (err error) {
			id1, err = addFoo(ctx, Foo{Name: "t1"})
			if err != nil {
				return err
			}

			// Ensure t1 created success
			suite.True(existFoo(ctx, id1))

			return retErr
		})(ctx)
	}
	t2 := func(ctx context.Context, retErr error) error {
		return WithTransaction(func(ctx context.Context) (err error) {
			id2, _ = addFoo(ctx, Foo{Name: "t2"})
			if err != nil {
				return err
			}

			// Ensure t2 created success
			suite.True(existFoo(ctx, id2))

			return retErr
		})(ctx)
	}

	if suite.Nil(t1(ctx, nil)) {
		suite.True(existFoo(ctx, id1))
	}

	if suite.Nil(t2(ctx, nil)) {
		suite.True(existFoo(ctx, id2))
	}

	// delete foo t1 and t2 in db
	suite.Nil(deleteFoo(ctx, id1))
	suite.Nil(deleteFoo(ctx, id2))

	if suite.Error(t1(ctx, errors.New("oops"))) {
		suite.False(existFoo(ctx, id1))
	}

	if suite.Nil(t2(ctx, nil)) {
		suite.True(existFoo(ctx, id2))
		suite.Nil(deleteFoo(ctx, id2))
	}
}

func (suite *OrmSuite) TestNestedTransaction() {
	ctx := NewContext(context.TODO(), orm.NewOrm())

	var id1, id2 int64
	nt1 := WithTransaction(func(ctx context.Context) (err error) {
		id1, err = addFoo(ctx, Foo{Name: "nt1"})
		return err
	})
	nt2 := WithTransaction(func(ctx context.Context) (err error) {
		id2, err = addFoo(ctx, Foo{Name: "nt2"})
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
			suite.True(existFoo(ctx, id1))
			suite.True(existFoo(ctx, id2))

			return retErr
		})(ctx)
	}

	if suite.Nil(nt(ctx, nil)) {
		suite.True(existFoo(ctx, id1))
		suite.True(existFoo(ctx, id2))

		// delete foo nt1 and nt2 in db
		suite.Nil(deleteFoo(ctx, id1))
		suite.Nil(deleteFoo(ctx, id2))
		suite.False(existFoo(ctx, id1))
		suite.False(existFoo(ctx, id2))
	}

	if suite.Error(nt(ctx, errors.New("oops"))) {
		suite.False(existFoo(ctx, id1))
		suite.False(existFoo(ctx, id2))
	}

	// test nt1 failed but we skip it and nt2 success
	suite.Nil(nt1(ctx))
	suite.True(existFoo(ctx, id1))

	// delete nt1 here because id1 will overwrite in the following transaction
	defer func(id int64) {
		suite.Nil(deleteFoo(ctx, id))
	}(id1)

	t := WithTransaction(func(ctx context.Context) error {
		suite.Error(nt1(ctx))

		if err := nt2(ctx); err != nil {
			return err
		}

		// Ensure t2 created success
		suite.True(existFoo(ctx, id2))

		return nil
	})

	if suite.Nil(t(ctx)) {
		suite.True(existFoo(ctx, id2))

		// delete foo t2 in db
		suite.Nil(deleteFoo(ctx, id2))
	}
}

func (suite *OrmSuite) TestNestedSavepoint() {
	ctx := NewContext(context.TODO(), orm.NewOrm())

	var id1, id2 int64
	ns1 := WithTransaction(func(ctx context.Context) (err error) {
		id1, err = addFoo(ctx, Foo{Name: "ns1"})
		return err
	})
	ns2 := WithTransaction(func(ctx context.Context) (err error) {
		id2, err = addFoo(ctx, Foo{Name: "ns2"})
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
			suite.True(existFoo(ctx, id1))
			suite.True(existFoo(ctx, id2))

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
	suite.True(existFoo(ctx, id1))
	suite.True(existFoo(ctx, id2))
	// delete foo nt1 and nt2 in db
	suite.Nil(deleteFoo(ctx, id1))
	suite.Nil(deleteFoo(ctx, id2))
	suite.False(existFoo(ctx, id1))
	suite.False(existFoo(ctx, id2))

	// transaction commit and s1s2 rollback
	suite.Nil(t(ctx, nil, errors.New("oops")))
	// Ensure nt1 and nt2 created failed
	suite.False(existFoo(ctx, id1))
	suite.False(existFoo(ctx, id2))

	// transaction rollback and s1s2 commit
	suite.Error(t(ctx, errors.New("oops"), nil))
	// Ensure nt1 and nt2 created failed
	suite.False(existFoo(ctx, id1))
	suite.False(existFoo(ctx, id2))
}

func (suite *OrmSuite) TestReadOrCreate() {
	ctx := NewContext(context.TODO(), orm.NewOrm())

	var id int64
	f1 := func(ctx context.Context) (err error) {
		created1, id1, err := ReadOrCreate(ctx, &Foo{Name: "n1"}, "name")
		suite.NoError(err)
		suite.True(created1)

		created2, id2, err := ReadOrCreate(ctx, &Foo{Name: "n1"}, "name")
		suite.NoError(err)
		suite.False(created2)

		suite.Equal(id2, id1)

		id = id1

		return nil
	}

	suite.NoError(WithTransaction(f1)(ctx))
	suite.True(existFoo(ctx, id))
}

func (suite *OrmSuite) TestReadOrCreateParallel() {
	count := 500

	arr := make([]int, count)

	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx := NewContext(context.TODO(), orm.NewOrm())
			created, _, err := ReadOrCreate(ctx, &Foo{Name: "n2"}, "name")
			suite.NoError(err)

			if created {
				arr[i] = 1
			}
		}(i)
	}

	wg.Wait()

	sum := 0
	for _, v := range arr {
		sum += v
	}

	suite.Equal(1, sum)
}

func (suite *OrmSuite) TestPaginationOnRawSQL() {
	query := &q.Query{
		PageNumber: 1,
		PageSize:   10,
	}
	sql := "select * from harbor_user where user_id > ? order by user_name "
	params := []interface{}{2}
	sql, params = PaginationOnRawSQL(query, sql, params)
	suite.Equal("select * from harbor_user where user_id > ? order by user_name  limit ? offset ?", sql)
	suite.Equal(int64(10), params[1])
	suite.Equal(int64(0), params[2])
}

func TestRunOrmSuite(t *testing.T) {
	suite.Run(t, new(OrmSuite))
}
