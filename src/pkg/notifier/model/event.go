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
	Resources   []*Resource       `json:"resources,omitempty"`
	Repository  *Repository       `json:"repository,omitempty"`
	Replication *Replication      `json:"replication,omitempty"`
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
	HarborHostname     string               `json:"harbor_hostname,omitempty"`
	JobStatus          string               `json:"job_status,omitempty"`
	Description        string               `json:"description,omitempty"`
	ArtifactType       string               `json:"artifact_type,omitempty"`
	AuthenticationType string               `json:"authentication_type,omitempty"`
	OverrideMode       bool                 `json:"override_mode,omitempty"`
	TriggerType        string               `json:"trigger_type,omitempty"`
	PolicyCreator      string               `json:"policy_creator,omitempty"`
	ExecutionTimestamp int64                `json:"execution_timestamp,omitempty"`
	SrcResource        *ReplicationResource `json:"src_resource,omitempty"`
	DestResource       *ReplicationResource `json:"dest_resource,omitempty"`
	SuccessfulArtifact []*ArtifactInfo      `json:"successful_artifact,omitempty"`
	FailedArtifact     []*ArtifactInfo      `json:"failed_artifact,omitempty"`
}

// ArtifactInfo describe info of artifact replicated
type ArtifactInfo struct {
	Type       string `json:"type"`
	Status     string `json:"status"`
	NameAndTag string `json:"name_tag"`
	FailReason string `json:"fail_reason,omitempty"`
}

// ReplicationResource describes replication resource info
type ReplicationResource struct {
	RegistryName string `json:"registry_name,omitempty"`
	RegistryType string `json:"registry_type"`
	Endpoint     string `json:"endpoint"`
	Provider     string `json:"provider,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
}
