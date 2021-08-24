package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/controller/scandataexport"
	"github.com/goharbor/harbor/src/pkg/registry"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	exporttesting "github.com/goharbor/harbor/src/testing/controller/scan/export"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type ScanExportTestSuite struct {
	htesting.Suite
	scanExportCtl *exporttesting.Controller
	regCli        registry.Client
}

func (suite *ScanExportTestSuite) SetupSuite() {
	suite.scanExportCtl = &exporttesting.Controller{}
	suite.regCli = &exporttesting.RegistryClient{}

	suite.Config = &restapi.Config{
		ScanDataExportAPI: &scanDataExportAPI{
			scanDataExportCtl: suite.scanExportCtl,
			regCli:            suite.regCli,
		},
	}

	suite.Suite.SetupSuite()
}

func (suite *ScanExportTestSuite) TestAuthorization() {
	{
		criteria := models.ScanDataExportCriteria{
			CVEIds:       []string{"CVE-123"},
			Labels:       []int64{100},
			Projects:     []int64{200},
			Repositories: []int64{300},
			Tags:         []int64{400, 500}}

		reqs := []struct {
			method  string
			url     string
			body    interface{}
			headers map[string]string
		}{
			{http.MethodPost, "/export/scan", criteria, map[string]string{"X-Scan-Data-Type": v1.MimeTypeGenericVulnerabilityReport}},
			{http.MethodGet, "/export/scan/execution/100", nil, nil},
			{http.MethodGet, "/export/scan/download/100", nil, nil},
		}

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
func (suite *ScanExportTestSuite) TestExportScanData() {
	// user authenticated and correct headers sent
	{
		suite.Security.On("IsAuthenticated").Return(true).Once()
		suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
		url := "/export/scan"

		criteria := models.ScanDataExportCriteria{
			CVEIds:       []string{"CVE-123"},
			Labels:       []int64{100},
			Projects:     []int64{200},
			Repositories: []int64{300},
			Tags:         []int64{400, 500}}

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
	}

	// user authenticated but incorrect/unsupported header sent across
	{
		suite.Security.On("IsAuthenticated").Return(true).Once()
		suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
		url := "/export/scan"

		criteria := models.ScanDataExportCriteria{
			CVEIds:       []string{"CVE-123"},
			Labels:       []int64{100},
			Projects:     []int64{200},
			Repositories: []int64{300},
			Tags:         []int64{400, 500}}

		data, err := json.Marshal(criteria)
		buffer := bytes.NewBuffer(data)

		headers := make(map[string]string)
		headers["X-Scan-Data-Type"] = "test"

		mock.OnAnything(suite.scanExportCtl, "Start").Return(int64(100), nil).Once()
		res, err := suite.DoReq(http.MethodPost, url, buffer, headers)
		suite.Equal(400, res.StatusCode)
		suite.Equal(nil, err)

	}
}

func (suite *ScanExportTestSuite) TestGetScanDataExportExecution() {

	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
	url := "/export/scan/execution/100"
	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)

	execution := &scandataexport.Execution{
		ID:               100,
		UserID:           3,
		Status:           "Success",
		StatusMessage:    "",
		Trigger:          "MANUAL",
		StartTime:        startTime,
		EndTime:          endTime,
		ExportDataDigest: "datadigest",
	}
	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(execution, nil).Once()
	res, err := suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(200, res.StatusCode)
	suite.Equal(nil, err)
	respData := scandataexport.Execution{}
	json.NewDecoder(res.Body).Decode(&respData)
	fmt.Println("Done")

}

func (suite *ScanExportTestSuite) TestDownloadScanData() {

	suite.Security.On("IsAuthenticated").Return(true).Once()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(1)
	url := "/export/scan/download/100"
	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)

	execution := &scandataexport.Execution{
		ID:               100,
		UserID:           3,
		Status:           "Success",
		StatusMessage:    "",
		Trigger:          "MANUAL",
		StartTime:        startTime,
		EndTime:          endTime,
		ExportDataDigest: "datadigest",
	}
	mock.OnAnything(suite.scanExportCtl, "GetExecution").Return(execution, nil).Once()
	// all BLOB related operations succeed
	mock.OnAnything(suite.regCli, "PushBlob").Return(nil).Once()
	mock.OnAnything(suite.regCli, "DeleteBlob").Return(nil).Once()

	sampleData := "test,hello,world"
	data := io.NopCloser(strings.NewReader(sampleData))
	mock.OnAnything(suite.regCli, "PullBlob").Return(int64(16), data, nil)

	res, err := suite.DoReq(http.MethodGet, url, nil)
	suite.Equal(200, res.StatusCode)
	suite.Equal(nil, err)

	// validate the content of the response
	var responseData bytes.Buffer
	if _, err := io.Copy(&responseData, res.Body); err == nil {
		suite.Equal(sampleData, responseData.String())
	}

	// assert that the delete method has been called on the repository
	client := suite.regCli.(*exporttesting.RegistryClient)
	client.AssertCalled(suite.T(), "DeleteBlob", "scandata_export_100", "datadigest")
}

func TestScanExportTestSuite(t *testing.T) {
	suite.Run(t, &ScanExportTestSuite{})
}
