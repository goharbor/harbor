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
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/jobservice/common/query"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	secretKey  = "CORE_SECRET"
	fakeSecret = "I'mfakesecret"
)

// APIHandlerTestSuite tests functions of API handler
type APIHandlerTestSuite struct {
	suite.Suite

	server     *Server
	controller *fakeController
	APIAddr    string
	client     *http.Client
	cancel     context.CancelFunc
}

// SetupSuite prepares test suite
func (suite *APIHandlerTestSuite) SetupSuite() {
	_ = os.Setenv(secretKey, fakeSecret)

	suite.client = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    20,
			IdleConnTimeout: 30 * time.Second,
		},
	}

	suite.createServer()

	go func() {
		_ = suite.server.Start()
	}()

	<-time.After(200 * time.Millisecond)
}

// TearDownSuite clears test suite
func (suite *APIHandlerTestSuite) TearDownSuite() {
	_ = os.Unsetenv(secretKey)
	_ = suite.server.Stop()
	suite.cancel()
}

// TestAPIHandlerTestSuite is suite entry for 'go test'
func TestAPIHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(APIHandlerTestSuite))
}

// TestUnAuthorizedAccess ...
func (suite *APIHandlerTestSuite) TestUnAuthorizedAccess() {
	_ = os.Unsetenv(secretKey)
	defer func() {
		_ = os.Setenv(secretKey, fakeSecret)
	}()

	_, code := suite.getReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job"))
	assert.Equal(suite.T(), 401, code, "expect '401' but got none 401 error")
}

// TestLaunchJobFailed ...
func (suite *APIHandlerTestSuite) TestLaunchJobFailed() {
	req := createJobReq()
	bytes, _ := json.Marshal(req)

	fc1 := &fakeController{}
	fc1.On("LaunchJob", req).Return(nil, errs.BadRequestError(req.Job.Name))
	suite.controller = fc1
	_, code := suite.postReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs"), bytes)
	assert.Equal(suite.T(), 400, code, "expect 400 bad request but got %d", code)

	fc2 := &fakeController{}
	fc2.On("LaunchJob", req).Return(nil, errs.ConflictError(req.Job.Name))
	suite.controller = fc2
	_, code = suite.postReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs"), bytes)
	assert.Equal(suite.T(), 409, code, "expect 409 conflict but got %d", code)

	fc3 := &fakeController{}
	fc3.On("LaunchJob", req).Return(nil, errs.LaunchJobError(errors.New("testing launch job")))
	suite.controller = fc3
	_, code = suite.postReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs"), bytes)
	assert.Equal(suite.T(), 500, code, "expect 500 internal server error but got %d", code)
}

// TestLaunchJobSucceed ...
func (suite *APIHandlerTestSuite) TestLaunchJobSucceed() {
	req := createJobReq()
	bytes, _ := json.Marshal(req)

	fc := &fakeController{}
	fc.On("LaunchJob", req).Return(createJobStats("sample", "Generic", ""), nil)
	suite.controller = fc

	_, code := suite.postReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs"), bytes)
	assert.Equal(suite.T(), 202, code, "expected 202 created but got %d when launching job", code)
}

// TestGetJobFailed ...
func (suite *APIHandlerTestSuite) TestGetJobFailed() {
	fc := &fakeController{}
	fc.On("GetJob", "fake_job_ID").Return(nil, errs.NoObjectFoundError("fake_job_ID"))
	suite.controller = fc

	_, code := suite.getReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID"))
	assert.Equal(suite.T(), 404, code, "expected 404 not found but got %d when getting job", code)
}

// TestGetJobSucceed ...
func (suite *APIHandlerTestSuite) TestGetJobSucceed() {
	fc := &fakeController{}
	fc.On("GetJob", "fake_job_ID").Return(createJobStats("sample", "Generic", ""), nil)
	suite.controller = fc

	res, code := suite.getReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID"))
	require.Equal(suite.T(), 200, code, "expected 200 ok but got %d when getting job", code)
	stats, err := getResult(res)
	require.Nil(suite.T(), err, "no error should be occurred when unmarshal job stats")
	assert.Equal(suite.T(), "fake_job_ID", stats.Info.JobID, "expected job ID 'fake_job_ID' but got %s", stats.Info.JobID)
}

