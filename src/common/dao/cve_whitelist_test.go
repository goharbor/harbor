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
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUpdateAndGetCVEWhitelist(t *testing.T) {
	require.Nil(t, ClearTable("cve_whitelist"))
	l2, err := GetCVEWhitelist(5)
	assert.Nil(t, err)
	assert.Nil(t, l2)

	longList := []models.CVEWhitelistItem{}
	for i := 0; i < 50; i++ {
		longList = append(longList, models.CVEWhitelistItem{CVEID: "CVE-1999-0067"})
	}

	e := int64(1573254000)
	in1 := models.CVEWhitelist{ProjectID: 3, Items: longList, ExpiresAt: &e}
	_, err = UpdateCVEWhitelist(in1)
	require.Nil(t, err)
	// assert.Equal(t, int64(1), n)
	out1, err := GetCVEWhitelist(3)
	require.Nil(t, err)
	assert.Equal(t, int64(3), out1.ProjectID)
	assert.Equal(t, longList, out1.Items)
	assert.Equal(t, e, *out1.ExpiresAt)

	sysCVEs := []models.CVEWhitelistItem{
		{CVEID: "CVE-2019-10164"},
		{CVEID: "CVE-2017-12345"},
	}
	in3 := models.CVEWhitelist{Items: sysCVEs}
	_, err = UpdateCVEWhitelist(in3)
	require.Nil(t, err)

	require.Nil(t, ClearTable("cve_whitelist"))
}
