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
	"fmt"

	"github.com/goharbor/harbor/src/controller/event"
	evtModel "github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/auditext"
	am "github.com/goharbor/harbor/src/pkg/auditext/model"
)

// Handler - audit log handler
type Handler struct {
}

// AuditResolver - interface to resolve to AuditLog
type AuditResolver interface {
	ResolveToAuditLog() (*am.AuditLogExt, error)
}

// Name ...
func (h *Handler) Name() string {
	return "AuditLog"
}

// Handle ...
func (h *Handler) Handle(ctx context.Context, value interface{}) error {
	var addAuditLog bool
	switch v := value.(type) {
	case *event.PushArtifactEvent, *event.DeleteArtifactEvent,
		*event.DeleteRepositoryEvent, *event.CreateProjectEvent, *event.DeleteProjectEvent,
		*event.DeleteTagEvent, *event.CreateTagEvent,
		*event.CreateRobotEvent, *event.DeleteRobotEvent, *evtModel.CommonEvent:
		addAuditLog = true
	case *event.PullArtifactEvent:
		addAuditLog = !config.PullAuditLogDisable(ctx)
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}

	if addAuditLog {
		resolver := value.(AuditResolver)
		auditLog, err := resolver.ResolveToAuditLog()
		if err != nil {
			log.Errorf("failed to handler event %v", err)
			return err
		}
		if auditLog != nil && config.AuditLogEventEnabled(ctx, fmt.Sprintf("%v_%v", auditLog.Operation, auditLog.ResourceType)) {
			_, err := auditext.Mgr.Create(ctx, auditLog)
			if err != nil {
				log.Debugf("add audit log err: %v", err)
			}
		}
	}
	return nil
}

// IsStateful ...
func (h *Handler) IsStateful() bool {
	return false
}