// TestJobActionFailed ...
func (suite *APIHandlerTestSuite) TestJobActionFailed() {
	actionReq := createJobActionReq("not-support")
	data, _ := json.Marshal(actionReq)
	_, code := suite.postReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID"), data)
	assert.Equal(suite.T(), 501, code, "expected 501 not implemented but got %d", code)

	fc1 := &fakeController{}
	fc1.On("StopJob", "fake_job_ID_not").Return(errs.NoObjectFoundError("fake_job_ID_not"))
	suite.controller = fc1
	actionReq = createJobActionReq("stop")
	data, _ = json.Marshal(actionReq)
	_, code = suite.postReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID_not"), data)
	assert.Equal(suite.T(), 404, code, "expected 404 not found but got %d", code)

	fc2 := &fakeController{}
	fc2.On("StopJob", "fake_job_ID").Return(errs.BadRequestError("fake_job_ID"))
	suite.controller = fc2
	_, code = suite.postReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID"), data)
	assert.Equal(suite.T(), 400, code, "expected 400 bad request but got %d", code)

	fc3 := &fakeController{}
	fc3.On("StopJob", "fake_job_ID").Return(errs.StopJobError(errors.New("testing error")))
	suite.controller = fc3
	_, code = suite.postReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID"), data)
	assert.Equal(suite.T(), 500, code, "expected 500 internal server but got %d", code)
}

// TestJobActionSucceed ...
func (suite *APIHandlerTestSuite) TestJobActionSucceed() {
	fc := &fakeController{}
	fc.On("StopJob", "fake_job_ID_not").Return(nil)
	suite.controller = fc
	actionReq := createJobActionReq("stop")
	data, _ := json.Marshal(actionReq)
	_, code := suite.postReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID_not"), data)
	assert.Equal(suite.T(), 204, code, "expected 204 no content but got %d", code)
}

// TestCheckStatus ...
func (suite *APIHandlerTestSuite) TestCheckStatus() {
	statsRes := &worker.Stats{
		Pools: []*worker.StatsData{
			{
				WorkerPoolID: "my-worker-pool-ID",
			},
		},
	}
	fc := &fakeController{}
	fc.On("CheckStatus").Return(statsRes, nil)
	suite.controller = fc

	bytes, code := suite.getReq(fmt.Sprintf("%s/%s", suite.APIAddr, "stats"))
	require.Equal(suite.T(), 200, code, "expected 200 ok when getting worker stats but got %d", code)

	poolStats := &worker.Stats{
		Pools: make([]*worker.StatsData, 0),
	}
	err := json.Unmarshal(bytes, poolStats)
	assert.Nil(suite.T(), err, "no error should be occurred when unmarshal worker stats")
	assert.Equal(suite.T(), 1, len(poolStats.Pools), "at least 1 pool exists but got %d", len(poolStats.Pools))
	assert.Equal(suite.T(), "my-worker-pool-ID", poolStats.Pools[0].WorkerPoolID, "expected pool ID 'my-worker-pool-ID' but got %s", poolStats.Pools[0].WorkerPoolID)
}

// TestGetJobLogInvalidID ...
func (suite *APIHandlerTestSuite) TestGetJobLogInvalidID() {
	fc := &fakeController{}
	fc.On("GetJobLogData", "fake_job_ID_not").Return(nil, errs.NoObjectFoundError("fake_job_ID_not"))
	suite.controller = fc

	_, code := suite.getReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID_not/log"))
	assert.Equal(suite.T(), 404, code, "expected 404 not found but got %d", code)
}

// TestGetJobLog ...
func (suite *APIHandlerTestSuite) TestGetJobLog() {
	fc := &fakeController{}
	fc.On("GetJobLogData", "fake_job_ID").Return([]byte("hello log"), nil)
	suite.controller = fc

	resData, code := suite.getReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID/log"))
	require.Equal(suite.T(), 200, code, "expected 200 ok but got %d", code)
	assert.Equal(suite.T(), "hello log", string(resData))
}

