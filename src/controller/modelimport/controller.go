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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/event/operator"
	projectctl "github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	modelimportpkg "github.com/goharbor/harbor/src/pkg/modelimport"
	"github.com/goharbor/harbor/src/pkg/modelimport/model"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

const callbackFuncName = "MODEL_IMPORT_CALLBACK"

var (
	// Ctl is the global model import controller.
	Ctl = NewController()
)

func init() {
	callbackFunc := func(ctx context.Context, param string) error {
		params := map[string]any{}
		if err := json.Unmarshal([]byte(param), &params); err != nil {
			return err
		}
		var policyID int64
		if id, ok := params["policy_id"].(float64); ok {
			policyID = int64(id)
		}
		if op, ok := params["operator"].(string); ok {
			ctx = context.WithValue(ctx, operator.ContextKey{}, op)
		}
		_, err := Ctl.Start(ctx, policyID, task.ExecutionTriggerSchedule)
		return err
	}
	if err := scheduler.RegisterCallbackFunc(callbackFuncName, callbackFunc); err != nil {
		log.Errorf("failed to register the callback function for model import: %v", err)
	}
	if err := task.RegisterCheckInProcessor(job.ModelImportVendorType, modelImportCheckIn); err != nil {
		log.Fatalf("failed to register the checkin processor for the model import job, error %v", err)
	}
}

// Controller manages Hugging Face model import policies and executions.
type Controller interface {
	CountPolicies(ctx context.Context, query *q.Query) (int64, error)
	ListPolicies(ctx context.Context, query *q.Query) ([]*model.Policy, error)
	ListPoliciesByProject(ctx context.Context, projectID int64, query *q.Query) ([]*model.Policy, error)
	GetPolicy(ctx context.Context, id int64) (*model.Policy, error)
	GetPolicyByName(ctx context.Context, projectID int64, name string) (*model.Policy, error)
	CreatePolicy(ctx context.Context, policy *model.Policy) (int64, error)
	UpdatePolicy(ctx context.Context, policy *model.Policy, props ...string) error
	DeletePolicy(ctx context.Context, id int64) error
	StartImport(ctx context.Context, req *modelimportpkg.ImportRequest, trigger string) (int64, error)
	Start(ctx context.Context, policyID int64, trigger string) (int64, error)
	Stop(ctx context.Context, executionID int64) error
	ExecutionCount(ctx context.Context, query *q.Query) (int64, error)
	ListExecutions(ctx context.Context, query *q.Query) ([]*task.Execution, error)
	GetExecution(ctx context.Context, executionID int64) (*task.Execution, error)
	TaskCount(ctx context.Context, query *q.Query) (int64, error)
	ListTasks(ctx context.Context, query *q.Query) ([]*task.Task, error)
	GetTask(ctx context.Context, taskID int64) (*task.Task, error)
	GetTaskLog(ctx context.Context, taskID int64) ([]byte, error)
}

type controller struct {
	policyMgr  modelimportpkg.Manager
	projectCtl projectctl.Controller
	taskMgr    task.Manager
	execMgr    task.ExecutionManager
	scheduler  scheduler.Scheduler
}

// NewController returns a model import controller.
func NewController() Controller {
	return &controller{
		policyMgr:  modelimportpkg.Mgr,
		projectCtl: projectctl.Ctl,
		taskMgr:    task.Mgr,
		execMgr:    task.ExecMgr,
		scheduler:  scheduler.Sched,
	}
}

func (c *controller) CountPolicies(ctx context.Context, query *q.Query) (int64, error) {
	return c.policyMgr.Count(ctx, query)
}

func (c *controller) ListPolicies(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	return c.policyMgr.List(ctx, query)
}

func (c *controller) ListPoliciesByProject(ctx context.Context, projectID int64, query *q.Query) ([]*model.Policy, error) {
	return c.policyMgr.ListByProject(ctx, projectID, query)
}

func (c *controller) GetPolicy(ctx context.Context, id int64) (*model.Policy, error) {
	policy, err := c.policyMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	policy.PlainHFToken = ""
	return policy, nil
}

func (c *controller) GetPolicyByName(ctx context.Context, projectID int64, name string) (*model.Policy, error) {
	policy, err := c.policyMgr.GetByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}
	policy.PlainHFToken = ""
	return policy, nil
}

func (c *controller) CreatePolicy(ctx context.Context, policy *model.Policy) (int64, error) {
	if err := c.preparePolicyForSave(ctx, policy, nil); err != nil {
		return 0, err
	}
	id, err := c.policyMgr.Create(ctx, policy)
	if err != nil {
		return 0, err
	}
	policy.ID = id
	if err := c.ensureSchedule(ctx, policy); err != nil {
		return 0, err
	}
	return id, nil
}

