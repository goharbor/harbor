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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/dao"
	. "github.com/goharbor/harbor/src/lib/orm"
)

// ReadOrCreateTransactionSuite compares beego's native ReadOrCreate with
// the custom orm.ReadOrCreate to demonstrate why the custom version is needed.
// This is a regression test for the LDAP login bug (SQLSTATE 25P02).
type ReadOrCreateTransactionSuite struct {
	suite.Suite
}

func (suite *ReadOrCreateTransactionSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()
	RegisterModel(&Foo{})

	o, err := FromContext(Context())
	suite.Require().NoError(err)

	_, err = o.Raw(`CREATE TABLE IF NOT EXISTS foo (
		id SERIAL PRIMARY KEY NOT NULL,
		name VARCHAR(30),
		UNIQUE (name)
	)`).Exec()
	suite.Require().NoError(err)
}

func (suite *ReadOrCreateTransactionSuite) TearDownSuite() {
	o, _ := FromContext(Context())
	o.Raw(`DROP TABLE IF EXISTS foo`).Exec()
}

// TestBeegoNativeCorruptsTransaction demonstrates that beego's native method
// corrupts the transaction when any database error occurs.
// This is the ROOT CAUSE of the LDAP login bug.
//
// EXPECTED: This test PASSES by proving the transaction IS corrupted (25P02).
func (suite *ReadOrCreateTransactionSuite) TestBeegoNativeCorruptsTransaction() {
	ctx := NewContext(context.TODO(), orm.NewOrm())
	uniqueName := fmt.Sprintf("beego%d", time.Now().UnixNano())
	var gotError25P02 bool

	WithTransaction(func(txCtx context.Context) error {
		o, _ := FromContext(txCtx)

		// Insert a record, then try to insert duplicate (simulates race condition)
		o.Insert(&Foo{Name: uniqueName})
		_, err := o.Insert(&Foo{Name: uniqueName}) // duplicate key error
		suite.Error(err, "Should get duplicate key error")

		// Try another operation - this WILL fail with 25P02
		_, err = o.Insert(&Foo{Name: fmt.Sprintf("next%d", time.Now().UnixNano())})
		if err != nil && strings.Contains(err.Error(), "25P02") {
			gotError25P02 = true
		}
		return err
	})(ctx)

	suite.True(gotError25P02, "Beego's native method SHOULD corrupt the transaction (25P02)")
}

// TestCustomReadOrCreateDoesNotCorruptTransaction demonstrates that the custom
// orm.ReadOrCreate handles errors gracefully without corrupting the transaction.
// This is the FIX for the LDAP login bug.
//
// EXPECTED: This test PASSES by proving the transaction is NOT corrupted.
func (suite *ReadOrCreateTransactionSuite) TestCustomReadOrCreateDoesNotCorruptTransaction() {
	ctx := NewContext(context.TODO(), orm.NewOrm())
	recordName := fmt.Sprintf("custom%d", time.Now().UnixNano())

	// Pre-insert a record (simulates concurrent request that won the race)
	_, err := orm.NewOrm().Insert(&Foo{Name: recordName})
	suite.Require().NoError(err)

	var transactionHealthy = true
	err = WithTransaction(func(txCtx context.Context) error {
		// Custom ReadOrCreate finds existing record (no error, no corruption)
		foo := &Foo{Name: recordName}
		created, _, err := ReadOrCreate(txCtx, foo, "Name")
		if err != nil {
			return err
		}
		suite.False(created, "Should find existing record")

		// Subsequent operations MUST still work
		foo2 := &Foo{Name: fmt.Sprintf("new%d", time.Now().UnixNano())}
		created2, _, err := ReadOrCreate(txCtx, foo2, "Name")
		if err != nil {
			if strings.Contains(err.Error(), "25P02") {
				transactionHealthy = false
			}
			return err
		}
		suite.True(created2, "Should create new record")
		return nil
	})(ctx)

	suite.NoError(err)
	suite.True(transactionHealthy, "Custom ReadOrCreate should NOT corrupt the transaction")
}

func TestReadOrCreateTransactionSuite(t *testing.T) {
	suite.Run(t, new(ReadOrCreateTransactionSuite))
}
