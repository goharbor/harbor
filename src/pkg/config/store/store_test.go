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

package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/errors"
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

// GetFromDriverTestSuite tests the GetFromDriver method in ConfigStore
type GetFromDriverTestSuite struct {
	suite.Suite
	ctx    context.Context
	store  *ConfigStore
	driver *MockDriver
}

func (suite *GetFromDriverTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.driver = &MockDriver{}
	suite.store = &ConfigStore{
		cfgDriver: suite.driver,
	}
}

// TestGetFromDriverSuccess tests successful retrieval from driver
func (suite *GetFromDriverTestSuite) TestGetFromDriverSuccess() {
	key := common.SkipAuditLogDatabase
	expectedResult := map[string]any{
		common.SkipAuditLogDatabase: true,
	}

	suite.driver.On("Get", suite.ctx, key).Return(expectedResult, nil)

	result, err := suite.store.GetFromDriver(suite.ctx, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.driver.AssertExpectations(suite.T())
}

// TestGetFromDriverNilDriver tests behavior when driver is nil
func (suite *GetFromDriverTestSuite) TestGetFromDriverNilDriver() {
	key := common.SkipAuditLogDatabase
	suite.store.cfgDriver = nil

	result, err := suite.store.GetFromDriver(suite.ctx, key)

	suite.Require().Error(err)
	suite.Contains(err.Error(), "failed to load store, cfgDriver is nil")
	suite.Nil(result)
}

// TestGetFromDriverError tests error handling when driver returns error
func (suite *GetFromDriverTestSuite) TestGetFromDriverError() {
	key := common.SkipAuditLogDatabase
	expectedError := errors.New("database connection failed")

	suite.driver.On("Get", suite.ctx, key).Return(map[string]any{}, expectedError)

	result, err := suite.store.GetFromDriver(suite.ctx, key)

	suite.Require().Error(err)
	suite.Equal(expectedError, err)
	suite.Empty(result)
	suite.driver.AssertExpectations(suite.T())
}

// TestGetFromDriverEmptyResult tests when driver returns empty result
func (suite *GetFromDriverTestSuite) TestGetFromDriverEmptyResult() {
	key := common.SkipAuditLogDatabase
	expectedResult := map[string]any{}

	suite.driver.On("Get", suite.ctx, key).Return(expectedResult, nil)

	result, err := suite.store.GetFromDriver(suite.ctx, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.driver.AssertExpectations(suite.T())
}

// TestGetFromDriverMultipleConfigs tests retrieval with multiple configurations
func (suite *GetFromDriverTestSuite) TestGetFromDriverMultipleConfigs() {
	key := common.AuditLogForwardEndpoint
	expectedResult := map[string]any{
		common.AuditLogForwardEndpoint: "syslog://localhost:514",
		common.SkipAuditLogDatabase:    false,
		"other_config":                 "value",
	}

	suite.driver.On("Get", suite.ctx, key).Return(expectedResult, nil)

	result, err := suite.store.GetFromDriver(suite.ctx, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.Equal("syslog://localhost:514", result[common.AuditLogForwardEndpoint])
	suite.Equal(false, result[common.SkipAuditLogDatabase])
	suite.Equal("value", result["other_config"])
	suite.driver.AssertExpectations(suite.T())
}

// TestGetFromDriverNilContext tests behavior with nil context
func (suite *GetFromDriverTestSuite) TestGetFromDriverNilContext() {
	key := common.SkipAuditLogDatabase
	expectedResult := map[string]any{
		common.SkipAuditLogDatabase: false,
	}

	suite.driver.On("Get", mock.Anything, key).Return(expectedResult, nil)

	result, err := suite.store.GetFromDriver(nil, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.driver.AssertExpectations(suite.T())
}

// TestGetFromDriverEmptyKey tests behavior with empty key
func (suite *GetFromDriverTestSuite) TestGetFromDriverEmptyKey() {
	key := ""
	expectedResult := map[string]any{}

	suite.driver.On("Get", suite.ctx, key).Return(expectedResult, nil)

	result, err := suite.store.GetFromDriver(suite.ctx, key)

	suite.Require().NoError(err)
	suite.Equal(expectedResult, result)
	suite.driver.AssertExpectations(suite.T())
}

// TestGetFromDriverDifferentKeys tests retrieval with different configuration keys
func (suite *GetFromDriverTestSuite) TestGetFromDriverDifferentKeys() {
	testCases := []struct {
		name           string
		key            string
		expectedResult map[string]any
	}{
		{
			name: "skip_audit_log_database",
			key:  common.SkipAuditLogDatabase,
			expectedResult: map[string]any{
				common.SkipAuditLogDatabase: true,
			},
		},
		{
			name: "audit_log_forward_endpoint",
			key:  common.AuditLogForwardEndpoint,
			expectedResult: map[string]any{
				common.AuditLogForwardEndpoint: "syslog://remote:514",
			},
		},
		{
			name: "pull_audit_log_disable",
			key:  common.PullAuditLogDisable,
			expectedResult: map[string]any{
				common.PullAuditLogDisable: false,
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.driver.On("Get", suite.ctx, tc.key).Return(tc.expectedResult, nil)

			result, err := suite.store.GetFromDriver(suite.ctx, tc.key)

			suite.Require().NoError(err)
			suite.Equal(tc.expectedResult, result)
			suite.driver.AssertExpectations(suite.T())

			// Reset mock for next iteration
			suite.driver.ExpectedCalls = nil
		})
	}
}

// Run the test suite
func TestGetFromDriverTestSuite(t *testing.T) {
	suite.Run(t, new(GetFromDriverTestSuite))
}
