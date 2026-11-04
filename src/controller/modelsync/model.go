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

package modelsync

import (
	"regexp"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	pkgmodel "github.com/goharbor/harbor/src/pkg/modelsync/policy/model"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

// DefaultRevision is synced when a policy names no revision.
const DefaultRevision = "main"

var (
	commitRe = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	// a source ref becomes a tag the way the Hugging Face adapter names it
	invalidTagChars = regexp.MustCompile(`[^\w.-]`)
)

// Policy is the controller-level model sync policy, with the source registry and
// destination project populated.
type Policy struct {
	ID          int64
	Name        string
	Description string
	Creator     string
	Enabled     bool
	// Registry is the source registry, e.g. a Hugging Face endpoint.
	Registry *model.Registry
	// SrcRepository is the source model ID, e.g. "HuggingFaceTB/SmolLM2-135M".
	SrcRepository string
	// SrcRevision is the branch or tag to sync; empty means DefaultRevision.
	SrcRevision string
	// FileFilter selects the files of the model; nil selects all of them.
	FileFilter    *model.FileFilter
	DestProjectID int64
	// DestProjectName is populated when reading a policy.
	DestProjectName string
	// DestRepository is the repository inside the destination project; empty means the
	// lower-cased source repository.
	DestRepository     string
	TriggerType        string
	Cron               string
	LastSyncedRevision string
	CreationTime       time.Time
	UpdateTime         time.Time
}

// IsScheduledTrigger returns whether the policy has a cron trigger.
func (p *Policy) IsScheduledTrigger() bool {
	return p.TriggerType == pkgmodel.TriggerTypeScheduled
}

// Revision returns the branch or tag the policy syncs.
func (p *Policy) Revision() string {
	if r := strings.TrimSpace(p.SrcRevision); r != "" {
		return r
	}
	return DefaultRevision
}

// RevisionTagFilter returns the tag the source gives the synced revision.
func (p *Policy) RevisionTagFilter() string {
	return invalidTagChars.ReplaceAllString(p.Revision(), "-")
}

// EffectiveDestRepository returns the destination repository inside the project.
func (p *Policy) EffectiveDestRepository() string {
	if r := strings.Trim(strings.TrimSpace(p.DestRepository), "/"); r != "" {
		return r
	}
	return pkgmodel.DefaultDestRepository(p.SrcRepository)
}

// FullDestRepository returns "<project>/<repository>".
func (p *Policy) FullDestRepository() string {
	return p.DestProjectName + "/" + p.EffectiveDestRepository()
}

// Validate checks the fields that need no lookup.
func (p *Policy) Validate() error {
	if len(strings.TrimSpace(p.Name)) == 0 {
		return errors.BadRequestError(nil).WithMessage("name cannot be empty")
	}
	if len(p.Name) > 256 {
		return errors.BadRequestError(nil).WithMessage("name is longer than 256 characters")
	}
	if p.Registry == nil || p.Registry.ID <= 0 {
		return errors.BadRequestError(nil).WithMessage("source registry is required")
	}
	if len(strings.TrimSpace(p.SrcRepository)) == 0 {
		return errors.BadRequestError(nil).WithMessage("source repository cannot be empty")
	}
	if commitRe.MatchString(p.Revision()) {
		return errors.BadRequestError(nil).
			WithMessage("the revision must be a branch or tag: a sync copies what a ref points at, a pinned commit is not supported")
	}
	if p.DestProjectID <= 0 {
		return errors.BadRequestError(nil).WithMessage("destination project is required")
	}
	if err := p.FileFilter.Validate(); err != nil {
		return err
	}
	switch p.TriggerType {
	case pkgmodel.TriggerTypeManual:
		if strings.TrimSpace(p.Cron) != "" {
			return errors.BadRequestError(nil).WithMessage("cron must be empty for manual trigger")
		}
	case pkgmodel.TriggerTypeScheduled:
		if strings.TrimSpace(p.Cron) == "" {
			return errors.BadRequestError(nil).WithMessage("cron is required for scheduled trigger")
		}
		if err := utils.ValidateCronString(p.Cron); err != nil {
			return errors.BadRequestError(nil).WithMessagef("invalid cron %q: %v", p.Cron, err)
		}
	default:
		return errors.BadRequestError(nil).WithMessagef("invalid trigger type %q", p.TriggerType)
	}
	return nil
}

// From converts the DAO model to the controller model.
func (p *Policy) From(policy *pkgmodel.Policy) error {
	filter, err := policy.GetFileFilter()
	if err != nil {
		return err
	}
	p.ID = policy.ID
	p.Name = policy.Name
	p.Description = policy.Description
	p.Creator = policy.Creator
	p.Enabled = policy.Enabled
	p.Registry = &model.Registry{ID: policy.RegistryID}
	p.SrcRepository = policy.SrcRepository
	p.SrcRevision = policy.SrcRevision
	p.FileFilter = filter
	p.DestProjectID = policy.DestProjectID
	p.DestRepository = policy.DestRepository
	p.TriggerType = policy.TriggerType
	p.Cron = policy.Cron
	p.LastSyncedRevision = policy.LastSyncedRevision
	p.CreationTime = policy.CreationTime
	p.UpdateTime = policy.UpdateTime
	return nil
}

// To converts the controller model to the DAO model.
func (p *Policy) To() (*pkgmodel.Policy, error) {
	policy := &pkgmodel.Policy{
		ID:                 p.ID,
		Name:               strings.TrimSpace(p.Name),
		Description:        p.Description,
		Creator:            p.Creator,
		Enabled:            p.Enabled,
		SrcRepository:      strings.Trim(strings.TrimSpace(p.SrcRepository), "/"),
		SrcRevision:        strings.TrimSpace(p.SrcRevision),
		DestProjectID:      p.DestProjectID,
		DestRepository:     strings.Trim(strings.TrimSpace(p.DestRepository), "/"),
		TriggerType:        p.TriggerType,
		Cron:               strings.TrimSpace(p.Cron),
		LastSyncedRevision: p.LastSyncedRevision,
		CreationTime:       p.CreationTime,
		UpdateTime:         p.UpdateTime,
	}
	if p.Registry != nil {
		policy.RegistryID = p.Registry.ID
	}
	if err := policy.SetFileFilter(p.FileFilter); err != nil {
		return nil, err
	}
	return policy, nil
}

// Execution is a model sync execution.
type Execution struct {
	ID            int64
	PolicyID      int64
	Status        string
	StatusMessage string
	Metrics       *dao.Metrics
	Trigger       string
	Operator      string
	StartTime     time.Time
	EndTime       time.Time
}

// Task is the copy task of a model sync execution. The source facts (revision, digest, tags,
// file count and size) are resolved when the execution starts and are the same for every task
// of the execution.
type Task struct {
	ID             int64
	ExecutionID    int64
	Status         string
	StatusMessage  string
	RunCount       int32
	JobID          string
	SrcRepository  string
	SrcRevision    string
	DestRepository string
	Revision       string
	Digest         string
	Tags           []string
	Files          int64
	Size           int64
	// Skipped is true when the destination held the artifact before the run.
	Skipped      bool
	CreationTime time.Time
	StartTime    time.Time
	UpdateTime   time.Time
	EndTime      time.Time
}

// PreviewRequest resolves a source without copying it.
type PreviewRequest struct {
	RegistryID    int64
	SrcRepository string
	SrcRevision   string
	FileFilter    *model.FileFilter
}

// PreviewFile is one file of the preview.
type PreviewFile struct {
	Path    string
	Size    int64
	SHA256  string
	Type    string
	Matched bool
}

// Preview is the result of resolving a source.
type Preview struct {
	Repository   string
	Ref          string
	Revision     string
	RevisionTag  string
	Digest       string
	SourceURL    string
	Metadata     map[string]string
	Files        []*PreviewFile
	MatchedFiles int64
	MatchedSize  int64
	TotalFiles   int64
	TotalSize    int64
	Tags         []string
}
