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

// Package modelsync imports models from a model hub into a project. A model sync policy names
// one model, revision and file filter; each run resolves the source and then hands a copy of
// that one repository to the replication copy flow, so the transfer, the job and the
// execution/task bookkeeping are the replication ones under the vendor type MODEL_SYNC.
package modelsync

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/replication/flow"
	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/modelsync/policy"
	pkgmodel "github.com/goharbor/harbor/src/pkg/modelsync/policy/model"
	"github.com/goharbor/harbor/src/pkg/reg"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

const callbackFuncName = "MODEL_SYNC_CALLBACK"

// execution extra attributes written when the source is resolved
const (
	attrOperator       = "operator"
	attrSrcRepository  = "src_repository"
	attrSrcRevision    = "src_revision"
	attrDestRepository = "dest_repository"
	attrRevision       = "revision"
	attrDigest         = "digest"
	attrTags           = "tags"
	attrFiles          = "files"
	attrSize           = "size"
	attrSkipped        = "skipped"
	attrSourceURL      = "source_url"
)

// Ctl is the global model sync controller.
var Ctl = NewController()

// Controller defines the operations of model sync.
type Controller interface {
	PolicyCount(ctx context.Context, query *q.Query) (int64, error)
	ListPolicies(ctx context.Context, query *q.Query) ([]*Policy, error)
	GetPolicy(ctx context.Context, id int64) (*Policy, error)
	CreatePolicy(ctx context.Context, policy *Policy) (int64, error)
	UpdatePolicy(ctx context.Context, policy *Policy) error
	DeletePolicy(ctx context.Context, id int64) error
	// Start starts a sync of the policy and returns the execution ID. The source is resolved
	// and copied in the background.
	Start(ctx context.Context, policy *Policy, trigger string) (int64, error)
	Stop(ctx context.Context, executionID int64) error
	ExecutionCount(ctx context.Context, query *q.Query) (int64, error)
	ListExecutions(ctx context.Context, query *q.Query) ([]*Execution, error)
	GetExecution(ctx context.Context, executionID int64) (*Execution, error)
	TaskCount(ctx context.Context, query *q.Query) (int64, error)
	ListTasks(ctx context.Context, query *q.Query) ([]*Task, error)
	GetTask(ctx context.Context, taskID int64) (*Task, error)
	GetTaskLog(ctx context.Context, taskID int64) ([]byte, error)
	// Preview resolves a source and lists its files without copying anything.
	Preview(ctx context.Context, req *PreviewRequest) (*Preview, error)
	// ListAdapterTypes lists the registry types that can be a model sync source.
	ListAdapterTypes(ctx context.Context) []string
}

// NewController creates a model sync controller.
func NewController() Controller {
	return &controller{
		policyMgr:  policy.Mgr,
		execMgr:    task.ExecMgr,
		taskMgr:    task.Mgr,
		regMgr:     reg.Mgr,
		projectCtl: project.Ctl,
		artCtl:     artifact.Ctl,
		scheduler:  scheduler.Sched,
		flowCtl:    flow.NewController(),
		ormCreator: orm.Crt,
		wp:         lib.NewWorkerPool(10),
		resolverTypes: sync.OnceValue(func() []string {
			return resolverTypes(adp.ListRegisteredAdapterTypes(), adp.GetFactory)
		}),
	}
}

type controller struct {
	policyMgr     policy.Manager
	execMgr       task.ExecutionManager
	taskMgr       task.Manager
	regMgr        reg.Manager
	projectCtl    project.Controller
	artCtl        artifact.Controller
	scheduler     scheduler.Scheduler
	flowCtl       flow.Controller
	ormCreator    orm.Creator
	wp            *lib.WorkerPool
	resolverTypes func() []string
}

