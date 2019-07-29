// Copyright 2018 Project Harbor Authors
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

package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/goharbor/harbor/src/chartserver"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/stretchr/testify/assert"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

var (
	crOldController *chartserver.Controller
	crMockServer    *httptest.Server
	uploadMu        sync.RWMutex
)

func TestIsMultipartFormData(t *testing.T) {
	req, err := createRequest(http.MethodPost, "/api/chartrepo/charts")
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(headerContentType, "application/json")
	if isMultipartFormData(req) {
		t.Fatal("expect false result but got true")
	}

	req.Header.Set(headerContentType, contentTypeMultipart)
	if !isMultipartFormData(req) {
		t.Fatalf("expect %s result but got %s", contentTypeMultipart, req.Header.Get(headerContentType))
	}
}

// Test namespace cheking
func TestRequireNamespace(t *testing.T) {
	chartAPI := &ChartRepositoryAPI{}
	chartAPI.ProjectMgr = &mockProjectManager{}

	if !chartAPI.requireNamespace("library") {
		t.Fatal("expect namespace 'library' existing but got false")
	}
}

// Prepare
func TestPrepareEnv(t *testing.T) {
	var err error
	crMockServer, crOldController, err = mockChartController()
	if err != nil {
		t.Fatalf("Failed to start mock chart service with error: %s", err)
	}
}

// Test get health
func TestGetHealthStatus(t *testing.T) {
	status := make(map[string]interface{})
	err := handleAndParse(&testingRequest{
		url:        "/api/chartrepo/health",
		method:     http.MethodGet,
		credential: sysAdmin,
	}, &status)

	if err != nil {
		t.Fatal(err)
	}

	if _, ok := status["health"]; !ok {
		t.Fatal("expect 'health' but got nil")
	}
}

// Test get index by repo
func TestGetIndexByRepo(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/chartrepo/library/index.yaml",
			method:     http.MethodGet,
			credential: projDeveloper,
		},
		code: http.StatusOK,
	})
}

// Test get index
func TestGetIndex(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/chartrepo/index.yaml",
			method:     http.MethodGet,
			credential: sysAdmin,
		},
		code: http.StatusOK,
	})
}

// Test download chart
func TestDownloadChart(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/chartrepo/library/charts/harbor-0.2.0.tgz",
			method:     http.MethodGet,
			credential: projDeveloper,
		},
		code: http.StatusOK,
	})
}

// Test get charts
func TestListCharts(t *testing.T) {
	charts := make([]*chartserver.ChartInfo, 0)
	err := handleAndParse(&testingRequest{
		url:        "/api/chartrepo/library/charts",
		method:     http.MethodGet,
		credential: projAdmin,
	}, &charts)

	if err != nil {
		t.Fatal(err)
	}

	if len(charts) != 2 {
		t.Fatalf("expect 2 charts but got %d", len(charts))
	}
}

// Test get chart versions
func TestListChartVersions(t *testing.T) {
	chartVersions := make(chartserver.ChartVersions, 0)
	err := handleAndParse(&testingRequest{
		url:        "/api/chartrepo/library/charts/harbor",
		method:     http.MethodGet,
		credential: projAdmin,
	}, &chartVersions)

	if err != nil {
		t.Fatal(err)
	}

	if len(chartVersions) != 2 {
		t.Fatalf("expect 2 chart versions but got %d", len(chartVersions))
	}
}

// Test get chart version details
func TestGetChartVersion(t *testing.T) {
	chartV := &chartserver.ChartVersionDetails{}
	err := handleAndParse(&testingRequest{
		url:        "/api/chartrepo/library/charts/harbor/0.2.0",
		method:     http.MethodGet,
		credential: projAdmin,
	}, chartV)

	if err != nil {
		t.Fatal(err)
	}

	if chartV.Metadata.GetName() != "harbor" {
		t.Fatalf("expect get chart 'harbor' but got %s", chartV.Metadata.GetName())
	}

	if chartV.Metadata.GetVersion() != "0.2.0" {
		t.Fatalf("expect get chart version '0.2.0' but got %s", chartV.Metadata.GetVersion())
	}
}

