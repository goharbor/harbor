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
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	ctltag "github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib"
	liberrors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	pkgmodel "github.com/goharbor/harbor/src/pkg/modelsync/policy/model"
	"github.com/goharbor/harbor/src/pkg/project/models"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/goharbor/harbor/src/pkg/task"
	testingartifact "github.com/goharbor/harbor/src/testing/controller/artifact"
	testingproject "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
	testingpolicy "github.com/goharbor/harbor/src/testing/pkg/modelsync/policy"
	testingreg "github.com/goharbor/harbor/src/testing/pkg/reg"
	testingadapter "github.com/goharbor/harbor/src/testing/pkg/reg/adapter"
	testingscheduler "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	testingtask "github.com/goharbor/harbor/src/testing/pkg/task"
)

// resolverAdapter is an adapter that can preview copies.
type resolverAdapter struct {
	testingadapter.Adapter
	result *model.ResolveResult
	err    error
	calls  []string
}

func (r *resolverAdapter) Resolve(repository, reference string, _ *model.FileFilter) (*model.ResolveResult, error) {
	r.calls = append(r.calls, repository+"@"+reference)
	return r.result, r.err
}

type fakeFlow struct {
	executionID int64
	policy      *repctlmodel.Policy
	err         error
}

func (f *fakeFlow) Start(_ context.Context, executionID int64, policy *repctlmodel.Policy, _ *model.Resource) error {
	f.executionID = executionID
	f.policy = policy
	return f.err
}

type controllerTestSuite struct {
	suite.Suite
	ctl        *controller
	policyMgr  *testingpolicy.Manager
	execMgr    *testingtask.ExecutionManager
	taskMgr    *testingtask.Manager
	regMgr     *testingreg.Manager
	projectCtl *testingproject.Controller
	artCtl     *testingartifact.Controller
	scheduler  *testingscheduler.Scheduler
	flow       *fakeFlow
	resolver   *resolverAdapter
	hf         *model.Registry
}

func (s *controllerTestSuite) SetupTest() {
	s.policyMgr = &testingpolicy.Manager{}
	s.execMgr = &testingtask.ExecutionManager{}
	s.taskMgr = &testingtask.Manager{}
	s.regMgr = &testingreg.Manager{}
	s.projectCtl = &testingproject.Controller{}
	s.artCtl = &testingartifact.Controller{}
	s.scheduler = &testingscheduler.Scheduler{}
	s.flow = &fakeFlow{}
	s.resolver = &resolverAdapter{result: &model.ResolveResult{
		Repository:   "huggingfacetb/smollm2-135m",
		SourceID:     "HuggingFaceTB/SmolLM2-135M",
		Reference:    "main",
		Revision:     "93efa2f097d58c2a74874c7e644dbc9b0cee75a2",
		RevisionTag:  "sha-93efa2f097d5",
		SourceURL:    "https://huggingface.co/HuggingFaceTB/SmolLM2-135M",
		Tags:         []string{"main"},
		Digest:       "sha256:2dc8fb68cff37ff1f1ae1bcef28039072cb9a012d4e7c7c817cee9d3d1b3b18e",
		Licenses:     []string{"apache-2.0"},
		Task:         "text-generation",
		TotalFiles:   9,
		MatchedFiles: 6,
		TotalSize:    272447115,
		MatchedSize:  2910522,
		Files: []*model.ResolvedFile{
			{Path: "config.json", Size: 704, Class: model.FileClassConfig, SHA256: "abc", Matched: true},
			{Path: "model.safetensors", Size: 269060552, Class: model.FileClassWeight, SHA256: "def"},
		},
	}}
	s.hf = &model.Registry{ID: 1, Name: "hf", Type: "huggingface", URL: "https://huggingface.co"}
	s.ctl = &controller{
		policyMgr:     s.policyMgr,
		execMgr:       s.execMgr,
		taskMgr:       s.taskMgr,
		regMgr:        s.regMgr,
		projectCtl:    s.projectCtl,
		artCtl:        s.artCtl,
		scheduler:     s.scheduler,
		flowCtl:       s.flow,
		wp:            lib.NewWorkerPool(1),
		resolverTypes: func() []string { return []string{"huggingface"} },
	}
}

