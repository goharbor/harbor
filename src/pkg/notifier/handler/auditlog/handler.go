package auditlog

import (
	beegoorm "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/audit"
	am "github.com/goharbor/harbor/src/pkg/audit/model"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// Handler - audit log handler
type Handler struct {
	AuditLogMgr audit.Manager
}

// AuditResolver - interface to resolve to AuditLog
type AuditResolver interface {
	ResolveToAuditLog() (*am.AuditLog, error)
}

// AuditHandler ...
var AuditHandler = Handler{AuditLogMgr: audit.Mgr}

// Handle ...
func (h *Handler) Handle(value interface{}) error {
	ctx := orm.NewContext(nil, beegoorm.NewOrm())
	var auditLog *am.AuditLog
	switch v := value.(type) {
	case *model.ProjectEvent, *model.RepositoryEvent, *model.ArtifactEvent, *model.TagEvent:
		resolver := value.(AuditResolver)
		al, err := resolver.ResolveToAuditLog()
		if err != nil {
			log.Errorf("failed to handler event %v", err)
			return err
		}
		auditLog = al
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	if auditLog != nil {
		h.AuditLogMgr.Create(ctx, auditLog)
	}
	return nil
}

// IsStateful ...
func (h *Handler) IsStateful() bool {
	return false
}
