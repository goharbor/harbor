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

package chart

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/types"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

func deleteChartVersion(projectName, chartName, version string) {
	url := fmt.Sprintf("/api/%s/chartrepo/%s/charts/%s/%s", api.APIVersion, projectName, chartName, version)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	h := New(next)
	h.ServeHTTP(util.NewCustomResponseWriter(rr), req)
}

func uploadChartVersion(projectID int64, projectName, chartName, version string) {
	url := fmt.Sprintf("/api/%s/chartrepo/%s/charts/", api.APIVersion, projectName)
	req, _ := http.NewRequest(http.MethodPost, url, nil)

	info := &util.ChartVersionInfo{
		ProjectID: projectID,
		Namespace: projectName,
		ChartName: chartName,
		Version:   version,
	}
	*req = *req.WithContext(util.NewChartVersionInfoContext(req.Context(), info))

	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	rr := httptest.NewRecorder()
	h := New(next)
	h.ServeHTTP(util.NewCustomResponseWriter(rr), req)
}

func mockChartController() (*httptest.Server, *chartserver.Controller, error) {
	mockServer := httptest.NewServer(htesting.MockChartRepoHandler)

	var oldController, newController *chartserver.Controller
	url, err := url.Parse(mockServer.URL)
	if err == nil {
		newController, err = chartserver.NewController(url)
	}

	if err != nil {
		mockServer.Close()
		return nil, nil, err
	}

	chartController() // Init chart controller

	// Override current controller and keep the old one for restoring
	oldController = controller
	controller = newController

	return mockServer, oldController, nil
}

type HandlerSuite struct {
	htesting.Suite
	oldController   *chartserver.Controller
	mockChartServer *httptest.Server
}

func (suite *HandlerSuite) SetupTest() {
	mockServer, oldController, err := mockChartController()
	suite.Nil(err, "Mock chart controller failed")

	suite.oldController = oldController
	suite.mockChartServer = mockServer
}

func (suite *HandlerSuite) TearDownTest() {
	for _, table := range []string{
		"quota", "quota_usage",
	} {
		dao.ClearTable(table)
	}

	controller = suite.oldController
	suite.mockChartServer.Close()
}

func (suite *HandlerSuite) TestUpload() {
	suite.WithProject(func(projectID int64, projectName string) {
		uploadChartVersion(projectID, projectName, "harbor", "0.2.1")
		suite.AssertResourceUsage(1, types.ResourceCount, projectID)

		// harbor:0.2.0 exists in repo1, upload it again
		uploadChartVersion(projectID, projectName, "harbor", "0.2.0")
		suite.AssertResourceUsage(1, types.ResourceCount, projectID)
	}, "repo1")

	suite.WithProject(func(projectID int64, projectName string) {
		uploadChartVersion(projectID, projectName, "harbor-ha", "dev")
		suite.AssertResourceUsage(1, types.ResourceCount, projectID)
	}, "harbor-contrib")

	suite.WithProject(func(projectID int64, projectName string) {
		uploadChartVersion(projectID, projectName, "acs-engine-autoscaler", "1.0.0")
		suite.AssertResourceUsage(1, types.ResourceCount, projectID)
	}, "cluster-autoscaler")

	suite.WithProject(func(projectID int64, projectName string) {
		uploadChartVersion(projectID, projectName, "123456", "1-0")
		suite.AssertResourceUsage(1, types.ResourceCount, projectID)
	}, "123456")
}

func (suite *HandlerSuite) TestDelete() {
	suite.WithProject(func(projectID int64, projectName string) {
		uploadChartVersion(projectID, projectName, "harbor", "0.2.1")
		suite.AssertResourceUsage(1, types.ResourceCount, projectID)

		deleteChartVersion(projectName, "harbor", "0.2.1")
		suite.AssertResourceUsage(0, types.ResourceCount, projectID)
	}, "repo1")

	suite.WithProject(func(projectID int64, projectName string) {
		uploadChartVersion(projectID, projectName, "harbor-ha", "dev")
		suite.AssertResourceUsage(1, types.ResourceCount, projectID)

		deleteChartVersion(projectName, "harbor-ha", "dev")
		suite.AssertResourceUsage(0, types.ResourceCount, projectID)
	}, "harbor-contrib")

	suite.WithProject(func(projectID int64, projectName string) {
		uploadChartVersion(projectID, projectName, "acs-engine-autoscaler", "1.0.0")
		suite.AssertResourceUsage(1, types.ResourceCount, projectID)

		deleteChartVersion(projectName, "acs-engine-autoscaler", "1.0.0")
		suite.AssertResourceUsage(0, types.ResourceCount, projectID)
	}, "cluster-autoscaler")

	suite.WithProject(func(projectID int64, projectName string) {
		uploadChartVersion(projectID, projectName, "123456", "1-0")
		suite.AssertResourceUsage(1, types.ResourceCount, projectID)

		deleteChartVersion(projectName, "123456", "1-0")
		suite.AssertResourceUsage(0, types.ResourceCount, projectID)
	}, "123456")
}

func TestRunHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}