func (c *controller) UpdatePolicy(ctx context.Context, policy *model.Policy, props ...string) error {
	existing, err := c.policyMgr.Get(ctx, policy.ID)
	if err != nil {
		return err
	}
	if err := c.scheduler.UnScheduleByVendor(ctx, job.ModelImportVendorType, policy.ID); err != nil {
		return err
	}
	if err := c.preparePolicyForSave(ctx, policy, existing); err != nil {
		return err
	}
	if err := c.policyMgr.Update(ctx, policy, props...); err != nil {
		return err
	}
	return c.ensureSchedule(ctx, policy)
}

func (c *controller) DeletePolicy(ctx context.Context, id int64) error {
	if err := c.scheduler.UnScheduleByVendor(ctx, job.ModelImportVendorType, id); err != nil {
		return err
	}
	if err := c.execMgr.DeleteByVendor(ctx, job.ModelImportVendorType, id); err != nil {
		return err
	}
	return c.policyMgr.Delete(ctx, id)
}

func (c *controller) Start(ctx context.Context, policyID int64, trigger string) (int64, error) {
	policy, err := c.policyMgr.Get(ctx, policyID)
	if err != nil {
		return 0, err
	}
	if !policy.Enabled {
		return 0, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("model import policy is disabled")
	}
	if err := c.normalizeTargetRepository(ctx, policy); err != nil {
		return 0, err
	}
	plainToken, err := decryptToken(policy.HuggingFaceToken)
	if err != nil {
		return 0, err
	}
	return c.startImport(ctx, &modelimportpkg.ImportRequest{
		PolicyID:           policy.ID,
		ProjectID:          policy.ProjectID,
		TargetRepository:   policy.TargetRepository,
		TargetTag:          policy.TargetTag,
		HuggingFaceRepoID:  policy.HuggingFaceRepoID,
		Revision:           policy.HuggingFaceRevision,
		Subpath:            policy.HuggingFaceSubpath,
		Token:              plainToken,
		LastSyncedCommit:   policy.LastSyncedCommit,
		LastArtifactDigest: policy.LastArtifactDigest,
	}, trigger)
}

func (c *controller) StartImport(ctx context.Context, req *modelimportpkg.ImportRequest, trigger string) (int64, error) {
	if req == nil {
		return 0, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("model import request is required")
	}
	if err := c.normalizeRequestTargetRepository(ctx, req); err != nil {
		return 0, err
	}
	return c.startImport(ctx, req, trigger)
}

func (c *controller) startImport(ctx context.Context, req *modelimportpkg.ImportRequest, trigger string) (int64, error) {
	execAttrs := sanitizedRequestParams(req)
	execID, err := c.execMgr.Create(ctx, job.ModelImportVendorType, req.PolicyID, trigger, execAttrs)
	if err != nil {
		return 0, err
	}
	params := sanitizedRequestParams(req)
	params["hf_token"] = req.Token
	_, err = c.taskMgr.Create(ctx, execID, &task.Job{
		Name: job.ModelImportVendorType,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
	}, execAttrs)
	if err != nil {
		return 0, err
	}
	return execID, nil
}

func (c *controller) Stop(ctx context.Context, executionID int64) error {
	return c.execMgr.Stop(ctx, executionID)
}

func (c *controller) ExecutionCount(ctx context.Context, query *q.Query) (int64, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.ModelImportVendorType
	return c.execMgr.Count(ctx, query)
}

func (c *controller) ListExecutions(ctx context.Context, query *q.Query) ([]*task.Execution, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.ModelImportVendorType
	return c.execMgr.List(ctx, query)
}

func (c *controller) GetExecution(ctx context.Context, executionID int64) (*task.Execution, error) {
	exec, err := c.execMgr.Get(ctx, executionID)
	if err != nil {
		return nil, err
	}
	if exec.VendorType != job.ModelImportVendorType {
		return nil, errors.NotFoundError(nil).WithMessagef("model import execution %d not found", executionID)
	}
	return exec, nil
}

func (c *controller) TaskCount(ctx context.Context, query *q.Query) (int64, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.ModelImportVendorType
	return c.taskMgr.Count(ctx, query)
}

func (c *controller) ListTasks(ctx context.Context, query *q.Query) ([]*task.Task, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.ModelImportVendorType
	return c.taskMgr.List(ctx, query)
}