func init() {
	callbackFunc := func(ctx context.Context, param string) error {
		params := struct {
			PolicyID int64  `json:"policy_id"`
			Operator string `json:"operator"`
		}{}
		if err := json.Unmarshal([]byte(param), &params); err != nil {
			return err
		}
		if params.Operator != "" {
			ctx = context.WithValue(ctx, operator.ContextKey{}, params.Operator)
		}
		p, err := Ctl.GetPolicy(ctx, params.PolicyID)
		if err != nil {
			return err
		}
		_, err = Ctl.Start(ctx, p, task.ExecutionTriggerSchedule)
		return err
	}
	if err := scheduler.RegisterCallbackFunc(callbackFuncName, callbackFunc); err != nil {
		log.Errorf("failed to register the callback function for model sync: %v", err)
	}
	if err := task.RegisterTaskStatusChangePostFunc(job.ModelSyncVendorType, recordSyncedRevision); err != nil {
		log.Errorf("failed to register the task status change function for model sync: %v", err)
	}
}

// recordSyncedRevision moves the policy's last synced revision when its copy task succeeds.
func recordSyncedRevision(ctx context.Context, taskID int64, status string) error {
	if status != job.SuccessStatus.String() {
		return nil
	}
	t, err := task.Mgr.Get(ctx, taskID)
	if err != nil {
		return err
	}
	exec, err := task.ExecMgr.Get(ctx, t.ExecutionID)
	if err != nil {
		return err
	}
	revision, _ := exec.ExtraAttrs[attrRevision].(string)
	if revision == "" {
		return nil
	}
	return policy.Mgr.Update(ctx, &pkgmodel.Policy{ID: exec.VendorID, LastSyncedRevision: revision}, "LastSyncedRevision")
}

// resolverTypes returns the adapter types whose adapters can preview a copy. An adapter is
// created from an empty registry only to inspect its type; factories that need a usable
// registry to create one are not resolvers.
func resolverTypes(types []string, getFactory func(string) (adp.Factory, error)) []string {
	result := []string{}
	for _, t := range types {
		if isResolverType(t, getFactory) {
			result = append(result, t)
		}
	}
	sort.Strings(result)
	return result
}

func isResolverType(t string, getFactory func(string) (adp.Factory, error)) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	factory, err := getFactory(t)
	if err != nil {
		return false
	}
	a, err := factory.Create(&model.Registry{Type: t, URL: "https://model-sync.invalid", Credential: &model.Credential{}})
	if err != nil {
		return false
	}
	_, ok = a.(adp.ArtifactResolver)
	return ok
}

func (c *controller) ListAdapterTypes(_ context.Context) []string {
	return c.resolverTypes()
}

func (c *controller) PolicyCount(ctx context.Context, query *q.Query) (int64, error) {
	return c.policyMgr.Count(ctx, query)
}

func (c *controller) ListPolicies(ctx context.Context, query *q.Query) ([]*Policy, error) {
	policies, err := c.policyMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	result := make([]*Policy, 0, len(policies))
	for _, p := range policies {
		cp, err := c.populate(ctx, p)
		if err != nil {
			return nil, err
		}
		result = append(result, cp)
	}
	return result, nil
}

func (c *controller) GetPolicy(ctx context.Context, id int64) (*Policy, error) {
	p, err := c.policyMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.populate(ctx, p)
}

// populate fills in the registry and destination project. A missing registry or project
// leaves the fields sparse instead of failing, so the policy can still be fixed or deleted.
func (c *controller) populate(ctx context.Context, p *pkgmodel.Policy) (*Policy, error) {
	cp := &Policy{}
	if err := cp.From(p); err != nil {
		return nil, err
	}
	registry, err := c.regMgr.Get(ctx, p.RegistryID)
	switch {
	case err == nil:
		registry.Credential = nil
		cp.Registry = registry
	case errors.IsNotFoundErr(err):
		log.Warningf("registry %d of model sync policy %d not found", p.RegistryID, p.ID)
	default:
		return nil, err
	}
	proj, err := c.projectCtl.Get(ctx, p.DestProjectID)
	switch {
	case err == nil:
		cp.DestProjectName = proj.Name
	case errors.IsNotFoundErr(err):
		log.Warningf("project %d of model sync policy %d not found", p.DestProjectID, p.ID)
	default:
		return nil, err
	}
	return cp, nil
}

