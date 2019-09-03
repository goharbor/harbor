// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/quota/driver"
	"github.com/goharbor/harbor/src/common/quota/driver/mocks"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/goharbor/harbor/src/testing/apitests/apilib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	reference  = "mock"
	hardLimits = types.ResourceList{types.ResourceCount: -1, types.ResourceStorage: -1}
)

func init() {
	mockDriver := &mocks.Driver{}

	mockHardLimitsFn := func() types.ResourceList {
		return hardLimits
	}

	mockLoadFn := func(key string) driver.RefObject {
		return driver.RefObject{"id": key}
	}

	mockValidateFn := func(hardLimits types.ResourceList) error {
		if len(hardLimits) == 0 {
			return fmt.Errorf("no resources found")
		}

		return nil
	}

	mockDriver.On("HardLimits").Return(mockHardLimitsFn)
	mockDriver.On("Load", mock.AnythingOfType("string")).Return(mockLoadFn, nil)
	mockDriver.On("Validate", mock.AnythingOfType("types.ResourceList")).Return(mockValidateFn)

	driver.Register(reference, mockDriver)
}

func TestQuotaAPIList(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	count := 10
	for i := 0; i < count; i++ {
		mgr, err := quota.NewManager(reference, fmt.Sprintf("%d", i))
		assert.Nil(err)

		_, err = mgr.NewQuota(hardLimits)
		assert.Nil(err)
	}

	code, quotas, err := apiTest.QuotasGet(&apilib.QuotaQuery{Reference: reference}, *admin)
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Len(quotas, count, fmt.Sprintf("quotas len should be %d", count))

	code, quotas, err = apiTest.QuotasGet(&apilib.QuotaQuery{Reference: reference, PageSize: 1}, *admin)
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Len(quotas, 1)
}

func TestQuotaAPIGet(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	mgr, err := quota.NewManager(reference, "quota-get")
	assert.Nil(err)

	quotaID, err := mgr.NewQuota(hardLimits)
	assert.Nil(err)

	code, quota, err := apiTest.QuotasGetByID(*admin, fmt.Sprintf("%d", quotaID))
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Equal(map[string]int64{"storage": -1, "count": -1}, quota.Hard)

	code, _, err = apiTest.QuotasGetByID(*admin, "100")
	assert.Nil(err)
	assert.Equal(int(404), code)
}

func TestQuotaPut(t *testing.T) {
	assert := assert.New(t)
	apiTest := newHarborAPI()

	mgr, err := quota.NewManager(reference, "quota-put")
	assert.Nil(err)

	quotaID, err := mgr.NewQuota(hardLimits)
	assert.Nil(err)

	code, quota, err := apiTest.QuotasGetByID(*admin, fmt.Sprintf("%d", quotaID))
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Equal(map[string]int64{"count": -1, "storage": -1}, quota.Hard)

	code, err = apiTest.QuotasPut(*admin, fmt.Sprintf("%d", quotaID), models.QuotaUpdateRequest{})
	assert.Nil(err, err)
	assert.Equal(int(400), code)

	code, err = apiTest.QuotasPut(*admin, fmt.Sprintf("%d", quotaID), models.QuotaUpdateRequest{Hard: types.ResourceList{types.ResourceCount: 100, types.ResourceStorage: 100}})
	assert.Nil(err)
	assert.Equal(int(200), code)

	code, quota, err = apiTest.QuotasGetByID(*admin, fmt.Sprintf("%d", quotaID))
	assert.Nil(err)
	assert.Equal(int(200), code)
	assert.Equal(map[string]int64{"count": 100, "storage": 100}, quota.Hard)
}
