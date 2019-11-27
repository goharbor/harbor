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
	quotaUsageReference     = "project"
	quotaUsageUserReference = "user"
	quotaUsageUsed          = models.QuotaUsed{"storage": 1024}
	quotaUsageUsedLarger    = models.QuotaUsed{"storage": 2048}
)

type QuotaUsageDaoSuite struct {
	suite.Suite
}

func (suite *QuotaUsageDaoSuite) equalUsed(usage1 *models.QuotaUsage, usage2 *models.QuotaUsage) {
	used1, err := usage1.GetUsed()
	suite.Nil(err, "used1 invalid")

	used2, err := usage2.GetUsed()
	suite.Nil(err, "used2 invalid")

	suite.Equal(used1, used2)
}

func (suite *QuotaUsageDaoSuite) TearDownTest() {
	ClearTable("quota_usage")
}

func (suite *QuotaUsageDaoSuite) TestAddQuotaUsage() {
	_, err1 := AddQuotaUsage(models.QuotaUsage{Reference: quotaUsageReference, ReferenceID: "1", Used: quotaUsageUsed.String()})
	suite.Nil(err1)

	// Will failed for reference and reference_id should unique in db
	_, err2 := AddQuotaUsage(models.QuotaUsage{Reference: quotaUsageReference, ReferenceID: "1", Used: quotaUsageUsed.String()})
	suite.Error(err2)

	_, err3 := AddQuotaUsage(models.QuotaUsage{Reference: quotaUsageUserReference, ReferenceID: "1", Used: quotaUsageUsed.String()})
	suite.Nil(err3)
}

func (suite *QuotaUsageDaoSuite) TestGetQuotaUsage() {
	quotaUsage1 := models.QuotaUsage{Reference: quotaUsageReference, ReferenceID: "1", Used: quotaUsageUsed.String()}
	id, err := AddQuotaUsage(quotaUsage1)
	suite.Nil(err)

	// Get the new added quotaUsage
	quotaUsage2, err := GetQuotaUsage(id)
	suite.Nil(err)
	suite.NotNil(quotaUsage2)

	// Get the quotaUsage which id is 10000 not found
	quotaUsage3, err := GetQuotaUsage(10000)
	suite.Nil(err)
	suite.Nil(quotaUsage3)
}

func (suite *QuotaUsageDaoSuite) TestUpdateQuotaUsage() {
	quotaUsage1 := models.QuotaUsage{Reference: quotaUsageReference, ReferenceID: "1", Used: quotaUsageUsed.String()}
	id, err := AddQuotaUsage(quotaUsage1)
	suite.Nil(err)

	// Get the new added quotaUsage
	quotaUsage2, err := GetQuotaUsage(id)
	suite.Nil(err)
	suite.equalUsed(&quotaUsage1, quotaUsage2)

	// Update the quotaUsage
	quotaUsage2.SetUsed(quotaUsageUsedLarger)
	time.Sleep(time.Millisecond * 10) // Ensure that UpdateTime changed
	suite.Nil(UpdateQuotaUsage(*quotaUsage2))

	// Get the updated quotaUsage
	quotaUsage3, err := GetQuotaUsage(id)
	suite.Nil(err)
	suite.equalUsed(quotaUsage2, quotaUsage3)
	suite.NotEqual(quotaUsage2.UpdateTime, quotaUsage3.UpdateTime)
}

func (suite *QuotaUsageDaoSuite) TestListQuotaUsages() {
	AddQuotaUsage(models.QuotaUsage{Reference: quotaUsageReference, ReferenceID: "1", Used: quotaUsageUsed.String()})
	AddQuotaUsage(models.QuotaUsage{Reference: quotaUsageReference, ReferenceID: "2", Used: quotaUsageUsed.String()})
	AddQuotaUsage(models.QuotaUsage{Reference: quotaUsageReference, ReferenceID: "3", Used: quotaUsageUsed.String()})
	AddQuotaUsage(models.QuotaUsage{Reference: quotaUsageUserReference, ReferenceID: "1", Used: quotaUsageUsedLarger.String()})

	// List all the quotaUsages
	quotaUsages, err := ListQuotaUsages()
	suite.Nil(err)
	suite.Equal(4, len(quotaUsages))
	suite.Equal(quotaUsageReference, quotaUsages[0].Reference)

	// List quotaUsages filter by reference
	quotaUsages, err = ListQuotaUsages(&models.QuotaUsageQuery{Reference: quotaUsageReference})
	suite.Nil(err)
	suite.Equal(3, len(quotaUsages))

	// List quotaUsages filter by reference ids
	quotaUsages, err = ListQuotaUsages(&models.QuotaUsageQuery{Reference: quotaUsageReference, ReferenceIDs: []string{"1", "2"}})
	suite.Nil(err)
	suite.Equal(2, len(quotaUsages))

	// List quotaUsages by pagination
	quotaUsages, err = ListQuotaUsages(&models.QuotaUsageQuery{Pagination: models.Pagination{Size: 2}})
	suite.Nil(err)
	suite.Equal(2, len(quotaUsages))

	// List quotaUsages by sorting
	quotaUsages, err = ListQuotaUsages(&models.QuotaUsageQuery{Sorting: models.Sorting{Sort: "-used.storage"}})
	suite.Nil(err)
	suite.Equal(quotaUsageUserReference, quotaUsages[0].Reference)
}

func TestRunQuotaUsageDaoSuite(t *testing.T) {
	suite.Run(t, new(QuotaUsageDaoSuite))
}

func Test_quotaUsageOrderBy(t *testing.T) {
	query := func(sort string) []*models.QuotaUsageQuery {
		return []*models.QuotaUsageQuery{
			{Sorting: models.Sorting{Sort: sort}},
		}
	}

	type args struct {
		query []*models.QuotaUsageQuery
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"no query", args{nil}, ""},
		{"order by unsupport field", args{query("unknow")}, ""},
		{"order by count of used", args{query("used.count")}, "(CAST( (CASE WHEN (used->>'count') IS NULL THEN '0' WHEN (used->>'count') = '-1' THEN '9223372036854775807' ELSE (used->>'count') END) AS BIGINT )) ASC"},
		{"order by storage of used", args{query("used.storage")}, "(CAST( (CASE WHEN (used->>'storage') IS NULL THEN '0' WHEN (used->>'storage') = '-1' THEN '9223372036854775807' ELSE (used->>'storage') END) AS BIGINT )) ASC"},
		{"order by unsupport used resource", args{query("used.unknow")}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quotaUsageOrderBy(tt.args.query...); got != tt.want {
				t.Errorf("quotaUsageOrderBy() = %v, want %v", got, tt.want)
			}
		})
	}
}
