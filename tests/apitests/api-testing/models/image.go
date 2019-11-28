package models

//Repository : For /api/repositories
type Repository struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//Tag : For /api/repositories/:repo/tags
type Tag struct {
	Digest       string                  `json:"digest"`
	Name         string                  `json:"name"`
	Signature    map[string]interface{}  `json:"signature, omitempty"`
	ScanOverview map[string]ScanOverview `json:"scan_overview, omitempty"`
}

//ScanOverview : For scanning
type ScanOverview struct {
	Status string `json:"scan_status"`
}
