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

package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ClientTestSuite tests the v1 client
type ClientTestSuite struct {
	suite.Suite

	testServer *httptest.Server
	client     Client
}

// TestClient is the entry of ClientTestSuite
func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

// SetupSuite prepares the test suite env
func (suite *ClientTestSuite) SetupSuite() {
	suite.testServer = httptest.NewServer(&mockHandler{})
	r := &scanner.Registration{
		ID:             1000,
		UUID:           "uuid",
		Name:           "TestClient",
		URL:            suite.testServer.URL,
		SkipCertVerify: true,
	}

	c, err := NewClient(r)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), c)

	suite.client = c
}

// TestClientMetadata tests the metadata of the client
func (suite *ClientTestSuite) TestClientMetadata() {
	m, err := suite.client.GetMetadata()
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), m)

	assert.Equal(suite.T(), m.Scanner.Name, "Clair")
}

// TestClientSubmitScan tests the scan submission of client
func (suite *ClientTestSuite) TestClientSubmitScan() {
	res, err := suite.client.SubmitScan(&ScanRequest{})
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), res)

	assert.Equal(suite.T(), res.ID, "123456789")
}

// TestClientGetScanReportError tests getting report failed
func (suite *ClientTestSuite) TestClientGetScanReportError() {
	_, err := suite.client.GetScanReport("id1", MimeTypeNativeReport)
	require.Error(suite.T(), err)
	assert.Condition(suite.T(), func() (success bool) {
		success = strings.Index(err.Error(), "error") != -1
		return
	})
}

// TestClientGetScanReport tests getting report
func (suite *ClientTestSuite) TestClientGetScanReport() {
	res, err := suite.client.GetScanReport("id2", MimeTypeNativeReport)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), res)
}

// TestClientGetScanReportNotReady tests the case that the report is not ready
func (suite *ClientTestSuite) TestClientGetScanReportNotReady() {
	_, err := suite.client.GetScanReport("id3", MimeTypeNativeReport)
	require.Error(suite.T(), err)
	require.Condition(suite.T(), func() (success bool) {
		_, success = err.(*ReportNotReadyError)
		return
	})
	assert.Equal(suite.T(), 10, err.(*ReportNotReadyError).RetryAfter)
}

// TearDownSuite clears the test suite env
func (suite *ClientTestSuite) TearDownSuite() {
	suite.testServer.Close()
}

type mockHandler struct{}

// ServeHTTP ...
func (mh *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.RequestURI {
	case "/metadata":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		m := &ScannerAdapterMetadata{
			Scanner: &Scanner{
				Name:    "Clair",
				Vendor:  "Harbor",
				Version: "0.1.0",
			},
			Capabilities: &ScannerCapability{
				ConsumesMimeTypes: []string{
					MimeTypeOCIArtifact,
					MimeTypeDockerArtifact,
				},
				ProducesMimeTypes: []string{
					MimeTypeNativeReport,
					MimeTypeRawReport,
				},
			},
			Properties: ScannerProperties{
				"extra": "testing",
			},
		}
		data, _ := json.Marshal(m)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
		break
	case "/scan":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		res := &ScanResponse{}
		res.ID = "123456789"

		data, _ := json.Marshal(res)

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(data)
		break
	case "/scan/id1/report":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		e := &ErrorResponse{
			&Error{
				Message: "error",
			},
		}

		data, _ := json.Marshal(e)

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(data)
		break
	case "/scan/id2/report":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
		break
	case "/scan/id3/report":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Add(refreshAfterHeader, fmt.Sprintf("%d", 10))
		w.Header().Add("Location", "/scan/id3/report")
		w.WriteHeader(http.StatusFound)
		break
	}
}
