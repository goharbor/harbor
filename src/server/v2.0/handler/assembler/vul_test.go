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
	"fmt"
	"testing"

	models "github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/testing/api/scan"
	"github.com/goharbor/harbor/src/testing/api/scanner"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type VulAssemblerTestSuite struct {
	suite.Suite
}

func (suite *VulAssemblerTestSuite) newVulAssembler(withScanOverview bool) (*VulAssembler, *scan.Controller, *scanner.Controller) {
	vulAssembler := NewVulAssembler(withScanOverview)

	scanCtl := &scan.Controller{}
	scannerCtl := &scanner.Controller{}

	vulAssembler.scanCtl = scanCtl
	vulAssembler.scannerCtl = scannerCtl

	return vulAssembler, scanCtl, scannerCtl
}

func (suite *VulAssemblerTestSuite) TestNotHasScanner() {
	{
		assembler, _, scannerCtl := suite.newVulAssembler(true)
		scannerCtl.On("GetRegistrationByProject", mock.AnythingOfType("int64")).Return(nil, nil)

		var artifact model.Artifact
		suite.Nil(assembler.WithArtifacts(&artifact).Assemble(context.TODO()))
		suite.Len(artifact.AdditionLinks, 0)
	}

	{
		assembler, _, scannerCtl := suite.newVulAssembler(true)
		scannerCtl.On("GetRegistrationByProject", mock.AnythingOfType("int64")).Return(nil, fmt.Errorf("error"))

		var artifact model.Artifact
		suite.Nil(assembler.WithArtifacts(&artifact).Assemble(context.TODO()))
		suite.Len(artifact.AdditionLinks, 0)
	}
}

func (suite *VulAssemblerTestSuite) TestHasScanner() {
	{
		assembler, scanCtl, scannerCtl := suite.newVulAssembler(true)
		scannerCtl.On("GetRegistrationByProject", mock.AnythingOfType("int64")).Return(&models.Registration{}, nil)

		summary := map[string]interface{}{"key": "value"}
		scanCtl.On("GetSummary", mock.AnythingOfType("*v1.Artifact"), mock.AnythingOfType("[]string")).Return(summary, nil)

		var artifact model.Artifact
		suite.Nil(assembler.WithArtifacts(&artifact).Assemble(context.TODO()))
		suite.Len(artifact.AdditionLinks, 1)
		suite.Equal(artifact.ScanOverview, summary)
	}

	{
		assembler, scanCtl, scannerCtl := suite.newVulAssembler(false)
		scannerCtl.On("GetRegistrationByProject", mock.AnythingOfType("int64")).Return(&models.Registration{}, nil)
		summary := map[string]interface{}{"key": "value"}
		scanCtl.On("GetSummary", mock.AnythingOfType("*v1.Artifact"), mock.AnythingOfType("[]string")).Return(summary, nil)

		var artifact model.Artifact
		suite.Nil(assembler.WithArtifacts(&artifact).Assemble(context.TODO()))
		suite.Len(artifact.AdditionLinks, 1)
		suite.Nil(artifact.ScanOverview)
	}

	{
		assembler, scanCtl, scannerCtl := suite.newVulAssembler(true)
		scannerCtl.On("GetRegistrationByProject", mock.AnythingOfType("int64")).Return(&models.Registration{}, nil)

		scanCtl.On("GetSummary", mock.AnythingOfType("*v1.Artifact"), mock.AnythingOfType("[]string")).Return(nil, fmt.Errorf("error"))

		var artifact model.Artifact
		suite.Nil(assembler.WithArtifacts(&artifact).Assemble(context.TODO()))
		suite.Len(artifact.AdditionLinks, 1)
		suite.Nil(artifact.ScanOverview)
	}
}

func TestVulAssemblerTestSuite(t *testing.T) {
	suite.Run(t, &VulAssemblerTestSuite{})
}
