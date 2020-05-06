package metadata

import (
	"time"

	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

// ReplicationMetaData defines replication related event data
type ReplicationMetaData struct {
	ReplicationTaskID int64
	Status            string
}

// Resolve replication metadata into replication event
func (r *ReplicationMetaData) Resolve(evt *event.Event) error {
	data := &event2.ReplicationEvent{
		ReplicationTaskID: r.ReplicationTaskID,
		EventType:         event2.TopicReplication,
		OccurAt:           time.Now(),
		Status:            r.Status,
	}

	evt.Topic = event2.TopicReplication
	evt.Data = data
	return nil
}
