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

package quota

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	htesting "github.com/goharbor/harbor/src/testing"
)

type RequestDBTestSuite struct {
	htesting.Suite

	ctl       *controller
	reference string
}

func (suite *RequestDBTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.ctl = &controller{quotaMgr: quota.Mgr}
	suite.reference = "request-db"
}

func (suite *RequestDBTestSuite) create(hard, used int64) *quota.Quota {
	ctx := orm.Context()
	id, err := quota.Mgr.Create(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: hard}, types.ResourceList{types.ResourceStorage: used})
	suite.Require().NoError(err)
	q, err := quota.Mgr.Get(ctx, id)
	suite.Require().NoError(err)
	return q
}

func (suite *RequestDBTestSuite) used(q *quota.Quota) int64 {
	q, err := quota.Mgr.Get(orm.Context(), q.ID)
	suite.Require().NoError(err)
	used, err := q.GetUsed()
	suite.Require().NoError(err)
	return used[types.ResourceStorage]
}

// TestInTransactionEnforcesUnderConcurrency: with the reservation after f,
// concurrent requests in their own transactions still never exceed the limit,
// and a denied request rolls back without leaving usage behind.
func (suite *RequestDBTestSuite) TestInTransactionEnforcesUnderConcurrency() {
	q := suite.create(100, 0)

	var ok atomic.Int64
	var wg sync.WaitGroup
	for range 25 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := orm.WithTransaction(func(ctx context.Context) error {
				return suite.ctl.Request(ctx, q.Reference, q.ReferenceID, types.ResourceList{types.ResourceStorage: 10}, func() error {
					time.Sleep(20 * time.Millisecond)
					return nil
				})
			})(orm.Context())
			if err == nil {
				ok.Add(1)
			}
		}()
	}
	wg.Wait()

	suite.Equal(int64(10), ok.Load())
	suite.Equal(int64(100), suite.used(q))
}

// TestFailureLeavesUsageUnchanged: whether the failed request runs in a
// transaction that then commits or rolls back, or in none, the usage ends where
// it started and is never subtracted twice.
func (suite *RequestDBTestSuite) TestFailureLeavesUsageUnchanged() {
	failed := errors.New("upload failed")
	request := func(ctx context.Context, q *quota.Quota) error {
		err := suite.ctl.Request(ctx, q.Reference, q.ReferenceID, types.ResourceList{types.ResourceStorage: 50}, func() error { return failed })
		suite.ErrorIs(err, failed)
		return err
	}

	q := suite.create(100, 10)
	suite.Error(request(orm.Context(), q))
	suite.Equal(int64(10), suite.used(q), "no transaction")

	for _, commit := range []bool{true, false} {
		q := suite.create(100, 10)
		txErr := orm.WithTransaction(func(ctx context.Context) error {
			err := request(ctx, q)
			if commit {
				return nil
			}
			return err
		})(orm.Context())
		suite.Equal(!commit, txErr != nil)
		suite.Equal(int64(10), suite.used(q), "commit=%v", commit)
	}
}

// TestRollbackAfterRefreshKeepsUsage: a refresh between the reservation and
// the failed upload already recomputed the usage without the reservation, so
// the rollback must not subtract it again and erase other usage.
func (suite *RequestDBTestSuite) TestRollbackAfterRefreshKeepsUsage() {
	ctx := orm.Context()
	q := suite.create(1000, 50)
	failed := errors.New("upload failed")
	err := suite.ctl.Request(ctx, q.Reference, q.ReferenceID, types.ResourceList{types.ResourceStorage: 100}, func() error {
		refreshed, err := quota.Mgr.Get(ctx, q.ID)
		suite.Require().NoError(err)
		refreshed.SetUsed(types.ResourceList{types.ResourceStorage: 50})
		suite.Require().NoError(quota.Mgr.Update(ctx, refreshed))
		return failed
	})
	suite.ErrorIs(err, failed)
	suite.Equal(int64(50), suite.used(q))
}

func TestRequestDBTestSuite(t *testing.T) {
	suite.Run(t, &RequestDBTestSuite{})
}
