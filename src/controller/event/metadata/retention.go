package metadata

import (
	"time"

	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

// RetentionMetaData defines tag retention related event data
type RetentionMetaData struct {
	Total    int
	Retained int
	Deleted  []*selector.Result
	Status   string
	TaskID   int64
}

// Resolve tag retention metadata into tag retention event
func (r *RetentionMetaData) Resolve(evt *event.Event) error {
	data := &event2.RetentionEvent{
		EventType: event2.TopicTagRetention,
		OccurAt:   time.Now(),
		Status:    r.Status,
		Deleted:   r.Deleted,
		TaskID:    r.TaskID,
	}

	evt.Topic = event2.TopicTagRetention
	evt.Data = data
	return nil
}
