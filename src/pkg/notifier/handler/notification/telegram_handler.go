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
	// TelegramContentType defines the content type for Telegram messages
	TelegramContentType = "application/json"
)

// TelegramHandler preprocess event data to telegram and start the hook processing
type TelegramHandler struct {
}

// Name ...
func (tg *TelegramHandler) Name() string {
	return "Telegram"
}

// Handle handles event to telegram
func (tg *TelegramHandler) Handle(ctx context.Context, value any) error {
	if value == nil {
		return fmt.Errorf("TelegramHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return fmt.Errorf("invalid notification telegram event")
	}

	return tg.process(ctx, event)
}

// IsStateful ...
func (tg *TelegramHandler) IsStateful() bool {
	return false
}

func (tg *TelegramHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	// Create a telegramJob to send message to telegram
	j.Name = job.TelegramJobVendorType

	// Convert payload to telegram format
	text, err := tg.convert(event.Payload)
	if err != nil {
		return fmt.Errorf("convert payload to telegram text failed: %v", err)
	}

	j.Parameters = map[string]any{
		"text":             text,
		"bot_token":       event.Target.AuthHeader, // Use auth header for bot token
		"chat_id":         event.Target.Address,    // Use address for chat ID
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}

func (tg *TelegramHandler) convert(payLoad *model.Payload) (string, error) {
	eventData, err := json.MarshalIndent(payLoad.EventData, "", "\t")
	if err != nil {
		return "", fmt.Errorf("marshal from eventData %v failed: %v", payLoad.EventData, err)
	}

	// Format as Telegram message text with Markdown
	text := fmt.Sprintf("*Harbor Event Notification*\n\n"+
		"*Event Type:* %s\n"+
		"*Occurred At:* %d\n"+
		"*Operator:* %s\n\n"+
		"*Event Data:*\n```json\n%s\n```",
		payLoad.Type, payLoad.OccurAt, payLoad.Operator, string(eventData))

	return text, nil
}