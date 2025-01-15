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

package auditext

import (
	"context"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/q"
	auditV1 "github.com/goharbor/harbor/src/pkg/audit"
	"github.com/goharbor/harbor/src/pkg/auditext/dao"
	"github.com/goharbor/harbor/src/pkg/auditext/model"
)

// Mgr is the global audit log manager instance
var Mgr = New()

// Manager is used for audit log management
type Manager interface {
	// Count returns the total count of audit logs according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List audit logs according to the query
	List(ctx context.Context, query *q.Query) (audits []*model.AuditLogExt, err error)
	// Get the audit log specified by ID
	Get(ctx context.Context, id int64) (audit *model.AuditLogExt, err error)
	// Create the audit log
	Create(ctx context.Context, audit *model.AuditLogExt) (id int64, err error)
	// Delete the audit log specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Purge delete the audit log with retention hours
	Purge(ctx context.Context, retentionHour int, includeOperations []string, dryRun bool) (int64, error)
	// UpdateUsername Replace all log records username with its hash
	UpdateUsername(ctx context.Context, username string, replaceWith string) error
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{
		dao: dao.New(),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) UpdateUsername(ctx context.Context, username string, replaceWith string) error {
	return m.dao.UpdateUsername(ctx, username, replaceWith)
}

// Count ...
func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

// List ...
func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.AuditLogExt, error) {
	return m.dao.List(ctx, query)
}

// Get ...
func (m *manager) Get(ctx context.Context, id int64) (*model.AuditLogExt, error) {
	return m.dao.Get(ctx, id)
}

// Create ...
func (m *manager) Create(ctx context.Context, audit *model.AuditLogExt) (int64, error) {
	if len(config.AuditLogForwardEndpoint(ctx)) > 0 {
		auditV1.LogMgr.DefaultLogger(ctx).WithField("operator", audit.Username).
			WithField("time", audit.OpTime).WithField("resourceType", audit.ResourceType).
			Infof("action:%s, resource:%s", audit.Operation, audit.Resource)
	}
	if config.SkipAuditLogDatabase(ctx) {
		return 0, nil
	}
	return m.dao.Create(ctx, audit)
}

// Purge ...
func (m *manager) Purge(ctx context.Context, retentionHour int, includeOperations []string, dryRun bool) (int64, error) {
	return m.dao.Purge(ctx, retentionHour, includeOperations, dryRun)
}

// Delete ...
func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}
