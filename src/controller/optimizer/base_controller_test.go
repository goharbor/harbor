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

package optimizer

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	ar "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	art "github.com/goharbor/harbor/src/pkg/artifact"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	dockerfileoptdao "github.com/goharbor/harbor/src/pkg/dockerfileoptimization/dao"
	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
	v1 "github.com/goharbor/harbor/src/pkg/optimizer/rest/v1"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	scandao "github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	robottesting "github.com/goharbor/harbor/src/testing/controller/robot"
	scantesting "github.com/goharbor/harbor/src/testing/controller/scan"
	"github.com/goharbor/harbor/src/testing/mock"
	optdaotesting "github.com/goharbor/harbor/src/testing/pkg/dockerfileoptimization/dao"
	optimizertesting "github.com/goharbor/harbor/src/testing/pkg/optimizer"
	v1testing "github.com/goharbor/harbor/src/testing/pkg/optimizer/rest/v1"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
)

// ControllerTestSuite is the test suite for the optimizer basicController.
type ControllerTestSuite struct {
	suite.Suite

	mMgr     *optimizertesting.Manager
	mOptDAO  *optdaotesting.DAO
	mRC      *robottesting.Controller
	mScanCtl *scantesting.Controller
	execMgr  *tasktesting.ExecutionManager
	taskMgr  *tasktesting.Manager
	mcp      *v1testing.ClientPool
	mc       *v1testing.Client

	meta *v1.OptimizerAdapterMetadata
	reg  *dao.Registration

	c *basicController
}

// TestController is the entry point of the ControllerTestSuite.
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

// SetupSuite initializes an in-memory config manager so config.ExtEndpoint(),
// config.InternalCoreURL() and config.ScannerRobotPrefix() do not panic.
func (suite *ControllerTestSuite) SetupSuite() {
	config.InitWithSettings(map[string]any{})
}

func (suite *ControllerTestSuite) SetupTest() {
	suite.mMgr = &optimizertesting.Manager{}
	suite.mOptDAO = &optdaotesting.DAO{}
	suite.mRC = &robottesting.Controller{}
	suite.mScanCtl = &scantesting.Controller{}
	suite.execMgr = &tasktesting.ExecutionManager{}
	suite.taskMgr = &tasktesting.Manager{}
	suite.mcp = &v1testing.ClientPool{}
	suite.mc = &v1testing.Client{}

	suite.meta = &v1.OptimizerAdapterMetadata{
		Optimizer: &v1.Optimizer{
			Name:    "Sleeko",
			Vendor:  "CERN",
			Version: "0.1.0",
		},
		Capabilities: []*v1.OptimizerCapability{{
			ConsumesMimeTypes: []string{"application/vnd.docker.distribution.manifest.v2+json"},
			ProducesMimeTypes: []string{v1.MimeTypeOptimizationReport},
		}},
	}

	suite.reg = &dao.Registration{
		Name: "forUT",
		URL:  "http://adapter:8080",
	}

	suite.c = &basicController{
		manager: suite.mMgr,
		optDAO:  suite.mOptDAO,
		rc:      suite.mRC,
		scanCtl: suite.mScanCtl,
		execMgr: suite.execMgr,
		taskMgr: suite.taskMgr,
		uuid: func() (string, error) {
			return "the-uuid-123", nil
		},
		clientPool: suite.mcp,
	}
}

func (suite *ControllerTestSuite) TestListRegistrations() {
	regs := []*dao.Registration{suite.reg}
	suite.mMgr.On("List", mock.Anything, mock.Anything).Return(regs, nil).Once()

	l, err := suite.c.ListRegistrations(context.TODO(), nil)
	suite.Require().NoError(err)
	suite.Len(l, 1)
}

func (suite *ControllerTestSuite) TestGetTotalOfRegistrations() {
	suite.mMgr.On("Count", mock.Anything, mock.Anything).Return(int64(3), nil).Once()

	total, err := suite.c.GetTotalOfRegistrations(context.TODO(), nil)
	suite.Require().NoError(err)
	suite.Equal(int64(3), total)
}

func (suite *ControllerTestSuite) TestCreateRegistration_ReservedName() {
	reg := &dao.Registration{Name: "Sleeko", URL: "http://adapter:8080"}

	_, err := suite.c.CreateRegistration(context.TODO(), reg)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "reserved")
}

