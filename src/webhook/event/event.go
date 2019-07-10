package event

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/webhook/event/topic"
	"github.com/goharbor/harbor/src/webhook/model"
	"github.com/goharbor/harbor/src/webhook/scheduler"
)

const (
	// MediaTypeContainerImage ...
	MediaTypeContainerImage = "containerImage"

	// MediaTypeHelmChart ...
	MediaTypeHelmChart = "helmChart"
)

// Event ...
type Event struct {
	HookType    string
	ProjectID   int64
	ProjectName string
	Payload     *model.Payload
}

// ImageEvent ...
type ImageEvent struct {
	HookType      string
	ProjectID     int64
	ProjectName   string
	ProjectPublic bool
	Events        []*models.Event
	OccurAt       time.Time
	Operator      string
	RepoName      string
}

// ChartEvent ...
type ChartEvent struct {
	HookType       string
	ProjectName    string
	ChartName      string
	ChartVersions  []string
	Operator       string
	OccurTime      time.Time
	RepoCreateTime time.Time
}

// Subscribe topics
func init() {
	handlersMap := map[string][]notifier.NotificationHandler{
		topic.WebhookEventTopicOnImage: {&ImageWebhookHandler{}},
		topic.WebhookEventTopicOnChart: {&ChartWebhookHandler{}},
		topic.WebhookSendTopicOnHTTP:   {&scheduler.HTTPScheduler{}},
	}

	for t, handlers := range handlersMap {
		for _, handler := range handlers {
			if err := notifier.Subscribe(t, handler); err != nil {
				log.Errorf("failed to subscribe topic %s: %v", t, err)
				continue
			}
			log.Debugf("topic %s is subscribed", t)
		}
	}
}
