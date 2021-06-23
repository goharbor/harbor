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

package quota

import (
	"context"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	policy_model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/testing/mock"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	testing_notification "github.com/goharbor/harbor/src/testing/pkg/notification/policy"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// QuotaPreprocessHandlerSuite ...
type QuotaPreprocessHandlerSuite struct {
	suite.Suite
	om  policy.Manager
	evt *event.QuotaEvent
}

// TestQuotaPreprocessHandler ...
func TestQuotaPreprocessHandler(t *testing.T) {
	suite.Run(t, &QuotaPreprocessHandlerSuite{})
}

// SetupSuite prepares env for test suite.
func (suite *QuotaPreprocessHandlerSuite) SetupSuite() {
	common_dao.PrepareTestForPostgresSQL()
	cfg := map[string]interface{}{
		common.NotificationEnable: true,
	}
	config.InitWithSettings(cfg)

	res := &event.ImgResource{
		Digest: "sha256:abcd",
		Tag:    "latest",
	}
	suite.evt = &event.QuotaEvent{
		EventType: event.TopicQuotaExceed,
		OccurAt:   time.Now().UTC(),
		RepoName:  "hello-world",
		Resource:  res,
		Project: &proModels.Project{
			ProjectID: 1,
			Name:      "library",
		},
		Msg: "this is a testing quota event",
	}

	suite.om = notification.PolicyMgr
	mp := &testing_notification.Manager{}
	notification.PolicyMgr = mp
	mp.On("GetRelatedPolices", mock.Anything, mock.Anything, mock.Anything).Return([]*policy_model.Policy{
		{
			ID: 1,
		},
	}, nil)

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
	handler := &Handler{}
	err := handler.Handle(context.TODO(), suite.evt)
	suite.NoError(err)
}

// MockHandler ...
type MockHandler struct{}

// Name ...
func (m *MockHandler) Name() string {
	return "Mock"
}

// Handle ...
func (m *MockHandler) Handle(ctx context.Context, value interface{}) error {
	return nil
}

// IsStateful ...
func (m *MockHandler) IsStateful() bool {
	return false
}
