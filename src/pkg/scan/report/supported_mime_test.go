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

package report

import (
	"encoding/json"
	"testing"
	"time"

	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SupportedMimesSuite is a suite to test SupportedMimes.
type SupportedMimesSuite struct {
	suite.Suite

	mockData []byte
}

// TestSupportedMimesSuite is the entry of SupportedMimesSuite.
func TestSupportedMimesSuite(t *testing.T) {
	suite.Run(t, new(SupportedMimesSuite))
}

// SetupSuite prepares the test suite env.
func (suite *SupportedMimesSuite) SetupSuite() {
	rp := vuln.Report{
		GeneratedAt: time.Now().UTC().String(),
		Scanner: &v1.Scanner{
			Name:    "Clair",
			Vendor:  "Harbor",
			Version: "0.1.0",
		},
		Severity: vuln.High,
		Vulnerabilities: []*vuln.VulnerabilityItem{
			{
				ID:          "2019-0980-0909",
				Package:     "dpkg",
				Version:     "0.9.1",
				FixVersion:  "0.9.2",
				Severity:    vuln.High,
				Description: "mock one",
				Links:       []string{"https://vuln.com"},
			},
		},
	}

	jsonData, err := json.Marshal(rp)
	require.NoError(suite.T(), err)
	suite.mockData = jsonData
}

// TestResolveData tests the ResolveData.
func (suite *SupportedMimesSuite) TestResolveData() {
	obj, err := ResolveData(v1.MimeTypeNativeReport, suite.mockData)
	require.NoError(suite.T(), err)
	require.Condition(suite.T(), func() (success bool) {
		rp, ok := obj.(*vuln.Report)
		success = ok && rp != nil && rp.Severity == vuln.High

		return
	})
}
