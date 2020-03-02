package notification

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/scan/all"

	sc "github.com/goharbor/harbor/src/api/scan"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
	nm "github.com/goharbor/harbor/src/pkg/notification/model"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ScanImagePreprocessHandlerSuite is a test suite to test scan image preprocess handler.
type ScanImagePreprocessHandlerSuite struct {
	suite.Suite

	om  policy.Manager
	pid int64
	evt *model.ScanImageEvent
	c   sc.Controller
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
		EventType: nm.EventTypeScanningCompleted,
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
	mc := &MockScanAPIController{}

	var options []report.Option
	s := make(map[string]interface{})
	mc.On("GetSummary", a, []string{v1.MimeTypeNativeReport}, options).Return(s, nil)
	mc.On("GetReport", a, []string{v1.MimeTypeNativeReport}).Return(reports, nil)

	sc.DefaultController = mc

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
}

// TestHandle ...
func (suite *ScanImagePreprocessHandlerSuite) TestHandle() {
	handler := &ScanImagePreprocessHandler{}

	err := handler.Handle(suite.evt)
	suite.NoError(err)
}

// Mock things

// MockScanAPIController ...
type MockScanAPIController struct {
	mock.Mock
}

// Scan ...
func (msc *MockScanAPIController) Scan(artifact *v1.Artifact, option ...sc.Option) error {
	args := msc.Called(artifact)

	return args.Error(0)
}

func (msc *MockScanAPIController) GetReport(artifact *v1.Artifact, mimeTypes []string) ([]*scan.Report, error) {
	args := msc.Called(artifact, mimeTypes)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*scan.Report), args.Error(1)
}

func (msc *MockScanAPIController) GetSummary(artifact *v1.Artifact, mimeTypes []string, options ...report.Option) (map[string]interface{}, error) {
	args := msc.Called(artifact, mimeTypes, options)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (msc *MockScanAPIController) GetScanLog(uuid string) ([]byte, error) {
	args := msc.Called(uuid)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

func (msc *MockScanAPIController) HandleJobHooks(trackID string, change *job.StatusChange) error {
	args := msc.Called(trackID, change)

	return args.Error(0)
}

func (msc *MockScanAPIController) DeleteReports(digests ...string) error {
	pl := make([]interface{}, 0)
	for _, d := range digests {
		pl = append(pl, d)
	}
	args := msc.Called(pl...)

	return args.Error(0)
}

func (msc *MockScanAPIController) GetStats(requester string) (*all.Stats, error) {
	args := msc.Called(requester)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*all.Stats), args.Error(1)
}

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
