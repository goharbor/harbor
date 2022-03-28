package dao

import (
	"context"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/notification/job/model"
)

// New creates a default implementation for Dao
func NewMysqlDao() DAO {
	return &mysqlDao{}
}

type mysqlDao struct {
	*dao
}

// GetLastTriggerJobsGroupByEventType get notification jobs info of policy, including event type and last trigger time
func (d *mysqlDao) GetLastTriggerJobsGroupByEventType(ctx context.Context, policyID int64) ([]*model.Job, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	// todo Yvonne more beauty
	setSQLMode := `set sql_mode="STRICT_TRANS_TABLES"`
	_, err = ormer.Raw(setSQLMode).Exec()
	if err != nil {
		log.Errorf("query last trigger info group by event type failed: %v", err)
		return nil, err
	}
	// get jobs last triggered(created) group by event_type.
	sql := `select event_type, id, creation_time, status, notify_type, job_uuid, update_time, 
			creation_time, job_detail from notification_job where policy_id = ? 
			group by event_type order by event_type, id desc, creation_time, status, notify_type, job_uuid, update_time, creation_time, job_detail`
	jobs := []*model.Job{}
	_, err = ormer.Raw(sql, policyID).QueryRows(&jobs)
	if err != nil {
		log.Errorf("query last trigger info group by event type failed: %v", err)
		return nil, err
	}

	return jobs, nil
}
