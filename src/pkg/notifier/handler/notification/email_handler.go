// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

const (
	// EmailSubjectTemplate defines email subject template
	EmailSubjectTemplate = "Harbor Event: %s"
	// EmailBodyTemplate defines email body template
	EmailBodyTemplate = `Harbor Event Notification

Event Type: %s
Occurred At: %d
Operator: %s

Event Data:
%s
`
)

// EmailHandler preprocess event data to email and start the hook processing
type EmailHandler struct {
}

// Name ...
func (e *EmailHandler) Name() string {
	return "Email"
}

// Handle handles event to email
func (e *EmailHandler) Handle(ctx context.Context, value any) error {
	if value == nil {
		return fmt.Errorf("EmailHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return fmt.Errorf("invalid notification email event")
	}

	return e.process(ctx, event)
}

// IsStateful ...
func (e *EmailHandler) IsStateful() bool {
	return false
}

func (e *EmailHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	// Create an emailJob to send email
	j.Name = job.EmailJobVendorType

	// Convert payload to email format
	subject, body, err := e.convert(event.Payload)
	if err != nil {
		return fmt.Errorf("convert payload to email failed: %v", err)
	}

	j.Parameters = map[string]any{
		"subject": subject,
		"body":    body,
		"to":      event.Target.Address, // Assume address is the recipient email
	}
	return notification.HookManager.StartHook(ctx, event, j)
}

func (e *EmailHandler) convert(payLoad *model.Payload) (string, string, error) {
	eventData, err := json.MarshalIndent(payLoad.EventData, "", "\t")
	if err != nil {
		return "", "", fmt.Errorf("marshal from eventData %v failed: %v", payLoad.EventData, err)
	}

	subject := fmt.Sprintf(EmailSubjectTemplate, payLoad.Type)
	body := fmt.Sprintf(EmailBodyTemplate, payLoad.Type, payLoad.OccurAt, payLoad.Operator, string(eventData))

	return subject, body, nil
}