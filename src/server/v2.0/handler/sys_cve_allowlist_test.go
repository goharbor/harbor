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

package handler

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	allowlistModels "github.com/goharbor/harbor/src/pkg/allowlist/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	allowlisttesting "github.com/goharbor/harbor/src/testing/pkg/allowlist"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

type SysCVEAllowlistTestSuite struct {
	htesting.Suite

	mgr *allowlisttesting.Manager
}

func (suite *SysCVEAllowlistTestSuite) SetupSuite() {
	suite.mgr = &allowlisttesting.Manager{}

	suite.Config = &restapi.Config{
		SystemCVEAllowlistAPI: &systemCVEAllowListAPI{mgr: suite.mgr},
	}

	suite.Suite.SetupSuite()
}

func (suite *SysCVEAllowlistTestSuite) TestPutSystemCVEAllowlist_NoBody() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
	suite.mgr.On("SetSys", mock.Anything, allowlistModels.CVEAllowlist{}).Return(nil).Once()

	res, err := suite.Put("/system/CVEAllowlist", nil)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
}

func (suite *SysCVEAllowlistTestSuite) TestPutSystemCVEAllowlist_NullItem() {
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()

	res, err := suite.Put("/system/CVEAllowlist", bytes.NewBuffer([]byte(`{"items":[null]}`)))
	suite.NoError(err)
	suite.Equal(400, res.StatusCode)
}

func TestSysCVEAllowlistTestSuite(t *testing.T) {
	suite.Run(t, &SysCVEAllowlistTestSuite{})
}
