package hook

import (
	"encoding/json"
	"fmt"
	"time"

	cJob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	cModels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/webhook/config"
	"github.com/goharbor/harbor/src/webhook/job"
	"github.com/goharbor/harbor/src/webhook/job/manager"
)

// Manager send hook
type Manager interface {
	StartHook(item *ScheduleItem, data *models.JobData) error
}

// DefaultManager ...
type DefaultManager struct {
	jobMgr job.Manager
	client cJob.Client
}

// NewHookManager ...
func NewHookManager() *DefaultManager {
	return &DefaultManager{
		jobMgr: manager.NewDefaultManager(),
		client: utils.GetJobServiceClient(),
	}
}

// ScheduleItem is an item that can be scheduled
type ScheduleItem struct {
	HookType string
	PolicyID int64
	Target   *cModels.HookTarget
	Payload  interface{}
}

// StartHook ...
func (hm *DefaultManager) StartHook(item *ScheduleItem, data *models.JobData) error {
	payload, err := json.Marshal(item.Payload)
	if err != nil {
		return err
	}

	t := time.Now()
	id, err := hm.jobMgr.Create(&cModels.WebhookJob{
		PolicyID:     item.PolicyID,
		HookType:     item.HookType,
		NotifyType:   item.Target.Type,
		Status:       cModels.JobPending,
		CreationTime: t,
		UpdateTime:   t,
		JobDetail:    string(payload),
	})
	if err != nil {
		return fmt.Errorf("failed to create the job record for webhook based on policy %d: %v", item.PolicyID, err)
	}
	statusHookURL := fmt.Sprintf("%s/service/notifications/jobs/webhook/%d", config.Config.CoreURL, id)
	data.StatusHook = statusHookURL

	log.Debugf("created a webhook job %d for the policy %d", id, item.PolicyID)

	// submit hook job to jobservice
	jobUUID, err := hm.client.SubmitJob(data)
	if err != nil {
		log.Errorf("failed to process the webhook event: %v", err)
		e := hm.jobMgr.UpdateJobStatus(id, cModels.JobError)
		if e != nil {
			log.Errorf("failed to update the webhook job status %d: %v", id, e)
		}
		return err
	}
	err = hm.jobMgr.UpdateJobStatus(id, cModels.JobRunning, cModels.JobPending)
	if err != nil {
		log.Errorf("failed to update the webhook job status %d: %v", id, err)
		return err
	}

	if err = hm.jobMgr.Update(&cModels.WebhookJob{
		ID:   id,
		UUID: jobUUID,
	}, "UUID"); err != nil {
		log.Errorf("failed to update the webhook job %d: %v", id, err)
		return err
	}
	return nil
}
