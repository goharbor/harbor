// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/rbac"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"time"

	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/audit/model"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// the event consumers can refer to this file to find all topics and the corresponding event structures

// const definition
const (
	TopicCreateProject     = "CREATE_PROJECT"
	TopicDeleteProject     = "DELETE_PROJECT"
	TopicPushArtifact      = "PUSH_ARTIFACT"
	TopicPullArtifact      = "PULL_ARTIFACT"
	TopicDeleteArtifact    = "DELETE_ARTIFACT"
	TopicDeleteRepository  = "DELETE_REPOSITORY"
	TopicCreateTag         = "CREATE_TAG"
	TopicDeleteTag         = "DELETE_TAG"
	TopicScanningFailed    = "SCANNING_FAILED"
	TopicScanningStopped   = "SCANNING_STOPPED"
	TopicScanningCompleted = "SCANNING_COMPLETED"
	// QuotaExceedTopic is topic for quota warning event, the usage reaches the warning bar of limitation, like 85%
	TopicQuotaWarning    = "QUOTA_WARNING"
	TopicQuotaExceed     = "QUOTA_EXCEED"
	TopicUploadChart     = "UPLOAD_CHART"
	TopicDownloadChart   = "DOWNLOAD_CHART"
	TopicDeleteChart     = "DELETE_CHART"
	TopicReplication     = "REPLICATION"
	TopicArtifactLabeled = "ARTIFACT_LABELED"
	TopicTagRetention    = "TAG_RETENTION"
)

// CreateProjectEvent is the creating project event
type CreateProjectEvent struct {
	EventType string
	ProjectID int64
	Project   string
	Operator  string
	OccurAt   time.Time
}

// ResolveToAuditLog ...
func (c *CreateProjectEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    c.ProjectID,
		OpTime:       c.OccurAt,
		Operation:    rbac.ActionCreate.String(),
		Username:     c.Operator,
		ResourceType: "project",
		Resource:     c.Project}
	return auditLog, nil
}

func (c *CreateProjectEvent) String() string {
	return fmt.Sprintf("ID-%d Name-%s Operator-%s OccurAt-%s",
		c.ProjectID, c.Project, c.Operator, c.OccurAt.Format("2006-01-02 15:04:05"))
}

// DeleteProjectEvent is the deleting project event
type DeleteProjectEvent struct {
	EventType string
	ProjectID int64
	Project   string
	Operator  string
	OccurAt   time.Time
}

// ResolveToAuditLog ...
func (d *DeleteProjectEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    rbac.ActionDelete.String(),
		Username:     d.Operator,
		ResourceType: "project",
		Resource:     d.Project}
	return auditLog, nil
}

func (d *DeleteProjectEvent) String() string {
	return fmt.Sprintf("ID-%d Name-%s Operator-%s OccurAt-%s",
		d.ProjectID, d.Project, d.Operator, d.OccurAt.Format("2006-01-02 15:04:05"))
}

// DeleteRepositoryEvent is the deleting repository event
type DeleteRepositoryEvent struct {
	EventType  string
	ProjectID  int64
	Repository string
	Operator   string
	OccurAt    time.Time
}

// ResolveToAuditLog ...
func (d *DeleteRepositoryEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    rbac.ActionDelete.String(),
		Username:     d.Operator,
		ResourceType: "repository",
		Resource:     d.Repository,
	}
	return auditLog, nil
}

func (d *DeleteRepositoryEvent) String() string {
	return fmt.Sprintf("ID-%d Repository-%s Operator-%s OccurAt-%s",
		d.ProjectID, d.Repository, d.Operator, d.OccurAt.Format("2006-01-02 15:04:05"))
}

// ArtifactEvent is the pushing/pulling artifact event
type ArtifactEvent struct {
	EventType  string
	Repository string
	Artifact   *artifact.Artifact
	Tags       []string // when the artifact is pushed by digest, the tag here will be null
	Operator   string
	OccurAt    time.Time
}

func (a *ArtifactEvent) String() string {
	return fmt.Sprintf("ID-%d, Repository-%s Tags-%s Digest-%s Operator-%s OccurAt-%s",
		a.Artifact.ID, a.Repository, a.Tags, a.Artifact.Digest, a.Operator,
		a.OccurAt.Format("2006-01-02 15:04:05"))
}

// PushArtifactEvent is the pushing artifact event
type PushArtifactEvent struct {
	*ArtifactEvent
}

// ResolveToAuditLog ...
func (p *PushArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    p.Artifact.ProjectID,
		OpTime:       p.OccurAt,
		Operation:    rbac.ActionCreate.String(),
		Username:     p.Operator,
		ResourceType: "artifact"}

	if len(p.Tags) == 0 {
		auditLog.Resource = fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Artifact.Digest)
	} else {
		auditLog.Resource = fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Tags[0])
	}

	return auditLog, nil
}

func (p *PushArtifactEvent) String() string {
	return p.ArtifactEvent.String()
}

// PullArtifactEvent is the pulling artifact event
type PullArtifactEvent struct {
	*ArtifactEvent
}

// ResolveToAuditLog ...
func (p *PullArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    p.Artifact.ProjectID,
		OpTime:       p.OccurAt,
		Operation:    rbac.ActionPull.String(),
		Username:     p.Operator,
		ResourceType: "artifact"}

	if len(p.Tags) == 0 {
		auditLog.Resource = fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Artifact.Digest)
	} else {
		auditLog.Resource = fmt.Sprintf("%s:%s",
			p.Artifact.RepositoryName, p.Tags[0])
	}

	// for pull public resource
	if p.Operator == "" {
		auditLog.Username = "anonymous"
	} else {
		auditLog.Username = p.Operator
	}

	return auditLog, nil
}

