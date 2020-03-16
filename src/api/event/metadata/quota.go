package metadata

import (
	event2 "github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/pkg/errors"
	"time"
)

// QuotaMetaData defines quota related event data
type QuotaMetaData struct {
	Project  *models.Project
	RepoName string
	Tag      string
	Digest   string
	// used to define the event topic
	Level int
	// the msg contains the limitation and current usage of quota
	Msg     string
	OccurAt time.Time
}

// Resolve quota exceed into common image event
func (q *QuotaMetaData) Resolve(evt *event.Event) error {
	var topic string
	data := &event2.QuotaEvent{
		EventType: event2.TopicQuotaExceed,
		Project:   q.Project,
		Resource: &event2.ImgResource{
			Tag:    q.Tag,
			Digest: q.Digest,
		},
		OccurAt:  q.OccurAt,
		RepoName: q.RepoName,
		Msg:      q.Msg,
	}

	switch q.Level {
	case 1:
		topic = event2.TopicQuotaExceed
	case 2:
		topic = event2.TopicQuotaWarning
	default:
		return errors.New("not supported quota status")
	}

	evt.Topic = topic
	evt.Data = data
	return nil
}
