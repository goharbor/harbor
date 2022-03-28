package dao

import (
	"context"
	"strings"

	beegorm "github.com/beego/beego/orm"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
)

// NewMysqlDao ...
func NewMysqlDao() DAO {
	return &mysqlDao{}
}

type mysqlDao struct {
	*dao
}

// Purge delete expired audit log
func (*mysqlDao) Purge(ctx context.Context, retentionHour int, includeOperations []string, dryRun bool) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	if dryRun {
		return dryRunPurgeForMysql(ormer, retentionHour, includeOperations)
	}
	sql := "DELETE FROM audit_log WHERE op_time < date_sub(CURRENT_TIMESTAMP(6), interval ? * 1 hour) "
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

func dryRunPurgeForMysql(ormer beegorm.Ormer, retentionHour int, includeOperations []string) (int64, error) {
	sql := "SELECT count(1) cnt FROM audit_log WHERE op_time < date_sub(CURRENT_TIMESTAMP(6), interval ? * 1 hour) "
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
