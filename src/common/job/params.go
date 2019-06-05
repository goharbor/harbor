package job

// ScanJobParams holds parameters used to submit jobs to jobservice
type ScanJobParams struct {
	JobID      int64  `json:"job_int_id"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Digest     string `json:"digest"`
}