func (p *PullArtifactEvent) String() string {
	return p.ArtifactEvent.String()
}

// DeleteArtifactEvent is the deleting artifact event
type DeleteArtifactEvent struct {
	*ArtifactEvent
}

// ResolveToAuditLog ...
func (d *DeleteArtifactEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.Artifact.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    rbac.ActionDelete.String(),
		Username:     d.Operator,
		ResourceType: "artifact",
		Resource:     fmt.Sprintf("%s:%s", d.Artifact.RepositoryName, d.Artifact.Digest)}
	return auditLog, nil
}

func (d *DeleteArtifactEvent) String() string {
	return d.ArtifactEvent.String()
}

// CreateTagEvent is the creating tag event
type CreateTagEvent struct {
	EventType        string
	Repository       string
	Tag              string
	AttachedArtifact *artifact.Artifact
	Operator         string
	OccurAt          time.Time
}

// ResolveToAuditLog ...
func (c *CreateTagEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    c.AttachedArtifact.ProjectID,
		OpTime:       c.OccurAt,
		Operation:    rbac.ActionCreate.String(),
		Username:     c.Operator,
		ResourceType: "tag",
		Resource:     fmt.Sprintf("%s:%s", c.Repository, c.Tag)}
	return auditLog, nil
}

func (c *CreateTagEvent) String() string {
	return fmt.Sprintf("ArtifactID-%d, Repository-%s Tag-%s Digest-%s Operator-%s OccurAt-%s",
		c.AttachedArtifact.ID, c.Repository, c.Tag, c.AttachedArtifact.Digest, c.Operator,
		c.OccurAt.Format("2006-01-02 15:04:05"))
}

// DeleteTagEvent is the deleting tag event
type DeleteTagEvent struct {
	EventType        string
	Repository       string
	Tag              string
	AttachedArtifact *artifact.Artifact
	Operator         string
	OccurAt          time.Time
}

// ResolveToAuditLog ...
func (d *DeleteTagEvent) ResolveToAuditLog() (*model.AuditLog, error) {
	auditLog := &model.AuditLog{
		ProjectID:    d.AttachedArtifact.ProjectID,
		OpTime:       d.OccurAt,
		Operation:    rbac.ActionDelete.String(),
		Username:     d.Operator,
		ResourceType: "tag",
		Resource:     fmt.Sprintf("%s:%s", d.Repository, d.Tag)}
	return auditLog, nil
}

func (d *DeleteTagEvent) String() string {
	return fmt.Sprintf("ArtifactID-%d, Repository-%s Tag-%s Digest-%s Operator-%s OccurAt-%s",
		d.AttachedArtifact.ID, d.Repository, d.Tag, d.AttachedArtifact.Digest, d.Operator,
		d.OccurAt.Format("2006-01-02 15:04:05"))
}

// ScanImageEvent is scanning image related event data to publish
type ScanImageEvent struct {
	EventType string
	Artifact  *v1.Artifact
	OccurAt   time.Time
	Operator  string
}

func (s *ScanImageEvent) String() string {
	return fmt.Sprintf("Artifact-%+v Operator-%s OccurAt-%s",
		s.Artifact, s.Operator, s.OccurAt.Format("2006-01-02 15:04:05"))
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

func (c *ChartEvent) String() string {
	return fmt.Sprintf("ProjectName-%s ChartName-%s Versions-%s Operator-%s OccurAt-%s",
		c.ProjectName, c.ChartName, c.Versions, c.Operator, c.OccurAt.Format("2006-01-02 15:04:05"))
}

// QuotaEvent is project quota related event data to publish
type QuotaEvent struct {
	EventType string
	Project   *proModels.Project
	Resource  *ImgResource
	OccurAt   time.Time
	RepoName  string
	Msg       string
}

func (q *QuotaEvent) String() string {
	return fmt.Sprintf("ProjectID-%d RepoName-%s Resource-%+v Msg-%s OccurAt-%s",
		q.Project.ProjectID, q.RepoName, q.Resource, q.Msg, q.OccurAt.Format("2006-01-02 15:04:05"))
}

// ImgResource include image digest and tag
type ImgResource struct {
	Digest string
	Tag    string
}

// ReplicationEvent is replication related event data to publish
type ReplicationEvent struct {
	EventType         string
	ReplicationTaskID int64
	OccurAt           time.Time
	Status            string
}

func (r *ReplicationEvent) String() string {
	return fmt.Sprintf("ReplicationTaskID-%d Status-%s OccurAt-%s",
		r.ReplicationTaskID, r.Status, r.OccurAt.Format("2006-01-02 15:04:05"))
}

// ArtifactLabeledEvent is event data of artifact labeled
type ArtifactLabeledEvent struct {
	ArtifactID int64
	LabelID    int64
	OccurAt    time.Time
	Operator   string
}

func (al *ArtifactLabeledEvent) String() string {
	return fmt.Sprintf("ArtifactID-%d LabelID-%d Operator-%s OccurAt-%s",
		al.ArtifactID, al.LabelID, al.Operator, al.OccurAt.Format("2006-01-02 15:04:05"))
}

// RetentionEvent is tag retention related event data to publish
type RetentionEvent struct {
	TaskID    int64
	EventType string
	OccurAt   time.Time
	Status    string
	Deleted   []*selector.Result
}

func (r *RetentionEvent) String() string {
	candidates := []string{}
	for _, candidate := range r.Deleted {
		candidates = append(candidates, fmt.Sprintf("%s:%s:%s", candidate.Target.Namespace,
			candidate.Target.Repository, candidate.Target.Tags))
	}

	return fmt.Sprintf("TaskID-%d Status-%s Deleted-%s OccurAt-%s",
		r.TaskID, r.Status, candidates, r.OccurAt.Format("2006-01-02 15:04:05"))
}
