package notification

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/pkg/errors"
)

// UpdateNotificationJob update notification job
func UpdateNotificationJob(job *models.NotificationJob, props ...string) (int64, error) {
	if job == nil {
		return 0, errors.New("nil job")
	}

	if job.ID == 0 {
		return 0, fmt.Errorf("notification job ID is empty")
	}

	o := dao.GetOrmer()
	return o.Update(job, props...)
}

// AddNotificationJob insert new notification job to DB
func AddNotificationJob(job *models.NotificationJob) (int64, error) {
	if job == nil {
		return 0, errors.New("nil job")
	}
	o := dao.GetOrmer()
	if len(job.Status) == 0 {
		job.Status = models.JobPending
	}
	return o.Insert(job)
}

// GetNotificationJob ...
func GetNotificationJob(id int64) (*models.NotificationJob, error) {
	o := dao.GetOrmer()
	j := &models.NotificationJob{
		ID: id,
	}
	err := o.Read(j)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return j, nil
}

// GetTotalCountOfNotificationJobs ...
func GetTotalCountOfNotificationJobs(query ...*models.NotificationJobQuery) (int64, error) {
	qs := notificationJobQueryConditions(query...)
	return qs.Count()
}

// GetNotificationJobs ...
func GetNotificationJobs(query ...*models.NotificationJobQuery) ([]*models.NotificationJob, error) {
	var jobs []*models.NotificationJob

	qs := notificationJobQueryConditions(query...)
	if len(query) > 0 && query[0] != nil {
		qs = dao.PaginateForQuerySetter(qs, query[0].Page, query[0].Size)
	}

	qs = qs.OrderBy("-UpdateTime")

	_, err := qs.All(&jobs)
	return jobs, err
}

// GetLastTriggerJobsGroupByEventType get notification jobs info of policy, including event type and last trigger time
func GetLastTriggerJobsGroupByEventType(policyID int64) ([]*models.NotificationJob, error) {
	o := dao.GetOrmer()
	// get jobs last triggered(created) group by event_type
	sql := `select distinct on (event_type) event_type, id, creation_time, status, notify_type, job_uuid, update_time, 
			creation_time, job_detail from notification_job where policy_id = ? 
			order by event_type, id desc, creation_time, status, notify_type, job_uuid, update_time, creation_time, job_detail`

	jobs := []*models.NotificationJob{}
	_, err := o.Raw(sql, policyID).QueryRows(&jobs)
	if err != nil {
		log.Errorf("query last trigger info group by event type failed: %v", err)
		return nil, err
	}

	return jobs, nil
}

// DeleteNotificationJob ...
func DeleteNotificationJob(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.NotificationJob{ID: id})
	return err
}

// DeleteAllNotificationJobsByPolicyID ...
func DeleteAllNotificationJobsByPolicyID(policyID int64) (int64, error) {
	o := dao.GetOrmer()
	return o.Delete(&models.NotificationJob{PolicyID: policyID}, "policy_id")
}

func notificationJobQueryConditions(query ...*models.NotificationJobQuery) orm.QuerySeter {
	qs := dao.GetOrmer().QueryTable(&models.NotificationJob{})
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
	if len(q.EventTypes) > 0 {
		qs = qs.Filter("EventType__in", q.EventTypes)
	}
	return qs
}
