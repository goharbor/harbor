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

package model

import (
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// WebhookJob ...
type WebhookJob struct {
	*task.Execution
}

// ToSwagger ...
func (n *WebhookJob) ToSwagger() *models.WebhookJob {
	webhookJob := &models.WebhookJob{
		ID:           n.ID,
		PolicyID:     n.VendorID,
		Status:       n.Status,
		CreationTime: strfmt.DateTime(n.StartTime),
		UpdateTime:   strfmt.DateTime(n.UpdateTime),
	}

	var notifyType string
	// do the conversion for compatible with old API
	if n.VendorType == job.WebhookJobVendorType {
		notifyType = "http"
	} else if n.VendorType == job.SlackJobVendorType {
		notifyType = "slack"
	} else if n.VendorType == job.TeamsJobVendorType {
		notifyType = "teams"
	}
	webhookJob.NotifyType = notifyType

	if n.ExtraAttrs != nil {
		if eventType, ok := n.ExtraAttrs["event_type"].(string); ok {
			webhookJob.EventType = eventType
		}

		if payload, ok := n.ExtraAttrs["payload"].(string); ok {
			webhookJob.JobDetail = payload
		}
	}

	return webhookJob
}

// NewWebhookJob ...
func NewWebhookJob(exec *task.Execution) *WebhookJob {
	return &WebhookJob{
		Execution: exec,
	}
}
