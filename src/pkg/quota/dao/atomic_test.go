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
	"sync"
	"sync/atomic"

	"github.com/google/uuid"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/quota/types"
)

func (suite *DaoTestSuite) TestAddUsageEnforcesUnderConcurrency() {
	ctx := suite.Context()
	reference := uuid.New().String()
	id, err := suite.dao.Create(ctx, reference, "1", types.ResourceList{types.ResourceStorage: 300}, types.ResourceList{types.ResourceStorage: 0})
	suite.Require().NoError(err)

	var applied atomic.Int64
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := suite.dao.AddUsage(orm.Context(), reference, "1", types.ResourceList{types.ResourceStorage: 10}, true)
			suite.NoError(err)
			if ok {
				applied.Add(1)
			}
		}()
	}
	wg.Wait()

	suite.Equal(int64(30), applied.Load())
	q, err := suite.dao.Get(ctx, id)
	suite.Require().NoError(err)
	used, _ := types.NewResourceList(q.Used)
	suite.Equal(int64(300), used[types.ResourceStorage])
	suite.Equal(int64(30), q.UsedVersion, "every applied statement bumps the CAS version")
}

func (suite *DaoTestSuite) TestAddUsageUnlimitedRollbackAndMissingLimit() {
	ctx := suite.Context()
	reference := uuid.New().String()
	id, err := suite.dao.Create(ctx, reference, "1", types.ResourceList{types.ResourceStorage: types.UNLIMITED}, types.ResourceList{types.ResourceStorage: 5})
	suite.Require().NoError(err)

	ok, err := suite.dao.AddUsage(ctx, reference, "1", types.ResourceList{types.ResourceStorage: 1 << 40}, true)
	suite.Require().NoError(err)
	suite.True(ok)

	ok, err = suite.dao.AddUsage(ctx, reference, "1", types.ResourceList{types.ResourceStorage: -(1 << 41)}, false)
	suite.Require().NoError(err)
	suite.False(ok, "a rollback larger than the usage is refused")

	q, err := suite.dao.Get(ctx, id)
	suite.Require().NoError(err)
	used, _ := types.NewResourceList(q.Used)
	suite.Equal(int64(5+1<<40), used[types.ResourceStorage], "a refused rollback leaves the usage alone")

	ok, err = suite.dao.AddUsage(ctx, reference, "1", types.ResourceList{types.ResourceStorage: -(1 << 40)}, false)
	suite.Require().NoError(err)
	suite.True(ok)
	q, err = suite.dao.Get(ctx, id)
	suite.Require().NoError(err)
	used, _ = types.NewResourceList(q.Used)
	suite.Equal(int64(5), used[types.ResourceStorage])

	missing := uuid.New().String()
	_, err = suite.dao.Create(ctx, missing, "1", types.ResourceList{"count": 10}, types.ResourceList{types.ResourceStorage: 0})
	suite.Require().NoError(err)
	ok, err = suite.dao.AddUsage(ctx, missing, "1", types.ResourceList{types.ResourceStorage: 1}, true)
	suite.Require().NoError(err)
	suite.False(ok, "a resource without a hard limit is denied, as IsSafe does")

	ok, err = suite.dao.AddUsage(ctx, uuid.New().String(), "1", types.ResourceList{types.ResourceStorage: 1}, true)
	suite.Require().NoError(err)
	suite.False(ok, "missing quota row")
}
