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
	beegorm "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/audit"
	am "github.com/goharbor/harbor/src/pkg/audit/model"
)

// Handler - audit log handler
type Handler struct {
}

// AuditResolver - interface to resolve to AuditLog
type AuditResolver interface {
	ResolveToAuditLog() (*am.AuditLog, error)
}

// Handle ...
func (h *Handler) Handle(value interface{}) error {
	ctx := orm.NewContext(context.Background(), beegorm.NewOrm())
	var auditLog *am.AuditLog
	switch v := value.(type) {
	case *event.PushArtifactEvent, *event.PullArtifactEvent, *event.DeleteArtifactEvent,
		*event.DeleteRepositoryEvent, *event.CreateProjectEvent, *event.DeleteProjectEvent,
		*event.DeleteTagEvent, *event.CreateTagEvent:
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
		_, err := audit.Mgr.Create(ctx, auditLog)
		if err != nil {
			log.Debugf("add audit log err: %v", err)
		}
	}
	return nil
}

// IsStateful ...
func (h *Handler) IsStateful() bool {
	return false
}
