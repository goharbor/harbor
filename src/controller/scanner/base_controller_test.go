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

package scanner

import (
	"context"
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	mocktesting "github.com/goharbor/harbor/src/testing/mock"
	metadatatesting "github.com/goharbor/harbor/src/testing/pkg/project/metadata"
	v1testing "github.com/goharbor/harbor/src/testing/pkg/scan/rest/v1"
	scannertesting "github.com/goharbor/harbor/src/testing/pkg/scan/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ControllerTestSuite is test suite to test the basic api controller.
type ControllerTestSuite struct {
	suite.Suite

	c     *basicController
	mMgr  *scannertesting.Manager
	mMeta *metadatatesting.Manager

	sample *scanner.Registration
}

// TestController is the entry of controller test suite
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

// SetupTest prepares env for the controller test suite
func (suite *ControllerTestSuite) SetupTest() {
	suite.mMgr = &scannertesting.Manager{}
	suite.mMeta = &metadatatesting.Manager{}

	m := &v1.ScannerAdapterMetadata{
		Scanner: &v1.Scanner{
			Name:    "Trivy",
			Vendor:  "Harbor",
			Version: "0.1.0",
		},
		Capabilities: []*v1.ScannerCapability{{
			ConsumesMimeTypes: []string{
				v1.MimeTypeOCIArtifact,
				v1.MimeTypeDockerArtifact,
			},
			ProducesMimeTypes: []string{
				v1.MimeTypeNativeReport,
				v1.MimeTypeRawReport,
				v1.MimeTypeGenericVulnerabilityReport,
			},
		}},
		Properties: v1.ScannerProperties{
			"extra": "testing",
		},
	}

	suite.sample = &scanner.Registration{
		Name:        "forUT",
		Description: "sample registration",
		URL:         "https://sample.scanner.com",
	}

	mc := &v1testing.Client{}
	mc.On("GetMetadata").Return(m, nil)

	mcp := &v1testing.ClientPool{}
	mocktesting.OnAnything(mcp, "Get").Return(mc, nil)
	suite.c = &basicController{
		manager:    suite.mMgr,
		proMetaMgr: suite.mMeta,
		clientPool: mcp,
	}
}

// Clear test case
func (suite *ControllerTestSuite) TearDownTest() {
	suite.sample.UUID = ""
}

// TestListRegistrations tests ListRegistrations
func (suite *ControllerTestSuite) TestListRegistrations() {
	query := &q.Query{
		PageSize:   10,
		PageNumber: 1,
	}

	suite.sample.UUID = "uuid"
	l := []*scanner.Registration{suite.sample}

	suite.mMgr.On("List", mock.Anything, query).Return(l, nil)

	rl, err := suite.c.ListRegistrations(context.TODO(), query)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(rl))
}

// TestCreateRegistration tests CreateRegistration
func (suite *ControllerTestSuite) TestCreateRegistration() {
	suite.mMgr.On("Create", mock.Anything, suite.sample).Return("uuid", nil)

	uid, err := suite.mMgr.Create(context.TODO(), suite.sample)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), uid, "uuid")
}

// TestGetRegistration tests GetRegistration
func (suite *ControllerTestSuite) TestGetRegistration() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Get", mock.Anything, "uuid").Return(suite.sample, nil)

	rr, err := suite.c.GetRegistration(context.TODO(), "uuid")
	require.NoError(suite.T(), err)
	assert.NotNil(suite.T(), rr)
	assert.Equal(suite.T(), "forUT", rr.Name)
}

// TestRegistrationExists tests RegistrationExists
func (suite *ControllerTestSuite) TestRegistrationExists() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Get", mock.Anything, "uuid").Return(suite.sample, nil)

	exists := suite.c.RegistrationExists(context.TODO(), "uuid")
	assert.Equal(suite.T(), true, exists)

	suite.mMgr.On("Get", mock.Anything, "uuid2").Return(nil, nil)

	exists = suite.c.RegistrationExists(context.TODO(), "uuid2")
	assert.Equal(suite.T(), false, exists)
}

// TestUpdateRegistration tests UpdateRegistration
func (suite *ControllerTestSuite) TestUpdateRegistration() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Update", mock.Anything, suite.sample).Return(nil)

	err := suite.c.UpdateRegistration(context.TODO(), suite.sample)
	require.NoError(suite.T(), err)
}

// TestDeleteRegistration tests DeleteRegistration
func (suite *ControllerTestSuite) TestDeleteRegistration() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Get", mock.Anything, "uuid").Return(suite.sample, nil)
	suite.mMgr.On("Delete", mock.Anything, "uuid").Return(nil)

	r, err := suite.c.DeleteRegistration(context.TODO(), "uuid")
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "forUT", r.Name)
}

// TestSetDefaultRegistration tests SetDefaultRegistration
func (suite *ControllerTestSuite) TestSetDefaultRegistration() {
	suite.mMgr.On("SetAsDefault", mock.Anything, "uuid").Return(nil)

	err := suite.c.SetDefaultRegistration(context.TODO(), "uuid")
	require.NoError(suite.T(), err)
}

