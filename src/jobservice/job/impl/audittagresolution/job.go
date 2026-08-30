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

package audittagresolution

import (
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/auditext"
	"github.com/goharbor/harbor/src/pkg/tag"
)

const (
	ParamAuditLogID  = "audit_log_id"
	ParamRepository  = "repository"
	ParamDigest      = "digest"
)

// Job resolves tags for audit log entries that were recorded with a digest
// instead of a tag (e.g. proxy cache pulls where the GET uses the digest).
type Job struct {
	auditMgr    auditext.Manager
	artifactMgr artifact.Manager
	tagMgr      tag.Manager
}

func (j *Job) MaxFails() uint {
	return 3
}

func (j *Job) MaxCurrency() uint {
	return 5
}

func (j *Job) ShouldRetry() bool {
	return true
}

func (j *Job) Validate(params job.Parameters) error {
	if params == nil {
		return errors.New("missing job parameters")
	}
	for _, key := range []string{ParamAuditLogID, ParamRepository, ParamDigest} {
		if _, ok := params[key]; !ok {
			return fmt.Errorf("missing required parameter %q", key)
		}
	}
	return nil
}

func (j *Job) init() {
	if j.auditMgr == nil {
		j.auditMgr = auditext.Mgr
	}
	if j.artifactMgr == nil {
		j.artifactMgr = artifact.NewManager()
	}
	if j.tagMgr == nil {
		j.tagMgr = tag.NewManager()
	}
}

func (j *Job) Run(ctx job.Context, params job.Parameters) error {
	j.init()
	logger := ctx.GetLogger()
	sysCtx := ctx.SystemContext()

	auditLogID, ok := params[ParamAuditLogID].(float64)
	if !ok {
		return fmt.Errorf("invalid type for %s", ParamAuditLogID)
	}
	repo, ok := params[ParamRepository].(string)
	if !ok {
		return fmt.Errorf("invalid type for %s", ParamRepository)
	}
	digest, ok := params[ParamDigest].(string)
	if !ok {
		return fmt.Errorf("invalid type for %s", ParamDigest)
	}

	art, err := j.artifactMgr.GetByDigest(sysCtx, repo, digest)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			logger.Infof("artifact %s@%s not found, skipping tag resolution", repo, digest)
			return nil
		}
		return err
	}

	tags, err := j.tagMgr.List(sysCtx, q.New(q.KeyWords{"ArtifactID": art.ID}))
	if err != nil {
		return err
	}
	if len(tags) == 0 {
		return nil
	}

	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}

	entry, err := j.auditMgr.Get(sysCtx, int64(auditLogID))
	if err != nil {
		return err
	}

	resolvedNote := fmt.Sprintf("resolved tags: %s", strings.Join(names, ", "))
	if entry.OperationDescription != "" {
		entry.OperationDescription += "; " + resolvedNote
	} else {
		entry.OperationDescription = resolvedNote
	}

	if err := j.auditMgr.Update(sysCtx, entry, "OperationDescription"); err != nil {
		return err
	}

	logger.Infof("resolved tags for audit log %d: %s", int64(auditLogID), strings.Join(names, ", "))
	return nil
}
