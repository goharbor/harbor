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

package dao

import (
	"context"
	"slices"
	"strings"

	beegorm "github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/auditext/model"
)

// DAO is the data access object for audit log
type DAO interface {
	// Create the audit log ext
	Create(ctx context.Context, access *model.AuditLogExt) (id int64, err error)
	// Count returns the total count of audit log ext according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List audit log ext according to the query
	List(ctx context.Context, query *q.Query) (access []*model.AuditLogExt, err error)
	// Get the audit log ext specified by ID
	Get(ctx context.Context, id int64) (access *model.AuditLogExt, err error)
	// Delete the audit log ext specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Purge the audit log ext
	Purge(ctx context.Context, retentionHour int, includeOperations []string, dryRun bool) (int64, error)
	// UpdateUsername replaces username in matched records
	UpdateUsername(ctx context.Context, username string, usernameReplace string) error
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) UpdateUsername(ctx context.Context, username string, usernameReplace string) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = o.Raw("UPDATE audit_log_ext SET username = ? WHERE username = ?", usernameReplace, username).Exec()
	return err
}

// Count ...
func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &model.AuditLogExt{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

// List ...
func (d *dao) List(ctx context.Context, query *q.Query) ([]*model.AuditLogExt, error) {
	audit := []*model.AuditLogExt{}
	qs, err := orm.QuerySetter(ctx, &model.AuditLogExt{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&audit); err != nil {
		return nil, err
	}
	return audit, nil
}

// Get ...
func (d *dao) Get(ctx context.Context, id int64) (*model.AuditLogExt, error) {
	audit := &model.AuditLogExt{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(audit); err != nil {
		if e := orm.AsNotFoundError(err, "audit %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return audit, nil
}

// Create ...
func (d *dao) Create(ctx context.Context, audit *model.AuditLogExt) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	// the max length of username in database is 255, replace the last
	// three charaters with "..." if the length is greater than 256
	if len(audit.Username) > 255 {
		audit.Username = audit.Username[:252] + "..."
	}
	id, err := ormer.Insert(audit)
	if err != nil {
		return 0, err
	}
	return id, err
}

// Delete ...
func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.AuditLogExt{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("access %d not found", id)
	}
	return nil
}

// Purge delete expired audit log ext
func (*dao) Purge(ctx context.Context, retentionHour int, includeEventTypes []string, dryRun bool) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	if dryRun {
		return dryRunPurge(ormer, retentionHour, includeEventTypes)
	}
	filterEvents := permitEventTypes(includeEventTypes)
	if len(filterEvents) == 0 {
		log.Infof("no operation selected, skip to purge audit log")
		return 0, nil
	}
	sql := "DELETE FROM audit_log_ext WHERE op_time < NOW() - ? * interval '1 hour' AND lower(operation || '_' || resource_type) IN ('" + strings.Join(filterEvents, "','") + "')"
	log.Debugf("purge audit logs raw sql: %v", sql)

	r, err := ormer.Raw(sql, retentionHour).Exec()
	if err != nil {
		log.Errorf("failed to purge audit log, error %v", err)
		return 0, err
	}
	delRows, rErr := r.RowsAffected()
	if rErr != nil {
		log.Errorf("failed to purge audit log, error %v", rErr)
		return 0, rErr
	}
	log.Infof("purged %d audit logs in the database", delRows)

	return delRows, err
}

func dryRunPurge(ormer beegorm.QueryExecutor, retentionHour int, includeEventTypes []string) (int64, error) {
	filterEvents := permitEventTypes(includeEventTypes)
	if len(filterEvents) == 0 {
		log.Infof("[DRYRUN]no operation selected, skip to purge audit log")
		return 0, nil
	}
	sql := "SELECT count(1) cnt FROM audit_log_ext WHERE op_time < NOW() - ? * interval '1 hour' AND lower(operation || '_' || resource_type) IN ('" + strings.Join(filterEvents, "','") + "')"
	log.Debugf("purge audit log count raw sql: %v", sql)

	var cnt int64
	err := ormer.Raw(sql, retentionHour).QueryRow(&cnt)
	if err != nil {
		log.Errorf("failed to dry run purge audit log, error %v", err)
		return 0, err
	}
	log.Infof("[DRYRUN]purged %d audit logs in the database", cnt)
	return cnt, nil
}

// permitEventTypes filter not allowed event type, if no event types specified, purge no operation, use this function to avoid SQL injection
func permitEventTypes(includeEventTypes []string) []string {
	if includeEventTypes == nil {
		return nil
	}
	var filterEvents []string
	for _, e := range includeEventTypes {
		event := strings.ToLower(e)
		if slices.Contains(model.EventTypes, event) {
			filterEvents = append(filterEvents, e)
		} else if event == model.OtherEvents { // include all other events
			filterEvents = append(filterEvents, model.OtherEventTypes...)
		}
	}
	return filterEvents
}
