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
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota/driver"
	"github.com/goharbor/harbor/src/common/quota/driver/mocks"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	hardLimits = types.ResourceList{types.ResourceStorage: 1000}
	reference  = "mock"
)

func init() {
	mockDriver := &mocks.Driver{}

	mockHardLimitsFn := func() types.ResourceList {
		return types.ResourceList{
			types.ResourceStorage: -1,
		}
	}

	mockLoadFn := func(key string) driver.RefObject {
		return driver.RefObject{"id": key}
	}

	mockDriver.On("HardLimits").Return(mockHardLimitsFn)
	mockDriver.On("Load", mock.AnythingOfType("string")).Return(mockLoadFn, nil)
	mockDriver.On("Validate", mock.AnythingOfType("types.ResourceList")).Return(nil)

	driver.Register(reference, mockDriver)
}

func mustResourceList(s string) types.ResourceList {
	resources, _ := types.NewResourceList(s)
	return resources
}

type ManagerSuite struct {
	suite.Suite
}

func (suite *ManagerSuite) SetupTest() {
	_, ok := driver.Get(reference)
	if !ok {
		suite.Fail("driver not found for %s", reference)
	}
}

func (suite *ManagerSuite) quotaManager(referenceIDs ...string) *Manager {
	referenceID := "1"
	if len(referenceIDs) > 0 {
		referenceID = referenceIDs[0]
	}

	mgr, _ := NewManager(reference, referenceID)
	return mgr
}

func (suite *ManagerSuite) TearDownTest() {
	dao.ClearTable("quota")
	dao.ClearTable("quota_usage")
}

func (suite *ManagerSuite) TestNewQuota() {
	mgr := suite.quotaManager()

	if id, err := mgr.NewQuota(hardLimits); suite.Nil(err) {
		quota, _ := dao.GetQuota(id)
		suite.Equal(hardLimits, mustResourceList(quota.Hard))
	}

	mgr = suite.quotaManager("2")
	used := types.ResourceList{types.ResourceStorage: 100}
	if id, err := mgr.NewQuota(hardLimits, used); suite.Nil(err) {
		quota, _ := dao.GetQuota(id)
		suite.Equal(hardLimits, mustResourceList(quota.Hard))

		usage, _ := dao.GetQuotaUsage(id)
		suite.Equal(used, mustResourceList(usage.Used))
	}
}

func (suite *ManagerSuite) TestDeleteQuota() {
	mgr := suite.quotaManager()

	id, err := mgr.NewQuota(hardLimits)
	if suite.Nil(err) {
		quota, _ := dao.GetQuota(id)
		suite.Equal(hardLimits, mustResourceList(quota.Hard))
	}

	if err := mgr.DeleteQuota(); suite.Nil(err) {
		quota, _ := dao.GetQuota(id)
		suite.Nil(quota)
	}
}

func (suite *ManagerSuite) TestUpdateQuota() {
	mgr := suite.quotaManager()

	id, _ := mgr.NewQuota(hardLimits)
	largeHardLimits := types.ResourceList{types.ResourceStorage: 1000000}

	if err := mgr.UpdateQuota(largeHardLimits); suite.Nil(err) {
		quota, _ := dao.GetQuota(id)
		suite.Equal(largeHardLimits, mustResourceList(quota.Hard))
	}
}

func (suite *ManagerSuite) TestSetResourceUsage() {
	mgr := suite.quotaManager()
	id, _ := mgr.NewQuota(hardLimits)

	if err := mgr.SetResourceUsage(types.ResourceStorage, 999999999999999999); suite.Nil(err) {
		quota, _ := dao.GetQuota(id)
		suite.Equal(hardLimits, mustResourceList(quota.Hard))

		usage, _ := dao.GetQuotaUsage(id)
		suite.Equal(types.ResourceList{types.ResourceStorage: 999999999999999999}, mustResourceList(usage.Used))
	}

	if err := mgr.SetResourceUsage(types.ResourceStorage, 234); suite.Nil(err) {
		usage, _ := dao.GetQuotaUsage(id)
		suite.Equal(types.ResourceList{types.ResourceStorage: 234}, mustResourceList(usage.Used))
	}
}