func (suite *ControllerTestSuite) TestCreateRegistration_PingFails() {
	suite.mcp.On("Get", suite.reg.URL, suite.reg.Auth, suite.reg.AccessCredential, suite.reg.SkipCertVerify).
		Return(nil, fmt.Errorf("connection refused")).Once()

	_, err := suite.c.CreateRegistration(context.TODO(), suite.reg)
	suite.Require().Error(err)
}

func (suite *ControllerTestSuite) TestCreateRegistration_FirstBecomesDefault() {
	suite.mcp.On("Get", suite.reg.URL, suite.reg.Auth, suite.reg.AccessCredential, suite.reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	suite.mc.On("GetMetadata").Return(suite.meta, nil).Once()

	suite.mMgr.On("List", mock.Anything, mock.Anything).Return([]*dao.Registration{}, nil).Once()
	suite.mMgr.On("Create", mock.Anything, testifymock.MatchedBy(func(r *dao.Registration) bool {
		return r.IsDefault
	})).Return("new-uuid", nil).Once()

	uid, err := suite.c.CreateRegistration(context.TODO(), suite.reg)
	suite.Require().NoError(err)
	suite.Equal("new-uuid", uid)
}

func (suite *ControllerTestSuite) TestCreateRegistration_NotFirstKeepsExplicitDefault() {
	suite.mcp.On("Get", suite.reg.URL, suite.reg.Auth, suite.reg.AccessCredential, suite.reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	suite.mc.On("GetMetadata").Return(suite.meta, nil).Once()

	suite.mMgr.On("List", mock.Anything, mock.Anything).Return([]*dao.Registration{{Name: "existing"}}, nil).Once()
	suite.mMgr.On("Create", mock.Anything, testifymock.MatchedBy(func(r *dao.Registration) bool {
		return !r.IsDefault
	})).Return("new-uuid", nil).Once()

	_, err := suite.c.CreateRegistration(context.TODO(), suite.reg)
	suite.Require().NoError(err)
}

func (suite *ControllerTestSuite) TestGetRegistration() {
	suite.reg.UUID = "uuid001"
	suite.mMgr.On("Get", mock.Anything, "uuid001").Return(suite.reg, nil).Once()

	r, err := suite.c.GetRegistration(context.TODO(), "uuid001")
	suite.Require().NoError(err)
	suite.Equal("forUT", r.Name)
}

func (suite *ControllerTestSuite) TestUpdateRegistration_DefaultAndDisabled() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", IsDefault: true, Disabled: true}

	err := suite.c.UpdateRegistration(context.TODO(), reg)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "can not be marked to deactivated")
}

func (suite *ControllerTestSuite) TestUpdateRegistration_ReservedName() {
	reg := &dao.Registration{UUID: "uuid001", Name: "Sleeko"}

	err := suite.c.UpdateRegistration(context.TODO(), reg)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "reserved")
}

func (suite *ControllerTestSuite) TestUpdateRegistration_Success() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT"}
	suite.mMgr.On("Update", mock.Anything, reg).Return(nil).Once()

	err := suite.c.UpdateRegistration(context.TODO(), reg)
	suite.Require().NoError(err)
}

func (suite *ControllerTestSuite) TestDeleteRegistration_NotFound() {
	suite.mMgr.On("Get", mock.Anything, "missing").Return(nil, nil).Once()

	r, err := suite.c.DeleteRegistration(context.TODO(), "missing")
	suite.Require().NoError(err)
	suite.Nil(r)
}

func (suite *ControllerTestSuite) TestDeleteRegistration_Success() {
	suite.reg.UUID = "uuid001"
	suite.mMgr.On("Get", mock.Anything, "uuid001").Return(suite.reg, nil).Once()
	suite.mMgr.On("Delete", mock.Anything, "uuid001").Return(nil).Once()

	r, err := suite.c.DeleteRegistration(context.TODO(), "uuid001")
	suite.Require().NoError(err)
	suite.Equal("forUT", r.Name)
}

func (suite *ControllerTestSuite) TestSetDefaultRegistration() {
	suite.mMgr.On("SetAsDefault", mock.Anything, "uuid001").Return(nil).Once()

	err := suite.c.SetDefaultRegistration(context.TODO(), "uuid001")
	suite.Require().NoError(err)
}

func (suite *ControllerTestSuite) TestGetDefaultRegistration() {
	suite.mMgr.On("GetDefault", mock.Anything).Return(suite.reg, nil).Once()

	r, err := suite.c.GetDefaultRegistration(context.TODO())
	suite.Require().NoError(err)
	suite.Equal(suite.reg, r)
}

func (suite *ControllerTestSuite) TestPing_NilRegistration() {
	_, err := suite.c.Ping(context.TODO(), nil)
	suite.Require().Error(err)
}

