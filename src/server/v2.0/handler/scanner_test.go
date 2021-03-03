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

package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	scannertesting "github.com/goharbor/harbor/src/testing/controller/scanner"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/suite"
)

type ScannerTestSuite struct {
	htesting.Suite

	scannerCtl *scannertesting.Controller
	reg        *scanner.Registration

	metadata v1.ScannerAdapterMetadata
}

func (suite *ScannerTestSuite) SetupSuite() {
	suite.reg = &scanner.Registration{
		Name: "reg",
		URL:  "http://reg:8080",
		UUID: "uuid",
	}

	suite.metadata = v1.ScannerAdapterMetadata{
		Scanner: &v1.Scanner{
			Name: "reg",
		},
	}

	suite.scannerCtl = &scannertesting.Controller{}

	suite.Config = &restapi.Config{
		ScannerAPI: &scannerAPI{
			scannerCtl: suite.scannerCtl,
		},
	}

	suite.Suite.SetupSuite()
}

func (suite *ScannerTestSuite) TestAuthorization() {
	newBody := func(body interface{}) io.Reader {
		if body == nil {
			return nil
		}

		buf, err := json.Marshal(body)
		suite.Require().NoError(err)
		return bytes.NewBuffer(buf)
	}

	reqs := []struct {
		method string
		url    string
		body   interface{}
	}{
		{http.MethodGet, "/scanners", nil},
		{http.MethodPost, "/scanners", suite.reg},
		{http.MethodPost, "/scanners/ping", suite.reg},
		{http.MethodGet, "/scanners/uuid1", nil},
		{http.MethodPut, "/scanners/uuid1", suite.reg},
		{http.MethodDelete, "/scanners/uuid1", nil},
		{http.MethodPatch, "/scanners/uuid1", map[string]interface{}{"is_default": true}},
		{http.MethodGet, "/scanners/uuid1/metadata", nil},
	}

	for _, req := range reqs {
		{
			// authorized required
			suite.Security.On("IsAuthenticated").Return(false).Once()

			res, err := suite.DoReq(req.method, req.url, newBody(req.body))
			suite.NoError(err)
			suite.Equal(401, res.StatusCode)
		}

		{
			// permission required
			suite.Security.On("IsAuthenticated").Return(true).Once()
			suite.Security.On("GetUsername").Return("username").Once()
			suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(false).Once()

			res, err := suite.DoReq(req.method, req.url, newBody(req.body))
			suite.NoError(err)
			suite.Equal(403, res.StatusCode)
		}
	}
}

func (suite *ScannerTestSuite) TestCreateScannerWithInvalidBody() {
	{
		// empty body
		res, err := suite.PostJSON("/scanners", nil)
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// name missing
		res, err := suite.PostJSON("/scanners", map[string]interface{}{
			"url": "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// url missing
		res, err := suite.PostJSON("/scanners", map[string]interface{}{
			"name": "reg",
		})
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// invalid url
		res, err := suite.PostJSON("/scanners", map[string]interface{}{
			"name": "reg",
			"url":  "invalid url",
		})
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}
}

func (suite *ScannerTestSuite) TestCreateScanner() {
	times := 4
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		mock.OnAnything(suite.scannerCtl, "CreateRegistration").Return("", fmt.Errorf("failed to create registration")).Once()
		res, err := suite.PostJSON("/scanners", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		mock.OnAnything(suite.scannerCtl, "CreateRegistration").Return("uuid", nil).Once()
		res, err := suite.PostJSON("/scanners", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(201, res.StatusCode)
		suite.Equal("/api/v2.0/scanners/uuid", res.Header.Get("Location"))
	}

	{
		// access_credential missing
		res, err := suite.PostJSON("/scanners", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
			"auth": "Basic",
		})
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}

	{
		mock.OnAnything(suite.scannerCtl, "CreateRegistration").Return("uuid", nil).Once()
		res, err := suite.PostJSON("/scanners", map[string]interface{}{
			"name":              "reg",
			"url":               "http://reg:8080",
			"auth":              "Basic",
			"access_credential": "username:password",
		})
		suite.NoError(err)
		suite.Equal(201, res.StatusCode)
		suite.Equal("/api/v2.0/scanners/uuid", res.Header.Get("Location"))
	}
}

func (suite *ScannerTestSuite) TestDeleteScanner() {
	times := 5
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// get scanner failed
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(nil, fmt.Errorf("failed to get registration")).Once()
		res, err := suite.Delete("/scanners/uuid")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// scanner not found
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(nil, nil).Once()
		res, err := suite.Delete("/scanners/uuid")
		suite.NoError(err)
		suite.Equal(404, res.StatusCode)
	}

	{
		// immutable scanner
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(&scanner.Registration{Immutable: true}, nil).Once()
		res, err := suite.Delete("/scanners/uuid")
		suite.NoError(err)
		suite.Equal(403, res.StatusCode)
	}

	{
		// delete scanner failed
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.scannerCtl, "DeleteRegistration").Return(nil, fmt.Errorf("failed to delete registration")).Once()
		res, err := suite.Delete("/scanners/uuid")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// delete scanner
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.scannerCtl, "DeleteRegistration").Return(suite.reg, nil).Once()
		res, err := suite.Delete("/scanners/uuid")
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func (suite *ScannerTestSuite) TestGetScanner() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// get scanner failed
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(nil, fmt.Errorf("failed to get registration")).Once()

		res, err := suite.Get("/scanners/uuid")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// scanner not found
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(nil, nil).Once()

		var scanner map[string]interface{}
		res, err := suite.GetJSON("/scanners/uuid", &scanner)
		suite.NoError(err)
		suite.Equal(404, res.StatusCode)
	}

	{
		// scanner found
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(suite.reg, nil).Once()

		var scanner map[string]interface{}
		res, err := suite.GetJSON("/scanners/uuid", &scanner)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Equal("uuid", scanner["uuid"])
	}
}