// TestSetRegistrationByProject tests SetRegistrationByProject
func (suite *ControllerTestSuite) TestSetRegistrationByProject() {
	m := make(map[string]string, 1)
	mm := make(map[string]string, 1)
	mmm := make(map[string]string, 1)
	mm[proScannerMetaKey] = "uuid"
	mmm[proScannerMetaKey] = "uuid2"

	var pid, pid2 int64 = 1, 2

	// not set before
	suite.mMeta.On("Get", mock.Anything, pid, proScannerMetaKey).Return(m, nil)
	suite.mMeta.On("Add", mock.Anything, pid, mm).Return(nil)

	err := suite.c.SetRegistrationByProject(context.TODO(), pid, "uuid")
	require.NoError(suite.T(), err)

	// Set before
	suite.mMeta.On("Get", mock.Anything, pid2, proScannerMetaKey).Return(mm, nil)
	suite.mMeta.On("Update", mock.Anything, pid2, mmm).Return(nil)

	err = suite.c.SetRegistrationByProject(context.TODO(), pid2, "uuid2")
	require.NoError(suite.T(), err)
}

// TestGetRegistrationByProject tests GetRegistrationByProject
func (suite *ControllerTestSuite) TestGetRegistrationByProject() {
	m := make(map[string]string, 1)
	m[proScannerMetaKey] = "uuid"

	// Configured at project level
	var pid int64 = 1
	suite.sample.UUID = "uuid"

	suite.mMeta.On("Get", mock.Anything, pid, proScannerMetaKey).Return(m, nil)
	suite.mMgr.On("Get", mock.Anything, "uuid").Return(suite.sample, nil)

	r, err := suite.c.GetRegistrationByProject(context.TODO(), pid)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "forUT", r.Name)

	// Not configured at project level, return system default
	suite.mMeta.On("Get", mock.Anything, pid, proScannerMetaKey).Return(nil, nil)
	suite.mMgr.On("GetDefault", mock.Anything).Return(suite.sample, nil)

	r, err = suite.c.GetRegistrationByProject(context.TODO(), pid)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), r)
	assert.Equal(suite.T(), "forUT", r.Name)
}

// TestGetRegistrationByProjectWhenPingError tests GetRegistrationByProject
func (suite *ControllerTestSuite) TestGetRegistrationByProjectWhenPingError() {
	m := make(map[string]string, 1)
	m[proScannerMetaKey] = "uuid"

	// Configured at project level
	var pid int64 = 1
	suite.sample.UUID = "uuid"

	suite.mMeta.On("Get", mock.Anything, pid, proScannerMetaKey).Return(m, nil)
	suite.mMgr.On("Get", mock.Anything, "uuid").Return(suite.sample, nil)

	// Ping error
	mc := &v1testing.Client{}
	mc.On("GetMetadata").Return(nil, fmt.Errorf("getMetadata error"))

	mcp := &v1testing.ClientPool{}
	mocktesting.OnAnything(mcp, "Get").Return(mc, nil)
	suite.c.clientPool = mcp

	r, err := suite.c.GetRegistrationByProject(context.TODO(), pid)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "unhealthy", r.Health)
}

// TestPing ...
func (suite *ControllerTestSuite) TestPing() {
	meta, err := suite.c.Ping(context.TODO(), suite.sample)
	require.NoError(suite.T(), err)
	suite.NotNil(meta)
}

// TestPingWithGenericMimeType tests ping for scanners supporting MIME type MimeTypeGenericVulnerabilityReport
func (suite *ControllerTestSuite) TestPingWithGenericMimeType() {
	m := &v1.ScannerAdapterMetadata{
		Scanner: &v1.Scanner{
			Name:    "Trivy",
			Vendor:  "Harbor",
			Version: "0.1.0",
		},
		Capabilities: []*v1.ScannerCapability{{
			ConsumesMimeTypes: []string{
				v1.MimeTypeOCIArtifact,
				v1.MimeTypeDockerArtifact,
			},
			ProducesMimeTypes: []string{
				v1.MimeTypeGenericVulnerabilityReport,
				v1.MimeTypeRawReport,
			},
		}},
		Properties: v1.ScannerProperties{
			"extra": "testing",
		},
	}
	mc := &v1testing.Client{}
	mc.On("GetMetadata").Return(m, nil)

	mcp := &v1testing.ClientPool{}
	mocktesting.OnAnything(mcp, "Get").Return(mc, nil)
	suite.c = &basicController{
		manager:    suite.mMgr,
		proMetaMgr: suite.mMeta,
		clientPool: mcp,
	}
	meta, err := suite.c.Ping(context.TODO(), suite.sample)
	require.NoError(suite.T(), err)
	suite.NotNil(meta)
}

// TestGetMetadata ...
func (suite *ControllerTestSuite) TestGetMetadata() {
	suite.sample.UUID = "uuid"
	suite.mMgr.On("Get", mock.Anything, "uuid").Return(suite.sample, nil)

	meta, err := suite.c.GetMetadata(context.TODO(), suite.sample.UUID)
	require.NoError(suite.T(), err)
	suite.NotNil(meta)
	suite.Equal(1, len(meta.Capabilities))
}
