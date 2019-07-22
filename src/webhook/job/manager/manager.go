package manager

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/webhook/job"
)

// DefaultManager ..
type DefaultManager struct {
}

// NewDefaultManager ...
func NewDefaultManager() job.Manager {
	return &DefaultManager{}
}

// Create ...
func (d *DefaultManager) Create(job *models.WebhookJob) (int64, error) {
	return dao.AddWebhookJob(job)
}

// List ...
func (d *DefaultManager) List(query ...*models.WebhookJobQuery) (int64, []*models.WebhookJob, error) {
	total, err := dao.GetTotalCountOfWebhookJobs(query...)
	if err != nil {
		return 0, nil, err
	}

	executions, err := dao.GetWebhookJobs(query...)
	if err != nil {
		return 0, nil, err
	}
	return total, executions, nil
}

// Update ...
func (d *DefaultManager) Update(job *models.WebhookJob, props ...string) error {
	n, err := dao.UpdateWebhookJob(job, props...)
	if err != nil {
		return err
	}

	if n == 0 {
		return fmt.Errorf("execution %d not found", job.ID)
	}
	return nil
}

// UpdateJobStatus ...
func (d *DefaultManager) UpdateJobStatus(jobID int64, status string, statusCondition ...string) error {
	n, err := dao.UpdateWebhookJobStatus(jobID, status, statusCondition...)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("Update webhook job %d status -> %s failed ", jobID, status)
	}
	return nil
}

// ListLastTriggerInfos ...
func (d *DefaultManager) ListLastTriggerInfos(policyID int64) ([]*models.LastTriggerInfo, error) {
	return dao.GetLastTriggerInfosGroupByHookType(policyID)
}