func (suite *ScannerTestSuite) TestGetScannerMetadata() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// get metadata failed
		mock.OnAnything(suite.scannerCtl, "GetMetadata").Return(nil, fmt.Errorf("failed to get metadata")).Once()

		res, err := suite.Get("/scanners/uuid/metadata")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		mock.OnAnything(suite.scannerCtl, "GetMetadata").Return(&suite.metadata, nil).Once()

		var md v1.ScannerAdapterMetadata
		res, err := suite.GetJSON("/scanners/uuid/metadata", &md)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Equal(suite.metadata, md)
	}
}

func (suite *ScannerTestSuite) TestListScanners() {
	times := 4
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// list scanners failed
		mock.OnAnything(suite.scannerCtl, "GetTotalOfRegistrations").Return(int64(0), fmt.Errorf("failed to count scanners")).Once()

		res, err := suite.Get("/scanners")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// list scanners failed
		mock.OnAnything(suite.scannerCtl, "GetTotalOfRegistrations").Return(int64(1), nil).Once()
		mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return(nil, fmt.Errorf("failed to list scanners")).Once()

		res, err := suite.Get("/scanners")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// scanners not found
		mock.OnAnything(suite.scannerCtl, "GetTotalOfRegistrations").Return(int64(0), nil).Once()
		mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return(nil, nil).Once()

		var scanners []interface{}
		res, err := suite.GetJSON("/scanners", &scanners)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(scanners, 0)
	}

	{
		// scanners found
		mock.OnAnything(suite.scannerCtl, "GetTotalOfRegistrations").Return(int64(3), nil).Once()
		mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return([]*scanner.Registration{suite.reg}, nil).Once()

		var scanners []interface{}
		res, err := suite.GetJSON("/scanners?page_size=1&page=2", &scanners)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(scanners, 1)
		suite.Equal("3", res.Header.Get("X-Total-Count"))
		suite.Contains(res.Header, "Link")
		suite.Equal(`</api/v2.0/scanners?page=1&page_size=1>; rel="prev" , </api/v2.0/scanners?page=3&page_size=1>; rel="next"`, res.Header.Get("Link"))
	}
}

func (suite *ScannerTestSuite) TestPingScanner() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// bad req
		res, err := suite.PostJSON("/scanners/ping", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
			"auth": "Basic",
		})
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}

	{
		// ping failed
		mock.OnAnything(suite.scannerCtl, "Ping").Return(nil, fmt.Errorf("failed to ping scanner")).Once()

		res, err := suite.PostJSON("/scanners/ping", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// ping
		mock.OnAnything(suite.scannerCtl, "Ping").Return(&suite.metadata, nil).Once()

		res, err := suite.PostJSON("/scanners/ping", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func (suite *ScannerTestSuite) TestSetScannerAsDefault() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		res, err := suite.PatchJSON("/scanners/uuid", map[string]interface{}{
			"is_default": false,
		})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		// set default failed
		mock.OnAnything(suite.scannerCtl, "SetDefaultRegistration").Return(fmt.Errorf("failed to set default")).Once()

		res, err := suite.PatchJSON("/scanners/uuid", map[string]interface{}{
			"is_default": true,
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// set default
		mock.OnAnything(suite.scannerCtl, "SetDefaultRegistration").Return(nil).Once()

		res, err := suite.PatchJSON("/scanners/uuid", map[string]interface{}{
			"is_default": true,
		})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func (suite *ScannerTestSuite) TestUpdateScanner() {
	times := 7
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)

	{
		// update scanner no body
		res, err := suite.Put("/scanners/uuid", nil)
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// get scanner failed
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(nil, fmt.Errorf("failed to get registration")).Once()

		res, err := suite.PutJSON("/scanners/uuid", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// scanner not found
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(nil, nil).Once()

		res, err := suite.PutJSON("/scanners/uuid", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(404, res.StatusCode)
	}

	{
		// immutable scanner
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(&scanner.Registration{Immutable: true}, nil).Once()

		res, err := suite.PutJSON("/scanners/uuid", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(403, res.StatusCode)
	}

	{
		// bad req
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(suite.reg, nil).Once()

		res, err := suite.PutJSON("/scanners/uuid", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
			"auth": "Basic",
		})
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}

	{
		// update scanner failed
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.scannerCtl, "UpdateRegistration").Return(fmt.Errorf("failed to update the scanner")).Once()

		res, err := suite.PutJSON("/scanners/uuid", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// update scanner
		mock.OnAnything(suite.scannerCtl, "GetRegistration").Return(suite.reg, nil).Once()
		mock.OnAnything(suite.scannerCtl, "UpdateRegistration").Return(nil).Once()

		res, err := suite.PutJSON("/scanners/uuid", map[string]interface{}{
			"name": "reg",
			"url":  "http://reg:8080",
		})
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}
}

func TestScannerTestSuite(t *testing.T) {
	suite.Run(t, &ScannerTestSuite{})
}
