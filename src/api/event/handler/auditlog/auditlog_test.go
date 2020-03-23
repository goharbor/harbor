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

package auditlog

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/api/event/metadata"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/audit/model"
	"github.com/goharbor/harbor/src/pkg/notifier"
	ne "github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
	suite.auditLogHandler = &Handler{}
}

func (suite *AuditLogHandlerTestSuite) TestSubscribeTagEvent() {

	suite.logMgr.On("Create", mock.Anything).Return(1, nil)
	suite.logMgr.On("Count", mock.Anything).Return(1, nil)

	// sample code to use the event framework.

	notifier.Subscribe(event.TopicCreateProject, suite.auditLogHandler)
	// event data should implement the interface TopicEvent
	ne.BuildAndPublish(&metadata.CreateProjectEventMetadata{
		ProjectID: 1,
		Project:   "test",
		Operator:  "admin",
	})
	cnt, err := suite.logMgr.Count(nil, nil)

	suite.Nil(err)
	suite.Equal(int64(1), cnt)

}

func TestAuditLogHandlerTestSuite(t *testing.T) {
	suite.Run(t, &AuditLogHandlerTestSuite{})
}