func (suite *ControllerTestSuite) TestPing_NewRegistrationSkipsCache() {
	// ID == 0: registration not yet persisted, cache path (which needs a real ID) is skipped.
	suite.mcp.On("Get", suite.reg.URL, suite.reg.Auth, suite.reg.AccessCredential, suite.reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	suite.mc.On("GetMetadata").Return(suite.meta, nil).Once()

	meta, err := suite.c.Ping(context.TODO(), suite.reg)
	suite.Require().NoError(err)
	suite.Equal(suite.meta, meta)
}

func (suite *ControllerTestSuite) TestPing_InvalidMetadata() {
	suite.mcp.On("Get", suite.reg.URL, suite.reg.Auth, suite.reg.AccessCredential, suite.reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	suite.mc.On("GetMetadata").Return(&v1.OptimizerAdapterMetadata{}, nil).Once()

	_, err := suite.c.Ping(context.TODO(), suite.reg)
	suite.Require().Error(err)
}

func (suite *ControllerTestSuite) TestGetMetadata_EmptyUUID() {
	_, err := suite.c.GetMetadata(context.TODO(), "")
	suite.Require().Error(err)
}

func (suite *ControllerTestSuite) TestGetMetadata_NotFound() {
	suite.mMgr.On("Get", mock.Anything, "missing").Return(nil, nil).Once()

	_, err := suite.c.GetMetadata(context.TODO(), "missing")
	suite.Require().Error(err)
}

func (suite *ControllerTestSuite) TestGetMetadata_Success() {
	suite.reg.UUID = "uuid001"
	suite.mMgr.On("Get", mock.Anything, "uuid001").Return(suite.reg, nil).Once()
	suite.mcp.On("Get", suite.reg.URL, suite.reg.Auth, suite.reg.AccessCredential, suite.reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	suite.mc.On("GetMetadata").Return(suite.meta, nil).Once()

	meta, err := suite.c.GetMetadata(context.TODO(), "uuid001")
	suite.Require().NoError(err)
	suite.Equal(suite.meta, meta)
}

// sampleArtifact returns an artifact with a project/repo shape valid for
// makeRobotAccount ("<project>/<repo>").
func sampleArtifact() *ar.Artifact {
	return &ar.Artifact{Artifact: art.Artifact{
		ID:                1,
		ProjectID:         1,
		RepositoryName:    "library/photon",
		Digest:            "sha256:digest",
		ManifestMediaType: "application/vnd.docker.distribution.manifest.v2+json",
	}}
}

func (suite *ControllerTestSuite) TestOptimize_NilArtifact() {
	_, err := suite.c.Optimize(context.TODO(), nil, "latest")
	suite.Require().Error(err)
}

func (suite *ControllerTestSuite) TestOptimize_NoDefaultRegistration() {
	suite.mMgr.On("GetDefault", mock.Anything).Return(nil, nil).Once()

	_, err := suite.c.Optimize(context.TODO(), sampleArtifact(), "latest")
	suite.Require().Error(err)
	suite.Contains(err.Error(), "no optimizer adapter is configured")
}

func (suite *ControllerTestSuite) TestOptimize_RegistrationDisabled() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", Disabled: true}
	suite.mMgr.On("GetDefault", mock.Anything).Return(reg, nil).Once()

	_, err := suite.c.Optimize(context.TODO(), sampleArtifact(), "latest")
	suite.Require().Error(err)
	suite.Contains(err.Error(), "deactivated")
}

func (suite *ControllerTestSuite) TestOptimize_PingFails() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", URL: "http://adapter:8080"}
	suite.mMgr.On("GetDefault", mock.Anything).Return(reg, nil).Once()
	suite.mcp.On("Get", reg.URL, reg.Auth, reg.AccessCredential, reg.SkipCertVerify).
		Return(nil, fmt.Errorf("unreachable")).Once()

	_, err := suite.c.Optimize(context.TODO(), sampleArtifact(), "latest")
	suite.Require().Error(err)
	suite.Contains(err.Error(), "not reachable")
}

func (suite *ControllerTestSuite) TestOptimize_UnsupportedMimeType() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", URL: "http://adapter:8080"}
	suite.mMgr.On("GetDefault", mock.Anything).Return(reg, nil).Once()
	suite.mcp.On("Get", reg.URL, reg.Auth, reg.AccessCredential, reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	// Metadata that does not consume the artifact's mime type.
	suite.mc.On("GetMetadata").Return(&v1.OptimizerAdapterMetadata{
		Optimizer: suite.meta.Optimizer,
		Capabilities: []*v1.OptimizerCapability{{
			ConsumesMimeTypes: []string{"some/other-type"},
			ProducesMimeTypes: []string{v1.MimeTypeOptimizationReport},
		}},
	}, nil).Once()

	_, err := suite.c.Optimize(context.TODO(), sampleArtifact(), "latest")
	suite.Require().Error(err)
	suite.Contains(err.Error(), "does not support optimizing")
}

func (suite *ControllerTestSuite) TestOptimize_Success() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", URL: "http://adapter:8080"}
	artifact := sampleArtifact()

	suite.mMgr.On("GetDefault", mock.Anything).Return(reg, nil).Once()
	suite.mcp.On("Get", reg.URL, reg.Auth, reg.AccessCredential, reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	suite.mc.On("GetMetadata").Return(suite.meta, nil).Once()

	suite.execMgr.On("Create", mock.Anything, "OPTIMIZE_ARTIFACT", artifact.ID, mock.Anything, mock.Anything).
		Return(int64(100), nil).Once()

	suite.mOptDAO.On("Upsert", mock.Anything, testifymock.MatchedBy(func(rec *dockerfileoptdao.DockerfileOptimization) bool {
		return rec.Status == dockerfileoptdao.StatusPending &&
			rec.RepositoryName == artifact.RepositoryName &&
			rec.ArtifactDigest == artifact.Digest &&
			rec.ExecutionID == int64(100)
	})).Return(nil).Once()

	suite.mScanCtl.On("GetReport", mock.Anything, artifact, mock.Anything).
		Return([]*scandao.Report{{Status: job.SuccessStatus.String(), Report: `{"vulnerabilities":[]}`}}, nil).Once()

	robotAccount := &robot.Robot{
		Robot: model.Robot{
			ID:     1,
			Name:   "robot$-forUT-the-uuid-123",
			Secret: "robot-secret",
		},
		Level: robot.LEVELPROJECT,
	}
	suite.mRC.On("Create", mock.Anything, mock.Anything).Return(int64(1), "robot-secret", nil).Once()
	suite.mRC.On("Get", mock.Anything, int64(1), &robot.Option{WithPermission: false}).Return(robotAccount, nil).Once()

	suite.taskMgr.On("Create", mock.Anything, int64(100), mock.Anything, mock.Anything).Return(int64(1), nil).Once()

	execID, err := suite.c.Optimize(context.TODO(), artifact, "latest")
	suite.Require().NoError(err)
	suite.Equal(int64(100), execID)
}

func (suite *ControllerTestSuite) TestCheckAvailability_NilArtifact() {
	err := suite.c.CheckAvailability(context.TODO(), nil)
	suite.Require().Error(err)
}

func (suite *ControllerTestSuite) TestCheckAvailability_NoDefaultRegistration() {
	suite.mMgr.On("GetDefault", mock.Anything).Return(nil, nil).Once()

	err := suite.c.CheckAvailability(context.TODO(), sampleArtifact())
	suite.Require().Error(err)
	suite.Contains(err.Error(), "no optimizer adapter is configured")
}

func (suite *ControllerTestSuite) TestCheckAvailability_RegistrationDisabled() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", Disabled: true}
	suite.mMgr.On("GetDefault", mock.Anything).Return(reg, nil).Once()

	err := suite.c.CheckAvailability(context.TODO(), sampleArtifact())
	suite.Require().Error(err)
	suite.Contains(err.Error(), "deactivated")
}

