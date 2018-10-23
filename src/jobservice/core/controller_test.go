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
package core

import (
	"errors"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/errs"

	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/models"
)

func TestLaunchGenericJob(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)
	req := createJobReq("Generic", false, false)
	res, err := c.LaunchJob(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Stats.JobID != "fake_ID" {
		t.Fatalf("expect enqueued job ID 'fake_ID' but got '%s'\n", res.Stats.JobID)
	}
}

func TestLaunchGenericJobUnique(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)
	req := createJobReq("Generic", true, false)
	res, err := c.LaunchJob(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Stats.JobID != "fake_ID" {
		t.Fatalf("expect enqueued job ID 'fake_ID' but got '%s'\n", res.Stats.JobID)
	}
}

func TestLaunchGenericJobWithHook(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)
	req := createJobReq("Generic", false, true)
	res, err := c.LaunchJob(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Stats.JobID != "fake_ID" {
		t.Fatalf("expect enqueued job ID 'fake_ID' but got '%s'\n", res.Stats.JobID)
	}
}

func TestLaunchScheduledJob(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)
	req := createJobReq("Scheduled", false, true)
	res, err := c.LaunchJob(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Stats.JobID != "fake_ID_Scheduled" {
		t.Fatalf("expect enqueued job ID 'fake_ID_Scheduled' but got '%s'\n", res.Stats.JobID)
	}
}

func TestLaunchScheduledUniqueJob(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)
	req := createJobReq("Scheduled", true, false)
	res, err := c.LaunchJob(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Stats.JobID != "fake_ID_Scheduled" {
		t.Fatalf("expect enqueued job ID 'fake_ID_Scheduled' but got '%s'\n", res.Stats.JobID)
	}
}

func TestLaunchPeriodicJob(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)
	req := createJobReq("Periodic", true, false)
	res, err := c.LaunchJob(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.Stats.JobID != "fake_ID_Periodic" {
		t.Fatalf("expect enqueued job ID 'fake_ID_Periodic' but got '%s'\n", res.Stats.JobID)
	}
}

func TestGetJobStats(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)
	stats, err := c.GetJob("fake_ID")
	if err != nil {
		t.Fatal(err)
	}

	if stats.Stats.Status != "running" {
		t.Fatalf("expect stauts 'running' but got '%s'\n", stats.Stats.Status)
	}
}

func TestJobActions(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)

	if err := c.StopJob("fake_ID"); err != nil {
		t.Fatal(err)
	}

	if err := c.CancelJob("fake_ID"); err != nil {
		t.Fatal(err)
	}

	if err := c.RetryJob("fake_ID"); err != nil {
		t.Fatal(err)
	}
}

func TestGetJobLogData(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)

	if _, err := c.GetJobLogData("fake_ID"); err != nil {
		if !errs.IsObjectNotFoundError(err) {
			t.Errorf("expect object not found error but got '%s'\n", err)
		}
	} else {
		t.Fatal("expect error but got nil")
	}
}

func TestCheckStatus(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)

	st, err := c.CheckStatus()
	if err != nil {
		t.Fatal(err)
	}

	if len(st.Pools) == 0 {
		t.Fatal("expect status data but got zero list")
	}

	if st.Pools[0].Status != "running" {
		t.Fatalf("expect status 'running' but got '%s'\n", st.Pools[0].Status)
	}
}

func TestInvalidCheck(t *testing.T) {
	pool := &fakePool{}
	c := NewController(pool)

	req := models.JobRequest{
		Job: &models.JobData{
			Name: "DEMO",
			Metadata: &models.JobMetadata{
				JobKind: "kind",
			},
		},
	}

	if _, err := c.LaunchJob(req); err == nil {
		t.Fatal("error expected but got nil")
	}

	req.Job.Name = "fake"
	if _, err := c.LaunchJob(req); err == nil {
		t.Fatal("error expected but got nil")
	}

	req.Job.Metadata.JobKind = "Scheduled"
	if _, err := c.LaunchJob(req); err == nil {
		t.Fatal("error expected but got nil")
	}

	req.Job.Metadata.JobKind = "Periodic"
	req.Job.Metadata.Cron = "x x x x x x"
	if _, err := c.LaunchJob(req); err == nil {
		t.Fatal("error expected but got nil")
	}
}

func createJobReq(kind string, isUnique bool, withHook bool) models.JobRequest {
	params := make(map[string]interface{})
	params["name"] = "testing"
	req := models.JobRequest{
		Job: &models.JobData{
			Name:       "DEMO",
			Parameters: params,
			Metadata: &models.JobMetadata{
				JobKind:       kind,
				IsUnique:      isUnique,
				ScheduleDelay: 100,
				Cron:          "5 * * * * *",
			},
		},
	}

	if withHook {
		req.Job.StatusHook = "http://localhost:9090"
	}

	return req
}

type fakePool struct{}

func (f *fakePool) Start() error {
	return nil
}

func (f *fakePool) RegisterJob(name string, job interface{}) error {
	return nil
}

func (f *fakePool) RegisterJobs(jobs map[string]interface{}) error {
	return nil
}

func (f *fakePool) Enqueue(jobName string, params models.Parameters, isUnique bool) (models.JobStats, error) {
	return models.JobStats{
		Stats: &models.JobStatData{
			JobID: "fake_ID",
		},
	}, nil
}

func (f *fakePool) Schedule(jobName string, params models.Parameters, runAfterSeconds uint64, isUnique bool) (models.JobStats, error) {
	return models.JobStats{
		Stats: &models.JobStatData{
			JobID: "fake_ID_Scheduled",
		},
	}, nil
}

func (f *fakePool) PeriodicallyEnqueue(jobName string, params models.Parameters, cronSetting string) (models.JobStats, error) {
	return models.JobStats{
		Stats: &models.JobStatData{
			JobID: "fake_ID_Periodic",
		},
	}, nil
}

func (f *fakePool) Stats() (models.JobPoolStats, error) {
	return models.JobPoolStats{
		Pools: []*models.JobPoolStatsData{
			{
				Status: "running",
			},
		},
	}, nil
}

func (f *fakePool) IsKnownJob(name string) (interface{}, bool) {
	return (*fakeJob)(nil), true
}

func (f *fakePool) ValidateJobParameters(jobType interface{}, params map[string]interface{}) error {
	return nil
}

func (f *fakePool) GetJobStats(jobID string) (models.JobStats, error) {
	return models.JobStats{
		Stats: &models.JobStatData{
			JobID:  "fake_ID",
			Status: "running",
		},
	}, nil
}

func (f *fakePool) StopJob(jobID string) error {
	return nil
}

func (f *fakePool) CancelJob(jobID string) error {
	return nil
}

func (f *fakePool) RetryJob(jobID string) error {
	return nil
}

func (f *fakePool) RegisterHook(jobID string, hookURL string) error {
	return nil
}

type fakeJob struct{}

func (j *fakeJob) MaxFails() uint {
	return 3
}

func (j *fakeJob) ShouldRetry() bool {
	return true
}

func (j *fakeJob) Validate(params map[string]interface{}) error {
	if p, ok := params["name"]; ok {
		if p == "testing" {
			return nil
		}
	}

	return errors.New("testing error")
}

func (j *fakeJob) Run(ctx env.JobContext, params map[string]interface{}) error {
	return nil
}
