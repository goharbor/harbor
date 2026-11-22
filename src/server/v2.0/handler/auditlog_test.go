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

package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/auditext/model"
	pkgModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/auditlog"
	projectop "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	auditexttesting "github.com/goharbor/harbor/src/testing/pkg/auditext"
)

type AuditLogTestSuite struct {
	suite.Suite
	auditextMgr *auditexttesting.Manager
	projectCtl  *projecttesting.Controller
	secCtx      *securitytesting.Context
	ctx         context.Context
	api         *auditlogAPI
	projApi     *projectAPI
}

func (suite *AuditLogTestSuite) SetupSuite() {
	suite.auditextMgr = &auditexttesting.Manager{}
	suite.projectCtl = &projecttesting.Controller{}
	suite.secCtx = &securitytesting.Context{}
	suite.ctx = security.NewContext(context.Background(), suite.secCtx)

	suite.api = &auditlogAPI{
		auditextMgr: suite.auditextMgr,
		projectCtl:  suite.projectCtl,
	}
	suite.projApi = &projectAPI{
		auditextMgr: suite.auditextMgr,
		projectCtl:  suite.projectCtl,
	}
	baseProjectCtl = suite.projectCtl
}

func (suite *AuditLogTestSuite) TearDownSuite() {
	baseProjectCtl = project.Ctl
}

func (suite *AuditLogTestSuite) SetupTest() {
	suite.auditextMgr.ExpectedCalls = nil
	suite.auditextMgr.Calls = nil
	suite.projectCtl.ExpectedCalls = nil
	suite.projectCtl.Calls = nil
	suite.secCtx.ExpectedCalls = nil
	suite.secCtx.Calls = nil

	suite.secCtx.On("IsAuthenticated").Return(true)
	suite.secCtx.On("GetUsername").Return("admin")
	suite.secCtx.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.secCtx.On("IsSysAdmin").Return(true)
}

func makeAuditLogs(n int) []*model.AuditLogExt {
	var logs []*model.AuditLogExt
	for i := 1; i <= n; i++ {
		logs = append(logs, &model.AuditLogExt{
			ID:                   int64(i),
			ProjectID:            1,
			Resource:             fmt.Sprintf("library/repo-%d", i),
			ResourceType:         "artifact",
			Username:             "admin",
			Operation:            "pull",
			OperationDescription: "pull artifact",
			IsSuccessful:         true,
			OpTime:               time.Now(),
		})
	}
	return logs
}

