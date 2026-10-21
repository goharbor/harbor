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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

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

func (suite *ControllerTestSuite) TestRequestUnlimitedSkipsReservation() {
	// unlimited hard limit with async refresh enabled: Request must not
	// write the quota usage at all
	restore := asyncRefreshConfigured
	asyncRefreshConfigured = true
	defer func() { asyncRefreshConfigured = restore }()

	q := &quota.Quota{
		Hard: types.ResourceList{types.ResourceStorage: types.UNLIMITED}.String(),
		Used: types.ResourceList{types.ResourceStorage: 0}.String(),
	}
	mock.OnAnything(suite.quotaMgr, "GetByRef").Return(q, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	called := false
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 100}, func() error {
		called = true
		return nil
	})
	suite.Nil(err)
	suite.True(called)
	suite.quotaMgr.AssertNotCalled(suite.T(), "Update")
}

func (suite *ControllerTestSuite) TestRequestUnlimitedReservesWhenAsyncDisabled() {
	// without QUOTA_ASYNC_REFRESH_DURATION the fast path must stay off:
	// even an unlimited quota goes through the reservation, keeping the
	// stored usage synchronously visible exactly as before
	restore := asyncRefreshConfigured
	asyncRefreshConfigured = false
	defer func() { asyncRefreshConfigured = restore }()

	q := &quota.Quota{
		Hard: types.ResourceList{types.ResourceStorage: types.UNLIMITED}.String(),
		Used: types.ResourceList{types.ResourceStorage: 0}.String(),
	}
	suite.PrepareForUpdate(q, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 100}, func() error {
		return nil
	})
	suite.Nil(err)
	suite.quotaMgr.AssertCalled(suite.T(), "Update", mock.Anything, mock.Anything)
}

func (suite *ControllerTestSuite) TestRequestLimitedStillReserves() {
	// finite hard limit: the reservation path must still run and write
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 10}, func() error {
		return nil
	})
	suite.Nil(err)
	suite.quotaMgr.AssertCalled(suite.T(), "Update", mock.Anything, mock.Anything)
}

func (suite *ControllerTestSuite) TestRequestLimitedDenies() {
	// finite hard limit exceeded: the request must be denied before f runs
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	called := false
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 1000}, func() error {
		called = true
		return nil
	})
	suite.Error(err)
	suite.False(called)
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
