package notification

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// QuotaPreprocessHandlerSuite ...
type QuotaPreprocessHandlerSuite struct {
	suite.Suite
	om  policy.Manager
	evt *model.QuotaEvent
}

// TestQuotaPreprocessHandler ...
func TestQuotaPreprocessHandler(t *testing.T) {
	suite.Run(t, &QuotaPreprocessHandlerSuite{})
}

// SetupSuite prepares env for test suite.
func (suite *QuotaPreprocessHandlerSuite) SetupSuite() {
	cfg := map[string]interface{}{
		common.NotificationEnable: true,
	}
	config.InitWithSettings(cfg)

	res := &model.ImgResource{
		Digest: "sha256:abcd",
		Tag:    "latest",
	}
	suite.evt = &model.QuotaEvent{
		EventType: model.EventTypeProjectQuota,
		OccurAt:   time.Now().UTC(),
		RepoName:  "hello-world",
		Resource:  res,
		Project: &models.Project{
			ProjectID: 1,
			Name:      "library",
		},
		Msg: "this is a testing quota event",
	}

	suite.om = notification.PolicyMgr
	mp := &fakedPolicyMgr{}
	notification.PolicyMgr = mp

	h := &MockHandler{}

	err := notifier.Subscribe(model.WebhookTopic, h)
	require.NoError(suite.T(), err)
}

// TearDownSuite ...
func (suite *QuotaPreprocessHandlerSuite) TearDownSuite() {
	notification.PolicyMgr = suite.om
}

// TestHandle ...
func (suite *QuotaPreprocessHandlerSuite) TestHandle() {
	handler := &QuotaPreprocessHandler{}
	err := handler.Handle(suite.evt)
	suite.NoError(err)
}

// MockHandler ...
type MockHandler struct{}

// Handle ...
func (m *MockHandler) Handle(value interface{}) error {
	return nil
}

// IsStateful ...
func (m *MockHandler) IsStateful() bool {
	return false
}