// TestGetPeriodicExecutionsWithoutQuery ...
func (suite *APIHandlerTestSuite) TestGetPeriodicExecutionsWithoutQuery() {
	q := &query.Parameter{
		PageNumber: 1,
		PageSize:   query.DefaultPageSize,
		Extras:     make(query.ExtraParameters),
	}

	fc := &fakeController{}
	fc.On("GetPeriodicExecutions", "fake_job_ID", q).
		Return([]*job.Stats{createJobStats("sample", "Generic", "")}, int64(1), nil)
	suite.controller = fc

	_, code := suite.getReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID/executions"))
	assert.Equal(suite.T(), 200, code, "expected 200 ok but got %d", code)
}

// TestGetPeriodicExecutionsWithQuery ...
func (suite *APIHandlerTestSuite) TestGetPeriodicExecutionsWithQuery() {
	extras := make(query.ExtraParameters)
	extras.Set(query.ExtraParamKeyNonStoppedOnly, true)
	q := &query.Parameter{
		PageNumber: 2,
		PageSize:   50,
		Extras:     extras,
	}

	fc := &fakeController{}
	fc.On("GetPeriodicExecutions", "fake_job_ID", q).
		Return([]*job.Stats{createJobStats("sample", "Generic", "")}, int64(1), nil)
	suite.controller = fc

	_, code := suite.getReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/fake_job_ID/executions?page_number=2&page_size=50&non_dead_only=true"))
	assert.Equal(suite.T(), 200, code, "expected 200 ok but got %d", code)
}

// TestScheduledJobs ...
func (suite *APIHandlerTestSuite) TestScheduledJobs() {
	q := &query.Parameter{
		PageNumber: 2,
		PageSize:   50,
		Extras:     make(query.ExtraParameters),
	}

	fc := &fakeController{}
	fc.On("ScheduledJobs", q).
		Return([]*job.Stats{createJobStats("sample", "Generic", "")}, int64(1), nil)
	suite.controller = fc

	_, code := suite.getReq(fmt.Sprintf("%s/%s", suite.APIAddr, "jobs/scheduled?page_number=2&page_size=50"))
	assert.Equal(suite.T(), 200, code, "expected 200 ok but got %d", code)
}

// createServer ...
func (suite *APIHandlerTestSuite) createServer() {
	port := uint(30000 + rand.Intn(1000))
	suite.APIAddr = fmt.Sprintf("http://localhost:%d/api/v1", port)

	config := ServerConfig{
		Protocol: "http",
		Port:     port,
	}
	ctx, cancel := context.WithCancel(context.Background())

	testingRouter := NewBaseRouter(
		NewDefaultHandler(suite),
		&SecretAuthenticator{},
	)
	suite.server = NewServer(ctx, testingRouter, config)
	suite.cancel = cancel
}

// postReq ...
func (suite *APIHandlerTestSuite) postReq(url string, data []byte) ([]byte, int) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(data)))
	if err != nil {
		return nil, 0
	}

	req.Header.Set(authHeader, fmt.Sprintf("%s %s", secretPrefix, fakeSecret))

	res, err := suite.client.Do(req)
	if err != nil {
		return nil, 0
	}

	var (
		resData []byte
	)

	defer func() {
		_ = res.Body.Close()
	}()
	if res.ContentLength > 0 {
		resData, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, 0
		}
	}

	return resData, res.StatusCode
}

// getReq ...
func (suite *APIHandlerTestSuite) getReq(url string) ([]byte, int) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0
	}

	req.Header.Set(authHeader, fmt.Sprintf("%s %s", secretPrefix, fakeSecret))

	res, err := suite.client.Do(req)
	if err != nil {
		return nil, 0
	}

	defer func() {
		_ = res.Body.Close()
	}()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, 0
	}

	return data, res.StatusCode
}

