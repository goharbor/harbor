package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	url2 "net/url"
	"strings"
	"testing"
	"time"

	beegoorm "github.com/beego/beego/v2/client/orm"
	"github.com/goharbor/harbor/src/lib/errors"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/scan_data_export"
	"github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/controller/scandataexport"
	"github.com/goharbor/harbor/src/testing/mock"
	systemartifacttesting "github.com/goharbor/harbor/src/testing/pkg/systemartifact"
	"github.com/goharbor/harbor/src/testing/pkg/user"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

type ScanExportTestSuite struct {
	htesting.Suite
	scanExportCtl  *scandataexport.Controller
	proCtl         *project.Controller
	sysArtifactMgr *systemartifacttesting.Manager
	userMgr        *user.Manager
}

func (suite *ScanExportTestSuite) SetupSuite() {

}

func (suite *ScanExportTestSuite) SetupTest() {

	suite.scanExportCtl = &scandataexport.Controller{}
	suite.proCtl = &project.Controller{}
	suite.sysArtifactMgr = &systemartifacttesting.Manager{}
	suite.userMgr = &user.Manager{}
	suite.Config = &restapi.Config{
		ScanDataExportAPI: &scanDataExportAPI{
			scanDataExportCtl: suite.scanExportCtl,
			proCtl:            suite.proCtl,
			sysArtifactMgr:    suite.sysArtifactMgr,
			userMgr:           suite.userMgr,
		},
	}
	mock.OnAnything(suite.proCtl, "Exists").Return(true, nil)
	suite.Suite.SetupSuite()
}

func (suite *ScanExportTestSuite) TestAuthorization() {
	{
		criteria := models.ScanDataExportRequest{
			CVEIds:       "CVE-123",
			Labels:       []int64{100},
			Projects:     []int64{200},
			Repositories: "test-repo",
			Tags:         "{test-tag1,test-tag2}",
		}

		reqs := []struct {
			method  string
			url     string
			body    interface{}
			headers map[string]string
		}{
			{http.MethodPost, "/export/cve", criteria, map[string]string{"X-Scan-Data-Type": v1.MimeTypeGenericVulnerabilityReport}},
			{http.MethodGet, "/export/cve/execution/100", nil, nil},
			{http.MethodGet, "/export/cve/download/100", nil, nil},
		}

		suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(false).Times(3)
		suite.Security.On("IsAuthenticated").Return(false).Times(3)
		for _, req := range reqs {

			if req.body != nil && req.method == http.MethodPost {
				data, _ := json.Marshal(criteria)
				buffer := bytes.NewBuffer(data)
				res, _ := suite.DoReq(req.method, req.url, buffer, req.headers)
				suite.Equal(401, res.StatusCode)
			} else {
				res, _ := suite.DoReq(req.method, req.url, nil)
				suite.Equal(401, res.StatusCode)
			}

		}
	}
}

