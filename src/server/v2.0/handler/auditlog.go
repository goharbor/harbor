package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/audit"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/auditlog"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/auditlog"
)

func newAuditLogAPI() *auditlogAPI {
	return &auditlogAPI{
		auditMgr:   audit.Mgr,
		projectCtl: project.Ctl,
	}
}

type auditlogAPI struct {
	BaseAPI
	auditMgr   audit.Manager
	projectCtl project.Controller
}

func (a *auditlogAPI) ListAuditLogs(ctx context.Context, params auditlog.ListAuditLogsParams) middleware.Responder {
	secCtx, ok := security.FromContext(ctx)
	if !ok {
		return a.SendError(ctx, errors.UnauthorizedError(errors.New("security context not found")))
	}
	if !secCtx.IsAuthenticated() {
		return a.SendError(ctx, errors.UnauthorizedError(nil).WithMessage(secCtx.GetUsername()))
	}
	query, err := a.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}

	if err := a.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceAuditLog); err != nil {
		ol := &q.OrList{}
		if sc, ok := secCtx.(*local.SecurityContext); ok && sc.IsAuthenticated() {
			user := sc.User()
			member := &project.MemberQuery{
				UserID:   user.UserID,
				GroupIDs: user.GroupIDs,
			}

			projects, err := a.projectCtl.List(ctx, q.New(q.KeyWords{"member": member}), project.Metadata(false))
			if err != nil {
				return a.SendError(ctx, fmt.Errorf(
					"failed to get projects of user %s: %v", secCtx.GetUsername(), err))
			}
			for _, project := range projects {
				if a.HasProjectPermission(ctx, project.ProjectID, rbac.ActionList, rbac.ResourceLog) {
					ol.Values = append(ol.Values, project.ProjectID)
				}
			}
		}
		// make sure no project will be selected with the query
		if len(ol.Values) == 0 {
			ol.Values = append(ol.Values, -1)
		}
		query.Keywords["ProjectID"] = ol
	}

	total, err := a.auditMgr.Count(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}
	logs, err := a.auditMgr.List(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}

	var auditLogs []*models.AuditLog
	for _, log := range logs {
		auditLogs = append(auditLogs, &models.AuditLog{
			ID:           log.ID,
			Resource:     log.Resource,
			ResourceType: log.ResourceType,
			Username:     log.Username,
			Operation:    log.Operation,
			OpTime:       strfmt.DateTime(log.OpTime),
		})
	}
	return operation.NewListAuditLogsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(auditLogs)
}
