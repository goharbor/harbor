package export

// CsvJobVendorID  specific type to be used in contexts
type CsvJobVendorID string

const (
	JobNameAttribute   = "job_name"
	UserNameAttribute  = "user_name"
	ScanDataExportDir  = "/var/scandata_exports"
	QueryPageSize      = 100
	DigestKey          = "artifact_digest"
	CreateTimestampKey = "create_ts"
	Vendor             = "SCAN_DATA_EXPORT"
	CsvJobVendorIDKey  = CsvJobVendorID("vendorId")
)
