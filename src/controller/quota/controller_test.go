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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	quotatesting "github.com/goharbor/harbor/src/testing/pkg/quota"
	drivertesting "github.com/goharbor/harbor/src/testing/pkg/quota/driver"
)

type ControllerTestSuite struct {
	suite.Suite

	reference string
	driver    *drivertesting.Driver
	quotaMgr  *quotatesting.Manager
	ctl       Controller

	quota *quota.Quota
}

func (suite *ControllerTestSuite) SetupTest() {
	suite.reference = "mock"

	suite.driver = &drivertesting.Driver{}
	driver.Register(suite.reference, suite.driver)

	suite.quotaMgr = &quotatesting.Manager{}
	suite.ctl = &controller{quotaMgr: suite.quotaMgr}

	hardLimits := types.ResourceList{types.ResourceStorage: 100}
	suite.quota = &quota.Quota{Hard: hardLimits.String(), Used: types.Zero(hardLimits).String()}
}

func (suite *ControllerTestSuite) PrepareForUpdate(q *quota.Quota, newUsage any) {
	mock.OnAnything(suite.quotaMgr, "GetByRef").Return(q, nil)

	mock.OnAnything(suite.driver, "CalculateUsage").Return(newUsage, nil)

	mock.OnAnything(suite.quotaMgr, "Update").Return(nil)
}

func (suite *ControllerTestSuite) TestRefresh() {
	suite.PrepareForUpdate(suite.quota, types.ResourceList{types.ResourceStorage: 0})

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Nil(suite.ctl.Refresh(ctx, suite.reference, referenceID))
}

func (suite *ControllerTestSuite) TestRefreshDriverNotFound() {
	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})

	suite.Error(suite.ctl.Refresh(ctx, uuid.New().String(), uuid.New().String()))
}

func (suite *ControllerTestSuite) TestRefershNegativeUsage() {
	suite.PrepareForUpdate(suite.quota, types.ResourceList{types.ResourceStorage: -1})

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Error(suite.ctl.Refresh(ctx, suite.reference, referenceID))
}

func (suite *ControllerTestSuite) TestRefreshUsageExceed() {
	suite.PrepareForUpdate(suite.quota, types.ResourceList{types.ResourceStorage: 101})

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Error(suite.ctl.Refresh(ctx, suite.reference, referenceID))
}

func (suite *ControllerTestSuite) TestUpdateUsageRetriesWithBackoff() {
	// regression test for the retry storm: optimistic-lock conflicts must
	// go through the retry loop with a non-zero backoff sleep, so losers
	// cannot re-CAS the contended quota_usage row in a zero-delay loop
	suite.PrepareForUpdate(suite.quota, types.ResourceList{types.ResourceStorage: 1})
	suite.quotaMgr.ExpectedCalls = nil
	mock.OnAnything(suite.quotaMgr, "GetByRef").Return(suite.quota, nil)
	mock.OnAnything(suite.quotaMgr, "Update").Return(orm.ErrOptimisticLock).Once()
	mock.OnAnything(suite.quotaMgr, "Update").Return(nil).Once()

	var sleeps []time.Duration
	opts := []retry.Option{
		retry.InitialInterval(time.Millisecond),
		retry.MaxInterval(5 * time.Millisecond),
		retry.Callback(func(_ error, sleep time.Duration) {
			sleeps = append(sleeps, sleep)
		}),
	}

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	suite.Nil(suite.ctl.Refresh(ctx, suite.reference, uuid.New().String(), IgnoreLimitation(true), WithRetryOptions(opts)))

	suite.Require().Len(sleeps, 1, "the optimistic-lock conflict must pass through the retry callback")
	suite.Greater(sleeps[0], time.Duration(0), "optimistic-lock retries must back off, not spin")
}

func (suite *ControllerTestSuite) TestRefreshIgnoreLimitation() {
	suite.PrepareForUpdate(suite.quota, types.ResourceList{types.ResourceStorage: 101})

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Nil(suite.ctl.Refresh(ctx, suite.reference, referenceID, IgnoreLimitation(true)))
}

func (suite *ControllerTestSuite) TestNoResourcesRequest() {
	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()

	suite.Nil(suite.ctl.Request(ctx, suite.reference, referenceID, nil, func() error { return nil }))
}

func (suite *ControllerTestSuite) TestRequest() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()
	resources := types.ResourceList{types.ResourceStorage: 100}

	suite.Nil(suite.ctl.Request(ctx, suite.reference, referenceID, resources, func() error { return nil }))
}

func (suite *ControllerTestSuite) TestRequestExceed() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()
	resources := types.ResourceList{types.ResourceStorage: 101}

	suite.Error(suite.ctl.Request(ctx, suite.reference, referenceID, resources, func() error { return nil }))
}

func (suite *ControllerTestSuite) TestRequestFunctionFailed() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()
	resources := types.ResourceList{types.ResourceStorage: 100}

	suite.Error(suite.ctl.Request(ctx, suite.reference, referenceID, resources, func() error { return fmt.Errorf("error") }))
}

func (suite *ControllerTestSuite) TestRequestResourceIsZero() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	referenceID := uuid.New().String()
	f := func() error {
		return nil
	}
	res := types.ResourceList{types.ResourceStorage: 0}
	err := suite.ctl.Request(ctx, suite.reference, referenceID, res, f)
	suite.Nil(err)
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}

