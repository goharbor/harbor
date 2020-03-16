package metadata

import (
	event2 "github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"time"
)

// ChartMetaData defines meta data of chart event
type ChartMetaData struct {
	ProjectName string
	ChartName   string
	Versions    []string
	OccurAt     time.Time
	Operator    string
}

func (cmd *ChartMetaData) convert(evt *event2.ChartEvent) {
	evt.ProjectName = cmd.ProjectName
	evt.OccurAt = cmd.OccurAt
	evt.Operator = cmd.Operator
	evt.ChartName = cmd.ChartName
	evt.Versions = cmd.Versions
}

// ChartUploadMetaData defines meta data of chart upload event
type ChartUploadMetaData struct {
	ChartMetaData
}

// Resolve chart uploading metadata into common chart event
func (cu *ChartUploadMetaData) Resolve(event *event.Event) error {
	data := &event2.ChartEvent{
		EventType: event2.TopicUploadChart,
	}
	cu.convert(data)

	event.Topic = event2.TopicUploadChart
	event.Data = data
	return nil
}

// ChartDownloadMetaData defines meta data of chart download event
type ChartDownloadMetaData struct {
	ChartMetaData
}

// Resolve chart download metadata into common chart event
func (cd *ChartDownloadMetaData) Resolve(evt *event.Event) error {
	data := &event2.ChartEvent{
		EventType: event2.TopicDownloadChart,
	}
	cd.convert(data)

	evt.Topic = event2.TopicDownloadChart
	evt.Data = data
	return nil
}

// ChartDeleteMetaData defines meta data of chart delete event
type ChartDeleteMetaData struct {
	ChartMetaData
}

// Resolve chart delete metadata into common chart event
func (cd *ChartDeleteMetaData) Resolve(evt *event.Event) error {
	data := &event2.ChartEvent{
		EventType: event2.TopicDeleteChart,
	}
	cd.convert(data)

	evt.Topic = event2.TopicDeleteChart
	evt.Data = data
	return nil
}
