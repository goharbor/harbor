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

package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	tmock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/modelsync"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	pkgmodel "github.com/goharbor/harbor/src/pkg/modelsync/policy/model"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/task"
	taskdao "github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	modelsynctesting "github.com/goharbor/harbor/src/testing/controller/modelsync"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

type ModelSyncTestSuite struct {
	htesting.Suite
	ctl *modelsynctesting.Controller
}

func (suite *ModelSyncTestSuite) SetupSuite() {
	suite.ctl = &modelsynctesting.Controller{}
	suite.Config = &restapi.Config{
		ModelSyncAPI: &modelSyncAPI{ctl: suite.ctl},
	}
	suite.Suite.SetupSuite()
}

func (suite *ModelSyncTestSuite) allow(times int) {
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("GetUsername").Return("admin").Maybe()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)
}

func body(v any) io.Reader {
	buf, _ := json.Marshal(v)
	return bytes.NewBuffer(buf)
}

func (suite *ModelSyncTestSuite) samplePolicy() *modelsync.Policy {
	return &modelsync.Policy{
		ID: 1, Name: "qwen", Enabled: true, Creator: "admin",
		Registry:      &model.Registry{ID: 2, Name: "hf", Type: model.RegistryTypeHuggingFace, URL: "https://huggingface.co"},
		SrcRepository: "Qwen/Qwen3-8B", FileFilter: &model.FileFilter{Include: []string{"*.safetensors"}, Exclude: []string{"original/**"}},
		DestProjectID: 1, DestProjectName: "library", TriggerType: pkgmodel.TriggerTypeScheduled, Cron: "0 0 * * * *",
		LastSyncedRevision: "abc", CreationTime: time.Now(), UpdateTime: time.Now(),
	}
}

func (suite *ModelSyncTestSuite) TestAuthorization() {
	reqs := []struct {
		method string
		url    string
		body   any
	}{
		{http.MethodGet, "/model-sync/policies", nil},
		{http.MethodPost, "/model-sync/policies", models.ModelSyncPolicy{Name: "p"}},
		{http.MethodGet, "/model-sync/policies/1", nil},
		{http.MethodPut, "/model-sync/policies/1", models.ModelSyncPolicy{Name: "p"}},
		{http.MethodDelete, "/model-sync/policies/1", nil},
		{http.MethodGet, "/model-sync/policies/1/executions", nil},
		{http.MethodPost, "/model-sync/policies/1/executions", nil},
		{http.MethodGet, "/model-sync/executions/1", nil},
		{http.MethodPut, "/model-sync/executions/1", nil},
		{http.MethodGet, "/model-sync/executions/1/tasks", nil},
		{http.MethodGet, "/model-sync/executions/1/tasks/1/log", nil},
		{http.MethodPost, "/model-sync/preview", models.ModelSyncPreviewRequest{RegistryID: new(int64), SrcRepository: new(string)}},
		{http.MethodGet, "/model-sync/adapters", nil},
	}
	for _, req := range reqs {
		suite.Security.On("IsAuthenticated").Return(false).Once()
		res, err := suite.DoReq(req.method, req.url, body(req.body))
		suite.NoError(err)
		suite.Equal(401, res.StatusCode, req.url)

		suite.Security.On("IsAuthenticated").Return(true).Once()
		suite.Security.On("GetUsername").Return("user").Once()
		suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(false).Once()
		res, err = suite.DoReq(req.method, req.url, body(req.body))
		suite.NoError(err)
		suite.Equal(403, res.StatusCode, req.url)
	}
}