// resolver returns the preview capability of a registry, with its credential.
func (c *controller) resolver(ctx context.Context, registryID int64) (*model.Registry, adp.ArtifactResolver, error) {
	registry, err := c.regMgr.Get(ctx, registryID)
	if err != nil {
		return nil, nil, err
	}
	a, err := c.regMgr.CreateAdapter(ctx, registry)
	if err != nil {
		return nil, nil, err
	}
	resolver, ok := a.(adp.ArtifactResolver)
	if !ok {
		return nil, nil, errors.BadRequestError(nil).
			WithMessagef("registry %s (%s) cannot be a model sync source, supported types: %s",
				registry.Name, registry.Type, strings.Join(c.resolverTypes(), ", "))
	}
	return registry, resolver, nil
}

func (c *controller) validate(ctx context.Context, p *Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if _, _, err := c.resolver(ctx, p.Registry.ID); err != nil {
		return err
	}
	proj, err := c.projectCtl.Get(ctx, p.DestProjectID)
	if err != nil {
		return err
	}
	if proj.IsProxy() {
		return errors.BadRequestError(nil).
			WithMessagef("project %s is a proxy cache project and cannot be the destination of model sync", proj.Name)
	}
	if dest := proj.Name + "/" + p.EffectiveDestRepository(); !lib.RepositoryNameRe.MatchString(dest) {
		return errors.BadRequestError(nil).WithMessagef("invalid destination repository %q", dest)
	}
	return nil
}

