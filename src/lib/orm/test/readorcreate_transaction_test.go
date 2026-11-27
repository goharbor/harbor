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
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/jackc/pgx/v5/pgconn"

	. "github.com/goharbor/harbor/src/lib/orm"
)

const postgresTransactionAbortedCode = "25P02"

// TestDuplicateKeyErrorAbortsTransactionWithoutSavepoint demonstrates that an
// unisolated duplicate-key error aborts a PostgreSQL transaction.
//
// EXPECTED: This test passes by proving the transaction is aborted (25P02).
func (suite *OrmSuite) TestDuplicateKeyErrorAbortsTransactionWithoutSavepoint() {
	ctx := NewContext(context.TODO(), orm.NewOrm())
	uniqueName := fmt.Sprintf("beego%d", time.Now().UnixNano())
	var gotError25P02 bool

	WithTransaction(func(txCtx context.Context) error {
		o, err := FromContext(txCtx)
		suite.Require().NoError(err)

		// Insert a record, then try to insert duplicate (simulates race condition)
		_, err = o.Insert(&Foo{Name: uniqueName})
		suite.Require().NoError(err)
		_, err = o.Insert(&Foo{Name: uniqueName}) // duplicate key error
		suite.Error(err, "Should get duplicate key error")

		// Try another operation - this WILL fail with 25P02
		_, err = o.Insert(&Foo{Name: fmt.Sprintf("next%d", time.Now().UnixNano())})
		if isPostgresErrorCode(err, postgresTransactionAbortedCode) {
			gotError25P02 = true
		}
		return err
	})(ctx)

	suite.True(gotError25P02, "duplicate-key error should abort the transaction (25P02) when not isolated")
}

// TestCustomReadOrCreateDoesNotAbortTransactionOnDuplicateKey demonstrates that
// the custom orm.ReadOrCreate isolates duplicate-key errors and retries the read.
//
// EXPECTED: This test PASSES by proving the transaction is NOT corrupted.
func (suite *OrmSuite) TestCustomReadOrCreateDoesNotAbortTransactionOnDuplicateKey() {
	recordName := fmt.Sprintf("custom%d", time.Now().UnixNano())
	releaseFirstTransaction := make(chan struct{})
	firstReady := make(chan error, 1)
	firstDone := make(chan error, 1)
	secondPID := make(chan backendPIDResult, 1)
	secondDone := make(chan error, 1)
	released := false
	defer func() {
		if !released {
			close(releaseFirstTransaction)
		}
	}()

	go func() {
		ctx := NewContext(context.TODO(), orm.NewOrm())
		firstDone <- WithTransaction(func(txCtx context.Context) error {
			foo := &Foo{Name: recordName}
			created, _, err := ReadOrCreate(txCtx, foo, "Name")
			if err != nil {
				firstReady <- fmt.Errorf("first ReadOrCreate: %w", err)
				return err
			}
			if !created {
				firstReady <- fmt.Errorf("first ReadOrCreate should create %q", recordName)
				return nil
			}

			firstReady <- nil
			<-releaseFirstTransaction
			return nil
		})(ctx)
	}()

	suite.Require().NoError(<-firstReady)

	go func() {
		ctx := NewContext(context.TODO(), orm.NewOrm())
		secondDone <- WithTransaction(func(txCtx context.Context) error {
			pid, err := postgresBackendPID(txCtx)
			if err != nil {
				secondPID <- backendPIDResult{err: err}
				return err
			}
			secondPID <- backendPIDResult{pid: pid}

			foo := &Foo{Name: recordName}
			created, _, err := ReadOrCreate(txCtx, foo, "Name")
			if err != nil {
				return fmt.Errorf("racing ReadOrCreate: %w", err)
			}
			if created {
				return fmt.Errorf("racing ReadOrCreate should find %q after duplicate-key retry", recordName)
			}

			foo2 := &Foo{Name: fmt.Sprintf("new%d", time.Now().UnixNano())}
			created, _, err = ReadOrCreate(txCtx, foo2, "Name")
			if err != nil {
				if isPostgresErrorCode(err, postgresTransactionAbortedCode) {
					return fmt.Errorf("transaction was aborted after duplicate-key retry: %w", err)
				}
				return fmt.Errorf("subsequent ReadOrCreate: %w", err)
			}
			if !created {
				return fmt.Errorf("subsequent ReadOrCreate should create %q", foo2.Name)
			}

			return nil
		})(ctx)
	}()

	pidResult := <-secondPID
	suite.Require().NoError(pidResult.err)
	suite.Require().Eventually(func() bool {
		waiting, err := postgresBackendWaitingOnLock(pidResult.pid)
		return err == nil && waiting
	}, 5*time.Second, 50*time.Millisecond, "second transaction should block on the first transaction's unique-key insert")

	close(releaseFirstTransaction)
	released = true
	suite.Require().NoError(<-firstDone)
	suite.NoError(<-secondDone)
}

type backendPIDResult struct {
	pid int
	err error
}

func postgresBackendPID(ctx context.Context) (int, error) {
	o, err := FromContext(ctx)
	if err != nil {
		return 0, err
	}

	var pid int
	if err := o.Raw("SELECT pg_backend_pid()").QueryRow(&pid); err != nil {
		return 0, err
	}
	return pid, nil
}

func postgresBackendWaitingOnLock(pid int) (bool, error) {
	o, err := FromContext(Context())
	if err != nil {
		return false, err
	}

	var waiting bool
	err = o.Raw(`SELECT EXISTS (
		SELECT 1
		FROM pg_stat_activity
		WHERE pid = ? AND wait_event_type = 'Lock'
	)`, pid).QueryRow(&waiting)
	return waiting, err
}

func isPostgresErrorCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}
