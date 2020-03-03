package model

// const definitions
const (
	EventTypePushImage         = "pushImage"
	EventTypePullImage         = "pullImage"
	EventTypeDeleteImage       = "deleteImage"
	EventTypeUploadChart       = "uploadChart"
	EventTypeDeleteChart       = "deleteChart"
	EventTypeDownloadChart     = "downloadChart"
	EventTypeScanningCompleted = "scanningCompleted"
	EventTypeScanningFailed    = "scanningFailed"
	EventTypeTestEndpoint      = "testEndpoint"
	EventTypeProjectQuota      = "projectQuota"

	NotifyTypeHTTP  = "http"
	NotifyTypeSlack = "slack"
)
