package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// UpdateWebhookExecution update webhook execution
func UpdateWebhookExecution(execution *models.WebhookExecution, props ...string) (int64, error) {
	if execution.ID == 0 {
		return 0, fmt.Errorf("webhook execution ID is empty")
	}

	o := GetOrmer()
	return o.Update(execution, props...)
}

//AddWebhookExecution insert new webhook execution to DB
func AddWebhookExecution(execution *models.WebhookExecution) (int64, error) {
	o := GetOrmer()
	if len(execution.Status) == 0 {
		execution.Status = models.JobPending
	}
	return o.Insert(execution)
}

// GetWebhookExecution ...
func GetWebhookExecution(id int64) (*models.WebhookExecution, error) {
	o := GetOrmer()
	j := &models.WebhookExecution{
		ID: id,
	}
	err := o.Read(j)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return j, nil
}

// GetTotalCountOfWebhookExecutions ...
func GetTotalCountOfWebhookExecutions(query ...*models.WebhookExecutionQuery) (int64, error) {
	qs := webhookExecutionQueryConditions(query...)
	return qs.Count()
}

// GetWebhookExecutions ...
func GetWebhookExecutions(query ...*models.WebhookExecutionQuery) ([]*models.WebhookExecution, error) {
	var jobs []*models.WebhookExecution

	qs := webhookExecutionQueryConditions(query...)
	if len(query) > 0 && query[0] != nil {
		qs = paginateForQuerySetter(qs, query[0].Page, query[0].Size)
	}

	qs = qs.OrderBy("-UpdateTime")

	_, err := qs.All(&jobs)
	return jobs, err
}

// GetLastTriggerInfosGroupByHookType get webhook executions info including hook type and last trigger time
func GetLastTriggerInfosGroupByHookType() ([]*models.LastTriggerInfo, error) {
	o := GetOrmer()
	sql := `select hook_type, max(creation_time) from webhook_execution group by hook_type`

	ltInfo := []*models.LastTriggerInfo{}
	_, err := o.Raw(sql).QueryRows(&ltInfo)
	if err != nil {
		log.Errorf("query last trigger info group by hook type failed: %v", err)
		return nil, err
	}

	return ltInfo, nil
}

// DeleteWebhookExecution ...
func DeleteWebhookExecution(id int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.WebhookExecution{ID: id})
	return err
}

// DeleteAllWebhookExecutions ...
func DeleteAllWebhookExecutions(policyID int64) (int64, error) {
	o := GetOrmer()
	return o.Delete(&models.WebhookExecution{PolicyID: policyID})
}

func webhookExecutionQueryConditions(query ...*models.WebhookExecutionQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(&models.WebhookExecution{})
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
	if q.StartTime != nil {
		qs = qs.Filter("CreationTime__gte", q.StartTime)
	}
	if q.EndTime != nil {
		qs = qs.Filter("EndTime_lte", q.EndTime)
	}
	return qs
}