func (c *controller) CreatePolicy(ctx context.Context, p *Policy) (int64, error) {
	if err := c.validate(ctx, p); err != nil {
		return 0, err
	}
	dp, err := p.To()
	if err != nil {
		return 0, err
	}
	dp.LastSyncedRevision = ""
	id, err := c.policyMgr.Create(ctx, dp)
	if err != nil {
		return 0, err
	}
	if p.IsScheduledTrigger() {
		if _, err := c.schedule(ctx, id, p.Cron); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func (c *controller) schedule(ctx context.Context, policyID int64, cron string) (int64, error) {
	cbParams := map[string]any{
		"policy_id": policyID,
		"operator":  secret.JobserviceUser,
	}
	return c.scheduler.Schedule(ctx, job.ModelSyncVendorType, policyID, "", cron, callbackFuncName, cbParams, map[string]any{})
}

func (c *controller) UpdatePolicy(ctx context.Context, p *Policy) error {
	if err := c.validate(ctx, p); err != nil {
		return err
	}
	old, err := c.policyMgr.Get(ctx, p.ID)
	if err != nil {
		return err
	}
	dp, err := p.To()
	if err != nil {
		return err
	}
	dp.Creator = old.Creator
	dp.CreationTime = old.CreationTime
	// the last synced revision describes the old source once the source changes
	if old.RegistryID != dp.RegistryID || !strings.EqualFold(old.SrcRepository, dp.SrcRepository) ||
		old.SrcRevision != dp.SrcRevision || old.FileFilter != dp.FileFilter {
		dp.LastSyncedRevision = ""
	} else {
		dp.LastSyncedRevision = old.LastSyncedRevision
	}
	if err := c.scheduler.UnScheduleByVendor(ctx, job.ModelSyncVendorType, p.ID); err != nil {
		return err
	}
	if err := c.policyMgr.Update(ctx, dp); err != nil {
		return err
	}
	if p.IsScheduledTrigger() {
		if _, err := c.schedule(ctx, p.ID, p.Cron); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) DeletePolicy(ctx context.Context, id int64) error {
	if err := c.execMgr.DeleteByVendor(ctx, job.ModelSyncVendorType, id); err != nil {
		return err
	}
	if err := c.scheduler.UnScheduleByVendor(ctx, job.ModelSyncVendorType, id); err != nil {
		return err
	}
	return c.policyMgr.Delete(ctx, id)
}

func (c *controller) Start(ctx context.Context, p *Policy, trigger string) (int64, error) {
	if !p.Enabled {
		return 0, errors.New(nil).WithCode(errors.PreconditionCode).WithMessagef("the policy %d is disabled", p.ID)
	}
	running, err := c.execMgr.Count(ctx, &q.Query{Keywords: map[string]any{
		"VendorType": job.ModelSyncVendorType,
		"VendorID":   p.ID,
		"Status":     job.RunningStatus.String(),
	}})
	if err != nil {
		return 0, err
	}
	extra := map[string]any{
		attrSrcRepository:  p.SrcRepository,
		attrSrcRevision:    p.Revision(),
		attrDestRepository: p.FullDestRepository(),
	}
	if op := operator.FromContext(ctx); op != "" {
		extra[attrOperator] = op
	}
	id, err := c.execMgr.Create(ctx, job.ModelSyncVendorType, p.ID, trigger, extra)
	if err != nil {
		return 0, err
	}
	if running > 0 {
		if err := c.execMgr.MarkError(ctx, id, "Execution skipped: another sync of this policy is still in progress."); err != nil {
			return 0, err
		}
		return id, nil
	}

	// The transaction of ctx may commit before the goroutine reads the execution, so the
	// goroutine gets its own orm context and waits for the record.
	backgroundCtx := context.WithoutCancel(ctx)
	go func() {
		c.wp.GetWorker()
		defer c.wp.ReleaseWorker()
		ctx := orm.NewContext(backgroundCtx, c.ormCreator.Create())
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("recovered from the panic of model sync execution %d: %v", id, r)
				c.markError(ctx, id, fmt.Errorf("panic during the model sync"))
			}
		}()
		if err := retry.Retry(func() error {
			_, err := c.execMgr.Get(ctx, id)
			return err
		}); err != nil {
			c.markError(ctx, id, fmt.Errorf("failed to wait for the execution record: %v", err))
			return
		}
		if err := c.run(ctx, id, p, extra); err != nil {
			c.markError(ctx, id, err)
		}
	}()
	return id, nil
}

// run resolves the source, records what the copy will produce on the execution, and starts
// the replication copy flow for the one model repository.
func (c *controller) run(ctx context.Context, executionID int64, p *Policy, extra map[string]any) error {
	registry, resolver, err := c.resolver(ctx, p.Registry.ID)
	if err != nil {
		return err
	}
	local, err := c.regMgr.Get(ctx, 0)
	if err != nil {
		return err
	}
	res, err := resolver.Resolve(p.SrcRepository, p.Revision(), p.FileFilter)
	if err != nil {
		return fmt.Errorf("failed to resolve %s@%s: %w", p.SrcRepository, p.Revision(), err)
	}
	if res.MatchedFiles == 0 {
		return errors.BadRequestError(nil).WithMessagef("no file of %s@%s matches the file filter", p.SrcRepository, p.Revision())
	}
	dest := p.FullDestRepository()
	tags := []string{p.RevisionTagFilter()}
	if res.RevisionTag != "" {
		tags = append(tags, res.RevisionTag)
	}
	extra[attrRevision] = res.Revision
	extra[attrDigest] = res.Digest
	extra[attrTags] = tags
	extra[attrFiles] = res.MatchedFiles
	extra[attrSize] = res.MatchedSize
	extra[attrSourceURL] = res.SourceURL
	extra[attrSkipped] = c.holds(ctx, dest, res.Digest, tags)
	if err := c.execMgr.UpdateExtraAttrs(ctx, executionID, extra); err != nil {
		return err
	}

	filters := []*model.Filter{
		{Type: model.FilterTypeName, Value: res.Repository},
		{Type: model.FilterTypeTag, Value: p.RevisionTagFilter()},
	}
	if !p.FileFilter.IsEmpty() {
		filters = append(filters, &model.Filter{Type: model.FilterTypeFile, Value: p.FileFilter})
	}
	copyPolicy := &repctlmodel.Policy{
		ID:             p.ID,
		Name:           p.Name,
		SrcRegistry:    registry,
		DestRegistry:   local,
		DestRepository: dest,
		Filters:        filters,
		Override:       true,
		Enabled:        true,
		RevisionTag:    true,
	}
	return c.flowCtl.Start(ctx, executionID, copyPolicy, nil)
}

// holds reports whether the destination already has the artifact with all the tags.
func (c *controller) holds(ctx context.Context, repository, digest string, tags []string) bool {
	if digest == "" {
		return false
	}
	art, err := c.artCtl.GetByReference(ctx, repository, digest, &artifact.Option{WithTag: true})
	if err != nil {
		return false
	}
	held := make([]string, 0, len(art.Tags))
	for _, t := range art.Tags {
		held = append(held, t.Name)
	}
	for _, t := range tags {
		if !slices.Contains(held, t) {
			return false
		}
	}
	return true
}

func (c *controller) markError(ctx context.Context, executionID int64, err error) {
	if e := c.execMgr.StopAndWaitWithError(ctx, executionID, 10*time.Second, err); e != nil {
		log.Errorf("failed to stop the model sync execution %d: %v", executionID, e)
	}
}

func (c *controller) Stop(ctx context.Context, id int64) error {
	if _, err := c.GetExecution(ctx, id); err != nil {
		return err
	}
	return c.execMgr.Stop(ctx, id)
}

func (c *controller) buildExecutionQuery(query *q.Query) *q.Query {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.ModelSyncVendorType
	for _, k := range []string{"PolicyID", "policy_id"} {
		if v, exist := query.Keywords[k]; exist {
			query.Keywords["VendorID"] = v
			delete(query.Keywords, k)
		}
	}
	return query
}

func (c *controller) ExecutionCount(ctx context.Context, query *q.Query) (int64, error) {
	return c.execMgr.Count(ctx, c.buildExecutionQuery(query))
}

func (c *controller) ListExecutions(ctx context.Context, query *q.Query) ([]*Execution, error) {
	execs, err := c.execMgr.List(ctx, c.buildExecutionQuery(query))
	if err != nil {
		return nil, err
	}
	result := make([]*Execution, 0, len(execs))
	for _, e := range execs {
		result = append(result, convertExecution(e))
	}
	return result, nil
}

func (c *controller) getExecution(ctx context.Context, id int64) (*task.Execution, error) {
	execs, err := c.execMgr.List(ctx, &q.Query{Keywords: map[string]any{
		"ID":         id,
		"VendorType": job.ModelSyncVendorType,
	}})
	if err != nil {
		return nil, err
	}
	if len(execs) == 0 {
		return nil, errors.NotFoundError(nil).WithMessagef("model sync execution %d not found", id)
	}
	return execs[0], nil
}

func (c *controller) GetExecution(ctx context.Context, id int64) (*Execution, error) {
	e, err := c.getExecution(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertExecution(e), nil
}

func (c *controller) TaskCount(ctx context.Context, query *q.Query) (int64, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.ModelSyncVendorType
	return c.taskMgr.Count(ctx, query)
}

func (c *controller) ListTasks(ctx context.Context, query *q.Query) ([]*Task, error) {
	query = q.MustClone(query)
	query.Keywords["VendorType"] = job.ModelSyncVendorType
	tasks, err := c.taskMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	execs := map[int64]*task.Execution{}
	result := make([]*Task, 0, len(tasks))
	for _, t := range tasks {
		e, ok := execs[t.ExecutionID]
		if !ok {
			if e, err = c.execMgr.Get(ctx, t.ExecutionID); err != nil {
				return nil, err
			}
			execs[t.ExecutionID] = e
		}
		result = append(result, convertTask(t, e))
	}
	return result, nil
}

func (c *controller) GetTask(ctx context.Context, id int64) (*Task, error) {
	tasks, err := c.ListTasks(ctx, &q.Query{Keywords: map[string]any{"ID": id}})
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.NotFoundError(nil).WithMessagef("model sync task %d not found", id)
	}
	return tasks[0], nil
}

func (c *controller) GetTaskLog(ctx context.Context, id int64) ([]byte, error) {
	if _, err := c.GetTask(ctx, id); err != nil {
		return nil, err
	}
	return c.taskMgr.GetLog(ctx, id)
}

func (c *controller) Preview(ctx context.Context, req *PreviewRequest) (*Preview, error) {
	if req == nil || req.RegistryID <= 0 {
		return nil, errors.BadRequestError(nil).WithMessage("registry id is required")
	}
	repository := strings.Trim(strings.TrimSpace(req.SrcRepository), "/")
	if repository == "" {
		return nil, errors.BadRequestError(nil).WithMessage("source repository is required")
	}
	if err := req.FileFilter.Validate(); err != nil {
		return nil, err
	}
	_, resolver, err := c.resolver(ctx, req.RegistryID)
	if err != nil {
		return nil, err
	}
	reference := strings.TrimSpace(req.SrcRevision)
	if reference == "" {
		reference = DefaultRevision
	}
	res, err := resolver.Resolve(repository, reference, req.FileFilter)
	if err != nil {
		return nil, err
	}
	return convertPreview(res), nil
}

func convertPreview(res *model.ResolveResult) *Preview {
	p := &Preview{
		Repository:   res.SourceID,
		Ref:          res.Reference,
		Revision:     res.Revision,
		RevisionTag:  res.RevisionTag,
		Digest:       res.Digest,
		SourceURL:    res.SourceURL,
		Metadata:     map[string]string{},
		Tags:         append([]string{}, res.Tags...),
		TotalFiles:   int64(res.TotalFiles),
		TotalSize:    res.TotalSize,
		MatchedFiles: int64(res.MatchedFiles),
		MatchedSize:  res.MatchedSize,
		Files:        make([]*PreviewFile, 0, len(res.Files)),
	}
	if res.RevisionTag != "" {
		p.Tags = append(p.Tags, res.RevisionTag)
	}
	if len(res.Licenses) > 0 {
		p.Metadata["license"] = strings.Join(res.Licenses, ", ")
	}
	if res.Task != "" {
		p.Metadata["task"] = res.Task
	}
	if res.LastModified != nil {
		p.Metadata["last_modified"] = res.LastModified.Format(time.RFC3339)
	}
	for _, f := range res.Files {
		p.Files = append(p.Files, &PreviewFile{Path: f.Path, Size: f.Size, SHA256: f.SHA256, Type: f.Class, Matched: f.Matched})
	}
	return p
}

func convertExecution(e *task.Execution) *Execution {
	exec := &Execution{
		ID:            e.ID,
		PolicyID:      e.VendorID,
		Status:        e.Status,
		StatusMessage: e.StatusMessage,
		Metrics:       e.Metrics,
		Trigger:       e.Trigger,
		StartTime:     e.StartTime,
		EndTime:       e.EndTime,
	}
	exec.Operator, _ = e.ExtraAttrs[attrOperator].(string)
	return exec
}

func convertTask(t *task.Task, e *task.Execution) *Task {
	tk := &Task{
		ID:            t.ID,
		ExecutionID:   t.ExecutionID,
		Status:        t.Status,
		StatusMessage: t.StatusMessage,
		RunCount:      t.RunCount,
		JobID:         t.JobID,
		CreationTime:  t.CreationTime,
		StartTime:     t.StartTime,
		UpdateTime:    t.UpdateTime,
		EndTime:       t.EndTime,
		Tags:          []string{},
	}
	if e == nil {
		return tk
	}
	str := func(k string) string {
		s, _ := e.ExtraAttrs[k].(string)
		return s
	}
	num := func(k string) int64 {
		// the execution manager decodes extra attributes with UseNumber
		switch v := e.ExtraAttrs[k].(type) {
		case json.Number:
			n, _ := v.Int64()
			return n
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		}
		return 0
	}
	tk.SrcRepository = str(attrSrcRepository)
	tk.SrcRevision = str(attrSrcRevision)
	tk.DestRepository = str(attrDestRepository)
	tk.Revision = str(attrRevision)
	tk.Digest = str(attrDigest)
	tk.Files = num(attrFiles)
	tk.Size = num(attrSize)
	tk.Skipped, _ = e.ExtraAttrs[attrSkipped].(bool)
	switch tags := e.ExtraAttrs[attrTags].(type) {
	case []any:
		for _, tag := range tags {
			if s, ok := tag.(string); ok {
				tk.Tags = append(tk.Tags, s)
			}
		}
	case []string:
		tk.Tags = append(tk.Tags, tags...)
	}
	return tk
}
