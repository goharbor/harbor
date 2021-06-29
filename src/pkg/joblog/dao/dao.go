package dao

import (
	"context"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/joblog/models"
	"time"
)

// DAO is the data access object for job log
type DAO interface {
	// Create the job log
	Create(ctx context.Context, jobLog *models.JobLog) (id int64, err error)
	// Get the job log specified by UUID
	Get(ctx context.Context, uuid string) (jobLog *models.JobLog, err error)
	// DeleteBefore the job log specified by time
	DeleteBefore(ctx context.Context, t time.Time) (id int64, err error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

// Create ...
func (d *dao) Create(ctx context.Context, jobLog *models.JobLog) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	count, err := ormer.InsertOrUpdate(jobLog, "job_uuid")
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Get ...
func (d *dao) Get(ctx context.Context, uuid string) (jobLog *models.JobLog, err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	jl := models.JobLog{UUID: uuid}
	err = ormer.Read(&jl, "UUID")
	if e := orm.AsNotFoundError(err, "no job log founded"); e != nil {
		log.Warningf("no job log founded. Query condition, uuid: %s, err: %v", uuid, e)
		return nil, err
	}
	return &jl, nil
}

// DeleteBefore ...
func (d *dao) DeleteBefore(ctx context.Context, t time.Time) (id int64, err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	sql := `delete from job_log where creation_time < ?`
	res, err := ormer.Raw(sql, t).Exec()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
