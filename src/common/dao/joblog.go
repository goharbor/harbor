package dao

import (
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"time"
)

// CreateOrUpdateJobLog ...
func CreateOrUpdateJobLog(log *models.JobLog) (int64, error) {
	o := GetOrmer()
	count, err := o.InsertOrUpdate(log, "job_uuid")
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetJobLog ...
func GetJobLog(uuid string) (*models.JobLog, error) {
	o := GetOrmer()
	jl := models.JobLog{UUID: uuid}
	err := o.Read(&jl, "UUID")
	if err == orm.ErrNoRows {
		return nil, err
	}
	return &jl, nil
}

// DeleteJobLogsBefore ...
func DeleteJobLogsBefore(t time.Time) (int64, error) {
	o := GetOrmer()
	sql := `delete from job_log where creation_time < ?`
	res, err := o.Raw(sql, t).Exec()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
