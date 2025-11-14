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

package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/config/store"
)

// MockDriver is a mock implementation of store.Driver
type MockDriver struct {
	mock.Mock
}

func (m *MockDriver) Load(ctx context.Context) (map[string]any, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockDriver) Save(ctx context.Context, cfg map[string]any) error {
	args := m.Called(ctx, cfg)
	return args.Error(0)
}

func (m *MockDriver) Get(ctx context.Context, key string) (map[string]any, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]any), args.Error(1)
}

// GetItemFromDriverTestSuite tests the GetItemFromDriver method
type GetItemFromDriverTestSuite struct {
	suite.Suite
	ctx     context.Context
	manager *CfgManager
	driver  *MockDriver
}

func (suite *GetItemFromDriverTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.driver = &MockDriver{}
	suite.manager = &CfgManager{
		Store: store.NewConfigStore(suite.driver),
	}
}

func (suite *GetItemFromDriverTestSuite) TestGetItemFromDriverSuccess() {
	key := common.SkipAuditLogDatabase
	expectedResult := map[string]any{
		common.SkipAuditLogDatabase: true,
	}

	suite.driver.On("Get", suite.ctx, key).Return(expectedResult, nil)

	result, err := suite.manager.GetItemFromDriver(suite.ctx, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.driver.AssertExpectations(suite.T())
}

func (suite *GetItemFromDriverTestSuite) TestGetItemFromDriverError() {
	key := common.SkipAuditLogDatabase
	expectedError := errors.New("database connection failed")

	suite.driver.On("Get", suite.ctx, key).Return(map[string]any{}, expectedError)

	result, err := suite.manager.GetItemFromDriver(suite.ctx, key)

	suite.Require().Error(err)
	suite.Equal(expectedError, err)
	suite.Empty(result)
	suite.driver.AssertExpectations(suite.T())
}

func (suite *GetItemFromDriverTestSuite) TestGetItemFromDriverEmptyResult() {
	key := common.SkipAuditLogDatabase
	expectedResult := map[string]any{}

	suite.driver.On("Get", suite.ctx, key).Return(expectedResult, nil)

	result, err := suite.manager.GetItemFromDriver(suite.ctx, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.driver.AssertExpectations(suite.T())
}

func (suite *GetItemFromDriverTestSuite) TestGetItemFromDriverMultipleKeys() {
	key := common.AuditLogForwardEndpoint
	expectedResult := map[string]any{
		common.AuditLogForwardEndpoint: "syslog://localhost:514",
		common.SkipAuditLogDatabase:    false,
	}

	suite.driver.On("Get", suite.ctx, key).Return(expectedResult, nil)

	result, err := suite.manager.GetItemFromDriver(suite.ctx, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.driver.AssertExpectations(suite.T())
}

func (suite *GetItemFromDriverTestSuite) TestGetItemFromDriverNilContext() {
	key := common.SkipAuditLogDatabase
	expectedResult := map[string]any{
		common.SkipAuditLogDatabase: false,
	}

	suite.driver.On("Get", mock.Anything, key).Return(expectedResult, nil)

	result, err := suite.manager.GetItemFromDriver(nil, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.driver.AssertExpectations(suite.T())
}

func (suite *GetItemFromDriverTestSuite) TestGetItemFromDriverEmptyKey() {
	key := ""
	expectedResult := map[string]any{}

	suite.driver.On("Get", suite.ctx, key).Return(expectedResult, nil)

	result, err := suite.manager.GetItemFromDriver(suite.ctx, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.driver.AssertExpectations(suite.T())
}

func TestGetItemFromDriverTestSuite(t *testing.T) {
	suite.Run(t, new(GetItemFromDriverTestSuite))
}