func (s *controllerTestSuite) policy() *Policy {
	return &Policy{
		ID:            7,
		Name:          "smollm",
		Enabled:       true,
		Registry:      &model.Registry{ID: 1},
		SrcRepository: "HuggingFaceTB/SmolLM2-135M",
		FileFilter:    &model.FileFilter{Include: []string{"*.json"}},
		DestProjectID: 3,
		TriggerType:   pkgmodel.TriggerTypeManual,
	}
}

func (s *controllerTestSuite) mockResolver() {
	s.regMgr.On("Get", mock.Anything, int64(1)).Return(s.hf, nil)
	s.regMgr.On("CreateAdapter", mock.Anything, s.hf).Return(s.resolver, nil)
}

func (s *controllerTestSuite) TestCreatePolicy() {
	s.mockResolver()
	s.projectCtl.On("Get", mock.Anything, int64(3)).Return(&models.Project{ProjectID: 3, Name: "models"}, nil)
	s.policyMgr.On("Create", mock.Anything, testifymock.MatchedBy(func(p *pkgmodel.Policy) bool {
		return p.RegistryID == 1 && p.SrcRepository == "HuggingFaceTB/SmolLM2-135M" &&
			p.FileFilter == `{"include":["*.json"]}` && p.LastSyncedRevision == ""
	})).Return(int64(7), nil)
	p := s.policy()
	p.TriggerType = pkgmodel.TriggerTypeScheduled
	p.Cron = "0 0 * * * *"
	s.scheduler.On("Schedule", mock.Anything, job.ModelSyncVendorType, int64(7), "", "0 0 * * * *",
		callbackFuncName, mock.Anything, mock.Anything).Return(int64(1), nil)

	id, err := s.ctl.CreatePolicy(context.TODO(), p)
	s.Require().NoError(err)
	s.Equal(int64(7), id)
	s.scheduler.AssertExpectations(s.T())
}

func (s *controllerTestSuite) TestCreatePolicyRejects() {
	// a registry whose adapter cannot resolve
	other := &model.Registry{ID: 2, Name: "hub", Type: "docker-hub"}
	s.regMgr.On("Get", mock.Anything, int64(2)).Return(other, nil)
	s.regMgr.On("CreateAdapter", mock.Anything, other).Return(&testingadapter.Adapter{}, nil)
	p := s.policy()
	p.Registry = &model.Registry{ID: 2}
	_, err := s.ctl.CreatePolicy(context.TODO(), p)
	s.True(liberrors.IsErr(err, liberrors.BadRequestCode), "%v", err)
	s.Contains(err.Error(), "cannot be a model sync source")

	// a proxy cache destination
	s.mockResolver()
	s.projectCtl.On("Get", mock.Anything, int64(4)).Return(&models.Project{ProjectID: 4, Name: "proxy", RegistryID: 1}, nil)
	p = s.policy()
	p.DestProjectID = 4
	_, err = s.ctl.CreatePolicy(context.TODO(), p)
	s.True(liberrors.IsErr(err, liberrors.BadRequestCode), "%v", err)

	// an invalid destination repository name
	s.projectCtl.On("Get", mock.Anything, int64(3)).Return(&models.Project{ProjectID: 3, Name: "models"}, nil)
	p = s.policy()
	p.DestRepository = "Upper/Case"
	_, err = s.ctl.CreatePolicy(context.TODO(), p)
	s.True(liberrors.IsErr(err, liberrors.BadRequestCode), "%v", err)
	s.policyMgr.AssertNotCalled(s.T(), "Create", mock.Anything, mock.Anything)
}

