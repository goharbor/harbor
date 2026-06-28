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
	// MatrixMessageTemplate defines Matrix message payload template
	MatrixMessageTemplate = `{
		"msgtype": "m.text",
		"body": "**Harbor Event:** %s\n**Type:** %s\n**Occurred At:** %s\n**Operator:** %s\n**Event Data:** %s"
	}`
)

// MatrixHandler preprocess event data to matrix and start the hook processing
type MatrixHandler struct {
}

// Name ...
func (m *MatrixHandler) Name() string {
	return "Matrix"
}

// Handle handles event to matrix
func (m *MatrixHandler) Handle(ctx context.Context, value any) error {
	if value == nil {
		return fmt.Errorf("MatrixHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return fmt.Errorf("invalid notification matrix event")
	}

	return m.process(ctx, event)
}

// IsStateful ...
func (m *MatrixHandler) IsStateful() bool {
	return false
}

func (m *MatrixHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	// Create a matrixJob to send message to matrix
	j.Name = job.MatrixJobVendorType

	// Convert payload to matrix format
	payload, err := m.convert(event.Payload)
	if err != nil {
		return fmt.Errorf("convert payload to matrix body failed: %v", err)
	}

	j.Parameters = map[string]any{
		"payload":          payload,
		"address":          event.Target.Address,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}

func (m *MatrixHandler) convert(payLoad *model.Payload) (string, error) {
	eventData, err := json.MarshalIndent(payLoad.EventData, "", "\t")
	if err != nil {
		return "", fmt.Errorf("marshal from eventData %v failed: %v", payLoad.EventData, err)
	}

	// Format as Matrix message
	message := map[string]string{
		"msgtype": "m.text",
		"body": fmt.Sprintf("**Harbor Event Notification**\n"+
			"**Event Type:** %s\n"+
			"**Occurred At:** %d\n"+
			"**Operator:** %s\n"+
			"**Event Data:**\n```json\n%s\n```",
			payLoad.Type, payLoad.OccurAt, payLoad.Operator, string(eventData)),
	}

	payloadBytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal matrix payload: %v", err)
	}
	return string(payloadBytes), nil
}