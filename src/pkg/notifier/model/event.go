package model

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
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
	Artifact  *v1.Artifact
	OccurAt   time.Time
	Operator  string
}

// QuotaEvent is project quota related event data to publish
type QuotaEvent struct {
	EventType string
	Project   *models.Project
	Resource  *ImgResource
	OccurAt   time.Time
	RepoName  string
	Msg       string
}

// ReplicationEvent is replication related event data to publish
type ReplicationEvent struct {
	EventType         string
	ReplicationTaskID int64
	OccurAt           time.Time
	Status            string
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
	Operator  string     `json:"operator"`
	EventData *EventData `json:"event_data,omitempty"`
}

// EventData of notification event payload
type EventData struct {
	Resources   []*Resource       `json:"resources"`
	Repository  *Repository       `json:"repository"`
	Replication *Replication      `json:"replication"`
	Custom      map[string]string `json:"custom_attributes,omitempty"`
}

// Resource describe infos of resource triggered notification
type Resource struct {
	Digest       string                 `json:"digest,omitempty"`
	Tag          string                 `json:"tag"`
	ResourceURL  string                 `json:"resource_url,omitempty"`
	ScanOverview map[string]interface{} `json:"scan_overview,omitempty"`
}

// Repository info of notification event
type Repository struct {
	DateCreated  int64  `json:"date_created,omitempty"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	RepoFullName string `json:"repo_full_name"`
	RepoType     string `json:"repo_type"`
}

// Replication describes replication infos
type Replication struct {
	HarborHostname     string                `json:"harbor_hostname"`
	JobStatus          string                `json:"job_status"`
	Description        string                `json:"description"`
	ArtifactType       string                `json:"artifact_type"`
	AuthenticationType string                `json:"authentication_type"`
	OverrideMode       bool                  `json:"override_mode"`
	TriggerType        string                `json:"trigger_type"`
	ExecutionTimestamp int64                 `json:"execution_timestamp"`
	Operator           string                `json:"operator"`
	SrcRegistryType    string                `json:"src_registry_type"`
	SrcRegistryName    string                `json:"src_registry_name"`
	SrcEndpoint        string                `json:"src_endpoint"`
	SrcProvider        string                `json:"src_provider"`
	SrcNamespace       string                `json:"src_namespace"`
	SrcProjectName     string                `json:"src_project_name"`
	DestRegistryType   string                `json:"dest_registry_type"`
	DestRegistryName   string                `json:"dest_registry_name"`
	DestEndpoint       string                `json:"dest_endpoint"`
	DestProvider       string                `json:"dest_provider"`
	DestNamespace      string                `json:"dest_namespace"`
	DestProjectName    string                `json:"dest_project_name"`
	SuccessfulArtifact []*SuccessfulArtifact `json:"successful_artifact"`
	FailedArtifact     []*FailedArtifact     `json:"failed_artifact"`
}

// SuccessfulArtifact describe info of artifact successfully replicated
type SuccessfulArtifact struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	NameTag string `json:"name_tag"`
}

// FailedArtifact describe info of artifact unsuccessfully replicated
type FailedArtifact struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	NameTag string `json:"name_tag"`
	Reason  string `json:"reason"`
}