func (suite *ScanExportTestSuite) TestValidateScanExportParams() {
	api := newScanDataExportAPI()
	api.proCtl = suite.proCtl
	ctx := context.TODO()
	// no scan data type should return error
	err := api.validateScanExportParams(ctx, operation.ExportScanDataParams{})
	suite.Error(err)
	suite.True(errors.IsErr(err, errors.BadRequestCode))

	xScanDataType := v1.MimeTypeGenericVulnerabilityReport
	// empty params should return error
	err = api.validateScanExportParams(ctx, operation.ExportScanDataParams{XScanDataType: xScanDataType})
	suite.Error(err)
	suite.True(errors.IsErr(err, errors.BadRequestCode))

	// multiple projects in input should return error
	criteria := models.ScanDataExportRequest{
		Projects: []int64{200, 300},
	}
	err = api.validateScanExportParams(ctx, operation.ExportScanDataParams{XScanDataType: xScanDataType, Criteria: &criteria})
	suite.Error(err)
	suite.True(errors.IsErr(err, errors.BadRequestCode))

	// spaces in input should return error
	criteria = models.ScanDataExportRequest{
		CVEIds:       "CVE-123, CVE-456",
		Labels:       []int64{100},
		Projects:     []int64{200},
		Repositories: "test-repo1, test-repo2",
		Tags:         "{test-tag1, test-tag2}",
	}
	err = api.validateScanExportParams(ctx, operation.ExportScanDataParams{XScanDataType: xScanDataType, Criteria: &criteria})
	suite.Error(err)
	suite.True(errors.IsErr(err, errors.BadRequestCode))

	// valid params should pass validator
	criteria = models.ScanDataExportRequest{
		CVEIds:       "CVE-123,CVE-456",
		Labels:       []int64{100},
		Projects:     []int64{200},
		Repositories: "test-repo1,test-repo2",
		Tags:         "{test-tag1,test-tag2}",
	}
	err = api.validateScanExportParams(ctx, operation.ExportScanDataParams{XScanDataType: xScanDataType, Criteria: &criteria})
	suite.NoError(err)

	// none exist project should return error
	api.proCtl = &project.Controller{}
	mock.OnAnything(api.proCtl, "Exists").Return(false, nil)
	criteria = models.ScanDataExportRequest{
		CVEIds:       "CVE-123,CVE-456",
		Labels:       []int64{100},
		Projects:     []int64{200},
		Repositories: "test-repo1,test-repo2",
		Tags:         "{test-tag1,test-tag2}",
	}
	err = api.validateScanExportParams(ctx, operation.ExportScanDataParams{XScanDataType: xScanDataType, Criteria: &criteria})
	suite.Error(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))
}

func (suite *ScanExportTestSuite) TestExportScanData() {
	suite.Security.On("GetUsername").Return("test-user")
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
	usr := commonmodels.User{UserID: 1000, Username: "test-user"}
	suite.userMgr.On("GetByName", mock.Anything, "test-user").Return(&usr, nil).Once()
	// user authenticated and correct headers sent
	{
		suite.Security.On("IsAuthenticated").Return(true).Once()
		url := "/export/cve"
		criteria := models.ScanDataExportRequest{
			JobName:      "test-job",
			CVEIds:       "CVE-123",
			Labels:       []int64{100},
			Projects:     []int64{200},
			Repositories: "test-repo",
			Tags:         "{test-tag1,test-tag2}",
		}

		data, err := json.Marshal(criteria)
		buffer := bytes.NewBuffer(data)

		headers := make(map[string]string)
		headers["X-Scan-Data-Type"] = v1.MimeTypeGenericVulnerabilityReport

		// data, err := json.Marshal(criteria)
		mock.OnAnything(suite.scanExportCtl, "Start").Return(int64(100), nil).Once()
		res, err := suite.DoReq(http.MethodPost, url, buffer, headers)
		suite.Equal(200, res.StatusCode)

		suite.Equal(nil, err)
		respData := make(map[string]interface{})
		json.NewDecoder(res.Body).Decode(&respData)
		suite.Equal(int64(100), int64(respData["id"].(float64)))

		// validate job name and user name set in the request for job execution
		jobRequestMatcher := testifymock.MatchedBy(func(req export.Request) bool {
			return req.UserName == "test-user" && req.JobName == "test-job" && req.Tags == "{test-tag1,test-tag2}" && req.UserID == 1000
		})
		suite.scanExportCtl.AssertCalled(suite.T(), "Start", mock.Anything, jobRequestMatcher)
	}

	// user authenticated but incorrect/unsupported header sent across
	{
		suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
		suite.Security.On("IsAuthenticated").Return(true).Once()
		url := "/export/cve"

		criteria := models.ScanDataExportRequest{
			CVEIds:       "CVE-123",
			Labels:       []int64{100},
			Projects:     []int64{200},
			Repositories: "test-repo",
			Tags:         "{test-tag1,test-tag2}",
		}

		data, err := json.Marshal(criteria)
		buffer := bytes.NewBuffer(data)

		headers := make(map[string]string)
		headers["X-Scan-Data-Type"] = "test"

		mock.OnAnything(suite.scanExportCtl, "Start").Return(int64(100), nil).Once()
		res, err := suite.DoReq(http.MethodPost, url, buffer, headers)
		suite.Equal(400, res.StatusCode)
		suite.Equal(nil, err)
	}

	// should return 400 if project id number is not one
	{
		suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
		suite.Security.On("IsAuthenticated").Return(true).Once()
		url := "/export/cve"

		criteria := models.ScanDataExportRequest{
			CVEIds:       "CVE-123",
			Labels:       []int64{100},
			Projects:     []int64{200, 300},
			Repositories: "test-repo",
			Tags:         "{test-tag1,test-tag2}",
		}

		data, err := json.Marshal(criteria)
		buffer := bytes.NewBuffer(data)

		headers := make(map[string]string)
		headers["X-Scan-Data-Type"] = v1.MimeTypeGenericVulnerabilityReport

		mock.OnAnything(suite.scanExportCtl, "Start").Return(int64(100), nil).Once()
		res, err := suite.DoReq(http.MethodPost, url, buffer, headers)
		suite.Equal(400, res.StatusCode)
		suite.Equal(nil, err)
	}

}

