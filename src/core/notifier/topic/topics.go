package topic

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/core/notifier/handler/webhook"
)

// Define global topic names
const (
	// ScanAllPolicyTopic is for notifying the change of scanning all policy.
	ScanAllPolicyTopic = common.ScanAllPolicy

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

	// WebhookHTTPTopic is topic for sending webhook payload by http
	WebhookHTTPTopic = "http"
	// WebhookEmailTopic is topic sending webhook payload by email
	WebhookEmailTopic = "email"
)

//Subscribe topics
func init() {
	handlersMap := map[string][]notifier.NotificationHandler{
		PushImageTopic:   {&webhook.PushImagePreprocessHandler{}},
		PullImageTopic:   {&webhook.PullImagePreprocessHandler{}},
		DeleteImageTopic: {&webhook.DeleteImagePreprocessHandler{}},
		WebhookHTTPTopic: {&webhook.HTTPScheduler{}},
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