func (suite *APIHandlerTestSuite) LaunchJob(req *job.Request) (*job.Stats, error) {
	return suite.controller.LaunchJob(req)
}

func (suite *APIHandlerTestSuite) GetJob(jobID string) (*job.Stats, error) {
	return suite.controller.GetJob(jobID)
}

func (suite *APIHandlerTestSuite) StopJob(jobID string) error {
	return suite.controller.StopJob(jobID)
}

func (suite *APIHandlerTestSuite) RetryJob(jobID string) error {
	return suite.controller.RetryJob(jobID)
}

func (suite *APIHandlerTestSuite) CheckStatus() (*worker.Stats, error) {
	return suite.controller.CheckStatus()
}

func (suite *APIHandlerTestSuite) GetJobLogData(jobID string) ([]byte, error) {
	return suite.controller.GetJobLogData(jobID)
}

func (suite *APIHandlerTestSuite) GetPeriodicExecutions(periodicJobID string, query *query.Parameter) ([]*job.Stats, int64, error) {
	return suite.controller.GetPeriodicExecutions(periodicJobID, query)
}

func (suite *APIHandlerTestSuite) ScheduledJobs(query *query.Parameter) ([]*job.Stats, int64, error) {
	return suite.controller.ScheduledJobs(query)
}

type fakeController struct {
	mock.Mock
}

func (fc *fakeController) LaunchJob(req *job.Request) (*job.Stats, error) {
	args := fc.Called(req)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*job.Stats), nil
}

func (fc *fakeController) GetJob(jobID string) (*job.Stats, error) {
	args := fc.Called(jobID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*job.Stats), nil
}

func (fc *fakeController) StopJob(jobID string) error {
	args := fc.Called(jobID)
	return args.Error(0)
}

func (fc *fakeController) RetryJob(jobID string) error {
	args := fc.Called(jobID)
	return args.Error(0)
}

func (fc *fakeController) CheckStatus() (*worker.Stats, error) {
	args := fc.Called()
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*worker.Stats), nil
}

func (fc *fakeController) GetJobLogData(jobID string) ([]byte, error) {
	args := fc.Called(jobID)
	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), nil
}

func (fc *fakeController) GetPeriodicExecutions(periodicJobID string, query *query.Parameter) ([]*job.Stats, int64, error) {
	args := fc.Called(periodicJobID, query)
	if args.Error(2) != nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}

	return args.Get(0).([]*job.Stats), args.Get(1).(int64), nil
}

func (fc *fakeController) ScheduledJobs(query *query.Parameter) ([]*job.Stats, int64, error) {
	args := fc.Called(query)
	if args.Error(2) != nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}

	return args.Get(0).([]*job.Stats), args.Get(1).(int64), nil
}

func createJobStats(name, kind, cron string) *job.Stats {
	now := time.Now()
	params := make(job.Parameters)
	params["image"] = "testing:v1"

	return &job.Stats{
		Info: &job.StatsInfo{
			JobID:       "fake_job_ID",
			Status:      job.PendingStatus.String(),
			JobName:     name,
			JobKind:     kind,
			IsUnique:    false,
			RefLink:     "/api/v1/jobs/fake_job_ID",
			CronSpec:    cron,
			RunAt:       now.Add(100 * time.Second).Unix(),
			EnqueueTime: now.Unix(),
			UpdateTime:  now.Unix(),
			Parameters:  params,
		},
	}
}

func getResult(res []byte) (*job.Stats, error) {
	obj := &job.Stats{}
	err := json.Unmarshal(res, obj)

	return obj, err
}

func createJobReq() *job.Request {
	params := make(job.Parameters)
	params["image"] = "testing:v1"

	return &job.Request{
		Job: &job.RequestBody{
			Name:       "my-testing-job",
			Parameters: params,
			Metadata: &job.Metadata{
				JobKind:  "Periodic",
				Cron:     "5 * * * * *",
				IsUnique: true,
			},
			StatusHook: "http://localhost:39999",
		},
	}
}

func createJobActionReq(action string) *job.ActionRequest {
	return &job.ActionRequest{
		Action: action,
	}
}