func (suite *ScanExportTestSuite) TestExportScanDataGetUserIdError() {
	suite.Security.On("GetUsername").Return("test-user")
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
	suite.userMgr.On("GetByName", mock.Anything, "test-user").Return(nil, errors.New("test error")).Once()
	// user authenticated and correct headers sent
	{
		suite.Security.On("IsAuthenticated").Return(true).Once()
		url := "/export/cve"
		criteria := models.ScanDataExportRequest{
			JobName:      "test-job",
			CVEIds:       "CVE-123",
			Labels:       []int64{100},
			Projects:     []int64{200},
			Repositories: "test-repo",
			Tags:         "{test-tag1,test-tag2}",
		}

		data, err := json.Marshal(criteria)
		buffer := bytes.NewBuffer(data)

		headers := make(map[string]string)
		headers["X-Scan-Data-Type"] = v1.MimeTypeGenericVulnerabilityReport

		// data, err := json.Marshal(criteria)
		mock.OnAnything(suite.scanExportCtl, "Start").Return(int64(100), nil).Once()
		res, err := suite.DoReq(http.MethodPost, url, buffer, headers)
		suite.Equal(http.StatusInternalServerError, res.StatusCode)
		suite.Equal(nil, err)

		suite.scanExportCtl.AssertNotCalled(suite.T(), "Start")
	}
}

func (suite *ScanExportTestSuite) TestExportScanDataGetUserIdNotFound() {
	suite.Security.On("GetUsername").Return("test-user")
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
	suite.userMgr.On("GetByName", mock.Anything, "test-user").Return(nil, nil).Once()
	// user authenticated and correct headers sent
	{
		suite.Security.On("IsAuthenticated").Return(true).Once()
		url := "/export/cve"
		criteria := models.ScanDataExportRequest{
			JobName:      "test-job",
			CVEIds:       "CVE-123",
			Labels:       []int64{100},
			Projects:     []int64{200},
			Repositories: "test-repo",
			Tags:         "{test-tag1,test-tag2}",
		}

		data, err := json.Marshal(criteria)
		buffer := bytes.NewBuffer(data)

		headers := make(map[string]string)
		headers["X-Scan-Data-Type"] = v1.MimeTypeGenericVulnerabilityReport

		// data, err := json.Marshal(criteria)
		mock.OnAnything(suite.scanExportCtl, "Start").Return(int64(100), nil).Once()
		res, err := suite.DoReq(http.MethodPost, url, buffer, headers)
		suite.Equal(http.StatusForbidden, res.StatusCode)
		suite.Equal(nil, err)

		suite.scanExportCtl.AssertNotCalled(suite.T(), "Start")
	}
}

