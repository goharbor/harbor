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

package notification

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/api/artifact"
	sc "github.com/goharbor/harbor/src/api/scan"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	artifacttesting "github.com/goharbor/harbor/src/testing/api/artifact"
	scantesting "github.com/goharbor/harbor/src/testing/api/scan"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ScanImagePreprocessHandlerSuite is a test suite to test scan image preprocess handler.
type ScanImagePreprocessHandlerSuite struct {
	suite.Suite

	om          policy.Manager
	pid         int64
	evt         *model.ScanImageEvent
	c           sc.Controller
	artifactCtl artifact.Controller
}

// TestScanImagePreprocessHandler is the entry point of ScanImagePreprocessHandlerSuite.
func TestScanImagePreprocessHandler(t *testing.T) {
	suite.Run(t, &ScanImagePreprocessHandlerSuite{})
}

// SetupSuite prepares env for test suite.
func (suite *ScanImagePreprocessHandlerSuite) SetupSuite() {
	cfg := map[string]interface{}{
		common.NotificationEnable: true,
	}
	config.InitWithSettings(cfg)

	a := &v1.Artifact{
		NamespaceID: int64(1),
		Repository:  "library/redis",
		Tag:         "latest",
		Digest:      "digest-code",
		MimeType:    v1.MimeTypeDockerArtifact,
	}
	suite.evt = &model.ScanImageEvent{
		EventType: model.EventTypeScanningCompleted,
		OccurAt:   time.Now().UTC(),
		Operator:  "admin",
		Artifact:  a,
	}

	reports := []*scan.Report{
		{
			Report: "{}",
		},
	}

	suite.c = sc.DefaultController
	mc := &scantesting.Controller{}

	var options []report.Option
	s := make(map[string]interface{})
	mc.On("GetSummary", a, []string{v1.MimeTypeNativeReport}, options).Return(s, nil)
	mock.OnAnything(mc, "GetSummary").Return(s, nil)
	mock.OnAnything(mc, "GetReport").Return(reports, nil)

	sc.DefaultController = mc

	suite.artifactCtl = artifact.Ctl

	artifactCtl := &artifacttesting.Controller{}

	art := &artifact.Artifact{}
	art.ProjectID = a.NamespaceID
	art.RepositoryName = a.Repository
	art.Digest = a.Digest

	mock.OnAnything(artifactCtl, "GetByReference").Return(art, nil)

	artifact.Ctl = artifactCtl

	suite.om = notification.PolicyMgr
	mp := &fakedPolicyMgr{}
	notification.PolicyMgr = mp

	h := &MockHTTPHandler{}

	err := notifier.Subscribe(model.WebhookTopic, h)
	require.NoError(suite.T(), err)
}

// TearDownSuite clears the env for test suite.
func (suite *ScanImagePreprocessHandlerSuite) TearDownSuite() {
	notification.PolicyMgr = suite.om
	sc.DefaultController = suite.c
	artifact.Ctl = suite.artifactCtl
}

// TestHandle ...
func (suite *ScanImagePreprocessHandlerSuite) TestHandle() {
	handler := &ScanImagePreprocessHandler{}

	err := handler.Handle(suite.evt)
	suite.NoError(err)
}

// Mock things

// MockHTTPHandler ...
type MockHTTPHandler struct{}

// Handle ...
func (m *MockHTTPHandler) Handle(value interface{}) error {
	return nil
}

// IsStateful ...
func (m *MockHTTPHandler) IsStateful() bool {
	return false
}
