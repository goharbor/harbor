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
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	dr "github.com/goharbor/harbor/src/common/quota/driver"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/stretchr/testify/suite"
)

type DriverSuite struct {
	suite.Suite
}

func (suite *DriverSuite) TestHardLimits() {
	driver := newDriver()

	suite.Equal(types.ResourceList{types.ResourceCount: -1, types.ResourceStorage: -1}, driver.HardLimits())
}

func (suite *DriverSuite) TestLoad() {
	driver := newDriver()

	if ref, err := driver.Load("1"); suite.Nil(err) {
		obj := dr.RefObject{
			"id":         int64(1),
			"name":       "library",
			"owner_name": "",
		}

		suite.Equal(obj, ref)
	}

	if ref, err := driver.Load("100000"); suite.Error(err) {
		suite.Empty(ref)
	}

	if ref, err := driver.Load("library"); suite.Error(err) {
		suite.Empty(ref)
	}
}

func (suite *DriverSuite) TestValidate() {
	driver := newDriver()

	suite.Nil(driver.Validate(types.ResourceList{types.ResourceCount: 1, types.ResourceStorage: 1024}))
	suite.Error(driver.Validate(types.ResourceList{}))
	suite.Error(driver.Validate(types.ResourceList{types.ResourceCount: 1}))
	suite.Error(driver.Validate(types.ResourceList{types.ResourceCount: 1, types.ResourceStorage: 0}))
	suite.Error(driver.Validate(types.ResourceList{types.ResourceCount: 1, types.ResourceName("foo"): 1}))
}

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()

	os.Exit(m.Run())
}

func TestRunDriverSuite(t *testing.T) {
	suite.Run(t, new(DriverSuite))
}
