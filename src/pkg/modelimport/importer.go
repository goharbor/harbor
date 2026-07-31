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

package modelimport

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/goharbor/harbor/src/common/utils"
	artifactctl "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/modelimport/huggingface"
	"github.com/goharbor/harbor/src/pkg/modelimport/model"
	"github.com/goharbor/harbor/src/pkg/modelimport/oci"
	repomodel "github.com/goharbor/harbor/src/pkg/repository/model"
)

// ImportRequest contains all inputs needed by one import job.
type ImportRequest struct {
	PolicyID           int64
	ProjectID          int64
	TargetRepository   string
	TargetTag          string
	HuggingFaceRepoID  string
	Revision           string
	Subpath            string
	Token              string
	LastSyncedCommit   string
	LastArtifactDigest string
}

// Progress is emitted through JobService check-ins.
type Progress struct {
	Stage          string `json:"stage"`
	Message        string `json:"message,omitempty"`
	CommitSHA      string `json:"commit_sha,omitempty"`
	ArtifactDigest string `json:"artifact_digest,omitempty"`
	Files          int    `json:"files,omitempty"`
}

// ProgressFunc receives import progress.
type ProgressFunc func(Progress)

// Importer imports models.
type Importer interface {
	Import(ctx context.Context, req *ImportRequest, progress ProgressFunc) error
}

type importer struct {
	hfClient  huggingface.Client
	converter *oci.Converter
	policyMgr Manager
}

// NewImporter returns the default importer.
func NewImporter() Importer {
	return &importer{
		hfClient:  huggingface.NewClient(),
		converter: oci.New(nil),
		policyMgr: Mgr,
	}
}

func (i *importer) Import(ctx context.Context, req *ImportRequest, progress ProgressFunc) error {
	if err := validateRequest(req); err != nil {
		return err
	}
	checkin(progress, Progress{Stage: "resolve", Message: "resolving Hugging Face model"})
	snapshot, err := i.hfClient.Snapshot(ctx, req.HuggingFaceRepoID, req.Revision, req.Subpath, req.Token)
	if err != nil {
		return err
	}
	if req.LastSyncedCommit != "" && req.LastSyncedCommit == snapshot.CommitSHA {
		checkin(progress, Progress{Stage: "noop", Message: "model is already in sync", CommitSHA: snapshot.CommitSHA})
		return nil
	}

	checkin(progress, Progress{Stage: "push", Message: "pushing OCI model artifact", CommitSHA: snapshot.CommitSHA, Files: len(snapshot.Files)})
	if err := ensureRepository(ctx, req.ProjectID, req.TargetRepository); err != nil {
		return err
	}
	result, err := i.converter.Push(ctx, req.TargetRepository, req.TargetTag, snapshot, func(ctx context.Context, filename string) (io.ReadCloser, error) {
		sourceName := filename
		if req.Subpath != "" {
			sourceName = req.Subpath + "/" + filename
		}
		return i.hfClient.OpenFile(ctx, req.HuggingFaceRepoID, snapshot.Revision, sourceName, req.Token)
	})
	if err != nil {
		return err
	}
	if _, _, err = artifactctl.Ctl.Ensure(ctx, req.TargetRepository, result.Digest, &artifactctl.ArtOption{
		Tags: []string{req.TargetTag},
	}); err != nil {
		return err
	}
	if req.PolicyID > 0 {
		if err = i.updatePolicySyncState(ctx, req.PolicyID, snapshot.CommitSHA, result.Digest); err != nil {
			return err
		}
	}
	checkin(progress, Progress{
		Stage:          "complete",
		Message:        "model import completed",
		CommitSHA:      snapshot.CommitSHA,
		ArtifactDigest: result.Digest,
		Files:          len(snapshot.Files),
	})
	return nil
}

func (i *importer) updatePolicySyncState(ctx context.Context, policyID int64, commit, artifactDigest string) error {
	policy, err := i.policyMgr.Get(ctx, policyID)
	if err != nil {
		return err
	}
	policy.LastSyncedCommit = commit
	policy.LastArtifactDigest = artifactDigest
	policy.UpdateTime = time.Now()
	return i.policyMgr.Update(ctx, policy, "LastSyncedCommit", "LastArtifactDigest", "UpdateTime")
}

func validateRequest(req *ImportRequest) error {
	if req == nil {
		return errors.New("nil model import request")
	}
	if req.ProjectID == 0 {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("project_id is required")
	}
	if req.TargetRepository == "" || req.HuggingFaceRepoID == "" {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("target_repository and hf_repo_id are required")
	}
	if req.TargetTag == "" {
		req.TargetTag = "latest"
	}
	if req.Revision == "" {
		req.Revision = "main"
	}
	return nil
}

func ensureRepository(ctx context.Context, projectID int64, repository string) error {
	projectName, _ := utils.ParseRepository(repository)
	if projectName == "" {
		return fmt.Errorf("target repository %q must include project namespace", repository)
	}
	if _, err := pkg.RepositoryMgr.GetByName(ctx, repository); err == nil {
		return nil
	} else if !errors.IsNotFoundErr(err) {
		return err
	}
	_, err := pkg.RepositoryMgr.Create(ctx, &repomodel.RepoRecord{
		Name:      repository,
		ProjectID: projectID,
	})
	if errors.IsConflictErr(err) {
		return nil
	}
	return err
}

func checkin(progress ProgressFunc, status Progress) {
	if progress != nil {
		progress(status)
	}
}

// NewPolicy returns a default policy for callers that need sensible defaults.
func NewPolicy() *model.Policy {
	return &model.Policy{
		Enabled: true,
		Trigger: &model.Trigger{
			Type: model.TriggerTypeManual,
		},
		TargetTag:           "latest",
		HuggingFaceRevision: "main",
	}
}
