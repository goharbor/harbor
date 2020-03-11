package model

import (
	"time"

	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/audit/model"
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

// HookEvent is hook related event data to publish
type HookEvent struct {
	PolicyID  int64
	EventType string
	Target    *models.EventTarget
	Payload   *Payload
}

// ProjectEvent info of Project related event
type ProjectEvent struct {
	// EventType - create/delete event
	TargetTopic string
	Project     *models.Project
	OccurAt     time.Time
	Operator    string
	Operation   string
}

// ResolveToAuditLog ...
func (p *ProjectEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    p.Project.ProjectID,
		OpTime:       p.OccurAt,
		Operation:    p.Operation,
		Username:     p.Operator,
		ResourceType: "project",
		Resource: fmt.Sprintf("/api/project/%v",
			p.Project.ProjectID)}
	return auditLog, nil
}

// Topic ...
func (p *ProjectEvent) Topic() string {
	return p.TargetTopic
}

// RepositoryEvent info of repository related event
type RepositoryEvent struct {
	TargetTopic string
	Project     *models.Project
	RepoName    string
	OccurAt     time.Time
	Operator    string
	Operation   string
}

// ResolveToAuditLog ...
func (r *RepositoryEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    r.Project.ProjectID,
		OpTime:       r.OccurAt,
		Operation:    r.Operation,
		Username:     r.Operator,
		ResourceType: "repository",
		Resource: fmt.Sprintf("/api/project/%v/repository/%v",
			r.Project.ProjectID, r.RepoName)}
	return auditLog, nil
}

// Topic ...
func (r *RepositoryEvent) Topic() string {
	return r.TargetTopic
}

// ArtifactEvent info of artifact related event
type ArtifactEvent struct {
	TargetTopic string
	Project     *models.Project
	RepoName    string
	Digest      string
	OccurAt     time.Time
	Operator    string
	Operation   string
}

// ResolveToAuditLog ...
func (a *ArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    a.Project.ProjectID,
		OpTime:       a.OccurAt,
		Operation:    a.Operation,
		Username:     a.Operator,
		ResourceType: "artifact",
		Resource: fmt.Sprintf("/api/project/%v/repository/%v/artifact/%v",
			a.Project.ProjectID, a.RepoName, a.Digest)}
	return auditLog, nil
}

// Topic ...
func (a *ArtifactEvent) Topic() string {
	return a.TargetTopic
}

// TagEvent info of tag related event
type TagEvent struct {
	TargetTopic string
	Project     *models.Project
	RepoName    string
	TagName     string
	Digest      string
	OccurAt     time.Time
	Operation   string
	Operator    string
}

// ResolveToAuditLog ...
func (t *TagEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    t.Project.ProjectID,
		OpTime:       t.OccurAt,
		Operation:    t.Operation,
		Username:     t.Operator,
		ResourceType: "tag",
		Resource: fmt.Sprintf("/api/project/%v/repository/%v/tag/%v",
			t.Project.ProjectID, t.RepoName, t.TagName)}
	log.Infof("create audit log %+v", auditLog)
	return auditLog, nil
}

// Topic ...
func (t *TagEvent) Topic() string {
	return t.TargetTopic
}

// ToImageEvent ...
func (t *TagEvent) ToImageEvent() *ImageEvent {
	var eventType string
	// convert tag operation to previous event type so that webhook can handle it
	if t.Operation == "push" {
		eventType = EventTypePushImage
	} else if t.Operation == "pull" {
		eventType = EventTypePullImage
	} else if t.Operation == "delete" {
		eventType = EventTypeDeleteImage
	}
	imgEvent := &ImageEvent{
		EventType: eventType,
		Project:   t.Project,
		Resource:  []*ImgResource{{Tag: t.TagName}},
		Operator:  t.Operator,
		OccurAt:   t.OccurAt,
		RepoName:  t.RepoName,
	}
	return imgEvent
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
	Resources  []*Resource       `json:"resources"`
	Repository *Repository       `json:"repository"`
	Custom     map[string]string `json:"custom_attributes,omitempty"`
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
