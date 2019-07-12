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

package scan

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	past      = int64(1561967574)
	vulnList1 = VulnerabilityList{}
	vulnList2 = VulnerabilityList{
		{ID: "CVE-2018-10754",
			Severity: models.SevLow,
			Pkg:      "ncurses",
			Version:  "6.0+20161126-1+deb9u2",
		},
		{
			ID:       "CVE-2018-6485",
			Severity: models.SevHigh,
			Pkg:      "glibc",
			Version:  "2.24-11+deb9u4",
		},
	}
	whiteList1 = models.CVEWhitelist{
		ExpiresAt: &past,
		Items: []models.CVEWhitelistItem{
			{CVEID: "CVE-2018-6485"},
		},
	}
	whiteList2 = models.CVEWhitelist{
		Items: []models.CVEWhitelistItem{
			{CVEID: "CVE-2018-6485"},
		},
	}
	whiteList3 = models.CVEWhitelist{
		Items: []models.CVEWhitelistItem{
			{CVEID: "CVE-2018-6485"},
			{CVEID: "CVE-2018-10754"},
			{CVEID: "CVE-2019-12817"},
		},
	}
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

func TestVulnerabilityList_HasCVE(t *testing.T) {
	cases := []struct {
		input  VulnerabilityList
		cve    string
		result bool
	}{
		{
			input:  vulnList1,
			cve:    "CVE-2018-10754",
			result: false,
		},
		{
			input:  vulnList2,
			cve:    "CVE-2018-10754",
			result: true,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.result, c.input.HasCVE(c.cve))
	}
}

func TestVulnerabilityList_Severity(t *testing.T) {
	cases := []struct {
		input  VulnerabilityList
		expect models.Severity
	}{
		{
			input:  vulnList1,
			expect: models.SevNone,
		}, {
			input:  vulnList2,
			expect: models.SevHigh,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expect, c.input.Severity())
	}
}

func TestVulnerabilityList_ApplyWhitelist(t *testing.T) {
	cases := []struct {
		vl             VulnerabilityList
		wl             models.CVEWhitelist
		expectFiltered VulnerabilityList
		expectSev      models.Severity
	}{
		{
			vl:             vulnList2,
			wl:             whiteList1,
			expectFiltered: VulnerabilityList{},
			expectSev:      models.SevHigh,
		},
		{
			vl: vulnList2,
			wl: whiteList2,
			expectFiltered: VulnerabilityList{
				{
					ID:       "CVE-2018-6485",
					Severity: models.SevHigh,
					Pkg:      "glibc",
					Version:  "2.24-11+deb9u4",
				},
			},
			expectSev: models.SevLow,
		},
		{
			vl:             vulnList1,
			wl:             whiteList3,
			expectFiltered: VulnerabilityList{},
			expectSev:      models.SevNone,
		},
		{
			vl: vulnList2,
			wl: whiteList3,
			expectFiltered: VulnerabilityList{
				{ID: "CVE-2018-10754",
					Severity: models.SevLow,
					Pkg:      "ncurses",
					Version:  "6.0+20161126-1+deb9u2",
				},
				{
					ID:       "CVE-2018-6485",
					Severity: models.SevHigh,
					Pkg:      "glibc",
					Version:  "2.24-11+deb9u4",
				},
			},
			expectSev: models.SevNone,
		},
	}
	for _, c := range cases {
		filtered := c.vl.ApplyWhitelist(c.wl)
		assert.Equal(t, c.expectFiltered, filtered)
		assert.Equal(t, c.vl.Severity(), c.expectSev)
	}
}

func TestVulnListByDigest(t *testing.T) {
	_, err := VulnListByDigest("notexist")
	assert.NotNil(t, err)
}

func TestVulnListFromClairResult(t *testing.T) {
	l := VulnListFromClairResult(nil)
	assert.Equal(t, VulnerabilityList{}, l)
	lv := &models.ClairLayerEnvelope{
		Layer: nil,
		Error: nil,
	}
	l2 := VulnListFromClairResult(lv)
	assert.Equal(t, VulnerabilityList{}, l2)
}
