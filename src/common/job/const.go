package job

const (
	//ImageScanJob is name of scan job it will be used as key to register to job service.
	ImageScanJob = "IMAGE_SCAN"
	// GenericKind marks the job as a generic job, it will be contained in job metadata and passed to job service.
	GenericKind = "Generic"
	// ImageTransfer : the name of image transfer job in job service
	ImageTransfer = "IMAGE_TRANSFER"
	// ImageDelete : the name of image delete job in job service
	ImageDelete = "IMAGE_DELETE"
	// ImageReplicate : the name of image replicate job in job service
	ImageReplicate = "IMAGE_REPLICATE"
)
