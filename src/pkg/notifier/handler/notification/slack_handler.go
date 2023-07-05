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
	"strings"
	"text/template"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

const (
	// SlackBodyTemplate defines Slack request body template
	SlackBodyTemplate = `{
	"blocks": [
		{
            "type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*Harbor webhook events*"
			}
        },
        {
            "type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*event_type:* {{.Type}}"
			}
        },
        {
            "type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*occur_at:* <!date^{{.OccurAt}}^{date} at {time}|February 18th, 2014 at 6:39 AM PST>"
			}
        },
        {	"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*operator:* {{.Operator}}"
			}
		},
        {	"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "*event_data:*"
			}
		},
		{	"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": "{{.EventData}}"
			}
		}
    ]}`
)

// SlackHandler preprocess event data to slack and start the hook processing
type SlackHandler struct {
}

// Name ...
func (s *SlackHandler) Name() string {
	return "Slack"
}

// Handle handles event to slack
func (s *SlackHandler) Handle(ctx context.Context, value interface{}) error {
	if value == nil {
		return errors.New("SlackHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return errors.New("invalid notification slack event")
	}

	return s.process(ctx, event)
}

// IsStateful ...
func (s *SlackHandler) IsStateful() bool {
	return false
}

func (s *SlackHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	// Create a slackJob to send message to slack
	j.Name = job.SlackJobVendorType

	// Convert payload to slack format
	payload, err := s.convert(event.Payload)
	if err != nil {
		return fmt.Errorf("convert payload to slack body failed: %v", err)
	}

	j.Parameters = map[string]interface{}{
		"payload":          payload,
		"address":          event.Target.Address,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}

func (s *SlackHandler) convert(payLoad *model.Payload) (string, error) {
	data := make(map[string]interface{})
	data["Type"] = payLoad.Type
	data["OccurAt"] = payLoad.OccurAt
	data["Operator"] = payLoad.Operator
	eventData, err := json.MarshalIndent(payLoad.EventData, "", "\t")
	if err != nil {
		return "", fmt.Errorf("marshal from eventData %v failed: %v", payLoad.EventData, err)
	}
	data["EventData"] = "```" + escapeEventData(string(eventData)) + "```"

	st, _ := template.New("slack").Parse(SlackBodyTemplate)
	var slackBuf bytes.Buffer
	if err := st.Execute(&slackBuf, data); err != nil {
		return "", fmt.Errorf("%v", err)
	}
	return slackBuf.String(), nil
}

func escapeEventData(str string) string {
	// escape " to \"
	str = strings.Replace(str, `"`, `\"`, -1)
	// escape \\" to \\\"
	str = strings.Replace(str, `\\"`, `\\\"`, -1)
	return str
}
