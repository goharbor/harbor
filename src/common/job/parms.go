package job

// ScanJobParms holds parameters used to submit jobs to jobservice
type ScanJobParms struct {
	JobID      int64  `json:"job_int_id"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	Digest     string `json:"digest"`
}
