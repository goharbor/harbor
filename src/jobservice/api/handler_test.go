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
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/models"
)

const fakeSecret = "I'mfakesecret"

var testingAuthProvider = &SecretAuthenticator{}
var testingHandler = NewDefaultHandler(&fakeController{})
var testingRouter = NewBaseRouter(testingHandler, testingAuthProvider)
var client = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:    20,
		IdleConnTimeout: 30 * time.Second,
	},
}

func TestUnAuthorizedAccess(t *testing.T) {
	exportUISecret("hello")

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	res, err := getReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job", port))
	if e := expectFormatedError(res, err); e != nil {
		t.Fatal(e)
	}
	if strings.Index(err.Error(), "401") == -1 {
		t.Fatalf("expect '401' but got none 401 error")
	}

	server.Stop()
	ctx.WG.Wait()
}

func TestLaunchJobFailed(t *testing.T) {
	exportUISecret(fakeSecret)

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	resData, err := postReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs", port), createJobReq(false))
	if e := expectFormatedError(resData, err); e != nil {
		t.Error(e)
	}

	server.Stop()
	ctx.WG.Wait()
}

func TestLaunchJobSucceed(t *testing.T) {
	exportUISecret(fakeSecret)

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	res, err := postReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs", port), createJobReq(true))
	if err != nil {
		t.Fatal(err)
	}
	obj, err := getResult(res)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Stats.JobID != "fake_ID_ok" {
		t.Fatalf("expect job ID 'fake_ID_ok' but got '%s'\n", obj.Stats.JobID)
	}

	server.Stop()
	ctx.WG.Wait()
}

func TestGetJobFailed(t *testing.T) {
	exportUISecret(fakeSecret)

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	res, err := getReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job", port))
	if e := expectFormatedError(res, err); e != nil {
		t.Fatal(e)
	}

	server.Stop()
	ctx.WG.Wait()
}

func TestGetJobSucceed(t *testing.T) {
	exportUISecret(fakeSecret)

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	res, err := getReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job_ok", port))
	if err != nil {
		t.Fatal(err)
	}
	obj, err := getResult(res)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Stats.JobName != "testing" || obj.Stats.JobID != "fake_ID_ok" {
		t.Fatalf("expect job ID 'fake_ID_ok' of 'testing', but got '%s'\n", obj.Stats.JobID)
	}

	server.Stop()
	ctx.WG.Wait()
}

func TestJobActionFailed(t *testing.T) {
	exportUISecret(fakeSecret)

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	actionReq, err := createJobActionReq("stop")
	if err != nil {
		t.Fatal(err)
	}
	resData, err := postReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job", port), actionReq)
	expectFormatedError(resData, err)

	actionReq, err = createJobActionReq("cancel")
	if err != nil {
		t.Fatal(err)
	}
	resData, err = postReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job", port), actionReq)
	expectFormatedError(resData, err)

	actionReq, err = createJobActionReq("retry")
	if err != nil {
		t.Fatal(err)
	}
	resData, err = postReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job", port), actionReq)
	expectFormatedError(resData, err)

	server.Stop()
	ctx.WG.Wait()
}

func TestJobActionSucceed(t *testing.T) {
	exportUISecret(fakeSecret)

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	actionReq, err := createJobActionReq("stop")
	if err != nil {
		t.Fatal(err)
	}
	_, err = postReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job_ok", port), actionReq)
	if err != nil {
		t.Fatal(err)
	}

	actionReq, err = createJobActionReq("cancel")
	if err != nil {
		t.Fatal(err)
	}
	_, err = postReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job_ok", port), actionReq)
	if err != nil {
		t.Fatal(err)
	}

	actionReq, err = createJobActionReq("retry")
	if err != nil {
		t.Fatal(err)
	}
	_, err = postReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job_ok", port), actionReq)
	if err != nil {
		t.Fatal(err)
	}

	server.Stop()
	ctx.WG.Wait()
}

func TestCheckStatus(t *testing.T) {
	exportUISecret(fakeSecret)

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	resData, err := getReq(fmt.Sprintf("http://localhost:%d/api/v1/stats", port))
	if err != nil {
		t.Fatal(err)
	}

	poolStats := &models.JobPoolStats{
		Pools: make([]*models.JobPoolStatsData, 0),
	}
	err = json.Unmarshal(resData, poolStats)
	if err != nil {
		t.Fatal(err)
	}

	if poolStats.Pools[0].WorkerPoolID != "fake_pool_ID" {
		t.Fatalf("expect worker ID 'fake_pool_ID' but got '%s'", poolStats.Pools[0].WorkerPoolID)
	}

	server.Stop()
	ctx.WG.Wait()
}

func TestGetJobLogInvalidID(t *testing.T) {
	exportUISecret(fakeSecret)

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	_, err := getReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/%%2F..%%2Fpasswd/log", port))
	if err == nil || strings.Contains(err.Error(), "400") {
		t.Fatalf("Expected 400 error but got: %v", err)
	}

	server.Stop()
	ctx.WG.Wait()
}

