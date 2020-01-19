package topic

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/handler/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// Subscribe topics
func init() {
	handlersMap := map[string][]notifier.NotificationHandler{
		model.PushImageTopic:         {&notification.ImagePreprocessHandler{}},
		model.PullImageTopic:         {&notification.ImagePreprocessHandler{}},
		model.DeleteImageTopic:       {&notification.ImagePreprocessHandler{}},
		model.WebhookTopic:           {&notification.HTTPHandler{}},
		model.UploadChartTopic:       {&notification.ChartPreprocessHandler{}},
		model.DownloadChartTopic:     {&notification.ChartPreprocessHandler{}},
		model.DeleteChartTopic:       {&notification.ChartPreprocessHandler{}},
		model.ScanningCompletedTopic: {&notification.ScanImagePreprocessHandler{}},
		model.ScanningFailedTopic:    {&notification.ScanImagePreprocessHandler{}},
		model.QuotaExceedTopic:       {&notification.QuotaPreprocessHandler{}},
		model.SlackTopic:             {&notification.SlackHandler{}},
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
