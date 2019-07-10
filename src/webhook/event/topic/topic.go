package topic

const (
	// WebhookEventTopicOnImage include pushImage, pullImage, deleteImage
	WebhookEventTopicOnImage = "OnImage"
	// WebhookEventTopicOnChart include uploadChart, deleteChart, downloadChart
	WebhookEventTopicOnChart = "OnChart"
	// WebhookEventTopicOnScan include scanningFailed, scanningCompleted
	WebhookEventTopicOnScan = "OnScan"

	// WebhookSendTopicOnHTTP send webhook payload by http
	WebhookSendTopicOnHTTP = "http"
	// WebhookSendTopicOnEmail send webhook payload by email
	WebhookSendTopicOnEmail = "email"
)