func (s *controllerTestSuite) TestUpdatePolicyLastSyncedRevision() {
	s.mockResolver()
	s.projectCtl.On("Get", mock.Anything, int64(3)).Return(&models.Project{ProjectID: 3, Name: "models"}, nil)
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	old := &pkgmodel.Policy{
		ID: 7, RegistryID: 1, SrcRepository: "HuggingFaceTB/SmolLM2-135M", FileFilter: `{"include":["*.json"]}`,
		LastSyncedRevision: "93efa2f097d58c2a74874c7e644dbc9b0cee75a2", Creator: "admin", CreationTime: created,
	}
	s.policyMgr.On("Get", mock.Anything, int64(7)).Return(old, nil)
	s.scheduler.On("UnScheduleByVendor", mock.Anything, job.ModelSyncVendorType, int64(7)).Return(nil)
	var updated []*pkgmodel.Policy
	s.policyMgr.On("Update", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		updated = append(updated, args.Get(1).(*pkgmodel.Policy))
	}).Return(nil)

	// same source: the revision stays
	p := s.policy()
	p.Description = "new"
	s.Require().NoError(s.ctl.UpdatePolicy(context.TODO(), p))
	// another filter: it describes the old source
	p = s.policy()
	p.FileFilter = &model.FileFilter{Include: []string{"*.json"}, Exclude: []string{"vocab.json"}}
	s.Require().NoError(s.ctl.UpdatePolicy(context.TODO(), p))

	s.Require().Len(updated, 2)
	s.Equal(old.LastSyncedRevision, updated[0].LastSyncedRevision)
	s.Equal("admin", updated[0].Creator)
	s.Equal(created, updated[0].CreationTime)
	s.Equal("", updated[1].LastSyncedRevision)
}

func (s *controllerTestSuite) TestRun() {
	s.mockResolver()
	local := &model.Registry{ID: 0, Name: "Local", Type: "harbor"}
	s.regMgr.On("Get", mock.Anything, int64(0)).Return(local, nil)
	s.artCtl.On("GetByReference", mock.Anything, "models/huggingfacetb/smollm2-135m", s.resolver.result.Digest, mock.Anything).
		Return(&artifact.Artifact{Tags: []*ctltag.Tag{{Tag: tag.Tag{Name: "main"}}, {Tag: tag.Tag{Name: "sha-93efa2f097d5"}}}}, nil)
	var extra map[string]any
	s.execMgr.On("UpdateExtraAttrs", mock.Anything, int64(11), mock.Anything).Run(func(args mock.Arguments) {
		extra = args.Get(2).(map[string]any)
	}).Return(nil)

	p := s.policy()
	p.DestProjectName = "models"
	s.Require().NoError(s.ctl.run(context.TODO(), 11, p, map[string]any{attrOperator: "admin"}))

	s.Equal([]string{"HuggingFaceTB/SmolLM2-135M@main"}, s.resolver.calls)
	s.Equal("admin", extra[attrOperator])
	s.Equal("93efa2f097d58c2a74874c7e644dbc9b0cee75a2", extra[attrRevision])
	s.Equal(s.resolver.result.Digest, extra[attrDigest])
	s.Equal([]string{"main", "sha-93efa2f097d5"}, extra[attrTags])
	s.Equal(true, extra[attrSkipped])

	s.Equal(int64(11), s.flow.executionID)
	cp := s.flow.policy
	s.Require().NotNil(cp)
	s.Equal(s.hf, cp.SrcRegistry)
	s.Equal(local, cp.DestRegistry)
	s.Equal("models/huggingfacetb/smollm2-135m", cp.DestRepository)
	s.True(cp.RevisionTag)
	s.True(cp.Override)
	s.Require().Len(cp.Filters, 3)
	s.Equal(model.FilterTypeName, cp.Filters[0].Type)
	s.Equal("huggingfacetb/smollm2-135m", cp.Filters[0].Value)
	s.Equal(model.FilterTypeTag, cp.Filters[1].Type)
	s.Equal("main", cp.Filters[1].Value)
	opts, err := cp.SourceOptions()
	s.Require().NoError(err)
	s.Equal([]string{"*.json"}, opts.FileFilter.Include)
	s.True(opts.RevisionTag)
}

func (s *controllerTestSuite) TestRunNothingMatches() {
	s.mockResolver()
	s.regMgr.On("Get", mock.Anything, int64(0)).Return(&model.Registry{}, nil)
	s.resolver.result.MatchedFiles = 0
	s.resolver.result.Digest = ""
	err := s.ctl.run(context.TODO(), 11, s.policy(), map[string]any{})
	s.True(liberrors.IsErr(err, liberrors.BadRequestCode), "%v", err)
	s.Nil(s.flow.policy)
	s.execMgr.AssertNotCalled(s.T(), "UpdateExtraAttrs", mock.Anything, mock.Anything, mock.Anything)
}

func (s *controllerTestSuite) TestRunResolveError() {
	s.mockResolver()
	s.regMgr.On("Get", mock.Anything, int64(0)).Return(&model.Registry{}, nil)
	s.resolver.err = liberrors.NotFoundError(nil).WithMessage("model not found")
	err := s.ctl.run(context.TODO(), 11, s.policy(), map[string]any{})
	s.Require().Error(err)
	s.Contains(err.Error(), "model not found")
	s.Nil(s.flow.policy)
}

