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
	// AMQPContentType defines the content type for AMQP messages
	AMQPContentType = "application/json"
)

// AMQPHandler preprocess event data to amqp and start the hook processing
type AMQPHandler struct {
}

// Name ...
func (a *AMQPHandler) Name() string {
	return "AMQP"
}

// Handle handles event to amqp
func (a *AMQPHandler) Handle(ctx context.Context, value any) error {
	if value == nil {
		return fmt.Errorf("AMQPHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return fmt.Errorf("invalid notification amqp event")
	}

	return a.process(ctx, event)
}

// IsStateful ...
func (a *AMQPHandler) IsStateful() bool {
	return false
}

func (a *AMQPHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	// Create an amqpJob to publish to amqp
	j.Name = job.AMQPJobVendorType

	// Convert payload to amqp format
	payload, err := a.convert(event.Payload)
	if err != nil {
		return fmt.Errorf("convert payload to amqp failed: %v", err)
	}

	j.Parameters = map[string]any{
		"payload":          payload,
		"queue":            event.Target.Address, // Assume address is the queue name
		"content_type":     AMQPContentType,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}

func (a *AMQPHandler) convert(payLoad *model.Payload) (string, error) {
	// For AMQP, send the full payload as JSON
	payloadBytes, err := json.Marshal(payLoad)
	if err != nil {
		return "", fmt.Errorf("failed to marshal amqp payload: %v", err)
	}
	return string(payloadBytes), nil
}