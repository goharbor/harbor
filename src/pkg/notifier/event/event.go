package event

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/pkg/errors"
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
	PolicyID  int64
	EventType string
	Target    *models.EventTarget
	Payload   *model.Payload
}

// Resolve hook metadata into hook event
func (h *HookMetaData) Resolve(evt *Event) error {
	data := &model.HookEvent{
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
func (e *Event) Build(metadata ...Metadata) error {
	for _, md := range metadata {
		if err := md.Resolve(e); err != nil {
			log.Debugf("failed to resolve event metadata: %v", md)
			return errors.Wrap(err, "failed to resolve event metadata")
		}
	}
	return nil
}

// Publish an event
func (e *Event) Publish() error {
	if err := notifier.Publish(e.Topic, e.Data); err != nil {
		log.Debugf("failed to publish topic %s with event: %v", e.Topic, e.Data)
		return errors.Wrap(err, "failed to publish event")
	}
	return nil
}

// BuildAndPublish builds the event according to the metadata and publish the event
// The process is done in a separated goroutine
func BuildAndPublish(metadata ...Metadata) {
	go func() {
		event := &Event{}
		if err := event.Build(metadata...); err != nil {
			log.Errorf("failed to build the event from metadata: %v", err)
			return
		}
		if err := event.Publish(); err != nil {
			log.Errorf("failed to publish the event %s: %v", event.Topic, err)
			return
		}
		log.Debugf("event %s published", event.Topic)
	}()
}
