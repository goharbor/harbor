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

package api

import (
	"fmt"
	"net/http"
	"testing"

	sc "github.com/goharbor/harbor/src/api/scanner"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	scannertesting "github.com/goharbor/harbor/src/testing/api/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ProScannerAPITestSuite is test suite for testing the project scanner API
type ProScannerAPITestSuite struct {
	suite.Suite

	originC sc.Controller
	mockC   *scannertesting.Controller
}

// TestProScannerAPI is the entry of ProScannerAPITestSuite
func TestProScannerAPI(t *testing.T) {
	suite.Run(t, new(ProScannerAPITestSuite))
}

// SetupSuite prepares testing env
func (suite *ProScannerAPITestSuite) SetupTest() {
	suite.originC = sc.DefaultController
	m := &scannertesting.Controller{}
	sc.DefaultController = m

	suite.mockC = m
}

// TearDownTest clears test case env
func (suite *ProScannerAPITestSuite) TearDownTest() {
	// Restore
	sc.DefaultController = suite.originC
}

// TestScannerAPIProjectScanner tests the API of getting/setting project level scanner
func (suite *ProScannerAPITestSuite) TestScannerAPIProjectScanner() {
	suite.mockC.On("SetRegistrationByProject", int64(1), "uuid").Return(nil)

	// Set
	body := make(map[string]interface{}, 1)
	body["uuid"] = "uuid"
	runCodeCheckingCases(suite.T(), &codeCheckingCase{
		request: &testingRequest{
			url:        fmt.Sprintf("/api/projects/%d/scanner", 1),
			method:     http.MethodPut,
			credential: projAdmin,
			bodyJSON:   body,
		},
		code: http.StatusOK,
	})

	r := &scanner.Registration{
		ID:          1004,
		UUID:        "uuid",
		Name:        "TestScannerAPIProjectScanner",
		Description: "JUST FOR TEST",
		URL:         "https://a.b.c",
	}
	suite.mockC.On("GetRegistrationByProject", int64(1)).Return(r, nil)

	// Get
	rr := &scanner.Registration{}
	err := handleAndParse(&testingRequest{
		url:        fmt.Sprintf("/api/projects/%d/scanner", 1),
		method:     http.MethodGet,
		credential: projAdmin,
	}, rr)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), r.Name, rr.Name)
	assert.Equal(suite.T(), r.UUID, rr.UUID)
}

// TestScannerAPIGetScannerCandidates ...
func (suite *ProScannerAPITestSuite) TestScannerAPIGetScannerCandidates() {
	query := &q.Query{
		PageNumber: 1,
		PageSize:   500,
	}

	ll := []*scanner.Registration{
		{
			ID:          1005,
			UUID:        "uuid",
			Name:        "TestScannerAPIGetScannerCandidates",
			Description: "JUST FOR TEST",
			URL:         "https://a.b.c",
		}}
	suite.mockC.On("ListRegistrations", query).Return(ll, nil)

	// Get
	l := make([]*scanner.Registration, 0)
	err := handleAndParse(&testingRequest{
		url:        fmt.Sprintf("/api/projects/%d/scanner/candidates", 1),
		method:     http.MethodGet,
		credential: projAdmin,
	}, &l)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), 1, len(l))
	assert.Equal(suite.T(), "uuid", l[0].UUID)
}
