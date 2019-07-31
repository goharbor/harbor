package notifier

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier/model"
)

// Event to publish
type Event struct {
	Topic string
	Data  interface{}
}

// EventMetadata is the event raw data to be processed
type EventMetadata interface {
	Resolve(event *Event) error
}

// ImageDelMetaData defines images deleting related event data
type ImageDelMetaData struct {
	Topic    string
	Project  *models.Project
	Tags     []string
	OccurAt  time.Time
	Operator string
	RepoName string
}

// Resolve image deleting metadata into common image event
func (i *ImageDelMetaData) Resolve(evt *Event) error {
	data := &model.ImageEvent{
		Project:  i.Project,
		OccurAt:  i.OccurAt,
		Operator: i.Operator,
		RepoName: i.RepoName,
	}
	for _, t := range i.Tags {
		res := &model.ImgResource{Tag: t}
		data.Resource = append(data.Resource, res)
	}
	evt.Topic = i.Topic
	evt.Data = data
	return nil
}

// ImagePushPullMetaData defines images pushing&pulling related event data
type ImagePushPullMetaData struct {
	Topic    string
	Project  *models.Project
	Tag      string
	Digest   string
	OccurAt  time.Time
	Operator string
	RepoName string
}

// Resolve image pushing&pulling metadata into common image event
func (i *ImagePushPullMetaData) Resolve(evt *Event) error {
	data := &model.ImageEvent{
		Project:  i.Project,
		OccurAt:  i.OccurAt,
		Operator: i.Operator,
		RepoName: i.RepoName,
		Resource: []*model.ImgResource{
			{
				Tag:    i.Tag,
				Digest: i.Digest,
			},
		},
	}

	evt.Topic = i.Topic
	evt.Data = data
	return nil
}

// HookMetaData defines hook notification related event data
type HookMetaData struct {
	Topic     string
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

	evt.Topic = h.Topic
	evt.Data = data
	return nil
}

// Build an event by metadata
func (e *Event) Build(metadata ...EventMetadata) error {
	for _, md := range metadata {
		if err := md.Resolve(e); err != nil {
			return err
		}
	}
	return nil
}

// Publish an event
func (e *Event) Publish() error {
	if err := Publish(e.Topic, e.Data); err != nil {
		log.Errorf("failed to publish topic %s with event: %v", e.Topic, err)
		return err
	}
	return nil
}