func (c *controller) GetTask(ctx context.Context, taskID int64) (*task.Task, error) {
	t, err := c.taskMgr.Get(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t.VendorType != job.ModelImportVendorType {
		return nil, errors.NotFoundError(nil).WithMessagef("model import task %d not found", taskID)
	}
	return t, nil
}

func (c *controller) GetTaskLog(ctx context.Context, taskID int64) ([]byte, error) {
	if _, err := c.GetTask(ctx, taskID); err != nil {
		return nil, err
	}
	return c.taskMgr.GetLog(ctx, taskID)
}

func (c *controller) preparePolicyForSave(ctx context.Context, policy *model.Policy, existing *model.Policy) error {
	if existing != nil && len(policy.PlainHFToken) == 0 {
		policy.HuggingFaceToken = existing.HuggingFaceToken
	}
	if len(policy.PlainHFToken) > 0 {
		encrypted, err := encrypt.Instance().Encrypt(policy.PlainHFToken)
		if err != nil {
			return err
		}
		policy.HuggingFaceToken = encrypted
	}
	policy.PlainHFToken = ""
	if err := c.normalizeTargetRepository(ctx, policy); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}
	return nil
}

func (c *controller) normalizeTargetRepository(ctx context.Context, policy *model.Policy) error {
	if policy == nil {
		return errors.New("nil model import policy")
	}
	project, err := c.projectCtl.Get(ctx, policy.ProjectID)
	if err != nil {
		return err
	}
	repository, err := normalizeTargetRepository(project.Name, policy.TargetRepository)
	if err != nil {
		return err
	}
	policy.TargetRepository = repository
	return nil
}

func (c *controller) normalizeRequestTargetRepository(ctx context.Context, req *modelimportpkg.ImportRequest) error {
	project, err := c.projectCtl.Get(ctx, req.ProjectID)
	if err != nil {
		return err
	}
	repository, err := normalizeTargetRepository(project.Name, req.TargetRepository)
	if err != nil {
		return err
	}
	req.TargetRepository = repository
	return nil
}

func normalizeTargetRepository(projectName, repository string) (string, error) {
	repository = strings.Trim(strings.TrimSpace(repository), "/")
	if repository == "" {
		return "", errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("target repository is required")
	}
	project, rest := utils.ParseRepository(repository)
	if project == "" {
		return fmt.Sprintf("%s/%s", projectName, rest), nil
	}
	if project != projectName {
		return "", errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("target repository %q must be under project %q", repository, projectName)
	}
	return repository, nil
}

func modelImportCheckIn(ctx context.Context, t *task.Task, sc *job.StatusChange) error {
	var progress modelimportpkg.Progress
	if err := json.Unmarshal([]byte(sc.CheckIn), &progress); err != nil {
		log.Errorf("failed to resolve checkin of model import task %d: %v", t.ID, err)
		return err
	}

	current, err := task.Mgr.Get(ctx, t.ID)
	if err != nil {
		return err
	}
	if current.ExtraAttrs == nil {
		current.ExtraAttrs = map[string]any{}
	}
	current.ExtraAttrs["progress_stage"] = progress.Stage
	current.ExtraAttrs["progress_message"] = progress.Message
	if progress.CommitSHA != "" {
		current.ExtraAttrs["commit_sha"] = progress.CommitSHA
	}
	if progress.ArtifactDigest != "" {
		current.ExtraAttrs["artifact_digest"] = progress.ArtifactDigest
	}
	if progress.Files > 0 {
		current.ExtraAttrs["files"] = progress.Files
	}

	return task.Mgr.UpdateExtraAttrs(ctx, t.ID, current.ExtraAttrs)
}

func (c *controller) ensureSchedule(ctx context.Context, policy *model.Policy) error {
	if policy.Trigger == nil || policy.Trigger.Type != model.TriggerTypeScheduled || len(policy.Trigger.Settings.Cron) == 0 {
		return nil
	}
	cbParams := map[string]any{
		"policy_id": policy.ID,
		"operator":  secret.JobserviceUser,
	}
	_, err := c.scheduler.Schedule(ctx, job.ModelImportVendorType, policy.ID, "", policy.Trigger.Settings.Cron,
		callbackFuncName, cbParams, map[string]any{})
	return err
}

func sanitizedRequestParams(req *modelimportpkg.ImportRequest) map[string]any {
	return map[string]any{
		"policy_id":            req.PolicyID,
		"project_id":           req.ProjectID,
		"target_repository":    req.TargetRepository,
		"target_tag":           req.TargetTag,
		"hf_repo_id":           req.HuggingFaceRepoID,
		"hf_revision":          req.Revision,
		"hf_subpath":           req.Subpath,
		"last_synced_commit":   req.LastSyncedCommit,
		"last_artifact_digest": req.LastArtifactDigest,
	}
}

func decryptToken(ciphertext string) (string, error) {
	if len(ciphertext) == 0 {
		return "", nil
	}
	return encrypt.Instance().Decrypt(ciphertext)
}
