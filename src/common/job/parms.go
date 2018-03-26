package job

// ScanJobParms holds parameters used to submit jobs to jobservice
type ScanJobParms struct {
	JobID         int64  `json:"job_int_id"`
	Repository    string `json:"repository"`
	Tag           string `json:"tag"`
	Digest        string `json:digest"`
	Secret        string `json:"job_service_secret"`
	RegistryURL   string `json:"registry_url"`
	ClairEndpoint string `json:"clair_endpoint"`
	TokenEndpoint string `json:"token_endpoint"`
}