func (suite *ModelSyncTestSuite) TestPolicyCRUD() {
	suite.allow(7)
	// not found
	mock.OnAnything(suite.ctl, "GetPolicy").Return(nil, errors.NotFoundError(nil)).Once()
	res, err := suite.Get("/model-sync/policies/99")
	suite.NoError(err)
	suite.Equal(404, res.StatusCode)

	var created *modelsync.Policy
	suite.ctl.On("CreatePolicy", mock.Anything, mock.Anything).Run(func(args tmock.Arguments) {
		created = args.Get(1).(*modelsync.Policy)
	}).Return(int64(1), nil).Once()
	res, err = suite.PostJSON("/model-sync/policies", models.ModelSyncPolicy{
		Name: "qwen", RegistryID: 2, SrcRepository: "Qwen/Qwen3-8B", DestProjectID: 1, Enabled: true,
		FileFilters: []string{"*.safetensors", " "}, FileExcludes: []string{"original/**"},
		Trigger: &models.ModelSyncTrigger{Type: "scheduled", TriggerSettings: &models.ModelSyncTriggerTriggerSettings{Cron: "0 0 * * * *"}},
	})
	suite.NoError(err)
	suite.Equal(201, res.StatusCode)
	suite.Equal("/api/v2.0/model-sync/policies/1", res.Header.Get("Location"))
	suite.Require().NotNil(created)
	suite.Equal("admin", created.Creator)
	suite.Equal(int64(2), created.Registry.ID)
	suite.Equal("0 0 * * * *", created.Cron)
	suite.Equal(pkgmodel.TriggerTypeScheduled, created.TriggerType)
	suite.Equal(&model.FileFilter{Include: []string{"*.safetensors"}, Exclude: []string{"original/**"}}, created.FileFilter)

	// create with nested registry and default trigger
	suite.ctl.On("CreatePolicy", mock.Anything, mock.Anything).Run(func(args tmock.Arguments) {
		created = args.Get(1).(*modelsync.Policy)
	}).Return(int64(2), nil).Once()
	res, err = suite.PostJSON("/model-sync/policies", models.ModelSyncPolicy{Name: "x", Registry: &models.Registry{ID: 5}, SrcRepository: "a/b", DestProjectID: 1})
	suite.NoError(err)
	suite.Equal(201, res.StatusCode)
	suite.Equal(int64(5), created.Registry.ID)
	suite.Equal(pkgmodel.TriggerTypeManual, created.TriggerType)
	suite.Nil(created.FileFilter)

	// list
	mock.OnAnything(suite.ctl, "PolicyCount").Return(int64(1), nil)
	mock.OnAnything(suite.ctl, "ListPolicies").Return([]*modelsync.Policy{suite.samplePolicy()}, nil)
	var policies []*models.ModelSyncPolicy
	res, err = suite.GetJSON("/model-sync/policies?q=name%3Dqwen", &policies)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Equal("1", res.Header.Get("X-Total-Count"))
	suite.Require().Len(policies, 1)
	suite.Equal("qwen/qwen3-8b", policies[0].DestRepository)
	suite.Equal("library", policies[0].DestProjectName)
	suite.Equal(int64(2), policies[0].RegistryID)
	suite.Equal("hf", policies[0].Registry.Name)
	suite.Equal("0 0 * * * *", policies[0].Trigger.TriggerSettings.Cron)
	suite.Equal("abc", policies[0].LastSyncedRevision)
	suite.Equal([]string{"*.safetensors"}, policies[0].FileFilters)
	suite.Equal([]string{"original/**"}, policies[0].FileExcludes)

	// get
	mock.OnAnything(suite.ctl, "GetPolicy").Return(suite.samplePolicy(), nil)
	var p models.ModelSyncPolicy
	res, err = suite.GetJSON("/model-sync/policies/1", &p)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Equal("qwen", p.Name)

	// update keeps creator
	var updated *modelsync.Policy
	suite.ctl.On("UpdatePolicy", mock.Anything, mock.Anything).Run(func(args tmock.Arguments) {
		updated = args.Get(1).(*modelsync.Policy)
	}).Return(nil).Once()
	res, err = suite.PutJSON("/model-sync/policies/1", models.ModelSyncPolicy{Name: "qwen2", RegistryID: 2, SrcRepository: "a/b", DestProjectID: 1})
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Equal(int64(1), updated.ID)
	suite.Equal("admin", updated.Creator)

	// delete
	mock.OnAnything(suite.ctl, "DeletePolicy").Return(nil).Once()
	res, err = suite.Delete("/model-sync/policies/1")
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
}

