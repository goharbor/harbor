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

package project

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/quota/types"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	blobtesting "github.com/goharbor/harbor/src/testing/controller/blob"
	"github.com/goharbor/harbor/src/testing/mock"
)

type DriverTestSuite struct {
	suite.Suite

	artifactCtl *artifacttesting.Controller
	blobCtl     *blobtesting.Controller

	d *driver
}

func (suite *DriverTestSuite) SetupTest() {
	suite.artifactCtl = &artifacttesting.Controller{}
	suite.blobCtl = &blobtesting.Controller{}

	suite.d = &driver{
		blobCtl: suite.blobCtl,
	}
}

func (suite *DriverTestSuite) TestValidate() {
	testCases := []struct {
		description string
		input       types.ResourceList
		hasErr      bool
	}{
		{
			description: "quota limit is 0",
			input:       map[types.ResourceName]int64{types.ResourceStorage: 0},
			hasErr:      true,
		},
		{
			description: "quota limit is -1",
			input:       map[types.ResourceName]int64{types.ResourceStorage: -1},
			hasErr:      false,
		},
		{
			description: "quota limit is -2",
			input:       map[types.ResourceName]int64{types.ResourceStorage: -2},
			hasErr:      true,
		},
		{
			description: "quota limit is types.MaxLimitedValue",
			input:       map[types.ResourceName]int64{types.ResourceStorage: int64(types.MaxLimitedValue)},
			hasErr:      false,
		},
		{
			description: "quota limit is types.MaxLimitedValue + 1",
			input:       map[types.ResourceName]int64{types.ResourceStorage: int64(types.MaxLimitedValue + 1)},
			hasErr:      true,
		},
		{
			description: "quota limit is 12345",
			input:       map[types.ResourceName]int64{types.ResourceStorage: int64(12345)},
			hasErr:      false,
		},
	}

	for _, tc := range testCases {
		gotErr := suite.d.Validate(tc.input)
		if tc.hasErr {
			suite.Errorf(gotErr, "test case: %s", tc.description)
		} else {
			suite.NoErrorf(gotErr, "test case: %s", tc.description)
		}
	}
}

func (suite *DriverTestSuite) TestCalculateUsage() {

	{
		mock.OnAnything(suite.blobCtl, "CalculateTotalSizeByProject").Return(int64(1000), nil).Once()

		resources, err := suite.d.CalculateUsage(context.TODO(), "1")
		if suite.Nil(err) {
			suite.Len(resources, 1)
			suite.Equal(resources[types.ResourceStorage], int64(1000))
		}
	}
}

func TestDriverTestSuite(t *testing.T) {
	suite.Run(t, &DriverTestSuite{})
}
