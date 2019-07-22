package webhook

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/webhook"
	"github.com/goharbor/harbor/src/webhook/hook"
)

// HTTPScheduler ...
type HTTPScheduler struct {
}

// Handle ...
func (h *HTTPScheduler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("HttpScheduler cannot handle nil value")
	}

	item, ok := value.(*hook.ScheduleItem)
	if !ok || item == nil {
		return errors.New("invalid webhook http schedule item")
	}

	return h.process(item)
}

// IsStateful ...
func (h *HTTPScheduler) IsStateful() bool {
	return false
}

func (h *HTTPScheduler) process(item *hook.ScheduleItem) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	j.Name = job.WebhookHTTPJob

	payload, err := json.Marshal(item.Payload)
	if err != nil {
		return fmt.Errorf("marshal from payload %v failed: %v", item.Payload, err)
	}

	j.Parameters = map[string]interface{}{
		"payload":          string(payload),
		"address":          item.Target.Address,
		"token":            item.Target.Token,
		"skip_cert_verify": item.Target.SkipCertVerify,
	}
	return webhook.HookManager.StartHook(item, j)
}
