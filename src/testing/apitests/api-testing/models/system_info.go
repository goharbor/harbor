package models

// SystemInfo : For GET /api/systeminfo
type SystemInfo struct {
	AuthMode    string `json:"auth_mode"`
	RegistryURL string `json:"registry_url"`
}
