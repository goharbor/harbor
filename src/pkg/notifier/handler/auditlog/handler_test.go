package auditlog

import (
	"context"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/audit/model"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	nm "github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type MockAuditLogManager struct {
	mock.Mock
}

func (m *MockAuditLogManager) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	args := m.Called()
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockAuditLogManager) Create(ctx context.Context, audit *model.AuditLog) (id int64, err error) {
	args := m.Called()
	return int64(args.Int(0)), args.Error(1)
}

func (m *MockAuditLogManager) Delete(ctx context.Context, id int64) (err error) {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAuditLogManager) Get(ctx context.Context, id int64) (audit *model.AuditLog, err error) {
	args := m.Called()
	return args.Get(0).(*model.AuditLog), args.Error(1)
}

func (m *MockAuditLogManager) List(ctx context.Context, query *q.Query) (audits []*model.AuditLog, err error) {
	args := m.Called()
	return args.Get(0).([]*model.AuditLog), args.Error(1)
}

type AuditLogHandlerTestSuite struct {
	suite.Suite
	auditLogHandler *Handler
	logMgr          *MockAuditLogManager
}

func (suite *AuditLogHandlerTestSuite) SetupSuite() {
	common_dao.PrepareTestForPostgresSQL()
	suite.logMgr = &MockAuditLogManager{}
	suite.auditLogHandler = &Handler{AuditLogMgr: suite.logMgr}
	log.SetLevel(log.DebugLevel)
}

func (suite *AuditLogHandlerTestSuite) TestSubscribeTagEvent() {

	suite.logMgr.On("Create", mock.Anything).Return(1, nil)
	suite.logMgr.On("Count", mock.Anything).Return(1, nil)

	// sample code to use the event framework.

	notifier.Subscribe(nm.PushTagTopic, suite.auditLogHandler)
	// event data should implement the interface TopicEvent
	data := &nm.TagEvent{
		TargetTopic: nm.PushTagTopic, // Topic is a attribute of event
		Project: &models.Project{
			ProjectID: 1,
			Name:      "library",
		},
		RepoName:  "busybox",
		Digest:    "abcdef",
		TagName:   "dev",
		OccurAt:   time.Now(),
		Operator:  "admin",
		Operation: "push", // Use Operation instead of event type.
	}
	// No EventMetadata anymore and there is no need to call resolve
	// The handler receives the TagEvent
	// The handler should use switch type interface to get TagEvent
	event.New().WithTopicEvent(data).Publish()

	cnt, err := suite.logMgr.Count(nil, nil)

	suite.Nil(err)
	suite.Equal(int64(1), cnt)

}

func TestAuditLogHandlerTestSuite(t *testing.T) {
	suite.Run(t, &AuditLogHandlerTestSuite{})
}