func (suite *ModelSyncTestSuite) TestExecutionsAndTasks() {
	suite.allow(8)
	// start + list executions read the policy
	mock.OnAnything(suite.ctl, "GetPolicy").Return(suite.samplePolicy(), nil).Times(2)
	mock.OnAnything(suite.ctl, "Start").Return(int64(10), nil).Once()
	res, err := suite.Post("/model-sync/policies/1/executions", nil)
	suite.NoError(err)
	suite.Equal(201, res.StatusCode)
	suite.Equal("/api/v2.0/model-sync/policies/1/executions/10", res.Header.Get("Location"))

	exec := &modelsync.Execution{ID: 10, PolicyID: 1, Status: job.SuccessStatus.String(), Trigger: task.ExecutionTriggerManual, Operator: "admin",
		Metrics: &taskdao.Metrics{TaskCount: 1, SuccessTaskCount: 1}}
	mock.OnAnything(suite.ctl, "ExecutionCount").Return(int64(1), nil)
	mock.OnAnything(suite.ctl, "ListExecutions").Return([]*modelsync.Execution{exec}, nil)
	var execs []*models.ModelSyncExecution
	res, err = suite.GetJSON("/model-sync/policies/1/executions?status=Succeed&trigger=manual", &execs)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Require().Len(execs, 1)
	suite.Equal("Succeed", execs[0].Status)
	suite.Equal("manual", execs[0].Trigger)
	suite.Equal(int64(1), execs[0].Succeed)

	mock.OnAnything(suite.ctl, "GetExecution").Return(exec, nil)
	var e models.ModelSyncExecution
	res, err = suite.GetJSON("/model-sync/executions/10", &e)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Equal("admin", e.Operator)

	mock.OnAnything(suite.ctl, "Stop").Return(nil).Once()
	res, err = suite.Put("/model-sync/executions/10", nil)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)

	tk := &modelsync.Task{ID: 11, ExecutionID: 10, Status: job.RunningStatus.String(), SrcRepository: "Qwen/Qwen3-8B",
		DestRepository: "library/qwen/qwen3-8b", Revision: "abc", Digest: "sha256:x", Tags: []string{"sha-abc"}, Files: 3, Size: 100}
	mock.OnAnything(suite.ctl, "TaskCount").Return(int64(1), nil)
	mock.OnAnything(suite.ctl, "ListTasks").Return([]*modelsync.Task{tk}, nil)
	var tasks []*models.ModelSyncTask
	res, err = suite.GetJSON("/model-sync/executions/10/tasks?status=InProgress", &tasks)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Require().Len(tasks, 1)
	suite.Equal("InProgress", tasks[0].Status)
	suite.Equal([]string{"sha-abc"}, tasks[0].Tags)
	suite.Equal(int64(3), tasks[0].Files)

	// the task doesn't belong to the execution
	mock.OnAnything(suite.ctl, "GetTask").Return(&modelsync.Task{ID: 11, ExecutionID: 99}, nil).Once()
	res, err = suite.Get("/model-sync/executions/10/tasks/11/log")
	suite.NoError(err)
	suite.Equal(404, res.StatusCode)

	mock.OnAnything(suite.ctl, "GetTask").Return(tk, nil)
	mock.OnAnything(suite.ctl, "GetTaskLog").Return([]byte("the log"), nil)
	res, err = suite.Get("/model-sync/executions/10/tasks/11/log")
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	data, _ := io.ReadAll(res.Body)
	suite.Equal("the log", string(data))
}

func (suite *ModelSyncTestSuite) TestPreviewAndAdapters() {
	suite.allow(2)
	var req *modelsync.PreviewRequest
	suite.ctl.On("Preview", mock.Anything, mock.Anything).Run(func(args tmock.Arguments) {
		req = args.Get(1).(*modelsync.PreviewRequest)
	}).Return(&modelsync.Preview{
		Repository: "Qwen/Qwen3-8B", Ref: "main", Revision: "abc", Tags: []string{"main", "sha-abc"},
		RevisionTag: "sha-abc", Digest: "sha256:d", Metadata: map[string]string{"license": "apache-2.0"},
		TotalFiles: 2, TotalSize: 110, MatchedFiles: 1, MatchedSize: 100,
		Files: []*modelsync.PreviewFile{{Path: "model.safetensors", Size: 100, Type: "weight", Matched: true}, {Path: "README.md", Size: 10, Type: "doc"}},
	}, nil).Once()
	regID := int64(2)
	repo := "Qwen/Qwen3-8B"
	var preview models.ModelSyncPreview
	res, err := suite.DoReq(http.MethodPost, "/model-sync/preview", body(models.ModelSyncPreviewRequest{RegistryID: &regID, SrcRepository: &repo,
		FileFilters: []string{"*.safetensors"}, FileExcludes: []string{"*.md"}}))
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.NoError(json.NewDecoder(res.Body).Decode(&preview))
	suite.Equal(int64(2), req.RegistryID)
	suite.Equal(&model.FileFilter{Include: []string{"*.safetensors"}, Exclude: []string{"*.md"}}, req.FileFilter)
	suite.Equal("abc", preview.Revision)
	suite.Equal("sha-abc", preview.RevisionTag)
	suite.Equal("sha256:d", preview.Digest)
	suite.Equal("apache-2.0", preview.Metadata["license"])
	suite.Require().Len(preview.Files, 2)
	suite.True(preview.Files[0].Matched)
	suite.Equal(int64(1), preview.MatchedFiles)

	mock.OnAnything(suite.ctl, "ListAdapterTypes").Return([]string{"huggingface"}).Once()
	var types []string
	res, err = suite.GetJSON("/model-sync/adapters", &types)
	suite.NoError(err)
	suite.Equal(200, res.StatusCode)
	suite.Equal([]string{"huggingface"}, types)
}

func TestModelSyncTestSuite(t *testing.T) {
	suite.Run(t, &ModelSyncTestSuite{})
}