func (suite *ScanExportTestSuite) TestExportScanDataNoPrivileges() {
	suite.Security.On("IsAuthenticated").Return(true).Times(2)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(false).Once()
	url := "/export/cve"

	criteria := models.ScanDataExportRequest{
		JobName:      "test-job",
		CVEIds:       "CVE-123",
		Labels:       []int64{100},
		Projects:     []int64{200},
		Repositories: "test-repo",
		Tags:         "{test-tag1,test-tag2}",
	}

	data, err := json.Marshal(criteria)
	buffer := bytes.NewBuffer(data)

	headers := make(map[string]string)
	headers["X-Scan-Data-Type"] = v1.MimeTypeGenericVulnerabilityReport

	mock.OnAnything(suite.scanExportCtl, "Start").Return(int64(100), nil).Once()
	res, err := suite.DoReq(http.MethodPost, url, buffer, headers)
	suite.Equal(http.StatusForbidden, res.StatusCode)
	suite.NoError(err)
}

func (suite *ScanExportTestSuite) TestGetScanDataExportExecution() {
	suite.Security.On("GetUsername").Return("test-user")
	suite.Security.On("IsAuthenticated").Return(true).Twice()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Twice()
	url := "/export/cve/execution/100"
	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)
	defaultStatusMessage := "Please contact the system administrator to check the logs of jobservice."
	customizeStatusMessage := "No vulnerabilities found or matched"

	execution := &export.Execution{
		ID:               100,
		UserID:           3,
		Status:           "Error",
		StatusMessage:    "",
		Trigger:          "MANUAL",
		StartTime:        startTime,
		EndTime:          endTime,
		ExportDataDigest: "datadigest",
		UserName:         "test-user",
		JobName:          "test-job",
		FilePresent:      false,
	}
	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(execution, nil).Once()
	res, err := suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(200, res.StatusCode)
	suite.Equal(nil, err)
	respData := models.ScanDataExportExecution{}
	json.NewDecoder(res.Body).Decode(&respData)
	suite.Equal("test-user", respData.UserName)
	suite.Equal(false, respData.FilePresent)
	suite.Equal(defaultStatusMessage, respData.StatusText)

	// test customize status message
	execution.StatusMessage = customizeStatusMessage
	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(execution, nil).Once()
	res, err = suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(200, res.StatusCode)
	suite.Equal(nil, err)
	respData = models.ScanDataExportExecution{}
	json.NewDecoder(res.Body).Decode(&respData)
	suite.Equal(customizeStatusMessage, respData.StatusText)
}

func (suite *ScanExportTestSuite) TestGetScanDataExportExecutionUserNotOwnerOfExport() {
	suite.Security.On("GetUsername").Return("test-user")
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
	url := "/export/cve/execution/100"
	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)

	execution := &export.Execution{
		ID:               100,
		UserID:           3,
		Status:           "Success",
		StatusMessage:    "",
		Trigger:          "MANUAL",
		StartTime:        startTime,
		EndTime:          endTime,
		ExportDataDigest: "datadigest",
		UserName:         "test-user1",
		JobName:          "test-job",
		FilePresent:      false,
	}
	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(execution, nil).Once()
	res, err := suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(http.StatusForbidden, res.StatusCode)
	suite.Equal(nil, err)
}

func (suite *ScanExportTestSuite) TestDownloadScanData() {
	suite.Security.On("GetUsername").Return("test-user")
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
	url := "/export/cve/download/100"
	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)

	execution := &export.Execution{
		ID:               int64(100),
		UserID:           int64(3),
		Status:           "Success",
		StatusMessage:    "",
		Trigger:          "MANUAL",
		StartTime:        startTime,
		EndTime:          endTime,
		ExportDataDigest: "datadigest",
		UserName:         "test-user",
		FilePresent:      true,
	}
	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(execution, nil)
	mock.OnAnything(suite.scanExportCtl, "DeleteExecution").Return(nil)

	// all BLOB related operations succeed
	mock.OnAnything(suite.sysArtifactMgr, "Create").Return(int64(1), nil)

	sampleData := "test,hello,world"
	data := io.NopCloser(strings.NewReader(sampleData))
	mock.OnAnything(suite.sysArtifactMgr, "Read").Return(data, nil)
	mock.OnAnything(suite.sysArtifactMgr, "Delete").Return(nil)

	res, err := suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(200, res.StatusCode)
	suite.Equal(nil, err)

	// validate the content of the response
	var responseData bytes.Buffer
	if _, err := io.Copy(&responseData, res.Body); err == nil {
		suite.Equal(sampleData, responseData.String())
	}
}

