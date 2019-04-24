package job

const (
	// ImageScanJob is name of scan job it will be used as key to register to job service.
	ImageScanJob = "IMAGE_SCAN"
	// ImageScanAllJob is the name of "scanall" job in job service
	ImageScanAllJob = "IMAGE_SCAN_ALL"
	// ImageGC the name of image garbage collection job in job service
	ImageGC = "IMAGE_GC"

	// JobKindGeneric : Kind of generic job
	JobKindGeneric = "Generic"
	// JobKindScheduled : Kind of scheduled job
	JobKindScheduled = "Scheduled"
	// JobKindPeriodic : Kind of periodic job
	JobKindPeriodic = "Periodic"

	// JobServiceStatusPending   : job status pending
	JobServiceStatusPending = "Pending"
	// JobServiceStatusRunning   : job status running
	JobServiceStatusRunning = "Running"
	// JobServiceStatusStopped   : job status stopped
	JobServiceStatusStopped = "Stopped"
	// JobServiceStatusCancelled : job status cancelled
	JobServiceStatusCancelled = "Cancelled"
	// JobServiceStatusError     : job status error
	JobServiceStatusError = "Error"
	// JobServiceStatusSuccess   : job status success
	JobServiceStatusSuccess = "Success"
	// JobServiceStatusScheduled : job status scheduled
	JobServiceStatusScheduled = "Scheduled"

	// JobActionStop : the action to stop the job
	JobActionStop = "stop"
)
