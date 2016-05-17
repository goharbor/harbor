package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/models"
)

func AddRepTarget(target models.RepTarget) (int64, error) {
	o := orm.NewOrm()
	return o.Insert(&target)
}
func GetRepTarget(id int64) (*models.RepTarget, error) {
	o := orm.NewOrm()
	t := models.RepTarget{ID: id}
	err := o.Read(&t)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &t, err
}
func DeleteRepTarget(id int64) error {
	o := orm.NewOrm()
	_, err := o.Delete(&models.RepTarget{ID: id})
	return err
}

func AddRepPolicy(policy models.RepPolicy) (int64, error) {
	o := orm.NewOrm()
	sqlTpl := `insert into replication_policy (name, project_id, target_id, enabled, description, cron_str, start_time, creation_time, update_time ) values (?, ?, ?, ?, ?, ?, %s, NOW(), NOW())`
	var sql string
	if policy.Enabled == 1 {
		sql = fmt.Sprintf(sqlTpl, "NOW()")
	} else {
		sql = fmt.Sprintf(sqlTpl, "NULL")
	}
	p, err := o.Raw(sql).Prepare()
	if err != nil {
		return 0, err
	}
	r, err := p.Exec(policy.Name, policy.ProjectID, policy.TargetID, policy.Enabled, policy.Description, policy.CronStr)
	if err != nil {
		return 0, err
	}
	id, err := r.LastInsertId()
	return id, err
}
func GetRepPolicy(id int64) (*models.RepPolicy, error) {
	o := orm.NewOrm()
	p := models.RepPolicy{ID: id}
	err := o.Read(&p)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &p, err
}
func GetRepPolicyByProject(projectID int64) ([]*models.RepPolicy, error) {
	var res []*models.RepPolicy
	o := orm.NewOrm()
	_, err := o.QueryTable("replication_policy").Filter("project_id", projectID).All(&res)
	return res, err
}
func DeleteRepPolicy(id int64) error {
	o := orm.NewOrm()
	_, err := o.Delete(&models.RepPolicy{ID: id})
	return err
}
func updateRepPolicyEnablement(id int64, enabled int) error {
	o := orm.NewOrm()
	p := models.RepPolicy{
		ID:      id,
		Enabled: enabled}
	num, err := o.Update(&p, "Enabled")
	if num == 0 {
		err = fmt.Errorf("Failed to update replication policy with id: %d", id)
	}
	return err
}
func EnableRepPolicy(id int64) error {
	return updateRepPolicyEnablement(id, 1)
}

func DisableRepPolicy(id int64) error {
	return updateRepPolicyEnablement(id, 0)
}

func AddRepJob(job models.RepJob) (int64, error) {
	o := orm.NewOrm()
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	return o.Insert(&job)
}
func GetRepJob(id int64) (*models.RepJob, error) {
	o := orm.NewOrm()
	j := models.RepJob{ID: id}
	err := o.Read(&j)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &j, err
}
func GetRepJobByPolicy(policyID int64) ([]*models.RepJob, error) {
	o := orm.NewOrm()
	var res []*models.RepJob
	_, err := o.QueryTable("replication_job").Filter("policy_id", policyID).All(&res)
	return res, err
}
func DeleteRepJob(id int64) error {
	o := orm.NewOrm()
	_, err := o.Delete(&models.RepJob{ID: id})
	return err
}
func UpdateRepJobStatus(id int64, status string) error {
	o := orm.NewOrm()
	j := models.RepJob{
		ID:     id,
		Status: status,
	}
	num, err := o.Update(&j, "Status")
	if num == 0 {
		err = fmt.Errorf("Failed to update replication job with id: %d %s", id, err.Error())
	}
	return err
}
