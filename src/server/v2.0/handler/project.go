package handler

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/pkg/audit"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project"
)

func newProjectAPI() *projectAPI {
	return &projectAPI{
		auditMgr: audit.Mgr,
		proCtl:   project.Ctl,
	}
}

type projectAPI struct {
	BaseAPI
	auditMgr audit.Manager
	proCtl   project.Controller
}

func (a *projectAPI) GetLogs(ctx context.Context, params operation.GetLogsParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceLog); err != nil {
		return a.SendError(ctx, err)
	}
	pro, err := a.proCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query, err := a.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query.Keywords["ProjectID"] = pro.ProjectID

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
	return operation.NewGetLogsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(auditLogs)
}
