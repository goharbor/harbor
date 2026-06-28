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
	// TeamsContentType defines the content type for Teams messages
	TeamsContentType = "application/json"
)

// TeamsHandler preprocess event data to teams and start the hook processing
type TeamsHandler struct {
}

// Name ...
func (t *TeamsHandler) Name() string {
	return "Teams"
}

// Handle handles event to teams
func (t *TeamsHandler) Handle(ctx context.Context, value any) error {
	if value == nil {
		return fmt.Errorf("TeamsHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return fmt.Errorf("invalid notification teams event")
	}

	return t.process(ctx, event)
}

// IsStateful ...
func (t *TeamsHandler) IsStateful() bool {
	return false
}

func (t *TeamsHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	// Create a teamsJob to send message to teams
	j.Name = job.TeamsJobVendorType

	// Convert payload to teams format
	payload, err := t.convert(event.Payload)
	if err != nil {
		return fmt.Errorf("convert payload to teams body failed: %v", err)
	}

	j.Parameters = map[string]any{
		"payload":          payload,
		"address":          event.Target.Address,
		"content_type":     TeamsContentType,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}

func (t *TeamsHandler) convert(payLoad *model.Payload) (string, error) {
	eventData, err := json.MarshalIndent(payLoad.EventData, "", "\t")
	if err != nil {
		return "", fmt.Errorf("marshal from eventData %v failed: %v", payLoad.EventData, err)
	}

	// Teams supports Adaptive Cards or simple text
	// Using simple text for simplicity
	message := map[string]string{
		"text": fmt.Sprintf("**Harbor Event Notification**\n\n"+
			"**Event Type:** %s\n"+
			"**Occurred At:** %d\n"+
			"**Operator:** %s\n\n"+
			"**Event Data:**\n```json\n%s\n```",
			payLoad.Type, payLoad.OccurAt, payLoad.Operator, string(eventData)),
	}

	payloadBytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal teams payload: %v", err)
	}
	return string(payloadBytes), nil
}