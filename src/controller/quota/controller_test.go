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

	// emulates the conditional UPDATE against q's hard limit
	mock.OnAnything(suite.quotaMgr, "AddUsage").Return(func(_ context.Context, _, _ string, delta types.ResourceList, enforce bool) (bool, error) {
		hard, _ := q.GetHard()
		used, _ := q.GetUsed()
		if enforce && quota.IsSafe(hard, used, types.Add(used, delta), false) != nil {
			return false, nil
		}
		return true, nil
	})
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

func (suite *ControllerTestSuite) TestRequestExceedIsDenied() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	called := false
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 101}, func() error {
		called = true
		return nil
	})
	suite.True(errors.IsErr(err, errors.DENIED), "a quota breach is reported as denied")
	suite.False(called)
}

func (suite *ControllerTestSuite) TestRequestFunctionFailedRollsBack() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	resources := types.ResourceList{types.ResourceStorage: 10}
	suite.Error(suite.ctl.Request(ctx, suite.reference, uuid.New().String(), resources, func() error { return fmt.Errorf("error") }))
	suite.quotaMgr.AssertCalled(suite.T(), "AddUsage", mock.Anything, mock.Anything, mock.Anything, resources, true)
	suite.quotaMgr.AssertCalled(suite.T(), "AddUsage", mock.Anything, mock.Anything, mock.Anything, types.ResourceList{types.ResourceStorage: -10}, false)
}

func (suite *ControllerTestSuite) TestRequestZeroDeniedWhenAlreadyOver() {
	q := &quota.Quota{
		Hard: types.ResourceList{types.ResourceStorage: 100}.String(),
		Used: types.ResourceList{types.ResourceStorage: 150}.String(),
	}
	suite.PrepareForUpdate(q, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 0}, func() error {
		return nil
	})
	suite.Error(err)
	suite.quotaMgr.AssertNotCalled(suite.T(), "AddUsage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *ControllerTestSuite) TestRequestInTransactionReservesAfterF() {
	// in a transaction the UPDATE must not hold the row lock across f
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeTxOrmer{})
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 10}, func() error {
		suite.quotaMgr.AssertNotCalled(suite.T(), "AddUsage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		return nil
	})
	suite.Nil(err)
	suite.quotaMgr.AssertNumberOfCalls(suite.T(), "AddUsage", 1)
	suite.quotaMgr.AssertNumberOfCalls(suite.T(), "GetByRef", 1)
}

func (suite *ControllerTestSuite) TestRequestInTransactionFailureReservesNothing() {
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeTxOrmer{})
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 10}, func() error {
		return fmt.Errorf("error")
	})
	suite.Error(err)
	suite.quotaMgr.AssertNotCalled(suite.T(), "AddUsage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *ControllerTestSuite) TestRequestInTransactionDeniesBeforeF() {
	// a request already over the limit is refused without running f
	suite.PrepareForUpdate(suite.quota, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeTxOrmer{})
	called := false
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 1000}, func() error {
		called = true
		return nil
	})
	suite.True(errors.IsErr(err, errors.DENIED))
	suite.False(called)
}

func (suite *ControllerTestSuite) TestRequestInTransactionDeniesWhenUpdateLoses() {
	// the read-only check passed but a concurrent request took the room
	suite.quotaMgr.On("GetByRef", mock.Anything, mock.Anything, mock.Anything).Return(suite.quota, nil)
	mock.OnAnything(suite.quotaMgr, "AddUsage").Return(false, nil)

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeTxOrmer{})
	called := false
	err := suite.ctl.Request(ctx, suite.reference, uuid.New().String(), types.ResourceList{types.ResourceStorage: 10}, func() error {
		called = true
		return nil
	})
	suite.True(called, "the handler runs before the late UPDATE")
	suite.True(errors.IsErr(err, errors.DENIED))
	var errs quota.Errors
	suite.True(errors.As(err, &errs), "a lost race still reports the overflow for the quota-exceeded event")
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
