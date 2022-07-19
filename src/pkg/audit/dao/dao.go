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
	"strings"

	beegorm "github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/common/rbac"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/audit/model"
)

// DAO is the data access object for audit log
type DAO interface {
	// Create the audit log
	Create(ctx context.Context, access *model.AuditLog) (id int64, err error)
	// Count returns the total count of audit logs according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List audit logs according to the query
	List(ctx context.Context, query *q.Query) (access []*model.AuditLog, err error)
	// Get the audit log specified by ID
	Get(ctx context.Context, id int64) (access *model.AuditLog, err error)
	// Delete the audit log specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Purge the audit log
	Purge(ctx context.Context, retentionHour int, includeOperations []string, dryRun bool) (int64, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

var allowedMaps = map[string]interface{}{
	strings.ToLower(rbac.ActionPull.String()):   struct{}{},
	strings.ToLower(rbac.ActionCreate.String()): struct{}{},
	strings.ToLower(rbac.ActionDelete.String()): struct{}{},
}

type dao struct{}

// Purge delete expired audit log
func (*dao) Purge(ctx context.Context, retentionHour int, includeOperations []string, dryRun bool) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	if dryRun {
		return dryRunPurge(ormer, retentionHour, includeOperations)
	}
	sql := "DELETE FROM audit_log WHERE op_time < NOW() - ? * interval '1 hour' "
	filterOps := permitOps(includeOperations)
	if len(filterOps) == 0 {
		log.Infof("no operation selected, skip to purge audit log")
		return 0, nil
	}
	sql = sql + "AND lower(operation) IN ('" + strings.Join(filterOps, "','") + "')"
	log.Debugf("the sql is %v", sql)

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

func dryRunPurge(ormer beegorm.Ormer, retentionHour int, includeOperations []string) (int64, error) {
	sql := "SELECT count(1) cnt FROM audit_log WHERE op_time < NOW() - ? * interval '1 hour' "
	filterOps := permitOps(includeOperations)
	if len(filterOps) == 0 {
		log.Infof("[DRYRUN]no operation selected, skip to purge audit log")
		return 0, nil
	}
	sql = sql + "AND lower(operation) IN ('" + strings.Join(filterOps, "','") + "')"
	log.Debugf("the sql is %v", sql)

	var cnt int64
	err := ormer.Raw(sql, retentionHour).QueryRow(&cnt)
	if err != nil {
		log.Errorf("failed to dry run purge audit log, error %v", err)
		return 0, err
	}
	log.Infof("[DRYRUN]purged %d audit logs in the database", cnt)
	return cnt, nil
}

// permitOps filter not allowed operation, if no operation specified, purge pull operation
func permitOps(includeOperations []string) []string {
	if includeOperations == nil {
		return nil
	}
	var filterOps []string
	for _, ops := range includeOperations {
		ops := strings.ToLower(ops)
		if _, exist := allowedMaps[ops]; exist {
			filterOps = append(filterOps, ops)
		}
	}
	return filterOps
}

// Count ...
func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &model.AuditLog{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

// List ...
func (d *dao) List(ctx context.Context, query *q.Query) ([]*model.AuditLog, error) {
	audit := []*model.AuditLog{}
	qs, err := orm.QuerySetter(ctx, &model.AuditLog{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&audit); err != nil {
		return nil, err
	}
	return audit, nil
}

// Get ...
func (d *dao) Get(ctx context.Context, id int64) (*model.AuditLog, error) {
	audit := &model.AuditLog{
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
func (d *dao) Create(ctx context.Context, audit *model.AuditLog) (int64, error) {
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
	n, err := ormer.Delete(&model.AuditLog{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("access %d not found", id)
	}
	return nil
}
