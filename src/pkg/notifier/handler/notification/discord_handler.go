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
	// DiscordEmbedTemplate defines Discord embed payload template
	DiscordEmbedTemplate = `{
		"embeds": [
			{
				"title": "Harbor webhook events",
				"fields": [
					{
						"name": "event_type",
						"value": "{{.Type}}"
					},
					{
						"name": "occur_at",
						"value": "<t:{{.OccurAt}}>"
					},
					{
						"name": "operator",
						"value": "{{.Operator}}"
					},
					{
						"name": "event_data",
						"value": "{{.EventData}}"
					}
				]
			}
		]
	}`
)

// DiscordHandler preprocess event data to discord and start the hook processing
type DiscordHandler struct {
}

// Name ...
func (d *DiscordHandler) Name() string {
	return "Discord"
}

// Handle handles event to discord
func (d *DiscordHandler) Handle(ctx context.Context, value any) error {
	if value == nil {
		return fmt.Errorf("DiscordHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return fmt.Errorf("invalid notification discord event")
	}

	return d.process(ctx, event)
}

// IsStateful ...
func (d *DiscordHandler) IsStateful() bool {
	return false
}

func (d *DiscordHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	// Create a discordJob to send message to discord
	j.Name = job.DiscordJobVendorType

	// Convert payload to discord format
	payload, err := d.convert(event.Payload)
	if err != nil {
		return fmt.Errorf("convert payload to discord body failed: %v", err)
	}

	j.Parameters = map[string]any{
		"payload":          payload,
		"address":          event.Target.Address,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}

func (d *DiscordHandler) convert(payLoad *model.Payload) (string, error) {
	data := make(map[string]any)
	data["Type"] = payLoad.Type
	data["OccurAt"] = payLoad.OccurAt
	data["Operator"] = payLoad.Operator
	eventData, err := json.MarshalIndent(payLoad.EventData, "", "\t")
	if err != nil {
		return "", fmt.Errorf("marshal from eventData %v failed: %v", payLoad.EventData, err)
	}
	data["EventData"] = "```" + string(eventData) + "```"

	// For Discord, we can use embeds or simple content
	// Here using embeds similar to Slack blocks
	embedPayload := map[string]any{
		"embeds": []map[string]any{
			{
				"title": "Harbor Webhook Event",
				"fields": []map[string]any{
					{"name": "Event Type", "value": payLoad.Type, "inline": true},
					{"name": "Occurred At", "value": fmt.Sprintf("<t:%d>", payLoad.OccurAt), "inline": true},
					{"name": "Operator", "value": payLoad.Operator, "inline": true},
					{"name": "Event Data", "value": "```json\n" + string(eventData) + "\n```", "inline": false},
				},
				"color": 3447003, // Blue color
			},
		},
	}

	payloadBytes, err := json.Marshal(embedPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal discord payload: %v", err)
	}
	return string(payloadBytes), nil
}