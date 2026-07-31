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
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/jobservice/job"
	modelimportpkg "github.com/goharbor/harbor/src/pkg/modelimport"
)

// Job imports and syncs Hugging Face models into Harbor as OCI model-spec artifacts.
type Job struct {
	importer modelimportpkg.Importer
}

// MaxFails implements job.Interface.
func (j *Job) MaxFails() uint {
	return 1
}

// MaxCurrency implements job.Interface.
func (j *Job) MaxCurrency() uint {
	return 2
}

// ShouldRetry implements job.Interface.
func (j *Job) ShouldRetry() bool {
	return false
}

// Validate implements job.Interface.
func (j *Job) Validate(params job.Parameters) error {
	if _, err := requestFromParams(params); err != nil {
		return err
	}
	return nil
}

// Run implements job.Interface.
func (j *Job) Run(ctx job.Context, params job.Parameters) error {
	logger := ctx.GetLogger()
	logger.Info("model import job starting")
	defer logger.Info("model import job finished")

	req, err := requestFromParams(params)
	if err != nil {
		return err
	}
	if j.importer == nil {
		j.importer = modelimportpkg.NewImporter()
	}
	checkin := func(status modelimportpkg.Progress) {
		data, err := json.Marshal(status)
		if err != nil {
			logger.Warningf("failed to marshal model import progress: %v", err)
			return
		}
		if err := ctx.Checkin(string(data)); err != nil {
			logger.Warningf("failed to check in model import progress: %v", err)
		}
	}
	return j.importer.Import(ctx.SystemContext(), req, checkin)
}

func requestFromParams(params job.Parameters) (*modelimportpkg.ImportRequest, error) {
	req := &modelimportpkg.ImportRequest{
		PolicyID:           int64FromParams(params, "policy_id"),
		ProjectID:          int64FromParams(params, "project_id"),
		TargetRepository:   stringFromParams(params, "target_repository"),
		TargetTag:          stringFromParams(params, "target_tag"),
		HuggingFaceRepoID:  stringFromParams(params, "hf_repo_id"),
		Revision:           stringFromParams(params, "hf_revision"),
		Subpath:            stringFromParams(params, "hf_subpath"),
		Token:              stringFromParams(params, "hf_token"),
		LastSyncedCommit:   stringFromParams(params, "last_synced_commit"),
		LastArtifactDigest: stringFromParams(params, "last_artifact_digest"),
	}
	if req.ProjectID == 0 {
		return nil, fmt.Errorf("missing project_id")
	}
	if req.TargetRepository == "" {
		return nil, fmt.Errorf("missing target_repository")
	}
	if req.TargetTag == "" {
		req.TargetTag = "latest"
	}
	if req.HuggingFaceRepoID == "" {
		return nil, fmt.Errorf("missing hf_repo_id")
	}
	if req.Revision == "" {
		req.Revision = "main"
	}
	return req, nil
}

func stringFromParams(params job.Parameters, key string) string {
	v, ok := params[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func int64FromParams(params job.Parameters, key string) int64 {
	v, ok := params[key]
	if !ok || v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case json.Number:
		i, _ := val.Int64()
		return i
	default:
		return 0
	}
}