func (suite *ScanExportTestSuite) TestDownloadScanDataUserNotOwnerofExport() {
	suite.Security.On("GetUsername").Return("test-user1")
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
	url := "/export/cve/download/100"
	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)

	execution := &export.Execution{
		ID:               int64(100),
		UserID:           int64(3),
		Status:           "Success",
		StatusMessage:    "",
		Trigger:          "MANUAL",
		StartTime:        startTime,
		EndTime:          endTime,
		ExportDataDigest: "datadigest",
		UserName:         "test-user",
		FilePresent:      true,
	}
	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(execution, nil)
	mock.OnAnything(suite.scanExportCtl, "DeleteExecution").Return(nil)

	// all BLOB related operations succeed
	mock.OnAnything(suite.sysArtifactMgr, "Create").Return(int64(1), nil)

	sampleData := "test,hello,world"
	data := io.NopCloser(strings.NewReader(sampleData))
	mock.OnAnything(suite.sysArtifactMgr, "Read").Return(data, nil)
	mock.OnAnything(suite.sysArtifactMgr, "Delete").Return(nil)

	res, err := suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(http.StatusForbidden, res.StatusCode)
	suite.Equal(nil, err)
}

func (suite *ScanExportTestSuite) TestDownloadScanDataNoCsvFilePresent() {
	suite.Security.On("GetUsername").Return("test-user1")
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
	url := "/export/cve/download/100"
	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)

	execution := &export.Execution{
		ID:               int64(100),
		UserID:           int64(3),
		Status:           "Success",
		StatusMessage:    "",
		Trigger:          "MANUAL",
		StartTime:        startTime,
		EndTime:          endTime,
		ExportDataDigest: "datadigest",
		UserName:         "test-user1",
		FilePresent:      false,
	}
	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(execution, nil)
	mock.OnAnything(suite.scanExportCtl, "DeleteExecution").Return(nil)

	// all BLOB related operations succeed
	mock.OnAnything(suite.sysArtifactMgr, "Create").Return(int64(1), nil)

	sampleData := "test,hello,world"
	data := io.NopCloser(strings.NewReader(sampleData))
	mock.OnAnything(suite.sysArtifactMgr, "Read").Return(data, nil)
	mock.OnAnything(suite.sysArtifactMgr, "Delete").Return(nil)

	res, err := suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(http.StatusNotFound, res.StatusCode)
	suite.Equal(nil, err)
}

func (suite *ScanExportTestSuite) TestDownloadScanDataExecutionNotPresent() {
	suite.Security.On("GetUsername").Return("test-user1")
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
	url := "/export/cve/download/100"

	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(nil, beegoorm.ErrNoRows)
	mock.OnAnything(suite.scanExportCtl, "DeleteExecution").Return(nil)

	// all BLOB related operations succeed
	mock.OnAnything(suite.sysArtifactMgr, "Create").Return(int64(1), nil)

	sampleData := "test,hello,world"
	data := io.NopCloser(strings.NewReader(sampleData))
	mock.OnAnything(suite.sysArtifactMgr, "Read").Return(data, nil)
	mock.OnAnything(suite.sysArtifactMgr, "Delete").Return(nil)

	res, err := suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(http.StatusNotFound, res.StatusCode)
	suite.Equal(nil, err)
}

