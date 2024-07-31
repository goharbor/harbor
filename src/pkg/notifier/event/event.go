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

package event

import (
	"context"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	policy_model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// Event to publish
type Event struct {
	Topic string
	Data  interface{}
}

// TopicEvent - Events that contains topic information
type TopicEvent interface {
	Topic() string
}

// New ...
func New() *Event {
	return &Event{}
}

// WithTopicEvent - builder method
func (e *Event) WithTopicEvent(topicEvent TopicEvent) *Event {
	e.Topic = topicEvent.Topic()
	e.Data = topicEvent
	return e
}

// Metadata is the event raw data to be processed
type Metadata interface {
	Resolve(event *Event) error
}

// HookMetaData defines hook notification related event data
type HookMetaData struct {
	ProjectID int64
	PolicyID  int64
	EventType string
	Target    *policy_model.EventTarget
	Payload   *model.Payload
}

// Resolve hook metadata into hook event
func (h *HookMetaData) Resolve(evt *Event) error {
	data := &model.HookEvent{
		ProjectID: h.ProjectID,
		PolicyID:  h.PolicyID,
		EventType: h.EventType,
		Target:    h.Target,
		Payload:   h.Payload,
	}

	evt.Topic = h.Target.Type
	evt.Data = data
	return nil
}

// Build an event by metadata
func (e *Event) Build(_ context.Context, metadata ...Metadata) error {
	for _, md := range metadata {
		if err := md.Resolve(e); err != nil {
			log.Debugf("failed to resolve event metadata: %v", md)
			return errors.Wrap(err, "failed to resolve event metadata")
		}
	}
	return nil
}

// Publish an event
func (e *Event) Publish(ctx context.Context) error {
	if err := notifier.Publish(ctx, e.Topic, e.Data); err != nil {
		log.Debugf("failed to publish topic %s with event: %v", e.Topic, e.Data)
		return errors.Wrap(err, "failed to publish event")
	}
	return nil
}

// BuildAndPublish builds the event according to the metadata and publish the event
// The process is done in a separated goroutine
func BuildAndPublish(ctx context.Context, metadata ...Metadata) {
	go func() {
		event := &Event{}
		if err := event.Build(ctx, metadata...); err != nil {
			log.Errorf("failed to build the event from metadata: %v", err)
			return
		}
		if err := event.Publish(ctx); err != nil {
			log.Errorf("failed to publish the event %s: %v", event.Topic, err)
			return
		}
		log.Debugf("event %s published", event.Topic)
	}()
}