func (suite *ManagerSuite) TestEnsureQuota() {
	// non-existent
	nonExistRefID := "3"
	mgr := suite.quotaManager(nonExistRefID)
	infinite := types.ResourceList{types.ResourceStorage: -1}
	usage := types.ResourceList{types.ResourceStorage: 10}
	err := mgr.EnsureQuota(usage)
	suite.Nil(err)
	query := &models.QuotaQuery{
		Reference:   reference,
		ReferenceID: nonExistRefID,
	}
	quotas, err := dao.ListQuotas(query)
	suite.Nil(err)
	suite.Equal(usage, mustResourceList(quotas[0].Used))
	suite.Equal(infinite, mustResourceList(quotas[0].Hard))

	// existent
	existRefID := "4"
	mgr = suite.quotaManager(existRefID)
	used := types.ResourceList{types.ResourceStorage: 11}
	if id, err := mgr.NewQuota(hardLimits, used); suite.Nil(err) {
		quota, _ := dao.GetQuota(id)
		suite.Equal(hardLimits, mustResourceList(quota.Hard))

		usage, _ := dao.GetQuotaUsage(id)
		suite.Equal(used, mustResourceList(usage.Used))
	}

	usage2 := types.ResourceList{types.ResourceStorage: 12}
	err = mgr.EnsureQuota(usage2)
	suite.Nil(err)
	query2 := &models.QuotaQuery{
		Reference:   reference,
		ReferenceID: existRefID,
	}
	quotas2, err := dao.ListQuotas(query2)
	suite.Equal(usage2, mustResourceList(quotas2[0].Used))
	suite.Equal(hardLimits, mustResourceList(quotas2[0].Hard))

}

func (suite *ManagerSuite) TestQuotaAutoCreation() {
	for i := 0; i < 10; i++ {
		mgr := suite.quotaManager(fmt.Sprintf("%d", i))
		resource := types.ResourceList{types.ResourceStorage: 100}

		suite.Nil(mgr.AddResources(resource))
	}
}

func (suite *ManagerSuite) TestAddResources() {
	mgr := suite.quotaManager()
	id, _ := mgr.NewQuota(hardLimits)

	resource := types.ResourceList{types.ResourceStorage: 100}

	if suite.Nil(mgr.AddResources(resource)) {
		usage, _ := dao.GetQuotaUsage(id)
		suite.Equal(resource, mustResourceList(usage.Used))
	}

	if suite.Nil(mgr.AddResources(resource)) {
		usage, _ := dao.GetQuotaUsage(id)
		suite.Equal(types.ResourceList{types.ResourceStorage: 200}, mustResourceList(usage.Used))
	}

	if err := mgr.AddResources(types.ResourceList{types.ResourceStorage: 10000}); suite.Error(err) {
		if errs, ok := err.(Errors); suite.True(ok) {
			for _, err := range errs {
				suite.IsType(&ResourceOverflow{}, err)
			}
		}
	}
}

func (suite *ManagerSuite) TestSubtractResources() {
	mgr := suite.quotaManager()
	id, _ := mgr.NewQuota(hardLimits)

	resource := types.ResourceList{types.ResourceStorage: 100}

	if suite.Nil(mgr.AddResources(resource)) {
		usage, _ := dao.GetQuotaUsage(id)
		suite.Equal(resource, mustResourceList(usage.Used))
	}

	if suite.Nil(mgr.SubtractResources(resource)) {
		usage, _ := dao.GetQuotaUsage(id)
		suite.Equal(types.ResourceList{types.ResourceStorage: 0}, mustResourceList(usage.Used))
	}
}

func (suite *ManagerSuite) TestRaceAddResources() {
	mgr := suite.quotaManager()
	mgr.NewQuota(hardLimits)

	resources := types.ResourceList{
		types.ResourceStorage: 100,
	}

	var wg sync.WaitGroup

	results := make([]bool, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			results[i] = mgr.AddResources(resources) == nil
		}(i)
	}
	wg.Wait()

	var success int
	for _, result := range results {
		if result {
			success++
		}
	}

	suite.Equal(10, success)
}

func (suite *ManagerSuite) TestRaceSubtractResources() {
	mgr := suite.quotaManager()
	mgr.NewQuota(hardLimits, types.ResourceList{types.ResourceStorage: 1000})

	resources := types.ResourceList{
		types.ResourceStorage: 100,
	}

	var wg sync.WaitGroup

	results := make([]bool, 100)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			results[i] = mgr.SubtractResources(resources) == nil
		}(i)
	}
	wg.Wait()

	var success int
	for _, result := range results {
		if result {
			success++
		}
	}

	suite.Equal(10, success)
}

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()

	if result := m.Run(); result != 0 {
		os.Exit(result)
	}
}

func TestRunManagerSuite(t *testing.T) {
	suite.Run(t, new(ManagerSuite))
}

func BenchmarkAddResources(b *testing.B) {
	defer func() {
		dao.ClearTable("quota")
		dao.ClearTable("quota_usage")
	}()

	mgr, _ := NewManager(reference, "1")
	mgr.NewQuota(types.ResourceList{types.ResourceStorage: int64(b.N)})

	resource := types.ResourceList{
		types.ResourceStorage: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.AddResources(resource)
	}
	b.StopTimer()
}

func BenchmarkAddResourcesParallel(b *testing.B) {
	defer func() {
		dao.ClearTable("quota")
		dao.ClearTable("quota_usage")
	}()

	mgr, _ := NewManager(reference, "1")
	mgr.NewQuota(types.ResourceList{})

	resource := types.ResourceList{
		types.ResourceStorage: 1,
	}

	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			mgr.AddResources(resource)
		}
	})
	b.StopTimer()
}
