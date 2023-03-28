package notification

import (
	"context"
	"encoding/json"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/formats"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// HTTPHandler preprocess http event data and start the hook processing
type HTTPHandler struct {
}

// Name ...
func (h *HTTPHandler) Name() string {
	return "HTTP"
}

// Handle handles http event
func (h *HTTPHandler) Handle(ctx context.Context, value interface{}) error {
	if value == nil {
		return errors.New("HTTPHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return errors.New("invalid notification http event")
	}
	return h.process(ctx, event)
}

// IsStateful ...
func (h *HTTPHandler) IsStateful() bool {
	return false
}

func (h *HTTPHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	j.Name = job.WebhookJobVendorType

	if event == nil || event.Payload == nil || event.Target == nil {
		return errors.Errorf("invalid event: %+v", event)
	}

	formatter, err := formats.GetFormatter(event.Target.PayloadFormat)
	if err != nil {
		return errors.Wrap(err, "error to get formatter")
	}

	header, payload, err := formatter.Format(ctx, event)
	if err != nil {
		return errors.Wrap(err, "error to format event")
	}

	if len(event.Target.AuthHeader) > 0 {
		header.Set("Authorization", event.Target.AuthHeader)
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return errors.Wrap(err, "error to marshal header")
	}

	j.Parameters = map[string]interface{}{
		"payload":          string(payload),
		"address":          event.Target.Address,
		"header":           string(headerBytes),
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}
