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
	"context"
	"testing"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	scannertesting "github.com/goharbor/harbor/src/testing/controller/scanner"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type CheckerTestSuite struct {
	suite.Suite
}

func (suite *CheckerTestSuite) new() *checker {
	artifactCtl := &artifacttesting.Controller{}
	scannerCtl := &scannertesting.Controller{}

	return &checker{
		artifactCtl:   artifactCtl,
		scannerCtl:    scannerCtl,
		registrations: map[int64]*scanner.Registration{},
	}
}

func (suite *CheckerTestSuite) TestScannerNotFound() {
	c := suite.new()

	{
		mock.OnAnything(c.scannerCtl, "GetRegistrationByProject").Return(nil, nil)

		isScannable, err := c.IsScannable(context.TODO(), &artifact.Artifact{})
		suite.Nil(err)
		suite.False(isScannable)
	}
}

func (suite *CheckerTestSuite) TestIsScannable() {
	c := suite.new()

	supportMimeType := "support mime type"

	mock.OnAnything(c.scannerCtl, "GetRegistrationByProject").Return(&scanner.Registration{
		Metadata: &v1.ScannerAdapterMetadata{
			Capabilities: []*v1.ScannerCapability{
				{ConsumesMimeTypes: []string{supportMimeType}},
			},
		},
	}, nil)

	{
		art := &artifact.Artifact{}

		mock.OnAnything(c.artifactCtl, "Walk").Return(nil).Once().Run(func(args mock.Arguments) {
			walkFn := args.Get(2).(func(*artifact.Artifact) error)
			walkFn(art)
		})

		isScannable, err := c.IsScannable(context.TODO(), art)
		suite.Nil(err)
		suite.False(isScannable)
	}

	{
		art := &artifact.Artifact{}
		art.Type = "IMAGE"
		art.ManifestMediaType = supportMimeType

		mock.OnAnything(c.artifactCtl, "Walk").Return(nil).Once().Run(func(args mock.Arguments) {
			walkFn := args.Get(2).(func(*artifact.Artifact) error)
			walkFn(art)
		})

		isScannable, err := c.IsScannable(context.TODO(), art)
		suite.Nil(err)
		suite.True(isScannable)
	}
}

func TestCheckerTestSuite(t *testing.T) {
	suite.Run(t, &CheckerTestSuite{})
}
