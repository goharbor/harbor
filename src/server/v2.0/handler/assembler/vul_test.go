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

package assembler

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/testing/api/scan"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type VulAssemblerTestSuite struct {
	suite.Suite
}

func (suite *VulAssemblerTestSuite) TestScannable() {
	checker := &scan.Checker{}
	scanCtl := &scan.Controller{}

	assembler := VulAssembler{
		scanChecker:      checker,
		scanCtl:          scanCtl,
		withScanOverview: true,
	}

	mock.OnAnything(checker, "IsScannable").Return(true, nil)

	summary := map[string]interface{}{"key": "value"}
	mock.OnAnything(scanCtl, "GetSummary").Return(summary, nil)

	var artifact model.Artifact

	suite.Nil(assembler.WithArtifacts(&artifact).Assemble(context.TODO()))
	suite.Len(artifact.AdditionLinks, 1)
	suite.Equal(artifact.ScanOverview, summary)
}

func (suite *VulAssemblerTestSuite) TestNotScannable() {
	checker := &scan.Checker{}
	scanCtl := &scan.Controller{}

	assembler := VulAssembler{
		scanChecker:      checker,
		scanCtl:          scanCtl,
		withScanOverview: true,
	}

	mock.OnAnything(checker, "IsScannable").Return(false, nil)

	summary := map[string]interface{}{"key": "value"}
	mock.OnAnything(scanCtl, "GetSummary").Return(summary, nil)

	var art model.Artifact

	suite.Nil(assembler.WithArtifacts(&art).Assemble(context.TODO()))
	suite.Len(art.AdditionLinks, 0)
	scanCtl.AssertNotCalled(suite.T(), "GetSummary")
}

func TestVulAssemblerTestSuite(t *testing.T) {
	suite.Run(t, &VulAssemblerTestSuite{})
}
