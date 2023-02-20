package export

// CsvJobVendorID  specific type to be used in contexts
type CsvJobVendorID string

const (
	ProjectIDsAttribute    = "project_ids"
	JobNameAttribute       = "job_name"
	UserNameAttribute      = "user_name"
	StatusMessageAttribute = "status_message"
	// the scan data is a temporary file, use /tmp directory to avoid the permission issue.
	ScanDataExportDir  = "/tmp"
	QueryPageSize      = 100000
	ArtifactGroupSize  = 10000
	DigestKey          = "artifact_digest"
	CreateTimestampKey = "create_ts"
	Vendor             = "SCAN_DATA_EXPORT"
	CsvJobVendorIDKey  = CsvJobVendorID("vendorId")
)
