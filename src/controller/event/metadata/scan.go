package metadata

import (
	"fmt"
	"time"

	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
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

	switch job.Status(si.Status) {
	case job.SuccessStatus:
		eventType = event2.TopicScanningCompleted
		topic = event2.TopicScanningCompleted
	case job.StoppedStatus:
		eventType = event2.TopicScanningStopped
		topic = event2.TopicScanningStopped
	case job.ErrorStatus:
		eventType = event2.TopicScanningFailed
		topic = event2.TopicScanningFailed
	default:
		return fmt.Errorf("not supported scan hook status %s", si.Status)
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
