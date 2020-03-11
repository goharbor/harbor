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
	"strconv"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/quota"
	"github.com/goharbor/harbor/src/pkg/quota/driver"
	"github.com/goharbor/harbor/src/pkg/types"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	quotatesting "github.com/goharbor/harbor/src/testing/pkg/quota"
	drivertesting "github.com/goharbor/harbor/src/testing/pkg/quota/driver"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
}

func (suite *ControllerTestSuite) TestGetReservedResources() {
	reservedExpiration := time.Second * 3
	ctl := &controller{reservedExpiration: reservedExpiration}

	reference, referenceID := "reference", uuid.New().String()

	{
		resources, err := ctl.getReservedResources(context.TODO(), reference, referenceID)
		suite.Nil(err)
		suite.Len(resources, 0)
	}

	suite.Nil(ctl.setReservedResources(context.TODO(), reference, referenceID, types.ResourceList{types.ResourceCount: 1}))

	{
		resources, err := ctl.getReservedResources(context.TODO(), reference, referenceID)
		suite.Nil(err)
		suite.Len(resources, 1)
	}

	time.Sleep(reservedExpiration)

	{
		resources, err := ctl.getReservedResources(context.TODO(), reference, referenceID)
		suite.Nil(err)
		suite.Len(resources, 0)
	}
}

func (suite *ControllerTestSuite) TestReserveResources() {
	quotaMgr := &quotatesting.Manager{}

	hardLimits := types.ResourceList{types.ResourceCount: 1}

	mock.OnAnything(quotaMgr, "GetForUpdate").Return(&quota.Quota{Hard: hardLimits.String(), Used: types.Zero(hardLimits).String()}, nil)

	ctl := &controller{quotaMgr: quotaMgr, reservedExpiration: defaultReservedExpiration}

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	reference, referenceID := "reference", uuid.New().String()
	resources := types.ResourceList{types.ResourceCount: 1}

	suite.Nil(ctl.reserveResources(ctx, reference, referenceID, resources))

	suite.Error(ctl.reserveResources(ctx, reference, referenceID, resources))
}

func (suite *ControllerTestSuite) TestUnreserveResources() {
	quotaMgr := &quotatesting.Manager{}

	hardLimits := types.ResourceList{types.ResourceCount: 1}

	mock.OnAnything(quotaMgr, "GetForUpdate").Return(&quota.Quota{Hard: hardLimits.String(), Used: types.Zero(hardLimits).String()}, nil)

	ctl := &controller{quotaMgr: quotaMgr, reservedExpiration: defaultReservedExpiration}

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	reference, referenceID := "reference", uuid.New().String()
	resources := types.ResourceList{types.ResourceCount: 1}

	suite.Nil(ctl.reserveResources(ctx, reference, referenceID, resources))

	suite.Error(ctl.reserveResources(ctx, reference, referenceID, resources))

	suite.Nil(ctl.unreserveResources(ctx, reference, referenceID, resources))

	suite.Nil(ctl.reserveResources(ctx, reference, referenceID, resources))
}

func (suite *ControllerTestSuite) TestRequest() {
	quotaMgr := &quotatesting.Manager{}

	hardLimits := types.ResourceList{types.ResourceCount: 1}

	q := &quota.Quota{Hard: hardLimits.String(), Used: types.Zero(hardLimits).String()}
	used := types.ResourceList{types.ResourceCount: 0}

	mock.OnAnything(quotaMgr, "GetForUpdate").Return(q, nil)

	mock.OnAnything(quotaMgr, "Update").Return(nil).Run(func(mock.Arguments) {
		q.SetUsed(used)
	})

	d := &drivertesting.Driver{}

	mock.OnAnything(d, "CalculateUsage").Return(used, nil).Run(func(args mock.Arguments) {
		used[types.ResourceCount]++
	})

	driver.Register("mock", d)

	ctl := &controller{quotaMgr: quotaMgr, reservedExpiration: defaultReservedExpiration}

	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	reference, referenceID := "mock", "1"
	resources := types.ResourceList{types.ResourceCount: 1}

	{
		suite.Nil(ctl.Request(ctx, reference, referenceID, resources, func() error { return nil }))
	}

	{
		suite.Error(ctl.Request(ctx, reference, referenceID, resources, func() error { return nil }))
	}
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}

func BenchmarkGetReservedResources(b *testing.B) {
	ctl := &controller{reservedExpiration: defaultReservedExpiration}

	ctx := context.TODO()
	reference, referenceID := "reference", uuid.New().String()
	ctl.setReservedResources(ctx, reference, referenceID, types.ResourceList{types.ResourceCount: 1})

	for i := 0; i < b.N; i++ {
		ctl.getReservedResources(ctx, reference, referenceID)
	}
}

func BenchmarkSetReservedResources(b *testing.B) {
	ctl := &controller{reservedExpiration: defaultReservedExpiration}

	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		s := strconv.Itoa(i)
		ctl.setReservedResources(ctx, "reference"+s, s, types.ResourceList{types.ResourceCount: 1})
	}
}
