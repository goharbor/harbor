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
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// HookClientTestSuite tests functions of hook client
type HookClientTestSuite struct {
	suite.Suite

	mockServer *httptest.Server
	client     Client
}

// SetupSuite prepares test suite
func (suite *HookClientTestSuite) SetupSuite() {
	suite.client = NewClient(context.Background())
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		m := &Event{}
		err = json.Unmarshal(bytes, m)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if m.Data.JobID == "job_ID_failed" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "ok")
	}))
}

// TearDownSuite clears test suite
func (suite *HookClientTestSuite) TearDownSuite() {
	suite.mockServer.Close()
}

// TestHookClientTestSuite is entry of go test
func TestHookClientTestSuite(t *testing.T) {
	suite.Run(t, new(HookClientTestSuite))
}

// TestHookClient ...
func (suite *HookClientTestSuite) TestHookClient() {
	changeData := &job.StatusChange{
		JobID:  "fake_job_ID",
		Status: "running",
	}
	evt := &Event{
		URL:       suite.mockServer.URL,
		Data:      changeData,
		Message:   fmt.Sprintf("Status of job %s changed to: %s", changeData.JobID, changeData.Status),
		Timestamp: time.Now().Unix(),
	}
	err := suite.client.SendEvent(evt)
	assert.Nil(suite.T(), err, "send event: nil error expected but got %s", err)
}

// TestReportStatusFailed ...
func (suite *HookClientTestSuite) TestReportStatusFailed() {
	changeData := &job.StatusChange{
		JobID:  "job_ID_failed",
		Status: "running",
	}
	evt := &Event{
		URL:       suite.mockServer.URL,
		Data:      changeData,
		Message:   fmt.Sprintf("Status of job %s changed to: %s", changeData.JobID, changeData.Status),
		Timestamp: time.Now().Unix(),
	}

	err := suite.client.SendEvent(evt)
	assert.NotNil(suite.T(), err, "send event: expected non nil error but got nil")
}