func (suite *ControllerTestSuite) TestCheckAvailability_PingFails() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", URL: "http://adapter:8080"}
	suite.mMgr.On("GetDefault", mock.Anything).Return(reg, nil).Once()
	suite.mcp.On("Get", reg.URL, reg.Auth, reg.AccessCredential, reg.SkipCertVerify).
		Return(nil, fmt.Errorf("unreachable")).Once()

	err := suite.c.CheckAvailability(context.TODO(), sampleArtifact())
	suite.Require().Error(err)
	suite.Contains(err.Error(), "not reachable")
}

func (suite *ControllerTestSuite) TestCheckAvailability_UnsupportedMimeType() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", URL: "http://adapter:8080"}
	suite.mMgr.On("GetDefault", mock.Anything).Return(reg, nil).Once()
	suite.mcp.On("Get", reg.URL, reg.Auth, reg.AccessCredential, reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	suite.mc.On("GetMetadata").Return(&v1.OptimizerAdapterMetadata{
		Optimizer: suite.meta.Optimizer,
		Capabilities: []*v1.OptimizerCapability{{
			ConsumesMimeTypes: []string{"some/other-type"},
			ProducesMimeTypes: []string{v1.MimeTypeOptimizationReport},
		}},
	}, nil).Once()

	err := suite.c.CheckAvailability(context.TODO(), sampleArtifact())
	suite.Require().Error(err)
	suite.Contains(err.Error(), "does not support optimizing")
}