func (s *controllerTestSuite) TestStartDisabledAndBusy() {
	p := s.policy()
	p.Enabled = false
	_, err := s.ctl.Start(context.TODO(), p, task.ExecutionTriggerManual)
	s.True(liberrors.IsErr(err, liberrors.PreconditionCode), "%v", err)

	s.execMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	s.execMgr.On("Create", mock.Anything, job.ModelSyncVendorType, int64(7), task.ExecutionTriggerManual, mock.Anything).Return(int64(12), nil)
	s.execMgr.On("MarkError", mock.Anything, int64(12), mock.Anything).Return(nil)
	id, err := s.ctl.Start(context.TODO(), s.policy(), task.ExecutionTriggerManual)
	s.Require().NoError(err)
	s.Equal(int64(12), id)
	s.execMgr.AssertCalled(s.T(), "MarkError", mock.Anything, int64(12), mock.Anything)
}

func (s *controllerTestSuite) TestPreview() {
	s.mockResolver()
	_, err := s.ctl.Preview(context.TODO(), &PreviewRequest{RegistryID: 1})
	s.True(liberrors.IsErr(err, liberrors.BadRequestCode), "%v", err)
	_, err = s.ctl.Preview(context.TODO(), &PreviewRequest{RegistryID: 1, SrcRepository: "a/b",
		FileFilter: &model.FileFilter{Include: []string{"[a"}}})
	s.True(liberrors.IsErr(err, liberrors.BadRequestCode), "%v", err)

	pv, err := s.ctl.Preview(context.TODO(), &PreviewRequest{RegistryID: 1, SrcRepository: " HuggingFaceTB/SmolLM2-135M/ "})
	s.Require().NoError(err)
	s.Equal([]string{"HuggingFaceTB/SmolLM2-135M@main"}, s.resolver.calls)
	s.Equal("HuggingFaceTB/SmolLM2-135M", pv.Repository)
	s.Equal([]string{"main", "sha-93efa2f097d5"}, pv.Tags)
	s.Equal("sha-93efa2f097d5", pv.RevisionTag)
	s.Equal(s.resolver.result.Digest, pv.Digest)
	s.Equal("apache-2.0", pv.Metadata["license"])
	s.Equal("text-generation", pv.Metadata["task"])
	s.Equal(int64(6), pv.MatchedFiles)
	s.Require().Len(pv.Files, 2)
	s.Equal(model.FileClassWeight, pv.Files[1].Type)
	s.False(pv.Files[1].Matched)
	// the resolver result's tags are not modified
	s.Equal([]string{"main"}, s.resolver.result.Tags)
}

func (s *controllerTestSuite) TestListTasks() {
	s.taskMgr.On("List", mock.Anything, testifymock.MatchedBy(func(query *q.Query) bool {
		return query.Keywords["VendorType"] == job.ModelSyncVendorType
	})).Return([]*task.Task{{ID: 21, ExecutionID: 11, Status: job.SuccessStatus.String()}}, nil)
	// as read back from the JSON column
	s.execMgr.On("Get", mock.Anything, int64(11)).Return(&task.Execution{ID: 11, ExtraAttrs: map[string]any{
		attrSrcRepository:  "HuggingFaceTB/SmolLM2-135M",
		attrDestRepository: "models/huggingfacetb/smollm2-135m",
		attrRevision:       "93efa2f097d58c2a74874c7e644dbc9b0cee75a2",
		attrDigest:         "sha256:2dc8",
		attrTags:           []any{"main", "sha-93efa2f097d5"},
		attrFiles:          json.Number("6"),
		attrSize:           float64(2910522),
		attrSkipped:        true,
	}}, nil)
	tasks, err := s.ctl.ListTasks(context.TODO(), q.New(q.KeyWords{"ExecutionID": int64(11)}))
	s.Require().NoError(err)
	s.Require().Len(tasks, 1)
	t := tasks[0]
	s.Equal("models/huggingfacetb/smollm2-135m", t.DestRepository)
	s.Equal([]string{"main", "sha-93efa2f097d5"}, t.Tags)
	s.Equal(int64(6), t.Files)
	s.Equal(int64(2910522), t.Size)
	s.True(t.Skipped)
}

