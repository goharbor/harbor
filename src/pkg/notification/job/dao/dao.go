package dao

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/job/model"
)

// DAO defines the interface to access the robot data model
type DAO interface {
	// Create ...
	Create(ctx context.Context, n *model.Job) (int64, error)

	// Update ...
	Update(ctx context.Context, n *model.Job, props ...string) error

	// Get ...
	Get(ctx context.Context, id int64) (*model.Job, error)

	// Count ...
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// List ...
	List(ctx context.Context, query *q.Query) ([]*model.Job, error)

	// Delete ...
	Delete(ctx context.Context, id int64) error

	// GetLastTriggerJobsGroupByEventType ...
	GetLastTriggerJobsGroupByEventType(ctx context.Context, policyID int64) ([]*model.Job, error)

	// DeleteByPolicyID
	DeleteByPolicyID(ctx context.Context, policyID int64) error
}

// New creates a default implementation for Dao
func New() DAO {
	return &dao{}
}

type dao struct{}

// UpdateNotificationJob update notification job
func (d *dao) Update(ctx context.Context, job *model.Job, props ...string) error {
	if job == nil {
		return errors.New("nil job")
	}

	if job.ID == 0 {
		return fmt.Errorf("notification job ID is empty")
	}

	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(job, props...)
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("notification %d not found", job.ID)
	}
	if err != nil {
		return err
	}
	return nil
}

// Create insert new notification job to DB
func (d *dao) Create(ctx context.Context, job *model.Job) (int64, error) {
	if job == nil {
		return 0, errors.New("nil job")
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	return ormer.Insert(job)
}

// Get ...
func (d *dao) Get(ctx context.Context, id int64) (*model.Job, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	j := &model.Job{
		ID: id,
	}
	if err := ormer.Read(j); err != nil {
		if e := orm.AsNotFoundError(err, "notificationJob %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return j, nil
}

// Count ...
func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &model.Job{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

// List ...
func (d *dao) List(ctx context.Context, query *q.Query) ([]*model.Job, error) {
	jobs := []*model.Job{}

	qs, err := orm.QuerySetter(ctx, &model.Job{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&jobs); err != nil {
		return nil, err
	}
	return jobs, nil
}

// GetLastTriggerJobsGroupByEventType get notification jobs info of policy, including event type and last trigger time
func (d *dao) GetLastTriggerJobsGroupByEventType(ctx context.Context, policyID int64) ([]*model.Job, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	// get jobs last triggered(created) group by event_type. postgres group by usage reference:
	// https://stackoverflow.com/questions/13325583/postgresql-max-and-group-by
	sql := `select distinct on (event_type) event_type, id, creation_time, status, notify_type, job_uuid, update_time, 
			creation_time, job_detail from notification_job where policy_id = ? 
			order by event_type, id desc, creation_time, status, notify_type, job_uuid, update_time, creation_time, job_detail`

	jobs := []*model.Job{}
	_, err = ormer.Raw(sql, policyID).QueryRows(&jobs)
	if err != nil {
		log.Errorf("query last trigger info group by event type failed: %v", err)
		return nil, err
	}

	return jobs, nil
}

func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.Job{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("notificationJob %d not found", id)
	}
	return nil
}

// DeleteByPolicyID ...
func (d *dao) DeleteByPolicyID(ctx context.Context, policyID int64) error {
	qs, err := orm.QuerySetter(ctx, &model.Job{}, &q.Query{
		Keywords: map[string]interface{}{
			"policy_id": policyID,
		},
	})
	if err != nil {
		return err
	}
	n, err := qs.Delete()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("notificationJob %d not found", policyID)
	}
	return nil
}
