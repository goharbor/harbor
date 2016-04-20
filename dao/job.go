package dao

import (
	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

const (
	JobPending  string = "pending"
	JobRunning  string = "running"
	JobError    string = "error"
	JobStopped  string = "stopped"
	JobFinished string = "finished"
)

func AddJob(entry models.JobEntry) (int64, error) {

	sql := `insert into job (job_type, status, options, parms, cron_str, creation_time, update_time) values (?,"pending",?,?,?,NOW(),NOW())`
	o := orm.NewOrm()
	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return 0, err
	}
	r, err := p.Exec(entry.Type, entry.OptionsStr, entry.ParmsStr, entry.CronStr)
	if err != nil {
		return 0, err
	}
	id, err := r.LastInsertId()
	return id, err
}

func AddJobLog(id int64, level string, message string) error {
	sql := `insert into job_log (job_id, level, message, creation_time, update_time) values (?, ?, ?, NOW(), NOW())`
	log.Debugf("trying to add a log for job:%d", id)
	o := orm.NewOrm()
	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return err
	}
	_, err = p.Exec(id, level, message)
	return err
}

func UpdateJobStatus(id int64, status string) error {
	o := orm.NewOrm()
	sql := "update job set status=?, update_time=NOW() where job_id=?"
	_, err := o.Raw(sql, status, id).Exec()
	return err
}

func ListJobs() ([]models.JobEntry, error) {
	o := orm.NewOrm()
	sql := `select j.job_id, j.job_type, j.status, j.enabled, j.creation_time, j.update_time from job j`
	var res []models.JobEntry
	_, err := o.Raw(sql).QueryRows(&res)
	if err != nil {
		return nil, err
	}
	return res, err
}

func GetJob(id int64) (*models.JobEntry, error) {
	o := orm.NewOrm()
	sql := `select j.job_id, j.job_type, j.status, j.enabled, j.creation_time, j.update_time from job j where j.job_id = ?`
	var res []models.JobEntry
	p := make([]interface{}, 1)
	p = append(p, id)
	n, err := o.Raw(sql, p).QueryRows(&res)
	if n == 0 {
		return nil, err
	}
	return &res[0], err
}

func GetJobLogs(jobID int64) ([]models.JobLog, error) {
	o := orm.NewOrm()
	var res []models.JobLog
	p := make([]interface{}, 1)
	p = append(p, jobID)
	sql := `select l.log_id, l.job_id, l.level, l.message, l.creation_time, l.update_time from job_log l where l.job_id = ?`
	_, err := o.Raw(sql, p).QueryRows(&res)
	return res, err
}
