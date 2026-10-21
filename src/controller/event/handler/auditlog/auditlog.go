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
	"strings"

	"github.com/goharbor/harbor/src/controller/event"
	evtModel "github.com/goharbor/harbor/src/controller/event/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	atrJob "github.com/goharbor/harbor/src/jobservice/job/impl/audittagresolution"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/auditext"
	am "github.com/goharbor/harbor/src/pkg/auditext/model"
	"github.com/goharbor/harbor/src/pkg/task"
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
func (h *Handler) Handle(ctx context.Context, value any) error {
	var addAuditLog bool
	var isPull bool
	switch v := value.(type) {
	case *event.PushArtifactEvent, *event.DeleteArtifactEvent,
		*event.DeleteRepositoryEvent, *event.CreateProjectEvent, *event.DeleteProjectEvent,
		*event.DeleteTagEvent, *event.CreateTagEvent,
		*event.CreateRobotEvent, *event.DeleteRobotEvent, *evtModel.CommonEvent:
		addAuditLog = true
	case *event.PullArtifactEvent:
		addAuditLog = !config.PullAuditLogDisable(ctx)
		isPull = true
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
			id, err := auditext.Mgr.Create(ctx, auditLog)
			if err != nil {
				log.Infof("add audit log err: %v", err)
			}
			// For digest-only pull entries, enqueue a background job to resolve tags.
			// The resource field uses "@" for digest references and ":" for tag references.
			if isPull && id > 0 && strings.Contains(auditLog.Resource, "@") {
				h.enqueueTagResolution(ctx, id, auditLog)
			}
		}
	}
	return nil
}

func (h *Handler) enqueueTagResolution(ctx context.Context, auditLogID int64, auditLog *am.AuditLogExt) {
	// Extract repository and digest from resource "repo@sha256:abc..."
	parts := strings.SplitN(auditLog.Resource, "@", 2)
	if len(parts) != 2 {
		return
	}
	repo, digest := parts[0], parts[1]

	params := map[string]any{
		atrJob.ParamAuditLogID: auditLogID,
		atrJob.ParamRepository: repo,
		atrJob.ParamDigest:     digest,
	}

	execID, err := task.ExecMgr.Create(ctx, job.AuditTagResolutionVendorType, -1, task.ExecutionTriggerEvent, params)
	if err != nil {
		log.Errorf("failed to create audit tag resolution execution: %v", err)
		return
	}
	_, err = task.Mgr.Create(ctx, execID, &task.Job{
		Name: job.AuditTagResolutionVendorType,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
	})
	if err != nil {
		log.Errorf("failed to create audit tag resolution task: %v", err)
	}
}

// IsStateful ...
func (h *Handler) IsStateful() bool {
	return false
}
