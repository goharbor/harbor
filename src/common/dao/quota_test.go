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
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/suite"
)

var (
	quotaReference     = "dao"
	quotaUserReference = "user"
	quotaHard          = models.QuotaHard{"storage": 1024}
	quotaHardLarger    = models.QuotaHard{"storage": 2048}
)

type QuotaDaoSuite struct {
	suite.Suite
}

func (suite *QuotaDaoSuite) equalHard(quota1 *models.Quota, quota2 *models.Quota) {
	hard1, err := quota1.GetHard()
	suite.Nil(err, "hard1 invalid")

	hard2, err := quota2.GetHard()
	suite.Nil(err, "hard2 invalid")

	suite.Equal(hard1, hard2)
}

func (suite *QuotaDaoSuite) TearDownTest() {
	ClearTable("quota")
	ClearTable("quota_usage")
}

func (suite *QuotaDaoSuite) TestAddQuota() {
	_, err1 := AddQuota(models.Quota{Reference: quotaReference, ReferenceID: "1", Hard: quotaHard.String()})
	suite.Nil(err1)

	// Will failed for reference and reference_id should unique in db
	_, err2 := AddQuota(models.Quota{Reference: quotaReference, ReferenceID: "1", Hard: quotaHard.String()})
	suite.Error(err2)

	_, err3 := AddQuota(models.Quota{Reference: quotaUserReference, ReferenceID: "1", Hard: quotaHard.String()})
	suite.Nil(err3)
}

func (suite *QuotaDaoSuite) TestGetQuota() {
	quota1 := models.Quota{Reference: quotaReference, ReferenceID: "1", Hard: quotaHard.String()}
	id, err := AddQuota(quota1)
	suite.Nil(err)

	// Get the new added quota
	quota2, err := GetQuota(id)
	suite.Nil(err)
	suite.NotNil(quota2)

	// Get the quota which id is 10000 not found
	quota3, err := GetQuota(10000)
	suite.Nil(err)
	suite.Nil(quota3)
}

func (suite *QuotaDaoSuite) TestUpdateQuota() {
	quota1 := models.Quota{Reference: quotaReference, ReferenceID: "1", Hard: quotaHard.String()}
	id, err := AddQuota(quota1)
	suite.Nil(err)

	// Get the new added quota
	quota2, err := GetQuota(id)
	suite.Nil(err)
	suite.equalHard(&quota1, quota2)

	// Update the quota
	quota2.SetHard(quotaHardLarger)
	time.Sleep(time.Millisecond * 10) // Ensure that UpdateTime changed
	suite.Nil(UpdateQuota(*quota2))

	// Get the updated quota
	quota3, err := GetQuota(id)
	suite.Nil(err)
	suite.equalHard(quota2, quota3)
	suite.NotEqual(quota2.UpdateTime, quota3.UpdateTime)
}

func (suite *QuotaDaoSuite) TestListQuotas() {
	id1, _ := AddQuota(models.Quota{Reference: quotaReference, ReferenceID: "1", Hard: quotaHard.String()})
	AddQuotaUsage(models.QuotaUsage{ID: id1, Reference: quotaReference, ReferenceID: "1", Used: "{}"})

	id2, _ := AddQuota(models.Quota{Reference: quotaReference, ReferenceID: "2", Hard: quotaHard.String()})
	AddQuotaUsage(models.QuotaUsage{ID: id2, Reference: quotaReference, ReferenceID: "2", Used: "{}"})

	id3, _ := AddQuota(models.Quota{Reference: quotaUserReference, ReferenceID: "1", Hard: quotaHardLarger.String()})
	AddQuotaUsage(models.QuotaUsage{ID: id3, Reference: quotaUserReference, ReferenceID: "1", Used: "{}"})

	id4, _ := AddQuota(models.Quota{Reference: quotaReference, ReferenceID: "3", Hard: quotaHard.String()})
	AddQuotaUsage(models.QuotaUsage{ID: id4, Reference: quotaReference, ReferenceID: "3", Used: "{}"})

	// List all the quotas
	quotas, err := ListQuotas()
	suite.Nil(err)
	suite.Equal(4, len(quotas))
	suite.Equal(quotaReference, quotas[0].Reference)

	// List quotas filter by reference
	quotas, err = ListQuotas(&models.QuotaQuery{Reference: quotaReference})
	suite.Nil(err)
	suite.Equal(3, len(quotas))

	// List quotas filter by reference ids
	quotas, err = ListQuotas(&models.QuotaQuery{Reference: quotaReference, ReferenceIDs: []string{"1", "2"}})
	suite.Nil(err)
	suite.Equal(2, len(quotas))

	// List quotas by pagination
	quotas, err = ListQuotas(&models.QuotaQuery{Pagination: models.Pagination{Size: 2}})
	suite.Nil(err)
	suite.Equal(2, len(quotas))

	// List quotas by sorting
	quotas, err = ListQuotas(&models.QuotaQuery{Sorting: models.Sorting{Sort: "-hard.storage"}})
	suite.Nil(err)
	suite.Equal(quotaUserReference, quotas[0].Reference)
}

func TestRunQuotaDaoSuite(t *testing.T) {
	suite.Run(t, new(QuotaDaoSuite))
}

func Test_quotaOrderBy(t *testing.T) {
	query := func(sort string) []*models.QuotaQuery {
		return []*models.QuotaQuery{
			{Sorting: models.Sorting{Sort: sort}},
		}
	}

	type args struct {
		query []*models.QuotaQuery
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"no query", args{nil}, "b.creation_time DESC"},
		{"order by unsupport field", args{query("unknow")}, "b.creation_time DESC"},
		{"order by storage of hard", args{query("hard.storage")}, "(CAST( (CASE WHEN (hard->>'storage') IS NULL THEN '0' WHEN (hard->>'storage') = '-1' THEN '9223372036854775807' ELSE (hard->>'storage') END) AS BIGINT )) ASC"},
		{"order by unsupport hard resource", args{query("hard.unknow")}, "b.creation_time DESC"},
		{"order by storage of used", args{query("used.storage")}, "(CAST( (CASE WHEN (used->>'storage') IS NULL THEN '0' WHEN (used->>'storage') = '-1' THEN '9223372036854775807' ELSE (used->>'storage') END) AS BIGINT )) ASC"},
		{"order by unsupport used resource", args{query("used.unknow")}, "b.creation_time DESC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quotaOrderBy(tt.args.query...); got != tt.want {
				t.Errorf("quotaOrderBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
