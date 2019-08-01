package notification

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/core/notifier/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
)

// HTTPHandler handles notification http topic and start the hook processing
type HTTPHandler struct {
}

// Handle handles http event
func (h *HTTPHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("HTTPHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return errors.New("invalid notification http event")
	}

	return h.process(event)
}

// IsStateful ...
func (h *HTTPHandler) IsStateful() bool {
	return false
}

func (h *HTTPHandler) process(event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	j.Name = job.WebhookJob

	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal from payload %v failed: %v", event.Payload, err)
	}

	j.Parameters = map[string]interface{}{
		"payload": string(payload),
		"address": event.Target.Address,
		// Users can define a secret in http statement in notification(webhook) policy. So it will be sent in header in http request.
		// The format will be like this:
		// Authorization: 'Secret eyJ0eXAiOiJKV1QiLCJhbGciOi'
		"auth_header":      "Secret " + event.Target.AuthHeader,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(event, j)
}
