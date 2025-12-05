package handler

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/rbac"
	auditmodel "github.com/goharbor/harbor/src/pkg/audit/model"
	auditextmodel "github.com/goharbor/harbor/src/pkg/auditext/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	audittesting "github.com/goharbor/harbor/src/testing/pkg/audit"
	auditexttesting "github.com/goharbor/harbor/src/testing/pkg/auditext"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

type AuditLogTestSuite struct {
	htesting.Suite

	auditMgr    *audittesting.Manager
	auditextMgr *auditexttesting.Manager
	projectCtl  *projecttesting.Controller
}

func (suite *AuditLogTestSuite) SetupSuite() {
	suite.auditMgr = &audittesting.Manager{}
	suite.auditextMgr = &auditexttesting.Manager{}
	suite.projectCtl = &projecttesting.Controller{}

	suite.Config = &restapi.Config{
		AuditlogAPI: &auditlogAPI{
			auditMgr:    suite.auditMgr,
			auditextMgr: suite.auditextMgr,
			projectCtl:  suite.projectCtl,
		},
	}

	suite.Suite.SetupSuite()
}

func (suite *AuditLogTestSuite) TestListAuditLogs() {
	// Mock security context
	suite.Security.On("IsAuthenticated").Return(true)
	suite.Security.On("GetUsername").Return("admin")
	// Mock system access check
	suite.Security.On("Can", rbac.ActionList, rbac.ResourceAuditLog).Return(true)

	// Case 1: Success
	{
		suite.auditMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil).Once()
		suite.auditMgr.On("List", mock.Anything, mock.Anything).Return([]*auditmodel.AuditLog{
			{
				ID:           1,
				Resource:     "library/ubuntu",
				ResourceType: "artifact",
				Username:     "admin",
				Operation:    "create",
				OpTime:       time.Now(),
			},
		}, nil).Once()

		var logs []*models.AuditLog
		res, err := suite.GetJSON("/audit-logs", &logs)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(logs, 1)
		suite.Equal(int64(1), logs[0].ID)
	}

	// Case 2: Count failed
	{
		suite.auditMgr.On("Count", mock.Anything, mock.Anything).Return(int64(0), fmt.Errorf("count error")).Once()

		res, err := suite.Get("/audit-logs")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	// Case 3: List failed
	{
		suite.auditMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil).Once()
		suite.auditMgr.On("List", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("list error")).Once()

		res, err := suite.Get("/audit-logs")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}
}

func (suite *AuditLogTestSuite) TestListAuditLogExts() {
	// Mock security context
	suite.Security.On("IsAuthenticated").Return(true)
	suite.Security.On("GetUsername").Return("admin")
	// Mock system access check
	suite.Security.On("Can", rbac.ActionList, rbac.ResourceAuditLog).Return(true)

	// Case 1: Success
	{
		suite.auditextMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil).Once()
		suite.auditextMgr.On("List", mock.Anything, mock.Anything).Return([]*auditextmodel.AuditLogExt{
			{
				ID:                   1,
				Resource:             "library/ubuntu",
				ResourceType:         "artifact",
				Username:             "admin",
				Operation:            "create",
				OperationDescription: "create artifact",
				IsSuccessful:         true,
				OpTime:               time.Now(),
			},
		}, nil).Once()

		var logs []*models.AuditLogExt
		res, err := suite.GetJSON("/audit-logs/ext", &logs)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Len(logs, 1)
		suite.Equal(int64(1), logs[0].ID)
	}

	// Case 2: Count failed
	{
		suite.auditextMgr.On("Count", mock.Anything, mock.Anything).Return(int64(0), fmt.Errorf("count error")).Once()

		res, err := suite.Get("/audit-logs/ext")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	// Case 3: List failed
	{
		suite.auditextMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil).Once()
		suite.auditextMgr.On("List", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("list error")).Once()

		res, err := suite.Get("/audit-logs/ext")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}
}

func (suite *AuditLogTestSuite) TestListAuditLogEventTypes() {
	// Mock security context
	suite.Security.On("IsAuthenticated").Return(true)
	suite.Security.On("GetUsername").Return("admin")
	// Mock system access check
	suite.Security.On("Can", rbac.ActionList, rbac.ResourceAuditLog).Return(true)

	var eventTypes []*models.AuditLogEventType
	res, err := suite.GetJSON("/audit-logs/event-types", &eventTypes)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.NotEmpty(eventTypes)
}

func TestAuditLogTestSuite(t *testing.T) {
	suite.Run(t, &AuditLogTestSuite{})
}