func (suite *ControllerTestSuite) TestCheckAvailability_Success() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", URL: "http://adapter:8080"}
	suite.mMgr.On("GetDefault", mock.Anything).Return(reg, nil).Once()
	suite.mcp.On("Get", reg.URL, reg.Auth, reg.AccessCredential, reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	suite.mc.On("GetMetadata").Return(suite.meta, nil).Once()

	err := suite.c.CheckAvailability(context.TODO(), sampleArtifact())
	suite.Require().NoError(err)
}

func (suite *ControllerTestSuite) TestOptimize_LaunchJobFailsMarksErrored() {
	reg := &dao.Registration{UUID: "uuid001", Name: "forUT", URL: "http://adapter:8080"}
	artifact := sampleArtifact()

	suite.mMgr.On("GetDefault", mock.Anything).Return(reg, nil).Once()
	suite.mcp.On("Get", reg.URL, reg.Auth, reg.AccessCredential, reg.SkipCertVerify).
		Return(suite.mc, nil).Once()
	suite.mc.On("GetMetadata").Return(suite.meta, nil).Once()

	suite.execMgr.On("Create", mock.Anything, "OPTIMIZE_ARTIFACT", artifact.ID, mock.Anything, mock.Anything).
		Return(int64(100), nil).Once()
	suite.mOptDAO.On("Upsert", mock.Anything, mock.Anything).Return(nil).Once()

	// robot creation fails -> launchOptimizeJob returns an error
	suite.mRC.On("Create", mock.Anything, mock.Anything).Return(int64(0), "", fmt.Errorf("robot creation failed")).Once()

	suite.mOptDAO.On("UpdateStatus", mock.Anything, artifact.RepositoryName, artifact.Digest,
		dockerfileoptdao.StatusError, mock.Anything).Return(nil).Once()

	_, err := suite.c.Optimize(context.TODO(), artifact, "latest")
	suite.Require().Error(err)
}

func (suite *ControllerTestSuite) TestFetchScanReport_SuccessReportReturned() {
	artifact := sampleArtifact()
	suite.mScanCtl.On("GetReport", mock.Anything, artifact, mock.Anything).
		Return([]*scandao.Report{{Status: job.SuccessStatus.String(), Report: `{"vulnerabilities":["CVE-1"]}`}}, nil).Once()

	report, err := suite.c.fetchScanReport(context.TODO(), artifact)
	suite.Require().NoError(err)
	suite.Equal(`{"vulnerabilities":["CVE-1"]}`, report)
}

func (suite *ControllerTestSuite) TestFetchScanReport_NoScannerConfiguredIsNotAnError() {
	artifact := sampleArtifact()
	suite.mScanCtl.On("GetReport", mock.Anything, artifact, mock.Anything).
		Return(nil, errors.NotFoundError(nil).WithMessage("no scanner registration configured")).Once()

	report, err := suite.c.fetchScanReport(context.TODO(), artifact)
	suite.Require().NoError(err)
	suite.Empty(report)
}

func (suite *ControllerTestSuite) TestFetchScanReport_NotYetSuccessfulReturnsEmpty() {
	artifact := sampleArtifact()
	suite.mScanCtl.On("GetReport", mock.Anything, artifact, mock.Anything).
		Return([]*scandao.Report{{Status: job.RunningStatus.String()}}, nil).Once()

	report, err := suite.c.fetchScanReport(context.TODO(), artifact)
	suite.Require().NoError(err)
	suite.Empty(report)
}

func (suite *ControllerTestSuite) TestFetchScanReport_OtherErrorPropagates() {
	artifact := sampleArtifact()
	suite.mScanCtl.On("GetReport", mock.Anything, artifact, mock.Anything).
		Return(nil, fmt.Errorf("db unavailable")).Once()

	_, err := suite.c.fetchScanReport(context.TODO(), artifact)
	suite.Require().Error(err)
}

func (suite *ControllerTestSuite) TestIsReservedName() {
	assert.True(suite.T(), isReservedName("Sleeko"))
	assert.False(suite.T(), isReservedName("something else"))
}