// Test delete chart version
func TestDeleteChartVersion(t *testing.T) {
	assert := assert.New(t)
	TestPrepareEnv(t)
	defer TestClearEnv(t)

	var count1 int64
	if _, err := uploadChartVersion("library", "harbor", "0.2.1"); assert.Nil(err, "Upload chart should be success") {
		count1, _ = getProjectCountUsage(1)
	}

	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/api/chartrepo/library/charts/harbor/0.2.1",
			method:     http.MethodDelete,
			credential: projAdmin,
		},
		code: http.StatusOK,
	})

	if count2, err := getProjectCountUsage(1); assert.Nil(err) {
		assert.Equal(int64(1), count1-count2, "Delete one chart version package should release 1 count usage")
	}
}

// Test delete chart
func TestDeleteChart(t *testing.T) {
	assert := assert.New(t)
	TestPrepareEnv(t)
	defer TestClearEnv(t)

	count, _ := getProjectCountUsage(1)

	for i, version := range []string{"0.2.0", "0.2.1"} {
		_, err := uploadChartVersion("library", "harbor", version)
		assert.Nil(err, "Upload helm chart ")
		if c, err := getProjectCountUsage(1); assert.Nil(err) {
			assert.Equal(int64(i+1)+count, c)
		}
	}

	count1, _ := getProjectCountUsage(1)

	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:        "/api/chartrepo/library/charts/harbor",
			method:     http.MethodDelete,
			credential: projAdmin,
		},
		code: http.StatusOK,
	})

	if count2, err := getProjectCountUsage(1); assert.Nil(err) {
		assert.Equal(int64(2), count1-count2, "Delete chart with two version packages should release 2 count usage")
	}
}

func TestUploadChart(t *testing.T) {
	assert := assert.New(t)
	TestPrepareEnv(t)
	defer TestClearEnv(t)

	uploadURL := "/api/chartrepo/library/charts"

	checkCountUsage := func(expected int64) {
		if count, err := getProjectCountUsage(1); assert.Nil(err) {
			assert.Equal(expected, count)
		}
	}

	count, err := getProjectCountUsage(1)
	assert.Nil(err, "Get project count usage should be success")

	chartVersionExists = func(string, string, string) bool {
		return false
	}

	parseChart = func(*http.Request) (*chart.Chart, error) {
		return nil, errors.New("parse chart failed")
	}

	// Parse chart failed from request
	res, _ := handle(&testingRequest{
		url:        uploadURL,
		method:     http.MethodPost,
		credential: sysAdmin,
	})
	assert.Equal(http.StatusBadRequest, res.Code)

	parseChart = func(*http.Request) (*chart.Chart, error) {
		metadata := &chart.Metadata{Name: "redis", Version: "1.0"}
		return &chart.Chart{Metadata: metadata}, nil
	}

	// Upload chart not exists in repo
	res, _ = handle(&testingRequest{
		url:        uploadURL,
		method:     http.MethodPost,
		credential: sysAdmin,
	})
	if assert.Equal(http.StatusCreated, res.Code) {
		checkCountUsage(1 + count)
	}

	// Upload chart exists in repo
	chartVersionExists = func(string, string, string) bool {
		return true
	}
	res, _ = handle(&testingRequest{
		url:        uploadURL,
		method:     http.MethodPost,
		credential: sysAdmin,
	})
	if assert.Equal(http.StatusCreated, res.Code) {
		checkCountUsage(1 + count)
	}

	// Upload another chart
	chartVersionExists = func(string, string, string) bool {
		return false
	}
	res, _ = handle(&testingRequest{
		url:        uploadURL,
		method:     http.MethodPost,
		credential: sysAdmin,
	})
	if assert.Equal(http.StatusCreated, res.Code) {
		checkCountUsage(2 + count)
	}
}

// Clear
func TestClearEnv(t *testing.T) {
	crMockServer.Close()
	chartController = crOldController
}

func createRequest(method string, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.RequestURI = url

	return req, nil
}

// Mock project manager
type mockProjectManager struct{}

func (mpm *mockProjectManager) Get(projectIDOrName interface{}) (*models.Project, error) {
	return nil, errors.New("Not implemented")
}

func (mpm *mockProjectManager) Create(*models.Project) (int64, error) {
	return -1, errors.New("Not implemented")
}