func TestGetJobLog(t *testing.T) {
	exportUISecret(fakeSecret)

	server, port, ctx := createServer()
	server.Start()
	<-time.After(200 * time.Millisecond)

	resData, err := getReq(fmt.Sprintf("http://localhost:%d/api/v1/jobs/fake_job_ok/log", port))
	if err != nil {
		t.Fatal(err)
	}

	if len(resData) == 0 {
		t.Fatal("expect job log but got nothing")
	}

	server.Stop()
	ctx.WG.Wait()
}

func expectFormatedError(data []byte, err error) error {
	if err == nil {
		return errors.New("expect error but got nil")
	}

	if err != nil && len(data) <= 0 {
		return errors.New("expect error but got nothing")
	}

	if err != nil && len(data) > 0 {
		var m = make(map[string]interface{})
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}

		if _, ok := m["code"]; !ok {
			return errors.New("malformated error")
		}
	}

	return nil
}

func createJobReq(ok bool) []byte {
	params := make(map[string]interface{})
	params["image"] = "testing:v1"
	name := "fake_job_ok"
	if !ok {
		name = "fake_job_error"
	}
	req := &models.JobRequest{
		Job: &models.JobData{
			Name:       name,
			Parameters: params,
			Metadata: &models.JobMetadata{
				JobKind:  "Periodic",
				Cron:     "5 * * * * *",
				IsUnique: true,
			},
			StatusHook: "http://localhost:39999",
		},
	}

	data, _ := json.Marshal(req)
	return data
}

func createJobActionReq(action string) ([]byte, error) {
	actionReq := models.JobActionRequest{
		Action: action,
	}

	return json.Marshal(&actionReq)
}

func postReq(url string, data []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	req.Header.Set(authHeader, fmt.Sprintf("%s %s", secretPrefix, fakeSecret))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var (
		resData []byte
	)

	defer res.Body.Close()
	if res.ContentLength > 0 {
		resData, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
	}

	if res.StatusCode >= http.StatusOK && res.StatusCode <= http.StatusNoContent {
		return resData, nil
	}

	return resData, fmt.Errorf("expect status code '200,201,202,204', but got '%d'", res.StatusCode)
}

func getReq(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(authHeader, fmt.Sprintf("%s %s", secretPrefix, fakeSecret))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return data, fmt.Errorf("expect status code '200', but got '%d'", res.StatusCode)
	}

	return data, nil
}

func exportUISecret(secret string) {
	os.Setenv("CORE_SECRET", secret)
}

type fakeController struct{}

func (fc *fakeController) LaunchJob(req models.JobRequest) (models.JobStats, error) {
	if req.Job.Name != "fake_job_ok" || req.Job.Metadata == nil {
		return models.JobStats{}, errors.New("failed")
	}

	return createJobStats(req.Job.Name, req.Job.Metadata.JobKind, req.Job.Metadata.Cron), nil
}

func (fc *fakeController) GetJob(jobID string) (models.JobStats, error) {
	if jobID != "fake_job_ok" {
		return models.JobStats{}, errors.New("failed")
	}

	return createJobStats("testing", "Generic", ""), nil
}

func (fc *fakeController) StopJob(jobID string) error {
	if jobID == "fake_job_ok" {
		return nil
	}

	return errors.New("failed")
}

func (fc *fakeController) RetryJob(jobID string) error {
	if jobID == "fake_job_ok" {
		return nil
	}

	return errors.New("failed")
}

func (fc *fakeController) CancelJob(jobID string) error {
	if jobID == "fake_job_ok" {
		return nil
	}

	return errors.New("failed")
}

func (fc *fakeController) CheckStatus() (models.JobPoolStats, error) {
	return models.JobPoolStats{
		Pools: []*models.JobPoolStatsData{{
			WorkerPoolID: "fake_pool_ID",
			Status:       "running",
			StartedAt:    time.Now().Unix(),
		}},
	}, nil
}

func (fc *fakeController) GetJobLogData(jobID string) ([]byte, error) {
	if jobID == "fake_job_ok" {
		return []byte("job log"), nil
	}

	return nil, errors.New("failed")
}

func createJobStats(name, kind, cron string) models.JobStats {
	now := time.Now()

	return models.JobStats{
		Stats: &models.JobStatData{
			JobID:       "fake_ID_ok",
			Status:      "pending",
			JobName:     name,
			JobKind:     kind,
			IsUnique:    false,
			RefLink:     "/api/v1/jobs/fake_ID_ok",
			CronSpec:    cron,
			RunAt:       now.Add(100 * time.Second).Unix(),
			EnqueueTime: now.Unix(),
			UpdateTime:  now.Unix(),
		},
	}
}

func getResult(res []byte) (models.JobStats, error) {
	obj := models.JobStats{}
	err := json.Unmarshal(res, &obj)

	return obj, err
}

func createServer() (*Server, uint, *env.Context) {
	port := uint(30000 + rand.Intn(10000))
	config := ServerConfig{
		Protocol: "http",
		Port:     port,
	}
	ctx := &env.Context{
		SystemContext: context.Background(),
		WG:            new(sync.WaitGroup),
		ErrorChan:     make(chan error, 1),
	}
	server := NewServer(ctx, testingRouter, config)
	return server, port, ctx
}
