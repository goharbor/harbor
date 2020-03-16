package metadata

import (
	event2 "github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/pkg/errors"
	"time"
)

const (
	autoTriggeredOperator = "auto"
)

// ScanImageMetaData defines meta data of image scanning event
type ScanImageMetaData struct {
	Artifact *v1.Artifact
	Status   string
}

// Resolve image scanning metadata into common chart event
func (si *ScanImageMetaData) Resolve(evt *event.Event) error {
	var eventType string
	var topic string

	switch si.Status {
	case models.JobFinished:
		eventType = event2.TopicScanningCompleted
		topic = event2.TopicScanningCompleted
	case models.JobError, models.JobStopped:
		eventType = event2.TopicScanningFailed
		topic = event2.TopicScanningFailed
	default:
		return errors.New("not supported scan hook status")
	}

	data := &event2.ScanImageEvent{
		EventType: eventType,
		Artifact:  si.Artifact,
		OccurAt:   time.Now(),
		Operator:  autoTriggeredOperator,
	}

	evt.Topic = topic
	evt.Data = data
	return nil
}
