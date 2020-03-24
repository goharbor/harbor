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

	"github.com/goharbor/harbor/src/pkg/types"
	artifacttesting "github.com/goharbor/harbor/src/testing/api/artifact"
	blobtesting "github.com/goharbor/harbor/src/testing/api/blob"
	charttesting "github.com/goharbor/harbor/src/testing/api/chartmuseum"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type DriverTestSuite struct {
	suite.Suite

	artifactCtl *artifacttesting.Controller
	blobCtl     *blobtesting.Controller
	chartCtl    *charttesting.Controller

	d *driver
}

func (suite *DriverTestSuite) SetupTest() {
	suite.artifactCtl = &artifacttesting.Controller{}
	suite.blobCtl = &blobtesting.Controller{}
	suite.chartCtl = &charttesting.Controller{}

	suite.d = &driver{
		artifactCtl: suite.artifactCtl,
		blobCtl:     suite.blobCtl,
		chartCtl:    suite.chartCtl,
	}
}

func (suite *DriverTestSuite) TestCalculateUsage() {

	{
		mock.OnAnything(suite.artifactCtl, "Count").Return(int64(10), nil).Once()
		mock.OnAnything(suite.blobCtl, "CalculateTotalSizeByProject").Return(int64(1000), nil).Once()
		mock.OnAnything(suite.chartCtl, "Count").Return(int64(10), nil).Once()

		resources, err := suite.d.CalculateUsage(context.TODO(), "1")
		if suite.Nil(err) {
			suite.Len(resources, 2)
			suite.Equal(resources[types.ResourceCount], int64(20))
		}
	}
}

func TestDriverTestSuite(t *testing.T) {
	suite.Run(t, &DriverTestSuite{})
}
