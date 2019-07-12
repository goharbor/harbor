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
	"github.com/goharbor/harbor/src/webhook/execution"
	"github.com/goharbor/harbor/src/webhook/execution/manager"
	"github.com/goharbor/harbor/src/webhook/model"
)

// Manager send hook
type Manager interface {
	StartHook(item *ScheduleItem, data *models.JobData) error
}

// DefaultManager ...
type DefaultManager struct {
	execMgr execution.Manager
	client  cJob.Client
}

// NewHookManager ...
func NewHookManager() *DefaultManager {
	return &DefaultManager{
		execMgr: manager.NewDefaultManager(),
		client:  utils.GetJobServiceClient(),
	}
}

// ScheduleItem is an item that can be scheduled
type ScheduleItem struct {
	PolicyID int64
	Target   *model.HookTarget
	Payload  interface{}
	IsTest   bool
}

// StartHook ...
func (hm *DefaultManager) StartHook(item *ScheduleItem, data *models.JobData) error {
	payload, err := json.Marshal(item.Payload)
	if err != nil {
		return err
	}

	t := time.Now()
	id, err := hm.execMgr.Create(&cModels.WebhookExecution{
		PolicyID:     item.PolicyID,
		HookType:     item.Target.Type,
		Status:       cModels.JobPending,
		CreationTime: t,
		UpdateTime:   t,
		JobDetail:    string(payload),
	})
	if err != nil {
		return fmt.Errorf("failed to create the execution record for webhook based on policy %d: %v", item.PolicyID, err)
	}
	statusHookURL := fmt.Sprintf("%s/service/notifications/jobs/webhook/%d", config.Config.CoreURL, id)
	data.StatusHook = statusHookURL

	log.Debugf("created a webhook execution %d for the policy %d", id, item.PolicyID)

	// submit hook job to jobservice
	go func() {
		whExecution := &cModels.WebhookExecution{
			ID:         id,
			UpdateTime: time.Now(),
		}

		jobUUID, err := hm.client.SubmitJob(data)
		if err != nil {
			log.Errorf("failed to process the webhook event: %v", err)
			e := hm.execMgr.Update(whExecution, "Status", "UpdateTime")
			if e != nil {
				log.Errorf("failed to update the webhook execution %d: %v", id, e)
			}
			return
		}
		whExecution.UUID = jobUUID
		e := hm.execMgr.Update(whExecution, "UUID", "UpdateTime")
		if e != nil {
			log.Errorf("failed to update the webhook execution %d: %v", id, e)
		}
	}()
	return nil
}