func (s *controllerTestSuite) TestDeletePolicy() {
	s.execMgr.On("DeleteByVendor", mock.Anything, job.ModelSyncVendorType, int64(7)).Return(nil)
	s.scheduler.On("UnScheduleByVendor", mock.Anything, job.ModelSyncVendorType, int64(7)).Return(nil)
	s.policyMgr.On("Delete", mock.Anything, int64(7)).Return(nil)
	s.Require().NoError(s.ctl.DeletePolicy(context.TODO(), 7))
	s.policyMgr.AssertExpectations(s.T())
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}

type factoryFunc func(*model.Registry) (adp.Adapter, error)

func (f factoryFunc) Create(r *model.Registry) (adp.Adapter, error) { return f(r) }
func (f factoryFunc) AdapterPattern() *model.AdapterPattern         { return nil }

func TestResolverTypes(t *testing.T) {
	factories := map[string]adp.Factory{
		"huggingface": factoryFunc(func(*model.Registry) (adp.Adapter, error) { return &resolverAdapter{}, nil }),
		"docker-hub":  factoryFunc(func(*model.Registry) (adp.Adapter, error) { return &testingadapter.Adapter{}, nil }),
		"broken":      factoryFunc(func(*model.Registry) (adp.Adapter, error) { return nil, errors.New("needs a real registry") }),
		"panicking":   factoryFunc(func(r *model.Registry) (adp.Adapter, error) { panic(r.Credential.AccessKey + "!") }),
	}
	get := func(t string) (adp.Factory, error) {
		if f, ok := factories[t]; ok {
			return f, nil
		}
		return nil, errors.New("unknown")
	}
	got := resolverTypes([]string{"panicking", "docker-hub", "huggingface", "broken", "unknown"}, get)
	if len(got) != 1 || got[0] != "huggingface" {
		t.Fatalf("resolverTypes = %v, want [huggingface]", got)
	}
}

func TestPolicyValidate(t *testing.T) {
	valid := func() *Policy {
		return &Policy{Name: "p", Registry: &model.Registry{ID: 1}, SrcRepository: "a/b", DestProjectID: 1,
			TriggerType: pkgmodel.TriggerTypeManual}
	}
	cases := map[string]func(p *Policy){
		"empty name":      func(p *Policy) { p.Name = " " },
		"no registry":     func(p *Policy) { p.Registry = nil },
		"no repository":   func(p *Policy) { p.SrcRepository = "" },
		"commit revision": func(p *Policy) { p.SrcRevision = "93efa2f097d58c2a74874c7e644dbc9b0cee75a2" },
		"no project":      func(p *Policy) { p.DestProjectID = 0 },
		"bad filter":      func(p *Policy) { p.FileFilter = &model.FileFilter{Exclude: []string{"[a"}} },
		"manual cron":     func(p *Policy) { p.Cron = "0 0 * * * *" },
		"scheduled empty": func(p *Policy) { p.TriggerType = pkgmodel.TriggerTypeScheduled },
		"bad trigger":     func(p *Policy) { p.TriggerType = "event_based" },
	}
	if err := valid().Validate(); err != nil {
		t.Fatalf("valid policy: %v", err)
	}
	for name, mutate := range cases {
		p := valid()
		mutate(p)
		if err := p.Validate(); !liberrors.IsErr(err, liberrors.BadRequestCode) {
			t.Errorf("%s: got %v, want a bad request", name, err)
		}
	}
}

func TestPolicyNaming(t *testing.T) {
	p := &Policy{SrcRepository: "HuggingFaceTB/SmolLM2-135M", DestProjectName: "models"}
	if got := p.FullDestRepository(); got != "models/huggingfacetb/smollm2-135m" {
		t.Errorf("FullDestRepository = %q", got)
	}
	if got := p.RevisionTagFilter(); got != "main" {
		t.Errorf("RevisionTagFilter = %q", got)
	}
	p.SrcRevision = "refs/pr/1"
	p.DestRepository = "/llm/smol/"
	if got := p.RevisionTagFilter(); got != "refs-pr-1" {
		t.Errorf("RevisionTagFilter = %q", got)
	}
	if got := p.FullDestRepository(); got != "models/llm/smol" {
		t.Errorf("FullDestRepository = %q", got)
	}
}