// TestListAuditLogExts_MultiPage_NextPageExists verifies that when a next page exists:
// - List is queried with ExtraLimit = 1
// - Count is NEVER called
// - The extra record is trimmed from the returned payload
// - X-Total-Count is set to pageNumber * pageSize + 1
// - Link header includes rel="next"
func (suite *AuditLogTestSuite) TestListAuditLogExts_MultiPage_NextPageExists() {
	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.PageNumber == 1 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return(makeAuditLogs(6), nil).Once()

	page := int64(1)
	pageSize := int64(5)
	params := auditlog.ListAuditLogExtsParams{
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.api.ListAuditLogExts(suite.ctx, params)
	okResp, ok := responder.(*auditlog.ListAuditLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(6), okResp.XTotalCount)
	suite.Contains(okResp.Link, `rel="next"`)
	suite.NotContains(okResp.Link, `rel="prev"`)
	suite.Len(okResp.Payload, 5)

	// Ensure Count was never called (no table scan)
	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestListAuditLogExts_SinglePage_ExactFit verifies that when all rows fit in one page:
// - List is queried with ExtraLimit = 1
// - Count is NEVER called
// - X-Total-Count reflects the exact count
// - Link header does NOT include rel="next"
func (suite *AuditLogTestSuite) TestListAuditLogExts_SinglePage_ExactFit() {
	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.PageNumber == 1 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return(makeAuditLogs(3), nil).Once()

	page := int64(1)
	pageSize := int64(5)
	params := auditlog.ListAuditLogExtsParams{
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.api.ListAuditLogExts(suite.ctx, params)
	okResp, ok := responder.(*auditlog.ListAuditLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(3), okResp.XTotalCount)
	suite.NotContains(okResp.Link, `rel="next"`)
	suite.NotContains(okResp.Link, `rel="prev"`)
	suite.Len(okResp.Payload, 3)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestListAuditLogExts_MiddlePage verifies that on a middle page:
// - Link header includes BOTH rel="prev" and rel="next"
// - X-Total-Count is pageNumber * pageSize + 1
// - Returned payload is trimmed to pageSize
func (suite *AuditLogTestSuite) TestListAuditLogExts_MiddlePage() {
	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.PageNumber == 2 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return(makeAuditLogs(6), nil).Once()

	page := int64(2)
	pageSize := int64(5)
	params := auditlog.ListAuditLogExtsParams{
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.api.ListAuditLogExts(suite.ctx, params)
	okResp, ok := responder.(*auditlog.ListAuditLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(11), okResp.XTotalCount)
	suite.Contains(okResp.Link, `rel="prev"`)
	suite.Contains(okResp.Link, `rel="next"`)
	suite.Len(okResp.Payload, 5)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestListAuditLogExts_LastPage verifies that on the last page:
// - Count is exact: (pageNumber - 1) * pageSize + len(logs)
// - Link header includes rel="prev" but NOT rel="next"
func (suite *AuditLogTestSuite) TestListAuditLogExts_LastPage() {
	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.PageNumber == 2 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return(makeAuditLogs(2), nil).Once()

	page := int64(2)
	pageSize := int64(5)
	params := auditlog.ListAuditLogExtsParams{
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.api.ListAuditLogExts(suite.ctx, params)
	okResp, ok := responder.(*auditlog.ListAuditLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(7), okResp.XTotalCount)
	suite.Contains(okResp.Link, `rel="prev"`)
	suite.NotContains(okResp.Link, `rel="next"`)
	suite.Len(okResp.Payload, 2)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestListAuditLogExts_Empty verifies an empty result:
// - X-Total-Count is 0
// - Link header has no next or prev
func (suite *AuditLogTestSuite) TestListAuditLogExts_Empty() {
	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.PageNumber == 1 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return([]*model.AuditLogExt{}, nil).Once()

	page := int64(1)
	pageSize := int64(5)
	params := auditlog.ListAuditLogExtsParams{
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.api.ListAuditLogExts(suite.ctx, params)
	okResp, ok := responder.(*auditlog.ListAuditLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(0), okResp.XTotalCount)
	suite.NotContains(okResp.Link, `rel="next"`)
	suite.NotContains(okResp.Link, `rel="prev"`)
	suite.Len(okResp.Payload, 0)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestListAuditLogExts_WithFilter verifies queries with filters (e.g., project_id):
// - Filter is correctly preserved on listQuery
// - Count is NOT called
func (suite *AuditLogTestSuite) TestListAuditLogExts_WithFilter() {
	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.Keywords["project_id"] == "3" && query.ExtraLimit == 1
	})).Return(makeAuditLogs(2), nil).Once()

	page := int64(1)
	pageSize := int64(5)
	qStr := "project_id=3"
	params := auditlog.ListAuditLogExtsParams{
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
		Q:           &qStr,
	}

	responder := suite.api.ListAuditLogExts(suite.ctx, params)
	okResp, ok := responder.(*auditlog.ListAuditLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(2), okResp.XTotalCount)
	suite.Len(okResp.Payload, 2)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestListAuditLogExts_Unbounded verifies queries without pageSize:
// - ExtraLimit is 0
// - Count is NOT called
func (suite *AuditLogTestSuite) TestListAuditLogExts_Unbounded() {
	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.PageSize == 0 && query.ExtraLimit == 0
	})).Return(makeAuditLogs(4), nil).Once()

	pageSize := int64(0)
	params := auditlog.ListAuditLogExtsParams{
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/auditlog-exts"}},
		PageSize:    &pageSize,
	}

	responder := suite.api.ListAuditLogExts(suite.ctx, params)
	okResp, ok := responder.(*auditlog.ListAuditLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(4), okResp.XTotalCount)
	suite.Len(okResp.Payload, 4)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestListAuditLogExts_DatabaseError verifies that DB errors are handled properly:
// - Returns error responder
// - Count is NOT called
func (suite *AuditLogTestSuite) TestListAuditLogExts_DatabaseError() {
	suite.auditextMgr.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("database connection failed")).Once()

	page := int64(1)
	pageSize := int64(5)
	params := auditlog.ListAuditLogExtsParams{
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.api.ListAuditLogExts(suite.ctx, params)
	suite.IsType(&ErrResponder{}, responder)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestProjectGetLogExts_MultiPage verifies that project-scoped GetLogExts:
// - Applies page-size-plus-one pagination
// - Never calls Count
// - Trims results and sets X-Total-Count and Link
func (suite *AuditLogTestSuite) TestProjectGetLogExts_MultiPage() {
	suite.projectCtl.On("GetByName", mock.Anything, "library").Return(&pkgModels.Project{
		ProjectID: 1,
		Name:      "library",
	}, nil)

	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.Keywords["ProjectID"] == int64(1) && query.PageNumber == 1 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return(makeAuditLogs(6), nil).Once()

	page := int64(1)
	pageSize := int64(5)
	params := projectop.GetLogExtsParams{
		ProjectName: "library",
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/projects/library/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.projApi.GetLogExts(suite.ctx, params)
	okResp, ok := responder.(*projectop.GetLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(6), okResp.XTotalCount)
	suite.Contains(okResp.Link, `rel="next"`)
	suite.Len(okResp.Payload, 5)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestProjectGetLogExts_LastPage verifies that project-scoped GetLogExts on the last page:
// - Never calls Count
// - Returns exact total
// - Link header has rel="prev" but no rel="next"
func (suite *AuditLogTestSuite) TestProjectGetLogExts_LastPage() {
	suite.projectCtl.On("GetByName", mock.Anything, "library").Return(&pkgModels.Project{
		ProjectID: 1,
		Name:      "library",
	}, nil)

	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.Keywords["ProjectID"] == int64(1) && query.PageNumber == 2 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return(makeAuditLogs(2), nil).Once()

	page := int64(2)
	pageSize := int64(5)
	params := projectop.GetLogExtsParams{
		ProjectName: "library",
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/projects/library/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.projApi.GetLogExts(suite.ctx, params)
	okResp, ok := responder.(*projectop.GetLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(7), okResp.XTotalCount)
	suite.Contains(okResp.Link, `rel="prev"`)
	suite.NotContains(okResp.Link, `rel="next"`)
	suite.Len(okResp.Payload, 2)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestListAuditLogExts_OutOfBoundsPage verifies that when navigating beyond the end of results:
// - Count is NOT called
// - Link header includes rel="prev" (page=pageNumber-1) allowing the user to navigate back
// - Link header does NOT include rel="next"
func (suite *AuditLogTestSuite) TestListAuditLogExts_OutOfBoundsPage() {
	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.PageNumber == 5 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return([]*model.AuditLogExt{}, nil).Once()

	page := int64(5)
	pageSize := int64(5)
	params := auditlog.ListAuditLogExtsParams{
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.api.ListAuditLogExts(suite.ctx, params)
	okResp, ok := responder.(*auditlog.ListAuditLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(20), okResp.XTotalCount) // (page - 1) * pageSize + len(logs) = (5 - 1) * 5 + 0
	suite.Contains(okResp.Link, `rel="prev"`)
	suite.Contains(okResp.Link, `page=4`)
	suite.NotContains(okResp.Link, `rel="next"`)
	suite.Len(okResp.Payload, 0)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestProjectGetLogExts_Empty verifies an empty result for project logs:
// - Count is NOT called
// - X-Total-Count is 0
// - Link header has no prev or next
func (suite *AuditLogTestSuite) TestProjectGetLogExts_Empty() {
	suite.projectCtl.On("GetByName", mock.Anything, "library").Return(&pkgModels.Project{
		ProjectID: 1,
		Name:      "library",
	}, nil)

	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.Keywords["ProjectID"] == int64(1) && query.PageNumber == 1 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return([]*model.AuditLogExt{}, nil).Once()

	page := int64(1)
	pageSize := int64(5)
	params := projectop.GetLogExtsParams{
		ProjectName: "library",
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/projects/library/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.projApi.GetLogExts(suite.ctx, params)
	okResp, ok := responder.(*projectop.GetLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(0), okResp.XTotalCount)
	suite.NotContains(okResp.Link, `rel="next"`)
	suite.NotContains(okResp.Link, `rel="prev"`)
	suite.Len(okResp.Payload, 0)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestProjectGetLogExts_OutOfBoundsPage verifies out-of-bounds page for project logs:
// - rel="prev" is present to navigate back
// - rel="next" is absent
func (suite *AuditLogTestSuite) TestProjectGetLogExts_OutOfBoundsPage() {
	suite.projectCtl.On("GetByName", mock.Anything, "library").Return(&pkgModels.Project{
		ProjectID: 1,
		Name:      "library",
	}, nil)

	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.Keywords["ProjectID"] == int64(1) && query.PageNumber == 4 && query.PageSize == 5 && query.ExtraLimit == 1
	})).Return([]*model.AuditLogExt{}, nil).Once()

	page := int64(4)
	pageSize := int64(5)
	params := projectop.GetLogExtsParams{
		ProjectName: "library",
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/projects/library/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.projApi.GetLogExts(suite.ctx, params)
	okResp, ok := responder.(*projectop.GetLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(15), okResp.XTotalCount) // (page - 1) * pageSize + len(logs) = (4 - 1) * 5 + 0
	suite.Contains(okResp.Link, `rel="prev"`)
	suite.Contains(okResp.Link, `page=3`)
	suite.NotContains(okResp.Link, `rel="next"`)
	suite.Len(okResp.Payload, 0)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestProjectGetLogExts_DatabaseError verifies DB error handling for project logs
func (suite *AuditLogTestSuite) TestProjectGetLogExts_DatabaseError() {
	suite.projectCtl.On("GetByName", mock.Anything, "library").Return(&pkgModels.Project{
		ProjectID: 1,
		Name:      "library",
	}, nil)

	suite.auditextMgr.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("connection reset")).Once()

	page := int64(1)
	pageSize := int64(5)
	params := projectop.GetLogExtsParams{
		ProjectName: "library",
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/projects/library/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.projApi.GetLogExts(suite.ctx, params)
	suite.IsType(&ErrResponder{}, responder)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestProjectGetLogExts_Unbounded verifies unbounded project logs query
func (suite *AuditLogTestSuite) TestProjectGetLogExts_Unbounded() {
	suite.projectCtl.On("GetByName", mock.Anything, "library").Return(&pkgModels.Project{
		ProjectID: 1,
		Name:      "library",
	}, nil)

	suite.auditextMgr.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
		return query.PageSize == 0 && query.ExtraLimit == 0
	})).Return(makeAuditLogs(3), nil).Once()

	pageSize := int64(0)
	params := projectop.GetLogExtsParams{
		ProjectName: "library",
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/projects/library/auditlog-exts"}},
		PageSize:    &pageSize,
	}

	responder := suite.projApi.GetLogExts(suite.ctx, params)
	okResp, ok := responder.(*projectop.GetLogExtsOK)
	suite.True(ok)
	suite.Equal(int64(3), okResp.XTotalCount)
	suite.Len(okResp.Payload, 3)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertExpectations(suite.T())
}

// TestProjectGetLogExts_ProjectNotFound verifies error when project does not exist
func (suite *AuditLogTestSuite) TestProjectGetLogExts_ProjectNotFound() {
	suite.projectCtl.On("GetByName", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	page := int64(1)
	pageSize := int64(5)
	params := projectop.GetLogExtsParams{
		ProjectName: "nonexistent",
		HTTPRequest: &http.Request{URL: &url.URL{Path: "/api/v2.0/projects/nonexistent/auditlog-exts"}},
		Page:        &page,
		PageSize:    &pageSize,
	}

	responder := suite.projApi.GetLogExts(suite.ctx, params)
	suite.IsType(&ErrResponder{}, responder)

	suite.auditextMgr.AssertNotCalled(suite.T(), "Count", mock.Anything, mock.Anything)
	suite.auditextMgr.AssertNotCalled(suite.T(), "List", mock.Anything, mock.Anything)
}

func TestAuditLogTestSuite(t *testing.T) {
	suite.Run(t, new(AuditLogTestSuite))
}
