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
	"os"
	"sync"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/stretchr/testify/suite"
)

var (
	hardLimits       = ResourceList{ResourceStorage: 1000}
	referenceProject = "project"
)

func mustResourceList(s string) ResourceList {
	resources, _ := NewResourceList(s)
	return resources
}

type ManagerSuite struct {
	suite.Suite
}

func (suite *ManagerSuite) quotaManager(referenceIDs ...string) *Manager {
	referenceID := "1"
	if len(referenceIDs) > 0 {
		referenceID = referenceIDs[0]
	}

	mgr, _ := NewManager(referenceProject, referenceID)
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
		suite.True(Equals(hardLimits, mustResourceList(quota.Hard)))
	}

	mgr = suite.quotaManager("2")
	used := ResourceList{ResourceStorage: 100}
	if id, err := mgr.NewQuota(hardLimits, used); suite.Nil(err) {
		quota, _ := dao.GetQuota(id)
		suite.True(Equals(hardLimits, mustResourceList(quota.Hard)))

		usage, _ := dao.GetQuotaUsage(id)
		suite.True(Equals(used, mustResourceList(usage.Used)))
	}
}

func (suite *ManagerSuite) TestAddResources() {
	mgr := suite.quotaManager()
	id, _ := mgr.NewQuota(hardLimits)

	resource := ResourceList{ResourceStorage: 100}

	if suite.Nil(mgr.AddResources(resource)) {
		usage, _ := dao.GetQuotaUsage(id)
		suite.True(Equals(resource, mustResourceList(usage.Used)))
	}

	if suite.Nil(mgr.AddResources(resource)) {
		usage, _ := dao.GetQuotaUsage(id)
		suite.True(Equals(ResourceList{ResourceStorage: 200}, mustResourceList(usage.Used)))
	}

	if err := mgr.AddResources(ResourceList{ResourceStorage: 10000}); suite.Error(err) {
		suite.True(IsUnsafeError(err))
	}
}

func (suite *ManagerSuite) TestSubtractResources() {
	mgr := suite.quotaManager()
	id, _ := mgr.NewQuota(hardLimits)

	resource := ResourceList{ResourceStorage: 100}

	if suite.Nil(mgr.AddResources(resource)) {
		usage, _ := dao.GetQuotaUsage(id)
		suite.True(Equals(resource, mustResourceList(usage.Used)))
	}

	if suite.Nil(mgr.SubtractResources(resource)) {
		usage, _ := dao.GetQuotaUsage(id)
		suite.True(Equals(ResourceList{ResourceStorage: 0}, mustResourceList(usage.Used)))
	}
}

func (suite *ManagerSuite) TestRaceAddResources() {
	mgr := suite.quotaManager()
	mgr.NewQuota(hardLimits)

	resources := ResourceList{
		ResourceStorage: 100,
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
	mgr.NewQuota(hardLimits, hardLimits)

	resources := ResourceList{
		ResourceStorage: 100,
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

	mgr, _ := NewManager(referenceProject, "1")
	mgr.NewQuota(ResourceList{ResourceStorage: int64(b.N)})

	resource := ResourceList{
		ResourceStorage: 1,
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

	mgr, _ := NewManager(referenceProject, "1")
	mgr.NewQuota(ResourceList{ResourceStorage: -1})

	resource := ResourceList{
		ResourceStorage: 1,
	}

	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			mgr.AddResources(resource)
		}
	})
	b.StopTimer()
}
