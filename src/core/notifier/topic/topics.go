package topic

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/core/notifier/handler/notification"
)

// Define global topic names
const (
	// PushImageTopic is topic for push image event
	PushImageTopic = "OnPushImage"
	// PullImageTopic is topic for pull image event
	PullImageTopic = "OnPullImage"
	// DeleteImageTopic is topic for delete image event
	DeleteImageTopic = "OnDeleteImage"
	// UploadChartTopic is topic for upload chart event
	UploadChartTopic = "OnUploadChart"
	// DownloadChartTopic is topic for download chart event
	DownloadChartTopic = "OnDownloadChart"
	// DeleteChartTopic is topic for delete chart event
	DeleteChartTopic = "OnDeleteChart"
	// ScanningFailedTopic is topic for scanning failed event
	ScanningFailedTopic = "OnScanningFailed"
	// ScanningCompletedTopic is topic for scanning completed event
	ScanningCompletedTopic = "OnScanningCompleted"

	// WebhookTopic is topic for sending webhook payload
	WebhookTopic = "http"
	// EmailTopic is topic for sending email payload
	EmailTopic = "email"
)

// Subscribe topics
func init() {
	handlersMap := map[string][]notifier.NotificationHandler{
		PushImageTopic:   {&notification.PushImagePreprocessHandler{}},
		PullImageTopic:   {&notification.PullImagePreprocessHandler{}},
		DeleteImageTopic: {&notification.DeleteImagePreprocessHandler{}},
		WebhookTopic:     {&notification.HTTPHandler{}},
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
