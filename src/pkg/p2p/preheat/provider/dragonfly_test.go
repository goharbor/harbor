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

package provider

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
)

// DragonflyTestSuite is a test suite of testing Dragonfly driver.
type DragonflyTestSuite struct {
	suite.Suite

	dragonfly *httptest.Server
	driver    *DragonflyDriver
}

// TestDragonfly is the entry method of running DragonflyTestSuite.
func TestDragonfly(t *testing.T) {
	suite.Run(t, &DragonflyTestSuite{})
}

// SetupSuite prepares the env for DragonflyTestSuite.
func (suite *DragonflyTestSuite) SetupSuite() {
	suite.dragonfly = MockDragonflyProvider()

	suite.dragonfly.StartTLS()

	suite.driver = &DragonflyDriver{
		instance: &provider.Instance{
			ID:       1,
			Name:     "test-instance",
			Vendor:   DriverDragonfly,
			Endpoint: suite.dragonfly.URL,
			AuthMode: auth.AuthModeNone,
			Enabled:  true,
			Default:  true,
			Insecure: true,
			Status:   DriverStatusHealthy,
		},
	}
}

// TearDownSuite clears the env for DragonflyTestSuite.
func (suite *DragonflyTestSuite) TearDownSuite() {
	suite.dragonfly.Close()
}

// TestSelf tests Self method.
func (suite *DragonflyTestSuite) TestSelf() {
	m := suite.driver.Self()
	suite.Equal(DriverDragonfly, m.ID, "self metadata")
}

// TestGetHealth tests GetHealth method.
func (suite *DragonflyTestSuite) TestGetHealth() {
	st, err := suite.driver.GetHealth()
	require.NoError(suite.T(), err, "get health")
	suite.Equal(DriverStatusHealthy, st.Status, "healthy status")
}

// TestPreheat tests Preheat method.
func (suite *DragonflyTestSuite) TestPreheat() {
	st, err := suite.driver.Preheat(&PreheatImage{
		Type:      "image",
		ImageName: "busybox",
		Tag:       "latest",
		URL:       "https://harbor.com",
		Digest:    "sha256:f3c97e3bd1e27393eb853a5c90b1132f2cda84336d5ba5d100c720dc98524c82",
		ExtraAttrs: map[string]interface{}{
			"scope":       "all_peers",
			"cluster_ids": []uint{1, 2, 3},
		},
	})
	require.NoError(suite.T(), err, "preheat image")
	suite.Equal(provider.PreheatingStatusPending, st.Status, "preheat status")
	suite.Equal("0", st.TaskID, "task id")
	suite.NotEmptyf(st.StartTime, "start time")
	suite.NotEmptyf(st.FinishTime, "finish time")
}

// TestCheckProgress tests CheckProgress method.
func (suite *DragonflyTestSuite) TestCheckProgress() {
	st, err := suite.driver.CheckProgress("1")
	require.NoError(suite.T(), err, "get image")
	suite.Equal(provider.PreheatingStatusRunning, st.Status, "preheat status")
	suite.Equal("1", st.TaskID, "task id")
	suite.NotEmptyf(st.StartTime, "start time")
	suite.NotEmptyf(st.FinishTime, "finish time")

	st, err = suite.driver.CheckProgress("2")
	require.NoError(suite.T(), err, "get image")
	suite.Equal(provider.PreheatingStatusSuccess, st.Status, "preheat status")
	suite.Equal("2", st.TaskID, "task id")
	suite.NotEmptyf(st.StartTime, "start time")
	suite.NotEmptyf(st.FinishTime, "finish time")

	st, err = suite.driver.CheckProgress("3")
	require.NoError(suite.T(), err, "get image")
	suite.Equal(provider.PreheatingStatusFail, st.Status, "preheat status")
	suite.Equal("3", st.TaskID, "task id")
	suite.NotEmptyf(st.StartTime, "start time")
	suite.NotEmptyf(st.FinishTime, "finish time")
}
