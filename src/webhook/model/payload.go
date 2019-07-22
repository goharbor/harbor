package model

// Payload of webhook event
type Payload struct {
	Type      string     `json:"type"`
	OccurAt   int64      `json:"occur_at"`
	EventData *EventData `json:"event_data,omitempty"`
	Operator  string     `json:"operator"`
}

// EventData of webhook event payload
type EventData struct {
	Resources  []*Resource `json:"resources"`
	Repository *Repository `json:"repository"`
}

// Resource describe infos of resource triggered webhook
type Resource struct {
	Digest       string        `json:"digest,omitempty"`
	Tag          string        `json:"tag"`
	ResourceURL  string        `json:"resource_url,omitempty"`
	ScanOverview *ScanOverview `json:"scan_overview,omitempty"`
}

// Repository info of webhook event
type Repository struct {
	DateCreated  int64  `json:"date_created,omitempty"`
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