func (suite *ScanExportTestSuite) TestDownloadScanDataExecutionError() {
	suite.Security.On("GetUsername").Return("test-user1")
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
	url := "/export/cve/download/100"

	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(nil, errors.New("test error"))
	mock.OnAnything(suite.scanExportCtl, "DeleteExecution").Return(nil)

	// all BLOB related operations succeed
	mock.OnAnything(suite.sysArtifactMgr, "Create").Return(int64(1), nil)

	sampleData := "test,hello,world"
	data := io.NopCloser(strings.NewReader(sampleData))
	mock.OnAnything(suite.sysArtifactMgr, "Read").Return(data, nil)
	mock.OnAnything(suite.sysArtifactMgr, "Delete").Return(nil)

	res, err := suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(http.StatusInternalServerError, res.StatusCode)
	suite.Equal(nil, err)
}

func (suite *ScanExportTestSuite) TestGetScanDataExportExecutionList() {
	suite.Security.On("GetUsername").Return("test-user")
	suite.Security.On("IsAuthenticated").Return(true).Twice()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Twice()
	url, err := url2.Parse("/export/cve/executions")
	params := url2.Values{}
	params.Add("user_name", "test-user")
	url.RawQuery = params.Encode()
	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)
	defaultStatusMessage := "Please contact the system administrator to check the logs of jobservice."
	customizeStatusMessage := "No vulnerabilities found or matched"

	execution := &export.Execution{
		ID:               100,
		UserID:           3,
		Status:           "Error",
		StatusMessage:    "",
		Trigger:          "MANUAL",
		StartTime:        startTime,
		EndTime:          endTime,
		ExportDataDigest: "datadigest",
		JobName:          "test-job",
		UserName:         "test-user",
	}
	fmt.Println("URL string : ", url.String())
	mock.OnAnything(suite.scanExportCtl, "ListExecutions").Return([]*export.Execution{execution}, nil).Once()
	res, err := suite.DoReq(http.MethodGet, url.String(), nil)
	suite.Equal(200, res.StatusCode)
	suite.Equal(nil, err)
	respData := models.ScanDataExportExecutionList{}
	json.NewDecoder(res.Body).Decode(&respData)
	suite.Equal(1, len(respData.Items))
	suite.Equal(int64(100), respData.Items[0].ID)
	suite.Equal(defaultStatusMessage, respData.Items[0].StatusText)
	// test customize status message
	execution.StatusMessage = customizeStatusMessage
	mock.OnAnything(suite.scanExportCtl, "ListExecutions").Return([]*export.Execution{execution}, nil).Once()
	res, err = suite.DoReq(http.MethodGet, url.String(), nil)
	suite.Equal(200, res.StatusCode)
	suite.Equal(nil, err)
	respData = models.ScanDataExportExecutionList{}
	json.NewDecoder(res.Body).Decode(&respData)
	suite.Equal(1, len(respData.Items))
	suite.Equal(int64(100), respData.Items[0].ID)
	suite.Equal(customizeStatusMessage, respData.Items[0].StatusText)
}

func (suite *ScanExportTestSuite) TestGetScanDataExportExecutionListFilterNotOwned() {
	suite.Security.On("GetUsername").Return("test-user")
	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
	url, err := url2.Parse("/export/cve/executions")
	params := url2.Values{}
	params.Add("user_name", "test-user")
	url.RawQuery = params.Encode()
	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)

	executionOwned := &export.Execution{
		ID:               100,
		UserID:           3,
		Status:           "Success",
		StatusMessage:    "",
		Trigger:          "MANUAL",
		StartTime:        startTime,
		EndTime:          endTime,
		ExportDataDigest: "datadigest",
		JobName:          "test-job",
		UserName:         "test-user",
	}

	fmt.Println("URL string : ", url.String())
	mock.OnAnything(suite.scanExportCtl, "ListExecutions").Return([]*export.Execution{executionOwned}, nil).Once()
	res, err := suite.DoReq(http.MethodGet, url.String(), nil)
	suite.Equal(200, res.StatusCode)
	suite.Equal(nil, err)
	respData := models.ScanDataExportExecutionList{}
	json.NewDecoder(res.Body).Decode(&respData)
	suite.Equal(1, len(respData.Items))
	suite.Equal(int64(100), respData.Items[0].ID)
}

func TestScanExportTestSuite(t *testing.T) {
	suite.Run(t, &ScanExportTestSuite{})
}
