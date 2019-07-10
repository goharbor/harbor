package model

// Payload of webhook event
type Payload struct {
	Type       string       `json:"type"`
	OccurAt    int64        `json:"occur_at"`
	MediaType  string       `json:"media_type"`
	EventData  []*EventData `json:"event_data,omitempty"`
	Repository *Repository  `json:"repository"`
	Operator   string       `json:"operator"`
}

// EventData of image webhook event
type EventData struct {
	Digest       string        `json:"digest,omitempty"`
	Tag          string        `json:"tag"`
	ResourceURL  string        `json:"resource_url,omitempty"`
	ScanOverview *ScanOverview `json:"scan_overview,omitempty"`
}

// Repository info of webhook event
type Repository struct {
	DateCreated  int64  `json:"date_created"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	RepoFullName string `json:"repo_full_name"`
	RepoType     string `json:"repo_type"`
}

// ScanOverview of scan result
type ScanOverview struct {
	Components *Components `json:"components,omitempty"`
}

// Components of scan result
type Components struct {
	Total   uint64     `json:"total"`
	Summary []*Summary `json:"summary"`
}

// Summary of scan result
type Summary struct {
	Count    uint64 `json:"count"`
	Severity uint64 `json:"severity"`
}
