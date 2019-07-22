package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// UpdateWebhookJob update webhook job
func UpdateWebhookJob(job *models.WebhookJob, props ...string) (int64, error) {
	if job.ID == 0 {
		return 0, fmt.Errorf("webhook job ID is empty")
	}

	o := GetOrmer()
	return o.Update(job, props...)
}

// UpdateWebhookJobStatus ...
func UpdateWebhookJobStatus(id int64, status string, statusCondition ...string) (int64, error) {
	qs := GetOrmer().QueryTable(&models.WebhookJob{}).Filter("id", id)
	if len(statusCondition) > 0 {
		qs = qs.Filter("status", statusCondition[0])
	}
	params := orm.Params{
		"status": status,
	}

	n, err := qs.Update(params)
	if err != nil {
		return 0, err
	}
	log.Debugf("update webhook job status %d: -> %s", id, status)
	return n, err
}

//AddWebhookJob insert new webhook job to DB
func AddWebhookJob(job *models.WebhookJob) (int64, error) {
	o := GetOrmer()
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	return o.Insert(job)
}

// GetWebhookJob ...
func GetWebhookJob(id int64) (*models.WebhookJob, error) {
	o := GetOrmer()
	j := &models.WebhookJob{
		ID: id,
	}
	err := o.Read(j)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return j, nil
}

// GetTotalCountOfWebhookJobs ...
func GetTotalCountOfWebhookJobs(query ...*models.WebhookJobQuery) (int64, error) {
	qs := webhookJobQueryConditions(query...)
	return qs.Count()
}

// GetWebhookJobs ...
func GetWebhookJobs(query ...*models.WebhookJobQuery) ([]*models.WebhookJob, error) {
	var jobs []*models.WebhookJob

	qs := webhookJobQueryConditions(query...)
	if len(query) > 0 && query[0] != nil {
		qs = paginateForQuerySetter(qs, query[0].Page, query[0].Size)
	}

	qs = qs.OrderBy("-UpdateTime")

	_, err := qs.All(&jobs)
	return jobs, err
}

// GetLastTriggerInfosGroupByHookType get webhook jobs info of policy, including hook type and last trigger time
func GetLastTriggerInfosGroupByHookType(policyID int64) ([]*models.LastTriggerInfo, error) {
	o := GetOrmer()
	sql := `select hook_type, max(creation_time) as ct from webhook_job where policy_id = ? group by hook_type`

	ltInfo := []*models.LastTriggerInfo{}
	_, err := o.Raw(sql, policyID).QueryRows(&ltInfo)
	if err != nil {
		log.Errorf("query last trigger info group by hook type failed: %v", err)
		return nil, err
	}

	return ltInfo, nil
}

// DeleteWebhookJob ...
func DeleteWebhookJob(id int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.WebhookJob{ID: id})
	return err
}

// DeleteAllWebhookJobs ...
func DeleteAllWebhookJobs(policyID int64) (int64, error) {
	o := GetOrmer()
	return o.Delete(&models.WebhookJob{PolicyID: policyID})
}

func webhookJobQueryConditions(query ...*models.WebhookJobQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(&models.WebhookJob{})
	if len(query) == 0 || query[0] == nil {
		return qs
	}

	q := query[0]
	if q.PolicyID != 0 {
		qs = qs.Filter("PolicyID", q.PolicyID)
	}
	if len(q.Statuses) > 0 {
		qs = qs.Filter("Status__in", q.Statuses)
	}
	if len(q.HookTypes) > 0 {
		qs = qs.Filter("HookType__in", q.HookTypes)
	}
	return qs
}
