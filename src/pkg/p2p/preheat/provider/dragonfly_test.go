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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider/auth"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
	suite.dragonfly = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case healthCheckEndpoint:
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			w.WriteHeader(http.StatusOK)
		case preheatEndpoint:
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			image := &PreheatImage{}
			if err := json.Unmarshal(data, image); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			if image.Type == "image" &&
				image.URL == "https://harbor.com" &&
				image.ImageName == "busybox" &&
				image.Tag == "latest" {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ID":"dragonfly-id"}`))
				return
			}

			w.WriteHeader(http.StatusBadRequest)
		case strings.Replace(preheatTaskEndpoint, "{task_id}", "dragonfly-id", 1):
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}
			status := &dragonflyPreheatInfo{
				ID:         "dragonfly-id",
				StartTime:  time.Now().UTC().String(),
				FinishTime: time.Now().Add(5 * time.Minute).UTC().String(),
				Status:     "SUCCESS",
			}
			bytes, _ := json.Marshal(status)
			_, _ = w.Write(bytes)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))

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
	})
	require.NoError(suite.T(), err, "preheat image")
	suite.Equal("dragonfly-id", st.TaskID, "preheat image result")
}

// TestCheckProgress tests CheckProgress method.
func (suite *DragonflyTestSuite) TestCheckProgress() {
	st, err := suite.driver.CheckProgress("dragonfly-id")
	require.NoError(suite.T(), err, "get preheat status")
	suite.Equal(provider.PreheatingStatusSuccess, st.Status, "preheat status")
}
