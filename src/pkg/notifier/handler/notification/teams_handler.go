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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

const (
	// TeamsBodyTemplate defines Teams request body template
	TeamsBodyTemplate = `{
		"type": "message",
		"attachments": [
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"contentUrl": null,
				"content": {
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"type": "AdaptiveCard",
					"version": "1.4",
					"body": [
						{
							"type": "TextBlock",
							"text": "**Harbor webhook events**"
						},
						{
							"type": "TextBlock",
							"text": "**event_type:** {{.Type}}"
						},
						{
							"type": "TextBlock",
							"text": "**occur_at:** {{.OccurAt}}"
						},
						{
							"type": "TextBlock",
							"text": "**operator:** {{.Operator}}"
						},
						{
							"type": "TextBlock",
							"text": "**event_data:**"
						},
						{
							"type": "TextBlock",
							"text": "{{.EventData}}",
							"wrap": true
						}
					]
				}
			}
		]
	}`
)

// TeamsHandler preprocess event data to teams and start the hook processing
type TeamsHandler struct {
}

// Name ...
func (s *TeamsHandler) Name() string {
	return "Teams"
}

// Handle handles event to teams
func (s *TeamsHandler) Handle(ctx context.Context, value interface{}) error {
	if value == nil {
		return errors.New("TeamsHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return errors.New("invalid notification teams event")
	}

	return s.process(ctx, event)
}

// IsStateful ...
func (s *TeamsHandler) IsStateful() bool {
	return false
}

func (s *TeamsHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	// Create a teamsJob to send message to teams
	j.Name = job.TeamsJobVendorType

	// Convert payload to teams format
	payload, err := s.convert(event.Payload)
	if err != nil {
		return fmt.Errorf("convert payload to teams body failed: %v", err)
	}

	j.Parameters = map[string]interface{}{
		"payload":          payload,
		"address":          event.Target.Address,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}

func (s *TeamsHandler) convert(payLoad *model.Payload) (string, error) {
	data := make(map[string]interface{})
	data["Type"] = payLoad.Type
	data["OccurAt"] = payLoad.OccurAt
	data["Operator"] = payLoad.Operator
	eventData, err := json.MarshalIndent(payLoad.EventData, "", "\t")
	if err != nil {
		return "", fmt.Errorf("marshal from eventData %v failed: %v", payLoad.EventData, err)
	}
	data["EventData"] = escapeEventData(string(eventData))

	tt, _ := template.New("teams").Parse(TeamsBodyTemplate)
	var teamsBuf bytes.Buffer
	if err := tt.Execute(&teamsBuf, data); err != nil {
		return "", fmt.Errorf("%v", err)
	}
	return teamsBuf.String(), nil
}
