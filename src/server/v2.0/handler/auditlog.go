package handler

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/pkg/audit"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/auditlog"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/auditlog"
)

func newAuditLogAPI() *auditlogAPI {
	return &auditlogAPI{
		auditMgr: audit.Mgr,
	}
}

type auditlogAPI struct {
	BaseAPI
	auditMgr audit.Manager
}

func (a *auditlogAPI) ListAuditLogs(ctx context.Context, params auditlog.ListAuditLogsParams) middleware.Responder {
	// ToDo enable permission check
	// if !a.HasPermission(ctx, rbac.ActionList, rbac.ResourceLog) {
	//	return a.SendError(ctx, ierror.ForbiddenError(nil))
	// }
	keywords := make(map[string]interface{})
	query := &q.Query{
		Keywords: keywords,
	}
	// TODO support fuzzy match and start end time
	if params.Username != nil {
		query.Keywords["Username"] = *(params.Username)
	}
	if params.Operation != nil {
		query.Keywords["Operation"] = *(params.Operation)
	}
	if params.Resource != nil {
		query.Keywords["Resource"] = *(params.Resource)
	}
	if params.ResourceType != nil {
		query.Keywords["ResourceType"] = *(params.ResourceType)
	}
	if params.Page != nil {
		query.PageNumber = *(params.Page)
	}
	if params.PageSize != nil {
		query.PageSize = *(params.PageSize)
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
			OpTime:       log.OpTime.String(),
		})
	}
	return operation.NewListAuditLogsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(auditLogs)
}
