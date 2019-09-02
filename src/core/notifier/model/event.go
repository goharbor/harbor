package model

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
)

// ImageEvent is image related event data to publish
type ImageEvent struct {
	EventType string
	Project   *models.Project
	Resource  []*ImgResource
	OccurAt   time.Time
	Operator  string
	RepoName  string
}

// ImgResource include image digest and tag
type ImgResource struct {
	Digest string
	Tag    string
}

// ChartEvent is chart related event data to publish
type ChartEvent struct {
	EventType   string
	ProjectName string
	ChartName   string
	Versions    []string
	OccurAt     time.Time
	Operator    string
}

// ScanImageEvent is scanning image related event data to publish
type ScanImageEvent struct {
	EventType string
	JobID     int64
	OccurAt   time.Time
	Operator  string
}

// HookEvent is hook related event data to publish
type HookEvent struct {
	PolicyID  int64
	EventType string
	Target    *models.EventTarget
	Payload   *Payload
}

// Payload of notification event
type Payload struct {
	Type      string     `json:"type"`
	OccurAt   int64      `json:"occur_at"`
	EventData *EventData `json:"event_data,omitempty"`
	Operator  string     `json:"operator"`
}

// EventData of notification event payload
type EventData struct {
	Resources  []*Resource `json:"resources"`
	Repository *Repository `json:"repository"`
}

// Resource describe infos of resource triggered notification
type Resource struct {
	Digest       string                  `json:"digest,omitempty"`
	Tag          string                  `json:"tag"`
	ResourceURL  string                  `json:"resource_url,omitempty"`
	ScanOverview *models.ImgScanOverview `json:"scan_overview,omitempty"`
}

// Repository info of notification event
type Repository struct {
	DateCreated  int64  `json:"date_created,omitempty"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	RepoFullName string `json:"repo_full_name"`
	RepoType     string `json:"repo_type"`
}
