package event

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/core/notifier/model"
	notifyModel "github.com/goharbor/harbor/src/pkg/notification/model"
	"github.com/pkg/errors"
)

// Event to publish
type Event struct {
	Topic string
	Data  interface{}
}

// Metadata is the event raw data to be processed
type Metadata interface {
	Resolve(event *Event) error
}

// ImageDelMetaData defines images deleting related event data
type ImageDelMetaData struct {
	Project  *models.Project
	Tags     []string
	OccurAt  time.Time
	Operator string
	RepoName string
}

// Resolve image deleting metadata into common image event
func (i *ImageDelMetaData) Resolve(evt *Event) error {
	data := &model.ImageEvent{
		EventType: notifyModel.EventTypeDeleteImage,
		Project:   i.Project,
		OccurAt:   i.OccurAt,
		Operator:  i.Operator,
		RepoName:  i.RepoName,
	}
	for _, t := range i.Tags {
		res := &model.ImgResource{Tag: t}
		data.Resource = append(data.Resource, res)
	}
	evt.Topic = model.DeleteImageTopic
	evt.Data = data
	return nil
}

// ImagePushMetaData defines images pushing related event data
type ImagePushMetaData struct {
	Project  *models.Project
	Tag      string
	Digest   string
	OccurAt  time.Time
	Operator string
	RepoName string
}

// Resolve image pushing metadata into common image event
func (i *ImagePushMetaData) Resolve(evt *Event) error {
	data := &model.ImageEvent{
		EventType: notifyModel.EventTypePushImage,
		Project:   i.Project,
		OccurAt:   i.OccurAt,
		Operator:  i.Operator,
		RepoName:  i.RepoName,
		Resource: []*model.ImgResource{
			{
				Tag:    i.Tag,
				Digest: i.Digest,
			},
		},
	}

	evt.Topic = model.PushImageTopic
	evt.Data = data
	return nil
}

// ImagePullMetaData defines images pulling related event data
type ImagePullMetaData struct {
	Project  *models.Project
	Tag      string
	Digest   string
	OccurAt  time.Time
	Operator string
	RepoName string
}

// Resolve image pulling metadata into common image event
func (i *ImagePullMetaData) Resolve(evt *Event) error {
	data := &model.ImageEvent{
		EventType: notifyModel.EventTypePullImage,
		Project:   i.Project,
		OccurAt:   i.OccurAt,
		Operator:  i.Operator,
		RepoName:  i.RepoName,
		Resource: []*model.ImgResource{
			{
				Tag:    i.Tag,
				Digest: i.Digest,
			},
		},
	}

	evt.Topic = model.PullImageTopic
	evt.Data = data
	return nil
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
