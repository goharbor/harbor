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
	"container/list"
	"context"

	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification/hook"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	n_event "github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/notifier/formats"
	notifier_model "github.com/goharbor/harbor/src/pkg/notifier/model"
)

type (
	// EventType is the type of event
	EventType string
	// NotifyType is the type of notify
	NotifyType string
	// PayloadFormatType is the type of payload format
	PayloadFormatType string
)

func (e EventType) String() string {
	return string(e)
}

func (n NotifyType) String() string {
	return string(n)
}

func (p PayloadFormatType) String() string {
	return string(p)
}

var (
	// PolicyMgr is a global notification policy manager
	PolicyMgr policy.Manager

	// HookManager is a hook manager
	HookManager hook.Manager

	// supportedEventTypes is a slice to store supported event type, eg. pushImage, pullImage etc
	supportedEventTypes []EventType

	// supportedNotifyTypes is a slice to store notification type, eg. HTTP, Email etc
	supportedNotifyTypes []NotifyType

	// supportedPayloadFormatTypes is a slice to store the supported payload formats. eg. Default, CloudEvents etc
	supportedPayloadFormatTypes []PayloadFormatType
)

// Init ...
func Init() {
	// init notification policy manager
	PolicyMgr = policy.Mgr
	// init hook manager
	HookManager = hook.NewHookManager()

	initSupportedNotifyType()

	log.Info("notification initialization completed")
}

func initSupportedNotifyType() {
	supportedEventTypes = make([]EventType, 0)
	supportedNotifyTypes = make([]NotifyType, 0)

	eventTypes := []string{
		event.TopicPushArtifact,
		event.TopicPullArtifact,
		event.TopicDeleteArtifact,
		event.TopicQuotaExceed,
		event.TopicQuotaWarning,
		event.TopicScanningFailed,
		event.TopicScanningStopped,
		event.TopicScanningCompleted,
		event.TopicReplication,
		event.TopicTagRetention,
	}
	for _, eventType := range eventTypes {
		supportedEventTypes = append(supportedEventTypes, EventType(eventType))
	}

	notifyTypes := []string{notifier_model.NotifyTypeHTTP, notifier_model.NotifyTypeSlack, notifier_model.NotifyTypeTeams}
	for _, notifyType := range notifyTypes {
		supportedNotifyTypes = append(supportedNotifyTypes, NotifyType(notifyType))
	}

	payloadFormats := []string{formats.DefaultFormat, formats.CloudEventsFormat}
	for _, payloadFormat := range payloadFormats {
		supportedPayloadFormatTypes = append(supportedPayloadFormatTypes, PayloadFormatType(payloadFormat))
	}
}

type eventKey struct{}

// EventCtx ...
type EventCtx struct {
	Events     *list.List
	MustNotify bool
}

// NewEventCtx returns instance of EventCtx
func NewEventCtx() *EventCtx {
	return &EventCtx{
		Events:     list.New(),
		MustNotify: false,
	}
}

// NewContext returns new context with event
func NewContext(ctx context.Context, ec *EventCtx) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, eventKey{}, ec)
}

// AddEvent add events into request context, the event will be sent by the notification middleware eventually.
func AddEvent(ctx context.Context, m n_event.Metadata, notify ...bool) {
	if m == nil {
		return
	}

	e, ok := ctx.Value(eventKey{}).(*EventCtx)
	if !ok {
		log.Debug("request has not event list, cannot add event into context")
		return
	}
	if len(notify) != 0 {
		e.MustNotify = notify[0]
	}
	e.Events.PushBack(m)
}

func GetSupportedEventTypes() []EventType {
	return supportedEventTypes
}

func GetSupportedNotifyTypes() []NotifyType {
	return supportedNotifyTypes
}

func GetSupportedPayloadFormats() []PayloadFormatType {
	return supportedPayloadFormatTypes
}