func (mpm *mockProjectManager) Delete(projectIDOrName interface{}) error {
	return errors.New("Not implemented")
}

func (mpm *mockProjectManager) Update(projectIDOrName interface{}, project *models.Project) error {
	return errors.New("Not implemented")
}
func (mpm *mockProjectManager) List(query *models.ProjectQueryParam) (*models.ProjectQueryResult, error) {
	results := &models.ProjectQueryResult{
		Total:    2,
		Projects: make([]*models.Project, 0),
	}

	results.Projects = append(results.Projects, &models.Project{ProjectID: 0, Name: "library"})
	results.Projects = append(results.Projects, &models.Project{ProjectID: 1, Name: "repo2"})

	return results, nil
}

func (mpm *mockProjectManager) IsPublic(projectIDOrName interface{}) (bool, error) {
	return false, errors.New("Not implemented")
}

func (mpm *mockProjectManager) Exists(projectIDOrName interface{}) (bool, error) {
	if projectIDOrName == nil {
		return false, errors.New("nil projectIDOrName")
	}

	if ns, ok := projectIDOrName.(string); ok {
		if ns == "library" {
			return true, nil
		}

		return false, nil
	}

	return false, errors.New("unknown type of projectIDOrName")
}

// get all public project
func (mpm *mockProjectManager) GetPublic() ([]*models.Project, error) {
	return nil, errors.New("Not implemented")
}

// if the project manager uses a metadata manager, return it, otherwise return nil
func (mpm *mockProjectManager) GetMetadataManager() metamgr.ProjectMetadataManager {
	return nil
}

// mock security context
type mockSecurityContext struct{}

// IsAuthenticated returns whether the context has been authenticated or not
func (msc *mockSecurityContext) IsAuthenticated() bool {
	return true
}

// GetUsername returns the username of user related to the context
func (msc *mockSecurityContext) GetUsername() string {
	return "amdin"
}

// IsSysAdmin returns whether the user is system admin
func (msc *mockSecurityContext) IsSysAdmin() bool {
	return true
}

// IsSolutionUser returns whether the user is solution user
func (msc *mockSecurityContext) IsSolutionUser() bool {
	return false
}

// Can returns whether the user can do action on resource
func (msc *mockSecurityContext) Can(action rbac.Action, resource rbac.Resource) bool {
	namespace, err := resource.GetNamespace()
	if err != nil || namespace.Kind() != "project" {
		return false
	}

	projectIDOrName := namespace.Identity()

	if projectIDOrName == nil {
		return false
	}

	if ns, ok := projectIDOrName.(string); ok {
		if ns == "library" {
			return true
		}
	}

	return false
}

// Get current user's all project
func (msc *mockSecurityContext) GetMyProjects() ([]*models.Project, error) {
	return []*models.Project{{ProjectID: 0, Name: "library"}}, nil
}

// Get user's role in provided project
func (msc *mockSecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	return []int{0, 1, 2, 3}
}

func uploadChartVersion(projectName, chartName, version string, exists ...bool) (*httptest.ResponseRecorder, error) {
	uploadMu.Lock()
	defer uploadMu.Unlock()

	oldchartVersionExists := chartVersionExists
	oldParseChart := parseChart

	defer func() {
		chartVersionExists = oldchartVersionExists
		parseChart = oldParseChart
	}()

	chartVersionExists = func(string, string, string) bool {
		if len(exists) > 0 {
			return exists[0]
		}
		return false
	}

	parseChart = func(*http.Request) (*chart.Chart, error) {
		metadata := &chart.Metadata{Name: chartName, Version: version}
		return &chart.Chart{Metadata: metadata}, nil
	}

	return handle(&testingRequest{
		url:        fmt.Sprintf("/api/chartrepo/%s/charts", projectName),
		method:     http.MethodPost,
		credential: sysAdmin,
	})
}

func getProjectCountUsage(projectID int64) (int64, error) {
	usage := models.QuotaUsage{Reference: "project", ReferenceID: fmt.Sprintf("%d", projectID)}
	err := dao.GetOrmer().Read(&usage, "reference", "reference_id")
	if err != nil {
		return 0, err
	}
	used, err := types.NewResourceList(usage.Used)
	if err != nil {
		return 0, err
	}

	return used[types.ResourceCount], nil
}