type FlushQuotaTestSuite struct {
	suite.Suite

	reference string
	quotaMgr  *quotatesting.Manager
}

func (suite *FlushQuotaTestSuite) SetupSuite() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisURL := fmt.Sprintf("redis://%s:6379/0", redisHost)

	// the flush reads the records with cache.Default() and writes them with the raw redis client
	suite.T().Setenv("_REDIS_URL_HARBOR", redisURL)
	suite.Require().NoError(cache.Initialize(cache.Redis, redisURL))

	suite.reference = "mock"
}

func (suite *FlushQuotaTestSuite) SetupTest() {
	suite.quotaMgr = &quotatesting.Manager{}
}

func (suite *FlushQuotaTestSuite) newQuota(referenceID string, hard, used int64) *quota.Quota {
	return &quota.Quota{
		ID:          1,
		Reference:   suite.reference,
		ReferenceID: referenceID,
		Hard:        types.ResourceList{types.ResourceStorage: hard}.String(),
		Used:        types.ResourceList{types.ResourceStorage: used}.String(),
	}
}

// cacheQuota saves a quota record in redis as a request does, and returns its referenceID and key.
func (suite *FlushQuotaTestSuite) cacheQuota(ctx context.Context, hard, used int64) (string, string) {
	referenceID := uuid.New().String()
	key := fmt.Sprintf("quota:%s:%s", suite.reference, referenceID)
	suite.Require().NoError(cache.Default().Save(ctx, key, suite.newQuota(referenceID, hard, used), time.Minute))
	suite.T().Cleanup(func() { _ = cache.Default().Delete(ctx, key) })

	return referenceID, key
}

func (suite *FlushQuotaTestSuite) fetchQuota(ctx context.Context, key string) *quota.Quota {
	q := &quota.Quota{}
	suite.Require().NoError(cache.Default().Fetch(ctx, key, q))

	return q
}

func storageList(v int64) string {
	return types.ResourceList{types.ResourceStorage: v}.String()
}

func (suite *FlushQuotaTestSuite) TestFlushKeepsHardLimitsOfDB() {
	ctx := context.TODO()
	// the quota API changed the hard limit in the db to 200, and requests changed the usage in redis to 50
	referenceID, key := suite.cacheQuota(ctx, 100, 50)
	mock.OnAnything(suite.quotaMgr, "GetByRef").Return(suite.newQuota(referenceID, 200, 10), nil)

	var updated *quota.Quota
	mock.OnAnything(suite.quotaMgr, "Update").Return(nil).Run(func(args mock.Arguments) {
		updated = args.Get(1).(*quota.Quota)
	})

	suite.Require().NoError(flushQuotaRecord(ctx, suite.quotaMgr, key))

	// the db gets the usage from redis and keeps its hard limit
	suite.Require().NotNil(updated)
	suite.False(updated.HardChanged, "the flush must not write the hard limits to the db")
	suite.Equal(storageList(200), updated.Hard)
	suite.Equal(storageList(50), updated.Used)

	// redis gets the hard limit from the db and keeps its usage
	cached := suite.fetchQuota(ctx, key)
	suite.Equal(storageList(200), cached.Hard)
	suite.Equal(storageList(50), cached.Used)
}

func (suite *FlushQuotaTestSuite) TestFlushUnchangedQuota() {
	ctx := context.TODO()
	referenceID, key := suite.cacheQuota(ctx, 100, 50)
	mock.OnAnything(suite.quotaMgr, "GetByRef").Return(suite.newQuota(referenceID, 100, 50), nil)

	suite.Require().NoError(flushQuotaRecord(ctx, suite.quotaMgr, key))

	suite.quotaMgr.AssertNotCalled(suite.T(), "Update", mock.Anything, mock.Anything)
	cached := suite.fetchQuota(ctx, key)
	suite.Equal(storageList(100), cached.Hard)
	suite.Equal(storageList(50), cached.Used)
}

func (suite *FlushQuotaTestSuite) TestFlushDropsRecordOfDeletedQuota() {
	ctx := context.TODO()
	_, key := suite.cacheQuota(ctx, 100, 50)
	mock.OnAnything(suite.quotaMgr, "GetByRef").Return(nil, errors.NotFoundError(nil))

	suite.Require().NoError(flushQuotaRecord(ctx, suite.quotaMgr, key))

	suite.quotaMgr.AssertNotCalled(suite.T(), "Update", mock.Anything, mock.Anything)
	suite.False(cache.Default().Contains(ctx, key), "the record of a deleted quota must be removed from redis")
}

func (suite *FlushQuotaTestSuite) TestFlushDoesNotRetryOptimisticLock() {
	ctx := context.TODO()
	referenceID, key := suite.cacheQuota(ctx, 100, 50)
	mock.OnAnything(suite.quotaMgr, "GetByRef").Return(suite.newQuota(referenceID, 200, 10), nil)
	mock.OnAnything(suite.quotaMgr, "Update").Return(orm.ErrOptimisticLock)

	// another core instance wrote the usage, the next flush syncs the record again
	suite.Require().NoError(flushQuotaRecord(ctx, suite.quotaMgr, key))

	suite.quotaMgr.AssertNumberOfCalls(suite.T(), "Update", 1)
	suite.Equal(storageList(200), suite.fetchQuota(ctx, key).Hard)
}

func TestFlushQuotaTestSuite(t *testing.T) {
	suite.Run(t, &FlushQuotaTestSuite{})
}
