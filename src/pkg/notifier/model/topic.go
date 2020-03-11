package model

// Define global topic names
const (
	// TagTopic
	PushTagTopic   = "PushTagTopic"
	PullTagTopic   = "PullTagTopic"
	DeleteTagTopic = "DeleteTagTopic"

	// ProjectTopic ...
	CreateProjectTopic = "CreateProjectTopic"
	DeleteProjectTopic = "DeleteProjectTopic"

	// RepositoryTopic ...
	CreateRepositoryTopic = "CreateRepositoryTopic"
	DeleteRepositoryTopic = "DeleteRepositoryTopic"

	// ArtifactTopic
	CreateArtifactTopic = "CreateArtifactTopic"
	DeleteArtifactTopic = "DeleteArtifactTopic"

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
	// QuotaExceedTopic is topic for quota warning event, the usage reaches the warning bar of limitation, like 85%
	QuotaWarningTopic = "OnQuotaWarning"
	// QuotaExceedTopic is topic for quota exceeded event
	QuotaExceedTopic = "OnQuotaExceed"

	// WebhookTopic is topic for sending webhook payload
	WebhookTopic = "http"
	// SlackTopic is topic for sending slack payload
	SlackTopic = "slack"
	// EmailTopic is topic for sending email payload
	EmailTopic = "email"
)
