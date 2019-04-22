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

package runtime

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// BootStrapTestSuite tests bootstrap
type BootStrapTestSuite struct {
	suite.Suite

	jobService *Bootstrap
	cancel     context.CancelFunc
	ctx        context.Context
}

// SetupSuite prepares test suite
func (suite *BootStrapTestSuite) SetupSuite() {
	// Load configurations
	err := config.DefaultConfig.Load("../config_test.yml", true)
	require.NoError(suite.T(), err, "load configurations error: %s", err)

	// Append node ID
	vCtx := context.WithValue(context.Background(), utils.NodeID, utils.GenerateNodeID())
	// Create the root context
	suite.ctx, suite.cancel = context.WithCancel(vCtx)

	// Initialize logger
	err = logger.Init(suite.ctx)
	require.NoError(suite.T(), err, "init logger: nil error expected but got %s", err)

	suite.jobService = &Bootstrap{}
	suite.jobService.SetJobContextInitializer(nil)
}

// TearDownSuite clears the test suite
func (suite *BootStrapTestSuite) TearDownSuite() {
	suite.cancel()

	pool := tests.GiveMeRedisPool()
	conn := pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	_ = tests.ClearAll(tests.GiveMeTestNamespace(), conn)
}

// TestBootStrapTestSuite is entry of go test
func TestBootStrapTestSuite(t *testing.T) {
	suite.Run(t, new(BootStrapTestSuite))
}

// TestBootStrap tests bootstrap
func (suite *BootStrapTestSuite) TestBootStrap() {
	go func() {
		var err error
		defer func() {
			require.NoError(suite.T(), err, "load and run: nil error expected but got %s", err)
		}()

		err = suite.jobService.LoadAndRun(suite.ctx, suite.cancel)
	}()

	<-time.After(1 * time.Second)
	suite.cancel()
}
